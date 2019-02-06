package datamart

import (
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil-compute/model"
	api "github.com/unchartedsoftware/distil/api/model"
	"github.com/unchartedsoftware/distil/api/task"
)

// ISISearchResults wraps the response from the ISI datamart.
type ISISearchResults struct {
	Code    string             `json:"code"`
	Message string             `json:"message"`
	Data    []*ISISearchResult `json:"data"`
}

// ISISearchResult contains a single result from a query to the ISI datamart.
type ISISearchResult struct {
	Summary    string                   `json:"summary"`
	Score      float64                  `json:"score"`
	DatamartID string                   `json:"datamart_id"`
	Metadata   *ISISearchResultMetadata `json:"metadata"`
}

// ISISearchResultMetadata specifies the structure of the datamart dataset.
type ISISearchResultMetadata struct {
	DatamartID      string                          `json:"datamart_id"`
	Title           string                          `json:"title"`
	Description     string                          `json:"description"`
	URL             string                          `json:"url"`
	DateUpdated     string                          `json:"date_updated"`
	Provenance      *ISISearchResultProvenance      `json:"provenance"`
	Materialization *ISISearchResultMaterialization `json:"materialization"`
	Variables       []*ISISearchResultVariable      `json:"variables"`
	Keywords        []string                        `json:"keywords"`
}

// ISISearchResultProvenance defines the source of the data.
type ISISearchResultProvenance struct {
	Source string `json:"source"`
}

// ISISearchResultMaterialization specifies how to materialize the dataset.
type ISISearchResultMaterialization struct {
	PythonPath string `json:"python_path"`
}

// ISISearchResultVariable has the specification for a variable in a dataset.
type ISISearchResultVariable struct {
	DatamartID    string   `json:"datamart_id"`
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	SemanticTypes []string `json:"semantic_type"`
}

func parseISISearchResult(responseRaw []byte) ([]*api.Dataset, error) {
	var dmResult ISISearchResults
	err := json.Unmarshal(responseRaw, &dmResult)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse NYU datamart search request")
	}

	datasets := make([]*api.Dataset, 0)

	for _, res := range dmResult.Data {
		vars := make([]*model.Variable, 0)
		for _, c := range res.Metadata.Variables {
			vars = append(vars, &model.Variable{
				Name:        c.Name,
				DisplayName: c.Name,
			})
		}
		datasets = append(datasets, &api.Dataset{
			ID:          res.DatamartID,
			Name:        res.Metadata.Title,
			Description: res.Metadata.Description,
			Variables:   vars,
			Provenance:  ProvenanceISI,
			Summary:     res.Summary,
		})
	}

	return datasets, nil
}

// materializeISIDataset pulls a csv file from the ISI datamart.
func materializeISIDataset(datamart *Storage, id string, uri string) (string, error) {
	// get the csv file
	params := map[string]string{
		"datamart_id": id,
	}
	data, err := datamart.client.Get(datamart.getFunction, params)
	if err != nil {
		return "", err
	}

	// create the dataset meeting the d3m spec
	datasetPath, err := task.CreateDataset(id, data, datamart.outputPath, datamart.config)
	if err != nil {
		return "", errors.Wrap(err, "unable to store dataset from datamart")
	}

	// return the location of the expanded dataset folder
	return datasetPath, nil
}
