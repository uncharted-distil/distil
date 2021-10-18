//
//   Copyright © 2021 Uncharted Software Inc.
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
	extrema, err := params.DataStorage.FetchExtrema(params.Meta.ID, params.Meta.StorageName, timestampVar)
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
func (p *PredictionTimeseriesDataset) CreateDataset(rootDataPath string, datasetName string, config *env.Config) (*serialization.RawDataset, error) {
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

	return &serialization.RawDataset{
		ID:       p.params.Dataset,
		Name:     p.params.Dataset,
		Data:     timeseriesData,
		Metadata: p.params.Meta,
	}, nil
}

// GetDefinitiveTypes returns an empty list as definitive types.
func (p *PredictionTimeseriesDataset) GetDefinitiveTypes() []*model.Variable {
	return []*model.Variable{}
}

// CleanupTempFiles does nothing.
func (p *PredictionTimeseriesDataset) CleanupTempFiles() {
}

// returns the index of the supplied variables that needs to be added to the data
func findMissingColumns(sourceHeaderNames []string, predictionHeaderNames []string) map[int]bool {
	result := map[int]bool{}
	predVarMap := map[string]bool{}
	for _, v := range predictionHeaderNames {
		predVarMap[strings.ToLower(v)] = true
	}
	for i, hn := range sourceHeaderNames {
		if !predVarMap[strings.ToLower(hn)] {
			result[i] = true
		}
	}
	return result
}

// CreateDataset creates a prediction dataset.
func (p *predictionDataset) CreateDataset(rootDataPath string, datasetName string, config *env.Config) (*serialization.RawDataset, error) {
	// need to do a bit of processing on the usual setup
	ds, err := p.params.DatasetConstructor.CreateDataset(rootDataPath, datasetName, config)
	if err != nil {
		return nil, err
	}
	predictionVariables := ds.Metadata.GetMainDataResource().Variables
	// updated the new dataset to match the var types and ordering of the source dataset - required
	// so that the model lines up
	variables := p.params.Meta.GetMainDataResource().Variables
	csvDataAugmented, err := augmentPredictionDataset(ds.Data, p.params.Target, variables, predictionVariables)
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

	return &serialization.RawDataset{
		ID:       p.params.Dataset,
		Name:     p.params.Dataset,
		Data:     csvDataAugmented,
		Metadata: p.params.Meta,
	}, nil
}

// GetDefinitiveTypes returns an empty list as definitive types.
func (p *predictionDataset) GetDefinitiveTypes() []*model.Variable {
	return []*model.Variable{}
}

// CleanupTempFiles calls the cleanup on the dataset constructor used.
func (p *predictionDataset) CleanupTempFiles() {
	p.params.DatasetConstructor.CleanupTempFiles()
}

// PredictParams contains all parameters passed to the predict function.
type PredictParams struct {
	Meta               *model.Metadata
	LearningDataMeta   *model.Metadata
	SourceDataset      *api.Dataset
	Dataset            string
	SchemaPath         string
	SourceDatasetID    string
	SolutionID         string
	FittedSolutionID   string
	DatasetConstructor DatasetConstructor
	OutputPath         string
	IndexFields        []string
	Target             *model.Variable
	MetaStorage        api.MetadataStorage
	DataStorage        api.DataStorage
	SolutionStorage    api.SolutionStorage
	ModelStorage       api.ExportedModelStorage
	IngestConfig       *IngestTaskConfig
	Config             *env.Config
}

func createPredictionDatasetID(existingDatasetID string, fittedSolutionID string) string {
	return fmt.Sprintf("%s-%s", existingDatasetID, fittedSolutionID)
}

