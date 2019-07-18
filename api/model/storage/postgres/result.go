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
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/model"
	api "github.com/uncharted-distil/distil/api/model"
	log "github.com/unchartedsoftware/plog"
)

func (s *Storage) getResultTable(dataset string) string {
	return fmt.Sprintf("%s_result", dataset)
}

func (s *Storage) getResultTargetName(storageName string, resultURI string) (string, error) {
	// Assume only a single target / result. Read the target name from the
	// database table.
	sql := fmt.Sprintf("SELECT target FROM %s WHERE result_id = $1 LIMIT 1;", storageName)

	rows, err := s.client.Query(sql, resultURI)
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("Unable to get target variable name from results for result URI `%s`", resultURI))
	}
	defer rows.Close()

	if rows.Next() {
		var targetName string
		err = rows.Scan(&targetName)
		if err != nil {
			return "", errors.Wrap(err, fmt.Sprintf("Unable to get target variable name for result URI `%s`", resultURI))
		}

		return targetName, nil
	}

	return "", errors.Errorf("Target feature for result URI `%s` not found", resultURI)
}

func (s *Storage) getResultTargetVariable(dataset string, targetName string) (*model.Variable, error) {
	variable, err := s.metadata.FetchVariable(dataset, targetName)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to get target variable information")
	}

	return variable, nil
}

// PersistResult stores the solution result to Postgres.
func (s *Storage) PersistResult(dataset string, storageName string, resultURI string, target string) error {
	// Read the results file.
	file, err := os.Open(resultURI)
	if err != nil {
		return errors.Wrap(err, "unable open solution result file")
	}
	csvReader := csv.NewReader(bufio.NewReader(file))
	csvReader.TrimLeadingSpace = true
	records, err := csvReader.ReadAll()
	if err != nil {
		return errors.Wrap(err, "unable load solution result as csv")
	}
	if len(records) <= 0 || len(records[0]) <= 0 {
		return errors.Wrap(err, "solution csv empty")
	}

	// currently only support a single result column.
	if len(records[0]) > 2 {
		log.Warnf("Result contains %d columns, expected 2.  Additional columns will be ignored.", len(records[0]))
	}

	targetName := target

	// Translate from display name to storage name.
	targetDisplayName, err := s.getDisplayName(dataset, targetName)
	if err != nil {
		return errors.Wrap(err, "unable to map target name")
	}

	// A target that is a grouping won't have the correct name for now.  We need to check to see
	// if the target is a grouping, and use the Y Col as the target for purposes of dealing with the TA2.
	targetVariable, err := s.getResultTargetVariable(dataset, target)
	if targetVariable.Grouping != nil && model.IsTimeSeries(targetVariable.Grouping.Type) {
		// extract the time series value column
		targetName = targetVariable.Grouping.Properties.YCol
		targetVariable, err = s.metadata.FetchVariable(dataset, targetName)
		if err != nil {
			return err
		}
		targetDisplayName = targetVariable.DisplayName
		if err != nil {
			return err
		}
	}

	// Header row will have the target. Find the index.
	targetIndex := -1
	d3mIndexIndex := -1
	for i, v := range records[0] {
		if v == targetDisplayName {
			targetIndex = i
		} else if v == model.D3MIndexFieldName {
			d3mIndexIndex = i
		}
	}
	// result is not in valid format - d3mIndex and target col need to have correct name
	if targetIndex == -1 {
		return errors.Wrapf(err, "unable to find target col '%s' in result header", targetDisplayName)
	}
	if d3mIndexIndex == -1 {
		return errors.Wrapf(err, "unabled to find d3m index col '%s' in result header", model.D3MIndexFieldName)
	}

	// build the batch data
	insertData := make([][]interface{}, 0)
	for i := 1; i < len(records); i++ {
		// Each data row is index, target.
		// handle the parsed result/error - should be an int some TA2 systems return floats
		parsedVal, err := strconv.ParseInt(records[i][d3mIndexIndex], 10, 64)
		if err != nil {
			parsedValFloat, err := strconv.ParseFloat(records[i][d3mIndexIndex], 64)
			if err != nil {
				return errors.Wrap(err, "failed csv index parsing")
			}
			parsedVal = int64(parsedValFloat)
		}

		insertData = append(insertData, []interface{}{resultURI, parsedVal, target, records[i][targetIndex]})
	}

	// store all results to the storage
	err = s.InsertBatch(s.getResultTable(storageName), []string{"result_id", "index", "target", "value"}, insertData)
	if err != nil {
		return errors.Wrap(err, "failed to insert result in database")
	}

	return nil
}

