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

const semiSupervisedThreshold = 0.1

// ResolveTask will determine the task and subtask given training and target variables, and the ability of the underlying target labels.
func ResolveTask(storage api.DataStorage, datasetStorageName string, targetVariable *model.Variable) (*Task, error) {
	// Given the target variable and dataset, compute the task and subtask.
	// If there's no target variable, we'll treat this as an unsupervised clustering task.
	if targetVariable == nil {
		return &Task{[]string{compute.ClusteringTask}}, nil
	}

	// If this is a timeseries target type the task will be a forecasting
	if model.IsTimeSeries(targetVariable.Type) {
		return &Task{[]string{compute.ForecastingTask, compute.TimeSeriesTask}}, nil
	}

	// If this is numerical target type we'll treat this a regression task.
	if model.IsNumerical(targetVariable.Type) {
		// Numerical regression.  Currently no support for multivariate, so we default
		// to univariate.
		return &Task{[]string{compute.RegressionTask, compute.UnivariateTask}}, nil
	}

	// Classification.  This can be binary, multiclass, or semi-supervised depending on what we have for
	// label distribution.
	if model.IsCategorical(targetVariable.Type) {
		task := []string{compute.ClassificationTask}

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
			if len(targetCounts) == 2 {
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
