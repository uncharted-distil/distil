package model

import (
	"time"
)

// DataStorageCtor represents a client constructor to instantiate a data
// storage client.
type DataStorageCtor func() (DataStorage, error)

// DataStorage defines the functions available to query the underlying data storage.
type DataStorage interface {
	FetchNumRows(dataset string, filters map[string]interface{}) (int, error)
	FetchData(dataset string, index string, filterParams *FilterParams, inclusive bool) (*FilteredData, error)
	FetchSummary(dataset string, index string, varName string) (*Histogram, error)
	FetchSummaryByResult(dataset string, index string, varName string, resultURI string, extrema *Extrema) (*Histogram, error)
	PersistResult(dataset string, resultURI string) error
	FetchResults(dataset string, index string, resultURI string) (*FilteredData, error)
	FetchFilteredResults(dataset string, index string, resultURI string, filterParams *FilterParams, inclusive bool) (*FilteredData, error)
	FetchResultsSummary(dataset string, resultURI string, index string, extrema *Extrema) (*Histogram, error)
	FetchResultsExtremaByURI(dataset string, resultURI string, index string) (*Extrema, error)
	FetchResidualsSummary(dataset string, resultURI string, index string, extrema *Extrema) (*Histogram, error)
	FetchResidualsExtremaByURI(dataset string, resultURI string, index string) (*Extrema, error)
	// Dataset manipulation
	SetDataType(dataset string, index string, field string, fieldType string) error
}

// PipelineStorageCtor represents a client constructor to instantiate a
// pipeline storage client.
type PipelineStorageCtor func() (PipelineStorage, error)

// PipelineStorage defines the functions available to query the underlying
// pipeline storage.
type PipelineStorage interface {
	PersistSession(sessionID string) error
	PersistRequest(sessionID string, requestID string, dataset string, progress string, createdTime time.Time) error
	PersistResultMetadata(requestID string, pipelineID string, resultUUID string, resultURI string, progress string, outputType string, createdTime time.Time) error
	PersistResultScore(pipelineID string, metric string, score float64) error
	PersistRequestFeature(requestID string, featureName string, featureType string) error
	PersistRequestFilters(requestID string, filters *FilterParams) error
	UpdateRequest(requestID string, progress string, updatedTime time.Time) error
	FetchRequest(requestID string) (*Request, error)
	FetchRequests(sessionID string) ([]*Request, error)
	FetchResultMetadata(requestID string) ([]*Result, error)
	FetchResultMetadataByUUID(resultUUID string) (*Result, error)
	FetchResultMetadataByPipelineID(pipelineID string) (*Result, error)
	FetchResultMetadataByDatasetTarget(sessionID string, dataset string, target string) ([]*Result, error)
	FetchResultScore(pipelineID string) ([]*ResultScore, error)
	FetchRequestFeatures(requestID string) ([]*RequestFeature, error)
	FetchRequestFilters(requestID string) (*FilterParams, error)
}

// MetadataStorageCtor represents a client constructor to instantiate a
// metadata storage client.
type MetadataStorageCtor func() (MetadataStorage, error)

// MetadataStorage defines the functions available to query the underlying
// metadata storage.
type MetadataStorage interface {
	FetchVariables(dataset string, index string, includeIndex bool) ([]*Variable, error)
	FetchVariablesDisplay(dataset string, index string) ([]*Variable, error)
	FetchVariable(dataset string, index string, varName string) (*Variable, error)
	FetchVariableDisplay(dataset string, index string, varName string) (*Variable, error)
	FetchDatasets(index string, includeIndex bool) ([]*Dataset, error)
	SearchDatasets(index string, terms string, includeIndex bool) ([]*Dataset, error)

	// Dataset manipulation
	SetDataType(dataset string, index string, field string, fieldType string) error
}
