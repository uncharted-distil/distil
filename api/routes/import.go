package routes

import (
	"fmt"
	"net/http"
	"path"

	"github.com/pkg/errors"
	"goji.io/pat"

	"github.com/unchartedsoftware/distil-ingest/metadata"
	"github.com/unchartedsoftware/distil/api/env"
	"github.com/unchartedsoftware/distil/api/model"
	"github.com/unchartedsoftware/distil/api/task"
	"github.com/unchartedsoftware/distil/api/util"
)

// ImportHandler imports a dataset to the local file system and then ingests it.
func ImportHandler(datamartMetaCtor model.MetadataStorageCtor, fileMetaCtor model.MetadataStorageCtor, esMetaCtor model.MetadataStorageCtor, config *task.IngestTaskConfig) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		datasetID := pat.Param(r, "datasetID")
		source := metadata.DatasetSource(pat.Param(r, "source"))

		// update ingest config to use ingest URI.
		cfg, err := env.LoadConfig()
		if err != nil {
			handleError(w, err)
			return
		}

		meta, err := createMetadataStorageForSource(source, datamartMetaCtor, fileMetaCtor, esMetaCtor)
		if err != nil {
			handleError(w, err)
			return
		}

		// import the dataset to the local filesystem.
		resolver := createResolverForSource(source, datasetID, &cfg, config)
		uri := resolver.ResolveInputAbsolute(datasetID)

		ingestConfig := *config
		ingestConfig.Resolver = resolver

		_, err = meta.ImportDataset(datasetID, uri)
		if err != nil {
			handleError(w, err)
			return
		}

		// ingest the imported dataset
		err = task.IngestDataset(datamartMetaCtor, cfg.ESDatasetsIndex, datasetID, source, &ingestConfig)
		if err != nil {
			handleError(w, err)
			return
		}

		// marshal data and sent the response back
		err = handleJSON(w, map[string]interface{}{"result": "ingested"})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal result histogram into JSON"))
			return
		}
	}
}

func createMetadataStorageForSource(datasetSource metadata.DatasetSource, datamartMetaCtor model.MetadataStorageCtor, fileMetaCtor model.MetadataStorageCtor, esMetaCtor model.MetadataStorageCtor) (model.MetadataStorage, error) {
	if datasetSource == metadata.Contrib {
		return datamartMetaCtor()
	}
	if datasetSource == metadata.Seed {
		return esMetaCtor()
	}
	if datasetSource == metadata.Augmented {
		return fileMetaCtor()
	}
	return nil, fmt.Errorf("unrecognized source `%v`", datasetSource)
}

func createResolverForSource(datasetSource metadata.DatasetSource, datasetID string, config *env.Config, taskConfig *task.IngestTaskConfig) *util.PathResolver {
	if datasetSource == metadata.Contrib {
		return util.NewPathResolver(&util.PathConfig{
			InputFolder:  path.Join(config.DatamartImportFolder, datasetID),
			OutputFolder: taskConfig.Resolver.Config.OutputFolder,
		})
	}
	if datasetSource == metadata.Seed {
		return util.NewPathResolver(&util.PathConfig{
			InputFolder:     config.D3MInputDir,
			InputSubFolders: "TRAIN/dataset_TRAIN",
			OutputFolder:    taskConfig.Resolver.Config.OutputFolder,
		})
	}
	if datasetSource == metadata.Augmented {
		return util.NewPathResolver(&util.PathConfig{
			InputFolder:  path.Join(config.TmpDataPath, "augmented", datasetID),
			OutputFolder: taskConfig.Resolver.Config.OutputFolder,
		})
	}
	return util.NewPathResolver(&util.PathConfig{
		InputFolder:     config.D3MInputDir,
		InputSubFolders: "TRAIN/dataset_TRAIN",
		OutputFolder:    taskConfig.Resolver.Config.OutputFolder,
	})
}
