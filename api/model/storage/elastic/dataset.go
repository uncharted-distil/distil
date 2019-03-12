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

package elastic

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil-ingest/metadata"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/util/json"
	elastic "gopkg.in/olivere/elastic.v5"
)

const (
	// DatasetSuffix is the suffix for the dataset entry when stored in
	// elasticsearch.
	metadataType = "metadata"
	// Provenance for elastic
	Provenance       = "elastic"
	datasetsListSize = 1000
)

// ImportDataset is not supported (ES datasets are already ingested).
func (s *Storage) ImportDataset(id string, uri string) (string, error) {
	return "", errors.Errorf("Not Supported")
}

func (s *Storage) parseDatasets(res *elastic.SearchResult, includeIndex bool, includeMeta bool) ([]*api.Dataset, error) {
	var datasets []*api.Dataset
	for _, hit := range res.Hits.Hits {
		// parse hit into JSON
		src, err := json.Unmarshal(*hit.Source)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse dataset")
		}
		// get id
		id, ok := json.String(src, "datasetID")
		if !ok {
			id = hit.Id
		}
		// extract the name
		name, ok := json.String(src, "datasetName")
		if !ok || name == "NULL" {
			name = id
		}
		// extract the storage name
		storageName, ok := json.String(src, "storageName")
		if !ok {
			storageName = model.NormalizeDatasetID(name)
		}
		// extract the description
		description, ok := json.String(src, "description")
		if !ok {
			description = ""
		}
		// extract the summary
		summary, ok := json.String(src, "summary")
		if !ok {
			summary = ""
		}
		// extract the folder
		folder, ok := json.String(src, "datasetFolder")
		if !ok {
			summary = ""
		}
		// extract the summary
		summaryMachine, ok := json.String(src, "summaryMachine")
		if !ok {
			summary = ""
		}
		// extract the number of rows
		numRows, ok := json.Int(src, "numRows")
		if !ok {
			summary = ""
		}
		// extract the number of bytes
		numBytes, ok := json.Int(src, "numBytes")
		if !ok {
			summary = ""
		}
		// extract the variables list
		variables, err := s.parseVariables(hit, includeIndex, includeMeta)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse dataset")
		}
		// extract sources
		source, ok := json.String(src, "source")
		if !ok {
			source = string(metadata.Seed)
		}

		// write everythign out to result struct
		datasets = append(datasets, &api.Dataset{
			ID:          id,
			Name:        name,
			StorageName: storageName,
			Description: description,
			Folder:      folder,
			Summary:     summary,
			SummaryML:   summaryMachine,
			NumRows:     int64(numRows),
			NumBytes:    int64(numBytes),
			Variables:   variables,
			Provenance:  Provenance,
			Source:      metadata.DatasetSource(source),
		})
	}
	return datasets, nil
}

// FetchDatasets returns all datasets in the provided index.
func (s *Storage) FetchDatasets(includeIndex bool, includeMeta bool) ([]*api.Dataset, error) {
	// execute the ES query
	res, err := s.client.Search().
		Index(s.index).
		FetchSource(true).
		Size(datasetsListSize).
		Do(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "elasticsearch dataset fetch query failed")
	}
	return s.parseDatasets(res, includeIndex, includeMeta)
}

// FetchDataset returns a dataset in the provided index.
func (s *Storage) FetchDataset(datasetName string, includeIndex bool, includeMeta bool) (*api.Dataset, error) {
	query := elastic.NewMatchQuery("_id", datasetName)
	// execute the ES query
	res, err := s.client.Search().
		Query(query).
		Index(s.index).
		FetchSource(true).
		Size(datasetsListSize).
		Do(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "elasticsearch dataset fetch query failed")
	}
	datasets, err := s.parseDatasets(res, includeIndex, includeMeta)
	if err != nil {
		return nil, err
	}
	return datasets[0], nil
}

// SearchDatasets returns the datasets that match the search criteria in the
// provided index.
func (s *Storage) SearchDatasets(terms string, includeIndex bool, includeMeta bool) ([]*api.Dataset, error) {
	query := elastic.NewMultiMatchQuery(terms, "_id", "datasetFolder", "datasetID", "datasetName", "variables.colName", "description", "summaryMachine").
		Analyzer("standard")
	// execute the ES query
	res, err := s.client.Search().
		Query(query).
		Index(s.index).
		FetchSource(true).
		Size(datasetsListSize).
		Do(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "elasticsearch dataset search query failed")
	}
	return s.parseDatasets(res, includeIndex, includeMeta)
}

