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
	"math"
	"strconv"
	"time"

	"github.com/uncharted-distil/distil-compute/metadata"
	"github.com/uncharted-distil/distil-compute/model"
)

// NullableFloat64 is float64 with custom JSON marshalling to allow for NaN values
// to be handled gracefully.
type NullableFloat64 float64

// MarshalJSON provides a custom float JSON marshaller that will handle a NaN float64
// value by replacing it with empty data.
func (f NullableFloat64) MarshalJSON() ([]byte, error) {
	if math.IsNaN(float64(f)) {
		return []byte("null"), nil
	}
	return []byte(strconv.FormatFloat(float64(f), 'f', -1, 32)), nil
}

// TimeseriesObservation represents a timeseries value along with confidences.
type TimeseriesObservation struct {
	Value          NullableFloat64 `json:"value"`
	Time           float64         `json:"time"`
	ConfidenceLow  NullableFloat64 `json:"confidenceLow"`
	ConfidenceHigh NullableFloat64 `json:"confidenceHigh"`
}

// TimeseriesData represents the result of a timeseries request.
type TimeseriesData struct {
	Timeseries []*TimeseriesObservation
	IsDateTime bool
	Min        float64
	Max        float64
	Mean       float64
}

// DataStorageCtor represents a client constructor to instantiate a data
// storage client.
type DataStorageCtor func() (DataStorage, error)

// DataStorage defines the functions available to query the underlying data storage.
type DataStorage interface {
	FetchNumRows(storageName string, variables []*model.Variable) (int, error)
	FetchData(dataset string, storageName string, filterParams *FilterParams, invert bool) (*FilteredData, error)
	FetchDataset(dataset string, storageName string, invert bool, filterParams *FilterParams) ([][]string, error)
	FetchSummary(dataset string, storageName string, varName string, filterParams *FilterParams, invert bool, mode SummaryMode) (*VariableSummary, error)
	FetchSummaryByResult(dataset string, storageName string, varName string, resultURI string, filterParams *FilterParams, extrema *Extrema, mode SummaryMode) (*VariableSummary, error)
	PersistResult(dataset string, storageName string, resultURI string, target string) error
	PersistExplainedResult(dataset string, storageName string, resultURI string, explainResult *SolutionExplainResult) error
	PersistSolutionFeatureWeight(dataset string, storageName string, solutionID string, weights [][]string) error
	FetchResults(dataset string, storageName string, resultURI string, solutionID string, filterParams *FilterParams, removeTargetColumn bool) (*FilteredData, error)
	FetchPredictedSummary(dataset string, storageName string, resultURI string, filterParams *FilterParams, extrema *Extrema, mode SummaryMode) (*VariableSummary, error)
	FetchResultsExtremaByURI(dataset string, storageName string, resultURI string) (*Extrema, error)
	FetchCorrectnessSummary(dataset string, storageName string, resultURI string, filterParams *FilterParams, mode SummaryMode) (*VariableSummary, error)
	FetchConfidenceSummary(dataset string, storageName string, resultURI string, filterParams *FilterParams, mode SummaryMode) (*VariableSummary, error)
	FetchResidualsSummary(dataset string, storageName string, resultURI string, filterParams *FilterParams, extrema *Extrema, mode SummaryMode) (*VariableSummary, error)
	FetchResidualsExtremaByURI(dataset string, storageName string, resultURI string) (*Extrema, error)
	FetchExtrema(storageName string, variable *model.Variable) (*Extrema, error)
	FetchExtremaByURI(dataset string, storageName string, resultURI string, variable string) (*Extrema, error)
	FetchTimeseries(dataset string, storageName string, timeseriesColName string, xColName string, yColName string, timeseriesURI []string, filterParams *FilterParams, invert bool) (*map[string]*TimeseriesData, error)
	FetchTimeseriesForecast(dataset string, storageName string, timeseriesColName string, xColName string, yColName string, timeseriesURIs []string, resultUUID string, filterParams *FilterParams) (*map[string]*TimeseriesData, error)
	FetchCategoryCounts(storageName string, variable *model.Variable) (map[string]int, error)
	FetchSolutionFeatureWeights(dataset string, storageName string, resultURI string, d3mIndex int64) (*SolutionFeatureWeight, error)
	// Dataset manipulation
	IsValidDataType(dataset string, storageName string, varName string, varType string) (bool, error)
	SetDataType(dataset string, storageName string, varName string, varType string) error
	AddVariable(dataset string, storageName string, varName string, varType string, defaultVal string) error
	AddField(dataset string, storageName string, varName string, varType string, defaultVal string) error
	DeleteVariable(dataset string, storageName string, varName string) error
	UpdateVariable(storageName string, varName string, d3mIndex string, value string) error
	UpdateVariableBatch(storageName string, varName string, updates map[string]string) error
	UpdateData(dataset string, storageName string, varName string, updates map[string]string, filterParams *FilterParams) error
	DoesVariableExist(dataset string, storageName string, varName string) (bool, error)
	VerifyData(datasetID string, tableName string) error
	// Raw data queries
	FetchRawDistinctValues(dataset string, storageName string, varNames []string) ([][]string, error)

	// Property queries
	GetStorageName(dataset string) (string, error)

	// CloneDataset creates a copy of an existing dataset
	CloneDataset(dataset string, storageName string, datasetNew string, storageNameNew string) error
}