func (s *Storage) executeInsertResultStatement(storageName string, resultID string, index int64, target string, value string) error {
	statement := fmt.Sprintf("INSERT INTO %s (result_id, index, target, value) VALUES ($1, $2, $3, $4);", s.getResultTable(storageName))

	_, err := s.client.Exec(statement, resultID, index, target, value)

	return err
}

func (s *Storage) parseFilteredResults(variables []*model.Variable, numRows int, rows *pgx.Rows, target *model.Variable) (*api.FilteredData, error) {
	result := &api.FilteredData{
		NumRows: numRows,
		Values:  make([][]interface{}, 0),
	}

	// Parse the columns.
	if rows != nil {
		fields := rows.FieldDescriptions()
		columns := make([]api.Column, len(fields))
		for i := 0; i < len(fields); i++ {
			key := fields[i].Name
			label := key
			typ := "unknown"
			if api.IsPredictedKey(key) {
				label = "Predicted " + api.StripKeySuffix(key)
				typ = target.Type
			} else if api.IsErrorKey(key) {
				label = "Error"
				typ = target.Type
			} else {
				v := getVariableByKey(key, variables)
				if v != nil {
					typ = v.Type
				}
			}

			columns[i] = api.Column{
				Key:   key,
				Label: label,
				Type:  typ,
			}
		}

		// Result type provided by DB needs to be overridden with defined target type.
		columns[0].Type = target.Type

		// Parse the row data.
		for rows.Next() {
			columnValues, err := rows.Values()
			if err != nil {
				return nil, errors.Wrap(err, "Unable to extract fields from query result")
			}
			result.Values = append(result.Values, columnValues)
			result.Columns = columns
		}
	} else {
		result.Columns = make([]api.Column, 0)
	}

	return result, nil
}

func appendAndClause(expression string, andClause string) string {
	if expression == "" {
		return andClause
	}
	if andClause == "" {
		return andClause
	}
	return fmt.Sprintf("%s AND %s", expression, andClause)
}

func isCorrectnessCategory(categoryName string) bool {
	return strings.EqualFold(CorrectCategory, categoryName) || strings.EqualFold(categoryName, IncorrectCategory)
}

func addIncludeCorrectnessFilterToWhere(wheres []string, params []interface{}, correctnessFilter *model.Filter, target *model.Variable) ([]string, []interface{}, error) {
	if len(correctnessFilter.Categories[0]) == 0 {
		return nil, nil, fmt.Errorf("no category")
	}
	// filter for result correctness which is based on well know category values
	where := ""
	op := ""
	if strings.EqualFold(correctnessFilter.Categories[0], CorrectCategory) {
		op = "="
	} else if strings.EqualFold(correctnessFilter.Categories[0], IncorrectCategory) {
		op = "!="
	}
	where = fmt.Sprintf("predicted.value %s data.\"%s\"", op, target.Name)
	wheres = append(wheres, where)
	return wheres, params, nil
}

func addExcludeCorrectnessFilterToWhere(wheres []string, params []interface{}, correctnessFilter *model.Filter, target *model.Variable) ([]string, []interface{}, error) {
	// filter for result correctness which is based on well know category values
	if len(correctnessFilter.Categories[0]) == 0 {
		return nil, nil, fmt.Errorf("no category")
	}
	where := ""
	op := ""
	if strings.EqualFold(correctnessFilter.Categories[0], CorrectCategory) {
		op = "!="
	} else if strings.EqualFold(correctnessFilter.Categories[0], IncorrectCategory) {
		op = "="
	}
	where = fmt.Sprintf("predicted.value %s data.\"%s\"", op, target.Name)
	wheres = append(wheres, where)
	return wheres, params, nil
}