// CloneDataset clones a dataset in metadata storage, data storage and on disk.
func CloneDataset(sourceDatasetID string, cloneDatasetID string, cloneFolder string,
	cloneLearningDataset bool, metaStorage api.MetadataStorage, dataStorage api.DataStorage) error {

	ds, err := metaStorage.FetchDataset(sourceDatasetID, false, false, false)
	if err != nil {
		return err
	}

	storageNameClone, err := dataStorage.GetStorageName(cloneDatasetID)
	if err != nil {
		return err
	}

	// clone metadata and data
	err = metaStorage.CloneDataset(sourceDatasetID, cloneDatasetID, storageNameClone, path.Base(cloneFolder))
	if err != nil {
		return err
	}

	err = dataStorage.CloneDataset(sourceDatasetID, ds.StorageName, cloneDatasetID, storageNameClone)
	if err != nil {
		return err
	}

	// clone dataset on disk
	dsDisk, err := api.LoadDiskDataset(ds)
	if err != nil {
		return err
	}
	dsDiskCloned, err := dsDisk.Clone(cloneFolder, cloneDatasetID, storageNameClone)
	if err != nil {
		return err
	}

	// update learning folder
	dsCloned, err := metaStorage.FetchDataset(cloneDatasetID, true, true, true)
	if err != nil {
		return err
	}
	dsCloned.LearningDataset = dsDiskCloned.GetLearningFolder()
	err = metaStorage.UpdateDataset(dsCloned)
	if err != nil {
		return err
	}

	return nil
}

// PrepExistingPredictionDataset sets up an existing dataset to be usable for predictions.
func PrepExistingPredictionDataset(params *PredictParams) (string, string, error) {
	// we need to clone the base dataset and use the clone for predictions
	// otherwise we would be updating the base dataset with new data
	cloneDatasetID := createPredictionDatasetID(params.Dataset, params.FittedSolutionID)
	dsCloned, err := params.MetaStorage.FetchDataset(cloneDatasetID, true, true, true)
	if err != nil {
		return "", "", err
	}
	if dsCloned != nil {
		// already cloned so assume everything is good to go
		return cloneDatasetID, dsCloned.GetLearningFolder(), nil
	}

	dsSource, err := params.MetaStorage.FetchDataset(params.Dataset, true, true, true)
	if err != nil {
		return "", "", err
	}
	targetFolder := fmt.Sprintf("%s-%s", env.ResolvePath(dsSource.Source, dsSource.Folder), params.FittedSolutionID)

	// clone the base dataset, then add the necessary fields
	log.Infof("cloning '%s' for predictions using '%s' as new id stored on disk at '%s'", params.Dataset, cloneDatasetID, targetFolder)
	err = CloneDataset(params.Dataset, cloneDatasetID, targetFolder, false, params.MetaStorage, params.DataStorage)
	if err != nil {
		return "", "", err
	}

	// pull the cloned dataset for updates
	dsCloned, err = params.MetaStorage.FetchDataset(cloneDatasetID, true, true, true)
	if err != nil {
		return "", "", err
	}

	// make sure the dataset on disk has the target variable!
	// (if performance here becomes an issue, only need data if adding target)
	dsDisk, err := api.LoadDiskDataset(dsCloned)
	if err != nil {
		return "", "", err
	}

	alignmentUpdates := getPredictionDatasetAlignmentUpdates(params.Meta.GetMainDataResource().Variables,
		dsDisk.Dataset.Metadata.GetMainDataResource().Variables, params.Target)
	err = updatePredictionAlignment(alignmentUpdates, dsCloned, dsDisk, params.MetaStorage, params.DataStorage)
	if err != nil {
		return "", "", err
	}

	// If there is learnign data present, we'll need to align the pre-featurized dataset as well
	if params.LearningDataMeta != nil {
		prefeaturizedAlignmentUpdates := getPredictionDatasetAlignmentUpdates(params.LearningDataMeta.GetMainDataResource().Variables,
			dsDisk.FeaturizedDataset.Dataset.Metadata.GetMainDataResource().Variables, params.Target)
		err = updatePredictionAlignment(prefeaturizedAlignmentUpdates, dsCloned, dsDisk.FeaturizedDataset, params.MetaStorage, params.DataStorage)
		if err != nil {
			return "", "", err
		}
	}

	// update the learning dataset
	learningFolder := dsDisk.GetLearningFolder()

	foundTarget := dsDisk.Dataset.FieldExists(params.Target)
	if !foundTarget {
		err = dsDisk.AddField(params.Target)
		if err != nil {
			return "", "", err
		}
		err = dsDisk.SaveDataset()
		if err != nil {
			return "", "", err
		}

		// need to append the target to the underlying data
		dsCloned.Variables = append(dsCloned.Variables, params.Target)

		// target field needs to exist in data storage as well
		err = params.DataStorage.AddVariable(dsCloned.ID, dsCloned.StorageName, params.Target.Key, params.Target.Type, "")
		if err != nil {
			return "", "", err
		}
	}
	dsCloned.Type = api.DatasetTypeInference
	err = params.MetaStorage.UpdateDataset(dsCloned)
	if err != nil {
		return "", "", err
	}

	return cloneDatasetID, learningFolder, nil
}

