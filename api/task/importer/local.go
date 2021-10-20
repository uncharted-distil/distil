package importer

import (
	"io/ioutil"
	"path"

	"github.com/pkg/errors"
	log "github.com/unchartedsoftware/plog"

	"github.com/uncharted-distil/distil-compute/metadata"
	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil/api/dataset"
	"github.com/uncharted-distil/distil/api/env"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/task"
	"github.com/uncharted-distil/distil/api/util"
)

// Local can be used to import datasets from the local filesystem.
type Local struct {
	sourcePath         string
	datasetID          string
	datasetDescription string
	config             *env.Config
	sample             bool
}

// NewLocal creates an importer for local datasets.
func NewLocal(config *env.Config) Importer {
	return &Local{
		config: config,
	}
}

// Initialize sets up the importer.
func (l *Local) Initialize(params map[string]interface{}, ingestParams *task.IngestParams) error {
	if params == nil {
		return errors.Errorf("no parameters specified")
	}

	if params["path"] == nil {
		return errors.Errorf("missing 'path' parameter")
	}

	// Check if we want to sample the dataset
	l.sample = true
	if params["nosample"] != nil {
		l.sample = false
	}

	if params["description"] != nil {
		l.datasetDescription = params["description"].(string)
	}

	l.sourcePath = params["path"].(string)
	l.datasetID = ingestParams.ID

	return nil
}

// PrepareImport prepares the dataset for import.
func (l *Local) PrepareImport() (*task.IngestSteps, *task.IngestParams, error) {
	ingestSteps := &task.IngestSteps{
		ClassificationOverwrite: false,
		VerifyMetadata:          false,
		FallbackMerged:          true,
		CheckMatch:              true,
		SkipFeaturization:       false,
	}

	log.Infof("Creating dataset '%s' from '%s'", l.datasetID, l.sourcePath)
	creationResult, err := createDataset(l.sourcePath, l.datasetID, l.config)
	if err != nil {
		return nil, nil, err
	}
	ingestParams := &task.IngestParams{
		ID:              creationResult.name,
		Path:            creationResult.path,
		RawGroupings:    creationResult.groups,
		IndexFields:     creationResult.indexFields,
		DefinitiveTypes: api.MapVariables(creationResult.definitiveVars, func(variable *model.Variable) string { return variable.Key }),
		Source:          metadata.Augmented,
	}

	log.Infof("Created dataset '%s' from local source '%s'", ingestParams.ID, ingestParams.Path)

	return ingestSteps, ingestParams, nil
}

// Import imports the dataset using the dataset importer.
func (l *Local) Import() error {
	return nil
}

// CleanupImport removes temporary files and structures created during the import.
func (l *Local) CleanupImport(ingestResult *task.IngestResult) error {
	if !util.IsInDirectory(env.GetPublicPath(), l.sourcePath) {
		util.Delete(l.sourcePath)
	}

	return nil
}

func rawDatasetIsTabular(datasetPath string) bool {
	// check if datasetPath is a folder
	if !util.FileExists(datasetPath) || util.IsDirectory(datasetPath) {
		return false
	}

	// check if it is a csv file
	return path.Ext(datasetPath) == ".csv"
}

func createTableDataset(datasetPath string, datasetName string) (task.DatasetConstructor, error) {
	data, err := ioutil.ReadFile(datasetPath)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to read raw tabular data")
	}

	ds, err := dataset.NewTableDataset(datasetName, data, true)
	if err != nil {
		return nil, err
	}

	return ds, nil
}

func createDataset(datasetPath string, datasetName string, config *env.Config) (*creationResult, error) {
	// create the dataset constructors for downstream processing
	var ds task.DatasetConstructor
	var err error
	var groups []map[string]interface{}
	var indexFields []string
	if util.IsDatasetDir(datasetPath) {
		ds, err = dataset.NewD3MDataset(datasetName, datasetPath)
	} else if rawDatasetIsTabular(datasetPath) {
		ds, err = createTableDataset(datasetPath, datasetName)
	} else {
		// check to see what type of files it contains
		var fileType string
		fileType, err = dataset.CheckFileType(datasetPath)
		if err != nil {
			return nil, err
		}
		if fileType == "png" || fileType == "jpeg" || fileType == "jpg" {
			ds, err = dataset.NewMediaDatasetFromExpanded(datasetName, fileType, "jpeg", "", datasetPath)
		} else if fileType == "tif" {
			ds, err = dataset.NewSatelliteDatasetFromExpanded(datasetName, fileType, "", datasetPath)
			groups = []map[string]interface{}{dataset.CreateSatelliteGrouping(), dataset.CreateGeoBoundsGrouping()}
			indexFields = dataset.GetSatelliteIndexFields()
		} else if fileType == "txt" {
			ds, err = dataset.NewMediaDatasetFromExpanded(datasetName, "txt", "txt", "", datasetPath)
		} else {
			err = errors.Errorf("unsupported archived file type %s", fileType)
		}
	}
	if err != nil {
		return nil, err
	}

	// create the formatted d3m dataset
	outputPath := path.Join(config.D3MOutputDir, config.AugmentedSubFolder)
	log.Infof("Creating final dataset '%s' from '%s'", datasetName, outputPath)
	datasetName, formattedPath, err := task.CreateDataset(datasetName, ds, outputPath, config)
	if err != nil {
		return nil, err
	}

	return &creationResult{
		name:           datasetName,
		path:           formattedPath,
		groups:         groups,
		indexFields:    indexFields,
		definitiveVars: ds.GetDefinitiveTypes(),
	}, nil
}

type creationResult struct {
	name           string
	path           string
	groups         []map[string]interface{}
	indexFields    []string
	definitiveVars []*model.Variable
}
