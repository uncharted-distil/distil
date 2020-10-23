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

package util

const (
	// Accuracy identifies model metric based on nearness to the original result.
	Accuracy = "accuracy"
	// F1 identifies model metric based on precision and recall
	F1 = "f1"
	// F1Micro identifies model metric based  on precision and recall
	F1Micro = "f1Micro"
	// F1Macro identifies model metric based on precision and recall.
	F1Macro = "f1Macro"
	// JaccardSimilarityScore identifies model metric based on
	JaccardSimilarityScore = "jaccardSimilarityScore"
	// MeanAbsoluteError identifies model metric based on
	MeanAbsoluteError = "meanAbsoluteError"
	// MeanSquaredError identifies model metric based on the quality of the estimator.
	MeanSquaredError = "meanSquaredError"
	// NormalizedMutualInformation identifies model metric based on the relationship between variables.
	NormalizedMutualInformation = "normalizedMutualInformation"
	// ObjectDetectionAP identifies model metric based on
	ObjectDetectionAP = "objectDetectionAP"
	// Precision identifies model metric based on nearness to expected result
	Precision = "precision"
	// PrecisionAtTopK identifies model metric based on
	PrecisionAtTopK = "precisionAtTopK"
	// Recall identifies model metric based on
	Recall = "recall"
	// RocAuc identifies model metric based on
	RocAuc = "rocAuc"
	// RocAucMicro identifies model metric based on
	RocAucMicro = "rocAucMicro"
	// RocAucMacro identifies model metric based on
	RocAucMacro = "rocAucMacro"
	// RootMeanSquaredError identifies model metric based on
	RootMeanSquaredError = "rootMeanSquaredError"
	// RootMeanSquaredErrorAvg identifies model metric based on
	RootMeanSquaredErrorAvg = "rootMeanSquaredErrorAvg"
	// RSquared identifies model metric based on
	RSquared = "rSquared"

	// ForecastingTask represents timeseries forcasting
	ForecastingTask = "forecasting"
	// ClassificationTask represents a classification task on image, timeseries or basic tabular data
	ClassificationTask = "classification"
	// RegressionTask represents a regression task on image, timeseries or basic tabular data
	RegressionTask = "regression"
	// ClusteringTask represents an unsupervised clustering task on image, timeseries or basic tabular data
	ClusteringTask = "clustering"
	// LinkPredictionTask represents a link prediction task on graph data
	LinkPredictionTask = "linkPrediction"
	// VertexClassificationTask represents a vertex nomination task on graph data
	VertexClassificationTask = "vertexClassification"
	// VertexNominationTask represents a vertex nomination task on graph data
	VertexNominationTask = "vertexNomination"
	// CommunityDetectionTask represents an unsupervised community detectiontask on on graph data
	CommunityDetectionTask = "communityDetection"
	// GraphMatchingTask represents an unsupervised matching task on graph data
	GraphMatchingTask = "graphMatching"
	// CollaborativeFilteringTask represents a collaborative filtering recommendation task on basic tabular data
	CollaborativeFilteringTask = "collaborativeFiltering"
	// ObjectDetectionTask represents an object detection task on image data
	ObjectDetectionTask = "objectDetection"
	// SemiSupervisedTask represents a semi-supervised classification task on tabular data
	SemiSupervisedTask = "semiSupervised"
	// BinaryTask represents task involving a single binary value for each prediction
	BinaryTask = "binary"
	// MultiClassTask represents a task involving a multi class value for each prediction
	MultiClassTask = "multiClass"
	// MultiLabelTask represents a task involving multiple lables for each each prediction
	MultiLabelTask = "multiLabel"
	// UnivariateTask represents a task involving predictions on a single variable
	UnivariateTask = "univariate"
	// MultivariateTask represents a task involving predictions on multiple variables
	MultivariateTask = "multivariate"
	// OverlappingTask represents a task involving overlapping predictions
	OverlappingTask = "overlapping"
	// NonOverlappingTask represents a task involving non-overlapping predictions
	NonOverlappingTask = "nonOverlapping"
	// TabularTask represents a task involving tabular data
	TabularTask = "tabular"
	// RelationalTask represents a task involving relational data
	RelationalTask = "relational"
	// ImageTask represents a task involving image data
	ImageTask = "image"
	// AudioTask represents a task involving audio data
	AudioTask = "audio"
	// VideoTask represents a task involving video data
	VideoTask = "video"
	// SpeechTask represents a task involving speech data
	SpeechTask = "speech"
	// TextTask represents a task involving text data
	TextTask = "text"
	// GraphTask represents a task involving graph data
	GraphTask = "graph"
	// MultiGraphTask represents a task involving multiple graph data
	MultiGraphTask = "multigraph"
	// TimeSeriesTask represents a task involving timeseries data
	TimeSeriesTask = "timeseries"
	// GroupedTask represents a task involving grouped data
	GroupedTask = "grouped"
	// GeospatialTask represents a task involving geospatial data
	GeospatialTask = "geospatial"
	// RemoteSensingTask represents a task involving remote sensing data
	RemoteSensingTask = "remoteSensing"
	// LupiTask represents a task involving LUPI (Learning Using Priveleged Information) data
	LupiTask = "lupi"
	// UndefinedTask is a flag for undefined/unknown task values
	UndefinedTask = "undefined"

	// UndefinedMetric is a flag for undefined/uknown metric values
	UndefinedMetric = "undefined"
)

