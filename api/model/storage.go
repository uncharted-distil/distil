package model

// StorageCtor represents a client constructor to instantiate a storage
// client.
type StorageCtor func() (Storage, error)

// Storage defines the functions available to query the underlying data storage.
type Storage interface {
	FetchData(dataset string, index string, filterParams *FilterParams) (*FilteredData, error)
	FetchSummary(dataset string, variable *Variable) (*Histogram, error)
	PersistResult(dataset string, resultURI string) error
	FetchResults(dataset string, resultURI string, index string) (*FilteredData, error)
	FetchResultsSummary(dataset string, resultURI string, index string) (*Histogram, error)

	// System data operations
	PersistSession(sessionID string) error
	PersistRequest(sessionID string, requestID string, pipelineID string, dataset string, progress string) error
	UpdateRequest(requestID string, progress string) error
	PersistResultMetadata(requestID string, resultUUID string, resultURI string) error
	FetchRequests(sessionID string) ([]*Request, error)
}
