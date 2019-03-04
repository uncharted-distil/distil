package model

import (
	"regexp"
	"strings"
	"time"

	"github.com/uncharted-distil/distil-compute/model"
)

var (
	suffixReg = regexp.MustCompile(`:\S+:error|:\S+:predicted$`)
)

// Request represents the request metadata.
type Request struct {
	RequestID       string        `json:"requestId"`
	Dataset         string        `json:"dataset"`
	Progress        string        `json:"progress"`
	CreatedTime     time.Time     `json:"timestamp"`
	LastUpdatedTime time.Time     `json:"lastUpdatedTime"`
	Features        []*Feature    `json:"features"`
	Filters         *FilterParams `json:"filters"`
	Solutions       []*Solution   `json:"solutions"`
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
	SolutionID  string           `json:"solutionId"`
	RequestID   string           `json:"requestId"`
	Progress    string           `json:"progress"`
	CreatedTime time.Time        `json:"timestamp"`
	Result      *SolutionResult  `json:"result"`
	Scores      []*SolutionScore `json:"scores"`
	IsBad       bool             `json:"isBad"`
}

// SolutionResult represents the solution result metadata.
type SolutionResult struct {
	FittedSolutionID string    `json:"fittedSolutionId"`
	SolutionID       string    `json:"solutionId"`
	Dataset          string    `json:"dataset"`
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

// GetPredictedKey returns a solutions predicted col key.
func GetPredictedKey(target string, solutionID string) string {
	return target + ":" + solutionID + ":predicted"
}

// GetErrorKey returns a solutions error col key.
func GetErrorKey(target string, solutionID string) string {
	return target + ":" + solutionID + ":error"
}

// IsPredictedKey returns true if the key matches a predicted key.
func IsPredictedKey(key string) bool {
	return strings.HasSuffix(key, ":predicted")
}

// IsErrorKey returns true if the key matches an error key.
func IsErrorKey(key string) bool {
	return strings.HasSuffix(key, ":error")
}

// IsResultKey returns true if the key matches an predicted or error key.
func IsResultKey(key string) bool {
	return IsPredictedKey(key) || IsErrorKey(key)
}

// StripKeySuffix removes any result key suffix.
func StripKeySuffix(key string) string {
	return suffixReg.ReplaceAllString(key, "")
}
