//
//    Copyright © 2021 Uncharted Software Inc.
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package compute

import (
	"errors"

	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil-compute/primitive/compute"
	api "github.com/uncharted-distil/distil/api/model"
)

// Task provides an array of task keywords.  These are mapped to string definitions
// derfined by the LL d3m problem schema.
type Task struct {
	Task []string `json:"task"`
}

// HasTaskType indicates whether or not a given Task includes the supplied task type.
func HasTaskType(task *Task, taskType string) bool {
	for _, typ := range task.Task {
		if typ == taskType {
			return true
		}
	}
	return false
}

const semiSupervisedThreshold = 0.1

// ResolveTask will determine the task and subtask given a dataset, training and target variables.
func ResolveTask(storage api.DataStorage, datasetStorageName string, targetVariable *model.Variable, features []*model.Variable) (*Task, error) {
	// Given the target variable and dataset, compute the task and subtask.
	// If there's no target variable, we'll treat this as an unsupervised clustering task.
	if targetVariable == nil {
		return &Task{[]string{compute.ClusteringTask}}, nil
	}

	// If this is a timeseries target type the task will be a forecasting
	if model.IsTimeSeries(targetVariable.Type) {
		return &Task{[]string{compute.ForecastingTask, compute.TimeSeriesTask}}, nil
	}

	// If this is numerical target type we'll treat this a regression task.  Determine whether its an image
	// regression or numerical regression.
	if model.IsNumerical(targetVariable.Type) {
		tasks := []string{compute.RegressionTask}
		for _, feature := range features {
			if model.IsImage(feature.Type) {
				tasks = append(tasks, compute.ImageTask)
				return &Task{tasks}, nil
			}
			if model.IsMultiBandImage(feature.Type) {
				tasks = append(tasks, compute.RemoteSensingTask)
				return &Task{tasks}, nil
			}
			if model.IsTimeSeries(feature.Type) {
				tasks = append(tasks, compute.TimeSeriesTask)
				return &Task{tasks}, nil
			}
		}

		// Numerical regression.  Currently no support for multivariate, so we default
		// to univariate.
		return &Task{[]string{compute.RegressionTask}}, nil
	}

	// Classification.  This can be binary, multiclass, or semi-supervised depending on what we have for
	// label distribution, and may involve image data.
	if model.IsCategorical(targetVariable.Type) {
		task := []string{compute.ClassificationTask}

		// Flag the presence of image features.
		for _, feature := range features {
			if model.IsImage(feature.Type) {
				task = append(task, compute.ImageTask)
			} else if model.IsMultiBandImage(feature.Type) {
				task = append(task, compute.RemoteSensingTask)
			} else if model.IsTimeSeries(feature.Type) {
				task = append(task, compute.TimeSeriesTask)
			}
		}

		// Fetch the counts for each category
		targetCounts, err := storage.FetchCategoryCounts(datasetStorageName, targetVariable)
		if err != nil {
			return nil, err
		}

		// If more than 10% of the labels are empty, treat this as a semi-supervised learning task
		total := 0
		for _, count := range targetCounts {
			total += count
		}
		if emptyCount, ok := targetCounts[""]; ok {
			if float32(emptyCount)/float32(total) > semiSupervisedThreshold {
				task = append(task, compute.SemiSupervisedTask)
			}
			// If there are 3 labels (2 + empty), update this as a binary classification task
			excludeEmptySize := len(targetCounts) - 1
			if excludeEmptySize == 2 {
				task = append(task, compute.BinaryTask)
			} else {
				task = append(task, compute.MultiClassTask)
			}
			return &Task{task}, nil
		}

		// If there are only two labels, update this as a binary classification task
		if len(targetCounts) == 2 {
			task = append(task, compute.BinaryTask)
		} else {
			task = append(task, compute.MultiClassTask)
		}
		return &Task{task}, nil
	}

	// if vector type, assume object detection
	if model.IsList(targetVariable.Type) {
		return &Task{[]string{compute.ObjectDetectionTask, compute.ImageTask}}, nil
	}

	return nil, errors.New("failed to determine task from dataset and target")
}
