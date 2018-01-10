package routes

import (
	"net/http"

	"github.com/pkg/errors"
	"goji.io/pat"

	"github.com/unchartedsoftware/distil/api/task"
)

// ResultsSummaryHandler bins predicted result data for consumption in a downstream summary view.
func IngestHandler(user string, password string, database string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// extract route parameters
		index := pat.Param(r, "index")
		dataset := pat.Param(r, "dataset")

		config := &task.ImportTaskConfig{
			Dataset:             dataset,
			ESMetadataIndexName: index,
		}

		err := task.Merge(config)
		if err != nil {
			handleError(w, err)
			return
		}

		err = task.Classify(config)
		if err != nil {
			handleError(w, err)
			return
		}

		err = task.Rank(config)
		if err != nil {
			handleError(w, err)
			return
		}

		err = task.Ingest(config)
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
