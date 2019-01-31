package routes

import (
	"net/http"

	"github.com/pkg/errors"
	"goji.io/pat"

	//"github.com/unchartedsoftware/distil-compute/model"

	api "github.com/unchartedsoftware/distil/api/model"
	"github.com/unchartedsoftware/distil/api/task"
)

// RankingResult represents a ranking response for a target variable.
type RankingResult struct {
	Rankings map[string]interface{} `json:"rankings"`
}

// VariableRankingHandler generates a route handler that allows to ranking
// variables of a dataset relative to the importance of a selected variable.
func VariableRankingHandler(metaCtor api.MetadataStorageCtor) func(http.ResponseWriter, *http.Request) {
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
		rankings, err := task.TargetRank(d.Folder, target, d.Variables, d.Source)
		if err != nil {
			handleError(w, err)
			return
		}

		res := make(map[string]interface{})
		for _, variable := range d.Variables {
			rank, ok := rankings[variable.Name]
			if ok {
				res[variable.Name] = rank
			} else {
				res[variable.Name] = 0
			}
		}

		// marshal output into JSON
		err = handleJSON(w, RankingResult{
			Rankings: res,
		})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal summary result into JSON"))
			return
		}
	}
}
