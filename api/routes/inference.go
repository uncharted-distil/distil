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
	"io/ioutil"
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
	"github.com/uncharted-distil/distil/api/dataset"
	"github.com/uncharted-distil/distil/api/env"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/task"
	"github.com/uncharted-distil/distil/api/util/json"
)

// InferenceHandler receives a file and produces results using the specified
// fitted solution id
func InferenceHandler(outputPath string, dataStorageCtor api.DataStorageCtor, solutionStorageCtor api.SolutionStorageCtor,
	metaStorageCtor api.MetadataStorageCtor, config *env.Config) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		datasetName := pat.Param(r, "dataset")
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
		datasetImported := false
		datasetIngested := false
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

			datasetImported, data, err = createTimeseriesFromRequest(dataStorage, datasetES, startStr, stepCount)
			if err != nil {
				handleError(w, errors.Wrap(err, "unable to create timeseries datat"))
				return
			}
			log.Infof("created timeseries data to use for predictions for dataset %s solution %s", datasetName, fittedSolutionID)
		} else if targetType == "image" {
			// type cant be a post param since the upload is the actual data
			queryValues := r.URL.Query()
			imageType := queryValues["image"]
			if len(imageType) == 0 {
				handleError(w, errors.Errorf("no image type specified"))
				return
			}

			// read the file from the request
			data, err = receiveFile(r)
			if err != nil {
				handleError(w, errors.Wrap(err, "unable to receive file from request"))
				return
			}

			datasetImported, data, err = createImageFromRequest(data, datasetName, outputPath, imageType[0], config)
			if err != nil {
				handleError(w, errors.Wrap(err, "unable to create image dataset from request"))
				return
			}
		} else {
			// read the file from the request
			data, err = receiveFile(r)
			if err != nil {
				handleError(w, errors.Wrap(err, "unable to receive file from request"))
				return
			}
			log.Infof("received data to use for predictions for dataset %s solution %s", datasetName, fittedSolutionID)
		}

		// get the source dataset from the fitted solution ID
		req, err := solutionStorage.FetchRequestByFittedSolutionID(fittedSolutionID)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to fetch request using fitted solution id"))
			return
		}

		schemaPath := path.Join(env.ResolvePath(datasetES.Source, datasetES.Folder), compute.D3MDataSchema)
		meta, err := metadata.LoadMetadataFromOriginalSchema(schemaPath, true)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to load metadata from source dataset schema doc"))
			return
		}

		target := getTarget(req)

		// In the case of grouped variables, the target will not be variable itself, but one of its property
		// values.  We need to fetch using the original dataset, since it will have grouped variable info,
		// and then resolve the actual target.
		targetVar, err := metaStorage.FetchVariable(meta.ID, target)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to get target var from metadata storage"))
			return
		}

		ds, err := dataset.NewTableDataset(datasetName, data, false)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to create inference dataset"))
			return
		}

		predictParams := &task.PredictParams{
			Meta:               meta,
			Dataset:            datasetName,
			SolutionID:         sr.SolutionID,
			FittedSolutionID:   fittedSolutionID,
			DatasetConstructor: ds,
			OutputPath:         outputPath,
			Target:             targetVar,
			MetaStorage:        metaStorage,
			DataStorage:        dataStorage,
			SolutionStorage:    solutionStorage,
			DatasetIngested:    datasetIngested,
			DatasetImported:    datasetImported,
			Config:             config,
		}

		res, err := task.Predict(predictParams)
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

func createImageFromRequest(data []byte, datasetName string, outputPath string, imageType string, config *env.Config) (bool, []byte, error) {
	// raw request is zip file of image dataset that needs to be imported
	datasetPath, err := dataset.StoreZipDataset(datasetName, data)
	if err != nil {
		return false, nil, err
	}

	expandedInfo, err := dataset.ExpandZipDataset(datasetName, datasetPath)
	if err != nil {
		return false, nil, err
	}
	ds, err := createMediaDataset(datasetName, imageType, expandedInfo.ExtractedFilePath)
	if err != nil {
		return false, nil, err
	}
	_, formattedPath, err := task.CreateDataset(datasetName, ds, outputPath, config)
	if err != nil {
		return false, nil, err
	}

	formattedPath = path.Join(formattedPath, "tables", "learningData.csv")

	// once imported, read the csv file as the data to use for the inference
	datasetData, err := ioutil.ReadFile(formattedPath)
	if err != nil {
		return false, nil, err
	}
	return true, datasetData, nil
}

func createTimeseriesFromRequest(dataStorage api.DataStorage, datasetES *api.Dataset, startStr string, stepCount int) (bool, []byte, error) {
	// need to create timeseries based on start time and step count
	var groupingVar *model.Variable
	for _, v := range datasetES.Variables {
		if v.IsGrouping() {
			groupingVar = v
			break
		}
	}

	// find the timsetamp column and id columns
	tsg := groupingVar.Grouping.(*model.TimeseriesGrouping)
	timestampCol := tsg.XCol
	var timestampVar *model.Variable
	for _, v := range datasetES.Variables {
		if v.Name == timestampCol {
			timestampVar = v
			break
		}
	}

	// get the distinct values for the id columns
	idValues := make(map[string][]string)
	for _, vID := range tsg.SubIDs {
		vals, err := dataStorage.FetchRawDistinctValues(datasetES.ID, datasetES.StorageName, []string{vID})
		if err != nil {
			return false, nil, errors.Wrapf(err, "unable to fetch distinct values for '%s' from data storage", vID)
		}
		idValues[vID] = vals[0]
	}

	// get the step duration
	timestampValues, err := dataStorage.FetchRawDistinctValues(datasetES.ID, datasetES.StorageName, []string{timestampCol})
	if err != nil {
		return false, nil, errors.Wrapf(err, "unable to fetch distinct timestamp values from data storage")
	}

	// generate timestamps to use for prediction based on type of timestamp
	var timestampPredictionValues []string
	if model.IsDateTime(timestampVar.Type) {
		timestampPredictionValues, err = generateTimestampValuesInference(timestampValues[0], startStr, stepCount)
	} else if model.IsNumerical(timestampVar.Type) {
		timestampPredictionValues, err = generateIntValuesInference(timestampValues[0], startStr, stepCount)
	} else {
		return false, nil, errors.Errorf("timestamp variable '%s' is type '%s' which is not supported for timeseries creation", timestampVar.Name, timestampVar.Type)
	}
	if err != nil {
		return false, nil, err
	}

	timeseriesData, err := createTimeseriesData(idValues, timestampCol, timestampPredictionValues)
	if err != nil {
		return false, nil, err
	}

	return false, timeseriesData, nil
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

func generateIntValuesInference(existingValues []string, startStr string, stepCount int) ([]string, error) {
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

func generateTimestampValuesInference(existingValues []string, startStr string, stepCount int) ([]string, error) {
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
