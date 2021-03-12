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

package postgres

import (
	"fmt"
	"strings"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/model"
	api "github.com/uncharted-distil/distil/api/model"
	log "github.com/unchartedsoftware/plog"
)

const (
	// CorrectCategory identifies the correct result meta-category.
	CorrectCategory = "correct"

	// IncorrectCategory identifies the incorrect result meta-category.
	IncorrectCategory = "incorrect"
)

var (
	pgRandomSeed = 0.2
)

// SetRandomSeed sets the random seed to use when reading a subset of data from the database.
func SetRandomSeed(seed float64) {
	pgRandomSeed = seed
}

func getVariableByKey(key string, variables []*model.Variable) *model.Variable {
	for _, variable := range variables {
		if variable.IsGrouping() && variable.Grouping.GetIDCol() == key {
			return variable
		}
		if variable.Key == key && variable.DistilRole != model.VarDistilRoleGrouping {
			return variable
		}
	}
	return nil
}

func (s *Storage) parseFilteredData(dataset string, filterVariables []*model.Variable, numRows int, includeGroupingCol bool, rows pgx.Rows) (*api.FilteredData, error) {
	result := &api.FilteredData{
		NumRows: numRows,
		Values:  make([][]*api.FilteredDataValue, 0),
	}

	if rows != nil {

		// Parse the columns.  We can potentially have multiple variables map to the same result
		// (timeries variables that use the same grouping column) so we iterate over the filter variable
		// list to find any that map.
		fields := rows.FieldDescriptions()
		columns := []*api.Column{}
		fieldIndexMap := []int{}
		for _, variable := range filterVariables {
			// loop through the filter vars and find the key associated with each
			for fieldIdx, f := range fields {
				fieldKey := string(f.Name)
				if variable.IsGrouping() && variable.Grouping.GetIDCol() == fieldKey {
					columns = append(columns, &api.Column{
						Key:   variable.Key,
						Label: variable.DisplayName,
						Type:  variable.Type,
					})
					fieldIndexMap = append(fieldIndexMap, fieldIdx)
				} else if fieldKey == variable.Key && (includeGroupingCol || variable.DistilRole != model.VarDistilRoleGrouping) {
					columns = append(columns, &api.Column{
						Key:   variable.Key,
						Label: variable.DisplayName,
						Type:  variable.Type,
					})
					fieldIndexMap = append(fieldIndexMap, fieldIdx)
				}
			}
		}
		result.Columns = columns

		// Parse the row data.
		for rows.Next() {
			columnValues, err := rows.Values()
			if err != nil {
				return nil, err
			}

			// filtered data has no weights associated with it
			// we use the field index map to ensure that the column structure and row data structures
			// align
			weightedValues := make([]*api.FilteredDataValue, len(fieldIndexMap))
			for colIdx, fieldIdx := range fieldIndexMap {
				parsedValue, err := parsePostgresType(columnValues[fieldIdx], fields[fieldIdx])
				if err != nil {
					return nil, err
				}
				weightedValues[colIdx] = &api.FilteredDataValue{
					Value: parsedValue,
				}
			}

			result.Values = append(result.Values, weightedValues)
		}
		err := rows.Err()
		if err != nil {
			return nil, errors.Wrapf(err, "error reading data from postgres")
		}
	} else {
		result.Columns = make([]*api.Column, 0)
	}

	return result, nil
}

func (s *Storage) formatFilterKey(alias string, key string) string {
	if api.IsResultKey(key) {
		return "result.value"
	}
	return getFullName(alias, key)
}

func featureVarName(varName string) string {
	return fmt.Sprintf("%s%s", model.ClusterVarPrefix, varName)
}

