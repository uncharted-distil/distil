package model

// Request represents the pipeline request metadata.
type Request struct {
	SessionID  string
	RequestID  string
	PipelineID string
	Dataset    string
	Progress   string
}
