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
	"strconv"
	"strings"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/model"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/serialization"
	log "github.com/unchartedsoftware/plog"
)

const (
	resultTableSuffix        = "_result"
	featureWeightTableSuffix = "_explain"
	dataTableAlias           = "data"
	confidenceName           = "confidence"
)

func (s *Storage) getResultTable(storageName string) string {
	return fmt.Sprintf("%s%s", storageName, resultTableSuffix)
}

func (s *Storage) getSolutionFeatureWeightTable(storageName string) string {
	return fmt.Sprintf("%s%s", storageName, featureWeightTableSuffix)
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
	err = rows.Err()
	if err != nil {
		return "", errors.Wrapf(err, "error reading data from postgres")
	}

	return "", errors.Errorf("Target feature for result URI `%s` not found", resultURI)
}

func (s *Storage) getResultTargetVariable(dataset string, targetName string) (*model.Variable, error) {
	variable, err := s.metadata.FetchVariable(dataset, targetName)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to get target variable information")
	}
	if variable.IsGrouping() && model.IsTimeSeries(variable.Type) {
		// extract the time series value column
		tsg := variable.Grouping.(*model.TimeseriesGrouping)
		return s.metadata.FetchVariable(dataset, tsg.YCol)
	}
	return variable, nil
}

// PersistSolutionFeatureWeight persists the solution feature weight to Postgres.
func (s *Storage) PersistSolutionFeatureWeight(dataset string, storageName string, resultURI string, weights [][]string) error {
	// weight structure is header row and then one row / d3m index
	// keep only weights that are tied to a variable in the base dataset
	fieldsDatabase, err := s.getDatabaseFields(fmt.Sprintf("%s_explain", storageName))
	if err != nil {
		return err
	}
	fieldsMetadata, err := s.metadata.FetchVariables(dataset, true, true)
	if err != nil {
		return err
	}
	fieldsMetadataMap := mapFields(fieldsMetadata)

	// build the map of shap output -> explain index
	fieldsWeight := weights[0]
	fieldsMap := make(map[int]int)
	fields := []string{"result_id"}
	for _, dbField := range fieldsDatabase {
		for i := 0; i < len(fieldsWeight); i++ {
			if fieldsMetadataMap[dbField] != nil && fieldsMetadataMap[dbField].Name == fieldsWeight[i] {
				// before we append the field, the length will give us the index for the field we will append
				fieldsMap[i] = len(fields)
				fields = append(fields, dbField)
			}
		}
	}

	values := make([][]interface{}, 0)
	for _, row := range weights[1:] {
		// parse into floats for storage (and add result uri)
		parsedWeights := make([]interface{}, len(fields))
		parsedWeights[0] = resultURI
		for i := 0; i < len(row); i++ {
			weightIndex, ok := fieldsMap[i]
			if ok {
				w, err := strconv.ParseFloat(row[i], 64)
				if err != nil {
					return errors.Wrap(err, "failed to parse feature weight")
				}
				parsedWeights[weightIndex] = w
			}
		}

		values = append(values, parsedWeights)
	}

	// batch the data to the storage
	err = s.InsertBatch(s.getSolutionFeatureWeightTable(storageName), fields, values)
	if err != nil {
		return errors.Wrap(err, "failed to insert result in database")
	}

	return nil
}

