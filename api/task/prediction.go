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

package task

import (
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/metadata"
	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil-compute/primitive/compute"
	comp "github.com/uncharted-distil/distil/api/compute"
	"github.com/uncharted-distil/distil/api/env"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/serialization"
	"github.com/uncharted-distil/distil/api/util"
	log "github.com/unchartedsoftware/plog"
)

const (
	// DefaultSeparator is the default separator to use when dealing with groupings.
	DefaultSeparator = "_"
)

// PredictionTimeseriesDataset has the paramaters necessary to create a timeseries dataset
// from minimal information.
type PredictionTimeseriesDataset struct {
	params               *PredictParams
	start                int64
	interval             float64
	count                int
	isDatetimeTimeseries bool
	idValues             [][]string
	idKeys               []string
	timestampVariable    *model.Variable
}

type predictionDataset struct {
	params *PredictParams
}

// NewPredictionTimeseriesDataset creates prediction timeseries dataset.
func NewPredictionTimeseriesDataset(params *PredictParams, interval float64, count int) (*PredictionTimeseriesDataset, error) {
	// get the timestamp variable
	variables, err := params.MetaStorage.FetchVariables(params.Dataset, true, true, false)
	if err != nil {
		return nil, err
	}
	var groupingVar *model.Variable
	for _, v := range variables {
		if v.IsGrouping() {
			groupingVar = v
			break
		}
	}

	tsg := groupingVar.Grouping.(*model.TimeseriesGrouping)
	var timestampVar *model.Variable
	for _, v := range variables {
		if v.Key == tsg.XCol {
			timestampVar = v
			break
		}
	}

	// determine the start date via timestamp extrema
	extrema, err := params.DataStorage.FetchExtrema(params.Meta.StorageName, timestampVar)
	if err != nil {
		return nil, err
	}

	// get existing id values
	idValues, err := params.DataStorage.FetchRawDistinctValues(params.Meta.ID, params.Meta.StorageName, tsg.SubIDs)
	if err != nil {
		return nil, err
	}

	return &PredictionTimeseriesDataset{
		params:               params,
		interval:             interval,
		count:                count,
		isDatetimeTimeseries: model.IsDateTime(extrema.Type),
		start:                int64(extrema.Max + interval),
		idValues:             idValues,
		idKeys:               tsg.SubIDs,
		timestampVariable:    timestampVar,
	}, nil
}

// CreateDataset creates a raw dataset based on minimum timeseries parameters.
func (p *PredictionTimeseriesDataset) CreateDataset(rootDataPath string, datasetName string, config *env.Config) (*api.RawDataset, error) {
	// generate timestamps to use for prediction based on type of timestamp
	var timestampPredictionValues []string
	if model.IsDateTime(p.timestampVariable.Type) {
		timestampPredictionValues = generateTimestampValues(p.interval, p.start, p.count)
	} else if model.IsNumerical(p.timestampVariable.Type) || model.IsTimestamp(p.timestampVariable.Type) {
		timestampPredictionValues = generateIntValues(p.interval, p.start, p.count)
	} else {
		return nil, errors.Errorf("timestamp variable '%s' is type '%s' which is not supported for timeseries creation", p.timestampVariable.Key, p.timestampVariable.Type)
	}

	timeseriesData := createTimeseriesData(p.idKeys, p.idValues, p.timestampVariable.Key, timestampPredictionValues)

	return &api.RawDataset{
		ID:       p.params.Dataset,
		Name:     p.params.Dataset,
		Data:     timeseriesData,
		Metadata: p.params.Meta,
	}, nil
}

