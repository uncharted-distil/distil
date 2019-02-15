package model

import (
	"time"

	"github.com/unchartedsoftware/distil-compute/model"
)

// DataStorageCtor represents a client constructor to instantiate a data
// storage client.
type DataStorageCtor func() (DataStorage, error)

// DataStorage defines the functions available to query the underlying data storage.
type DataStorage interface {
	FetchNumRows(storageName string, filters map[string]interface{}) (int, error)
	FetchData(dataset string, storageName string, filterParams *FilterParams, invert bool) (*FilteredData, error)
	FetchSummary(dataset string, storageName string, varName string, filterParams *FilterParams) (*Histogram, error)
	FetchSummaryByResult(dataset string, storageName string, varName string, resultURI string, filterParams *FilterParams, extrema *Extrema) (*Histogram, error)
	PersistResult(dataset string, storageName string, resultURI string, target string) error
	FetchResults(dataset string, storageName string, resultURI string, solutionID string, filterParams *FilterParams) (*FilteredData, error)
	FetchPredictedSummary(dataset string, storageName string, resultURI string, filterParams *FilterParams, extrema *Extrema) (*Histogram, error)
	FetchResultsExtremaByURI(dataset string, storageName string, resultURI string) (*Extrema, error)
	FetchCorrectnessSummary(dataset string, storageName string, resultURI string, filterParams *FilterParams) (*Histogram, error)
	FetchResidualsSummary(dataset string, storageName string, resultURI string, filterParams *FilterParams, extrema *Extrema) (*Histogram, error)
	FetchResidualsExtremaByURI(dataset string, storageName string, resultURI string) (*Extrema, error)
	FetchExtremaByURI(dataset string, storageName string, resultURI string, variable string) (*Extrema, error)
	FetchTimeseries(dataset string, storageName string, timeseriesColName string, xColName string, yColName string, timeseriesURI string, filterParams *FilterParams) ([][]float64, error)
	// Dataset manipulation
	SetDataType(dataset string, storageName string, varName string, varType string) error
	AddVariable(dataset string, storageName string, varName string, varType string) error
	DeleteVariable(dataset string, storageName string, varName string) error
	UpdateVariable(storageName string, varName string, d3mIndex string, value string) error
	UpdateVariableBatch(storageName string, varName string, updates map[string]string) error
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
	PersistSolutionResult(solutionID string, fittedSolutionID, resultUUID string, resultURI string, progress string, createdTime time.Time) error
	PersistSolutionScore(solutionID string, metric string, score float64) error
	UpdateRequest(requestID string, progress string, updatedTime time.Time) error
	FetchRequest(requestID string) (*Request, error)
	FetchRequestBySolutionID(requestID string) (*Request, error)
	FetchRequestByDatasetTarget(dataset string, target string, solutionID string) ([]*Request, error)
	FetchRequestFeatures(requestID string) ([]*Feature, error)
	FetchRequestFilters(requestID string, features []*Feature) (*FilterParams, error)
	FetchSolution(solutionID string) (*Solution, error)
	FetchSolutionResultByUUID(resultUUID string) (*SolutionResult, error)
	FetchSolutionResult(solutionID string) (*SolutionResult, error)
	FetchSolutionScores(solutionID string) ([]*SolutionScore, error)
}

// MetadataStorageCtor represents a client constructor to instantiate a
// metadata storage client.
type MetadataStorageCtor func() (MetadataStorage, error)

// MetadataStorage defines the functions available to query the underlying
// metadata storage.
type MetadataStorage interface {
	FetchVariables(dataset string, includeIndex bool, includeMeta bool) ([]*model.Variable, error)
	FetchVariablesDisplay(dataset string) ([]*model.Variable, error)
	DoesVariableExist(dataset string, varName string) (bool, error)
	FetchVariable(dataset string, varName string) (*model.Variable, error)
	FetchVariableDisplay(dataset string, varName string) (*model.Variable, error)
	FetchDataset(dataset string, includeIndex bool, includeMeta bool) (*Dataset, error)
	FetchDatasets(includeIndex bool, includeMeta bool) ([]*Dataset, error)
	SearchDatasets(terms string, includeIndex bool, includeMeta bool) ([]*Dataset, error)
	ImportDataset(id string, uri string) (string, error)

	// Dataset manipulation
	SetDataType(dataset string, varName string, varType string) error
	AddVariable(dataset string, varName string, varType string, varDistilRole string) error
	DeleteVariable(dataset string, varName string) error
}