func addIncludePredictedFilterToWhere(wheres []string, params []interface{}, predictedFilter *model.Filter, target *model.Variable) ([]string, []interface{}, error) {
	// Handle the predicted column, which is accessed as `value` in the result query
	where := ""
	switch predictedFilter.Type {
	case model.NumericalFilter:
		// numerical range-based filter
		where = fmt.Sprintf("cast(predicted.value AS double precision) >= $%d AND cast(predicted.value AS double precision) <= $%d", len(params)+1, len(params)+2)
		params = append(params, *predictedFilter.Min)
		params = append(params, *predictedFilter.Max)

	case model.BivariateFilter:
		// cast to double precision in case of string based representation
		// hardcode [lat, lon] format for now
		where := fmt.Sprintf("predicted.value[2] >= $%d AND predicted.value[2] <= $%d predicted.value[1] >= $%d AND predicted.value[1] <= $%d", len(params)+1, len(params)+2, len(params)+3, len(params)+4)
		wheres = append(wheres, where)
		params = append(params, predictedFilter.Bounds.MinX)
		params = append(params, predictedFilter.Bounds.MaxX)
		params = append(params, predictedFilter.Bounds.MinY)
		params = append(params, predictedFilter.Bounds.MaxY)

	case model.CategoricalFilter:
		// categorical label based filter, with checks for special correct/incorrect metafilters
		categories := make([]string, 0)
		offset := len(params) + 1

		for i, category := range predictedFilter.Categories {
			if !isCorrectnessCategory(category) {
				categories = append(categories, fmt.Sprintf("$%d", offset+i))
				params = append(params, category)
			}
		}

		if len(categories) >= 1 {
			where = fmt.Sprintf("predicted.value IN (%s)", strings.Join(categories, ", "))
		}

	case model.RowFilter:
		// row index based filter
		indices := make([]string, 0)
		offset := len(params) + 1
		for i, d3mIndex := range predictedFilter.D3mIndices {
			indices = append(indices, fmt.Sprintf("$%d", offset+i))
			params = append(params, d3mIndex)

		}
		if len(indices) >= 1 {
			where = fmt.Sprintf("predicted.value IN (%s)", strings.Join(indices, ", "))
		}

	default:
		return nil, nil, errors.Errorf("unexpected type %s for variable %s", predictedFilter.Type, predictedFilter.Key)
	}

	// Append the AND clause
	wheres = append(wheres, where)
	return wheres, params, nil
}

func addExcludePredictedFilterToWhere(wheres []string, params []interface{}, predictedFilter *model.Filter, target *model.Variable) ([]string, []interface{}, error) {
	// Handle the predicted column, which is accessed as `value` in the result query
	where := ""
	switch predictedFilter.Type {
	case model.NumericalFilter:
		// numerical range-based filter
		where = fmt.Sprintf("(cast(predicted.value AS double precision) < $%d OR cast(predicted.value AS double precision) > $%d)", len(params)+1, len(params)+2)
		params = append(params, *predictedFilter.Min)
		params = append(params, *predictedFilter.Max)

	case model.BivariateFilter:
		// bivariate
		// cast to double precision in case of string based representation
		// hardcode [lat, lon] format for now
		where := fmt.Sprintf("(predicted.value[2] < $%d OR predicted.value[2] > $%d) OR (predicted.value[1] < $%d OR predicted.value[1] > $%d)", len(params)+1, len(params)+2, len(params)+3, len(params)+4)
		wheres = append(wheres, where)
		params = append(params, predictedFilter.Bounds.MinX)
		params = append(params, predictedFilter.Bounds.MaxX)
		params = append(params, predictedFilter.Bounds.MinY)
		params = append(params, predictedFilter.Bounds.MaxY)

	case model.CategoricalFilter:
		// categorical label based filter, with checks for special correct/incorrect metafilters
		categories := make([]string, 0)
		offset := len(params) + 1

		for i, category := range predictedFilter.Categories {
			if !isCorrectnessCategory(category) {
				categories = append(categories, fmt.Sprintf("$%d", offset+i))
				params = append(params, category)
			}
		}

		if len(categories) >= 1 {
			where = fmt.Sprintf("predicted.value NOT IN (%s)", strings.Join(categories, ", "))
		}

	case model.RowFilter:
		// row index based filter
		indices := make([]string, 0)
		offset := len(params) + 1
		for i, d3mIndex := range predictedFilter.D3mIndices {
			indices = append(indices, fmt.Sprintf("$%d", offset+i))
			params = append(params, d3mIndex)

		}
		if len(indices) >= 1 {
			where = fmt.Sprintf("predicted.value NOT IN (%s)", strings.Join(indices, ", "))
		}

	default:
		return nil, nil, errors.Errorf("unexpected type %s for variable %s", predictedFilter.Type, predictedFilter.Key)
	}

	// Append the AND clause
	wheres = append(wheres, where)
	return wheres, params, nil
}

