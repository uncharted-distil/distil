package elastic

import (
	"context"

	"github.com/pkg/errors"
	"gopkg.in/olivere/elastic.v5"

	"github.com/unchartedsoftware/distil/api/model"
	"github.com/unchartedsoftware/distil/api/util/json"
)

const (
	// Variables is the field name which stores the variables in elasticsearch.
	Variables = "variables"
	// VarNameField is the field name for the variable name.
	VarNameField = "colName"
	// VarRoleField is the field name for the variable role.
	VarRoleField = "selectedRole"
	// VarDisplayVariableField is the field name for the display variable.
	VarDisplayVariableField = "colDisplayName"
	// VarOriginalVariableField is the field name for the original variable.
	VarOriginalVariableField = "colOriginalName"
	// VarTypeField is the field name for the variable type.
	VarTypeField = "colType"
	// VarImportanceField is the field name for the variable importnace.
	VarImportanceField = "importance"
	// VarSuggestedTypesField is the field name for the suggested variable types.
	VarSuggestedTypesField = "suggestedTypes"
	// VarRoleIndex is the variable role of an index field.
	VarRoleIndex = "index"
)

func (s *Storage) parseRawVariable(child map[string]interface{}) (*model.Variable, error) {
	name, ok := json.String(child, VarNameField)
	if !ok {
		return nil, errors.New("unable to parse name from variable data")
	}
	typ, ok := json.String(child, VarTypeField)
	if !ok {
		return nil, errors.New("unable to parse type from variable data")
	}
	importance, ok := json.Int(child, VarImportanceField)
	if !ok {
		importance = 0
	}
	role, ok := json.String(child, VarRoleField)
	if !ok {
		role = ""
	}
	originalVariable, ok := json.String(child, VarOriginalVariableField)
	if !ok {
		originalVariable = name
	}
	displayVariable, ok := json.String(child, VarDisplayVariableField)
	if !ok {
		displayVariable = ""
	}
	suggestedTypes, ok := json.Array(child, VarSuggestedTypesField)
	if !ok {
		suggestedTypes = make([]map[string]interface{}, 0)
	}

	// default the display name to the normalized name
	if displayVariable == "" {
		displayVariable = name
	}

	return &model.Variable{
		Name:             name,
		Type:             typ,
		Importance:       importance,
		Role:             role,
		SuggestedTypes:   suggestedTypes,
		OriginalVariable: originalVariable,
		DisplayVariable:  displayVariable,
	}, nil
}

func (s *Storage) parseVariable(searchHit *elastic.SearchHit, varName string) (*model.Variable, error) {
	// unmarshal the hit source
	src, err := json.Unmarshal(*searchHit.Source)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse search result")
	}
	// get the variables array
	children, ok := json.Array(src, Variables)
	if !ok {
		return nil, errors.New("unable to parse variables from search result")
	}
	// find the matching var name
	for _, child := range children {
		variable, err := s.parseRawVariable(child)
		if err != nil {
			return nil, errors.Wrap(err, "unable to parse variable")
		}
		if variable != nil {
			if variable.Name == varName {
				return variable, nil
			}
		}
	}
	return nil, errors.Errorf("unable to find variable `%s`", varName)
}

func (s *Storage) parseVariables(searchHit *elastic.SearchHit, includeIndex bool) ([]*model.Variable, error) {
	// unmarshal the hit source
	src, err := json.Unmarshal(*searchHit.Source)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse search result")
	}
	// get the variables array
	children, ok := json.Array(src, Variables)
	if !ok {
		return nil, errors.New("unable to parse variables from search result")
	}
	// for each variable, extract the `colName` and `colType`
	var variables []*model.Variable
	for _, child := range children {
		variable, err := s.parseRawVariable(child)
		if err != nil {
			return nil, errors.Wrap(err, "unable to parse variable")
		}
		if !includeIndex && variable.Role == VarRoleIndex {
			continue
		}
		if variable != nil {
			variables = append(variables, variable)
		}
	}
	return variables, nil
}

// FetchVariable returns the variable for the provided index, dataset, and variable.
func (s *Storage) FetchVariable(dataset string, varName string) (*model.Variable, error) {
	// get dataset id
	datasetID := dataset + DatasetSuffix
	// create match query
	query := elastic.NewMatchQuery("_id", datasetID)
	// create fetch context
	fetchContext := elastic.NewFetchSourceContext(true)
	fetchContext.Include(Variables)
	// execute the ES query
	res, err := s.client.Search().
		Query(query).
		Index(s.index).
		FetchSource(true).
		FetchSourceContext(fetchContext).
		Do(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "elasticSearch variable fetch query failed")
	}
	// check that we have only one hit (should only ever be one matching dataset)
	if len(res.Hits.Hits) != 1 {
		return nil, errors.New("elasticSearch variable fetch query len(hits) != 1")
	}
	// extract output into JSON ready structs
	variables, err := s.parseVariable(res.Hits.Hits[0], varName)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse search result")
	}
	return variables, err
}

// FetchVariableDisplay returns the display variable for the provided index, dataset, and variable.
func (s *Storage) FetchVariableDisplay(dataset string, varName string) (*model.Variable, error) {
	// get the indicated variable.
	variable, err := s.FetchVariable(dataset, varName)
	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch variable")
	}

	// DisplayVariable will identify the variable to return.
	// If not set, no other fetch is needed.
	if variable.DisplayVariable != "" && variable.DisplayVariable != varName {
		return s.FetchVariable(dataset, variable.DisplayVariable)
	}

	return variable, nil
}

// FetchVariables returns all the variables for the provided index and dataset.
func (s *Storage) FetchVariables(dataset string, includeIndex bool) ([]*model.Variable, error) {
	// get dataset id
	datasetID := dataset + DatasetSuffix
	// create match query
	query := elastic.NewMatchQuery("_id", datasetID)
	// create fetch context
	fetchContext := elastic.NewFetchSourceContext(true)
	fetchContext.Include(Variables)
	// execute the ES query
	res, err := s.client.Search().
		Query(query).
		Index(s.index).
		FetchSource(true).
		FetchSourceContext(fetchContext).
		Do(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "elasticSearch variable fetch query failed")
	}
	// check that we have only one hit (should only ever be one matching dataset)
	if len(res.Hits.Hits) != 1 {
		return nil, errors.New("elasticSearch variable fetch query len(hits) != 1")
	}
	// extract output into JSON ready structs
	return s.parseVariables(res.Hits.Hits[0], includeIndex)
}

// FetchVariablesDisplay returns all the display variables for the provided index and dataset.
func (s *Storage) FetchVariablesDisplay(dataset string) ([]*model.Variable, error) {
	// get all variables.
	vars, err := s.FetchVariables(dataset, false)
	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch dataset variables")
	}

	// create a lookup for the variables.
	varsLookup := make(map[string]*model.Variable)
	for _, v := range vars {
		varsLookup[v.Name] = v
	}

	// build the slice by cycling through the variables and using the lookup
	// for the display variables. Only include a variable once.
	resultIncludes := make(map[string]bool)
	result := make([]*model.Variable, 0)
	for _, v := range vars {
		name := v.Name
		if !resultIncludes[name] {
			result = append(result, varsLookup[name])
			resultIncludes[name] = true
		}
	}

	return result, nil
}
