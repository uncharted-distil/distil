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

	elastic "github.com/olivere/elastic/v7"
	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/metadata"
	"github.com/uncharted-distil/distil-compute/model"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/util/json"
)

const (
	// Provenance for elastic
	Provenance       = "elastic"
	datasetsListSize = 1000
)

// ImportDataset is not supported (ES datasets are already ingested).
func (s *Storage) ImportDataset(id string, uri string) (string, error) {
	return "", errors.Errorf("Not Supported")
}

// CloneDataset is not supported (ES datasets are already ingested).
func (s *Storage) CloneDataset(dataset string, datasetNew string, storageNameNew string, folderNew string) error {
	ds, err := s.FetchDataset(dataset, true, true, false)
	if err != nil {
		return err
	}

	// update the id to match the new info
	ds.ID = datasetNew
	ds.StorageName = storageNameNew
	ds.Name = datasetNew
	ds.Folder = folderNew
	ds.Source = metadata.Augmented
	// cloned datasets CAN be altered
	ds.Immutable = false
	ds.Clone = true

	return s.UpdateDataset(ds)
}

// UpdateDataset updates a dataset already stored in ES.
func (s *Storage) UpdateDataset(dataset *api.Dataset) error {
	source := map[string]interface{}{
		"datasetName":     dataset.Name,
		"datasetID":       dataset.ID,
		"storageName":     dataset.StorageName,
		"description":     dataset.Description,
		"summary":         dataset.Summary,
		"summaryMachine":  dataset.SummaryML,
		"numRows":         dataset.NumRows,
		"numBytes":        dataset.NumBytes,
		"variables":       dataset.Variables,
		"datasetFolder":   dataset.Folder,
		"source":          dataset.Source,
		"datasetOrigins":  dataset.JoinSuggestions,
		"type":            dataset.Type,
		"learningDataset": dataset.LearningDataset,
		"clone":           dataset.Clone,
		"immutable":       dataset.Immutable,
	}

	bytes, err := json.Marshal(source)
	if err != nil {
		return errors.Wrapf(err, "failed to marshal document source")
	}

	// update the document in the metadata index
	_, err = s.client.Index().
		Index(s.datasetIndex).
		Id(dataset.ID).
		BodyString(string(bytes)).
		Refresh("true").
		Do(context.Background())
	if err != nil {
		return errors.Wrapf(err, "failed to update document in index `%s`", s.datasetIndex)
	}
	return nil
}

// IngestDataset adds a document consisting of the metadata to ES.
func (s *Storage) IngestDataset(datasetSource metadata.DatasetSource, meta *model.Metadata) error {

	if len(meta.DataResources) > 1 && meta.SchemaSource != model.SchemaSourceOriginal {
		return errors.New("metadata variables not merged into a single dataset")
	}

	mainDR := meta.GetMainDataResource()

	// clear refers to
	for _, v := range mainDR.Variables {
		v.RefersTo = nil
	}
	var origins []map[string]interface{}
	if meta.DatasetOrigins != nil {
		origins = make([]map[string]interface{}, len(meta.DatasetOrigins))
		for i, ds := range meta.DatasetOrigins {
			origins[i] = map[string]interface{}{
				"searchResult":  ds.SearchResult,
				"provenance":    ds.Provenance,
				"sourceDataset": ds.SourceDataset,
			}
		}
	}

	source := map[string]interface{}{
		"datasetName":      meta.Name,
		"datasetID":        meta.ID,
		"parentDatasetIDs": meta.ParentDatasetIDs,
		"storageName":      meta.StorageName,
		"description":      meta.Description,
		"summary":          meta.Summary,
		"summaryMachine":   meta.SummaryMachine,
		"numRows":          meta.NumRows,
		"numBytes":         meta.NumBytes,
		"variables":        mainDR.Variables,
		"datasetFolder":    meta.DatasetFolder,
		"source":           datasetSource,
		"datasetOrigins":   origins,
		"type":             meta.Type,
		"learningDataset":  meta.LearningDataset,
		"clone":            meta.Clone,
		"immutable":        meta.Immutable,
	}

	bytes, err := json.Marshal(source)
	if err != nil {
		return errors.Wrapf(err, "failed to marshal document source")
	}

	// push the document into the metadata index
	_, err = s.client.Index().
		Index(s.datasetIndex).
		Id(meta.ID).
		BodyString(string(bytes)).
		Refresh("true").
		Do(context.Background())
	if err != nil {
		return errors.Wrapf(err, "failed to add document to index `%s`", s.datasetIndex)
	}
	return nil
}

