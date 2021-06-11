//
//   Copyright © 2020 Uncharted Software Inc.
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
	"net/http"
	"strings"

	"github.com/pkg/errors"
	"goji.io/v3/pat"

	"github.com/uncharted-distil/distil-compute/primitive/compute"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/util"
)

// MultiBandCombinationDesc provides a band combination ID and display name.
type MultiBandCombinationDesc struct {
	ID          util.BandCombinationID `json:"id"`
	DisplayName string                 `json:"displayName"`
}

// ModelMetricDesc provides a scoring ID, display name, and description.
type ModelMetricDesc struct {
	ID          util.MetricID `json:"id"`
	DisplayName string        `json:"displayName"`
	Description string        `json:"description"`
}

// Combinations provides a lit of combinations to be serialized to JSON for transport to the
// client.
type Combinations struct {
	Combinations []interface{} `json:"combinations"`
}

// IndexDataHandler fetches a list of index data using the supplied parameters
// to determine the type of index data to return.
func IndexDataHandler(ctor api.MetadataStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		typ := pat.Param(r, "type")

		// parse POST params
		params, err := getPostParameters(r)
		if err != nil {
			handleError(w, errors.Wrap(err, "Unable to parse post parameters"))
			return
		}

		var combinations *Combinations
		switch typ {
		case "metrics":
			task := params["task"].(string)
			combinations = getModelMetrics(task)
		case "bands":
			combinations = getBandCombinations()
		default:
			err = errors.Errorf("unrecognized index data type '%s'", typ)
		}
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal result histogram into JSON"))
			return
		}

		err = handleJSON(w, combinations)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal result histogram into JSON"))
			return
		}
	}
}

func getBandCombinations() *Combinations {
	combinationsList := make([]interface{}, len(util.SentinelBandCombinations))
	idx := 0
	for _, value := range util.SentinelBandCombinations {
		combinationsList[idx] = &MultiBandCombinationDesc{value.ID, value.DisplayName}
		idx++
	}
	return &Combinations{combinationsList}
}

func getModelMetrics(task string) *Combinations {
	var taskMetrics map[string]util.Metric

	if strings.Contains(task, compute.ClassificationTask) {
		if strings.Contains(task, compute.MultiClassTask) {
			taskMetrics = util.TaskMetricMap[compute.MultiClassTask]
		} else if strings.Contains(task, compute.BinaryTask) {
			taskMetrics = util.TaskMetricMap[compute.BinaryTask]
		} else {
			taskMetrics = util.TaskMetricMap[compute.ClassificationTask]
		}
	} else if strings.Contains(task, compute.RegressionTask) || strings.Contains(task, compute.ForecastingTask) {
		taskMetrics = util.TaskMetricMap[compute.RegressionTask]
	} else {
		taskMetrics = util.AllModelMetrics
	}

	idx := 0
	metricsList := make([]interface{}, len(taskMetrics))
	for _, value := range taskMetrics {
		metricsList[idx] = &ModelMetricDesc{value.ID, value.DisplayName, value.Description}
		idx++
	}

	return &Combinations{metricsList}
}