type alignmentUpdates struct {
	adds     []*model.Variable
	reorders map[int]int
}

func getPredictionDatasetAlignmentUpdates(modelVariables []*model.Variable, predictionVariables []*model.Variable,
	targetVariable *model.Variable) *alignmentUpdates {
	// The model may have variables that our input dataset is missing.  To maximize compatibility, we will create any missing
	// columns, set their type metadata appropriately, and leave them empty.  It may also be the case that the datasets have
	// their variables ordered differently, which is a problem because D3M pipelines are driven by column index rather than
	// name (we argued against this and lost).

	generatedFeatureIdx := len(predictionVariables)
	predictVars := map[string]*model.Variable{}
	for _, v := range predictionVariables {
		predictVars[v.Key] = v
		if v.DistilRole == model.VarDistilRoleSystemData && v.Index < generatedFeatureIdx {
			generatedFeatureIdx = v.Index
		}
	}

	modelVars := map[string]*model.Variable{}
	for _, v := range modelVariables {
		modelVars[v.Key] = v
	}

	// First find the variables in the model list that are not in the predict list and store them.
	missingList := []*model.Variable{}
	for _, v := range modelVariables {
		if _, ok := predictVars[v.Key]; !ok {
			missingList = append(missingList, v)
		}
	}

	// Insert the missing variables into to a temp list to use for determining re-ordering.  The insert happens
	// prior to any generated features.
	extPredictionVariables := make([]*model.Variable, len(predictionVariables)+len(missingList))
	copy(extPredictionVariables, predictionVariables[:generatedFeatureIdx])
	copy(extPredictionVariables[generatedFeatureIdx:], missingList)
	copy(extPredictionVariables[generatedFeatureIdx+len(missingList):], predictionVariables[generatedFeatureIdx:])

	// Next, make sure the order of columns in the prediction dataset match the expected model ordering, and
	// store a mapping for those that don't match.
	nextVarIdx := len(modelVariables)
	reorders := map[int]int{}
	for _, p := range extPredictionVariables {
		if m, ok := modelVars[p.Key]; ok {
			if m.Index != p.Index {
				reorders[p.Index] = m.Index
			}
		} else {
			reorders[p.Index] = nextVarIdx
			nextVarIdx++
		}
	}

	return &alignmentUpdates{missingList, reorders}
}

func updatePredictionAlignment(updates *alignmentUpdates, dataset *api.Dataset, diskDataset *api.DiskDataset,
	metaStorage api.MetadataStorage, dataStorage api.DataStorage) error {

	// align the CSV data for the base prediction dataset
	sourceData := diskDataset.Dataset.Data
	alignedData := make([][]string, len(sourceData))

	for i, row := range sourceData {
		// Create a new row that will include the original prediction data, along with the columns for data
		// that needs to be added to align with the dataset data.
		alignedData[i] = make([]string, len(row)+len(updates.adds))

		// First, add empty data for the missing variables at the locations the model expects them
		for _, addVar := range updates.adds {
			if i > 0 {
				alignedData[i][addVar.Index] = ""
			} else {
				alignedData[i][addVar.Index] = addVar.Key
			}
		}

		// Next, add the existing data, but re-map the column order so they line up with what the model expects
		for j, value := range row {
			if mappedIndex, ok := updates.reorders[j]; ok {
				alignedData[i][mappedIndex] = value
			} else {
				alignedData[i][j] = value
			}
		}
	}
	// update the disk dataset structure with aligned data
	diskDataset.Dataset.Data = alignedData

	if len(updates.adds) > 0 {
		// update the in-memory dataset variables to reflect the new additions
		dataset.Variables = append(dataset.Variables, updates.adds...)

		// update the disk dataset variables to reflect the new additions
		diskDataset.Dataset.Metadata.GetMainDataResource().Variables = append(diskDataset.Dataset.Metadata.GetMainDataResource().Variables, updates.adds...)

		// update data storage with the added variables
		for _, addVariable := range updates.adds {
			addVariable.DistilRole = model.VarDistilRolePadding
			err := dataStorage.AddVariable(dataset.ID, dataset.StorageName, addVariable.Key, addVariable.Type, "")
			if err != nil {
				return err
			}
		}
	}

	if len(updates.reorders) > 0 {
		// update the in-memory dataset and disk-based dataset variables to reflect any re-ordering that takes place
		updateVariableIndices(diskDataset, dataset.Variables)
		updateVariableIndices(diskDataset, diskDataset.Dataset.Metadata.GetMainDataResource().Variables)
	}

	// save the updates to disk - should hopefully just work
	err := diskDataset.SaveDataset()
	if err != nil {
		return err
	}

	// save the updates to the metadata store
	err = metaStorage.UpdateDataset(dataset)
	if err != nil {
		return err
	}

	return nil
}

