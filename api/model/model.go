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

package model

import (
	"regexp"
	"strings"
	"time"

	"github.com/uncharted-distil/distil-compute/model"
)

const (
	// SolutionResultTypeInference is the solution result type for inferences.
	SolutionResultTypeInference = "inference"
	// SolutionResultTypeTest is the solution result type for tests.
	SolutionResultTypeTest = "test"
)

var (
	suffixReg = regexp.MustCompile(`:error|:predicted|:confidence$`)
)

// ExportedModel represents a description of an exported model.
type ExportedModel struct {
	ModelName        string   `json:"modelName"`
	ModelDescription string   `json:"modelDescription"`
	FilePath         string   `json:"filePath"`
	FittedSolutionID string   `json:"fittedSolutionId"`
	DatasetID        string   `json:"datasetId"`
	DatasetName      string   `json:"datasetName"`
	Target           string   `json:"target"`
	Variables        []string `json:"variables"`
}

// Request represents the request metadata.
type Request struct {
	RequestID       string        `json:"requestId"`
	Dataset         string        `json:"dataset"`
	Progress        string        `json:"progress"`
	CreatedTime     time.Time     `json:"timestamp"`
	LastUpdatedTime time.Time     `json:"lastUpdatedTime"`
	Features        []*Feature    `json:"features"`
	Filters         *FilterParams `json:"filters"`
}

// Prediction represents the prediction metadata.
type Prediction struct {
	RequestID        string    `json:"requestId"`
	Dataset          string    `json:"dataset"`
	Target           string    `json:"target"`
	FittedSolutionID string    `json:"fittedSolutionId"`
	Progress         string    `json:"progress"`
	CreatedTime      time.Time `json:"timestamp"`
	LastUpdatedTime  time.Time `json:"lastUpdatedTime"`
}

// TargetFeature returns the target feature out of the feature set.
func (r *Request) TargetFeature() string {
	for _, f := range r.Features {
		if f.FeatureType == model.FeatureTypeTarget {
			return f.FeatureName
		}
	}
	return ""
}

// Feature represents a request feature metadata.
type Feature struct {
	RequestID   string `json:"requestId"`
	FeatureName string `json:"featureName"`
	FeatureType string `json:"featureType"`
}

// Solution is a container for a TA2 solution.
type Solution struct {
	SolutionID          string            `json:"solutionId"`
	ExplainedSolutionID string            `json:"explainedSolutionId"`
	RequestID           string            `json:"requestId"`
	CreatedTime         time.Time         `json:"timestamp"`
	State               *SolutionState    `json:"state"`
	Results             []*SolutionResult `json:"results"`
	Scores              []*SolutionScore  `json:"scores"`
	IsBad               bool              `json:"isBad"`
}

// SolutionState represents the state updates for a solution.
type SolutionState struct {
	SolutionID  string    `json:"solutionId"`
	Progress    string    `json:"progress"`
	CreatedTime time.Time `json:"timestamp"`
}

// SolutionExplainResult captures the explainable output by row.
type SolutionExplainResult struct {
	ResultURI     string
	Values        [][]string
	D3MIndexIndex int
	ParsingParams []interface{}
}

// SolutionFeatureWeight captures the weights for a given d3m index and result.
type SolutionFeatureWeight struct {
	ResultURI string
	D3MIndex  int64
	Weights   map[string]float64
}

// SolutionWeight captures the weights for a given d3m index and result.
type SolutionWeight struct {
	SolutionID   string
	FeatureIndex int64
	FeatureName  string
	Weight       float64
}

// SolutionResult represents the solution result metadata.
type SolutionResult struct {
	FittedSolutionID string    `json:"fittedSolutionId"`
	ProduceRequestID string    `json:"produceRequestId"`
	SolutionID       string    `json:"solutionId"`
	Dataset          string    `json:"dataset"`
	ResultType       string    `json:"result_type"`
	ResultURI        string    `json:"requestUri"`
	ResultUUID       string    `json:"resultId"`
	Progress         string    `json:"progress"`
	OutputType       string    `json:"outputType"`
	CreatedTime      time.Time `json:"timestamp"`
}

// SolutionScore represents the result score data.
type SolutionScore struct {
	SolutionID     string  `json:"solutionId"`
	Metric         string  `json:"metric"`
	Label          string  `json:"label"`
	Score          float64 `json:"value"`
	SortMultiplier float64 `json:"sortMultiplier"`
}

// PredictionResult represents the output from a model prediction.
type PredictionResult struct {
	*FilteredData
	FittedSolutionID string `json:"fittedSolutionId"`
	ProduceRequestID string `json:"produceRequestId"`
}

// GetPredictedKey returns a solutions predicted col key.
func GetPredictedKey(solutionID string) string {
	return solutionID + ":predicted"
}

// GetErrorKey returns a solutions error col key.
func GetErrorKey(solutionID string) string {
	return solutionID + ":error"
}

// GetConfidenceKey returns a solutions error col key.
func GetConfidenceKey(solutionID string) string {
	return solutionID + ":confidence"
}

// IsPredictedKey returns true if the key matches a predicted key.
func IsPredictedKey(key string) bool {
	return strings.HasSuffix(key, ":predicted")
}

// IsErrorKey returns true if the key matches an error key.
func IsErrorKey(key string) bool {
	return strings.HasSuffix(key, ":error")
}

// IsConfidenceKey returns true if the key matches an error key.
func IsConfidenceKey(key string) bool {
	return strings.HasSuffix(key, ":confidence")
}

// IsResultKey returns true if the key matches an predicted or error key.
func IsResultKey(key string) bool {
	return IsPredictedKey(key) || IsErrorKey(key)
}

// StripKeySuffix removes any result key suffix.
func StripKeySuffix(key string) string {
	return suffixReg.ReplaceAllString(key, "")
}