func (s *Storage) buildIncludeFilter(dataset string, wheres []string, params []interface{}, alias string, filter *model.Filter) ([]string, []interface{}) {

	name := s.formatFilterKey(alias, filter.Key)

	switch filter.Type {
	case model.DatetimeFilter:
		// datetime
		// extract epoch for comparison
		where := fmt.Sprintf("cast(extract(epoch from %s) as double precision) >= $%d AND cast(extract(epoch from %s) as double precision) < $%d", name, len(params)+1, name, len(params)+2)
		wheres = append(wheres, where)
		params = append(params, *filter.Min)
		params = append(params, *filter.Max)

	case model.NumericalFilter:
		// numerical
		// cast to double precision in case of string based representation
		where := fmt.Sprintf("cast(%s as double precision) >= $%d AND cast(%s as double precision) < $%d", name, len(params)+1, name, len(params)+2)
		wheres = append(wheres, where)
		params = append(params, *filter.Min)
		params = append(params, *filter.Max)

	case model.VectorFilter:
		// vector
		// cast to double precision array in case of string based representation
		nestedCast := ""
		if filter.NestedType == model.NumericalFilter {
			nestedCast = "double precision"
		}
		where := fmt.Sprintf("%s @> CAST(ARRAY[$%d, $%d] AS %s[])", name, len(params)+1, len(params)+2, nestedCast)
		wheres = append(wheres, where)
		params = append(params, *filter.Min)
		params = append(params, *filter.Max)

	case model.BivariateFilter:
		// bivariate
		// cast to double precision in case of string based representation
		fields, err := s.getBivariateFilterKeys(dataset, filter.Key, alias)
		if err != nil {
			log.Warnf("%+v", err)
		} else {
			where := fmt.Sprintf("cast(%s as double precision) >= $%d AND cast(%s as double precision) < $%d AND cast(%s as double precision) >= $%d AND cast(%s as double precision) < $%d",
				fields[0], len(params)+1, fields[0], len(params)+2, fields[1], len(params)+3, fields[1], len(params)+4)
			wheres = append(wheres, where)
			params = append(params, filter.Bounds.MinX)
			params = append(params, filter.Bounds.MaxX)
			params = append(params, filter.Bounds.MinY)
			params = append(params, filter.Bounds.MaxY)
		}

	case model.CategoricalFilter:
		// categorical
		categories := make([]string, 0)
		offset := len(params) + 1
		for i, category := range filter.Categories {
			categories = append(categories, fmt.Sprintf("$%d", offset+i))
			if category != "<none>" {
				params = append(params, category)
			} else {
				params = append(params, "")
			}
		}
		where := fmt.Sprintf("%s IN (%s)", name, strings.Join(categories, ", "))
		wheres = append(wheres, where)

	case model.GeoBoundsFilter:
		// geo bounds
		where := fmt.Sprintf("ST_INTERSECTS(%s, $%d)", name, len(params)+1)
		params = append(params, buildBoundsGeometryString(filter.Bounds))
		wheres = append(wheres, where)

	case model.ClusterFilter:
		// cluster
		name = s.formatFilterKey(alias, featureVarName(filter.Key))
		categories := make([]string, 0)
		offset := len(params) + 1
		for i, category := range filter.Categories {
			categories = append(categories, fmt.Sprintf("$%d", offset+i))
			params = append(params, category)
		}
		where := fmt.Sprintf("%s IN (%s)", name, strings.Join(categories, ", "))
		wheres = append(wheres, where)
	case model.RowFilter:
		// row
		indices := make([]string, 0)
		offset := len(params) + 1
		for i, d3mIndex := range filter.D3mIndices {
			indices = append(indices, fmt.Sprintf("$%d", offset+i))
			params = append(params, d3mIndex)
		}
		where := fmt.Sprintf("%s IN (%s)", name, strings.Join(indices, ", "))
		wheres = append(wheres, where)
	case model.TextFilter:
		// text
		offset := len(params) + 1
		for i, category := range filter.Categories {
			where := fmt.Sprintf("%s ~* (%s)", name, fmt.Sprintf("$%d", offset+i))
			params = append(params, category)
			wheres = append(wheres, where)
		}
	}
	return wheres, params
}