func (p *predictionDataset) CreateDataset(rootDataPath string, datasetName string, config *env.Config) (*api.RawDataset, error) {
	// need to do a bit of processing on the usual setup
	ds, err := p.params.DatasetConstructor.CreateDataset(rootDataPath, datasetName, config)
	if err != nil {
		return nil, err
	}

	// updated the new dataset to match the var types and ordering of the source dataset - required
	// so that the model lines up
	variables := p.params.Meta.GetMainDataResource().Variables
	csvDataAugmented, err := augmentPredictionDataset(ds.Data, variables, ds.Metadata.GetMainDataResource().Variables)
	if err != nil {
		return nil, err
	}
	dataResourcesMap := map[string]*model.DataResource{}
	for _, dataResource := range ds.Metadata.DataResources {
		dataResourcesMap[dataResource.ResID] = dataResource
	}
	// update the data resources to match those from the created dataset - they may have changed file types
	for i := range ds.Metadata.DataResources {
		dataResource := dataResourcesMap[p.params.Meta.DataResources[i].ResID]
		p.params.Meta.DataResources[i].ResFormat = dataResource.ResFormat
		p.params.Meta.DataResources[i].ResPath = dataResource.ResPath
	}

	return &api.RawDataset{
		ID:       p.params.Dataset,
		Name:     p.params.Dataset,
		Data:     csvDataAugmented,
		Metadata: p.params.Meta,
	}, nil
}

// PredictParams contains all parameters passed to the predict function.
type PredictParams struct {
	Meta               *model.Metadata
	SourceDataset      *api.Dataset
	Dataset            string
	SchemaPath         string
	SourceDatasetID    string
	SolutionID         string
	FittedSolutionID   string
	DatasetConstructor DatasetConstructor
	OutputPath         string
	Target             *model.Variable
	MetaStorage        api.MetadataStorage
	DataStorage        api.DataStorage
	SolutionStorage    api.SolutionStorage
	ModelStorage       api.ExportedModelStorage
	IngestConfig       *IngestTaskConfig
	Config             *env.Config
}

// ImportPredictionDataset imports a dataset to be used for predictions.
func ImportPredictionDataset(params *PredictParams) (string, string, error) {
	meta := params.Meta
	schemaPath := ""
	datasetName := fmt.Sprintf("pred_%s", params.Dataset)

	predictionDatasetCtor := &predictionDataset{
		params: params,
	}

	// create the dataset to be used for predictions
	datasetName, datasetPath, err := CreateDataset(datasetName, predictionDatasetCtor, params.OutputPath, params.Config)
	if err != nil {
		return "", "", err
	}
	log.Infof("created dataset for new data with id '%s' found at location '%s'", datasetName, datasetPath)

	// read the header of the new dataset to get the field names
	// if they dont match the original, then cant use the same pipeline
	rawDataPath := path.Join(datasetPath, compute.D3MDataFolder, compute.D3MLearningData)
	rawCSVData, err := util.ReadCSVFile(rawDataPath, false)
	if err != nil {
		return "", "", errors.Wrap(err, "unable to parse header result")
	}
	rawHeader := rawCSVData[0]
	mainDR := meta.GetMainDataResource()
	for i, f := range rawHeader {
		// TODO: col index not necessarily the same as index and thats what needs checking
		// We check both name and display name as the pre-ingested datasets are keyed of display name
		if mainDR.Variables[i].Key != f && mainDR.Variables[i].HeaderName != f {
			return "", "", errors.Errorf("variables in new prediction file do not match variables in original dataset")
		}
	}
	log.Infof("dataset fields match original dataset fields")

	// update the dataset doc to reflect original types
	meta.ID = datasetName
	meta.Name = datasetName
	meta.StorageName = model.NormalizeDatasetID(datasetName)
	meta.DatasetFolder = path.Base(datasetPath)
	schemaPath = path.Join(datasetPath, compute.D3MDataSchema)
	datasetStorage := serialization.GetStorage(rawDataPath)
	err = datasetStorage.WriteMetadata(schemaPath, meta, true, false)
	if err != nil {
		return "", "", errors.Wrap(err, "unable to update dataset doc")
	}
	log.Infof("wrote out schema doc for new dataset with id '%s' at location '%s'", meta.ID, schemaPath)

	return datasetName, schemaPath, nil
}

