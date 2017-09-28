package routes

import (
	"net/http"

	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil/api/model"
	"github.com/unchartedsoftware/distil/api/util/json"
	"goji.io/pat"
)

const (
	defaultSearchSize = 100
	searchSizeLimit   = 1000
	// NumericalFilter represents a numerical type of filter.
	NumericalFilter = "numerical"
	// CategoricalFilter represents a categorcial type of filter.
	CategoricalFilter = "categorical"
)

// FilteredDataHandler creates a route that fetches filtered data from backing storage instance.
func FilteredDataHandler(ctor model.StorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		dataset := pat.Param(r, "dataset")
		esIndex := pat.Param(r, "esIndex")

		// get variable names and ranges out of the params
		filterParams, err := ParseFilterParams(r)
		if err != nil {
			handleError(w, err)
			return
		}

		// get filter client
		client, err := ctor()
		if err != nil {
			handleError(w, err)
			return
		}

		// fetch filtered data based on the supplied search parameters
		data, err := model.FetchFilteredData(client, dataset, esIndex, filterParams)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal summary result into JSON"))
			return
		}

		// marshall output into JSON
		bytes, err := json.Marshal(data)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal summary result into JSON"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(bytes)
	}
}
