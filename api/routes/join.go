package routes

import (
	"net/http"

	"goji.io/pat"

	"github.com/pkg/errors"
	api "github.com/unchartedsoftware/distil/api/model"
	"github.com/unchartedsoftware/distil/api/task"
	"github.com/unchartedsoftware/distil/api/util/json"
)

// JoinHandler generates a route handler that joins two datasets using caller supplied
// columns.  The joined data is returned to the caller, but is NOT added to storage.
func JoinHandler(metaCtor api.MetadataStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// get dataset name
		datasetLeft := pat.Param(r, "dataset-left")
		datasetRight := pat.Param(r, "dataset-right")
		columnLeft := pat.Param(r, "column-left")
		columnRight := pat.Param(r, "column-right")

		// get storage client
		storage, err := metaCtor()
		if err != nil {
			handleError(w, err)
			return
		}

		// fetch vars for each dataset
		varsLeft, err := storage.FetchVariables(datasetLeft, false, true)
		if err != nil {
			handleError(w, err)
			return
		}

		varsRight, err := storage.FetchVariables(datasetRight, false, true)
		if err != nil {
			handleError(w, err)
		}

		// run joining pipeline
		data, err := task.Join(datasetLeft, datasetRight, columnLeft, columnRight, varsLeft, varsRight)
		if err != nil {
			handleError(w, err)
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