// IngestPredictionDataset ingests a dataset to be used for predictions.
func IngestPredictionDataset(params *PredictParams) error {
	schemaPath := params.SchemaPath
	// ingest the dataset but without running simon, duke, etc.
	err := IngestPostgres(schemaPath, schemaPath, metadata.Augmented, nil, params.IngestConfig, true, false, false)
	if err != nil {
		return errors.Wrap(err, "unable to ingest prediction data")
	}
	log.Infof("finished ingesting dataset '%s'", params.Dataset)

	// copy the metadata from the source dataset as it should be an exact match
	log.Infof("using datase '%s' as source for metadata", params.SourceDatasetID)
	metaClone, err := params.MetaStorage.FetchDataset(params.SourceDatasetID, true, true, true)
	if err != nil {
		return err
	}
	metaClone.ID = params.Dataset
	metaClone.StorageName = params.Meta.StorageName
	metaClone.Folder = params.Meta.DatasetFolder
	metaClone.Source = metadata.Augmented
	metaClone.Type = api.DatasetTypeInference

	err = params.MetaStorage.UpdateDataset(metaClone)
	if err != nil {
		return err
	}

	// only featurize if the source dataset was featurized
	if params.Meta.LearningDataset != "" {
		if err = Featurize(schemaPath, schemaPath, params.DataStorage, params.MetaStorage, params.Dataset, params.IngestConfig); err != nil {
			return errors.Wrap(err, "unabled to featurize prediction data")
		}
	}

	// Apply the var types associated with the fitted solution to the inference data - the model types and input types should
	// should match.
	if err := updateVariableTypes(params.SolutionStorage, params.MetaStorage, params.DataStorage, params.FittedSolutionID, params.Dataset, metaClone.StorageName); err != nil {
		return err
	}

	// Handle grouped variables.
	target := params.Target
	if target.IsGrouping() && model.IsTimeSeries(target.Grouping.GetType()) {
		tsg := target.Grouping.(*model.TimeseriesGrouping)
		log.Infof("target is a timeseries so need to extract the prediction target from the grouping")
		target, err = params.MetaStorage.FetchVariable(metaClone.ID, tsg.YCol)
		if err != nil {
			return err
		}

		// need to run the grouping compose to create the needed ID column
		log.Infof("creating composed variables on prediction dataset '%s'", params.Dataset)
		err = CreateComposedVariable(params.MetaStorage, params.DataStorage, params.Dataset,
			metaClone.StorageName, tsg.IDCol, target.DisplayName, tsg.SubIDs)
		if err != nil {
			return err
		}

		err = params.MetaStorage.AddGroupedVariable(params.Dataset, params.Target.Key, params.Target.DisplayName,
			params.Target.Type, params.Target.DistilRole, tsg)
		if err != nil {
			return err
		}
		log.Infof("done creating compose variables")
	}

	// add feature groups
	err = copyFeatureGroups(params.FittedSolutionID, metaClone.ID, params.SolutionStorage, params.MetaStorage)
	if err != nil {
		return err
	}

	return nil
}

