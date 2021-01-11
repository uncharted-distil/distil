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

	"github.com/olivere/elastic/v7"
	"github.com/pkg/errors"
	log "github.com/unchartedsoftware/plog"

	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil/api/util/json"
)

func (s *Storage) parseRawVariable(child map[string]interface{}) (*model.Variable, error) {
	headerName, ok := json.String(child, model.VarNameField)
	if !ok {
		return nil, errors.New("unable to parse header name from variable data")
	}
	key, ok := json.String(child, model.VarKeyField)
	if !ok {
		key = headerName
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
	description, ok := json.String(child, model.VarDescriptionField)
	if !ok {
		description = ""
	}
	importance, ok := json.Float(child, model.VarImportanceField)
	if !ok {
		importance = 0
	}
	role, ok := json.StringArray(child, model.VarRoleField)
	if !ok || role == nil {
		role = make([]string, 0)
	}
	selectedRole, ok := json.String(child, model.VarSelectedRoleField)
	if !ok {
		selectedRole = ""
	}
	originalVariable, ok := json.String(child, model.VarOriginalVariableField)
	if !ok {
		originalVariable = key
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
	immutable, ok := json.Bool(child, model.VarImmutableField)
	if !ok {
		immutable = false
	}
	min, ok := json.Float(child, model.VarMinField)
	if !ok {
		min = 0
	}
	max, ok := json.Float(child, model.VarMaxField)
	if !ok {
		max = 0
	}

	grouping, err := s.parseGrouping(child)
	if err != nil {
		log.Warnf("grouping parsing error: %+v", err)
		grouping = nil
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
		displayVariable = headerName
	}

	return &model.Variable{
		Key:              key,
		HeaderName:       headerName,
		Index:            index,
		Type:             typ,
		OriginalType:     originalType,
		Description:      description,
		Importance:       importance,
		Role:             role,
		SelectedRole:     selectedRole,
		SuggestedTypes:   suggestedTypesParsed,
		OriginalVariable: originalVariable,
		DisplayName:      displayVariable,
		DistilRole:       distilRole,
		Deleted:          deleted,
		Grouping:         grouping,
		Min:              min,
		Max:              max,
		Immutable:        immutable,
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

func (s *Storage) parseVariable(searchHit *elastic.SearchHit, key string) (*model.Variable, error) {
	// unmarshal the hit source
	src, err := json.Unmarshal(searchHit.Source)
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
			if variable.Key == key {
				return variable, nil
			}
		}
	}
	return nil, errors.Errorf("unable to find variable `%s`", key)
}

func (s *Storage) parseVariables(searchHit *elastic.SearchHit, includeIndex bool, includeMeta bool, includeSystemData bool) ([]*model.Variable, error) {
	// unmarshal the hit source
	src, err := json.Unmarshal(searchHit.Source)
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
		if !includeIndex && len(variable.Role) > 0 && model.IsIndexRole(variable.Role[0]) {
			continue
		}
		if !includeMeta && variable.DistilRole == model.VarDistilRoleMetadata {
			continue
		}
		if !includeSystemData && variable.DistilRole == model.VarDistilRoleSystemData {
			continue
		}
		if variable != nil {
			variables = append(variables, variable)
		}
	}

	// hide hidden variables
	var filtered []*model.Variable
	for _, v := range variables {
		if !v.Deleted {
			filtered = append(filtered, v)
		}
	}
	return filtered, nil
}

func (s *Storage) parseGrouping(variable map[string]interface{}) (model.BaseGrouping, error) {
	if !json.Exists(variable, model.VarGroupingField, "type") {
		return nil, nil
	}

	groupingType, ok := json.String(variable, model.VarGroupingField, "type")
	if !ok {
		return nil, errors.New("unable to read grouping type")
	}

	var grouping model.BaseGrouping
	if model.IsTimeSeries(groupingType) {
		grouping = &model.TimeseriesGrouping{}
		ok = json.Struct(variable, grouping, model.VarGroupingField)
		if !ok {
			return nil, errors.New("unable to parse timeseries grouping")
		}
	} else if model.IsGeoCoordinate(groupingType) {
		grouping = &model.GeoCoordinateGrouping{}
		ok = json.Struct(variable, grouping, model.VarGroupingField)
		if !ok {
			return nil, errors.New("unable to parse geocoordinate grouping")
		}
	} else if model.IsMultiBandImage(groupingType) {
		grouping = &model.MultiBandImageGrouping{}
		ok = json.Struct(variable, grouping, model.VarGroupingField)
		if !ok {
			return nil, errors.New("unable to parse remote sensing grouping")
		}
	} else if model.IsGeoBounds(groupingType) {
		grouping = &model.GeoBoundsGrouping{}
		ok = json.Struct(variable, grouping, model.VarGroupingField)
		if !ok {
			return nil, errors.New("unable to parse geobounds sensing grouping")
		}
	} else {
		return nil, errors.Errorf("unrecognized grouping type '%s'", groupingType)
	}
	return grouping, nil
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
		Index(s.datasetIndex).
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
	v, err := s.parseVariable(res.Hits.Hits[0], varName)
	if err != nil {
		return false, nil
	}
	return !v.Deleted, nil
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
		Index(s.datasetIndex).
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
func (s *Storage) FetchVariables(dataset string, includeIndex bool, includeMeta bool, includeSystemData bool) ([]*model.Variable, error) {
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
		Index(s.datasetIndex).
		FetchSource(true).
		FetchSourceContext(fetchContext).
		Do(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "elasticSearch variable fetch query failed")
	}
	// check that we have only one hit (should only ever be one matching dataset)
	if len(res.Hits.Hits) != 1 {
		return nil, errors.Errorf("elasticSearch variable fetch query len(hits) != 1 (len == %d) for dataset '%s'", len(res.Hits.Hits), datasetID)
	}
	// extract output into JSON ready structs
	return s.parseVariables(res.Hits.Hits[0], includeIndex, includeMeta, includeSystemData)
}

// FetchVariablesDisplay returns all the display variables for the provided index and dataset.
func (s *Storage) FetchVariablesDisplay(dataset string) ([]*model.Variable, error) {
	// get all variables.
	vars, err := s.FetchVariables(dataset, false, true, false)
	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch dataset variables")
	}

	// create a lookup for the variables.
	varsLookup := make(map[string]*model.Variable)
	for _, v := range vars {
		varsLookup[v.Key] = v
	}

	// build the slice by cycling through the variables and using the lookup
	// for the display variables. Only include a variable once.
	resultIncludes := make(map[string]bool)
	result := make([]*model.Variable, 0)
	for _, v := range vars {
		name := v.Key
		if !resultIncludes[name] {
			result = append(result, varsLookup[name])
			resultIncludes[name] = true
		}
	}

	return result, nil
}

// FetchVariablesByName returns all the caller supplied variables.
func (s *Storage) FetchVariablesByName(dataset string, varKeys []string, includeIndex bool, includeMeta bool, includeSystemData bool) ([]*model.Variable, error) {
	fetchedVariables, err := s.FetchVariables(dataset, includeIndex, includeMeta, includeSystemData)
	if err != nil {
		return nil, err
	}

	// put the var names into a set for quick lookup
	varKeySet := map[string]bool{}
	for _, key := range varKeys {
		varKeySet[key] = true
	}

	// filter the returned variables to match our input list
	filteredVariables := []*model.Variable{}
	for _, variable := range fetchedVariables {
		if varKeySet[variable.Key] {
			filteredVariables = append(filteredVariables, variable)
		}
	}
	return filteredVariables, nil
}
