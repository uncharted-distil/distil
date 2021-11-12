package routes

import (
	"net/http"

	"github.com/pkg/errors"
	"goji.io/v3/pat"

	"github.com/uncharted-distil/distil/api/env"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/task"
)

// SegmentationHandler will segment the specified remote sensing dataset.
func SegmentationHandler(metaCtor api.MetadataStorageCtor, dataCtor api.DataStorageCtor, config env.Config) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// get dataset name
		dataset := pat.Param(r, "dataset")
		// get variable name
		variable := pat.Param(r, "variable")

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

		ds, err := metaStorage.FetchDataset(dataset, false, false, false)
		if err != nil {
			handleError(w, err)
			return
		}

		outputURI, err := task.Segment(ds, dataStorage, variable)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable segment dataset"))
			return
		}

		// marshal output into JSON
		err = handleJSON(w, map[string]interface{}{
			"uri": outputURI,
		})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal clustering result into JSON"))
			return
		}
	}
}
