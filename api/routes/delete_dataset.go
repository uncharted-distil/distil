package routes

import (
	"github.com/uncharted-distil/distil/api/util"
	"net/http"

	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil/api/env"
	api "github.com/uncharted-distil/distil/api/model"
	"goji.io/v3/pat"
)

// DeletingDatasetHandler attempts to delete mutable datasets
func DeletingDatasetHandler(metaCtor api.MetadataStorageCtor, dataCtor api.DataStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// get params
		dataset := pat.Param(r, "dataset")
		// get meta and data storage
		metaStorage, err := metaCtor()
		if err != nil {
			handleError(w, err)
			return
		}
		dataStorage, err := dataCtor()
		if err != nil {
			handleError(w, err)
			return
		}

		ds, err := metaStorage.FetchDataset(dataset, true, true, false)
		if err != nil {
			handleError(w, err)
			return
		}
		folder := env.ResolvePath(ds.Source, ds.Folder)
		// verify dataset is a clone
		if ds.Immutable {
			handleError(w, errors.New("cannot delete Immutable dataset"))
			return
		}
		// delete db tables
		err = dataStorage.DeleteDataset(ds.StorageName)
		if err != nil {
			handleError(w, err)
			return
		}
		// delete meta
		err = metaStorage.DeleteDataset(dataset)
		if err != nil {
			handleError(w, err)
			return
		}
		// delete files
		err = util.RemoveContents(folder, true)
		if err != nil {
			handleError(w, err)
			return
		}
		// send json
		err = handleJSON(w, map[string]interface{}{"success": true})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal clustering result into JSON"))
			return
		}
	}
}
