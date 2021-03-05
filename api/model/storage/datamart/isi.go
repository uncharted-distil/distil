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
	"strings"

	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil/api/dataset"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/task"
	log "github.com/unchartedsoftware/plog"
)

// ISISearchResults is the basic search result container for ISI searches.
type ISISearchResults struct {
	Results []*ISISearchResult `json:"results"`
}

// ISISearchResult contains a single result from a query to the ISI datamart.
type ISISearchResult struct {
	Summary         *ISISearchResultSummary     `json:"summary"`
	Score           float64                     `json:"score"`
	Metadata        []*ISISearchResultMetadata  `json:"metadata"`
	MaterializeInfo string                      `json:"materialize_info"`
	ID              string                      `json:"id"`
	Augmentation    *SearchResultAugmentation   `json:"augmentation,omitempty"`
	ColumnNames     *ISISearchResultColumnNames `json:"all_column_names,omitempty"`
	Sample          string                      `json:"sample"`
}

// ISISearchResultSummary has a summary of the search result.
type ISISearchResultSummary struct {
	Title                string   `json:"title"`
	DatamartID           string   `json:"Datamart ID"`
	Score                string   `json:"Score"`
	URL                  string   `json:"URL"`
	Columns              []string `json:"Columns"`
	RecommendJoinColumns string   `json:"Recommend Join Columns"`
}

// ISISearchResultMetadata specifies the metadata of the datamart dataset.
type ISISearchResultMetadata struct {
	Selector []interface{}                    `json:"selector"`
	Metadata *ISISearchResultMetadataMetadata `json:"metadata"`
}

// ISISearchResultMetadataMetadata specifies the structure of the datamart dataset.
type ISISearchResultMetadataMetadata struct {
	StructuralType string                                    `json:"structural_type"`
	SemanticTypes  []interface{}                             `json:"semantic_types"`
	Dimension      *ISISearchResultMetadataMetadataDimension `json:"dimension"`
	Schema         string                                    `json:"schema"`
	Name           string                                    `json:"name"`
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
	QueryJSON     interface{}                                         `json:"query_json"`
	SearchType    string                                              `json:"search_type"`
}

// ISISearchResultMaterializationMetadataSearchResult specifies the materialization
// search results.
type ISISearchResultMaterializationMetadataSearchResult struct {
	PNodesNeeded          []string `json:"p_nodes_needed"`
	TargetQNodeColumnName string   `json:"target_q_node_column_name"`
	NumberOfVectors       string   `json:"number_of_vectors"`
	QNodesList            []string `json:"q_nodes_list"`
}

// ISISearchResultMaterializationAugmentation specifies the materialization augmentation.
type ISISearchResultMaterializationAugmentation struct {
	Properties   string      `json:"properties"`
	LeftColumns  [][]float64 `json:"left_columns"`
	RightColumns [][]float64 `json:"right_columns"`
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

// ISISearchResultColumnNames is the name of struct that Phil did not comment
// and it caused linting warnings that were bothering me so now I've fixed it.
type ISISearchResultColumnNames struct {
	LeftNames  []string `json:"left_names"`
	RightNames []string `json:"right_names"`
}

func isiSearch(datamart *Storage, query *SearchQuery, baseDataPath string) ([]byte, error) {
	log.Infof("querying ISI datamart")
	params := make(map[string]string)
	if len(query.Dataset.Keywords) > 0 {
		queryISI := map[string]interface{}{
			"keywords": query.Dataset.Keywords,
		}
		queryJSON, err := json.Marshal(queryISI)
		if err != nil {
			return nil, errors.Wrap(err, "unable to marshal datamart query")
		}
		params["query_json"] = string(queryJSON)
	}

	var responseRaw []byte
	var err error
	if baseDataPath != "" {
		responseRaw, err = datamart.client.PostFile(isiSearchFunction, "data", baseDataPath, params)
	} else {
		responseRaw, err = datamart.client.PostRequest(isiSearchFunctionNoData, params)
	}
	if err != nil {
		return nil, errors.Wrap(err, "unable to post to ISI datamart search request")
	}

	return responseRaw, nil
}

func parseISISearchResult(responseRaw []byte, baseDataset *api.Dataset) ([]*api.Dataset, error) {
	var dmResults *ISISearchResults
	err := json.Unmarshal(responseRaw, &dmResults)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse ISI datamart search request")
	}

	datasets := make([]*api.Dataset, 0)

	for _, res := range dmResults.Results {
		vars := make([]*model.Variable, 0)
		// for now, assume that a var has a selector with at least 2 elements.
		for _, c := range res.Summary.Columns {
			vars = append(vars, &model.Variable{
				Key:         c,
				DisplayName: c,
			})
		}
		joinSuggestions, joinScore, err := parseISIJoinSuggestion(res, baseDataset, vars)
		if err != nil {
			return nil, errors.Wrap(err, "unable to parse ISI datamart join suggestions")
		}

		datasets = append(datasets, &api.Dataset{
			ID:              res.ID,
			Name:            res.ID,
			Description:     res.Summary.Title,
			Variables:       vars,
			Provenance:      ProvenanceISI,
			Summary:         res.Summary.Title,
			JoinSuggestions: joinSuggestions,
			JoinScore:       joinScore,
		})
	}

	return datasets, nil
}

