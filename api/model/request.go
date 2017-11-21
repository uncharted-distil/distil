package model

import (
	"time"
)

// Request represents the pipeline request metadata.
type Request struct {
	SessionID       string
	RequestID       string
	Dataset         string
	Progress        string
	CreatedTime     time.Time
	LastUpdatedTime time.Time
	Results         []*Result
	Features        []*RequestFeature
	Filters         *FilterParams
}

// Result represents the pipeline result metadata.
type Result struct {
	RequestID   string
	PipelineID  string
	ResultURI   string
	ResultUUID  string
	Progress    string
	OutputType  string
	CreatedTime time.Time
	Scores      []*ResultScore
}

// RequestFeature represents the request feature metadata.
type RequestFeature struct {
	RequestID   string
	FeatureName string
	FeatureType string
}

// ResultScore represents the result score data.
type ResultScore struct {
	PipelineID string  `json:"pipelineId"`
	Metric     string  `json:"metric"`
	Score      float64 `json:"value"`
}
