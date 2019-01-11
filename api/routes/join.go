package routes

import (
	"net/http"

	"goji.io/pat"

	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil-ingest/metadata"
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
		sourceLeft := pat.Param(r, "source-left")
		columnLeft := pat.Param(r, "column-left")
		datasetRight := pat.Param(r, "dataset-right")
		columnRight := pat.Param(r, "column-right")
		sourceRight := pat.Param(r, "source-right")

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

		leftJoin := &task.JoinSpec{
			Column:        columnLeft,
			DatasetFolder: datasetLeft,
			DatasetSource: metadata.DatasetSource(sourceLeft),
		}

		rightJoin := &task.JoinSpec{
			Column:        columnRight,
			DatasetFolder: datasetRight,
			DatasetSource: metadata.DatasetSource(sourceRight),
		}

		// run joining pipeline
		data, err := task.Join(leftJoin, rightJoin, varsLeft, varsRight)
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