// Predict processes input data to generate predictions.
func Predict(params *PredictParams) (string, error) {
	log.Infof("generating predictions for fitted solution ID %s", params.FittedSolutionID)
	datasetPath := path.Join(params.OutputPath, params.Dataset)
	schemaPath := params.SchemaPath
	datasetName := params.Dataset

	// the dataset id needs to match the original dataset id for TA2 to be able to use the model
	// read from source in case any step has updated it along the way
	meta, err := metadata.LoadMetadataFromOriginalSchema(schemaPath, false)
	if err != nil {
		return "", errors.Wrap(err, "unable to read latest dataset doc")
	}
	meta.ID = params.SourceDatasetID
	datasetStorage := serialization.GetStorage(schemaPath)
	err = datasetStorage.WriteMetadata(schemaPath, meta, true, false)
	if err != nil {
		return "", errors.Wrap(err, "unable to update dataset doc")
	}

	// get the explained solution id
	solution, err := params.SolutionStorage.FetchSolution(params.SolutionID)
	if err != nil {
		return "", err
	}

	// Ensure the ta2 has fitted solution loaded.  If the model wasn't saved, it should be available
	// as part of the session.
	exportedModel, err := params.ModelStorage.FetchModelByID(params.FittedSolutionID)
	if err != nil {
		return "", err
	}
	if exportedModel != nil {
		_, err = LoadFittedSolution(exportedModel.FilePath, params.SolutionStorage, params.MetaStorage)
		if err != nil {
			return "", err
		}
	}

	// submit the new dataset for predictions
	log.Infof("generating predictions using data found at '%s'", datasetPath)
	predictionResult, err := comp.GeneratePredictions(datasetPath, solution.SolutionID, params.FittedSolutionID, client)
	if err != nil {
		return "", err
	}
	log.Infof("generated predictions stored at %v", predictionResult.ResultURI)

	// get the result UUID. NOTE: Doing sha1 for now.
	resultID, err := util.Hash(predictionResult.ResultURI)
	if err != nil {
		return "", err
	}

	err = persistPredictionResults(datasetName, params, meta, resultID, predictionResult)
	if err != nil {
		return "", err
	}

	return predictionResult.ProduceRequestID, nil
}

func persistPredictionResults(datasetName string, params *PredictParams, meta *model.Metadata, resultID string, predictionResult *comp.PredictionResult) error {
	if predictionResult.StepFeatureWeightURI != "" {
		featureWeights, err := comp.ExplainFeatureOutput(predictionResult.ResultURI, predictionResult.StepFeatureWeightURI)
		if err != nil {
			return err
		}
		err = params.DataStorage.PersistSolutionFeatureWeight(datasetName, meta.StorageName, featureWeights.ResultURI, featureWeights.Values)
		if err != nil {
			return err
		}
	}
	log.Infof("stored feature weights to the database")

	createdTime := time.Now()
	err := params.SolutionStorage.PersistPrediction(predictionResult.ProduceRequestID, datasetName, params.Target.Key, params.FittedSolutionID, "PREDICT_COMPLETED", createdTime)
	if err != nil {
		return err
	}
	err = params.SolutionStorage.PersistSolutionResult(params.SolutionID, params.FittedSolutionID, predictionResult.ProduceRequestID, api.SolutionResultTypeInference, resultID, predictionResult.ResultURI, "PREDICT_COMPLETED", createdTime)
	if err != nil {
		return err
	}

	target, err := resolveTarget(datasetName, params.Target, params.MetaStorage)
	if err != nil {
		return err
	}

	err = params.DataStorage.PersistResult(datasetName, meta.StorageName, predictionResult.ResultURI, target.Key)
	if err != nil {
		return err
	}

	err = params.DataStorage.PersistExplainedResult(datasetName, meta.StorageName, predictionResult.ResultURI, predictionResult.Confidences)
	if err != nil {
		return err
	}
	log.Infof("stored prediction results for %s to the database", predictionResult.ProduceRequestID)

	return nil
}

