//
//   Copyright © 2019 Uncharted Software Inc.
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

	"github.com/mitchellh/hashstructure"
	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil/api/env"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/task"
	"github.com/uncharted-distil/distil/api/util"
	log "github.com/unchartedsoftware/plog"
)

// ISISearchResult contains a single result from a query to the ISI datamart.
type ISISearchResult struct {
	Summary         string                     `json:"summary"`
	Score           float64                    `json:"score"`
	DatamartID      string                     `json:"datamart_id"`
	Metadata        []*ISISearchResultMetadata `json:"metadata"`
	MaterializeInfo string                     `json:"materialize_info"`
}

// ISISearchResultMetadata specifies the metadata of the datamart dataset.
type ISISearchResultMetadata struct {
	Selector []interface{}                    `json:"selector"`
	Metadata *ISISearchResultMetadataMetadata `json:"metadata"`
}

// ISISearchResultMetadataMetadata specifies the structure of the datamart dataset.
type ISISearchResultMetadataMetadata struct {
	StructuralType string                                    `json:"structural_type"`
	SemanticTypes  []string                                  `json:"semantic_types"`
	Dimension      *ISISearchResultMetadataMetadataDimension `json:"dimension"`
	Schema         string                                    `json:"schema"`
}

// ISISearchResultProvenance defines the source of the data.
type ISISearchResultProvenance struct {
	Source string `json:"source"`
}

// ISISearchResultMaterialization specifies how to materialize the dataset.
type ISISearchResultMaterialization struct {
	ID           string                                      `json:"id"`
	Score        float64                                     `json:"score"`
	Metadata     *ISISearchResultMaterializationMetadata     `json:"metadata"`
	Augmentation *ISISearchResultMaterializationAugmentation `json:"augmentation"`
	DatamartType string                                      `json:"datamart_type"`
}

// ISISearchResultMaterializationMetadata specifies the materialization metadata.
type ISISearchResultMaterializationMetadata struct {
	ConnectionURL string                                              `json:"connection_url"`
	SearchResult  *ISISearchResultMaterializationMetadataSearchResult `json:"search_result"`
	QueryJSON     string                                              `json:"query_json"`
	SearchType    string                                              `json:"search_type"`
}

// ISISearchResultMaterializationMetadataSearchResult specifies the materialization
// search results.
type ISISearchResultMaterializationMetadataSearchResult struct {
	PNodesNeeded          []string `json:"p_nodes_needed"`
	TargetQNodeColumnName string   `json:"target_q_node_column_name"`
}

// ISISearchResultMaterializationAugmentation specifies the materialization augmentation.
type ISISearchResultMaterializationAugmentation struct {
	Properties   string    `json:"properties"`
	LeftColumns  []float64 `json:"left_columns"`
	RightColumns []float64 `json:"right_columns"`
}

// ISISearchResultMetadataMetadataDimension has the specification for a dimension in a dataset.
type ISISearchResultMetadataMetadataDimension struct {
	Name          string   `json:"name"`
	SemanticTypes []string `json:"semantic_types"`
	Length        float64  `json:"length"`
}

// ISIMaterializedDataset container for the raw response from a materialize
// call to the ISI datamart.
type ISIMaterializedDataset struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

func isiSearch(datamart *Storage, query *SearchQuery, baseDataPath string) ([]byte, error) {
	log.Infof("querying ISI datamart")
	queryJSON, err := json.Marshal(query)
	if err != nil {
		return nil, errors.Wrap(err, "unable to marshal datamart query")
	}

	// need to store the query to file and send the file
	hash, err := hashstructure.Hash(query, nil)
	if err != nil {
		return nil, errors.Wrap(err, "unable to hash datamart query")
	}
	filepath := path.Join(env.GetTmpPath(), fmt.Sprintf("%v.json", hash))

	err = util.WriteFileWithDirs(filepath, queryJSON, os.ModePerm)
	if err != nil {
		return nil, errors.Wrap(err, "unable to store datamart query")
	}
	log.Infof("stored ISI query to filepath %s", filepath)

	responseRaw, err := datamart.client.PostFile(isiSearchFunction, "query", filepath, nil)
	if err != nil {
		return nil, errors.Wrap(err, "unable to post to ISI datamart search request")
	}

	return responseRaw, nil
}

