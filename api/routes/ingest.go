package routes

import (
	"net/http"

	"github.com/pkg/errors"
	"goji.io/pat"

	"github.com/unchartedsoftware/distil/api/task"
)

// IngestHandler ingests a dataset into ES & postgres. It assumes that SetHttpClient
// raw data is on the distil instance.
func IngestHandler(config *task.IngestTaskConfig) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// extract route parameters
		index := pat.Param(r, "index")
		dataset := pat.Param(r, "dataset")

		err := task.Merge(index, dataset, config)
		if err != nil {
			handleError(w, err)
			return
		}

		err = task.Classify(index, dataset, config)
		if err != nil {
			handleError(w, err)
			return
		}

		err = task.Rank(index, dataset, config)
		if err != nil {
			handleError(w, err)
			return
		}

		err = task.Ingest(index, dataset, config)
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
