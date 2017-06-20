package routes

import (
	"net/http"

	"github.com/pkg/errors"
	"goji.io/pat"
	"gopkg.in/olivere/elastic.v3"

	"github.com/unchartedsoftware/distil/api/model"
)

// VariableResult represents the result of a datasets response.
type VariableResult struct {
	Variables []model.Variable `json:"variables"`
}

// VariablesHandler generates a variable listing route handler associated with
// the caller supplied ES endpoint.  The handler returns a list of name/type
// tuples for the given dataset.
func VariablesHandler(client *elastic.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// get index name
		index := pat.Param(r, "index")
		// get dataset name
		dataset := pat.Param(r, "dataset")
		// fetch the variables
		variables, err := model.FetchVariables(client, index, dataset)
		if err != nil {
			handleError(w, err)
			return
		}
		// marshall output into JSON
		err = handleJSON(w, VariableResult{
			Variables: variables,
		})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal variables result into JSON"))
			return
		}
	}
}
