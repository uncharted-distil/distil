package elastic

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil/api/model"
	"github.com/unchartedsoftware/distil/api/util/json"
	"gopkg.in/olivere/elastic.v5"
)

const (
	// DatasetSuffix is the suffix for the dataset entry when stored in
	// elasticsearch.
	DatasetSuffix    = "_dataset"
	metadataType     = "metadata"
	datasetsListSize = 1000
)

func (s *Storage) parseDatasets(res *elastic.SearchResult, includeIndex bool) ([]*model.Dataset, error) {
	var datasets []*model.Dataset
	for _, hit := range res.Hits.Hits {
		// parse hit into JSON
		src, err := json.Unmarshal(*hit.Source)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse dataset")
		}
		// extract dataset id
		name := strings.TrimSuffix(hit.Id, DatasetSuffix)
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
		variables, err := s.parseVariables(hit, includeIndex)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse dataset")
		}
		// write everythign out to result struct
		datasets = append(datasets, &model.Dataset{
			Name:        name,
			Description: description,
			Summary:     summary,
			SummaryML:   summaryMachine,
			NumRows:     int64(numRows),
			NumBytes:    int64(numBytes),
			Variables:   variables,
		})
	}
	return datasets, nil
}

// FetchDatasets returns all datasets in the provided index.
func (s *Storage) FetchDatasets(includeIndex bool) ([]*model.Dataset, error) {
	// execute the ES query
	res, err := s.client.Search().
		Index(s.index).
		FetchSource(true).
		Size(datasetsListSize).
		Do(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "elasticsearch dataset fetch query failed")
	}
	return s.parseDatasets(res, includeIndex)
}

// SearchDatasets returns the datasets that match the search criteria in the
// provided index.
func (s *Storage) SearchDatasets(terms string, includeIndex bool) ([]*model.Dataset, error) {
	query := elastic.NewMultiMatchQuery(terms, "_id", "description", "variables.varName").
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
	return s.parseDatasets(res, includeIndex)
}

func (s *Storage) updateVariables(dataset string, variables []*model.Variable) error {
	// reserialize the data
	var serialized []map[string]interface{}
	for _, v := range variables {
		serialized = append(serialized, map[string]interface{}{
			VarNameField:             v.Name,
			VarRoleField:             v.Role,
			VarTypeField:             v.Type,
			VarImportanceField:       v.Importance,
			VarSuggestedTypesField:   v.SuggestedTypes,
			VarOriginalVariableField: v.OriginalVariable,
			VarDisplayVariableField:  v.DisplayVariable,
			VarDistilRole:            v.DistilRole,
			VarDeleted:               v.Deleted,
		})
	}

	source := map[string]interface{}{
		Variables: serialized,
	}

	// push the document into the metadata index
	_, err := s.client.Update().
		Index(s.index).
		Type(metadataType).
		Id(dataset + DatasetSuffix).
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
	vars, err := s.FetchVariables(dataset, true)
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
	vars, err := s.FetchVariables(dataset, true)
	if err != nil {
		return errors.Wrapf(err, "failed to fetch existing variable")
	}

	// add the new variables
	vars = append(vars, &model.Variable{
		Name:             varName,
		Type:             varType,
		OriginalVariable: varName,
		DisplayVariable:  varName,
		DistilRole:       varRole,
	})

	return s.updateVariables(dataset, vars)
}

// DeleteVariable flags a variable as deleted.
func (s *Storage) DeleteVariable(dataset string, varName string) error {
	// query for existing variables
	vars, err := s.FetchVariables(dataset, true)
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
