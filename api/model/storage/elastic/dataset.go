package elastic

import (
	"context"

	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil-compute/model"
	api "github.com/unchartedsoftware/distil/api/model"
	"github.com/unchartedsoftware/distil/api/util/json"
	"gopkg.in/olivere/elastic.v5"
)

const (
	// DatasetSuffix is the suffix for the dataset entry when stored in
	// elasticsearch.
	metadataType     = "metadata"
	provenance       = "elastic"
	datasetsListSize = 1000
)

// ImportDataset is not supported (ES datasets are already ingested).
func (s *Storage) ImportDataset(uri string) (string, error) {
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
		// extract dataset id
		name := hit.Id
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
		// write everythign out to result struct
		datasets = append(datasets, &api.Dataset{
			Name:        name,
			Description: description,
			Folder:      folder,
			Summary:     summary,
			SummaryML:   summaryMachine,
			NumRows:     int64(numRows),
			NumBytes:    int64(numBytes),
			Variables:   variables,
			Provenance:  provenance,
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
	query := elastic.NewMultiMatchQuery(terms, "_id", "description", "variables.colName", "summaryMachine").
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