func updateVariableIndices(diskDataset *api.DiskDataset, variables []*model.Variable) {
	headerRow := diskDataset.Dataset.Data[0]
	varIndices := map[string]int{}
	for i, headerValue := range headerRow {
		varIndices[headerValue] = i
	}
	for _, variable := range variables {
		if idx, ok := varIndices[variable.HeaderName]; ok {
			variable.Index = idx
		} else {
			variable.Index = -1
		}
	}
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
		// only the first n fields need to match, with n being the number of fields in the source dataset
		if i < len(mainDR.Variables) && mainDR.Variables[i].Key != f && mainDR.Variables[i].HeaderName != f {
			return "", "", errors.Errorf("variables in new prediction file do not match variables in original dataset")
		}
	}
	log.Infof("dataset fields compatible with original dataset fields")

	// read the metadata from the created prediction dataset since it needs to be updated
	datasetStorage := serialization.GetStorage(rawDataPath)
	schemaPath = path.Join(datasetPath, compute.D3MDataSchema)
	meta, err = datasetStorage.ReadMetadata(schemaPath)
	if err != nil {
		return "", "", errors.Wrap(err, "unable to read metadata")
	}

	// update the dataset doc to reflect original types
	meta.ID = datasetName
	meta.Name = datasetName
	meta.StorageName = model.NormalizeDatasetID(datasetName)
	meta.DatasetFolder = path.Base(datasetPath)
	variables := updateMetaDataTypes(params.SolutionStorage, params.MetaStorage, params.DataStorage, meta, params.FittedSolutionID, params.Dataset, meta.StorageName)
	if err != nil {
		return "", "", errors.Wrap(err, "unable to update metadata types")
	}
	err = datasetStorage.WriteMetadata(schemaPath, meta, true, false)
	if err != nil {
		return "", "", errors.Wrap(err, "unable to update dataset doc")
	}
	log.Infof("wrote out schema doc for new dataset with id '%s' at location '%s'", meta.ID, schemaPath)
	err = createClassification(params, datasetPath, variables)
	if err != nil {
		return "", "", errors.Wrap(err, "unable to create classification")
	}
	params.Meta = meta
	return datasetName, schemaPath, nil
}

// IngestPredictionDataset ingests a dataset to be used for predictions.
func IngestPredictionDataset(params *PredictParams) error {
	schemaPath := params.SchemaPath
	// ingest the dataset but without running simon, duke, etc.
	steps := &IngestSteps{
		VerifyMetadata:       true,
		FallbackMerged:       false,
		CreateMetadataTables: false,
	}
	ingestParams := &IngestParams{
		Source: metadata.Augmented,
	}
	err := IngestPostgres(schemaPath, schemaPath, ingestParams, params.IngestConfig, steps)
	if err != nil {
		return errors.Wrap(err, "unable to ingest prediction data")
	}
	log.Infof("finished ingesting dataset '%s'", params.Dataset)
	rawDataPath := path.Join(params.Meta.DatasetFolder, compute.D3MDataFolder, compute.D3MLearningData)
	datasetStorage := serialization.GetStorage(rawDataPath)
	err = datasetStorage.WriteMetadata(params.SchemaPath, params.Meta, true, false)
	if err != nil {
		return err
	}
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
		varExists, err := params.MetaStorage.DoesVariableExist(params.Dataset, params.Target.Key)
		if err != nil {
			return err
		}
		if !varExists {
			err = params.MetaStorage.AddGroupedVariable(params.Dataset, params.Target.Key, params.Target.DisplayName,
				params.Target.Type, params.Target.DistilRole, tsg)
			if err != nil {
				return err
			}
		}
		log.Infof("done creating compose variables")
	}

	// add feature groups
	err = copyFeatureGroups(params.FittedSolutionID, metaClone.ID, params.SolutionStorage, params.MetaStorage)
	if err != nil {
		return err
	}
	err = params.DataStorage.CreateIndices(params.Dataset, params.IndexFields)
	if err != nil {
		return err
	}
	return nil
}

