package importer

import (
	"strings"

	"github.com/pkg/errors"
	log "github.com/unchartedsoftware/plog"

	"github.com/uncharted-distil/distil-compute/metadata"
	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil/api/env"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/task"
	"github.com/uncharted-distil/distil/api/util"
	"github.com/uncharted-distil/distil/api/util/json"
)

// Joined can be used to import datasets that have been joined.
type Joined struct {
	sourcePath            string
	datasetID             string
	config                *env.Config
	joinedDataset         map[string]interface{}
	originalDataset       map[string]interface{}
	leftCols              []string
	rightCols             []string
	sourceLearningDataset string
	updateDatasetID       string
	meta                  api.MetadataStorage
}

// NewJoined creates a new importer for joined datasets.
func NewJoined(meta api.MetadataStorage, config *env.Config) Importer {
	return &Joined{
		meta:   meta,
		config: config,
	}
}

// Initialize sets up the importer.
func (j *Joined) Initialize(params map[string]interface{}, ingestParams *task.IngestParams) error {
	if params == nil {
		return errors.Errorf("no parameters specified")
	}

	if params["path"] == nil {
		return errors.Errorf("missing 'path' parameter")
	}

	if params["joinedDataset"] == nil {
		return errors.Errorf("missing 'joinedDataset' parameter")
	}

	if params["originalDataset"] == nil {
		return errors.Errorf("missing 'originalDataset' parameter")
	}

	j.sourcePath = params["path"].(string)

	originalDataset, ok := params["originalDataset"].(map[string]interface{})
	if !ok {
		return errors.Errorf("unable to parse original dataset")
	}
	j.originalDataset = originalDataset
	joinedDataset, ok := params["joinedDataset"].(map[string]interface{})
	if !ok {
		return errors.Errorf("unable to parse joined dataset")
	}
	j.joinedDataset = joinedDataset

	j.leftCols, ok = json.StringArray(params, "leftCols")
	if !ok {
		return errors.Errorf("unable to parse left cols")
	}
	j.rightCols, ok = json.StringArray(params, "rightCols")
	if !ok {
		return errors.Errorf("unable to parse right cols")
	}

	j.datasetID = ingestParams.ID

	return nil
}

