package importer

import (
	"fmt"
	"path"
	"strings"

	"github.com/pkg/errors"
	log "github.com/unchartedsoftware/plog"

	"github.com/uncharted-distil/distil-compute/metadata"
	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil/api/env"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/task"
	"github.com/uncharted-distil/distil/api/util"
)

// Union can be used to import datasets that have been combined using the union operation.
type Union struct {
	sourcePath            string
	datasetID             string
	config                *env.Config
	joinedDataset         map[string]interface{}
	originalDataset       map[string]interface{}
	sourceLearningDataset string
	updateDatasetID       string
	meta                  api.MetadataStorage
}

// NewUnion creates a new importer for union datasets.
func NewUnion(meta api.MetadataStorage, config *env.Config) Importer {
	return &Union{
		meta:   meta,
		config: config,
	}
}

// Initialize sets up the importer.
func (u *Union) Initialize(params map[string]interface{}, ingestParams *task.IngestParams) error {
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

	u.sourcePath = params["path"].(string)

	originalDataset, ok := params["originalDataset"].(map[string]interface{})
	if !ok {
		return errors.Errorf("unable to parse original dataset")
	}
	u.originalDataset = originalDataset
	joinedDataset, ok := params["joinedDataset"].(map[string]interface{})
	if !ok {
		return errors.Errorf("unable to parse joined dataset")
	}
	u.joinedDataset = joinedDataset

	u.datasetID = ingestParams.ID

	return nil
}

// PrepareImport prepares the dataset for import.
func (u *Union) PrepareImport() (*task.IngestSteps, *task.IngestParams, error) {
	ingestSteps := &task.IngestSteps{
		ClassificationOverwrite: false,
		VerifyMetadata:          true,
		FallbackMerged:          true,
		CheckMatch:              true,
		SkipFeaturization:       false,
	}

	// make sure either neither dataset is prefeaturized or both are prefeaturized
	// NOTE: THAT IS ONE BIG ASSUMPTION ON THE LEARNING DATASET NAME!
	originalLearningDataset := u.originalDataset["learningDataset"].(string)
	joinedLearningDataset := u.joinedDataset["learningDataset"].(string)
	if originalLearningDataset != "" && joinedLearningDataset != "" {
		ingestSteps.SkipFeaturization = true
		u.sourceLearningDataset = fmt.Sprintf("%s-union-%s", path.Base(originalLearningDataset), path.Base(joinedLearningDataset))
		u.updateDatasetID = strings.Join([]string{u.originalDataset["id"].(string), u.joinedDataset["id"].(string)}, "-")
	} else if originalLearningDataset != "" {
		return nil, nil, errors.Errorf("both the original and joining datasets need to be prefeaturized")
	} else if joinedLearningDataset != "" {
		return nil, nil, errors.Errorf("both the original and joining datasets need to be prefeaturized")
	}

	log.Infof("creating union dataset '%s' from '%s'", u.datasetID, u.sourcePath)
	creationResult, err := createDataset(u.sourcePath, u.datasetID, u.config)
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
	origins, err := getOriginsFromMaps(u.originalDataset, u.joinedDataset)
	if err != nil {
		return nil, nil, err
	}
	ingestParams.Origins = origins

	// if no groups were created, copy them from the passed in datasets
	if len(ingestParams.RawGroupings) == 0 {
		log.Infof("copying groupings from source datasets")
		groups, err := getDatasetGroups(u.originalDataset)
		if err != nil {
			return nil, nil, err
		}
		groupsJoin, err := getDatasetGroups(u.joinedDataset)
		if err != nil {
			return nil, nil, err
		}
		ingestParams.RawGroupings = combineDatasetGroupings(groups, groupsJoin)
	}

	// set the definitive types based on the currently stored metadata
	definitiveVars := append(
		getVariablesDefault(u.originalDataset["id"].(string), u.meta),
		getVariablesDefault(u.joinedDataset["id"].(string), u.meta)...,
	)
	ingestParams.DefinitiveTypes = api.MapVariables(definitiveVars, func(variable *model.Variable) string { return variable.Key })

	log.Infof("Created dataset '%s' from local source '%s'", ingestParams.ID, ingestParams.Path)
	return ingestSteps, ingestParams, nil
}

// CleanupImport removes temporary files and structures created during the import.
func (u *Union) CleanupImport(ingestResult *task.IngestResult) error {
	// update dataset to set learning data
	ds, err := u.meta.FetchDataset(u.datasetID, true, true, true)
	if err != nil {
		return err
	}
	ds.LearningDataset = u.sourceLearningDataset
	err = u.meta.UpdateDataset(ds)
	if err != nil {
		return err
	}

	util.Delete(u.sourcePath)
	return nil
}
