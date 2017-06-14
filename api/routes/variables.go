package routes

import (
	"encoding/json"
	"net/http"

	"github.com/jeffail/gabs"
	"github.com/pkg/errors"
	"github.com/unchartedsoftware/plog"
	"goji.io/pat"
	"gopkg.in/olivere/elastic.v2"
)

type varDesc struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type variableList struct {
	Variables []varDesc `json:"variables"`
}

type dataset struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	variableList
}

type datasetList struct {
	Datasets []dataset `json:"datasets"`
}

func parseVariables(searchHit *elastic.SearchHit) (*variableList, error) {
	var result variableList

	resultJSON, err := gabs.ParseJSON(*searchHit.Source)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to parse search result")
	}
	children, err := resultJSON.Path("variables").Children()
	if err != nil {
		return nil, errors.Wrap(err, "Unable to parse variables from search result result")
	}
	for _, varData := range children {
		d := varData.Data().(map[string]interface{})
		result.Variables = append(result.Variables, varDesc{
			Name: d["varName"].(string),
			Type: d["varType"].(string),
		})
	}

	return &result, nil
}

func fetchVariables(client *elastic.Client, index string, dataset string) ([]byte, error) {
	boolQuery := elastic.NewBoolQuery().
		Must(elastic.NewMatchQuery("_id", dataset))

	fetchContext := elastic.NewFetchSourceContext(true)
	fetchContext.Include("variables")

	// execute the ES query
	searchResult, err := client.Search().
		Query(boolQuery).
		Index(index).
		FetchSource(true).
		FetchSourceContext(fetchContext).
		Do()

	if err != nil {
		return nil, errors.Wrap(err, "ElasticSearch variable fetch query failed")
	}
	if len(searchResult.Hits.Hits) != 1 {
		return nil, errors.New("ElasticSearch variable fetch query returned > 1 results")
	}

	// Extract output into JSON ready structs
	variables, err := parseVariables(searchResult.Hits.Hits[0])
	if err != nil {
		return nil, errors.Wrap(err, "Unable to parse search result JSON")
	}

	// Marshall output into JSON
	js, err := json.Marshal(variables)
	if err != nil {
		return nil, errors.Wrap(err, "Unable marshal result into JSON")
	}

	return js, err
}

// VariablesHandler generates a variable listing route handler associated with the caller supplied
// ES endpoint.  The handler returns a list of name/type tuples for the given dataset.
func VariablesHandler(client *elastic.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		// get index name
		index := pat.Param(r, "index")

		// get dataset name
		dataset := pat.Param(r, "dataset")

		log.Infof("Processing variables request for %s", dataset)
		datasetID := dataset + "_dataset"

		js, err := fetchVariables(client, index, datasetID)
		if err != nil {
			handleServerError(err, w)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	}
}