func (s *Storage) getBivariateFilterKeys(dataset string, key string, alias string) ([]string, error) {

	fields := make([]string, 2)

	// assume the name is a grouping and get it
	g, err := s.metadata.FetchVariable(dataset, key)
	if err != nil {
		return nil, err
	}

	if model.IsGeoBounds(g.Type) {
		// only checking top left for now
		name := s.formatFilterKey(alias, g.Key)
		fields[0] = fmt.Sprintf("%s[1]", name)
		fields[1] = fmt.Sprintf("%s[2]", name)
		return fields, nil
	}

	if g.IsGrouping() && model.IsGeoCoordinate(g.Grouping.GetType()) {
		cg := g.Grouping.(*model.GeoCoordinateGrouping)
		fields[0] = s.formatFilterKey(alias, cg.XCol)
		fields[1] = s.formatFilterKey(alias, cg.YCol)
		return fields, nil
	}

	return nil, errors.Errorf("unsupported field type %s for bivariate filter", g.Type)
}

func (s *Storage) buildExcludeFilter(dataset string, wheres []string, params []interface{}, alias string, filter *model.Filter) ([]string, []interface{}) {

	name := s.formatFilterKey(alias, filter.Key)

	switch filter.Type {
	case model.DatetimeFilter:
		// datetime
		// extract epoch for comparison
		where := fmt.Sprintf("cast(extract(epoch from %s) as double precision) < $%d OR cast(extract(epoch from %s) as double precision) >= $%d", name, len(params)+1, name, len(params)+2)
		wheres = append(wheres, where)
		params = append(params, *filter.Min)
		params = append(params, *filter.Max)

	case model.NumericalFilter:
		// numerical
		//TODO: WHY DOES THIS QUERY NOT CAST TO DOUBLE LIKE THE INCLUDE???
		where := fmt.Sprintf("(%s < $%d OR %s >= $%d)", name, len(params)+1, name, len(params)+2)
		wheres = append(wheres, where)
		params = append(params, *filter.Min)
		params = append(params, *filter.Max)

	case model.VectorFilter:
		// vector
		// cast to double precision array in case of string based representation
		nestedCast := ""
		if filter.NestedType == model.NumericalFilter {
			nestedCast = "double precision"
		}
		where := fmt.Sprintf("NOT(%s @> CAST(ARRAY[$%d, $%d] AS %s[]))", name, len(params)+1, len(params)+2, nestedCast)
		wheres = append(wheres, where)
		params = append(params, *filter.Min)
		params = append(params, *filter.Max)

	case model.BivariateFilter:
		// bivariate
		// cast to double precision in case of string based representation
		fields, err := s.getBivariateFilterKeys(dataset, filter.Key, alias)
		if err != nil {
			log.Warnf("%+v", err)
		} else {
			where := fmt.Sprintf("(cast(%s as double precision) < $%d OR cast(%s as double precision) >= $%d) OR (cast(%s as double precision) < $%d OR cast(%s as double precision) >= $%d)",
				fields[0], len(params)+1, fields[0], len(params)+2, fields[1], len(params)+3, fields[1], len(params)+4)
			wheres = append(wheres, where)
			params = append(params, filter.Bounds.MinX)
			params = append(params, filter.Bounds.MaxX)
			params = append(params, filter.Bounds.MinY)
			params = append(params, filter.Bounds.MaxY)
		}

	case model.CategoricalFilter:
		// categorical
		categories := make([]string, 0)
		offset := len(params) + 1
		for i, category := range filter.Categories {
			categories = append(categories, fmt.Sprintf("$%d", offset+i))
			params = append(params, category)
		}
		where := fmt.Sprintf("%s NOT IN (%s)", name, strings.Join(categories, ", "))
		wheres = append(wheres, where)

	case model.GeoBoundsFilter:
		// geo bounds
		where := fmt.Sprintf("ST_INTERSECTS(%s, $%d)=false", name, len(params)+1)
		params = append(params, buildBoundsGeometryString(filter.Bounds))
		wheres = append(wheres, where)

	case model.ClusterFilter:
		// cluster
		name = s.formatFilterKey(alias, featureVarName(filter.Key))
		categories := make([]string, 0)
		offset := len(params) + 1
		for i, category := range filter.Categories {
			categories = append(categories, fmt.Sprintf("$%d", offset+i))
			params = append(params, category)
		}
		where := fmt.Sprintf("%s NOT IN (%s)", name, strings.Join(categories, ", "))
		wheres = append(wheres, where)
	case model.RowFilter:
		// row
		indices := make([]string, 0)
		offset := len(params) + 1
		for i, d3mIndex := range filter.D3mIndices {
			indices = append(indices, fmt.Sprintf("$%d", offset+i))
			params = append(params, d3mIndex)
		}
		where := fmt.Sprintf("%s NOT IN (%s)", name, strings.Join(indices, ", "))
		wheres = append(wheres, where)
	case model.TextFilter:
		// text
		offset := len(params) + 1
		for i, category := range filter.Categories {
			where := fmt.Sprintf("%s !~* (%s)", name, fmt.Sprintf("$%d", offset+i))
			params = append(params, category)
			wheres = append(wheres, where)
		}
	}
	return wheres, params
}

