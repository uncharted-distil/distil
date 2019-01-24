package routes

import (
	"net/http"
	"path"

	"github.com/pkg/errors"
	"goji.io/pat"

	"github.com/unchartedsoftware/distil-compute/model"
	"github.com/unchartedsoftware/distil-compute/primitive/compute"

	"github.com/unchartedsoftware/distil-ingest/metadata"
	api "github.com/unchartedsoftware/distil/api/model"
)

// UploadHanler uploads a file to the local file system and then imports it.
func UploadHanler(metaCtor api.MetadataStorageCtor, localMetaCtor api.MetadataStorageCtor, outputPath string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		dataset := pat.Param(r, "dataset")

		// save the csv file in the file system datasets folder

		// create the raw dataset schema doc
		datasetID := model.NormalizeDatasetID(dataset)
		meta := model.NewMetadata(datasetID, dataset, "", datasetID)
		dr := model.NewDataResource("0", model.ResTypeRaw, []string{compute.D3MResourceFormat})
		dr.ResPath = path.Join(compute.D3MDataFolder, compute.D3MLearningData)
		meta.DataResources = []*model.DataResource{dr}

		schemaPath := path.Join(outputPath, datasetID, compute.D3MDataSchema)
		err := metadata.WriteSchema(meta, schemaPath)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to output schema"))
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
