//
//   Copyright © 2021 Uncharted Software Inc.
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package datamart

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil-compute/primitive/compute"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/task"
	"github.com/uncharted-distil/distil/api/util"
	log "github.com/unchartedsoftware/plog"
)

// SearchResults is the basic search result container.
type SearchResults struct {
	Results []*SearchResult `json:"results"`
}

// SearchResult contains the basic dataset info.
type SearchResult struct {
	ID                 string                    `json:"id"`
	Score              float64                   `json:"score"`
	Discoverer         string                    `json:"discoverer"`
	Metadata           *SearchResultMetadata     `json:"metadata"`
	Augmentation       *SearchResultAugmentation `json:"augmentation,omitempty"`
	SuppliedID         string                    `json:"supplied_id"`
	SuppliedResourceID string                    `json:"supplied_resource_id"`
}

// SearchResultAugmentation contains data augmentation info.
type SearchResultAugmentation struct {
	Type             string     `json:"type"`
	LeftColumns      [][]int    `json:"left_columns"`
	RightColumns     [][]int    `json:"right_columns"`
	LeftColumnsNames [][]string `json:"left_columns_names"`
}

// SearchResultMetadata represents the dataset metadata.
type SearchResultMetadata struct {
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	Size        float64                  `json:"size"`
	NumRows     float64                  `json:"nb_rows"`
	Columns     []*SearchResultColumn    `json:"columns"`
	Materialize *SearchResultMaterialize `json:"materialize"`
	Date        string                   `json:"date"`
}

// SearchResultColumn has information on a dataset column.
type SearchResultColumn struct {
	Name           string   `json:"name"`
	StructuralType string   `json:"structural_type"`
	SemanticTypes  []string `json:"semantic_types"`
}

// SearchResultMaterialize contains the materialization info.
type SearchResultMaterialize struct {
	DirectURL string `json:"direct_url"`
	ID        string `json:"identifier"`
}

func nyuSearch(datamart *Storage, query *SearchQuery, baseDataPath string) ([]byte, error) {
	queryNYU := map[string]interface{}{
		"keywords": query.Dataset.Keywords,
	}
	queryJSON, err := json.Marshal(queryNYU)
	if err != nil {
		return nil, errors.Wrap(err, "unable to marshal datamart query")
	}
	params := map[string]string{"query": string(queryJSON)}

	var responseRaw []byte
	if baseDataPath != "" {
		responseRaw, err = datamart.client.PostFile(nyuSearchFunction, "data", baseDataPath, params)
	} else {
		responseRaw, err = datamart.client.PostRequest(nyuSearchFunction, params)
	}
	if err != nil {
		return nil, errors.Wrap(err, "unable to post to NYU datamart search request")
	}

	return responseRaw, nil
}

func parseNYUJoinSuggestion(result *SearchResult, baseDataset *api.Dataset) ([]*api.JoinSuggestion, error) {
	// need to get the specific search result string
	searchResultRaw, err := json.Marshal(result)
	if err != nil {
		return nil, errors.Wrap(err, "unable to marshal NYU search result")
	}

	origin := &model.DatasetOrigin{
		SearchResult: string(searchResultRaw),
		Provenance:   ProvenanceNYU,
	}

	joins := make([]*api.JoinSuggestion, 0)
	if result.Augmentation != nil && result.Augmentation.Type == "join" {
		rightColumnNames := []string{}
		for _, joinColumns := range result.Augmentation.RightColumns {
			colNames := []string{}
			for _, colIndex := range joinColumns {
				if colIndex >= 0 {
					colNames = append(colNames, result.Metadata.Columns[colIndex].Name)
				} else {
					log.Warnf("invalid column index (%d) received for join suggestion from NYU", colIndex)
				}
			}
			rightColumnNames = append(rightColumnNames, strings.Join(colNames[:], ", "))
		}
		leftColumnNames := []string{}
		for _, joinColumns := range result.Augmentation.LeftColumnsNames {
			leftColumnNames = append(leftColumnNames, strings.Join(joinColumns, ", "))
		}

		joins = append(joins, &api.JoinSuggestion{
			BaseDataset:   baseDataset.ID,
			BaseColumns:   leftColumnNames,
			JoinDataset:   result.Metadata.Name,
			JoinColumns:   rightColumnNames,
			JoinScore:     result.Score,
			DatasetOrigin: origin,
			Index:         len(joins),
		})
	}
	return joins, nil
}

func mapNYUDataTypesToDistil(nyuType string) string {
	switch nyuType {
	case "http://schema.org/Boolean":
		return model.BoolType
	case "http://schema.org/DateTime":
		return model.DateTimeType
	case "http://schema.org/Float":
		return model.RealType
	case "http://schema.org/Integer":
		return model.IntegerType
	case "http://schema.org/latitude":
		return model.LatitudeType
	case "http://schema.org/longitude":
		return model.LongitudeType
	case "http://schema.org/Text":
		return model.StringType
	default:
		return model.UnknownType
	}
}

func parseNYUSearchResult(responseRaw []byte, baseDataset *api.Dataset) ([]*api.Dataset, error) {
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
				Key:          c.Name,
				DisplayName:  c.Name,
				OriginalType: mapNYUDataTypesToDistil(c.StructuralType),
				DistilRole:   []string{model.VarDistilRoleData},
			})
		}

		joinSuggestions, err := parseNYUJoinSuggestion(res, baseDataset)
		if err != nil {
			return nil, errors.Wrap(err, "unable to parse NYU datamart join suggestions")
		}

		datasets = append(datasets, &api.Dataset{
			ID:              res.ID,
			Name:            res.Metadata.Name,
			Description:     res.Metadata.Description,
			NumRows:         int64(res.Metadata.NumRows),
			NumBytes:        int64(res.Metadata.Size),
			Variables:       vars,
			Provenance:      ProvenanceNYU,
			Source:          "contrib",
			JoinSuggestions: joinSuggestions,
			JoinScore:       res.Score,
			// parse out more information for type
		})
	}
	return datasets, nil
}

// materializeNYUDataset pulls a d3m directory and extracts its contents.
func materializeNYUDataset(datamart *Storage, id string, uri string) (string, error) {
	name := path.Base(uri)
	// get the compressed dataset
	requestURI := fmt.Sprintf("%s/%s", nyuGetFunction, id)
	params := map[string]string{
		"format":         "d3m",
		"format_version": "4.0.0",
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
	formattedPath, err := task.Format(extractedSchema, name, datamart.ingestConfig)
	if err != nil {
		return "", errors.Wrap(err, "unable to format datamart dataset")
	}

	return formattedPath, nil
}