func augmentPredictionDataset(csvData [][]string, sourceVariables []*model.Variable, predictionVariables []*model.Variable) ([][]string, error) {
	log.Infof("augmenting prediction dataset fields")

	// map fields to indices
	headerSource := make([]string, len(sourceVariables))
	sourceVariableMap := make(map[string]*model.Variable)
	sourceVariableHeaderMap := make(map[string]*model.Variable)
	for _, v := range sourceVariables {
		sourceVariableMap[strings.ToLower(v.Key)] = v
		sourceVariableHeaderMap[strings.ToLower(v.HeaderName)] = v
		headerSource[v.Index] = v.HeaderName
	}

	addIndex := true
	predictVariablesMap := make(map[int]int)

	// If the variable list for prediction set is empty (as is the case for tabular data) then we just use the
	// header values as the list of variable names to build the map.
	if len(predictionVariables) == 0 {
		for i, pv := range csvData[0] {
			varName := strings.ToLower(pv)
			if sourceVariableHeaderMap[varName] != nil {
				predictVariablesMap[i] = sourceVariableHeaderMap[varName].Index
				log.Infof("mapped '%s' to index %d", varName, predictVariablesMap[i])
			}
		}
	} else {
		// Otherwise, we have the variables defined, and leverage the extra info provided to help map columns between model
		// and prediction datasets.
		for i, predictVariable := range predictionVariables {
			varName := strings.ToLower(predictVariable.Key)
			if sourceVariableMap[varName] != nil {
				predictVariablesMap[i] = sourceVariableMap[varName].Index
				log.Infof("mapped '%s' to index %d", varName, predictVariablesMap[i])
			} else if predictVariable.IsMediaReference() {
				log.Warnf("media reference field '%s' not found in source dataset - attempting to match by type", predictVariable.Key)
				// loop back over the source vars utnil we find one that is also a media reference
				for _, sourceVariable := range sourceVariables {
					if sourceVariable.IsMediaReference() {
						predictVariablesMap[i] = sourceVariableMap[strings.ToLower(sourceVariable.Key)].Index
						break
					}
				}
			} else {
				log.Warnf("field '%s' not found in source dataset - column will be empty", predictVariable.Key)
				predictVariablesMap[i] = -1
			}
			if predictVariable.Key == model.D3MIndexName {
				addIndex = false
			}
		}
	}

	// check if a variable the model needs is missing
	if len(sourceVariables) > len(predictVariablesMap) {
		return nil, errors.Errorf("missing some variables for model so unable to get predictions (missing column count: %d)", len(sourceVariables)-len(predictVariablesMap))
	}

	// read the rest of the data
	log.Infof("rewriting prediction dataset to match source dataset structure")
	count := 0

	// read the d3m field index if present
	d3mFieldIndex := -1
	if variable, ok := sourceVariableMap[strings.ToLower(model.D3MIndexName)]; ok {
		d3mFieldIndex = variable.Index
	}

	outputData := [][]string{headerSource}
	for _, line := range csvData[1:] {
		// write the columns in the same order as the source dataset
		output := make([]string, len(predictVariablesMap))
		for i, f := range line {
			sourceIndex := predictVariablesMap[i]
			if sourceIndex >= 0 {
				output[sourceIndex] = f
			}
		}

		if addIndex && d3mFieldIndex >= 0 {
			output[d3mFieldIndex] = fmt.Sprintf("%d", count)
		}
		count = count + 1
		outputData = append(outputData, output)
	}

	log.Infof("done augmenting prediction dataset")

	return outputData, nil
}

