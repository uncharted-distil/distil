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
	FetchData(dataset string, filterParams *FilterParams, invert bool) (*FilteredData, error)
	FetchSummary(dataset string, varName string, filterParams *FilterParams) (*Histogram, error)
	FetchSummaryByResult(dataset string, varName string, resultURI string, filterParams *FilterParams, extrema *Extrema) (*Histogram, error)
	PersistResult(dataset string, resultURI string) error
	FetchResults(dataset string, resultURI string) (*FilteredData, error)
	FetchFilteredResults(dataset string, resultURI string, filterParams *FilterParams) (*FilteredData, error)
	FetchResultsSummary(dataset string, resultURI string, filterParams *FilterParams, extrema *Extrema) (*Histogram, error)
	FetchResultsExtremaByURI(dataset string, resultURI string) (*Extrema, error)
	FetchResidualsSummary(dataset string, resultURI string, filterParams *FilterParams, extrema *Extrema) (*Histogram, error)
	FetchResidualsExtremaByURI(dataset string, resultURI string) (*Extrema, error)
	FetchExtremaByURI(dataset string, resultURI string, variable string) (*Extrema, error)

	// Dataset manipulation
	SetDataType(dataset string, varName string, varType string) error
	AddVariable(dataset string, varName string, varType string) error
	DeleteVariable(dataset string, varName string) error
	UpdateVariable(dataset string, varName string, d3mIndex string, value string) error
}

// SolutionStorageCtor represents a client constructor to instantiate a
// solution storage client.
type SolutionStorageCtor func() (SolutionStorage, error)

// SolutionStorage defines the functions available to query the underlying
// solution storage.
type SolutionStorage interface {
	PersistRequest(requestID string, dataset string, progress string, createdTime time.Time) error
	PersistRequestFeature(requestID string, featureName string, featureType string) error
	PersistRequestFilters(requestID string, filters *FilterParams) error
	PersistSolution(requestID string, solutionID string, progress string, createdTime time.Time) error
	PersistSolutionResult(solutionID string, resultUUID string, resultURI string, progress string, createdTime time.Time) error
	PersistSolutionScore(solutionID string, metric string, score float64) error
	UpdateRequest(requestID string, progress string, updatedTime time.Time) error
	FetchRequest(requestID string) (*Request, error)
	FetchRequestBySolutionID(requestID string) (*Request, error)
	FetchRequestFeatures(requestID string) ([]*Feature, error)
	FetchRequestFilters(requestID string, features []*Feature) (*FilterParams, error)
	FetchSolution(solutionID string) (*Solution, error)
	FetchSolutionResultByRequestID(requestID string) ([]*SolutionResult, error)
	FetchSolutionResultByUUID(resultUUID string) (*SolutionResult, error)
	FetchSolutionResult(solutionID string) (*SolutionResult, error)
	FetchSolutionResultByDatasetTarget(dataset string, target string, solutionID string) ([]*Request, error)
	FetchSolutionScore(solutionID string) ([]*SolutionScore, error)
}

// MetadataStorageCtor represents a client constructor to instantiate a
// metadata storage client.
type MetadataStorageCtor func() (MetadataStorage, error)

// MetadataStorage defines the functions available to query the underlying
// metadata storage.
type MetadataStorage interface {
	FetchVariables(dataset string, includeIndex bool) ([]*Variable, error)
	FetchVariablesDisplay(dataset string) ([]*Variable, error)
	FetchVariable(dataset string, varName string) (*Variable, error)
	FetchVariableDisplay(dataset string, varName string) (*Variable, error)
	FetchDatasets(includeIndex bool) ([]*Dataset, error)
	SearchDatasets(terms string, includeIndex bool) ([]*Dataset, error)

	// Dataset manipulation
	SetDataType(dataset string, varName string, varType string) error
	AddVariable(dataset string, varName string, varType string, varDistilRole string) error
	DeleteVariable(dataset string, varName string) error
}
