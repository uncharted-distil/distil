package routes

import (
	"net/http"

	"github.com/pkg/errors"
	//"goji.io/pat"

	"github.com/unchartedsoftware/distil/api/model"
)

// RankingResult represents a ranking response for a target variable.
type RankingResult struct {
	Rankings map[string]interface{} `json:"rankings"`
}

// VariableRankingHandler generates a route handler that allows to ranking
// variables of a dataset relative to the importance of a selected variable.
func VariableRankingHandler(ctorStorage model.DataStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// // get dataset name
		// dataset := pat.Param(r, "dataset")
		// // get variabloe name
		// target := pat.Param(r, "target")

		// // get storage client
		// storage, err := ctorStorage()
		// if err != nil {
		// 	handleError(w, err)
		// 	return
		// }

		// marshall output into JSON
		err := handleJSON(w, RankingResult{})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal summary result into JSON"))
			return
		}
	}
}
