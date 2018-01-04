package routes

import (
	"net/http"

	"github.com/pkg/errors"
	"goji.io/pat"

	"github.com/unchartedsoftware/distil-ingest/rest"
	"github.com/unchartedsoftware/distil/api/model"
	"github.com/unchartedsoftware/distil/api/pipeline"
)

// RankingHandler generates a route handler that will rank importance of
// variables for a given target.
func RankingHandler(ctor model.DataStorageCtor, restClient *rest.Client, dataDir string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// get index name
		index := pat.Param(r, "index")
		// get dataset name
		dataset := pat.Param(r, "dataset")
		// get target name
		target := pat.Param(r, "target")
		// get elasticsearch client
		client, err := ctor()
		if err != nil {
			handleError(w, err)
			return
		}
		// calculate importance
		importance, err := pipeline.Rank(restClient, client, dataset, index, target, dataDir)
		if err != nil {
			handleError(w, err)
			return
		}
		// marshall data
		err = handleJSON(w, importance)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal importance result into JSON"))
			return
		}
	}
}
