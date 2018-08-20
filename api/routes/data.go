package routes

import (
	"net/http"

	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil/api/model"
	"github.com/unchartedsoftware/distil/api/util/json"
	"goji.io/pat"
)

// DataHandler creates a route that fetches filtered data from backing storage instance.
func DataHandler(storageCtor model.DataStorageCtor, metaCtor model.MetadataStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// parse POST params
		params, err := getPostParameters(r)
		if err != nil {
			handleError(w, errors.Wrap(err, "Unable to parse post parameters"))
			return
		}

		// get variable names and ranges out of the params
		filterParams, err := model.ParseFilterParamsFromJSON(params)
		if err != nil {
			handleError(w, err)
			return
		}

		dataset := pat.Param(r, "dataset")
		invert := pat.Param(r, "invert")
		invertBool := false
		if invert == "true" {
			invertBool = true
		}

		// get filter client
		storage, err := storageCtor()
		if err != nil {
			handleError(w, err)
			return
		}

		// fetch filtered data based on the supplied search parameters
		data, err := storage.FetchData(dataset, filterParams, invertBool)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable fetch filtered data"))
			return
		}

		// marshall output into JSON
		bytes, err := json.Marshal(data)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal filtered data result into JSON"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(bytes)
	}
}
