package datamart

import (
	"encoding/json"
	"strings"

	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil-compute/model"
	api "github.com/unchartedsoftware/distil/api/model"
)

const (
	// DatasetSuffix is the suffix for the dataset entry when stored in
	// elasticsearch.
	metadataType       = "metadata"
	datasetsListSize   = 1000
	provenance         = "datamart"
	searchRESTFunction = "search"
)

// SearchQuery is the basic search query container.
type SearchQuery struct {
	Query *SearchQueryProperties `json:"query,omitempty"`
}

// SearchQueryProperties contains the basic properties to query.
type SearchQueryProperties struct {
	Dataset *SearchQueryDatasetProperties `json:"dataset,omitempty"`
}

// SearchQueryDatasetProperties represents queryin on metadata.
type SearchQueryDatasetProperties struct {
	About       string   `json:"about,omitempty"`
	Description []string `json:"description,omitempty"`
	Name        []string `json:"name,omitempty"`
	Keywords    []string `json:"keywords,omitempty"`
}

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

// ImportDataset makes the dataset available for ingest and returns
// the URI to use for ingest.
func (s *Storage) ImportDataset(uri string) (string, error) {
	// dataset is already on local file system and accessible for ingest
	return uri, nil
}

// FetchDatasets returns all datasets in the provided index.
func (s *Storage) FetchDatasets(includeIndex bool, includeMeta bool) ([]*api.Dataset, error) {
	// use default string in search to get complete list
	return s.SearchDatasets("", includeIndex, includeMeta)
}

// FetchDataset returns a dataset in the provided index.
func (s *Storage) FetchDataset(datasetName string, includeIndex bool, includeMeta bool) (*api.Dataset, error) {
	return nil, errors.Errorf("Not implemented")
}

// SearchDatasets returns the datasets that match the search criteria in the
// provided index.
func (s *Storage) SearchDatasets(terms string, includeIndex bool, includeMeta bool) ([]*api.Dataset, error) {
	rawSets, err := s.searchREST(terms)
	if err != nil {
		return nil, err
	}

	return s.parseDatasets(rawSets)
}

// SetDataType is not supported by the datamart.
func (s *Storage) SetDataType(dataset string, varName string, varType string) error {
	return errors.Errorf("Not supported")
}

// AddVariable is not supported by the datamart.
func (s *Storage) AddVariable(dataset string, varName string, varType string, varRole string) error {
	return errors.Errorf("Not supported")
}

// DeleteVariable is not supported by the datamart.
func (s *Storage) DeleteVariable(dataset string, varName string) error {
	return errors.Errorf("Not supported")
}

func (s *Storage) parseDatasets(raw *SearchResults) ([]*api.Dataset, error) {
	datasets := make([]*api.Dataset, 0)

	for _, res := range raw.Results {
		vars := make([]*model.Variable, 0)
		for _, c := range res.Metadata.Columns {
			vars = append(vars, &model.Variable{
				Name:         c.Name,
				OriginalName: c.Name,
				DisplayName:  c.Name,
			})
		}
		datasets = append(datasets, &api.Dataset{
			Name:        res.Metadata.Name,
			Description: res.Metadata.Description,
			NumRows:     int64(res.Metadata.NumRows),
			NumBytes:    int64(res.Metadata.Size),
			Variables:   vars,
			Provenance:  provenance,
		})
	}

	return datasets, nil
}

func (s *Storage) searchREST(searchText string) (*SearchResults, error) {
	terms := strings.Fields(searchText)

	// get complete URI for the endpoint
	query := &SearchQuery{
		Query: &SearchQueryProperties{
			Dataset: &SearchQueryDatasetProperties{
				About: searchText,
				//Name:        terms,
				Description: terms,
				//Keywords:    terms,
			},
		},
	}
	queryJSON, err := json.Marshal(query)
	if err != nil {
		return nil, errors.Wrap(err, "unable to marshal datamart query")
	}

	responseRaw, err := s.client.PostJSON(searchRESTFunction, queryJSON)
	if err != nil {
		return nil, errors.Wrap(err, "unable to post datamart search request")
	}

	// parse result
	var dmResult SearchResults
	err = json.Unmarshal(responseRaw, &dmResult)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse datamart search request")
	}

	return &dmResult, nil
}