// PrepareImport prepares the dataset for import.
func (j *Joined) PrepareImport() (*task.IngestSteps, *task.IngestParams, error) {
	ingestSteps := &task.IngestSteps{
		ClassificationOverwrite: false,
		VerifyMetadata:          true,
		FallbackMerged:          true,
		CheckMatch:              true,
		SkipFeaturization:       false,
	}

	// updateDatasetID is the ID of the dataset to use to sync a prefeaturized
	// dataset. Note that it will be one of two datasets joined, whichever is NOT
	// prefeaturized. This is used instead of using the joined dataset because we do
	// NOT want to sync the data that is already prefeaturized!

	// make sure only one of the datasets has a prefeaturized version
	originalLearningDataset := j.originalDataset["learningDataset"].(string)
	joinedLearningDataset := j.joinedDataset["learningDataset"].(string)
	if originalLearningDataset != "" && joinedLearningDataset != "" {
		return nil, nil, errors.Errorf("joining datasets that both have learning datasets is not supported")
	} else if originalLearningDataset != "" {
		ingestSteps.SkipFeaturization = true
		j.sourceLearningDataset = originalLearningDataset
		j.updateDatasetID = strings.Join([]string{j.originalDataset["id"].(string), j.joinedDataset["id"].(string)}, "-")
	} else if joinedLearningDataset != "" {
		ingestSteps.SkipFeaturization = true
		j.sourceLearningDataset = joinedLearningDataset
		j.updateDatasetID = strings.Join([]string{j.originalDataset["id"].(string), j.joinedDataset["id"].(string)}, "-")
	}

	log.Infof("creating joined dataset '%s' from '%s'", j.datasetID, j.sourcePath)
	creationResult, err := createDataset(j.sourcePath, j.datasetID, j.config)
	if err != nil {
		return nil, nil, err
	}
	ingestParams := &task.IngestParams{
		ID:           creationResult.name,
		Path:         creationResult.path,
		RawGroupings: creationResult.groups,
		IndexFields:  creationResult.indexFields,
		Source:       metadata.Augmented,
	}
	ingestSteps.VerifyMetadata = false

	// combine the origin and joined dateset into an array of structs
	origins, err := getOriginsFromMaps(j.originalDataset, j.joinedDataset)
	if err != nil {
		return nil, nil, err
	}
	ingestParams.Origins = origins

	// if no groups were created, copy them from the passed in datasets
	if len(ingestParams.RawGroupings) == 0 {
		log.Infof("copying groupings from source datasets")
		groups, err := getDatasetGroups(j.originalDataset)
		if err != nil {
			return nil, nil, err
		}
		groupsJoin, err := getDatasetGroups(j.joinedDataset)
		if err != nil {
			return nil, nil, err
		}
		ingestParams.RawGroupings = combineDatasetGroupings(groups, groupsJoin)

		// final step - on join we drop the right column, which may be referred to by a group
		// we replace any refs to the right col with a ref to the left col
		nameUpdates := make([]nameUpdate, len(j.rightCols))
		for rightIdx, rightVar := range j.rightCols {
			nameUpdates[rightIdx] = nameUpdate{old: rightVar, new: j.leftCols[rightIdx]}
		}
		ingestParams.RawGroupings = remapDatasetGroups(nameUpdates, ingestParams.RawGroupings)
	}

	// set the definitive types based on the currently stored metadata
	definitiveVars := append(
		getVariablesDefault(j.originalDataset["id"].(string), j.meta),
		getVariablesDefault(j.joinedDataset["id"].(string), j.meta)...,
	)
	ingestParams.DefinitiveTypes = api.MapVariables(definitiveVars, func(variable *model.Variable) string { return variable.Key })

	log.Infof("Created dataset '%s' from local source '%s'", ingestParams.ID, ingestParams.Path)
	return ingestSteps, ingestParams, nil
}

// CleanupImport removes temporary files and structures created during the import.
func (j *Joined) CleanupImport(ingestResult *task.IngestResult) error {
	// if there is a source learning dataset, then sync it properly with the newly imported dataset
	err := syncPrefeaturizedDataset(ingestResult.DatasetID, j.updateDatasetID, j.sourceLearningDataset, j.meta)
	if err != nil {
		return err
	}

	util.Delete(j.sourcePath)

	return nil
}

// Given a raw JSON group definition, find any references to the "old" variable in the update list,
// and replace it with the "new" value.  This is mainly needed for joins, where the right hand join columns
// get removed, but may part of a group in the joined data.
func remapDatasetGroups(updates []nameUpdate, groups []map[string]interface{}) []map[string]interface{} {
	remappedGroups := []map[string]interface{}{}
	for _, g := range groups {
		rg := map[string]interface{}{}
		for groupField, value := range g {
			for _, update := range updates {
				if value == update.old {
					rg[groupField] = update.new
				} else {
					rg[groupField] = value
				}
			}
		}
		remappedGroups = append(remappedGroups, rg)
	}
	return remappedGroups
}

func getOriginsFromMaps(originalDataset map[string]interface{}, joinedDataset map[string]interface{}) ([]*model.DatasetOrigin, error) {
	joinSuggestions := make([]interface{}, 0)
	if originalDataset["joinSuggestion"] != nil {
		joinSuggestions = append(joinSuggestions, originalDataset["joinSuggestion"].([]interface{})...)
	}
	if joinedDataset["joinSuggestion"] != nil {
		joinSuggestions = append(joinSuggestions, joinedDataset["joinSuggestion"].([]interface{})...)
	}

	origins := make([]*model.DatasetOrigin, len(joinSuggestions))
	for i, js := range joinSuggestions {
		targetOriginModel := model.DatasetOrigin{}
		targetJoin := js.(map[string]interface{})
		targetJoinOrigin := targetJoin["datasetOrigin"].(map[string]interface{})
		err := json.MapToStruct(&targetOriginModel, targetJoinOrigin)
		if err != nil {
			return nil, err
		}
		origins[i] = &targetOriginModel
	}

	return origins, nil
}

