package model

// Request represents the pipeline request metadata.
type Request struct {
	SessionID string
	RequestID string
	Dataset   string
	Progress  string

	Results  []*Result
	Features []*RequestFeature
}

// Result represents the pipeline result metadata.
type Result struct {
	RequestID  string
	PipelineID string
	ResultURI  string
	ResultUUID string
	Progress   string

	Scores []*ResultScore
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