func (s *Storage) updateVariables(dataset string, variables []*model.Variable) error {
	// reserialize the data
	var serialized []map[string]interface{}
	for _, v := range variables {
		serialized = append(serialized, map[string]interface{}{
			model.VarNameField:             v.Name,
			model.VarIndexField:            v.Index,
			model.VarRoleField:             v.Role,
			model.VarSelectedRoleField:     v.SelectedRole,
			model.VarTypeField:             v.Type,
			model.VarOriginalTypeField:     v.OriginalType,
			model.VarImportanceField:       v.Importance,
			model.VarSuggestedTypesField:   v.SuggestedTypes,
			model.VarOriginalVariableField: v.OriginalVariable,
			model.VarDisplayVariableField:  v.DisplayName,
			model.VarDistilRole:            v.DistilRole,
			model.VarDeleted:               v.Deleted,
		})
	}

	source := map[string]interface{}{
		model.Variables: serialized,
	}

	// push the document into the metadata index
	_, err := s.client.Update().
		Index(s.index).
		Type(metadataType).
		Id(dataset).
		Doc(source).
		Do(context.Background())
	if err != nil {
		return errors.Wrapf(err, "failed to add document to index `%s`", s.index)
	}

	return nil
}

// SetDataType updates the data type of the field in ES.
func (s *Storage) SetDataType(dataset string, varName string, varType string) error {
	// Fetch all existing variables
	vars, err := s.FetchVariables(dataset, true, true)
	if err != nil {
		return errors.Wrapf(err, "failed to fetch existing variable")
	}

	// Update only the variable we care about
	for _, v := range vars {
		if v.Name == varName {
			v.Type = varType
		}
	}

	return s.updateVariables(dataset, vars)
}

// AddVariable adds a new variable to the dataset.
func (s *Storage) AddVariable(dataset string, varName string, varType string, varRole string) error {
	// query for existing variables
	vars, err := s.FetchVariables(dataset, true, true)
	if err != nil {
		return errors.Wrapf(err, "failed to fetch existing variable")
	}

	// add the new variables
	vars = append(vars, &model.Variable{
		Name:             varName,
		Index:            len(vars),
		Type:             varType,
		OriginalType:     varType,
		OriginalVariable: varName,
		DisplayName:      varName,
		DistilRole:       varRole,
		SuggestedTypes:   make([]*model.SuggestedType, 0),
	})

	return s.updateVariables(dataset, vars)
}

// DeleteVariable flags a variable as deleted.
func (s *Storage) DeleteVariable(dataset string, varName string) error {
	// query for existing variables
	vars, err := s.FetchVariables(dataset, true, true)
	if err != nil {
		return errors.Wrapf(err, "failed to fetch existing variable")
	}

	// soft delete the variable
	for _, v := range vars {
		if v.Name == varName {
			v.Deleted = true
		}
	}

	return s.updateVariables(dataset, vars)
}

// AddGrouping adds a grouping to the metadata.
func (s *Storage) AddGrouping(datasetName string, grouping model.Grouping) error {

	query := elastic.NewMatchQuery("_id", datasetName)
	// execute the ES query
	res, err := s.client.Search().
		Query(query).
		Index(s.index).
		FetchSource(true).
		Size(datasetsListSize).
		Do(context.Background())
	if err != nil {
		return errors.Wrap(err, "elasticsearch dataset fetch query failed")
	}

	if len(res.Hits.Hits) != 1 {
		return fmt.Errorf("default dataset meta not found")
	}
	hit := res.Hits.Hits[0]

	// parse hit into JSON
	source, err := json.Unmarshal(*hit.Source)
	if err != nil {
		return errors.Wrap(err, "elasticsearch dataset unmarshal failed")
	}

	variables, ok := json.Array(source, "variables")
	if !ok {
		return errors.Wrap(err, "variables unmarshal failed")
	}

	found := false
	for _, variable := range variables {
		name, ok := json.String(variable, "colName")
		if ok && name == grouping.IDCol {
			variable["grouping"] = json.StructToMap(grouping)
			found = true
		}
	}
	if !found {
		return fmt.Errorf("no variable match found for grouping")
	}

	// push the document into the metadata index
	_, err = s.client.Index().
		Index(s.index).
		Type(metadataType).
		Id(datasetName).
		BodyJson(source).
		Do(context.Background())
	if err != nil {
		return errors.Wrapf(err, "failed to add document to index `%s`", s.index)
	}

	return nil
}