func addIncludeErrorFilterToWhere(wheres []string, params []interface{}, targetName string, residualFilter *model.Filter) ([]string, []interface{}, error) {
	// Add a clause to filter residuals to the existing where
	typedError := getErrorTyped(targetName)
	where := fmt.Sprintf("(%s >= $%d AND %s <= $%d)", typedError, len(params)+1, typedError, len(params)+2)
	params = append(params, *residualFilter.Min)
	params = append(params, *residualFilter.Max)

	// Append the AND clause
	wheres = append(wheres, where)
	return wheres, params, nil
}

func addExcludeErrorFilterToWhere(wheres []string, params []interface{}, targetName string, residualFilter *model.Filter) ([]string, []interface{}, error) {
	// Add a clause to filter residuals to the existing where
	typedError := getErrorTyped(targetName)
	where := fmt.Sprintf("(%s < $%d OR %s > $%d)", typedError, len(params)+1, typedError, len(params)+2)
	params = append(params, *residualFilter.Min)
	params = append(params, *residualFilter.Max)

	// Append the AND clause
	wheres = append(wheres, where)
	return wheres, params, nil
}

// FetchResults pulls the results from the Postgres database.
func (s *Storage) FetchResults(dataset string, storageName string, resultURI string, solutionID string, filterParams *api.FilterParams) (*api.FilteredData, error) {
	storageNameResult := s.getResultTable(storageName)
	targetName, err := s.getResultTargetName(storageNameResult, resultURI)
	if err != nil {
		return nil, err
	}

	// fetch the variable info to resolve its type - skip the first column since that will be the d3m_index value
	variable, err := s.getResultTargetVariable(dataset, targetName)
	if err != nil {
		return nil, err
	}

	// fetch variable metadata
	variables, err := s.metadata.FetchVariables(dataset, false, false, false)
	if err != nil {
		return nil, errors.Wrap(err, "Could not pull variables from ES")
	}

	// generate variable list for inclusion in query select
	distincts, fields, err := s.buildFilteredResultQueryField(variables, variable, filterParams.Variables)
	if err != nil {
		return nil, errors.Wrap(err, "Could not build field list")
	}

	// break filters out groups for specific handling
	filters := s.splitFilters(filterParams)

	genericFilterParams := &api.FilterParams{
		Filters: filters.genericFilters,
	}

	// if filterParams != nil {
	// 	genericFilterParams.Highlight = filterParams.Highlight
	// }

	// Create the filter portion of the where clause.
	wheres := make([]string, 0)
	params := make([]interface{}, 0)
	wheres, params = s.buildFilteredQueryWhere(wheres, params, genericFilterParams, false)

	// Add the predicted filter into the where clause if it was included in the filter set
	if filters.predictedFilter != nil {
		if filters.predictedFilter.Mode == model.IncludeFilter {
			wheres, params, err = addIncludePredictedFilterToWhere(wheres, params, filters.predictedFilter, variable)
			if err != nil {
				return nil, errors.Wrap(err, "Could not add result to where clause")
			}
		} else {
			wheres, params, err = addExcludePredictedFilterToWhere(wheres, params, filters.predictedFilter, variable)
			if err != nil {
				return nil, errors.Wrap(err, "Could not add result to where clause")
			}
		}
	}

	// Add the correctness filter into the where clause if it was included in the filter set
	if filters.correctnessFilter != nil {
		if filters.correctnessFilter.Mode == model.IncludeFilter {
			wheres, params, err = addIncludeCorrectnessFilterToWhere(wheres, params, filters.correctnessFilter, variable)
			if err != nil {
				return nil, errors.Wrap(err, "Could not add result to where clause")
			}
		} else {
			wheres, params, err = addExcludeCorrectnessFilterToWhere(wheres, params, filters.correctnessFilter, variable)
			if err != nil {
				return nil, errors.Wrap(err, "Could not add result to where clause")
			}
		}
	}

	// Add the error filter into the where clause if it was included in the filter set
	if filters.residualFilter != nil {
		if filters.residualFilter.Mode == model.IncludeFilter {
			wheres, params, err = addIncludeErrorFilterToWhere(wheres, params, targetName, filters.residualFilter)
			if err != nil {
				return nil, errors.Wrap(err, "Could not add error to where clause")
			}
		} else {
			wheres, params, err = addExcludeErrorFilterToWhere(wheres, params, targetName, filters.residualFilter)
			if err != nil {
				return nil, errors.Wrap(err, "Could not add error to where clause")
			}
		}
	}

	// If our results are numerical we need to compute residuals and store them in a column called 'error'
	predictedCol := api.GetPredictedKey(targetName, solutionID)
	errorCol := api.GetErrorKey(targetName, solutionID)
	targetCol := targetName

	errorExpr := ""
	if model.IsNumerical(variable.Type) {
		errorExpr = fmt.Sprintf("%s as \"%s\",", getErrorTyped(variable.Name), errorCol)
	}

	query := fmt.Sprintf(
		"SELECT %s predicted.value as \"%s\", "+
			"\"%s\" as \"%s\", "+
			"%s "+
			"%s "+
			"FROM %s as predicted inner join %s as data on data.\"%s\" = predicted.index "+
			"WHERE result_id = $%d AND target = $%d",
		distincts, predictedCol, targetName, targetCol, errorExpr, fields, storageNameResult, storageName,
		model.D3MIndexFieldName, len(params)+1, len(params)+2)

	params = append(params, resultURI)
	params = append(params, targetName)

	if len(wheres) > 0 {
		query = fmt.Sprintf("%s AND %s", query, strings.Join(wheres, " AND "))
	}

	// Do not return the whole result set to the client.
	query = fmt.Sprintf("%s LIMIT %d;", query, filterParams.Size)

	rows, err := s.client.Query(query, params...)
	if err != nil {
		return nil, errors.Wrap(err, "Error querying results")
	}
	defer rows.Close()

	countFilter := map[string]interface{}{
		"result_id": resultURI,
	}
	joinDef := &joinDefinition{
		baseColumn:    model.D3MIndexFieldName,
		joinTableName: storageNameResult,
		joinAlias:     "joined",
		joinColumn:    "index",
	}
	numRows, err := s.fetchNumRowsJoined(storageNameResult, variables, countFilter, joinDef)
	if err != nil {
		return nil, errors.Wrap(err, "Could not pull num rows")
	}

	return s.parseFilteredResults(variables, numRows, rows, variable)
}

