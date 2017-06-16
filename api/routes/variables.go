package routes

import (
	"net/http"

	"github.com/pkg/errors"
	"github.com/unchartedsoftware/plog"
	"goji.io/pat"
	"gopkg.in/olivere/elastic.v3"

	"github.com/unchartedsoftware/distil/api/util/json"
)

// Variable represents a single variable description within a dataset.
type Variable struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// VariableResult represents the result of a datasets response.
type VariableResult struct {
	Variables []Variable `json:"variables"`
}

func parseVariables(searchHit *elastic.SearchHit) ([]Variable, error) {
	// unmarshal the hit source
	src, err := json.Unmarshal(*searchHit.Source)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse search result")
	}
	// get the variables array
	children, ok := json.Array(src, "variables")
	if !ok {
		return nil, errors.New("unable to parse variables from search result")
	}
	// for each variable, extract the `varName` and `varType`
	var variables []Variable
	for _, child := range children {
		name, ok := json.String(child, "varName")
		if !ok {
			continue
		}
		typ, ok := json.String(child, "varType")
		if !ok {
			continue
		}
		variables = append(variables, Variable{
			Name: name,
			Type: typ,
		})
	}
	return variables, nil
}

func fetchVariables(client *elastic.Client, index string, dataset string) ([]Variable, error) {
	log.Infof("Processing variables request for %s", dataset)
	// get dataset id
	datasetID := dataset + "_dataset"
	// create match query
	query := elastic.NewMatchQuery("_id", datasetID)
	// create fetch context
	fetchContext := elastic.NewFetchSourceContext(true)
	fetchContext.Include("variables")
	// execute the ES query
	res, err := client.Search().
		Query(query).
		Index(index).
		FetchSource(true).
		FetchSourceContext(fetchContext).
		Do()
	if err != nil {
		return nil, errors.Wrap(err, "elasticSearch variable fetch query failed")
	}
	// check that we have only one hit (should only ever be one matching dataset)
	if len(res.Hits.Hits) != 1 {
		return nil, errors.New("elasticSearch variable fetch query len(hits) != 1")
	}
	// extract output into JSON ready structs
	variables, err := parseVariables(res.Hits.Hits[0])
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse search result")
	}
	return variables, err
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
		variables, err := fetchVariables(client, index, dataset)
		if err != nil {
			handleError(w, err)
			return
		}
		// marshall output into JSON
		bytes, err := json.Marshal(VariableResult{
			Variables: variables,
		})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal variables result into JSON"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(bytes)
	}
}
