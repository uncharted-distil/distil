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
	"context"
	"encoding/csv"
	"fmt"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
	"goji.io/v3/pat"

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

		// Expand any grouped variables defined in filters into their subcomponents
		dataset := predictResult.Dataset

		// if no result, return an empty map
		if predictResult == nil {
			err = handleJSON(w, make(map[string]interface{}))
			if err != nil {
				handleError(w, errors.Wrap(err, "unable marshal version into JSON and write response"))
			}
			return
		}

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
		filterParams := &api.FilterParams{
			Size: rowCount,
		}
		filterParams.Merge(req.Filters)
		filterParams, err = api.ExpandFilterParams(dataset, filterParams, false, meta)
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
		wr := csv.NewWriter(w)
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", "attachment;filename=TheCSVFileName.csv")

		header := make([]string, len(results.Columns))
		for i, c := range results.Columns {
			header[i] = c.Label
		}
		err = wr.Write(header)
		if err != nil {
			handleError(w, err)
			return
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
				handleError(w, err)
				return
			}
		}
		wr.Flush()
	}
}