// CreateComposedVariable creates a new variable to use as group id.
func CreateComposedVariable(metaStorage api.MetadataStorage, dataStorage api.DataStorage,
	dataset string, storageName string, composedVarName string, composedVarDisplayName string, sourceVarNames []string) error {

	// create the variable data store entry
	varExists, err := metaStorage.DoesVariableExist(dataset, composedVarName)
	if err != nil {
		return err
	}

	if !varExists {
		// create the variable metadata entry
		err := metaStorage.AddVariable(dataset, composedVarName, composedVarDisplayName, model.StringType, model.VarDistilRoleGrouping)
		if err != nil {
			return err
		}

		err = dataStorage.AddVariable(dataset, storageName, composedVarName, model.StringType, "")
		if err != nil {
			return err
		}
	}

	composedData := map[string]string{}
	var filter *api.FilterParams
	if len(sourceVarNames) > 0 {
		// Fetch data using the source names as the filter
		filter = &api.FilterParams{
			Variables: sourceVarNames,
		}
	} else {
		// No grouping column - just use the d3mIndex as we'll just stick some placeholder
		// data in.
		filter = &api.FilterParams{
			Variables: []string{model.D3MIndexName},
		}
	}
	rawData, err := dataStorage.FetchData(dataset, storageName, filter, false, nil)
	if err != nil {
		return err
	}

	// Create a map of the retreived fields to column number.  Store d3mIndex since it needs to be directly referenced
	// further along.
	d3mIndexFieldindex := -1
	colNameToIdx := make(map[string]int)
	for i, c := range rawData.Columns {
		if c.Label == model.D3MIndexName {
			d3mIndexFieldindex = i
		} else {
			colNameToIdx[c.Label] = i
		}
	}

	if len(sourceVarNames) > 0 {
		// Loop over the fetched data, composing each column value into a single new column value using the
		// separator.
		for _, r := range rawData.Values {
			// create the hash from the specified columns
			composed := createComposedFields(r, sourceVarNames, colNameToIdx, DefaultSeparator)
			composedData[fmt.Sprintf("%v", r[d3mIndexFieldindex].Value)] = composed
		}
	} else {
		// Loop over the fetched d3mIndex values and set a placeholder value.
		for _, r := range rawData.Values {
			composedData[fmt.Sprintf("%v", r[d3mIndexFieldindex].Value)] = "__timeseries"
		}
	}

	// Save the new column
	err = dataStorage.UpdateVariableBatch(storageName, composedVarName, composedData)
	if err != nil {
		return err
	}

	return nil
}

func createComposedFields(data []*api.FilteredDataValue, fields []string, mappedFields map[string]int, separator string) string {
	dataToJoin := make([]string, len(fields))
	for i, field := range fields {
		dataToJoin[i] = fmt.Sprintf("%v", data[mappedFields[field]].Value)
	}
	return strings.Join(dataToJoin, separator)
}

