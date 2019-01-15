package elastic

import (
	"context"

	"github.com/pkg/errors"
	"gopkg.in/olivere/elastic.v5"

	"github.com/unchartedsoftware/distil-compute/model"
	"github.com/unchartedsoftware/distil/api/util/json"
)

func (s *Storage) parseRawVariable(child map[string]interface{}) (*model.Variable, error) {
	name, ok := json.String(child, model.VarNameField)
	if !ok {
		return nil, errors.New("unable to parse name from variable data")
	}
	index, ok := json.Int(child, model.VarIndexField)
	if !ok {
		return nil, errors.New("unable to parse index from variable data")
	}
	typ, ok := json.String(child, model.VarTypeField)
	if !ok {
		return nil, errors.New("unable to parse type from variable data")
	}
	originalType, ok := json.String(child, model.VarOriginalTypeField)
	if !ok {
		return nil, errors.New("unable to parse original type from variable data")
	}
	importance, ok := json.Int(child, model.VarImportanceField)
	if !ok {
		importance = 0
	}
	role, ok := json.StringArray(child, model.VarRoleField)
	if !ok {
		role = make([]string, 0)
	}
	selectedRole, ok := json.String(child, model.VarSelectedRoleField)
	if !ok {
		selectedRole = ""
	}
	originalVariable, ok := json.String(child, model.VarOriginalVariableField)
	if !ok {
		originalVariable = name
	}
	displayVariable, ok := json.String(child, model.VarDisplayVariableField)
	if !ok {
		displayVariable = ""
	}
	distilRole, ok := json.String(child, model.VarDistilRole)
	if !ok {
		distilRole = ""
	}
	deleted, ok := json.Bool(child, model.VarDeleted)
	if !ok {
		deleted = false
	}
	suggestedTypes, ok := json.Array(child, model.VarSuggestedTypesField)
	suggestedTypesParsed := make([]*model.SuggestedType, 0)
	if ok {
		for _, t := range suggestedTypes {
			suggestedType, err := s.parseSuggestedType(t)
			if err != nil {
				return nil, errors.Wrap(err, "unable to parse suggested type")
			}
			suggestedTypesParsed = append(suggestedTypesParsed, suggestedType)
		}
	}

	// default the display name to the normalized name
	if displayVariable == "" {
		displayVariable = name
	}

	return &model.Variable{
		Name:             name,
		Index:            index,
		Type:             typ,
		OriginalType:     originalType,
		Importance:       importance,
		Role:             role,
		SelectedRole:     selectedRole,
		SuggestedTypes:   suggestedTypesParsed,
		OriginalVariable: originalVariable,
		DisplayName:      displayVariable,
		DistilRole:       distilRole,
		Deleted:          deleted,
	}, nil
}

func (s *Storage) parseSuggestedType(json map[string]interface{}) (*model.SuggestedType, error) {
	typ, ok := json[model.TypeTypeField].(string)
	if !ok {
		return nil, errors.New("unable to parse type from suggested type data")
	}
	probability, ok := json[model.TypeProbabilityField].(float64)
	if !ok {
		return nil, errors.New("unable to parse probability from suggested type data")
	}
	provenance, ok := json[model.TypeProvenanceField].(string)
	if !ok {
		return nil, errors.New("unable to parse provenance from suggested type data")
	}

	return &model.SuggestedType{
		Type:        typ,
		Probability: probability,
		Provenance:  provenance,
	}, nil
}

func (s *Storage) parseVariable(searchHit *elastic.SearchHit, varName string) (*model.Variable, error) {
	// unmarshal the hit source
	src, err := json.Unmarshal(*searchHit.Source)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse search result")
	}
	// get the variables array
	children, ok := json.Array(src, model.Variables)
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

func (s *Storage) parseVariables(searchHit *elastic.SearchHit, includeIndex bool, includeMeta bool) ([]*model.Variable, error) {
	// unmarshal the hit source
	src, err := json.Unmarshal(*searchHit.Source)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse search result")
	}
	// get the variables array
	children, ok := json.Array(src, model.Variables)
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
		if !includeIndex && len(variable.Role) > 0 && variable.Role[0] == model.VarRoleIndex {
			continue
		}
		if !includeMeta && variable.DistilRole == model.VarRoleMetadata {
			continue
		}
		if variable != nil {
			variables = append(variables, variable)
		}
	}
	return variables, nil
}

// DoesVariableExist returns whether or not a variable exists.
func (s *Storage) DoesVariableExist(dataset string, varName string) (bool, error) {
	// get dataset id
	datasetID := dataset
	// create match query
	query := elastic.NewMatchQuery("_id", datasetID)
	// create fetch context
	fetchContext := elastic.NewFetchSourceContext(true)
	fetchContext.Include(model.Variables)
	// execute the ES query
	res, err := s.client.Search().
		Query(query).
		Index(s.index).
		FetchSource(true).
		FetchSourceContext(fetchContext).
		Do(context.Background())
	if err != nil {
		return false, errors.Wrap(err, "elasticSearch variable fetch query failed")
	}
	// check that we have only one hit (should only ever be one matching dataset)
	if len(res.Hits.Hits) == 0 {
		return false, nil
	}
	if len(res.Hits.Hits) > 1 {
		return false, errors.New("elasticSearch variable fetch query len(hits) > 1")
	}
	_, err = s.parseVariable(res.Hits.Hits[0], varName)
	if err != nil {
		return false, nil
	}
	return true, nil
}

// FetchVariable returns the variable for the provided index, dataset, and variable.
func (s *Storage) FetchVariable(dataset string, varName string) (*model.Variable, error) {
	// get dataset id
	datasetID := dataset
	// create match query
	query := elastic.NewMatchQuery("_id", datasetID)
	// create fetch context
	fetchContext := elastic.NewFetchSourceContext(true)
	fetchContext.Include(model.Variables)
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
	variable, err := s.parseVariable(res.Hits.Hits[0], varName)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse search result")
	}
	return variable, err
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
	if variable.DisplayName != "" && variable.DisplayName != varName {
		return s.FetchVariable(dataset, variable.DisplayName)
	}

	return variable, nil
}

// FetchVariables returns all the variables for the provided index and dataset.
func (s *Storage) FetchVariables(dataset string, includeIndex bool, includeMeta bool) ([]*model.Variable, error) {
	// get dataset id
	datasetID := dataset
	// create match query
	query := elastic.NewMatchQuery("_id", datasetID)
	// create fetch context
	fetchContext := elastic.NewFetchSourceContext(true)
	fetchContext.Include(model.Variables)
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
	return s.parseVariables(res.Hits.Hits[0], includeIndex, includeMeta)
}

// FetchVariablesDisplay returns all the display variables for the provided index and dataset.
func (s *Storage) FetchVariablesDisplay(dataset string) ([]*model.Variable, error) {
	// get all variables.
	vars, err := s.FetchVariables(dataset, false, true)
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