// DeleteDataset deletes a dataset from ES.
func (s *Storage) DeleteDataset(dataset string) error {
	_, err := s.client.Delete().Index(s.datasetIndex).Id(dataset).Refresh("true").Do(context.Background())

	return err
}

func (s *Storage) parseDatasets(res *elastic.SearchResult, includeIndex bool, includeMeta bool, includeSystemData bool) ([]*api.Dataset, error) {
	var datasets []*api.Dataset
	for _, hit := range res.Hits.Hits {
		// parse hit into JSON
		src, err := json.Unmarshal(hit.Source)
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
			folder = ""
		}
		// extract the learning dataset
		learningDataset, ok := json.String(src, "learningDataset")
		if !ok {
			learningDataset = ""
		}
		// extract the machine learned summary
		summaryMachine, ok := json.String(src, "summaryMachine")
		if !ok {
			summaryMachine = ""
		}
		// extract the type (default to modelling)
		typStr, ok := json.String(src, "type")
		typ := api.DatasetTypeModelling
		if ok && typStr != "" {
			typ = api.DatasetType(typStr)
		}
		// extract the number of rows
		numRows, ok := json.Int(src, "numRows")
		if !ok {
			numRows = 0
		}
		// extract the number of bytes
		numBytes, ok := json.Int(src, "numBytes")
		if !ok {
			numBytes = 0
		}

		// extract the number of bytes
		immutable, ok := json.Bool(src, "immutable")
		if !ok {
			immutable = false
		}
		// extract the number of bytes
		clone, ok := json.Bool(src, "clone")
		if !ok {
			clone = false
		}

		// extract the variables list
		variables, err := s.parseVariables(hit, includeIndex, includeMeta, includeSystemData)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse dataset")
		}
		// extract sources
		source, ok := json.String(src, "source")
		if !ok {
			source = string(metadata.Seed)
		}

		// extract dataset source information
		var datasetOrigins []*api.JoinSuggestion
		if src["datasetOrigins"] != nil {
			origins, ok := json.Array(src, "datasetOrigins")
			if ok {
				datasetOrigins = make([]*api.JoinSuggestion, len(origins))
				for i, origin := range origins {
					searchResult, ok := json.String(origin, "searchResult")
					if !ok {
						searchResult = ""
					}
					searchProvenance, ok := json.String(origin, "provenance")
					if !ok {
						searchProvenance = ""
					}
					sourceDataset, ok := json.String(origin, "sourceDataset")
					if !ok {
						sourceDataset = ""
					}

					datasetOrigins[i] = &api.JoinSuggestion{
						DatasetOrigin: &model.DatasetOrigin{
							SearchResult:  searchResult,
							Provenance:    searchProvenance,
							SourceDataset: sourceDataset,
						},
					}
				}
			}
		}

		// write everythign out to result struct
		datasets = append(datasets, &api.Dataset{
			ID:              id,
			Name:            name,
			StorageName:     storageName,
			Description:     description,
			Folder:          folder,
			Summary:         summary,
			SummaryML:       summaryMachine,
			NumRows:         int64(numRows),
			NumBytes:        int64(numBytes),
			Variables:       variables,
			Provenance:      Provenance,
			Source:          metadata.DatasetSource(source),
			JoinSuggestions: datasetOrigins,
			Type:            typ,
			LearningDataset: learningDataset,
			Immutable:       immutable,
			Clone:           clone,
		})
	}
	return datasets, nil
}

// FetchDatasets returns all datasets in the provided index.
func (s *Storage) FetchDatasets(includeIndex bool, includeMeta bool, includeSystemData bool) ([]*api.Dataset, error) {
	query := elastic.NewBoolQuery().MustNot(elastic.NewTermQuery("type", api.DatasetTypeInference))
	// execute the ES query
	res, err := s.client.Search().
		Query(query).
		Index(s.datasetIndex).
		FetchSource(true).
		Size(datasetsListSize).
		Do(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "elasticsearch dataset fetch query failed")
	}
	return s.parseDatasets(res, includeIndex, includeMeta, includeSystemData)
}

