package datamart

import (
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil-compute/model"
	api "github.com/unchartedsoftware/distil/api/model"
)

// SearchResults is the basic search result container.
type SearchResults struct {
	Results []*SearchResult `json:"results"`
}

// SearchResult contains the basic dataset info.
type SearchResult struct {
	ID         string                `json:"id"`
	Score      float64               `json:"score"`
	Discoverer string                `json:"discoverer"`
	Metadata   *SearchResultMetadata `json:"metadata"`
}

// SearchResultMetadata represents the dataset metadata.
type SearchResultMetadata struct {
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Size        float64               `json:"size"`
	NumRows     float64               `json:"nb_rows"`
	Columns     []*SearchResultColumn `json:"columns"`
	Date        string                `json:"date"`
}

// SearchResultColumn has information on a dataset column.
type SearchResultColumn struct {
	Name           string `json:"name"`
	StructuralType string `json:"structural_type"`
}

func parseNYUSearchResult(responseRaw []byte) ([]*api.Dataset, error) {
	var dmResult SearchResults
	err := json.Unmarshal(responseRaw, &dmResult)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse NYU datamart search request")
	}

	datasets := make([]*api.Dataset, 0)

	for _, res := range dmResult.Results {
		vars := make([]*model.Variable, 0)
		for _, c := range res.Metadata.Columns {
			vars = append(vars, &model.Variable{
				Name:        c.Name,
				DisplayName: c.Name,
			})
		}
		datasets = append(datasets, &api.Dataset{
			ID:          res.ID,
			Name:        res.Metadata.Name,
			Description: res.Metadata.Description,
			NumRows:     int64(res.Metadata.NumRows),
			NumBytes:    int64(res.Metadata.Size),
			Variables:   vars,
			Provenance:  Provenance,
		})
	}

	return datasets, nil
}
