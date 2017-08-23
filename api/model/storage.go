package model

// StorageCtor represents a client constructor to instantiate a storage
// client.
type StorageCtor func() (Storage, error)

// Storage defines the functions available to query the underlying data storage.
type Storage interface {
	FetchData(string, *FilterParams) (*FilteredData, error)
	FetchSummary(variable *Variable, dataset string) (*Histogram, error)
}