func parseISISearchResult(responseRaw []byte, baseDataset *api.Dataset) ([]*api.Dataset, error) {
	var dmResults []*ISISearchResult
	err := json.Unmarshal(responseRaw, &dmResults)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse ISI datamart search request")
	}

	datasets := make([]*api.Dataset, 0)

	for _, res := range dmResults {
		vars := make([]*model.Variable, 0)
		for _, c := range res.Metadata {
			vars = append(vars, &model.Variable{
				Name:        c.Metadata.Dimension.Name,
				DisplayName: c.Metadata.Dimension.Name,
			})
		}
		joinSuggestions, joinScore, err := parseISIJoinSuggestion(res, baseDataset)
		if err != nil {
			return nil, errors.Wrap(err, "unable to parse ISI datamart join suggestions")
		}

		// need to get the specific search result string
		searchResultRaw, err := json.Marshal(res)
		if err != nil {
			return nil, errors.Wrap(err, "unable to marshal NYU search result")
		}

		datasets = append(datasets, &api.Dataset{
			ID:              res.DatamartID,
			Name:            res.DatamartID,
			Description:     res.Summary,
			Variables:       vars,
			Provenance:      ProvenanceISI,
			Summary:         res.Summary,
			JoinSuggestions: joinSuggestions,
			JoinScore:       joinScore,
			SearchResult:    string(searchResultRaw),
		})
	}

	return datasets, nil
}

func parseISIJoinSuggestion(result *ISISearchResult, baseDataset *api.Dataset) ([]*api.JoinSuggestion, float64, error) {
	// materialize_info has the join data in json structure
	var materialization ISISearchResultMaterialization
	err := json.Unmarshal([]byte(result.MaterializeInfo), &materialization)
	if err != nil {
		return nil, 0, errors.Wrap(err, "unable to unmarshal ISI datamart join suggestions")
	}

	joins := make([]*api.JoinSuggestion, 0)
	if materialization.Augmentation != nil && materialization.Augmentation.Properties == "join" {
		rightColumnNames := []string{}
		colNames := []string{}
		for _, colIndex := range materialization.Augmentation.RightColumns {
			colNames = append(colNames, result.Metadata[int(colIndex)].Metadata.Dimension.Name)
		}
		rightColumnNames = append(rightColumnNames, strings.Join(colNames[:], ", "))

		leftColumnNames := []string{}
		colNames = []string{}
		for _, colIndex := range materialization.Augmentation.LeftColumns {
			colNames = append(colNames, baseDataset.Variables[int(colIndex)].Name)
		}
		rightColumnNames = append(rightColumnNames, strings.Join(colNames[:], ", "))

		joins = append(joins, &api.JoinSuggestion{
			BaseDataset: baseDataset.ID,
			BaseColumns: leftColumnNames,
			JoinColumns: rightColumnNames,
		})
	}
	return joins, materialization.Score, nil
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

	// parse out the raw data
	var dataset ISIMaterializedDataset
	err = json.Unmarshal(data, &dataset)
	if err != nil {
		return "", errors.Wrap(err, "unable to parse ISI datamart materialized dataset")
	}

	// create the dataset meeting the d3m spec
	datasetPath, err := task.CreateDataset(id, []byte(dataset.Data), datamart.outputPath, datamart.config)
	if err != nil {
		return "", errors.Wrap(err, "unable to store dataset from ISI datamart")
	}

	// return the location of the expanded dataset folder
	return datasetPath, nil
}
