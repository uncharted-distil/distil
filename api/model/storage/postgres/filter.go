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
		if variable.Name == key {
			return variable
		}
	}
	return nil
}

func (s *Storage) parseFilteredData(dataset string, variables []*model.Variable, numRows int, rows pgx.Rows) (*api.FilteredData, error) {
	result := &api.FilteredData{
		NumRows: numRows,
		Values:  make([][]*api.FilteredDataValue, 0),
	}

	// Parse the columns.
	if rows != nil {
		// Parse the row data.
		for rows.Next() {
			columnValues, err := rows.Values()
			if err != nil {
				return nil, err
			}

			// filtered data has no weights associated with it
			weightedValues := make([]*api.FilteredDataValue, len(columnValues))
			for i, cv := range columnValues {
				weightedValues[i] = &api.FilteredDataValue{
					Value: cv,
				}
			}

			result.Values = append(result.Values, weightedValues)
		}
		err := rows.Err()
		if err != nil {
			return nil, errors.Wrapf(err, "error reading data from postgres")
		}

		fields := rows.FieldDescriptions()
		columns := make([]*api.Column, len(fields))
		for i := 0; i < len(fields); i++ {
			key := string(fields[i].Name)

			v := getVariableByKey(key, variables)
			if v == nil {
				return nil, fmt.Errorf("unable to lookup variable for %s", key)
			}
			columns[i] = &api.Column{
				Key:   key,
				Label: v.DisplayName,
				Type:  v.Type,
			}
		}
		result.Columns = columns
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
		where := fmt.Sprintf("ST_WITHIN(%s, $%d)", name, len(params)+1)
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
		where := fmt.Sprintf("\"%s\" IN (%s)", model.D3MIndexFieldName, strings.Join(indices, ", "))
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
		name := s.formatFilterKey(alias, g.Name)
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
		where := fmt.Sprintf("ST_WITHIN(%s, $%d)=false", name, len(params)+1)
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
		where := fmt.Sprintf("\"%s\" NOT IN (%s)", model.D3MIndexFieldName, strings.Join(indices, ", "))
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

	highlight := filterParams.Highlight
	if highlight != nil {
		switch highlight.Mode {
		case model.IncludeFilter:
			wheres, params = s.buildIncludeFilter(dataset, wheres, params, alias, highlight)
		case model.ExcludeFilter:
			wheres, params = s.buildExcludeFilter(dataset, wheres, params, alias, highlight)
		}
	}

	var filterWheres []string
	for _, filter := range filterParams.Filters {
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

func (s *Storage) buildFilteredQueryField(variables []*model.Variable, filterVariables []string) (string, error) {

	distincts := make([]string, 0)
	fields := make([]string, 0)
	indexIncluded := false
	for _, variable := range api.GetFilterVariables(filterVariables, variables) {

		if variable.DistilRole == model.VarDistilRoleGrouping {
			distincts = append(distincts, fmt.Sprintf("DISTINCT ON (\"%s\")", variable.Name))
		}

		// derived metadata variables (ex: postgis geometry) should use the original variables
		varName := variable.Name
		if variable.DistilRole == model.VarDistilRoleMetadata && variable.OriginalVariable != variable.Name {
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

func (s *Storage) buildFilteredResultQueryField(variables []*model.Variable, targetVariable *model.Variable, filterVariables []string) (string, []string, error) {

	distincts := make([]string, 0)
	fields := make([]string, 0)
	for _, variable := range api.GetFilterVariables(filterVariables, variables) {

		if strings.Compare(targetVariable.Name, variable.Name) != 0 {

			if variable.DistilRole == model.VarDistilRoleGrouping {
				distincts = append(distincts, fmt.Sprintf("DISTINCT ON (\"%s\")", variable.Name))
			}

			fields = append(fields, fmt.Sprintf("\"%s\"", variable.Name))
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

	targetVariable, err := s.getResultTargetVariable(request.Dataset, request.TargetFeature())
	if err != nil {
		return nil, nil, err
	}

	typedError := getErrorTyped("", targetVariable.Name)

	where := fmt.Sprintf("(%s >= $%d AND %s <= $%d)", typedError, len(params)+1, typedError, len(params)+2)
	params = append(params, *residualFilter.Min)
	params = append(params, *residualFilter.Max)

	// Append the AND clause
	wheres = append(wheres, where)
	return wheres, params, nil
}

func (s *Storage) buildConfidenceResultWhere(wheres []string, params []interface{}, confidenceFilter *model.Filter) ([]string, []interface{}, error) {
	// Add a clause to filter confidence to the existing where
	where := fmt.Sprintf("(result.confidence >= $%d AND result.confidence <= $%d)", len(params)+1, len(params)+2)
	params = append(params, *confidenceFilter.Min)
	params = append(params, *confidenceFilter.Max)

	// Append the AND clause
	wheres = append(wheres, where)
	return wheres, params, nil
}

func (s *Storage) buildPredictedResultWhere(dataset string, wheres []string, params []interface{}, alias string, resultURI string, resultFilter *model.Filter) ([]string, []interface{}, error) {
	// handle the general category case

	filterParams := &api.FilterParams{
		Filters: []*model.Filter{resultFilter},
	}

	wheres, params = s.buildFilteredQueryWhere(dataset, wheres, params, alias, filterParams, false)
	return wheres, params, nil
}

func (s *Storage) buildResultQueryFilters(dataset string, storageName string, resultURI string, filterParams *api.FilterParams, alias string) ([]string, []interface{}, error) {
	// pull filters generated against the result facet out for special handling
	filters := splitFilters(filterParams)

	genericFilterParams := &api.FilterParams{
		Filters: filters.genericFilters,
	}

	// create the filter for the query
	wheres := make([]string, 0)
	params := make([]interface{}, 0)
	wheres, params = s.buildFilteredQueryWhere(dataset, wheres, params, alias, genericFilterParams, false)

	// assemble split filters
	var err error
	if filters.predictedFilter != nil {
		wheres, params, err = s.buildPredictedResultWhere(dataset, wheres, params, alias, resultURI, filters.predictedFilter)
		if err != nil {
			return nil, nil, err
		}
	} else if filters.correctnessFilter != nil {
		wheres, params, err = s.buildCorrectnessResultWhere(wheres, params, storageName, resultURI, filters.correctnessFilter)
		if err != nil {
			return nil, nil, err
		}
	} else if filters.residualFilter != nil {
		wheres, params, err = s.buildErrorResultWhere(wheres, params, filters.residualFilter)
		if err != nil {
			return nil, nil, err
		}
	} else if filters.confidenceFilter != nil {
		wheres, params, err = s.buildConfidenceResultWhere(wheres, params, filters.confidenceFilter)
		if err != nil {
			return nil, nil, err
		}
	}
	return wheres, params, nil
}

type filters struct {
	genericFilters    []*model.Filter
	predictedFilter   *model.Filter
	residualFilter    *model.Filter
	correctnessFilter *model.Filter
	confidenceFilter  *model.Filter
}

func splitFilters(filterParams *api.FilterParams) *filters {
	// Groups filters for handling downstream
	var predictedFilter *model.Filter
	var residualFilter *model.Filter
	var correctnessFilter *model.Filter
	var confidenceFilter *model.Filter
	var remaining []*model.Filter

	if filterParams == nil {
		return &filters{}
	}

	if filterParams.Highlight != nil {
		highlight := filterParams.Highlight
		if api.IsPredictedKey(highlight.Key) {
			predictedFilter = highlight
		} else if api.IsErrorKey(highlight.Key) {
			if highlight.Type == model.NumericalFilter {
				residualFilter = highlight
			} else if highlight.Type == model.CategoricalFilter {
				correctnessFilter = highlight
			}
		} else if api.IsConfidenceKey(highlight.Key) {
			confidenceFilter = highlight
		} else {
			remaining = append(remaining, highlight)
		}
	}

	for _, filter := range filterParams.Filters {
		if api.IsPredictedKey(filter.Key) {
			predictedFilter = filter
		} else if api.IsErrorKey(filter.Key) {
			if filter.Type == model.NumericalFilter {
				residualFilter = filter
			} else if filter.Type == model.CategoricalFilter {
				correctnessFilter = filter
			}
		} else if api.IsConfidenceKey(filter.Key) {
			confidenceFilter = filter
		} else {
			remaining = append(remaining, filter)
		}
	}

	return &filters{
		genericFilters:    remaining,
		predictedFilter:   predictedFilter,
		residualFilter:    residualFilter,
		correctnessFilter: correctnessFilter,
		confidenceFilter:  confidenceFilter,
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

	// match order by for distinct
	var groupings []string
	for _, v := range variables {
		if v.IsGrouping() && v.Grouping.GetIDCol() != "" {
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

func (s *Storage) filterIncludesIndex(filterParams *api.FilterParams) bool {
	for _, v := range filterParams.Filters {
		if v.Key == model.D3MIndexFieldName {
			return true
		}
	}

	return false
}

// FetchData creates a postgres query to fetch a set of rows.  Applies filters to restrict the
// results to a user selected set of fields, with rows further filtered based on allowed ranges and
// categories.
func (s *Storage) FetchData(dataset string, storageName string, filterParams *api.FilterParams, invert bool) (*api.FilteredData, error) {
	variables, err := s.metadata.FetchVariables(dataset, true, true)
	if err != nil {
		return nil, errors.Wrap(err, "Could not pull variables from ES")
	}

	numRows, err := s.FetchNumRows(storageName, variables)
	if err != nil {
		return nil, errors.Wrap(err, "Could not pull num rows")
	}

	// if there are no filters, and we are returning the exclude set, we expect
	// no results in the filtered set
	if invert && filterParams.Filters == nil {
		return &api.FilteredData{
			NumRows: numRows,
			Columns: make([]*api.Column, 0),
			Values:  make([][]*api.FilteredDataValue, 0),
		}, nil
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
	for _, v := range variables {
		if v.IsGrouping() && v.Grouping.GetIDCol() != "" {
			groupings = append(groupings, "\""+v.Grouping.GetIDCol()+"\"")
		}
	}
	groupings = append(groupings, "\""+model.D3MIndexFieldName+"\"")
	orderBy := strings.Join(groupings, ",")

	// order & limit the filtered data.
	query = fmt.Sprintf("SELECT * FROM (%s ORDER BY %s) data ORDER BY random()", query, orderBy)
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

	// parse the result
	filteredData, err := s.parseFilteredData(dataset, variables, numRows, res)
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