// SolutionStorageCtor represents a client constructor to instantiate a
// solution storage client.
type SolutionStorageCtor func() (SolutionStorage, error)

// SolutionStorage defines the functions available to query the underlying
// solution storage.
type SolutionStorage interface {
	PersistPrediction(requestID string, dataset string, target string, fittedSolutionID string, progress string, createdTime time.Time) error
	PersistRequest(requestID string, dataset string, progress string, createdTime time.Time) error
	PersistRequestFeature(requestID string, featureName string, featureType string) error
	PersistRequestFilters(requestID string, filters *FilterParams) error
	PersistSolution(requestID string, solutionID string, explainedSolutionID string, createdTime time.Time) error
	PersistSolutionWeight(solutionID string, featureName string, featureIndex int64, weight float64) error
	PersistSolutionState(solutionID string, progress string, createdTime time.Time) error
	PersistSolutionResult(solutionID string, fittedSolutionID string, produceRequestID string, resultType string, resultUUID string, resultURI string, progress string, createdTime time.Time) error
	PersistSolutionExplainedOutput(resultUUID string, explainOutput map[string]*SolutionExplainResult) error
	PersistSolutionScore(solutionID string, metric string, score float64) error
	UpdateRequest(requestID string, progress string, updatedTime time.Time) error
	UpdateSolution(solutionID string, explainedSolutionID string) error
	FetchRequest(requestID string) (*Request, error)
	FetchRequestBySolutionID(solutionID string) (*Request, error)
	FetchRequestByFittedSolutionID(fittedSolutionID string) (*Request, error)
	FetchRequestByDatasetTarget(dataset string, target string) ([]*Request, error)
	FetchRequestFeatures(requestID string) ([]*Feature, error)
	FetchRequestFilters(requestID string, features []*Feature) (*FilterParams, error)
	FetchSolution(solutionID string) (*Solution, error)
	FetchExplainValues(dataset string, storageName string, d3mIndex []int, resultUUID string) ([]SolutionExplainValues, error)
	FetchSolutionsByDatasetTarget(dataset string, target string) ([]*Solution, error)
	FetchSolutionsByRequestID(requestID string) ([]*Solution, error)
	FetchSolutionWeights(solutionID string) ([]*SolutionWeight, error)
	FetchSolutionResultByUUID(resultUUID string) (*SolutionResult, error)
	FetchSolutionResults(solutionID string) ([]*SolutionResult, error)
	FetchSolutionResultsByFittedSolutionID(fittedSolutionID string) ([]*SolutionResult, error)
	FetchSolutionResultByProduceRequestID(produceRequestID string) (*SolutionResult, error)
	FetchPredictionResultByProduceRequestID(produceRequestID string) (*SolutionResult, error)
	FetchPredictionResultByUUID(reusultUUID string) (*SolutionResult, error)
	FetchSolutionScores(solutionID string) ([]*SolutionScore, error)
	FetchPrediction(requestID string) (*Prediction, error)
	FetchPredictionsByFittedSolutionID(fittedSolutionID string) ([]*Prediction, error)
}

// MetadataStorageCtor represents a client constructor to instantiate a
// metadata storage client.
type MetadataStorageCtor func() (MetadataStorage, error)

// MetadataStorage defines the functions available to query the underlying
// metadata storage.
type MetadataStorage interface {
	FetchVariables(dataset string, includeIndex bool, includeMeta bool) ([]*model.Variable, error)
	FetchVariablesByName(dataset string, varNames []string, includeIndex bool, includeMeta bool) ([]*model.Variable, error)
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
	AddGroupedVariable(dataset string, varName string, varDisplayName string, varType string, varRole string, grouping model.BaseGrouping) error
	RemoveGroupedVariable(datasetName string, grouping model.BaseGrouping) error
	DeleteDataset(dataset string) error
	IngestDataset(datasetSource metadata.DatasetSource, meta *model.Metadata) error
	UpdateDataset(dataset *Dataset) error

	// CloneDataset creates a copy of an existing dataset
	CloneDataset(dataset string, datasetNew string, storageNameNew string, folderNew string) error
}

// ExportedModelStorageCtor represents a client constructor to instantiate a
// model storage client.
type ExportedModelStorageCtor func() (ExportedModelStorage, error)

// ExportedModelStorage defines the functions available to query the underlying
// model storage.
type ExportedModelStorage interface {
	PersistExportedModel(exportedModel *ExportedModel) error
	FetchModel(model string) (*ExportedModel, error)
	FetchModelByID(fittedSolutionID string) (*ExportedModel, error)
	FetchModels() ([]*ExportedModel, error)
	SearchModels(terms string) ([]*ExportedModel, error)
}
