package routes

import (
	"net/http"

	"github.com/pkg/errors"
	"goji.io/pat"

	"github.com/unchartedsoftware/distil/api/elastic"
	"github.com/unchartedsoftware/distil/api/model"
)

// VariablesResult represents the result of a variables response.
type VariablesResult struct {
	Variables []*model.Variable `json:"variables"`
}

// VariablesHandler generates a route handler that facilitates a search of
// variable names and descriptions, returning a variable list for the specified
// dataset.
func VariablesHandler(ctor elastic.ClientCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// get index name
		index := pat.Param(r, "index")
		// get dataset name
		dataset := pat.Param(r, "dataset")
		// get elasticsearch client
		client, err := ctor()
		if err != nil {
			handleError(w, err)
			return
		}
		// fetch variables
		variables, err := model.FetchVariables(client, index, dataset)
		if err != nil {
			handleError(w, err)
			return
		}
		// marshall data
		err = handleJSON(w, VariablesResult{
			Variables: variables,
		})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal dataset result into JSON"))
			return
		}
	}
}
