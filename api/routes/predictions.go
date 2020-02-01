//
//   Copyright © 2019 Uncharted Software Inc.
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package routes

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"net/http"
	"path"
	"sort"
	"strconv"
	"time"

	"github.com/araddon/dateparse"
	"github.com/pkg/errors"
	log "github.com/unchartedsoftware/plog"
	"goji.io/v3/pat"

	"github.com/uncharted-distil/distil-compute/metadata"
	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil-compute/primitive/compute"
	"github.com/uncharted-distil/distil/api/env"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/task"
	"github.com/uncharted-distil/distil/api/util/json"
)

// PredictionsHandler receives a file and produces results using the specified
// fitted solution id
func PredictionsHandler(outputPath string, dataStorageCtor api.DataStorageCtor, solutionStorageCtor api.SolutionStorageCtor,
	metaStorageCtor api.MetadataStorageCtor, config *env.Config, ingestConfig *task.IngestTaskConfig) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		dataset := pat.Param(r, "dataset")
		fittedSolutionID := pat.Param(r, "fitted-solution-id")
		targetType := pat.Param(r, "target-type")

		solutionStorage, err := solutionStorageCtor()
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to initialize solution storage"))
			return
		}
		metaStorage, err := metaStorageCtor()
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to initialize metadata storage"))
			return
		}
		dataStorage, err := dataStorageCtor()
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to initialize data storage"))
			return
		}

		// get the solution id from the fitted solution ID
		solutionResults, err := solutionStorage.FetchSolutionResultsByFittedSolutionID(fittedSolutionID)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to fetch solution results fitted solution id"))
			return
		}
		if len(solutionResults) == 0 {
			handleError(w, errors.Errorf("unable to map fitted solution id to dataset or solution id"))
			return
		}
		sr := solutionResults[0]

		// read the metadata of the original dataset
		datasetES, err := metaStorage.FetchDataset(sr.Dataset, false, false)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to fetch dataset from es"))
			return
		}

		var data []byte
		if targetType == "timeseries" {
			// passed in params will be start and step count
			params, err := getPostParameters(r)
			if err != nil {
				handleError(w, errors.Wrap(err, "Unable to parse post parameters"))
				return
			}
			stepCount, ok := json.Int(params, "count")
			if !ok {
				handleError(w, errors.Errorf("Unable to parse count parameter"))
				return
			}
			startStr, ok := json.String(params, "start")
			if !ok {
				handleError(w, errors.Errorf("Unable to parse start parameter"))
				return
			}

			data, err = createTimeseriesFromRequest(dataStorage, datasetES, startStr, stepCount)
			if err != nil {
				handleError(w, errors.Wrap(err, "unable to create timeseries datat"))
				return
			}
			log.Infof("created timeseries data to use for predictions for dataset %s solution %s", dataset, fittedSolutionID)
		} else {
			// read the file from the request
			data, err = receiveFile(r)
			if err != nil {
				handleError(w, errors.Wrap(err, "unable to receive file from request"))
				return
			}
			log.Infof("received data to use for predictions for dataset %s solution %s", dataset, fittedSolutionID)
		}

		// get the source dataset from the fitted solution ID
		req, err := solutionStorage.FetchRequestByFittedSolutionID(fittedSolutionID)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to fetch request using fitted solution id"))
			return
		}

		schemaPath := path.Join(env.ResolvePath(datasetES.Source, datasetES.Folder), compute.D3MDataSchema)
		meta, err := metadata.LoadMetadataFromOriginalSchema(schemaPath)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to load metadata from source dataset schema doc"))
			return
		}

		res, err := task.Predict(meta, dataset, sr.SolutionID, fittedSolutionID, data, outputPath, config.ESDatasetsIndex, getTarget(req), metaStorage, dataStorage, solutionStorage, ingestConfig)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to generate predictions"))
			return
		}

		// marshal data and sent the response back
		err = handleJSON(w, res)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal result histogram into JSON"))
			return
		}
	}
}

func getTarget(request *api.Request) string {
	for _, f := range request.Features {
		if f.FeatureType == "target" {
			return f.FeatureName
		}
	}

	return ""
}