func parseISIJoinSuggestion(result *ISISearchResult, baseDataset *api.Dataset, vars []*model.Variable) ([]*api.JoinSuggestion, float64, error) {
	// need to get the specific search result string
	searchResultRaw, err := json.Marshal(result)
	if err != nil {
		return nil, 0, errors.Wrap(err, "unable to marshal ISI search result")
	}

	origin := &model.DatasetOrigin{
		SearchResult: string(searchResultRaw),
		Provenance:   ProvenanceISI,
	}

	// materialize_info has the join score
	var materialization ISISearchResultMaterialization
	err = json.Unmarshal([]byte(result.MaterializeInfo), &materialization)
	if err != nil {
		return nil, 0, errors.Wrap(err, "unable to unmarshal ISI datamart join suggestions")
	}

	// column names and indices stored separately in the search result
	joins := make([]*api.JoinSuggestion, 0)
	if result.Augmentation != nil && result.Augmentation.Type == "join" && result.ColumnNames != nil {
		leftColNames := make([]string, 0)
		for _, lc := range result.Augmentation.LeftColumns[0] {
			if lc < len(result.ColumnNames.LeftNames) {
				leftColNames = append(leftColNames, result.ColumnNames.LeftNames[lc])
			}
		}

		rightColNames := make([]string, 0)
		for _, rc := range result.Augmentation.RightColumns[0] {
			if rc < len(result.ColumnNames.RightNames) {
				rightColNames = append(rightColNames, result.ColumnNames.RightNames[rc])
			}
		}

		rightColumnNames := make([]string, 0)
		leftColumnNames := make([]string, 0)
		if len(rightColNames) == len(leftColNames) {
			rightColumnNames = append(rightColumnNames, strings.Join(rightColNames[:], ", "))
			leftColumnNames = append(leftColumnNames, strings.Join(leftColNames[:], ", "))
		} else {
			log.Warnf("right dataset (%s) join columns (%v) do not match left dataset (%s) join columns (%v)", result.ID, rightColNames, baseDataset.ID, leftColNames)
		}

		joins = append(joins, &api.JoinSuggestion{
			BaseDataset:   baseDataset.ID,
			BaseColumns:   leftColumnNames,
			JoinDataset:   result.ID,
			JoinColumns:   rightColumnNames,
			JoinScore:     result.Score,
			DatasetOrigin: origin,
			Index:         len(joins),
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
	var datasetRaw ISIMaterializedDataset
	err = json.Unmarshal(data, &datasetRaw)
	if err != nil {
		return "", errors.Wrap(err, "unable to parse ISI datamart materialized dataset")
	}

	// create the dataset meeting the d3m spec
	ds, err := dataset.NewTableDataset(id, []byte(datasetRaw.Data), true)
	if err != nil {
		return "", errors.Wrap(err, "unable to create raw dataset from ISI datamart materialized dataset")
	}
	_, datasetPath, err := task.CreateDataset(id, ds, datamart.outputPath, datamart.config)
	if err != nil {
		return "", errors.Wrap(err, "unable to store dataset from ISI datamart")
	}

	// return the location of the expanded dataset folder
	return datasetPath, nil
}
