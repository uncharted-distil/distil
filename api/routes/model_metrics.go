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
	"github.com/uncharted-distil/distil-compute/primitive/compute"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/util"
	"goji.io/v3/pat"
)

// ModelMetricDesc provides a scoring ID, display name, and description.
type ModelMetricDesc struct {
	ID          util.MetricID `json:"id"`
	DisplayName string        `json:"displayName"`
	Description string        `json:"description"`
}

// ModelMetrics provides a lit of combinations to be serialized to JSON for transport to the
// client.
type ModelMetrics struct {
	Combinations []*ModelMetricDesc `json:"metrics"`
}

// ModelMetricsHandler fetches a list of available model metric methods for an analysis task.
func ModelMetricsHandler(ctor api.MetadataStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		task := pat.Param(r, "task")
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
		metricsList := make([]*ModelMetricDesc, len(taskMetrics))
		for _, value := range taskMetrics {
			metricsList[idx] = &ModelMetricDesc{value.ID, value.DisplayName, value.Description}
			idx++
		}

		metrics := ModelMetrics{metricsList}
		err := handleJSON(w, metrics)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal result histogram into JSON"))
			return
		}
	}
}