func getDatasetGroups(dsRaw map[string]interface{}) ([]map[string]interface{}, error) {
	// cycle through variables, pulling groups out as they are found
	groups := []map[string]interface{}{}
	for _, v := range dsRaw["variables"].([]interface{}) {
		rawVariables := v.(map[string]interface{})
		if rawVariables["grouping"] != nil {
			groups = append(groups, rawVariables["grouping"].(map[string]interface{}))
		}
	}

	return groups, nil
}

func combineDatasetGroupings(groupingsFirst []map[string]interface{}, groupingsSecond []map[string]interface{}) []map[string]interface{} {
	// keep the first set of groupings, and add the second set provided that type of group does not already exist
	combined := append([]map[string]interface{}{}, groupingsFirst...)
	groupingTypes := map[string]bool{}
	for _, g := range groupingsFirst {
		groupingTypes[g["type"].(string)] = true
	}

	for _, g := range groupingsSecond {
		if !groupingTypes[g["type"].(string)] {
			combined = append(combined, g)
		}
	}

	return combined
}

func syncPrefeaturizedDataset(datasetID string, updateDatasetID string, sourceLearningDataset string, metaStorage api.MetadataStorage) error {
	// make sure there is a prefeaturized dataset to sync
	if sourceLearningDataset == "" {
		return nil
	}
	log.Infof("syncing prefeaturized dataset '%s' using prefeaturized data found at '%s' and updating with data from dataset '%s'",
		datasetID, sourceLearningDataset, updateDatasetID)

	ds, err := metaStorage.FetchDataset(datasetID, true, true, true)
	if err != nil {
		return err
	}
	// this is the join folder created before this ingest task
	dsUpdateLearningFolder := env.ResolvePath(metadata.Augmented, updateDatasetID)
	// TODO: CHECK UNIQUENESS!!!!
	joinedLearningDataset := task.CreateFeaturizedDatasetID(datasetID)
	joinedLearningDataset = env.ResolvePath(ds.Source, joinedLearningDataset)
	// copy prefeaturized data from the original source to the joined folder
	dsDisk, err := task.CopyDiskDataset(sourceLearningDataset, joinedLearningDataset, ds.ID, ds.StorageName)
	if err != nil {
		return err
	}
	log.Infof("copied prefeaturized data to '%s'", joinedLearningDataset)
	// load the recently copied prefeaturized dataset from the original to join as the join's FeaturizedDataset
	dsDisk.FeaturizedDataset, err = api.LoadDiskDatasetFromFolder(joinedLearningDataset)
	if err != nil {
		return err
	}
	// sync the dataset on disk
	// read the unfeaturized dataset from disk
	dsDiskUpdate, err := api.LoadDiskDatasetFromFolder(dsUpdateLearningFolder)
	if err != nil {
		return err
	}

	// update the featurized dataset using the unfeaturized data
	log.Infof("updating dataset on disk using data from '%s'", dsUpdateLearningFolder)
	// do not filter dataset by updates
	err = dsDisk.UpdateOnDisk(ds, dsDiskUpdate.Dataset.Data, true, false)
	if err != nil {
		return err
	}

	// update the metadata in ES
	ds.LearningDataset = joinedLearningDataset
	err = metaStorage.UpdateDataset(ds)
	if err != nil {
		return err
	}

	log.Infof("done syncing prefeaturized dataset '%s' on disk", datasetID)

	return nil
}

func getVariablesDefault(datasetID string, metaStorage api.MetadataStorage) []*model.Variable {
	ds, err := metaStorage.FetchDataset(datasetID, true, true, true)
	if err != nil {
		log.Infof("unable to fetch variables so defaulting to empty list")
		return []*model.Variable{}
	}

	return ds.Variables
}

type nameUpdate struct {
	old string
	new string
}
