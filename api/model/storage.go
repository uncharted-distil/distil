package model

// StorageCtor represents a client constructor to instantiate a storage
// client.
type StorageCtor func() (Storage, error)

// Storage defines the functions available to query the underlying data storage.
type Storage interface {
	FetchData(string, *FilterParams) (*FilteredData, error)
	FetchSummary(dataset string, variable *Variable) (*Histogram, error)
	PersistResult(dataset string, pipelineID string, resultURI string) error
	FetchResults(pipelineURI string, resultURI string, index string, dataset string, targetName string) (*FilteredData, error)
	FetchResultsSummary(pipelineURI string, resultURI string, index string, dataset string, targetName string) (*Histogram, error)
}
