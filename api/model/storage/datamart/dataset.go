package datamart

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil-compute/model"
	"github.com/unchartedsoftware/distil-compute/primitive/compute"
	api "github.com/unchartedsoftware/distil/api/model"
	"github.com/unchartedsoftware/distil/api/task"
	"github.com/unchartedsoftware/distil/api/util"
)

const (
	// DatasetSuffix is the suffix for the dataset entry when stored in
	// elasticsearch.
	metadataType     = "metadata"
	datasetsListSize = 1000
	// Provenance for datamart
	Provenance      = "datamart"
	getRESTFunction = "download"
)

// SearchQuery contains the basic properties to query.
type SearchQuery struct {
	Dataset *SearchQueryDatasetProperties `json:"dataset,omitempty"`
}

// SearchQueryDatasetProperties represents queryin on metadata.
type SearchQueryDatasetProperties struct {
	About       string   `json:"about"`
	Description []string `json:"description"`
	Name        []string `json:"name,omitempty"`
	Keywords    []string `json:"keywords,omitempty"`
}

// ImportDataset makes the dataset available for ingest and returns
// the URI to use for ingest.
func (s *Storage) ImportDataset(id string, uri string) (string, error) {
	//TODO: MAKE THIS WORK ON APIs OTHER THAN NYU!
	name := path.Base(uri)
	// get the compressed dataset
	requestURI := fmt.Sprintf("%s/%s", getRESTFunction, id)
	params := map[string]string{
		"format": "d3m",
	}
	data, err := s.client.Get(requestURI, params)
	if err != nil {
		return "", err
	}

	// write the compressed dataset to disk
	zipFilename := path.Join(s.outputPath, fmt.Sprintf("%s.zip", name))
	err = util.WriteFileWithDirs(zipFilename, data, os.ModePerm)
	if err != nil {
		return "", errors.Wrap(err, "unable to store dataset from datamart")
	}

	// expand the archive into a dataset folder
	extractedArchivePath := path.Join(s.outputPath, name)
	err = util.Unzip(zipFilename, extractedArchivePath)
	if err != nil {
		return "", errors.Wrap(err, "unable to extract datamart archive")
	}

	// format the dataset
	extractedSchema := path.Join(extractedArchivePath, compute.D3MDataSchema)
	formattedPath, err := task.Format(extractedSchema, s.config)
	if err != nil {
		return "", errors.Wrap(err, "unable to format datamart dataset")
	}

	// copy the formatted output to the datamart output path (delete existing copy)
	err = util.RemoveContents(s.outputPath)
	if err != nil {
		return "", errors.Wrap(err, "unable to delete raw datamart dataset")
	}

	err = util.Copy(formattedPath, extractedArchivePath)
	if err != nil {
		return "", errors.Wrap(err, "unable to copy formatted datamart dataset")
	}

	// return the location of the expanded dataset folder
	return formattedPath, nil
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

func (s *Storage) searchREST(searchText string) (*SearchResults, error) {
	terms := strings.Fields(searchText)

	// get complete URI for the endpoint
	query := &SearchQuery{
		Dataset: &SearchQueryDatasetProperties{
			About: searchText,
			//Name:        terms,
			Description: terms,
			//Keywords:    terms,
		},
	}
	queryJSON, err := json.Marshal(query)
	if err != nil {
		return nil, errors.Wrap(err, "unable to marshal datamart query")
	}

	responseRaw, err := s.client.PostRequest(s.searchFunction, map[string]string{"query": string(queryJSON)})
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
