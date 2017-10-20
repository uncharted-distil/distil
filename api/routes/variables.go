package routes

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
	"goji.io/pat"

	"github.com/unchartedsoftware/distil/api/elastic"
	"github.com/unchartedsoftware/distil/api/model"
	"github.com/unchartedsoftware/plog"
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

// VariableTypeHandler generates a route handler that facilitates the update
// of a variable type.
func VariableTypeHandler(storageCtor model.StorageCtor, ctor elastic.ClientCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		params, err := getPostParameters(r)
		if err != nil {
			handleError(w, errors.Wrap(err, "Unable to parse post parameters"))
			return
		}
		field := params["field"].(string)
		typ := params["type"].(string)
		index := pat.Param(r, "index")
		dataset := pat.Param(r, "dataset")

		log.Infof("index: %s\tdataset: %s\tfield: %s\ttype: %s", index, dataset, field, typ)

		// get clients
		client, err := storageCtor()
		if err != nil {
			handleError(w, err)
			return
		}
		clientES, err := ctor()
		if err != nil {
			handleError(w, err)
			return
		}

		// update the variable type in the storage
		err = client.SetDataType(dataset, index, field, typ)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to update the data type in storage"))
			return
		}

		// update the variable type in the metadata
		err = model.SetDataType(clientES, dataset, index, field, typ)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to update the data type in metadata"))
			return
		}
	}
}

func getPostParameters(r *http.Request) (map[string]interface{}, error) {
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse POST request")
	}

	params := make(map[string]interface{})
	err = json.Unmarshal(body, &params)

	return params, err
}