func (s *Storage) buildFilteredQueryWhere(dataset string, wheres []string, params []interface{}, alias string, filterParams *api.FilterParams, invert bool) ([]string, []interface{}) {

	if filterParams == nil {
		return wheres, params
	}

	var highlightWheres []string
	for _, highlight := range filterParams.Highlights.List {
		switch highlight.Mode {
		case model.IncludeFilter:
			highlightWheres, params = s.buildIncludeFilter(dataset, highlightWheres, params, alias, highlight)
		case model.ExcludeFilter:
			highlightWheres, params = s.buildExcludeFilter(dataset, highlightWheres, params, alias, highlight)
		}
	}
	if len(highlightWheres) > 0 {
		where := ""
		if invert {
			// highlights are always treated as or (adding Not(...or...) makes it and)
			where = fmt.Sprintf("(%s)", strings.Join(highlightWheres, " OR "))
		} else {
			where = fmt.Sprintf("(%s)", strings.Join(highlightWheres, " OR "))
		}
		wheres = append(wheres, where)
	}

	var filterWheres []string
	for _, filter := range filterParams.Filters.List {
		switch filter.Mode {
		case model.IncludeFilter:
			filterWheres, params = s.buildIncludeFilter(dataset, filterWheres, params, alias, filter)
		case model.ExcludeFilter:
			filterWheres, params = s.buildExcludeFilter(dataset, filterWheres, params, alias, filter)
		}
	}
	if len(filterWheres) > 0 {
		where := ""
		if invert {
			where = fmt.Sprintf("NOT(%s)", strings.Join(filterWheres, " AND "))
		} else {
			where = strings.Join(filterWheres, " AND ")
		}
		wheres = append(wheres, where)
	}
	return wheres, params
}
func (s *Storage) buildSelectStatement(variables []*model.Variable, filterVariables []string) (string, error) {

	distincts := make([]string, 0)
	fields := make([]string, 0)
	indexIncluded := false
	for _, variable := range api.GetFilterVariables(filterVariables, variables) {
		if variable.IsGrouping() {
			continue
		}

		// derived metadata variables (ex: postgis geometry) should use the original variables
		varName := variable.Key
		if variable.DistilRole == model.VarDistilRoleMetadata && variable.OriginalVariable != variable.Key {
			varName = variable.OriginalVariable
		}

		fields = append(fields, fmt.Sprintf("\"%s\"", varName))
		if varName == model.D3MIndexFieldName {
			indexIncluded = true
		}

	}
	// if the index is not already in the field list, then append it
	if !indexIncluded {
		fields = append(fields, fmt.Sprintf("\"%s\"", model.D3MIndexFieldName))
	}
	return strings.Join(distincts, ",") + " " + strings.Join(fields, ","), nil
}
func (s *Storage) buildFilteredQueryField(variables []*model.Variable, filterVariables []string) (string, error) {

	distincts := make([]string, 0)
	fields := make([]string, 0)
	indexIncluded := false
	for _, variable := range api.GetFilterVariables(filterVariables, variables) {
		if variable.IsGrouping() {
			continue
		}

		if variable.DistilRole == model.VarDistilRoleGrouping {
			distincts = append(distincts, fmt.Sprintf("DISTINCT ON (\"%s\")", variable.Key))
		}

		// derived metadata variables (ex: postgis geometry) should use the original variables
		varKey := variable.Key
		if variable.DistilRole == model.VarDistilRoleMetadata && variable.OriginalVariable != variable.Key {
			varKey = variable.OriginalVariable
		}

		fields = append(fields, fmt.Sprintf("\"%s\"", varKey))
		if varKey == model.D3MIndexFieldName {
			indexIncluded = true
		}

	}
	// if the index is not already in the field list, then append it
	if !indexIncluded {
		fields = append(fields, fmt.Sprintf("\"%s\"", model.D3MIndexFieldName))
	}
	return strings.Join(distincts, ",") + " " + strings.Join(fields, ","), nil
}

