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

import (
	"github.com/uncharted-distil/distil-compute/primitive/compute"
)

const (
	// Accuracy identifies model metric based on nearness to the original result.
	Accuracy = "accuracy"
	// F1 identifies model metric based on precision and recall
	F1 = "f1"
	// F1Micro identifies model metric based  on precision and recall
	F1Micro = "f1Micro"
	// F1Macro identifies model metric based on precision and recall.
	F1Macro = "f1Macro"
	// MeanAbsoluteError identifies model metric based on
	MeanAbsoluteError = "meanAbsoluteError"
	// MeanSquaredError identifies model metric based on the quality of the estimator.
	MeanSquaredError = "meanSquaredError"
	// NormalizedMutualInformation identifies model metric based on the relationship between variables.
	NormalizedMutualInformation = "normalizedMutualInformation"
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
	allModelLabels = map[string]string{
		Accuracy:                    compute.GetMetricLabel(compute.ConvertProblemMetricToTA2(Accuracy)),
		F1:                          compute.GetMetricLabel(compute.ConvertProblemMetricToTA2(F1)),
		F1Macro:                     compute.GetMetricLabel(compute.ConvertProblemMetricToTA2(F1Macro)),
		F1Micro:                     compute.GetMetricLabel(compute.ConvertProblemMetricToTA2(F1Micro)),
		RocAuc:                      compute.GetMetricLabel(compute.ConvertProblemMetricToTA2(RocAuc)),
		RocAucMacro:                 compute.GetMetricLabel(compute.ConvertProblemMetricToTA2(RocAucMacro)),
		RocAucMicro:                 compute.GetMetricLabel(compute.ConvertProblemMetricToTA2(RocAucMicro)),
		MeanAbsoluteError:           compute.GetMetricLabel(compute.ConvertProblemMetricToTA2(MeanAbsoluteError)),
		MeanSquaredError:            compute.GetMetricLabel(compute.ConvertProblemMetricToTA2(MeanSquaredError)),
		NormalizedMutualInformation: compute.GetMetricLabel(compute.ConvertProblemMetricToTA2(NormalizedMutualInformation)),
		RootMeanSquaredError:        compute.GetMetricLabel(compute.ConvertProblemMetricToTA2(RootMeanSquaredError)),
		RootMeanSquaredErrorAvg:     compute.GetMetricLabel(compute.ConvertProblemMetricToTA2(RootMeanSquaredErrorAvg)),
		RSquared:                    compute.GetMetricLabel(compute.ConvertProblemMetricToTA2(RSquared)),
	}

	// AllModelMetrics defines a list of model scoring metrics
	AllModelMetrics = map[string]Metric{
		Accuracy: {
			Accuracy,
			allModelLabels[Accuracy],
			allModelLabels[Accuracy] + " scores the result based only on the percentage of correct predictions.",
		},
		F1: {
			F1,
			allModelLabels[F1],
			allModelLabels[F1] + " scoring averages true positives, false negatives and false positives for binary classifications, balancing precision and recall.",
		},
		F1Macro: {
			F1Macro,
			allModelLabels[F1Macro],
			"F1 Macro scoring averages true positives, false negatives and false positives for all multi-class classification options, balancing precision and recall.",
		},
		F1Micro: {
			F1Micro,
			allModelLabels[F1Micro],
			allModelLabels[F1Micro] + " scoring averages true positives, false negatives and false positives for each multi-class classification options for multi-class problems, balancing precision and recall.",
		},
		RocAuc: {
			RocAuc,
			allModelLabels[RocAuc],
			allModelLabels[RocAuc] + " scoring compares relationship between inputs on result for binary classifications.",
		},
		RocAucMacro: {
			RocAucMacro,
			allModelLabels[RocAucMacro],
			allModelLabels[RocAucMacro] + " scoring compares the relationship between inputs and the result for all multi-class classification options.",
		},
		RocAucMicro: {
			RocAucMicro,
			allModelLabels[RocAucMicro],
			allModelLabels[RocAucMicro] + " scoring compares hte relationship between inputs and the result for each multi-class classification options.",
		},
		MeanAbsoluteError: {
			MeanAbsoluteError,
			allModelLabels[MeanAbsoluteError],
			allModelLabels[MeanAbsoluteError] + " measures the average magnitude of errors in a set of predictions.",
		},
		MeanSquaredError: {
			MeanSquaredError,
			allModelLabels[MeanSquaredError],
			allModelLabels[MeanSquaredError] + " measures the quality of an estimator where values closer to 0 are better.",
		},
		NormalizedMutualInformation: {
			NormalizedMutualInformation,
			allModelLabels[NormalizedMutualInformation],
			allModelLabels[NormalizedMutualInformation] + " scores the relationship / lack of entropy between variables where 0 is no relationship and 1 is a strong relationship.",
		},
		RootMeanSquaredError: {
			RootMeanSquaredError,
			allModelLabels[RootMeanSquaredError],
			allModelLabels[RootMeanSquaredError] + " measures the quality of an estimator and the average magnitude of the error.",
		},
		RootMeanSquaredErrorAvg: {
			RootMeanSquaredErrorAvg,
			allModelLabels[RootMeanSquaredErrorAvg],
			allModelLabels[RootMeanSquaredErrorAvg] + " measures the quality of an estimator and the average magnitude of the error averaged across classifcations options.",
		},
		RSquared: {
			RSquared,
			allModelLabels[RSquared],
			allModelLabels[RSquared] + " measures the relationship between predictions and their inputs where values closer to 1 suggest a strong correlation.",
		},
	}

	//TaskMetricMap maps tasks to metrics
	TaskMetricMap = map[string]map[string]Metric{
		compute.BinaryTask: {
			Accuracy: AllModelMetrics[Accuracy],
			F1:       AllModelMetrics[F1],
			RocAuc:   AllModelMetrics[RocAuc],
		},
		compute.MultiClassTask: {
			Accuracy:    AllModelMetrics[Accuracy],
			F1Macro:     AllModelMetrics[F1Micro],
			F1Micro:     AllModelMetrics[F1Macro],
			RocAucMacro: AllModelMetrics[RocAucMacro],
			RocAucMicro: AllModelMetrics[RocAucMicro],
		},
		compute.ClassificationTask: {
			Accuracy:    AllModelMetrics[Accuracy],
			F1:          AllModelMetrics[F1],
			F1Macro:     AllModelMetrics[F1Micro],
			F1Micro:     AllModelMetrics[F1Macro],
			RocAuc:      AllModelMetrics[RocAuc],
			RocAucMacro: AllModelMetrics[RocAucMacro],
			RocAucMicro: AllModelMetrics[RocAucMicro],
		},
		compute.RegressionTask: {
			MeanAbsoluteError:       AllModelMetrics[MeanAbsoluteError],
			MeanSquaredError:        AllModelMetrics[MeanSquaredError],
			RootMeanSquaredError:    AllModelMetrics[RootMeanSquaredError],
			RootMeanSquaredErrorAvg: AllModelMetrics[RootMeanSquaredErrorAvg],
			RSquared:                AllModelMetrics[RSquared],
		},
	}
)
