package datamart

import (
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil-compute/model"
	"github.com/unchartedsoftware/distil-compute/primitive/compute"
	api "github.com/unchartedsoftware/distil/api/model"
	"github.com/unchartedsoftware/distil/api/task"
	"github.com/unchartedsoftware/distil/api/util"
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

func nyuSearch(datamart *Storage, query *SearchQuery) ([]byte, error) {
	queryJSON, err := json.Marshal(query)
	if err != nil {
		return nil, errors.Wrap(err, "unable to marshal datamart query")
	}

	responseRaw, err := datamart.client.PostRequest(nyuSearchFunction, map[string]string{"query": string(queryJSON)})
	if err != nil {
		return nil, errors.Wrap(err, "unable to post to NYU datamart search request")
	}

	return responseRaw, nil
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
			Provenance:  ProvenanceNYU,
		})
	}

	return datasets, nil
}

// materializeNYUDataset pulls a d3m directory and extracts its contents.
func materializeNYUDataset(datamart *Storage, id string, uri string) (string, error) {
	name := path.Base(uri)
	// get the compressed dataset
	requestURI := fmt.Sprintf("%s/%s", getRESTFunction, id)
	params := map[string]string{
		"format": "d3m",
	}
	data, err := datamart.client.Get(requestURI, params)
	if err != nil {
		return "", err
	}

	// write the compressed dataset to disk
	zipFilename := path.Join(datamart.outputPath, fmt.Sprintf("%s.zip", name))
	err = util.WriteFileWithDirs(zipFilename, data, os.ModePerm)
	if err != nil {
		return "", errors.Wrap(err, "unable to store dataset from datamart")
	}

	// expand the archive into a dataset folder
	extractedArchivePath := path.Join(datamart.outputPath, name)
	err = util.Unzip(zipFilename, extractedArchivePath)
	if err != nil {
		return "", errors.Wrap(err, "unable to extract datamart archive")
	}

	// format the dataset
	extractedSchema := path.Join(extractedArchivePath, compute.D3MDataSchema)
	formattedPath, err := task.Format(extractedSchema, datamart.config)
	if err != nil {
		return "", errors.Wrap(err, "unable to format datamart dataset")
	}

	// copy the formatted output to the datamart output path (delete existing copy)
	err = util.RemoveContents(datamart.outputPath)
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