func (s *Storage) buildFilteredResultQueryField(variables []*model.Variable, targetVariable *model.Variable, filterVariables []string) (string, []string, error) {

	distincts := make([]string, 0)
	fields := make([]string, 0)
	groupingCols := map[string]bool{}
	for _, variable := range api.GetFilterVariables(filterVariables, variables) {
		if variable.IsGrouping() {
			continue
		}

		if strings.Compare(targetVariable.Key, variable.Key) != 0 {

			if variable.DistilRole == model.VarDistilRoleGrouping && !groupingCols[variable.Key] {
				groupingCols[variable.Key] = true // don't duplicate columns in our distinct
				distincts = append(distincts, fmt.Sprintf("DISTINCT ON (\"%s\")", variable.Key))
			}

			fields = append(fields, fmt.Sprintf("\"%s\"", variable.Key))
		}
	}
	fields = append(fields, fmt.Sprintf("\"%s\"", model.D3MIndexFieldName))
	return strings.Join(distincts, ","), fields, nil
}

func (s *Storage) buildCorrectnessResultWhere(wheres []string, params []interface{}, storageName string, resultURI string, resultFilter *model.Filter) ([]string, []interface{}, error) {
	// get the target variable name
	storageNameResult := s.getResultTable(storageName)
	targetName, err := s.getResultTargetName(storageNameResult, resultURI)
	if err != nil {
		return nil, nil, err
	}

	// correct/incorrect are well known categories that require the predicted category to be compared
	// to the target category
	op := ""
	for _, category := range resultFilter.Categories {
		if strings.EqualFold(category, CorrectCategory) {
			op = "="
			break
		} else if strings.EqualFold(category, IncorrectCategory) {
			op = "!="
			break
		}
	}
	if op == "" {
		return nil, nil, err
	}
	where := fmt.Sprintf("result.value %s data.\"%s\"", op, targetName)
	wheres = append(wheres, where)
	return wheres, params, nil
}

func (s *Storage) buildErrorResultWhere(wheres []string, params []interface{}, residualFilter *model.Filter) ([]string, []interface{}, error) {
	// Add a clause to filter residuals to the existing where

	// Error keys are a string of the form <solutionID>:error.  We need to pull the solution ID out so we can find the name of the target var.
	solutionID := api.StripKeySuffix(residualFilter.Key)

	request, err := s.FetchRequestBySolutionID(solutionID)
	if err != nil {
		return nil, nil, err
	}

	// Fetch the target variable.  For grouped variables, the target will be one of the component
	// variables.
	targetVariable, err := s.getResultTargetVariable(request.Dataset, request.TargetFeature())
	if err != nil {
		return nil, nil, err
	}
	if targetVariable.IsGrouping() && model.IsTimeSeries(targetVariable.Grouping.GetType()) {
		tsg := targetVariable.Grouping.(*model.TimeseriesGrouping)
		targetVariable, err = s.getResultTargetVariable(request.Dataset, tsg.YCol)
		if err != nil {
			return nil, nil, err
		}
	}

	typedError := getErrorTyped("", targetVariable.Key)

	where := fmt.Sprintf("(%s >= $%d AND %s <= $%d)", typedError, len(params)+1, typedError, len(params)+2)
	params = append(params, *residualFilter.Min)
	params = append(params, *residualFilter.Max)

	// Append the AND clause
	wheres = append(wheres, where)
	return wheres, params, nil
}

