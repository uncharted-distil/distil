package model

// StorageCtor represents a client constructor to instantiate a storage
// client.
type StorageCtor func() (Storage, error)

// Storage defines the functions available to query the underlying data storage.
type Storage interface {
	FetchData(string, *FilterParams) (*FilteredData, error)
	FetchSummary(dataset string, variable *Variable) (*Histogram, error)
	PersistResult(dataset string, resultURI string) error
	FetchResults(dataset string, resultURI string, index string) (*FilteredData, error)
	FetchResultsSummary(dataset string, resultURI string, index string) (*Histogram, error)
}
