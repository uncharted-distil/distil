package routes

import (
	"net/http"

	"github.com/pkg/errors"
	"goji.io/pat"

	"github.com/unchartedsoftware/distil/api/model"
	"github.com/unchartedsoftware/distil/api/task"
)

// RankingResult represents a ranking response for a target variable.
type RankingResult struct {
	Rankings map[string]interface{} `json:"rankings"`
}

// VariableRankingHandler generates a route handler that allows to ranking
// variables of a dataset relative to the importance of a selected variable.
func VariableRankingHandler(metaCtor model.MetadataStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// get dataset name
		dataset := pat.Param(r, "dataset")
		// get variabloe name
		target := pat.Param(r, "target")

		// get storage client
		storage, err := metaCtor()
		if err != nil {
			handleError(w, err)
			return
		}

		d, err := storage.FetchDataset(dataset, false, false)
		if err != nil {
			handleError(w, err)
			return
		}

		// compute rankings
		rankings, err := task.TargetRankPrimitive(d.Folder, target, d.Variables)
		if err != nil {
			handleError(w, err)
			return
		}

		var res map[string]interface{}

		for index, variable := range d.Variables {
			res[variable.Key] = rankings[index]
		}

		// marshall output into JSON
		err = handleJSON(w, RankingResult{
			Rankings: res,
		})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal summary result into JSON"))
			return
		}
	}
}
