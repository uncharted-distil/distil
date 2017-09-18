package model

// Request represents the pipeline request metadata.
type Request struct {
	SessionID  string
	RequestID  string
	PipelineID string
	Dataset    string
	Progress   string

	Results  []*Result
	Features []*RequestFeature
}

// Result represents the pipeline result metadata.
type Result struct {
	RequestID  string
	ResultURI  string
	ResultUUID string
	Progress   string
}

// RequestFeature represents the request feature metadata.
type RequestFeature struct {
	RequestID   string
	FeatureName string
	FeatureType string
}
