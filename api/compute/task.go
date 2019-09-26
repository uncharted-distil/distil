package compute

import (
	"errors"

	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil-compute/primitive/compute"
	api "github.com/uncharted-distil/distil/api/model"
)

// Task provides a task and subtask field.  These are mapped to string definitions
// derfined by the LL d3m problem schema.
type Task struct {
	Task    string `json:"task"`
	SubTask string `json:"subtask"`
}

const semiSupervisedThreshold = 0.1

// ResolveTask will determine the task and subtask given training and target variables, and the ability of the underlying target labels.
func ResolveTask(storage api.DataStorage, datasetStorageName string, targetVariable *model.Variable) (*Task, error) {
	// Given the target variable and dataset, compute the task and subtask.
	// If there's no target variable, we'll treat this as an unsupervised clustering task.
	if targetVariable == nil {
		return &Task{compute.ClusteringTask, compute.NoneSubTask}, nil
	}

	// If this is a timeseries target type the task will be a forecasting
	if model.IsTimeSeries(targetVariable.Type) {
		return &Task{compute.TimeseriesForecastingTask, compute.NoneSubTask}, nil
	}

	// If this is numerical target type we'll treat this a regression task.
	if model.IsNumerical(targetVariable.Type) {
		// Numerical regression.  Currently no support for multivariate, so we default
		// to univariate.
		return &Task{compute.RegressionTask, compute.UnivariateSubTask}, nil
	}

	// Classification.  This can be binary, multiclass, or semi-supervised depending on what we have for
	// label distribution.
	if model.IsCategorical(targetVariable.Type) {
		task := compute.ClassificationTask
		subTask := compute.MultiClassSubTask

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
				task = compute.SemiSupervisedClassificationTask
			}
			// If there are 3 labels (2 + empty), update this as a binary classification task
			if len(targetCounts) == 2 {
				subTask = compute.BinarySubTask
			}
			return &Task{task, subTask}, nil
		}

		// If there are only two labels, update this as a binary classification task
		if len(targetCounts) == 2 {
			subTask = compute.BinarySubTask
		}
		return &Task{task, subTask}, nil
	}
	return nil, errors.New("failed to determine task from dataset and target")
}