// Predict processes input data to generate predictions.
func Predict(params *PredictParams) (string, error) {
	log.Infof("generating predictions for fitted solution ID %s found at '%s'", params.FittedSolutionID, params.SchemaPath)
	schemaPath := params.SchemaPath
	datasetName := params.Dataset

	// the dataset id needs to match the original dataset id for TA2 to be able to use the model
	// read from source in case any step has updated it along the way
	meta, err := metadata.LoadMetadataFromOriginalSchema(schemaPath, false)
	if err != nil {
		return "", errors.Wrap(err, "unable to read latest dataset doc")
	}
	meta.ID = params.SourceDatasetID
	datasetStorage := serialization.GetStorage(meta.GetMainDataResource().ResPath)
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
	log.Infof("generating predictions using data found at '%s'", params.SchemaPath)
	predictionResult, err := comp.GeneratePredictions(params.SchemaPath, solution.SolutionID, params.FittedSolutionID, client)
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
	log.Infof("persisting prediction results for %s using storage name %s", datasetName, meta.StorageName)
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

func augmentPredictionDataset(csvData [][]string, target *model.Variable,
	sourceVariables []*model.Variable, predictionVariables []*model.Variable) ([][]string, error) {
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
	addTarget := true
	predictVariablesMap := make(map[int]int)
	isTimeseries := model.IsTimeSeries(target.Type)
	newPredictionFields := map[int]*model.Variable{}
	// If the variable list for prediction set is empty (as is the case for tabular data) then we just use the
	// header values as the list of variable names to build the map.
	if len(predictionVariables) == 0 {
		for i, pv := range csvData[0] {
			varName := strings.ToLower(pv)
			if sourceVariableHeaderMap[varName] != nil {
				predictVariablesMap[i] = sourceVariableHeaderMap[varName].Index
				log.Infof("mapped '%s' to index %d", varName, predictVariablesMap[i])
				if sourceVariableHeaderMap[varName].Key == target.Key {
					addTarget = false
				}
			} else {
				newFieldIndex := len(sourceVariables) + len(newPredictionFields)
				log.Infof("new prediction field '%s' found and mapped to index %d", pv, newFieldIndex)
				predictVariablesMap[i] = -1
				newPredictionFields[i] = model.NewVariable(newFieldIndex, pv, pv, pv, pv, model.StringType,
					model.StringType, "", []string{model.RoleAttribute}, model.VarDistilRoleData, nil, sourceVariables, true)
			}
			if pv == model.D3MIndexFieldName {
				addIndex = false
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
				log.Warnf("field '%s' not found in source dataset - column will be appended to the dataset", predictVariable.Key)
				predictVariablesMap[i] = -1
				newPredictionFields[i] = predictVariable
			}
			if predictVariable.Key == model.D3MIndexFieldName {
				addIndex = false
			} else if predictVariable.Key == target.Key {
				addTarget = false
			}
		}
	}

	// add target if it isnt part of prediction dataset
	if addTarget && !isTimeseries {
		// some how the index from source to es can be off?
		// so find the source variable and use it instead
		index := target.Index
		for _, v := range sourceVariables {
			if v.Key == target.Key {
				index = v.Index
			}
		}
		predictVariablesMap[len(csvData[0])] = index
	}

	// add the new prediction fields
	maxIndex := len(sourceVariables)
	for _, v := range newPredictionFields {
		v.Index = maxIndex
		headerSource = append(headerSource, v.HeaderName)
		maxIndex++
	}

	// read the rest of the data
	log.Infof("rewriting prediction dataset to match source dataset structure")
	count := 0

	// read the d3m field index if present
	d3mFieldIndex := -1
	if variable, ok := sourceVariableMap[strings.ToLower(model.D3MIndexFieldName)]; ok {
		d3mFieldIndex = variable.Index
	}
	missingColumns := findMissingColumns(headerSource, csvData[0])
	outputData := [][]string{headerSource}
	for _, line := range csvData[1:] {
		offset := 0
		// write the columns in the same order as the source dataset
		output := make([]string, maxIndex)
		i := 0
		for range predictVariablesMap {
			sourceIndex := predictVariablesMap[i]
			if sourceIndex >= 0 {
				// need to handle the case where the field does not exist in source so needs to be appened
				// the variable in newPredictionFields has the correct index to use
				if _, ok := missingColumns[predictVariablesMap[i]]; !ok {
					output[sourceIndex] = line[i-offset]
				} else {
					output[sourceIndex] = ""
					offset++
				}
			} else {
				// new var so append accordingly
				newVar := newPredictionFields[i]
				output[newVar.Index] = line[i-offset]
			}
			i++
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
	}
	varExists, err = dataStorage.DoesVariableExist(dataset, storageName, composedVarName)
	if err != nil {
		return err
	}
	if !varExists {
		err = dataStorage.AddVariable(dataset, storageName, composedVarName, model.StringType, "")
		if err != nil {
			return err
		}
	}
	composedData := map[string]string{}
	// No grouping column - just use the d3mIndex as we'll just stick some placeholder
	// data in.
	filter := &api.FilterParams{
		Variables: []string{model.D3MIndexFieldName},
	}
	// Fetch data using the source names as the filter
	filter.Variables = append(filter.Variables, sourceVarNames...)
	rawData, err := dataStorage.FetchData(dataset, storageName, filter, false, nil)
	if err != nil {
		return err
	}

	// Create a map of the retreived fields to column number.  Store d3mIndex since it needs to be directly referenced
	// further along.
	d3mIndexFieldindex := rawData.Columns[model.D3MIndexFieldName].Index

	if len(sourceVarNames) > 0 {
		// Loop over the fetched data, composing each column value into a single new column value using the
		// separator.
		for _, r := range rawData.Values {
			// create the hash from the specified columns
			composed := createComposedFields(r, sourceVarNames, rawData.Columns, DefaultSeparator)
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

func createComposedFields(data []*api.FilteredDataValue, fields []string, mappedFields map[string]*api.Column, separator string) string {
	dataToJoin := make([]string, len(fields))
	for i, field := range fields {
		dataToJoin[i] = fmt.Sprintf("%v", data[mappedFields[field].Index].Value)
	}
	return strings.Join(dataToJoin, separator)
}

func createClassification(params *PredictParams, datasetPath string, variables []*model.Variable) error {
	outputPath := path.Join(datasetPath, params.Config.ClassificationOutputPath)
	log.Info("writing predicted dataset type classification to file")
	classification := buildClassificationFromMetadata(variables)
	classification.Path = outputPath
	err := metadata.WriteClassification(classification, outputPath)
	if err != nil {
		return err
	}
	return nil
}

func updateMetaDataTypes(solutionStorage api.SolutionStorage, metaStorage api.MetadataStorage,
	dataStorage api.DataStorage, meta *model.Metadata, fittedSolutionID string, dataset string, storageName string) []*model.Variable {
	variables := meta.GetMainDataResource().Variables
	solutionRequest, err := solutionStorage.FetchRequestByFittedSolutionID(fittedSolutionID)
	if err != nil {
		return nil
	}
	// get a variable map for quick look up
	trainVariables, err := metaStorage.FetchVariables(solutionRequest.Dataset, true, true, false)
	if err != nil {
		return nil
	}
	varMap := map[string]*model.Variable{}
	for _, v := range trainVariables {
		varMap[v.Key] = v
	}
	result := []*model.Variable{}
	for _, v := range variables {
		tmp := varMap[v.Key]
		if tmp != nil {
			tmp.Index = v.Index
			result = append(result, tmp)
		}
	}
	return result
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
	variableMap := api.MapVariables(variables, func(variable *model.Variable) string { return variable.Key })
	variablesPrediction, err := metaStorage.FetchVariables(datasetName, false, true, false)
	if err != nil {
		return err
	}
	variablePredictionMap := api.MapVariables(variablesPrediction, func(variable *model.Variable) string { return variable.Key })

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