// FetchDataset returns a dataset in the provided index.
func (s *Storage) FetchDataset(datasetName string, includeIndex bool, includeMeta bool, includeSystemData bool) (*api.Dataset, error) {
	query := elastic.NewMatchQuery("_id", datasetName)
	// execute the ES query
	res, err := s.client.Search().
		Query(query).
		Index(s.datasetIndex).
		FetchSource(true).
		Size(datasetsListSize).
		Do(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "elasticsearch dataset fetch query failed")
	}
	datasets, err := s.parseDatasets(res, includeIndex, includeMeta, includeSystemData)
	if err != nil {
		return nil, err
	}
	if len(datasets) < 1 {
		return nil, nil
	}

	return datasets[0], nil
}

// SearchDatasets returns the datasets that match the search criteria in the
// provided index.
func (s *Storage) SearchDatasets(terms string, baseDataset *api.Dataset, includeIndex bool, includeMeta bool, includeSystemData bool) ([]*api.Dataset, error) {
	query := elastic.NewMultiMatchQuery(terms, "_id", "datasetFolder", "datasetID", "datasetName", "variables.colName", "description", "summaryMachine").
		Analyzer("standard")
	// execute the ES query
	res, err := s.client.Search().
		Query(query).
		Index(s.datasetIndex).
		FetchSource(true).
		Size(datasetsListSize).
		Do(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "elasticsearch dataset search query failed")
	}
	return s.parseDatasets(res, includeIndex, includeMeta, includeSystemData)
}

func (s *Storage) updateVariables(dataset string, variables []*model.Variable) error {
	// reserialize the data
	// HACK: DO NOT STORE EXTREMA VALUES AS THERE ARE CURRENTLY ISSUES WITH
	// SCIENTIFIC NOTATION FLOATS SINCE THEY ARE TYPED AS LONG IN ES.
	//   TO REPRODUCE: run startup ingest on SEMI_1217_click_prediction_small
	var serialized []map[string]interface{}
	for _, v := range variables {
		serialized = append(serialized, map[string]interface{}{
			model.VarNameField:             v.HeaderName,
			model.VarStorageNameField:      v.StorageName,
			model.VarIndexField:            v.Index,
			model.VarRoleField:             v.Role,
			model.VarSelectedRoleField:     v.SelectedRole,
			model.VarTypeField:             v.Type,
			model.VarOriginalTypeField:     v.OriginalType,
			model.VarDescriptionField:      v.Description,
			model.VarImportanceField:       v.Importance,
			model.VarSuggestedTypesField:   v.SuggestedTypes,
			model.VarOriginalVariableField: v.OriginalVariable,
			model.VarDisplayVariableField:  v.DisplayName,
			model.VarDistilRole:            v.DistilRole,
			model.VarDeleted:               v.Deleted,
			model.VarGroupingField:         v.Grouping,
			model.VarImmutableField:        v.Immutable,
		})
	}

	source := map[string]interface{}{
		model.Variables: serialized,
	}

	// push the document into the metadata index
	_, err := s.client.Update().
		Index(s.datasetIndex).
		Id(dataset).
		Doc(source).
		Refresh("true").
		Do(context.Background())
	if err != nil {
		return errors.Wrapf(err, "failed to add document to index `%s`", s.datasetIndex)
	}

	return nil
}

// SetDataType updates the data type of the field in ES.
func (s *Storage) SetDataType(dataset string, varName string, varType string) error {
	// Fetch all existing variables
	vars, err := s.FetchVariables(dataset, true, true, false)
	if err != nil {
		return errors.Wrapf(err, "failed to fetch existing variable")
	}

	// Update only the variable we care about
	for _, v := range vars {
		if v.StorageName == varName {
			v.Type = varType
		}
	}

	return s.updateVariables(dataset, vars)
}

// SetExtrema updates the min & max values of a field in ES.
func (s *Storage) SetExtrema(dataset string, varName string, extrema *api.Extrema) error {
	// Fetch all existing variables
	vars, err := s.FetchVariables(dataset, true, true, false)
	if err != nil {
		return errors.Wrapf(err, "failed to fetch existing variable")
	}

	// Update only the variable we care about
	for _, v := range vars {
		if v.StorageName == varName {
			v.Min = extrema.Min
			v.Max = extrema.Max
		}
	}

	return s.updateVariables(dataset, vars)
}

