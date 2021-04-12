//
//   Copyright © 2021 Uncharted Software Inc.
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
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/paulmach/orb"
	geo "github.com/paulmach/orb/geojson"
	"github.com/pkg/errors"
	"goji.io/v3/pat"

	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil-compute/primitive/compute"
	"github.com/uncharted-distil/distil/api/env"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/util"
	log "github.com/unchartedsoftware/plog"
)

// ExportHandler exports the caller supplied solution by calling through to the compute
// server export functionality.
func ExportHandler(client *compute.Client, exportPath string, logger *env.DiscoveryLogger) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// extract route parameters
		solutionID := pat.Param(r, "solution-id")

		err := client.ExportSolution(context.Background(), solutionID)
		if err != nil {
			log.Infof("Failed solution export request for %s", solutionID)
		} else {
			log.Infof("Completed export request for %s", solutionID)
		}

		_, err = logger.InitializeLog("event-" + util.GenerateTimeFileNameStr() + ".csv")
		if err != nil {
			log.Infof("error initializing log after export: %v", err)
		}
	}
}

// ExportResultHandler will return a CSV file containing the results of a prediction.
// Data is transformed into a string using a naive print statement.
func ExportResultHandler(solutionCtor api.SolutionStorageCtor, dataCtor api.DataStorageCtor, metaCtor api.MetadataStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		produceRequestID, err := url.PathUnescape(pat.Param(r, "produce-request-id"))
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to unescape produce request id"))
			return
		}

		format, err := url.PathUnescape(pat.Param(r, "format"))
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to unescape format"))
			return
		}

		solution, err := solutionCtor()
		if err != nil {
			handleError(w, err)
			return
		}

		data, err := dataCtor()
		if err != nil {
			handleError(w, err)
			return
		}

		meta, err := metaCtor()
		if err != nil {
			handleError(w, err)
			return
		}

		// get the solution result (which is actually the prediction result) using the predict request ID
		predictResult, err := solution.FetchPredictionResultByProduceRequestID(produceRequestID)
		if err != nil {
			handleError(w, err)
			return
		}

		// get the original solution ID out of the result
		solutionID := predictResult.SolutionID

		// get the filters
		req, err := solution.FetchRequestBySolutionID(solutionID)
		if err != nil {
			handleError(w, err)
			return
		}
		if req == nil {
			handleError(w, errors.Errorf("solution id `%s` cannot be mapped to result URI", solutionID))
			return
		}

		// if no result, return an empty map
		if predictResult == nil {
			err = handleJSON(w, make(map[string]interface{}))
			if err != nil {
				handleError(w, errors.Wrap(err, "unable marshal version into JSON and write response"))
			}
			return
		}

		// Expand any grouped variables defined in filters into their subcomponents
		dataset := predictResult.Dataset

		ds, err := meta.FetchDataset(dataset, false, false, false)
		if err != nil {
			handleError(w, err)
			return
		}
		storageName := ds.StorageName

		// get row count for export
		rowCount, err := data.FetchNumRows(storageName, ds.Variables)
		if err != nil {
			handleError(w, err)
			return
		}
		filterParamsRaw := &api.FilterParamsRaw{
			Size: rowCount,
		}
		req.Filters.Merge(filterParamsRaw)
		filterParams, err := api.ExpandFilterParams(dataset, req.Filters, false, meta)
		if err != nil {
			handleError(w, err)
			return
		}

		results, err := data.FetchResults(dataset, storageName, predictResult.ResultURI, produceRequestID, filterParams, true)
		if err != nil {
			handleError(w, err)
			return
		}

		// replace any NaN values with an empty string
		results = api.ReplaceNaNs(results, api.EmptyString)

		// write out the result to CSV
		contentType, extension, output, err := createExportedData(req.TargetFeature(), format, results)
		if err != nil {
			handleError(w, err)
			return
		}

		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment;filename=TheCSVFileName.%s", extension))
		_, err = w.Write(output)
		if err != nil {
			handleError(w, err)
			return
		}
	}
}

func createExportedData(target string, format string, results *api.FilteredData) (string, string, []byte, error) {
	switch format {
	case "csv":
		return exportCSV(results)
	case "geojson":
		return exportGeoJSON(target, results)
	default:
		return "", "", nil, errors.Errorf("unsupported export format '%s'", format)
	}
}

func exportCSV(results *api.FilteredData) (string, string, []byte, error) {
	outputBuffer := &bytes.Buffer{}
	wr := csv.NewWriter(outputBuffer)

	header := make([]string, len(results.Columns))
	for i, c := range results.Columns {
		header[i] = c.Label
	}
	err := wr.Write(header)
	if err != nil {
		return "", "", nil, errors.Wrapf(err, "unable to write csv header")
	}

	for _, row := range results.Values {
		record := make([]string, len(row))
		for i, v := range row {
			if v != nil {
				record[i] = fmt.Sprintf("%v", v.Value)
			}
		}
		err = wr.Write(record)
		if err != nil {
			return "", "", nil, errors.Wrapf(err, "unable to write csv record")
		}
	}
	wr.Flush()

	return "text/csv", "csv", outputBuffer.Bytes(), nil
}

func exportGeoJSON(target string, results *api.FilteredData) (string, string, []byte, error) {
	if !canExportGeoJSON(results) {
		return "", "", nil, errors.Errorf("unable to export results to geo json")
	}

	coordinateColumnIndex := -1
	predictionColumnIndex := -1
	for i, c := range results.Columns {
		if model.IsVector(c.Type) {
			coordinateColumnIndex = i
		} else if strings.Contains(c.Key, ":predicted") {
			predictionColumnIndex = i
		}
	}

	// build the geojson content
	output := []*geo.Feature{}
	for _, row := range results.Values {
		geometry := createGeometry(row[coordinateColumnIndex].Value.([]float64))
		feature := geo.NewFeature(geometry)
		feature.Properties[target] = row[predictionColumnIndex].Value.(string)
		output = append(output, feature)
	}

	outputBin, err := json.Marshal(output)
	if err != nil {
		return "", "", nil, errors.Wrapf(err, "unable to marshal geojson output")
	}

	return "application/json", "json", outputBin, nil
}

func getPointsFromVector(polygon []float64) [][]float64 {
	points := [][]float64{}
	for i := 0; i < len(polygon); i += 2 {
		points = append(points, []float64{polygon[i], polygon[i+1]})
	}
	points = append(points, points[0])

	return points
}

func canExportGeoJSON(results *api.FilteredData) bool {
	// would expect a multiband image and a real vector
	types := map[string]bool{}
	for _, c := range results.Columns {
		types[c.Type] = true
	}
	return types[model.MultiBandImageType] && types[model.RealVectorType]
}

func createGeometry(coordinates []float64) orb.Geometry {
	pointsCoordinates := getPointsFromVector(coordinates)
	points := make([]orb.Point, len(pointsCoordinates))
	for i, p := range pointsCoordinates {
		points[i] = [2]float64{p[0], p[1]}
	}
	rings := []orb.Ring{points}

	// make sure we follow the right hand rule (ie counter-clockwise direction)
	if rings[0].Orientation() == orb.CW {
		rings[0].Reverse()
	}

	return orb.Polygon{rings[0]}
}
