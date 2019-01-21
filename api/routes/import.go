package routes

import (
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
func ImportHandler(metaCtor model.MetadataStorageCtor, localMetaCtor model.MetadataStorageCtor, config *task.IngestTaskConfig) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		index := pat.Param(r, "index")
		dataset := pat.Param(r, "dataset")
		source := metadata.DatasetSource(pat.Param(r, "source"))

		params, err := getPostParameters(r)
		if err != nil {
			handleError(w, err)
			return
		}
		id := ""
		if params["id"] != nil {
			id = params["id"].(string)
		}

		// update ingest config to use ingest URI.
		serverConfig, err := env.LoadConfig()
		if err != nil {
			handleError(w, err)
			return
		}

		meta, err := metaCtor()
		if err != nil {
			handleError(w, err)
			return
		}

		// import the dataset to the local filesystem.
		resolver := createResolverForSource(source, dataset, &serverConfig, config)
		uri := resolver.ResolveInputAbsolute(dataset)

		ingestConfig := *config
		ingestConfig.Resolver = resolver

		_, err = meta.ImportDataset(id, uri)
		if err != nil {
			handleError(w, err)
			return
		}

		// ingest the imported dataset.
		err = task.IngestDataset(localMetaCtor, index, dataset, source, &ingestConfig)
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

func createResolverForSource(datasetSource metadata.DatasetSource, dataset string, config *env.Config, taskConfig *task.IngestTaskConfig) *util.PathResolver {
	if datasetSource == metadata.Contrib {
		return util.NewPathResolver(&util.PathConfig{
			InputFolder:  path.Join(config.DatamartImportFolder, dataset),
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
			InputFolder:  path.Join(config.TmpDataPath, "augmented", dataset),
			OutputFolder: taskConfig.Resolver.Config.OutputFolder,
		})
	}
	return util.NewPathResolver(&util.PathConfig{
		InputFolder:     config.D3MInputDir,
		InputSubFolders: "TRAIN/dataset_TRAIN",
		OutputFolder:    taskConfig.Resolver.Config.OutputFolder,
	})
}
