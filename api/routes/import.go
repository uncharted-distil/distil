package routes

import (
	"net/http"

	"github.com/pkg/errors"

	"github.com/unchartedsoftware/distil/api/model"
	"github.com/unchartedsoftware/distil/api/task"
	"github.com/unchartedsoftware/distil/api/util"
)

// ImportHandler imports a dataset to the local file system and then ingests it.
func ImportHandler(metaCtor model.MetadataStorageCtor, config *task.IngestTaskConfig) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		params, err := getPostParameters(r)
		if err != nil {
			handleError(w, errors.Wrap(err, "Unable to parse post parameters"))
			return
		}
		uri := params["uri"].(string)
		index := params["index"].(string)
		dataset := params["dataset"].(string)

		meta, err := metaCtor()
		if err != nil {
			handleError(w, err)
			return
		}

		// import the dataset to the local filesystem.
		ingestURI, err := meta.ImportDataset(uri)
		if err != nil {
			handleError(w, err)
			return
		}

		// update ingest config to use ingest URI.
		resolver := util.NewPathResolver(&util.PathConfig{
			InputFolder:  ingestURI,
			OutputFolder: config.Resolver.Config.OutputFolder,
		})
		ingestConfig := *config
		ingestConfig.Resolver = resolver

		// ingest the imported dataset.
		err = task.IngestDataset(metaCtor, index, dataset, &ingestConfig)
		if err != nil {
			handleError(w, err)
			return
		}

		// marshall data and sent the response back
		err = handleJSON(w, map[string]interface{}{"result": "ingested"})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal result histogram into JSON"))
			return
		}
	}
}
