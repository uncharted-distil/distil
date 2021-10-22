package importer

import (
	"github.com/uncharted-distil/distil-compute/metadata"
	"github.com/uncharted-distil/distil/api/env"
	"github.com/uncharted-distil/distil/api/task"
)

// Datamart can be used to import datasets from datamarts. Note that all the work
// is done by the datamart metadata storage so this is empty.
type Datamart struct {
	source    metadata.DatasetSource
	datasetID string
}

// NewDatamart creates an importer for datamart datasets.
func NewDatamart() Importer {
	return &Datamart{}
}

// Initialize sets up the importer.
func (d *Datamart) Initialize(params map[string]interface{}, ingestParams *task.IngestParams) error {
	d.datasetID = ingestParams.ID
	d.source = ingestParams.Source
	return nil
}

// PrepareImport prepares the datamart dataset for import.
func (d *Datamart) PrepareImport() (*task.IngestSteps, *task.IngestParams, error) {
	return &task.IngestSteps{
		ClassificationOverwrite: false,
		VerifyMetadata:          true,
		FallbackMerged:          true,
		CheckMatch:              true,
		SkipFeaturization:       false,
	}, &task.IngestParams{Path: env.ResolvePath(d.source, d.datasetID), Source: d.source}, nil
}

// Import imports the dataset using the datamart importer..
func (d *Datamart) Import() error {
	return nil
}

// CleanupImport removes temporary files and structures created during the import.
func (d *Datamart) CleanupImport(ingestResult *task.IngestResult) error {
	return nil
}