func (s *Storage) getResultMinMaxAggsQuery(variable *model.Variable, resultVariable *model.Variable) string {
	// get min / max agg names
	minAggName := api.MinAggPrefix + resultVariable.Name
	maxAggName := api.MaxAggPrefix + resultVariable.Name

	// Only numeric types should occur.
	fieldTyped := fmt.Sprintf("cast(\"%s\" as double precision)", resultVariable.Name)

	// create aggregations
	queryPart := fmt.Sprintf("MIN(%s) AS \"%s\", MAX(%s) AS \"%s\"", fieldTyped, minAggName, fieldTyped, maxAggName)
	// add aggregations
	return queryPart
}

func (s *Storage) fetchResultsExtrema(resultURI string, dataset string, variable *model.Variable, resultVariable *model.Variable) (*api.Extrema, error) {
	// add min / max aggregation
	aggQuery := s.getResultMinMaxAggsQuery(variable, resultVariable)

	// create a query that does min and max aggregations for each variable
	queryString := fmt.Sprintf("SELECT %s FROM %s WHERE result_id = $1 AND target = $2;", aggQuery, dataset)

	// execute the postgres query
	res, err := s.client.Query(queryString, resultURI, variable.Name)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch extrema for result from postgres")
	}
	defer res.Close()

	return s.parseExtrema(res, variable)
}

// FetchResultsExtremaByURI fetches the results extrema by resultURI.
func (s *Storage) FetchResultsExtremaByURI(dataset string, storageName string, resultURI string) (*api.Extrema, error) {
	storageNameResult := s.getResultTable(storageName)
	targetName, err := s.getResultTargetName(storageNameResult, resultURI)
	if err != nil {
		return nil, err
	}
	targetVariable, err := s.getResultTargetVariable(dataset, targetName)
	if err != nil {
		return nil, err
	}
	resultVariable := &model.Variable{
		Name: "value",
		Type: model.StringType,
	}

	field := NewNumericalField(s, storageName, targetVariable.Name, targetVariable.DisplayName, targetVariable.Type)
	return field.fetchResultsExtrema(resultURI, storageNameResult, resultVariable)
}

