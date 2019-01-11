package routes

import (
	"net/http"

	"github.com/pkg/errors"
	"goji.io/pat"

	"github.com/unchartedsoftware/distil-compute/model"
	api "github.com/unchartedsoftware/distil/api/model"
)

// VariablesResult represents the result of a variables response.
type VariablesResult struct {
	Variables []*model.Variable `json:"variables"`
}

// VariablesHandler generates a route handler that facilitates a search of
// variable names and descriptions, returning a variable list for the specified
// dataset.
func VariablesHandler(metaCtor api.MetadataStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// get dataset name
		dataset := pat.Param(r, "dataset")
		// get elasticsearch client
		meta, err := metaCtor()
		if err != nil {
			handleError(w, err)
			return
		}
		// fetch variables
		variables, err := meta.FetchVariables(dataset, false, false)
		if err != nil {
			handleError(w, err)
			return
		}
		// marshal data
		err = handleJSON(w, VariablesResult{
			Variables: variables,
		})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal dataset result into JSON"))
			return
		}
	}
}
