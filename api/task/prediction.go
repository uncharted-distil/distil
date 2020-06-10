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
	"github.com/uncharted-distil/distil/api/util"
	log "github.com/unchartedsoftware/plog"
)

const (
	// DefaultSeparator is the default separator to use when dealing with groupings.
	DefaultSeparator = "_"
)

type predictionDataset struct {
	params *PredictParams
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

	// update the data resources to match those from the created dataset - they may have changed file types
	for i, dataResource := range ds.Metadata.DataResources {
		p.params.Meta.DataResources[i].ResFormat = dataResource.ResFormat
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
	SolutionID         string
	FittedSolutionID   string
	DatasetConstructor DatasetConstructor
	OutputPath         string
	Index              string
	Target             *model.Variable
	MetaStorage        api.MetadataStorage
	DataStorage        api.DataStorage
	SolutionStorage    api.SolutionStorage
	ModelStorage       api.ExportedModelStorage
	DatasetIngested    bool
	DatasetImported    bool
	IngestConfig       *IngestTaskConfig
	Config             *env.Config
}

// Predict processes input data to generate predictions.
func Predict(params *PredictParams) (*api.SolutionResult, error) {
	log.Infof("generating predictions for fitted solution ID %s", params.FittedSolutionID)
	meta := params.Meta
	sourceDatasetID := meta.ID
	datasetPath := path.Join(params.OutputPath, params.Dataset)
	schemaPath := ""
	var err error

	// if the dataset was already imported, then just produce on it
	if params.DatasetImported {
		schemaPath = path.Join(datasetPath, compute.D3MDataSchema)
		log.Infof("dataset already imported at %s", datasetPath)
	} else {
		predictionDatasetCtor := &predictionDataset{
			params: params,
		}

		// create the dataset to be used for predictions
		_, _, err = CreateDataset(params.Dataset, predictionDatasetCtor, params.OutputPath, params.Config)
		if err != nil {
			return nil, err
		}
		log.Infof("created dataset for new data")

		// read the header of the new dataset to get the field names
		// if they dont match the original, then cant use the same pipeline
		rawDataPath := path.Join(datasetPath, compute.D3MDataFolder, compute.D3MLearningData)
		rawCSVData, err := util.ReadCSVFile(rawDataPath, false)
		if err != nil {
			return nil, errors.Wrap(err, "unable to parse header result")
		}
		rawHeader := rawCSVData[0]
		for i, f := range rawHeader {
			// TODO: may have to check the name rather than display name
			// TODO: col index not necessarily the same as index and thats what needs checking
			if meta.GetMainDataResource().Variables[i].DisplayName != f {
				return nil, errors.Errorf("variables in new prediction file do not match variables in original dataset")
			}
		}
		log.Infof("dataset fields match original dataset fields")

		// update the dataset doc to reflect original types
		meta.ID = params.Dataset
		meta.StorageName = model.NormalizeDatasetID(params.Dataset)
		meta.DatasetFolder = path.Base(datasetPath)
		schemaPath = path.Join(datasetPath, compute.D3MDataSchema)
		err = metadata.WriteSchema(meta, schemaPath, true)
		if err != nil {
			return nil, errors.Wrap(err, "unable to update dataset doc")
		}
		log.Infof("wrote out schema doc for new dataset")
	}

	if !params.DatasetIngested {
		// ingest the dataset but without running simon, duke, etc.
		_, err = Ingest(schemaPath, schemaPath, params.MetaStorage, params.Index, params.Dataset,
			metadata.Augmented, nil, api.DatasetTypeInference, params.IngestConfig, false, false)
		if err != nil {
			return nil, errors.Wrap(err, "unable to ingest ranked data")
		}
		log.Infof("finished ingesting the dataset")
	}

	// Apply the var types associated with the fitted solution to the inference data - the model types and input types should
	// should match.
	if err := updateVariableTypes(params.SolutionStorage, params.MetaStorage, params.DataStorage, params.FittedSolutionID, params.Dataset); err != nil {
		return nil, err
	}

	// Handle grouped variables.
	target := params.Target
	if target.IsGrouping() && model.IsTimeSeries(target.Grouping.GetType()) {
		tsg := target.Grouping.(*model.TimeseriesGrouping)
		log.Infof("target is a timeseries so need to extract the prediction target from the grouping")
		target, err = params.MetaStorage.FetchVariable(meta.ID, tsg.YCol)
		if err != nil {
			return nil, err
		}

		// need to run the grouping compose to create the needed ID column
		log.Infof("creating composed variables on inferrence dataset '%s'", params.Dataset)
		err = CreateComposedVariable(params.MetaStorage, params.DataStorage, params.Dataset,
			tsg.IDCol, tsg.IDCol, tsg.SubIDs)
		if err != nil {
			return nil, err
		}

		err = params.MetaStorage.AddGroupedVariable(params.Dataset, params.Target.Name, params.Target.DisplayName,
			params.Target.Type, params.Target.DistilRole, tsg)
		if err != nil {
			return nil, err
		}
		log.Infof("done creating compose variables")
	}

	// the dataset id needs to match the original dataset id for TA2 to be able to use the model
	meta.ID = sourceDatasetID
	err = metadata.WriteSchema(meta, schemaPath, true)
	if err != nil {
		return nil, errors.Wrap(err, "unable to update dataset doc")
	}

	// get the explained solution id
	solution, err := params.SolutionStorage.FetchSolution(params.SolutionID)
	if err != nil {
		return nil, err
	}

	// Ensure the ta2 has fitted solution loaded.  If the model wasn't saved, it should be available
	// as part of the session.
	exportedModel, err := params.ModelStorage.FetchModelByID(params.FittedSolutionID)
	if err != nil {
		return nil, err
	}
	if exportedModel != nil {
		_, err = LoadFittedSolution(exportedModel.FilePath, params.SolutionStorage, params.MetaStorage)
		if err != nil {
			return nil, err
		}
	}

	// submit the new dataset for predictions
	predictionResult, err := comp.GeneratePredictions(datasetPath, solution.ExplainedSolutionID, params.FittedSolutionID, client)
	if err != nil {
		return nil, err
	}
	log.Infof("generated predictions stored at %v", predictionResult.ResultURI)

	if predictionResult.StepFeatureWeightURI != "" {
		featureWeights, err := comp.ExplainFeatureOutput(predictionResult.ResultURI, schemaPath, predictionResult.StepFeatureWeightURI)
		if err != nil {
			return nil, err
		}
		err = params.DataStorage.PersistSolutionFeatureWeight(params.Dataset, model.NormalizeDatasetID(params.Dataset), featureWeights.ResultURI, featureWeights.Values)
		if err != nil {
			return nil, err
		}
	}
	log.Infof("stored feature weights to the database")

	// get the result UUID. NOTE: Doing sha1 for now.
	resultID, err := util.Hash(predictionResult.ResultURI)
	if err != nil {
		return nil, err
	}

	// Persist the prediction request metadata
	createdTime := time.Now()
	err = params.SolutionStorage.PersistPrediction(predictionResult.ProduceRequestID, params.Dataset, params.Target.Name, params.FittedSolutionID, "PREDICT_COMPLETED", createdTime)
	if err != nil {
		return nil, err
	}
	err = params.SolutionStorage.PersistSolutionResult(params.SolutionID, params.FittedSolutionID, predictionResult.ProduceRequestID, api.SolutionResultTypeInference, resultID, predictionResult.ResultURI, "PREDICT_COMPLETED", createdTime)
	if err != nil {
		return nil, err
	}

	err = params.DataStorage.PersistResult(params.Dataset, model.NormalizeDatasetID(params.Dataset), predictionResult.ResultURI, target.Name, nil)
	if err != nil {
		return nil, err
	}
	log.Infof("stored prediction results to the database")

	// set the dataset to the inference dataset
	res, err := params.SolutionStorage.FetchPredictionResultByProduceRequestID(predictionResult.ProduceRequestID)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func augmentPredictionDataset(csvData [][]string, sourceVariables []*model.Variable, predictionVariables []*model.Variable) ([][]string, error) {
	log.Infof("augment inference dataset fields")

	// map fields to indices
	headerSource := make([]string, len(sourceVariables))
	sourceVariableMap := make(map[string]*model.Variable)
	for _, v := range sourceVariables {
		sourceVariableMap[v.DisplayName] = v
		headerSource[v.Index] = v.DisplayName
	}

	addIndex := true
	predictVariablesMap := make(map[int]int)

	// If the variable list for prediction set is empty (as is the case for tabular data) then we just use the
	// header values as the list of variable names to build the map.
	if len(predictionVariables) == 0 {
		for i, pv := range csvData[0] {
			if sourceVariableMap[pv] != nil {
				predictVariablesMap[i] = sourceVariableMap[pv].Index
				log.Infof("mapped '%s' to index %d", pv, predictVariablesMap[i])
			}
		}
	} else {
		// Otherwise, we have the variables defined, and leverage the extra info provided to help map columns between model
		// and prediction datasets.
		for i, predictVariable := range predictionVariables {
			if sourceVariableMap[predictVariable.Name] != nil {
				predictVariablesMap[i] = sourceVariableMap[predictVariable.Name].Index
				log.Infof("mapped '%s' to index %d", predictVariable.Name, predictVariablesMap[i])
			} else if predictVariable.IsMediaReference() {
				log.Warnf("media reference field '%s' not found in source dataset - attempting to match by type", predictVariable.Name)
				// loop back over the source vars utnil we find one that is also a media reference
				for _, sourceVariable := range sourceVariables {
					if sourceVariable.IsMediaReference() {
						predictVariablesMap[i] = sourceVariableMap[sourceVariable.Name].Index
						break
					}
				}
			} else {
				log.Warnf("field '%s' not found in source dataset - column will be empty", predictVariable.Name)
				predictVariablesMap[i] = -1
			}
			if predictVariable.Name == model.D3MIndexName {
				addIndex = false
			}
		}
	}

	// read the rest of the data
	log.Infof("rewriting inference dataset to match source dataset structure")
	count := 0

	// read the d3m field index if present
	d3mFieldIndex := -1
	if variable, ok := sourceVariableMap[model.D3MIndexName]; ok {
		d3mFieldIndex = variable.Index
	}

	outputData := [][]string{headerSource}
	for _, line := range csvData[1:] {
		// write the columns in the same order as the source dataset
		output := make([]string, len(sourceVariableMap))
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

	log.Infof("done augmenting inference dataset")

	return outputData, nil
}

// CreateComposedVariable creates a new variable to use as group id.
func CreateComposedVariable(metaStorage api.MetadataStorage, dataStorage api.DataStorage,
	dataset string, composedVarName string, composedVarDisplayName string, sourceVarNames []string) error {

	// create the variable data store entry
	datasetStorageName := model.NormalizeDatasetID(dataset)

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

		err = dataStorage.AddVariable(dataset, datasetStorageName, composedVarName, model.StringType)
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
	rawData, err := dataStorage.FetchData(dataset, datasetStorageName, filter, false)
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
	err = dataStorage.UpdateVariableBatch(datasetStorageName, composedVarName, composedData)
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
func updateVariableTypes(solutionStorage api.SolutionStorage, metaStorage api.MetadataStorage, dataStorage api.DataStorage, fittedSolutionID string, dataset string) error {
	solutionRequest, err := solutionStorage.FetchRequestByFittedSolutionID(fittedSolutionID)
	if err != nil {
		return err
	}

	// get a variable map for quick look up
	variables, err := metaStorage.FetchVariables(solutionRequest.Dataset, false, true)
	if err != nil {
		return err
	}
	variableMap := map[string]*model.Variable{}
	for _, variable := range variables {
		variableMap[variable.Name] = variable
	}

	//
	storageName := model.NormalizeDatasetID(dataset)
	for _, feature := range solutionRequest.Features {
		// if this is a grouped variable we need to treat its components separately
		if variable, ok := variableMap[feature.FeatureName]; ok {
			componentVarNames := getComponentVariables(variable)
			for _, componentVarName := range componentVarNames {
				if componentVar, ok := variableMap[componentVarName]; ok {
					// update variable type
					if err := metaStorage.SetDataType(dataset, componentVar.Name, componentVar.Type); err != nil {
						return err
					}
					if err := dataStorage.SetDataType(dataset, storageName, componentVar.Name, componentVar.Type); err != nil {
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
	// only implemented for geo coordinate groups
	if variable.IsGrouping() && model.IsGeoCoordinate(variable.Grouping.GetType()) {
		gcg := variable.Grouping.(*model.GeoCoordinateGrouping)
		// Include X and Y col
		componentVars = append(componentVars, gcg.XCol, gcg.YCol)

		// include the grouping sub-ids if the ID is created from mutliple columns
		componentVars = append(componentVars, variable.Grouping.GetSubIDs()...)
		if variable.Grouping.GetIDCol() != "" {
			// include the grouping ID if present and there were no sub IDs
			componentVars = append(componentVars, variable.Grouping.GetIDCol())
		}
		return componentVars
	}
	return append(componentVars, variable.Name)
}
