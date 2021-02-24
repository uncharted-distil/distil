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
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/jackc/pgproto3/v2"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/model"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/serialization"
	jsonu "github.com/uncharted-distil/distil/api/util/json"
	log "github.com/unchartedsoftware/plog"
)

const (
	resultTableSuffix        = "_result"
	featureWeightTableSuffix = "_explain"
	dataTableAlias           = "data"
	confidenceName           = "confidence"
	rankName                 = "rank"
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
	fieldsMetadata, err := s.metadata.FetchVariables(dataset, true, true, false)
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
			if fieldsMetadataMap[dbField] != nil && fieldsMetadataMap[dbField].Key == fieldsWeight[i] {
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

// PersistExplainedResult stores the additional explained output.
func (s *Storage) PersistExplainedResult(dataset string, storageName string, resultURI string, explainResult *api.SolutionExplainResult) error {
	fieldName := "explain_values"
	params := make([][]interface{}, 0)
	if explainResult != nil {
		// build the explain lookup
		for _, row := range explainResult.Values[1:] {
			parsedExplainValues, err := explainResult.ParsingFunction(row)
			if err != nil {
				return err
			}
			params = append(params, []interface{}{row[explainResult.D3MIndexIndex], parsedExplainValues})
		}
	}

	// do a bulk update by creating the temp table, then doing an insert, then an update
	tx, err := s.batchClient.Begin()
	if err != nil {
		if rbErr := tx.Rollback(context.Background()); rbErr != nil {
			log.Error("rollback failed")
		}
		return errors.Wrap(err, "unable to create transaction")
	}

	tableNameTmp := fmt.Sprintf("%s_utmp", storageName)
	dataSQL := fmt.Sprintf("CREATE TEMP TABLE \"%s\" (\"%s\" TEXT NOT NULL, \"%s\" JSONB) ON COMMIT DROP;",
		tableNameTmp, model.D3MIndexName, fieldName)
	_, err = tx.Exec(context.Background(), dataSQL)
	if err != nil {
		if rbErr := tx.Rollback(context.Background()); rbErr != nil {
			log.Error("rollback failed")
		}
		return errors.Wrap(err, "unable to create temp table")
	}

	err = s.insertBulkCopyTransaction(tx, tableNameTmp, []string{model.D3MIndexName, fieldName}, params)
	if err != nil {
		if rbErr := tx.Rollback(context.Background()); rbErr != nil {
			log.Error("rollback failed")
		}
		return errors.Wrap(err, "unable to insert into temp table")
	}

	// build the filter structure
	wheres := []string{
		fmt.Sprintf("t.\"%s\" = b.index::text", model.D3MIndexName),
		"b.\"result_id\" = $1",
	}
	paramsFilter := []interface{}{resultURI}

	// run the update
	updateSQL := fmt.Sprintf("UPDATE %s.%s.\"%s\" AS b SET \"%s\" = t.\"%s\" FROM \"%s\" AS t WHERE %s",
		"distil", "public", s.getResultTable(storageName), fieldName, fieldName, tableNameTmp, strings.Join(wheres, " AND "))
	_, err = tx.Exec(context.Background(), updateSQL, paramsFilter...)
	if err != nil {
		if rbErr := tx.Rollback(context.Background()); rbErr != nil {
			log.Error("rollback failed")
		}
		return errors.Wrap(err, "unable to update base data")
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return errors.Wrap(err, "unable to commit bulk update")
	}

	return nil
}

// FetchResultDataset extracts the complete results and base table data.
func (s *Storage) FetchResultDataset(dataset string, storageName string, predictionName string, features []string, resultURI string) ([][]string, error) {
	fields := []string{}
	for _, v := range features {
		fields = append(fields, fmt.Sprintf("COALESCE(\"%s\", '') AS \"%s\"", v, v))
	}
	fields = append(fields, fmt.Sprintf("COALESCE(result.value) AS \"%s\"", predictionName))
	sql := fmt.Sprintf(`
		SELECT %s
		FROM %s base
		INNER JOIN %s result on CAST(base."d3mIndex" AS double precision) = result.index
		WHERE result.result_id = $1;`,
		strings.Join(fields, ", "), getBaseTableName(storageName), s.getResultTable(storageName))
	res, err := s.client.Query(sql, resultURI)
	if err != nil {
		return nil, errors.Wrapf(err, "unable execute query to extract dataset")
	}

	return s.parseData(res)
}

// FetchExplainValues : fetches fetches explain values
func (s *Storage) FetchExplainValues(dataset string, storageName string, d3mIndex []int, resultUUID string) ([]api.SolutionExplainValues, error) {
	// create the filter for the query.
	wheres := make([]string, 0)
	params := make([]interface{}, 0)
	params = append(params, resultUUID)
	wheres = append(wheres, fmt.Sprintf("s.solution_id = $%d", len(params)))
	params = append(params, d3mIndex)
	wheres = append(wheres, fmt.Sprintf("r.index=ANY($%d)", len(params)))

	where := fmt.Sprintf("WHERE %s", strings.Join(wheres, " AND "))

	query := fmt.Sprintf(`
	SELECT %s
	FROM %s AS r INNER JOIN solution_result AS s ON r.result_id=s.result_uri
	%s
	LIMIT %d`,
		model.ExplainValues, s.getResultTable(storageName),
		where, len(d3mIndex))
	res, err := s.client.Query(query, params...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch explanation values from postgres")
	}
	if res != nil {
		defer res.Close()
	}
	result := make([]api.SolutionExplainValues, 0)
	for res.Next() {
		buffer := api.SolutionExplainValues{GradCAM: [][]float64{{0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}}}
		err := res.Scan(&buffer)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan explanation values from postgres")
		}
		result = append(result, buffer)
	}
	return result, nil
}

// PersistResult stores the solution result to Postgres.
func (s *Storage) PersistResult(dataset string, storageName string, resultURI string, target string) error {
	// Read the results file.
	datasetStorage := serialization.GetStorage(resultURI)
	records, err := datasetStorage.ReadData(resultURI)
	if err != nil {
		return err
	}

	// currently only support a single result column.
	if len(records[0]) > 4 {
		log.Warnf("Result contains %d columns, expected 2, 3 or 4 (explanations).  Additional columns will be ignored.", len(records[0]))
	}

	// Fetch the actual target variable (this can be different than the requested target for grouped variables)
	targetVariable, err := s.getResultTargetVariable(dataset, target)
	if err != nil {
		return err
	}

	// Translate from display name to storage name.
	targetHeaderName, err := s.getHeaderName(dataset, targetVariable.Key)
	if err != nil {
		return errors.Wrap(err, "unable to map target name")
	}

	// We can't guarantee that the order of the variables in the returned result matches our
	// locally stored indices, so we fetch by name from the source to be safe.
	fieldsHeader := map[string]int{}
	for i, v := range records[0] {
		fieldsHeader[v] = i
	}
	targetIndex, ok := fieldsHeader[targetHeaderName]
	if !ok {
		return errors.Wrapf(err, "unable to find target col '%s' in result header", targetHeaderName)
	}
	d3mIndexIndex, ok := fieldsHeader[model.D3MIndexFieldName]
	if !ok {
		return errors.Wrapf(err, "unabled to find d3m index col '%s' in result header", model.D3MIndexFieldName)
	}
	confidenceIndex, ok := fieldsHeader[confidenceName]
	if !ok {
		confidenceIndex = -1
	}
	rankIndex, ok := fieldsHeader[rankName]
	if !ok {
		rankIndex = -1
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

		dataForInsert := []interface{}{resultURI, parsedVal, targetHeaderName, records[i][targetIndex]}
		explainValues, err := s.parseExplainValues(records[i], confidenceIndex, rankIndex)
		if err != nil {
			return err
		}
		if explainValues != nil {
			dataForInsert = append(dataForInsert, explainValues)
		}

		insertData = append(insertData, dataForInsert)
	}

	fields := []string{"result_id", "index", "target", "value"}
	if confidenceIndex+rankIndex > -2 {
		fields = append(fields, "explain_values")
	}

	// store all results to the storage
	err = s.insertBulkCopy(s.getResultTable(storageName), fields, insertData)
	if err != nil {
		return errors.Wrap(err, "failed to insert result in database")
	}

	return nil
}

func (s *Storage) parseExplainValues(record []string, confidenceIndex int, rankIndex int) (*api.SolutionExplainValues, error) {
	// -1 + -1 = -2 => no confidence nor ranking
	if confidenceIndex+rankIndex == -2 {
		return nil, nil
	}

	// can have ranking, confidence or both
	explain := &api.SolutionExplainValues{}
	if confidenceIndex >= 0 {
		cfs, err := strconv.ParseFloat(record[confidenceIndex], 64)
		if err != nil {
			return nil, errors.Wrap(err, "failed confidence value parsing")
		}
		explain.Confidence = cfs
	}
	if rankIndex >= 0 {
		rs, err := strconv.ParseFloat(record[rankIndex], 64)
		if err != nil {
			return nil, errors.Wrap(err, "failed rank value parsing")
		}
		explain.Rank = rs
	}

	return explain, nil
}

func (s *Storage) parseFilteredResults(variables []*model.Variable, rows pgx.Rows, target *model.Variable) (*api.FilteredData, error) {
	result := &api.FilteredData{
		Values: make([][]*api.FilteredDataValue, 0),
	}

	// Parse the columns (skipping weights columns)
	if rows != nil {
		var columns []*api.Column
		var fields []pgproto3.FieldDescription
		weightCount := 0
		predictedCol := -1
		explainCol := -1
		// Parse the row data.
		for rows.Next() {
			if columns == nil {
				fields = rows.FieldDescriptions()
				columns = make([]*api.Column, 0)
				for i := 0; i < len(fields); i++ {
					key := string(fields[i].Name)
					var label, typ string
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
					} else if key == "__predicted_explain" {
						explainCol = i
						continue
					} else {
						v := getVariableByKey(key, variables)
						key = v.Key
						label = v.DisplayName
						if key == target.Key {
							typ = target.Type
						} else {
							typ = v.Type
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
				if i == explainCol {
					if i < weightCount {
						// explain column IS NOT variable and so indices need to be adjusted
						varIndex--
					}
				} else if varIndex < len(weightedValues) {
					parsedValue, err := parsePostgresType(columnValues[i], fields[i])
					if err != nil {
						return nil, err
					}
					weightedValues[varIndex] = &api.FilteredDataValue{
						Value: parsedValue,
					}
				} else if columnValues[i] != nil && columns[varIndex-weightCount].Key != model.D3MIndexFieldName {
					weightedValues[varIndex-weightCount].Weight = columnValues[i].(float64)
				}
				varIndex++
			}
			if explainCol >= 0 && columnValues[explainCol] != nil {
				explainValuesRaw := columnValues[explainCol].(map[string]interface{})
				explainValuesParsed := &api.SolutionExplainValues{}
				err = jsonu.MapToStruct(explainValuesParsed, explainValuesRaw)
				if err != nil {
					return nil, err
				}
				weightedValues[predictedCol].Rank = api.NullableFloat64(explainValuesParsed.Rank)
				weightedValues[predictedCol].Confidence = api.NullableFloat64(explainValuesParsed.Confidence)
			} else if predictedCol >= 0 {
				weightedValues[predictedCol].Confidence = api.NullableFloat64(math.NaN())
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

func parsePostgresType(columnValue interface{}, description pgproto3.FieldDescription) (interface{}, error) {
	// do not want to return postgres specific types here
	if description.DataTypeOID == pgtype.Float8ArrayOID {
		output := []float64{}
		typed := columnValue.(pgtype.Float8Array)
		err := typed.AssignTo(&output)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to parse float array")
		}
		return output, nil
	}
	return columnValue, nil
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
	where = fmt.Sprintf("predicted.value %s data.\"%s\"", op, target.Key)
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
	where = fmt.Sprintf("predicted.value %s data.\"%s\"", op, target.Key)
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

func addExcludeConfidenceResultToWhere(wheres []string, params []interface{}, confidenceFilter *model.Filter) ([]string, []interface{}) {
	where := fmt.Sprintf("((explain_values -> 'confidence')::double precision < $%d OR (explain_values -> 'confidence')::double precision > $%d)", len(params)+1, len(params)+2)
	params = append(params, *confidenceFilter.Min)
	params = append(params, *confidenceFilter.Max)

	// Append the AND clause
	wheres = append(wheres, where)
	return wheres, params
}

func addExcludeRankResultToWhere(wheres []string, params []interface{}, rankFilter *model.Filter) ([]string, []interface{}) {
	where := fmt.Sprintf("((explain_values -> 'rank')::double precision < $%d OR (explain_values -> 'rank')::double precision > $%d)", len(params)+1, len(params)+2)
	params = append(params, *rankFilter.Min)
	params = append(params, *rankFilter.Max)

	// Append the AND clause
	wheres = append(wheres, where)
	return wheres, params
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
	variables, err := s.metadata.FetchVariables(dataset, true, true, false)
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

	// Add the error filter into the where clause if it was included in the filter set
	if filters.confidenceFilter != nil {
		if filters.confidenceFilter.Mode == model.IncludeFilter {
			wheres, params = s.buildConfidenceResultWhere(wheres, params, filters.confidenceFilter, "predicted")
		} else {
			wheres, params = addExcludeConfidenceResultToWhere(wheres, params, filters.confidenceFilter)
		}
	}
	if filters.rankFilter != nil {
		if filters.rankFilter.Mode == model.IncludeFilter {
			wheres, params = s.buildRankResultWhere(wheres, params, filters.rankFilter, "predicted")
		} else {
			wheres, params = addExcludeRankResultToWhere(wheres, params, filters.rankFilter)
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
			errorExpr = fmt.Sprintf("%s as \"%s\",", getErrorTyped(dataTableAlias, variable.Key), errorCol)
		}

		targetColumnQuery := ""
		if !predictionResultMode {
			targetColumnQuery = fmt.Sprintf("data.\"%s\" as \"%s\", ", targetName, targetName)
		}

		selectedVars = fmt.Sprintf("%s predicted.value as \"%s\", predicted.explain_values as \"__predicted_explain\", %s %s %s, %s ",
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
		if model.IsTA2Field(v.DistilRole, v.SelectedRole) && !model.IsIndexRole(v.SelectedRole) && !v.IsGrouping() {
			variablesSQL = append(variablesSQL, fmt.Sprintf("AVG(weights.\"%s\") as \"%s\"", v.Key, v.Key))
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
		Key:  "value",
		Type: model.StringType,
	}

	field := NewNumericalField(s, dataset, storageName, targetVariable.Key, targetVariable.DisplayName, targetVariable.Type, "")
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
		field = NewNumericalField(s, dataset, storageName, variable.Key, variable.DisplayName, variable.Type, countCol)
	} else if model.IsCategorical(variable.Type) {
		field = NewCategoricalField(s, dataset, storageName, variable.Key, variable.DisplayName, variable.Type, countCol)
	} else if model.IsVector(variable.Type) {
		field = NewVectorField(s, dataset, storageName, variable.Key, variable.DisplayName, variable.Type)
	} else {
		return nil, errors.Errorf("variable %s of type %s does not support summary", variable.Key, variable.Type)
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
	if err != nil {
		return false, errors.Wrap(err, "Unable to query weight state")
	}
	defer rows.Close()

	rows.Next()
	values, err := rows.Values()
	if err != nil {
		return false, errors.Wrap(err, "Unable to extract weight from query")
	}

	return bool(values[0].(bool)), nil
}

func (s *Storage) getHeaderName(dataset string, key string) (string, error) {
	variable, err := s.metadata.FetchVariable(dataset, key)
	if err != nil {
		return "", errors.Wrap(err, "unable fetch variable for name mapping")
	}

	if variable.IsGrouping() && model.IsTimeSeries(variable.Grouping.GetType()) {
		tsg := variable.Grouping.(*model.TimeseriesGrouping)
		variable, err = s.metadata.FetchVariable(dataset, tsg.YCol)
		if err != nil {
			return "", errors.Wrap(err, "unable fetch variable for name mapping")
		}
	}

	return variable.HeaderName, nil
}

func mapFields(fields []*model.Variable) map[string]*model.Variable {
	mapped := make(map[string]*model.Variable)
	for _, f := range fields {
		mapped[f.Key] = f
	}

	return mapped
}

func isTimeSeriesValue(variables []*model.Variable, targetVariable *model.Variable) bool {
	for _, v := range variables {
		if v.IsGrouping() && model.IsTimeSeries(v.Grouping.GetType()) {
			tsg := v.Grouping.(*model.TimeseriesGrouping)
			if tsg.YCol == targetVariable.Key {
				return true
			}
		}
	}
	return false
}