// Apply the var types associated with the fitted solution to the inference data - the model types and input types should
// should match.
func updateVariableTypes(solutionStorage api.SolutionStorage, metaStorage api.MetadataStorage,
	dataStorage api.DataStorage, fittedSolutionID string, dataset string, storageName string) error {
	solutionRequest, err := solutionStorage.FetchRequestByFittedSolutionID(fittedSolutionID)
	if err != nil {
		return err
	}

	// get a variable map for quick look up
	variables, err := metaStorage.FetchVariables(solutionRequest.Dataset, false, true, false)
	if err != nil {
		return err
	}
	variableMap := map[string]*model.Variable{}
	for _, variable := range variables {
		variableMap[variable.Key] = variable
	}

	for _, feature := range solutionRequest.Features {
		// if this is a grouped variable we need to treat its components separately
		if variable, ok := variableMap[feature.FeatureName]; ok {
			componentVarNames := getComponentVariables(variable)
			for _, componentVarName := range componentVarNames {
				if componentVar, ok := variableMap[componentVarName]; ok {
					// update variable type
					if err := dataStorage.SetDataType(dataset, storageName, componentVar.Key, componentVar.Type); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

// Extracts the list of components that used to create a compound variable.
func getComponentVariables(variable *model.Variable) []string {
	componentVars := []string{}
	if variable.IsGrouping() {
		if model.IsGeoCoordinate(variable.Grouping.GetType()) {
			gcg := variable.Grouping.(*model.GeoCoordinateGrouping)
			// Include X and Y col
			componentVars = append(componentVars, gcg.XCol, gcg.YCol)

			// include the grouping sub-ids if the ID is created from mutliple columns
			componentVars = append(componentVars, variable.Grouping.GetSubIDs()...)
			if variable.Grouping.GetIDCol() != "" {
				// include the grouping ID if present and there were no sub IDs
				componentVars = append(componentVars, variable.Grouping.GetIDCol())
			}
		} else if model.IsMultiBandImage(variable.Grouping.GetType()) {
			rsg := variable.Grouping.(*model.MultiBandImageGrouping)
			componentVars = append(componentVars, rsg.BandCol, rsg.IDCol, rsg.ImageCol)
		} else if model.IsTimeSeries(variable.Grouping.GetType()) {
			tsg := variable.Grouping.(*model.TimeseriesGrouping)
			componentVars = append(componentVars, tsg.XCol, tsg.YCol)
		} else if model.IsGeoBounds(variable.Grouping.GetType()) {
			componentVars = append(componentVars, variable.Grouping.GetHidden()...)
		}
	} else {
		componentVars = append(componentVars, variable.Key)
	}

	return componentVars
}

func copyFeatureGroups(fittedSolutionID string, datasetName string, solutionStorage api.SolutionStorage, metaStorage api.MetadataStorage) error {
	// get the features in the solution
	solutionRequest, err := solutionStorage.FetchRequestByFittedSolutionID(fittedSolutionID)
	if err != nil {
		return err
	}

	// get a variable map for quick look up
	variables, err := metaStorage.FetchVariables(solutionRequest.Dataset, false, true, false)
	if err != nil {
		return err
	}
	variableMap := createVarMap(variables, false, false)
	variablesPrediction, err := metaStorage.FetchVariables(datasetName, false, true, false)
	if err != nil {
		return err
	}
	variablePredictionMap := createVarMap(variablesPrediction, false, false)

	// copy over the groups that are found and dont already exist in the prediction dataset
	for _, feature := range solutionRequest.Features {
		if feature.FeatureType == "train" && variablePredictionMap[feature.FeatureName] == nil {
			sf := variableMap[feature.FeatureName]
			if sf.IsGrouping() && model.IsMultiBandImage(sf.Grouping.GetType()) {
				rsg := sf.Grouping.(*model.MultiBandImageGrouping)
				rsg.Dataset = datasetName
				err = metaStorage.AddGroupedVariable(datasetName, sf.Key, sf.DisplayName, sf.Type, sf.DistilRole, rsg)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func generateIntValues(interval float64, start int64, stepCount int) []string {
	// iterate until all required steps are created
	currentValue := start
	timeData := make([]string, 0)
	for i := 0; i < stepCount; i++ {
		timeData = append(timeData, fmt.Sprintf("%d", currentValue))
		currentValue = currentValue + int64(interval)
	}

	return timeData
}

func generateTimestampValues(interval float64, start int64, stepCount int) []string {
	// parse the start time
	startDate := time.Unix(start, 0)

	// iterate until all required steps are created
	currentTime := startDate
	intervalDuration := time.Duration(int64(interval)) * time.Second
	timeData := make([]string, 0)
	for i := 0; i < stepCount; i++ {
		timeData = append(timeData, currentTime.String())
		currentTime = currentTime.Add(intervalDuration)
	}

	return timeData
}

func createTimeseriesData(idFields []string, idValues [][]string, timestampFieldName string, timestampPredictionValues []string) [][]string {
	// create the header
	header := []string{model.D3MIndexFieldName}
	header = append(header, idFields...)
	header = append(header, timestampFieldName)

	// cycle through the distinct id values and generate one row / timestamp
	rowCount := 0
	generatedData := [][]string{header}
	for _, row := range idValues {
		for _, ts := range timestampPredictionValues {
			rowData := []string{fmt.Sprintf("%d", rowCount)}
			rowData = append(rowData, row...)
			rowData = append(rowData, ts)
			generatedData = append(generatedData, rowData)
			rowCount++
		}
	}

	return generatedData
}

func resolveTarget(datasetID string, target *model.Variable, metaStorage api.MetadataStorage) (*model.Variable, error) {
	trueTarget := target
	if target.IsGrouping() && model.IsTimeSeries(target.Grouping.GetType()) {
		tsg := target.Grouping.(*model.TimeseriesGrouping)
		log.Infof("target is a timeseries so need to extract the prediction target from the grouping")
		resolvedTarget, err := metaStorage.FetchVariable(datasetID, tsg.YCol)
		if err != nil {
			return nil, err
		}
		trueTarget = resolvedTarget
	}

	return trueTarget, nil
}