// PersistResult stores the solution result to Postgres.
func (s *Storage) PersistResult(dataset string, storageName string, resultURI string, target string, explainResult *api.SolutionExplainResult) error {
	// Read the results file.
	datasetStorage := serialization.GetStorage(resultURI)
	records, err := datasetStorage.ReadData(resultURI)
	if err != nil {
		return err
	}

	// read the confidence file
	explainValues := make(map[string]*api.SolutionExplainValues)
	if explainResult != nil {
		// build the confidence lookup
		for _, row := range explainResult.Values[1:] {
			parsedExplainValues, err := explainResult.ParsingFunction(row)
			if err != nil {
				return err
			}
			explainValues[row[explainResult.D3MIndexIndex]] = parsedExplainValues
		}
	}

	// currently only support a single result column.
	if len(records[0]) > 3 {
		log.Warnf("Result contains %d columns, expected 2 or 3 (confidence).  Additional columns will be ignored.", len(records[0]))
	}

	// Fetch the actual target variable (this can be different than the requested target for grouped variables)
	targetVariable, err := s.getResultTargetVariable(dataset, target)
	if err != nil {
		return err
	}

	// Translate from display name to storage name.
	targetDisplayName, err := s.getDisplayName(dataset, targetVariable.Name)
	if err != nil {
		return errors.Wrap(err, "unable to map target name")
	}

	// We can't guarantee that the order of the variables in the returned result matches our
	// locally stored indices, so we fetch by name from the source to be safe.
	targetIndex := -1
	d3mIndexIndex := -1
	confidenceIndex := -1
	for i, v := range records[0] {
		if v == targetDisplayName {
			targetIndex = i
		} else if v == model.D3MIndexFieldName {
			d3mIndexIndex = i
		} else if v == confidenceName {
			confidenceIndex = i
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
	indicesParsed := make(map[int64]bool)
	for i := 1; i < len(records); i++ {
		// Each data row is index, target.
		// handle the parsed result/error - should be an int some TA2 systems return floats
		parsedValFloat, err := strconv.ParseFloat(records[i][d3mIndexIndex], 64)
		if err != nil {
			return errors.Wrap(err, "failed csv index parsing")
		}
		parsedVal := int64(parsedValFloat)

		// assume (FOR NOW!!!) multi index results will be the same for a given
		// d3m index so store only 1 / index to not have duplicate query results
		if indicesParsed[parsedVal] {
			continue
		}
		indicesParsed[parsedVal] = true

		dataForInsert := []interface{}{resultURI, parsedVal, targetVariable.Name, records[i][targetIndex]}
		if confidenceIndex >= 0 {
			cfs, err := strconv.ParseFloat(records[i][confidenceIndex], 64)
			if err != nil {
				return errors.Wrap(err, "failed confidence value parsing")
			}
			dataForInsert = append(dataForInsert, cfs)
		}
		if explainValues[records[i][d3mIndexIndex]] != nil {
			dataForInsert = append(dataForInsert, explainValues[records[i][d3mIndexIndex]])
		}

		insertData = append(insertData, dataForInsert)
	}

	fields := []string{"result_id", "index", "target", "value"}
	if confidenceIndex >= 0 {
		fields = append(fields, "confidence")
	}
	if len(explainValues) > 0 {
		fields = append(fields, "explain_values")
	}

	// store all results to the storage
	err = s.insertBulkCopy(s.getResultTable(storageName), fields, insertData)
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

func (s *Storage) parseFilteredResults(variables []*model.Variable, rows pgx.Rows, target *model.Variable) (*api.FilteredData, error) {
	result := &api.FilteredData{
		Values: make([][]*api.FilteredDataValue, 0),
	}

	// Parse the columns (skipping weights columns)
	if rows != nil {
		var columns []*api.Column
		weightCount := 0
		confidenceCol := -1
		predictedCol := -1
		// Parse the row data.
		for rows.Next() {
			if columns == nil {
				fields := rows.FieldDescriptions()
				columns = make([]*api.Column, 0)
				for i := 0; i < len(fields); i++ {
					key := string(fields[i].Name)
					label := key
					typ := "unknown"
					if api.IsPredictedKey(key) {
						label = "Predicted " + api.StripKeySuffix(key)
						typ = target.Type
						predictedCol = i
					} else if api.IsErrorKey(key) {
						label = "Error"
						typ = target.Type
					} else if strings.HasPrefix(key, "__weights_") {
						weightCount = weightCount + 1
						continue
					} else if key == "__predicted_confidence" {
						confidenceCol = i
						continue
					} else {
						if key == target.Name {
							typ = target.Type
						} else {
							v := getVariableByKey(key, variables)
							if v != nil {
								typ = v.Type
								label = v.DisplayName
							}
						}
					}

					columns = append(columns, &api.Column{
						Key:   key,
						Label: label,
						Type:  typ,
					})
				}
			}
			columnValues, err := rows.Values()
			if err != nil {
				return nil, errors.Wrap(err, "Unable to extract fields from query result")
			}

			// match values with weights
			// weights are always the last columns, and match in order
			// with the value columns (skip the d3m index weight)
			weightedValues := make([]*api.FilteredDataValue, len(columns))
			varIndex := 0
			for i := 0; i < len(columnValues); i++ {
				if i == confidenceCol {
					if i < weightCount {
						// confidence column IS NOT a variable and so indices need to be adjusted
						varIndex--
					}
				} else if varIndex < len(weightedValues) {
					weightedValues[varIndex] = &api.FilteredDataValue{
						Value: columnValues[i],
					}
				} else if columnValues[i] != nil && columns[varIndex-weightCount].Key != model.D3MIndexFieldName {
					weightedValues[varIndex-weightCount].Weight = columnValues[i].(float64)
				}
				varIndex++
			}

			if confidenceCol >= 0 {
				weightedValues[predictedCol].Confidence = api.NullableFloat64(columnValues[confidenceCol].(float64))
			}
			result.Values = append(result.Values, weightedValues)
		}
		err := rows.Err()
		if err != nil {
			return nil, errors.Wrapf(err, "error reading data from postgres")
		}
		result.Columns = columns
	} else {
		result.Columns = make([]*api.Column, 0)
	}

	return result, nil
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

func getFullName(alias string, column string) string {
	fullName := fmt.Sprintf("\"%s\"", column)
	if alias != "" {
		fullName = fmt.Sprintf("%s.%s", alias, fullName)
	}

	return fullName
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

func addIncludeErrorFilterToWhere(wheres []string, params []interface{}, alias string, targetName string, residualFilter *model.Filter) ([]string, []interface{}, error) {
	// Add a clause to filter residuals to the existing where
	typedError := getErrorTyped(alias, targetName)
	where := fmt.Sprintf("(%s >= $%d AND %s <= $%d)", typedError, len(params)+1, typedError, len(params)+2)
	params = append(params, *residualFilter.Min)
	params = append(params, *residualFilter.Max)

	// Append the AND clause
	wheres = append(wheres, where)
	return wheres, params, nil
}

func addExcludeErrorFilterToWhere(wheres []string, params []interface{}, alias string, targetName string, residualFilter *model.Filter) ([]string, []interface{}, error) {
	// Add a clause to filter residuals to the existing where
	typedError := getErrorTyped(alias, targetName)
	where := fmt.Sprintf("(%s < $%d OR %s > $%d)", typedError, len(params)+1, typedError, len(params)+2)
	params = append(params, *residualFilter.Min)
	params = append(params, *residualFilter.Max)

	// Append the AND clause
	wheres = append(wheres, where)
	return wheres, params, nil
}

func addTableAlias(prefix string, fields []string, addToColumn bool) []string {
	fieldsPrepended := make([]string, len(fields))
	for i, f := range fields {
		// field name is quoted so need to prefix accordingly
		name := f
		if addToColumn {
			unquoted := name[1 : len(name)-1]
			name = fmt.Sprintf("%s as \"__%s_%s\"", name, prefix, unquoted)
		}
		fieldsPrepended[i] = fmt.Sprintf("%s.%s", prefix, name)
	}

	return fieldsPrepended
}

// FetchResults pulls the results from the Postgres database.  Note the generalized `id` string parameter - this will be
// a solution ID if this is a solution result, or a produce request ID is this is a predictions result.
func (s *Storage) FetchResults(dataset string, storageName string, resultURI string, id string, filterParams *api.FilterParams, predictionResultMode bool) (*api.FilteredData, error) {
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
	variables, err := s.metadata.FetchVariables(dataset, false, true)
	if err != nil {
		return nil, errors.Wrap(err, "Could not pull variables from ES")
	}

	// generate variable list for inclusion in query select
	distincts, fields, err := s.buildFilteredResultQueryField(variables, variable, filterParams.Variables)
	if err != nil {
		return nil, errors.Wrap(err, "Could not build field list")
	}
	fieldsData := addTableAlias(dataTableAlias, fields, false)
	fieldsExplain := addTableAlias("weights", fields, true)

	// break filters out groups for specific handling
	filters := splitFilters(filterParams)

	genericFilterParams := &api.FilterParams{
		Filters: filters.genericFilters,
	}

	// Create the filter portion of the where clause.
	wheres := make([]string, 0)
	params := make([]interface{}, 0)
	wheres, params = s.buildFilteredQueryWhere(dataset, wheres, params, dataTableAlias, genericFilterParams, false)

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
			wheres, params, err = addIncludeErrorFilterToWhere(wheres, params, dataTableAlias, targetName, filters.residualFilter)
			if err != nil {
				return nil, errors.Wrap(err, "Could not add error to where clause")
			}
		} else {
			wheres, params, err = addExcludeErrorFilterToWhere(wheres, params, dataTableAlias, targetName, filters.residualFilter)
			if err != nil {
				return nil, errors.Wrap(err, "Could not add error to where clause")
			}
		}
	}

	// If this is a timeseries forecast we don't want to include the target, predicted target or error
	// info in the returned data.  That information is fetched on a per-timeseries basis using the info
	// provided by this call.
	selectedVars := ""
	if isTimeSeriesValue(variables, variable) {
		selectedVars = fmt.Sprintf("%s %s ", distincts, strings.Join(fieldsData, ", "))
	} else {
		predictedCol := api.GetPredictedKey(id)

		// If our results are numerical we need to compute residuals and store them in a column called 'error'
		errorCol := api.GetErrorKey(id)
		errorExpr := ""
		if model.IsNumerical(variable.Type) && !predictionResultMode {
			errorExpr = fmt.Sprintf("%s as \"%s\",", getErrorTyped(dataTableAlias, variable.Name), errorCol)
		}

		targetColumnQuery := ""
		if !predictionResultMode {
			targetColumnQuery = fmt.Sprintf("data.\"%s\" as \"%s\", ", targetName, targetName)
		}

		selectedVars = fmt.Sprintf("%s predicted.value as \"%s\", COALESCE(predicted.confidence, 'NaN') as \"__predicted_confidence\", %s %s %s, %s ",
			distincts, predictedCol, targetColumnQuery, errorExpr, strings.Join(fieldsData, ", "), strings.Join(fieldsExplain, ", "))
	}

	wheres = append(wheres, fmt.Sprintf("predicted.result_id = $%d", len(params)+1))
	wheres = append(wheres, fmt.Sprintf("predicted.target = $%d", len(params)+2))
	wheres = append(wheres, "predicted.value != ''")
	params = append(params, resultURI)
	params = append(params, targetName)

	whereStatement := strings.Join(wheres, " AND ")

	query := fmt.Sprintf(
		"SELECT %s"+
			"FROM %s as predicted inner join %s as data on data.\"%s\" = predicted.index "+
			"LEFT OUTER JOIN %s as weights on weights.\"%s\" = predicted.index AND weights.result_id = predicted.result_id "+
			"WHERE %s",
		selectedVars, storageNameResult, storageName, model.D3MIndexFieldName,
		s.getSolutionFeatureWeightTable(storageName), model.D3MIndexFieldName,
		whereStatement)

	// Do not return the whole result set to the client.
	query = fmt.Sprintf("%s LIMIT %d;", query, filterParams.Size)

	rows, err := s.client.Query(query, params...)
	if err != nil {
		return nil, errors.Wrap(err, "Error querying results")
	}
	defer rows.Close()

	joinDef := &joinDefinition{
		baseAlias:     "data",
		baseColumn:    model.D3MIndexFieldName,
		joinTableName: storageNameResult,
		joinAlias:     "predicted",
		joinColumn:    "index",
	}
	numRows, err := s.fetchNumRowsJoined(storageName, variables, []string{"result_id = $1"}, []interface{}{resultURI}, joinDef)
	if err != nil {
		return nil, errors.Wrap(err, "Could not pull num rows")
	}
	numRowsFiltered, err := s.fetchNumRowsJoined(storageName, variables, wheres, params, joinDef)
	if err != nil {
		return nil, errors.Wrap(err, "Could not pull filtered num rows")
	}

	filteredData, err := s.parseFilteredResults(variables, rows, variable)
	if err != nil {
		return nil, err
	}
	filteredData.NumRows = numRows
	filteredData.NumRowsFiltered = numRowsFiltered

	weights, err := s.getAverageWeights(dataset, storageName, storageNameResult, resultURI, variables, whereStatement, params)
	if err != nil {
		return nil, err
	}
	for _, c := range filteredData.Columns {
		c.Weight = weights[c.Key]
	}

	return filteredData, nil
}

func (s *Storage) getAverageWeights(dataset string, storageName string, storageNameResult string, resultURI string,
	variables []*model.Variable, whereStatement string, params []interface{}) (map[string]float64, error) {
	variablesSQL := []string{}
	for _, v := range variables {
		if model.IsTA2Field(v.DistilRole, v.SelectedRole) {
			variablesSQL = append(variablesSQL, fmt.Sprintf("AVG(weights.\"%s\") as \"%s\"", v.Name, v.Name))
		}
	}

	sql := fmt.Sprintf("SELECT %s FROM %s AS weights INNER JOIN %s AS data on data.\"%s\" = weights.\"%s\" "+
		"INNER JOIN %s as predicted on data.\"%s\" = predicted.index WHERE %s",
		strings.Join(variablesSQL, ", "), s.getSolutionFeatureWeightTable(storageName),
		storageName, model.D3MIndexFieldName, model.D3MIndexFieldName, storageNameResult, model.D3MIndexFieldName, whereStatement)
	rows, err := s.client.Query(sql, params...)
	if err != nil {
		return nil, errors.Wrap(err, "unable to query for average result weights")
	}
	defer rows.Close()

	featureWeights, err := s.parseSolutionFeatureWeight(resultURI, rows)
	if err != nil {
		return nil, err
	}

	return featureWeights.Weights, nil
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

	field := NewNumericalField(s, dataset, storageName, targetVariable.Name, targetVariable.DisplayName, targetVariable.Type, "")
	return field.fetchResultsExtrema(resultURI, storageNameResult, resultVariable)
}

// FetchPredictedSummary gets the summary data about a target variable from the
// results table.
func (s *Storage) FetchPredictedSummary(dataset string, storageName string, resultURI string, filterParams *api.FilterParams, extrema *api.Extrema, mode api.SummaryMode) (*api.VariableSummary, error) {
	storageNameResult := s.getResultTable(storageName)
	weightTableName := s.getSolutionFeatureWeightTable(storageName)
	targetName, err := s.getResultTargetName(storageNameResult, resultURI)
	if err != nil {
		return nil, err
	}

	variable, err := s.getResultTargetVariable(dataset, targetName)
	if err != nil {
		return nil, err
	}

	countCol, err := s.getCountCol(dataset, mode)
	if err != nil {
		return nil, err
	}

	// use the variable type to guide the summary creation
	var field Field
	if model.IsNumerical(variable.Type) {
		field = NewNumericalField(s, dataset, storageName, variable.Name, variable.DisplayName, variable.Type, countCol)
	} else if model.IsCategorical(variable.Type) {
		field = NewCategoricalField(s, dataset, storageName, variable.Name, variable.DisplayName, variable.Type, countCol)
	} else if model.IsVector(variable.Type) {
		field = NewVectorField(s, dataset, storageName, variable.Name, variable.DisplayName, variable.Type)
	} else {
		return nil, errors.Errorf("variable %s of type %s does not support summary", variable.Name, variable.Type)
	}

	summary, err := field.FetchPredictedSummaryData(resultURI, storageNameResult, filterParams, extrema, mode)
	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch result summary")
	}

	// add dataset
	summary.Dataset = dataset

	weighted, err := s.getIsWeighted(resultURI, weightTableName)
	if err != nil {
		return nil, err
	}
	summary.Weighted = weighted
	return summary, nil
}

func (s *Storage) getIsWeighted(resultURI string, weightTableName string) (bool, error) {
	sql := fmt.Sprintf("SELECT EXISTS (SELECT * FROM %s WHERE result_id = $1 limit 1);", weightTableName)

	rows, err := s.client.Query(sql, resultURI)
	defer rows.Close()
	if err != nil {
		return false, errors.Wrap(err, "Unable to query weight state")
	}

	rows.Next()
	values, err := rows.Values()
	if err != nil {
		return false, errors.Wrap(err, "Unable to extract weight from query")
	}

	return bool(values[0].(bool)), nil
}

func (s *Storage) getDisplayName(dataset string, columnName string) (string, error) {
	displayName := ""
	variables, err := s.metadata.FetchVariables(dataset, false, false)
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

func mapFields(fields []*model.Variable) map[string]*model.Variable {
	mapped := make(map[string]*model.Variable)
	for _, f := range fields {
		mapped[f.Name] = f
	}

	return mapped
}

func isTimeSeriesValue(variables []*model.Variable, targetVariable *model.Variable) bool {
	for _, v := range variables {
		if v.IsGrouping() && model.IsTimeSeries(v.Grouping.GetType()) {
			tsg := v.Grouping.(*model.TimeseriesGrouping)
			if tsg.YCol == targetVariable.Name {
				return true
			}
		}
	}
	return false
}
