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
	DatasetSuffix = "_dataset"
	metadataType  = "metadata"
)

func (s *Storage) parseDatasets(res *elastic.SearchResult) ([]*model.Dataset, error) {
	var datasets []*model.Dataset
	for _, hit := range res.Hits.Hits {
		// parse hit into JSON
		src, err := json.Unmarshal(*hit.Source)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse dataset")
		}
		// extract dataset name (ID is mirror of name)
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
		variables, err := s.parseVariables(hit)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse dataset")
		}
		// write everythign out to result struct
		datasets = append(datasets, &model.Dataset{
			Name:        name,
			Description: description,
			Summary:     summary,
			NumRows:     int64(numRows),
			NumBytes:    int64(numBytes),
			Variables:   variables,
		})
	}
	return datasets, nil
}

// FetchDatasets returns all datasets in the provided index.
func (s *Storage) FetchDatasets(index string) ([]*model.Dataset, error) {
	// execute the ES query
	res, err := s.client.Search().
		Index(index).
		FetchSource(true).
		Do(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "elasticsearch dataset fetch query failed")
	}
	return s.parseDatasets(res)
}

// SearchDatasets returns the datasets that match the search criteria in the
// provided index.
func (s *Storage) SearchDatasets(index string, terms string) ([]*model.Dataset, error) {
	query := elastic.NewMultiMatchQuery(terms, "_id", "description", "variables.varName").
		Analyzer("standard")
	// execute the ES query
	res, err := s.client.Search().
		Query(query).
		Index(index).
		FetchSource(true).
		Do(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "elasticsearch dataset search query failed")
	}
	return s.parseDatasets(res)
}

// SetDataType updates the data type of the field in ES.
func (s *Storage) SetDataType(dataset string, index string, field string, fieldType string) error {
	// Fetch all existing variables
	vars, err := s.FetchVariables(dataset, index, true)
	if err != nil {
		return errors.Wrapf(err, "failed to fetch existing variable")
	}

	// Update only the variable we care about
	for _, v := range vars {
		if v.Name == field {
			v.Type = fieldType
		}
	}

	// re-serialize the vars
	var serialized []map[string]interface{}
	for _, v := range vars {
		serialized = append(serialized, map[string]interface{}{
			VarNameField:             v.Name,
			VarRoleField:             v.Role,
			VarTypeField:             v.Type,
			VarImportanceField:       v.Importance,
			VarSuggestedTypesField:   v.SuggestedTypes,
			VarOriginalVariableField: v.OriginalVariable,
			VarDisplayVariableField:  v.DisplayVariable,
		})
	}

	source := map[string]interface{}{
		Variables: serialized,
	}

	// push the document into the metadata index
	_, err = s.client.Update().
		Index(index).
		Type(metadataType).
		Id(dataset + DatasetSuffix).
		Doc(source).
		Do(context.Background())
	if err != nil {
		return errors.Wrapf(err, "failed to add document to index `%s`", index)
	}
	return nil
}
