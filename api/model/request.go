package model

import (
	"time"
)

// Request represents the pipeline request metadata.
type Request struct {
	SessionID       string            `json:"sessionId"`
	RequestID       string            `json:"requestId"`
	Dataset         string            `json:"dataset"`
	Progress        string            `json:"progress"`
	CreatedTime     time.Time         `json:"createdTime"`
	LastUpdatedTime time.Time         `json:"lastUpdatedTime"`
	Results         []*Result         `json:"results"`
	Features        []*RequestFeature `json:"features"`
	Filters         *FilterParams     `json:"filters"`
}

// Result represents the pipeline result metadata.
type Result struct {
	RequestID   string         `json:"requestId"`
	PipelineID  string         `json:"pipelineId"`
	ResultURI   string         `json:"requestUri"`
	ResultUUID  string         `json:"resultUuid"`
	Progress    string         `json:"progress"`
	OutputType  string         `json:"outputType"`
	CreatedTime time.Time      `json:"createdTime"`
	Scores      []*ResultScore `json:"scores"`
	Filters     *FilterParams  `json:"filters"`
}

// RequestFeature represents a request feature metadata.
type RequestFeature struct {
	RequestID   string `json:"requestId"`
	FeatureName string `json:"featureName"`
	FeatureType string `json:"featureType"`
}

// ResultScore represents the result score data.
type ResultScore struct {
	PipelineID string  `json:"pipelineId"`
	Metric     string  `json:"metric"`
	Score      float64 `json:"value"`
}