// FetchPredictedSummary gets the summary data about a target variable from the
// results table.
func (s *Storage) FetchPredictedSummary(dataset string, storageName string, resultURI string, filterParams *api.FilterParams, extrema *api.Extrema) (*api.VariableSummary, error) {
	storageNameResult := s.getResultTable(storageName)
	targetName, err := s.getResultTargetName(storageNameResult, resultURI)
	if err != nil {
		return nil, err
	}

	variable, err := s.getResultTargetVariable(dataset, targetName)
	if err != nil {
		return nil, err
	}

	var field Field

	if variable.Grouping != nil {
		if model.IsTimeSeries(variable.Grouping.Type) {

			timeColVar, err := s.metadata.FetchVariable(dataset, variable.Grouping.Properties.XCol)
			if err != nil {
				return nil, errors.Wrap(err, "failed to fetch variable description for summary")
			}

			field = NewTimeSeriesField(s, storageName, variable.Grouping.Properties.ClusterCol, variable.Grouping.IDCol, variable.Grouping.IDCol, variable.Grouping.Type, timeColVar.Name, timeColVar.Type)

		} else {
			return nil, errors.Errorf("variable grouping `%s` of type `%s` does not support summary", variable.Grouping.IDCol, variable.Grouping.Type)
		}

	} else {

		// use the variable type to guide the summary creation

		if model.IsNumerical(variable.Type) {
			field = NewNumericalField(s, storageName, variable.Name, variable.DisplayName, variable.Type)
		} else if model.IsCategorical(variable.Type) {
			field = NewCategoricalField(s, storageName, variable.Name, variable.DisplayName, variable.Type)
		} else if model.IsVector(variable.Type) {
			field = NewVectorField(s, storageName, variable.Name, variable.DisplayName, variable.Type)
		} else {
			return nil, errors.Errorf("variable %s of type %s does not support summary", variable.Name, variable.Type)
		}
	}

	summary, err := field.FetchPredictedSummaryData(resultURI, storageNameResult, filterParams, extrema)
	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch result summary")
	}

	// add dataset
	summary.Dataset = dataset

	return summary, nil
}

func (s *Storage) getDisplayName(dataset string, columnName string) (string, error) {
	displayName := ""
	variables, err := s.metadata.FetchVariables(dataset, false, false, false)
	if err != nil {
		return "", errors.Wrap(err, "unable fetch variables for name mapping")
	}

	for _, v := range variables {
		if v.Name == columnName {
			displayName = v.DisplayName
		}
	}

	return displayName, nil
}

// FetchForecastingSummary fetches a timeseries for a given result.
func (s *Storage) FetchForecastingSummary(dataset string, storageName string, xColName string, yColName string, interval int, resultURI string, filterParams *api.FilterParams) (*api.VariableSummary, error) {

	// need description of the variables to request aggregation against.
	timeColVar, err := s.metadata.FetchVariable(dataset, xColName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch variable description for summary")
	}
	variable, err := s.metadata.FetchVariable(dataset, yColName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch variable description for summary")
	}

	// get the histogram by using the variable type.
	var field Field

	if variable.Grouping != nil {
		return nil, errors.Errorf("not implemented")
	}

	if model.IsNumerical(variable.Type) {
		field = NewNumericalField(s, storageName, variable.Name, variable.DisplayName, variable.Type)
	} else if model.IsCategorical(variable.Type) {
		field = NewCategoricalField(s, storageName, variable.Name, variable.DisplayName, variable.Type)
	} else if model.IsVector(variable.Type) {
		field = NewVectorField(s, storageName, variable.Name, variable.DisplayName, variable.Type)
	} else if model.IsText(variable.Type) {
		field = NewTextField(s, storageName, variable.Name, variable.DisplayName, variable.Type)
	} else if model.IsImage(variable.Type) {
		field = NewImageField(s, storageName, variable.Name, variable.DisplayName, variable.Type)
	} else if model.IsDateTime(variable.Type) {
		field = NewDateTimeField(s, storageName, variable.Name, variable.DisplayName, variable.Type)
	} else {
		return nil, errors.Errorf("variable `%s` of type `%s` does not support summary", variable.Name, variable.Type)
	}

	timeseries, err := field.FetchForecastingSummaryData(timeColVar, interval, resultURI, filterParams)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch summary data")
	}

	timeseries.Type = "timeseries"

	// add dataset
	timeseries.Dataset = dataset

	return timeseries, err
}