func (s *Storage) buildConfidenceResultWhere(wheres []string, params []interface{}, confidenceFilter *model.Filter, alias string) ([]string, []interface{}) {
	// Add a clause to filter confidence to the existing where
	if alias != "" {
		alias = alias + "."
	}
	where := fmt.Sprintf("((%sexplain_values -> 'confidence')::double precision >= $%d AND (%sexplain_values -> 'confidence')::double precision <= $%d)", alias, len(params)+1, alias, len(params)+2)
	params = append(params, *confidenceFilter.Min)
	params = append(params, *confidenceFilter.Max)

	// Append the AND clause
	wheres = append(wheres, where)
	return wheres, params
}
func (s *Storage) buildRankResultWhere(wheres []string, params []interface{}, rankFilter *model.Filter, alias string) ([]string, []interface{}) {
	// Add a clause to filter confidence to the existing where
	if alias != "" {
		alias = alias + "."
	}
	where := fmt.Sprintf("((%sexplain_values -> 'rank')::double precision >= $%d AND (%sexplain_values -> 'rank')::double precision <= $%d)", alias, len(params)+1, alias, len(params)+2)
	params = append(params, *rankFilter.Min)
	params = append(params, *rankFilter.Max)

	// Append the AND clause
	wheres = append(wheres, where)
	return wheres, params
}
func (s *Storage) buildPredictedResultWhere(dataset string, wheres []string, params []interface{}, alias string, resultURI string, resultFilter *model.Filter) ([]string, []interface{}) {
	// handle the general category case

	filterParams := &api.FilterParams{
		Filters: api.FilterObject{List: []*model.Filter{resultFilter},
			Invert: false,
		},
	}

	wheres, params = s.buildFilteredQueryWhere(dataset, wheres, params, alias, filterParams, false)
	return wheres, params
}

func (s *Storage) buildResultQueryFilters(dataset string, storageName string, resultURI string, filterParams *api.FilterParams, alias string) ([]string, []interface{}, error) {
	// pull filters generated against the result facet out for special handling
	filters := splitFilters(filterParams)

	genericFilterParams := &api.FilterParams{
		Filters: filters.genericFilters,
	}
	genericFilterParams.Highlights = filters.genericHighlights
	// create the filter for the query
	wheres := make([]string, 0)
	params := make([]interface{}, 0)
	wheres, params = s.buildFilteredQueryWhere(dataset, wheres, params, alias, genericFilterParams, false)

	// assemble split filters
	var err error
	for _, predictedFilter := range filters.predictedFilters.List {
		wheres, params = s.buildPredictedResultWhere(dataset, wheres, params, alias, resultURI, predictedFilter)
	}
	for _, correctnessFilter := range filters.correctnessFilters.List {
		wheres, params, err = s.buildCorrectnessResultWhere(wheres, params, storageName, resultURI, correctnessFilter)
		if err != nil {
			return nil, nil, err
		}
	}
	for _, residualFilter := range filters.residualFilters.List {
		wheres, params, err = s.buildErrorResultWhere(wheres, params, residualFilter)
		if err != nil {
			return nil, nil, err
		}
	}
	for _, confidenceFilter := range filters.confidenceFilters.List {
		wheres, params = s.buildConfidenceResultWhere(wheres, params, confidenceFilter, "result")
	}
	for _, rankFilter := range filters.rankFilters.List {
		wheres, params = s.buildRankResultWhere(wheres, params, rankFilter, "result")
	}
	return wheres, params, nil
}

type filters struct {
	genericFilters     api.FilterObject
	genericHighlights  api.FilterObject
	predictedFilters   api.FilterObject
	residualFilters    api.FilterObject
	correctnessFilters api.FilterObject
	confidenceFilters  api.FilterObject
	rankFilters        api.FilterObject
}

