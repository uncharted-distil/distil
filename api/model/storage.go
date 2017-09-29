package model

import (
	"time"
)

// StorageCtor represents a client constructor to instantiate a storage
// client.
type StorageCtor func() (Storage, error)

// Storage defines the functions available to query the underlying data storage.
type Storage interface {
	FetchData(dataset string, index string, filterParams *FilterParams, inclusive bool) (*FilteredData, error)
	FetchSummary(dataset string, variable *Variable) (*Histogram, error)
	PersistResult(dataset string, resultURI string) error
	FetchResults(dataset string, resultURI string, index string) (*FilteredData, error)
	FetchResultsSummary(dataset string, resultURI string, index string) (*Histogram, error)

	// System data operations NOTE: Note sure if this should be split off in a different interface.
	PersistSession(sessionID string) error
	PersistRequest(sessionID string, requestID string, dataset string, progress string, createdTime time.Time) error
	PersistResultMetadata(requestID string, pipelineID string, resultUUID string, resultURI string, progress string, outputType string, createdTime time.Time) error
	PersistResultScore(pipelineID string, metric string, score float64) error
	PersistRequestFeature(requestID string, featureName string, featureType string) error
	UpdateRequest(requestID string, progress string, updatedTime time.Time) error
	FetchRequests(sessionID string) ([]*Request, error)
	FetchResultMetadata(requestID string) ([]*Result, error)
	FetchResultScore(pipelineID string) ([]*ResultScore, error)
	FetchRequestFeature(requestID string) ([]*RequestFeature, error)
}