// MetricID uniquely identifies model metric methods
type MetricID string

// Metric defines the ID, display name and description for various model metrics
type Metric struct {
	ID          MetricID
	DisplayName string
	Description string
}

var (
	// AllModelMetrics defines a list of model scoring metrics
	AllModelMetrics = map[string]Metric{
		Accuracy:                    {Accuracy, "Accuracy", "Accuracy scores the result based only on the percentage of correct predictions."},
		F1:                          {F1, "F1", "F1 scoring averages true positives, false negatives and false positives for binary classifications, balancing precision and recall."},
		F1Macro:                     {F1Macro, "F1 Macro", "F1 Macro scoring averages true positives, false negatives and false positives for all multi-class classification options, balancing precision and recall."},
		F1Micro:                     {F1Micro, "F1 Micro", "F1 Micro scoring averages true positives, false negatives and false positives for each multi-class classification options for multi-class problems, balancing precision and recall."},
		RocAuc:                      {RocAuc, "RocAuc", "RocAuc scoring compares relationship between inputs on result for binary classifications."},
		RocAucMacro:                 {RocAucMacro, "RocAuc Macro", "RocAuc scoring compares the relationship between inputs and the result for all multi-class classification options."},
		RocAucMicro:                 {RocAucMicro, "RocAuc Micro", "RocAuc scoring compares hte relationship between inputs and the result for each multi-class classification options."},
		MeanAbsoluteError:           {MeanAbsoluteError, "MAE", "The mean absolute error (MAE) measures the average magnitude of errors in a set of predictions."},
		MeanSquaredError:            {MeanSquaredError, "MSE", "The mean squared error measures the quality of an estimator where values closer to 0 are better."},
		NormalizedMutualInformation: {NormalizedMutualInformation, "NMI", "Normalized Mutual Information scores the relationship / lack of entropy between variables where 0 is no relationship and 1 is a strong relationship."},
		RootMeanSquaredError:        {RootMeanSquaredError, "RMSE", "The root mean squared error measures the quality of an estimator and the average magnitude of the error."},
		RootMeanSquaredErrorAvg:     {RootMeanSquaredErrorAvg, "RMSEA", "The root mean squared error average measures the quality of an estimator and the average magnitude of the error averaged across classifcations options."},
		RSquared:                    {RSquared, "RQ", "The root squared measures the relationship between predictions and their inputs where values closer to 1 suggest a strong correlation."},
	}

	//TaskMetricMap maps tasks to metrics
	TaskMetricMap = map[string]map[string]Metric{
		ClassificationTask: {
			Accuracy:    AllModelMetrics[Accuracy],
			F1:          AllModelMetrics[F1],
			F1Macro:     AllModelMetrics[F1Micro],
			F1Micro:     AllModelMetrics[F1Macro],
			RocAuc:      AllModelMetrics[RocAuc],
			RocAucMacro: AllModelMetrics[RocAucMacro],
			RocAucMicro: AllModelMetrics[RocAucMicro],
		},
		RegressionTask: {
			MeanAbsoluteError:       AllModelMetrics[MeanAbsoluteError],
			MeanSquaredError:        AllModelMetrics[MeanSquaredError],
			RootMeanSquaredError:    AllModelMetrics[RootMeanSquaredError],
			RootMeanSquaredErrorAvg: AllModelMetrics[RootMeanSquaredErrorAvg],
			RSquared:                AllModelMetrics[RSquared],
		},
	}
)
