package model

import (
	"time"
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
		if f.FeatureType == FeatureTypeTarget {
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
	SolutionID  string            `json:"solutionId"`
	RequestID   string            `json:"requestId"`
	Progress    string            `json:"progress"`
	CreatedTime time.Time         `json:"timestamp"`
	Results     []*SolutionResult `json:"results"`
	Scores      []*SolutionScore  `json:"scores"`
}

// SolutionResult represents the solution result metadata.
type SolutionResult struct {
	SolutionID  string    `json:"solutionId"`
	Dataset     string    `json:"dataset"`
	ResultURI   string    `json:"requestUri"`
	ResultUUID  string    `json:"resultId"`
	Progress    string    `json:"progress"`
	OutputType  string    `json:"outputType"`
	CreatedTime time.Time `json:"timestamp"`
}

// SolutionScore represents the result score data.
type SolutionScore struct {
	SolutionID string  `json:"solutionId"`
	Metric     string  `json:"metric"`
	Score      float64 `json:"value"`
}
