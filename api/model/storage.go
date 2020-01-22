//
//   Copyright © 2019 Uncharted Software Inc.
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package model

import (
	"time"

	"github.com/uncharted-distil/distil-compute/model"
)

// DataStorageCtor represents a client constructor to instantiate a data
// storage client.
type DataStorageCtor func() (DataStorage, error)

// DataStorage defines the functions available to query the underlying data storage.
type DataStorage interface {
	FetchNumRows(storageName string, variables []*model.Variable, filters map[string]interface{}) (int, error)
	FetchData(dataset string, storageName string, filterParams *FilterParams, invert bool) (*FilteredData, error)
	FetchSummary(dataset string, storageName string, varName string, filterParams *FilterParams, invert bool, mode SummaryMode) (*VariableSummary, error)
	FetchSummaryByResult(dataset string, storageName string, varName string, resultURI string, filterParams *FilterParams, extrema *Extrema, mode SummaryMode) (*VariableSummary, error)
	PersistResult(dataset string, storageName string, resultURI string, target string) error
	PersistSolutionFeatureWeight(dataset string, storageName string, solutionID string, weights [][]string) error
	FetchResults(dataset string, storageName string, resultURI string, solutionID string, filterParams *FilterParams, removeTargetColumn bool) (*FilteredData, error)
	FetchPredictedSummary(dataset string, storageName string, resultURI string, filterParams *FilterParams, extrema *Extrema) (*VariableSummary, error)
	FetchResultsExtremaByURI(dataset string, storageName string, resultURI string) (*Extrema, error)
	FetchCorrectnessSummary(dataset string, storageName string, resultURI string, filterParams *FilterParams) (*VariableSummary, error)
	FetchResidualsSummary(dataset string, storageName string, resultURI string, filterParams *FilterParams, extrema *Extrema) (*VariableSummary, error)
	FetchResidualsExtremaByURI(dataset string, storageName string, resultURI string) (*Extrema, error)
	FetchExtrema(storageName string, variable *model.Variable) (*Extrema, error)
	FetchExtremaByURI(dataset string, storageName string, resultURI string, variable string) (*Extrema, error)
	FetchTimeseries(dataset string, storageName string, timeseriesColName string, xColName string, yColName string, timeseriesURI string, filterParams *FilterParams, invert bool) ([][]float64, error)
	FetchTimeseriesForecast(dataset string, storageName string, timeseriesColName string, xColName string, yColName string, timeseriesURI string, resultUUID string, filterParams *FilterParams) ([][]float64, error)
	FetchCategoryCounts(storageName string, variable *model.Variable) (map[string]int, error)
	FetchSolutionFeatureWeights(dataset string, resultURI string, d3mIndex int64) (*SolutionFeatureWeight, error)

	// Dataset manipulation
	IsValidDataType(dataset string, storageName string, varName string, varType string) (bool, error)
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
	PersistSolution(requestID string, solutionID string, initialSearchSolutionID string, createdTime time.Time) error
	PersistSolutionWeight(solutionID string, featureName string, featureIndex int64, weight float64) error
	PersistSolutionState(solutionID string, progress string, createdTime time.Time) error
	PersistSolutionResult(solutionID string, fittedSolutionID string, produceRequestID string, resultType string, resultUUID string, resultURI string, progress string, createdTime time.Time) error
	PersistSolutionScore(solutionID string, metric string, score float64) error
	UpdateRequest(requestID string, progress string, updatedTime time.Time) error
	FetchRequest(requestID string) (*Request, error)
	FetchRequestBySolutionID(solutionID string) (*Request, error)
	FetchRequestByFittedSolutionID(fittedSolutionID string) (*Request, error)
	FetchRequestByDatasetTarget(dataset string, target string, solutionID string) ([]*Request, error)
	FetchRequestFeatures(requestID string) ([]*Feature, error)
	FetchRequestFilters(requestID string, features []*Feature) (*FilterParams, error)
	FetchSolution(solutionID string) (*Solution, error)
	FetchSolutionWeights(solutionID string) ([]*SolutionWeight, error)
	FetchSolutionResultByUUID(resultUUID string) (*SolutionResult, error)
	FetchSolutionResults(solutionID string) ([]*SolutionResult, error)
	FetchSolutionResultsByFittedSolutionID(fittedSolutionID string) ([]*SolutionResult, error)
	FetchSolutionResultByProduceRequestID(produceRequestID string) (*SolutionResult, error)
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
	SearchDatasets(terms string, baseDataset *Dataset, includeIndex bool, includeMeta bool) ([]*Dataset, error)
	ImportDataset(id string, uri string) (string, error)

	// Dataset manipulation
	SetDataType(dataset string, varName string, varType string) error
	SetExtrema(dataset string, varName string, extrema *Extrema) error
	AddVariable(dataset string, varName string, varDisplayName string, varType string, varDistilRole string) error
	DeleteVariable(dataset string, varName string) error
	AddGroupedVariable(dataset string, varName string, varDisplayName string, varType string, varRole string, grouping model.Grouping) error
	RemoveGroupedVariable(datasetName string, grouping model.Grouping) error
}