func createTimeseriesFromRequest(dataStorage api.DataStorage, datasetES *api.Dataset, startStr string, stepCount int) ([]byte, error) {
	// need to create timeseries based on start time and step count
	var groupingVar *model.Variable
	for _, v := range datasetES.Variables {
		if v.Grouping != nil {
			groupingVar = v
			break
		}
	}

	// find the timsetamp column and id columns
	timestampCol := groupingVar.Grouping.Properties.XCol
	var timestampVar *model.Variable
	for _, v := range datasetES.Variables {
		if v.Name == timestampCol {
			timestampVar = v
			break
		}
	}

	// get the distinct values for the id columns
	idValues := make(map[string][]string)
	for _, vID := range groupingVar.Grouping.SubIDs {
		vals, err := dataStorage.FetchRawDistinctValues(datasetES.ID, datasetES.StorageName, vID)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to fetch distinct values for '%s' from data storage", vID)
		}
		idValues[vID] = vals
	}

	// get the step duration
	timestampValues, err := dataStorage.FetchRawDistinctValues(datasetES.ID, datasetES.StorageName, timestampCol)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to fetch distinct timestamp values from data storage")
	}

	// generate timestamps to use for prediction based on type of timestamp
	var timestampPredictionValues []string
	if model.IsTimestamp(timestampVar.Type) {
		timestampPredictionValues, err = generateTimestampValues(timestampValues, startStr, stepCount)
	} else if model.IsNumerical(timestampVar.Type) {
		timestampPredictionValues, err = generateIntValues(timestampValues, startStr, stepCount)
	} else {
		return nil, errors.Errorf("timestamp variable '%s' is type '%s' which is not supported for timeseries creation", timestampVar.Name, timestampVar.Type)
	}
	if err != nil {
		return nil, err
	}

	return createTimeseriesData(idValues, timestampCol, timestampPredictionValues)
}

func createTimeseriesData(seriesFields map[string][]string, timestampFieldName string, timestampPredictionValues []string) ([]byte, error) {
	// create the header and the ids to use to generate the timeseries
	header := make([]string, 0)
	ids := make([][]string, 0)
	for name, field := range seriesFields {
		ids = append(ids, field)
		header = append(header, name)
	}
	header = append(header, timestampFieldName)

	// write the header
	outputBytes := &bytes.Buffer{}
	writerOutput := csv.NewWriter(outputBytes)
	err := writerOutput.Write(header)
	if err != nil {
		return nil, err
	}

	// treat the timestamp values as just another set of values to generate on
	ids = append(ids, timestampPredictionValues)

	// the cartesian product will generate all the values needed for the timeseries
	cartesianData := createGroupings(ids)
	err = writerOutput.WriteAll(cartesianData)
	if err != nil {
		return nil, err
	}

	writerOutput.Flush()

	return outputBytes.Bytes(), nil
}

func createGroupings(ids [][]string) [][]string {
	// end condition when empty list passed in
	if len(ids) == 0 {
		return [][]string{nil}
	}

	// use recursion to get cartesian product
	nested := createGroupings(ids[1:])

	// create the combined output
	output := make([][]string, 0)
	for _, id := range ids[0] {
		for _, product := range nested {
			output = append(output, append([]string{id}, product...))
		}
	}
	return output
}

func generateIntValues(existingValues []string, startStr string, stepCount int) ([]string, error) {
	start, err := strconv.Atoi(startStr)
	if err != nil {
		return nil, errors.Errorf("Unable to parse start into int")
	}

	// order the existing values and derive the duration between steps
	existingValuesParsed := make([]int, 0)
	for _, vs := range existingValues {
		v, err := strconv.Atoi(vs)
		if err != nil {
			continue
		}
		existingValuesParsed = append(existingValuesParsed, v)
	}
	sort.Slice(existingValuesParsed, func(i int, j int) bool {
		return existingValuesParsed[i] < existingValuesParsed[j]
	})
	stepDuration := 0
	if len(existingValuesParsed) > 1 {
		stepDuration = existingValuesParsed[1] - existingValuesParsed[0]
	}

	// iterate until all required steps are created
	currentValue := start
	timeData := make([]string, 0)
	for i := 0; i < stepCount; i++ {
		timeData = append(timeData, fmt.Sprintf("%d", currentValue))
		currentValue = currentValue + stepDuration
	}

	return timeData, nil
}

func generateTimestampValues(existingValues []string, startStr string, stepCount int) ([]string, error) {
	// parse the start time
	start, err := dateparse.ParseAny(startStr)
	if err != nil {
		return nil, errors.Errorf("Unable to parse start into time")
	}

	// order the timestamp values and derive the duration between steps (assumes a format that can be parsed)
	timestampsParsed := make([]time.Time, 0)
	for _, ts := range existingValues {
		t, err := dateparse.ParseAny(ts)
		if err != nil {
			continue
		}
		timestampsParsed = append(timestampsParsed, t)
	}
	sort.Slice(timestampsParsed, func(i int, j int) bool {
		return timestampsParsed[i].Before(timestampsParsed[j])
	})
	stepDuration := time.Duration(0)
	if len(timestampsParsed) > 1 {
		stepDuration = timestampsParsed[1].Sub(timestampsParsed[0])
	}

	// iterate until all required steps are created
	currentTime := start
	timeData := make([]string, 0)
	for i := 0; i < stepCount; i++ {
		timeData = append(timeData, currentTime.String())
		currentTime = currentTime.Add(stepDuration)
	}

	return timeData, nil
}
