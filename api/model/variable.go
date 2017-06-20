package model

import (
	"github.com/pkg/errors"
	"gopkg.in/olivere/elastic.v3"

	"github.com/unchartedsoftware/distil/api/util/json"
)

const (
	// Variables is the field name which stores the variables in elasticsearch.
	Variables = "variables"
	// VarNameField is the field name for the variable name.
	VarNameField = "varName"
	// VarTypeField is the field name for the variable type.
	VarTypeField = "varType"
)

// Variable represents a single variable description within a dataset.
type Variable struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

func parseVariables(searchHit *elastic.SearchHit) ([]Variable, error) {
	// unmarshal the hit source
	src, err := json.Unmarshal(*searchHit.Source)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse search result")
	}
	// get the variables array
	children, ok := json.Array(src, Variables)
	if !ok {
		return nil, errors.New("unable to parse variables from search result")
	}
	// for each variable, extract the `varName` and `varType`
	var variables []Variable
	for _, child := range children {
		name, ok := json.String(child, VarNameField)
		if !ok {
			continue
		}
		typ, ok := json.String(child, VarTypeField)
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

// FetchVariables returns all the variables for the provided index and dataset.
func FetchVariables(client *elastic.Client, index string, dataset string) ([]Variable, error) {
	// get dataset id
	datasetID := dataset + DatasetSuffix
	// create match query
	query := elastic.NewMatchQuery("_id", datasetID)
	// create fetch context
	fetchContext := elastic.NewFetchSourceContext(true)
	fetchContext.Include(Variables)
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
