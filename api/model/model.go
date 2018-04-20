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
	Pipelines       []*Pipeline   `json:"pipelines"`
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

// Pipeline is a container for a TA2 pipeline.
type Pipeline struct {
	PipelineID  string            `json:"pipelineId"`
	RequestID   string            `json:"requestId"`
	Progress    string            `json:"progress"`
	CreatedTime time.Time         `json:"timestamp"`
	Results     []*PipelineResult `json:"results"`
	Scores      []*PipelineScore  `json:"scores"`
}

// PipelineResult represents the pipeline result metadata.
type PipelineResult struct {
	PipelineID  string    `json:"pipelineId"`
	Dataset     string    `json:"dataset"`
	ResultURI   string    `json:"requestUri"`
	ResultUUID  string    `json:"resultId"`
	Progress    string    `json:"progress"`
	OutputType  string    `json:"outputType"`
	CreatedTime time.Time `json:"timestamp"`
}

// PipelineScore represents the result score data.
type PipelineScore struct {
	PipelineID string  `json:"pipelineId"`
	Metric     string  `json:"metric"`
	Score      float64 `json:"value"`
}
