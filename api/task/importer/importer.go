package importer

import "github.com/uncharted-distil/distil/api/task"

// Importer can be used to import datasets into the system.
type Importer interface {
	Initialize(params map[string]interface{}, ingestParams *task.IngestParams) error
	PrepareImport() (*task.IngestSteps, *task.IngestParams, error)
	Import() error
	CleanupImport(ingestResult *task.IngestResult) error
}