// AddVariable adds a new variable to the dataset.  If the varDisplayName is left blank it will be set to the varName value.
func (s *Storage) AddVariable(dataset string, varName string, varDisplayName string, varType string, varRole string) error {

	if varDisplayName == "" {
		varDisplayName = varName
	}

	// new variable definition
	variable := &model.Variable{
		StorageName:      varName,
		HeaderName:       varName,
		Type:             varType,
		OriginalType:     varType,
		OriginalVariable: varName,
		DisplayName:      varDisplayName,
		DistilRole:       varRole,
		Deleted:          false,
		Immutable:        false,
		SuggestedTypes:   make([]*model.SuggestedType, 0),
	}

	// query for existing variables
	vars, err := s.FetchVariables(dataset, true, true, true)
	if err != nil {
		return errors.Wrapf(err, "failed to fetch existing variable")
	}

	// check if var already exists
	found := false
	for index, v := range vars {
		if v.StorageName == varName {
			// check if it has been deleted
			if !v.Deleted {
				return errors.Errorf("variable already exists under this key")
			}

			// deleted, add the new var in its place
			variable.Index = index
			vars[index] = variable
			found = true
		}
	}

	if !found {
		// add the new variable
		variable.Index = len(vars)
		vars = append(vars, variable)
	}

	return s.updateVariables(dataset, vars)
}

// DeleteVariable flags a variable as deleted.
func (s *Storage) DeleteVariable(dataset string, varName string) error {
	// query for existing variables
	vars, err := s.FetchVariables(dataset, true, true, true)
	if err != nil {
		return errors.Wrapf(err, "failed to fetch existing variable")
	}

	// soft delete the variable
	for _, v := range vars {
		if v.StorageName == varName {
			v.Deleted = true
		}
	}

	return s.updateVariables(dataset, vars)
}

// AddGroupedVariable adds a grouping to the metadata.
func (s *Storage) AddGroupedVariable(dataset string, varName string, varDisplayName string, varType string, varRole string, grouping model.BaseGrouping) error {

	// Create a new grouping variable
	err := s.AddVariable(dataset, varName, varDisplayName, varType, varRole)
	if err != nil {
		return err
	}

	// Add the grouping related info to it.
	query := elastic.NewMatchQuery("_id", dataset)
	// execute the ES query
	res, err := s.client.Search().
		Query(query).
		Index(s.datasetIndex).
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
	source, err := json.Unmarshal(hit.Source)
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
		if ok && name == varName {
			variable[model.VarGroupingField] = json.StructToMap(grouping)
			found = true
		}
	}
	if !found {
		return fmt.Errorf("no variable match found for grouping")
	}

	// push the document into the metadata index
	_, err = s.client.Index().
		Index(s.datasetIndex).
		Id(dataset).
		BodyJson(source).
		Refresh("true").
		Do(context.Background())
	if err != nil {
		return errors.Wrapf(err, "failed to add document to index `%s`", s.datasetIndex)
	}

	return nil
}

// RemoveGroupedVariable removes a grouping to the metadata.
func (s *Storage) RemoveGroupedVariable(datasetName string, grouping model.BaseGrouping) error {

	query := elastic.NewMatchQuery("_id", datasetName)
	// execute the ES query
	res, err := s.client.Search().
		Query(query).
		Index(s.datasetIndex).
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
	source, err := json.Unmarshal(hit.Source)
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
		if ok && name == grouping.GetIDCol() {
			delete(variable, model.VarGroupingField)
			variable["colType"] = variable["colOriginalType"]
			found = true
		}
	}
	if !found {
		return fmt.Errorf("no variable match found for grouping")
	}

	// push the document into the metadata index
	_, err = s.client.Index().
		Index(s.datasetIndex).
		Id(datasetName).
		BodyJson(source).
		Refresh("true").
		Do(context.Background())
	if err != nil {
		return errors.Wrapf(err, "failed to add document to index `%s`", s.datasetIndex)
	}

	return nil
}