func splitFilters(filterParams *api.FilterParams) *filters {
	// Groups filters for handling downstream
	var predictedFilters api.FilterObject
	var residualFilters api.FilterObject
	var correctnessFilters api.FilterObject
	var confidenceFilters api.FilterObject
	var rankFilters api.FilterObject
	var remaining api.FilterObject
	var remainingHighlights api.FilterObject

	if filterParams == nil {
		return &filters{}
	}

	for _, highlight := range filterParams.Highlights.List {
		if api.IsPredictedKey(highlight.Key) {
			predictedFilters.List = append(predictedFilters.List, highlight)
		} else if api.IsErrorKey(highlight.Key) {
			if highlight.Type == model.NumericalFilter {
				residualFilters.List = append(residualFilters.List, highlight)
			} else if highlight.Type == model.CategoricalFilter {
				correctnessFilters.List = append(correctnessFilters.List, highlight)
			}
		} else if api.IsConfidenceKey(highlight.Key) {
			confidenceFilters.List = append(confidenceFilters.List, highlight)
		} else if api.IsRankKey(highlight.Key) {
			rankFilters.List = append(rankFilters.List, highlight)
		} else {
			remainingHighlights.List = append(remainingHighlights.List, highlight)
		}
	}

	for _, filter := range filterParams.Filters.List {
		if api.IsPredictedKey(filter.Key) {
			predictedFilters.List = append(predictedFilters.List, filter)
		} else if api.IsErrorKey(filter.Key) {
			if filter.Type == model.NumericalFilter {
				residualFilters.List = append(residualFilters.List, filter)
			} else if filter.Type == model.CategoricalFilter {
				correctnessFilters.List = append(correctnessFilters.List, filter)
			}
		} else if api.IsConfidenceKey(filter.Key) {
			confidenceFilters.List = append(confidenceFilters.List, filter)
		} else if api.IsRankKey(filter.Key) {
			rankFilters.List = append(rankFilters.List, filter)
		} else {
			remaining.List = append(remaining.List, filter)
		}
	}

	return &filters{
		genericFilters:     remaining,
		genericHighlights:  remainingHighlights,
		predictedFilters:   predictedFilters,
		residualFilters:    residualFilters,
		correctnessFilters: correctnessFilters,
		confidenceFilters:  confidenceFilters,
		rankFilters:        rankFilters,
	}
}

// FetchNumRows pulls the number of rows in the table.
func (s *Storage) FetchNumRows(storageName string, variables []*model.Variable) (int, error) {

	return s.fetchNumRowsJoined(storageName, variables, nil, nil, nil)
}

// FetchNumRowsFiltered pulls the number of filtered rows in the table.
func (s *Storage) FetchNumRowsFiltered(storageName string, variables []*model.Variable, filters []string, params []interface{}) (int, error) {

	return s.fetchNumRowsJoined(storageName, variables, filters, params, nil)
}

// fetchNumRowsJoined pulls the number of rows in the table.
func (s *Storage) fetchNumRowsJoined(storageName string, variables []*model.Variable, filters []string, params []interface{}, join *joinDefinition) (int, error) {

	countTarget := "*"

	// ensure distinct ordering matches order by
	groupings := []string{}
	groupingSet := map[string]bool{}
	for _, v := range variables {
		if v.IsGrouping() && v.Grouping.GetIDCol() != "" && !groupingSet[v.Grouping.GetIDCol()] {
			groupingSet[v.Grouping.GetIDCol()] = true
			groupings = append(groupings, v.Grouping.GetIDCol())
		}
	}

	if len(groupings) > 0 {
		countTarget = "DISTINCT "
		for i, g := range groupings {
			countTarget += "\"" + g + "\""
			if len(groupings)-1 > i {
				countTarget += ", "
			}
		}
	}

	joinSQL := ""
	tableAlias := "base_data"
	if join != nil {
		tableAlias = join.baseAlias
		joinSQL = getJoinSQL(join, true)
	}

	query := fmt.Sprintf("SELECT count(%s) FROM %s AS %s %s", countTarget, storageName, tableAlias, joinSQL)
	if len(filters) > 0 {
		query = fmt.Sprintf("%s WHERE %s", query, strings.Join(filters, " AND "))
	}
	var numRows int
	err := s.client.QueryRow(query, params...).Scan(&numRows)
	if err != nil {
		return -1, errors.Wrap(err, "postgres row query failed")
	}
	return numRows, nil
}

// FetchData creates a postgres query to fetch a set of rows.  Applies filters to restrict the
// results to a user selected set of fields, with rows further filtered based on allowed ranges and
// categories.
func (s *Storage) FetchData(dataset string, storageName string, filterParams *api.FilterParams, invert bool, includeGroupingCol bool, orderByVar *model.Variable) (*api.FilteredData, error) {
	variables, err := s.metadata.FetchVariables(dataset, true, true, true)
	if err != nil {
		return nil, errors.Wrap(err, "Could not pull variables from ES")
	}

	numRows, err := s.FetchNumRows(storageName, variables)
	if err != nil {
		return nil, errors.Wrap(err, "Could not pull num rows")
	}

	// if there are no filters, and we are returning the exclude set, we expect
	// no results in the filtered set
	if invert && filterParams.Filters.List == nil {
		return &api.FilteredData{
			NumRows: numRows,
			Columns: make([]*api.Column, 0),
			Values:  make([][]*api.FilteredDataValue, 0),
		}, nil
	}
	selectStatement, err := s.buildSelectStatement(variables, filterParams.Variables)
	if err != nil {
		return nil, errors.Wrap(err, "Could not build select statement")
	}
	// standard order by
	orderByClause := "random()"
	if orderByVar != nil {
		// if exist change order by clause
		orderByClause = "\"" + orderByVar.HeaderName + "\" DESC"
		// check if the order by variable exists in the supplied list of vars
		existInFilter := api.GetFilterVariables(filterParams.Variables, []*model.Variable{orderByVar})
		if len(existInFilter) == 0 {
			// if it does not exist add it for the inner query (in order to sort from the outer query)
			filterParams.Variables = append(filterParams.Variables, orderByVar.HeaderName)
		}
	}
	fields, err := s.buildFilteredQueryField(variables, filterParams.Variables)
	if err != nil {
		return nil, errors.Wrap(err, "Could not build field list")
	}

	// construct a Postgres query that fetches documents from the dataset with the supplied variable filters applied
	batch := &pgx.Batch{}
	batch.Queue(fmt.Sprintf("SELECT setseed(%v);", pgRandomSeed))
	query := fmt.Sprintf(" SELECT %s FROM %s", fields, storageName)

	wheres := make([]string, 0)
	params := make([]interface{}, 0)
	wheres, params = s.buildFilteredQueryWhere(dataset, wheres, params, "", filterParams, invert)

	if len(wheres) > 0 {
		query = fmt.Sprintf("%s WHERE %s", query, strings.Join(wheres, " AND "))
	}

	// match order by for distinct
	var groupings []string
	includedGroupings := map[string]bool{}
	for _, v := range variables {
		if v.IsGrouping() && v.Grouping.GetIDCol() != "" && !includedGroupings[v.Grouping.GetIDCol()] {
			includedGroupings[v.Grouping.GetIDCol()] = true
			groupings = append(groupings, "\""+v.Grouping.GetIDCol()+"\"")
		}
	}
	groupings = append(groupings, "\""+model.D3MIndexFieldName+"\"")
	orderBy := strings.Join(groupings, ",")
	// order & limit the filtered data.
	query = fmt.Sprintf("SELECT %s FROM (%s ORDER BY %s) data ORDER BY %s", selectStatement, query, orderBy, orderByClause)
	if filterParams.Size > 0 {
		query = fmt.Sprintf("%s LIMIT %d", query, filterParams.Size)
	}
	query = query + ";"
	// execute the postgres query
	batch.Queue(query, params...)
	resBatch := s.client.SendBatch(batch)
	defer resBatch.Close()
	_, err = resBatch.Exec()
	if err != nil {
		return nil, errors.Wrap(err, "postgres filtered data set seed query failed")
	}
	res, err := resBatch.Query()
	if err != nil {
		return nil, errors.Wrap(err, "postgres filtered data query failed")
	}
	if res != nil {
		defer res.Close()
	}

	// get the list of variables that are include by our current filter state
	filterVariables, err := s.metadata.FetchVariablesByName(dataset, filterParams.Variables, true, false, false)
	if err != nil {
		return nil, err
	}

	// parse the result
	filteredData, err := s.parseFilteredData(dataset, filterVariables, numRows, includeGroupingCol, res)
	if err != nil {
		return nil, err
	}

	// Add the num filtered rows
	numRowsFiltered, err := s.FetchNumRowsFiltered(storageName, variables, wheres, params)
	if err != nil {
		return nil, errors.Wrap(err, "Could not pull filtered num rows")
	}
	filteredData.NumRowsFiltered = numRowsFiltered

	return filteredData, nil
}
