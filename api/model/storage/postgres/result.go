package postgres

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil/api/model"
	log "github.com/unchartedsoftware/plog"
)

const (
	predictedSuffix = "_predicted"
	errorSuffix     = "_error"
	targetSuffix    = "_target"
)

func (s *Storage) getResultTable(dataset string) string {
	return fmt.Sprintf("%s_result", dataset)
}

func (s *Storage) getResultTargetName(dataset string, resultURI string, index string) (string, error) {
	// Assume only a single target / result. Read the target name from the
	// database table.
	sql := fmt.Sprintf("SELECT target FROM %s WHERE result_id = $1 LIMIT 1;", dataset)

	rows, err := s.client.Query(sql, resultURI)
	if err != nil {
		return "", errors.Wrap(err, "Unable to get target variable name from results")
	}
	defer rows.Close()

	if rows.Next() {
		var targetName string
		err = rows.Scan(&targetName)
		if err != nil {
			return "", errors.Wrap(err, "Unable to get target variable name from results")
		}

		return targetName, nil
	}

	return "", errors.New("Result URI not found")
}

func (s *Storage) getResultTargetVariable(dataset string, index string, targetName string) (*model.Variable, error) {
	variable, err := s.metadata.FetchVariable(dataset, index, targetName)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to get target variable information")
	}

	return variable, nil
}

// PersistResult stores the pipeline result to Postgres.
func (s *Storage) PersistResult(dataset string, resultURI string) error {
	// Read the results file.
	file, err := os.Open(resultURI)
	if err != nil {
		return errors.Wrap(err, "unable open pipeline result file")
	}
	csvReader := csv.NewReader(bufio.NewReader(file))
	csvReader.TrimLeadingSpace = true
	records, err := csvReader.ReadAll()
	if err != nil {
		return errors.Wrap(err, "unable load pipeline result as csv")
	}
	if len(records) <= 0 || len(records[0]) <= 0 {
		return errors.Wrap(err, "pipeline csv empty")
	}

	// currently only support a single result column.
	if len(records[0]) > 2 {
		log.Warnf("Result contains %s columns, expected 2.  Additional columns will be ignored.", len(records[0]))
	}

	// Header row will have the target.
	targetName := records[0][1]

	// store all results to the storage
	for i := 1; i < len(records); i++ {
		// Each data row is index, target.
		err = nil

		// handle the parsed result/error - should be an int some TA2 systems return floats
		if err != nil {
			return errors.Wrap(err, "failed csv value parsing")
		}
		parsedVal, err := strconv.ParseInt(records[i][0], 10, 64)
		if err != nil {
			parsedValFloat, err := strconv.ParseFloat(records[i][0], 64)
			if err != nil {
				return errors.Wrap(err, "failed csv index parsing")
			}
			parsedVal = int64(parsedValFloat)
		}

		// store the result to the storage
		err = s.executeInsertResultStatement(dataset, resultURI, parsedVal, targetName, records[i][1])
		if err != nil {
			return errors.Wrap(err, "failed to insert result in database")
		}
	}

	return nil
}

func (s *Storage) executeInsertResultStatement(dataset string, resultID string, index int64, target string, value string) error {
	statement := fmt.Sprintf("INSERT INTO %s (result_id, index, target, value) VALUES ($1, $2, $3, $4);", s.getResultTable(dataset))

	_, err := s.client.Exec(statement, resultID, index, target, value)

	return err
}

func (s *Storage) parseVariableValue(value string, variable *model.Variable) (interface{}, error) {
	// Integer types can be returned as floats.
	switch variable.Type {
	case model.IntegerType:
		return strconv.ParseFloat(value, 64)
	case model.FloatType:
		return strconv.ParseFloat(value, 64)
	case model.LongitudeType:
		return strconv.ParseFloat(value, 64)
	case model.LatitudeType:
		return strconv.ParseFloat(value, 64)
	case model.CategoricalType:
		fallthrough
	case model.TextType:
		fallthrough
	case model.DateTimeType:
		fallthrough
	case model.OrdinalType:
		return value, nil
	case model.BoolType:
		return strconv.ParseBool(value)
	default:
		return value, nil
	}
}

func (s *Storage) parseFilteredResults(dataset string, numRows int, rows *pgx.Rows, target *model.Variable) (*model.FilteredData, error) {
	result := &model.FilteredData{
		Name:    dataset,
		NumRows: numRows,
		Values:  make([][]interface{}, 0),
	}

	// Parse the columns.
	if rows != nil {
		fields := rows.FieldDescriptions()
		columns := make([]string, len(fields))
		types := make([]string, len(fields))
		for i := 0; i < len(fields); i++ {
			columns[i] = fields[i].Name
			types[i] = fields[i].DataTypeName
		}

		// Result type provided by DB needs to be overridden with defined target type.
		types[0] = target.Type

		// Parse the row data.
		for rows.Next() {
			columnValues, err := rows.Values()
			if err != nil {
				return nil, errors.Wrap(err, "Unable to extract fields from query result")
			}
			result.Values = append(result.Values, columnValues)
			result.Columns = columns
			result.Types = types
		}
	} else {
		result.Columns = make([]string, 0)
		result.Types = make([]string, 0)
	}

	return result, nil
}

func (s *Storage) parseResults(dataset string, numRows int, rows *pgx.Rows, variable *model.Variable) (*model.FilteredData, error) {
	// Scan the rows. Each row has only the value as a string.
	values := [][]interface{}{}
	for rows.Next() {
		var value string
		err := rows.Scan(&value)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to parse result row")
		}

		val, err := s.parseVariableValue(value, variable)
		if err != nil {
			return nil, errors.Wrap(err, "failed string value parsing")
		}
		values = append(values, []interface{}{val})
	}
	// Build the filtered data.
	return &model.FilteredData{
		Name:    dataset,
		NumRows: numRows,
		Columns: []string{variable.Name},
		Types:   []string{variable.Type},
		Values:  values,
	}, nil
}

type resultFilters struct {
	Predicted *model.Filter
	Error     *model.Filter
}

func removeResultFilters(filterParams *model.FilterParams) *resultFilters {
	// Strip the predicted and error filters out of the list - they need special handling
	var predictedFilter *model.Filter
	var errorFilter *model.Filter
	for i, filter := range filterParams.Filters {
		if strings.HasSuffix(filter.Name, predictedSuffix) {
			predictedFilter = filter
			filterParams.Filters = append(filterParams.Filters[:i], filterParams.Filters[i+1:]...)
		}
		if strings.HasSuffix(filter.Name, errorSuffix) {
			errorFilter = filter
			filterParams.Filters = append(filterParams.Filters[:i], filterParams.Filters[i+1:]...)
		}
	}

	return &resultFilters{
		Predicted: predictedFilter,
		Error:     errorFilter,
	}
}

func addPredictedFilterToWhere(dataset string, predictedFilter *model.Filter, wheres string, params []interface{}) (string, []interface{}, error) {
	// Handle the predicted column, which is accessed as `value` in the result query
	where := ""
	switch predictedFilter.Type {
	case model.NumericalFilter:
		// numerical
		where = fmt.Sprintf("cast(value AS double precision) >= $%d AND cast(value AS double precision) <= $%d", len(params)+1, len(params)+2)
		params = append(params, *predictedFilter.Min)
		params = append(params, *predictedFilter.Max)
	case model.CategoricalFilter:
		// categorical
		categories := make([]string, 0)
		offset := len(params) + 1
		for i, category := range predictedFilter.Categories {
			categories = append(categories, fmt.Sprintf("$%d", offset+i))
			params = append(params, category)
		}
		where = fmt.Sprintf("value IN (%s)", strings.Join(categories, ", "))
	default:
		return "", nil, errors.Errorf("unexpected type %s for variable %s", predictedFilter.Type, predictedFilter.Name)
	}

	// Append the AND clause
	if wheres != "" {
		wheres = " AND " + where
	} else {
		wheres = where
	}
	return wheres, params, nil
}

func addErrorFilterToWhere(dataset string, targetName string, errorFilter *model.Filter, wheres string, params []interface{}) (string, []interface{}, error) {
	// Add a clause to filter residuals to the existing where
	typedError := getErrorTyped(targetName)
	where := fmt.Sprintf("%s >= $%d AND %s <= $%d", typedError, len(params)+1, typedError, len(params)+2)
	params = append(params, *errorFilter.Min)
	params = append(params, *errorFilter.Max)

	// Append the AND clause
	if wheres != "" {
		wheres = " AND " + where
	} else {
		wheres = where
	}
	return wheres, params, nil
}

// FetchFilteredResults pulls the results from the Postgres database.
func (s *Storage) FetchFilteredResults(dataset string, index string, resultURI string, filterParams *model.FilterParams, inclusive bool) (*model.FilteredData, error) {
	datasetResult := s.getResultTable(dataset)
	targetName, err := s.getResultTargetName(datasetResult, resultURI, index)
	if err != nil {
		return nil, err
	}

	// fetch the variable info to resolve its type - skip the first column since that will be the d3m_index value
	variable, err := s.getResultTargetVariable(dataset, index, targetName)
	if err != nil {
		return nil, err
	}

	// fetch variable metadata
	variables, err := s.metadata.FetchVariables(dataset, index, false)
	if err != nil {
		return nil, errors.Wrap(err, "Could not pull variables from ES")
	}

	// remove result specific filters (predicted, error) from filters - they have their own handling
	resultFilters := removeResultFilters(filterParams)

	// generate variable list for inclusion in query select
	fields, err := s.buildFilteredResultQueryField(dataset, variables, variable, filterParams, inclusive)
	if err != nil {
		return nil, errors.Wrap(err, "Could not build field list")
	}

	// Create the filter portion of the where clause.
	where, params, err := s.buildFilteredQueryWhere(dataset, filterParams)
	if err != nil {
		return nil, errors.Wrap(err, "Could not build where clause")
	}

	// Add the predicted filter into the where clause if it was included in the filter set
	if resultFilters.Predicted != nil {
		where, params, err = addPredictedFilterToWhere(dataset, resultFilters.Predicted, where, params)
		if err != nil {
			return nil, errors.Wrap(err, "Could not add result to where clause")
		}
	}

	// Add the error filter into the where clause if it was included in the filter set
	if resultFilters.Error != nil {
		where, params, err = addErrorFilterToWhere(dataset, targetName, resultFilters.Error, where, params)
		if err != nil {
			return nil, errors.Wrap(err, "Could not add error to where clause")
		}
	}

	// If our results are numerical we need to compute residuals and store them in a column called 'error'
	errorExpr := ""
	errorCol := targetName + errorSuffix
	if model.IsNumerical(variable.Type) {
		errorExpr = fmt.Sprintf("%s as \"%s\",", getErrorTyped(variable.Name), errorCol)
	}

	predictedCol := targetName + predictedSuffix
	targetCol := targetName + targetSuffix

	query := fmt.Sprintf(
		"SELECT value as \"%s\", "+
			"\"%s\" as \"%s\", "+
			"%s "+
			"%s "+
			"FROM %s as predicted inner join %s as data on data.\"%s\" = predicted.index "+
			"WHERE result_id = $%d AND target = $%d",
		predictedCol, targetName, targetCol, errorExpr, fields, datasetResult, dataset, d3mIndexFieldName, len(params)+1, len(params)+2)
	params = append(params, resultURI)
	params = append(params, targetName)

	if len(where) > 0 {
		query = fmt.Sprintf("%s AND %s", query, where)
	}

	// Do not return the whole result set to the client.
	query = fmt.Sprintf("%s LIMIT %d;", query, filterParams.Size)

	rows, err := s.client.Query(query, params...)
	if err != nil {
		return nil, errors.Wrap(err, "Error querying results")
	}
	defer rows.Close()

	numRows, err := s.FetchNumRows(datasetResult)
	if err != nil {
		return nil, errors.Wrap(err, "Could not pull num rows")
	}

	return s.parseFilteredResults(dataset, numRows, rows, variable)
}

// FetchResults pulls the results from the Postgres database.
func (s *Storage) FetchResults(dataset string, index string, resultURI string) (*model.FilteredData, error) {

	// fetch the variable info to resolve its type - skip the first column since that will be the d3m_index value
	datasetResult := s.getResultTable(dataset)
	targetName, err := s.getResultTargetName(datasetResult, resultURI, index)
	variable, err := s.getResultTargetVariable(dataset, index, targetName)
	if err != nil {
		return nil, err
	}

	predictedCol := variable.Name + predictedSuffix
	sql := fmt.Sprintf("SELECT value FROM %s as %s WHERE result_id = $1 AND target = $2;", datasetResult, predictedCol)

	rows, err := s.client.Query(sql, resultURI, targetName)
	if err != nil {
		return nil, errors.Wrap(err, "Error querying results")
	}
	defer rows.Close()

	numRows, err := s.FetchNumRows(datasetResult)
	if err != nil {
		return nil, errors.Wrap(err, "Could not pull num rows")
	}

	return s.parseResults(dataset, numRows, rows, variable)
}

func (s *Storage) getResultMinMaxAggsQuery(variable *model.Variable, resultVariable *model.Variable) string {
	// get min / max agg names
	minAggName := model.MinAggPrefix + resultVariable.Name
	maxAggName := model.MaxAggPrefix + resultVariable.Name

	// Only numeric types should occur.
	fieldTyped := fmt.Sprintf("cast(\"%s\" as double precision)", resultVariable.Name)

	// create aggregations
	queryPart := fmt.Sprintf("MIN(%s) AS \"%s\", MAX(%s) AS \"%s\"", fieldTyped, minAggName, fieldTyped, maxAggName)
	// add aggregations
	return queryPart
}

func (s *Storage) getResultHistogramAggQuery(extrema *model.Extrema, variable *model.Variable, resultVariable *model.Variable) (string, string, string) {
	// compute the bucket interval for the histogram
	interval := s.calculateInterval(extrema)

	// Only numeric types should occur.
	fieldTyped := fmt.Sprintf("cast(\"%s\" as double precision)", resultVariable.Name)

	// get histogram agg name & query string.
	histogramAggName := fmt.Sprintf("\"%s%s\"", model.HistogramAggPrefix, extrema.Name)
	bucketQueryString := fmt.Sprintf("width_bucket(%s, %g, %g, %d) -1",
		fieldTyped, extrema.Min, extrema.Max, model.MaxNumBuckets-1)
	histogramQueryString := fmt.Sprintf("(%s) * %g + %g", bucketQueryString, interval, extrema.Min)

	return histogramAggName, bucketQueryString, histogramQueryString
}

func (s *Storage) fetchResultExtrema(resultURI string, dataset string, variable *model.Variable, resultVariable *model.Variable) (*model.Extrema, error) {
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

func (s *Storage) fetchNumericalResultHistogram(resultURI string, dataset string, variable *model.Variable) (*model.Histogram, error) {
	resultVariable := &model.Variable{
		Name: "value",
		Type: model.TextType,
	}

	// need the extrema to calculate the histogram interval
	extrema, err := s.fetchResultExtrema(resultURI, dataset, variable, resultVariable)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch result variable extrema for summary")
	}
	// for each returned aggregation, create a histogram aggregation. Bucket
	// size is derived from the min/max and desired bucket count.
	histogramName, bucketQuery, histogramQuery := s.getResultHistogramAggQuery(extrema, variable, resultVariable)

	// Create the complete query string.
	query := fmt.Sprintf(`
		SELECT %s as bucket, CAST(%s as double precision) AS %s, COUNT(*) AS count FROM %s
		WHERE result_id = $1 AND target = $2
		GROUP BY %s ORDER BY %s;`, bucketQuery, histogramQuery, histogramName, dataset, bucketQuery, histogramName)

	// execute the postgres query
	res, err := s.client.Query(query, resultURI, variable.Name)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch histograms for result variable summaries from postgres")
	}
	defer res.Close()

	return s.parseNumericHistogram(variable.Type, res, extrema)
}

func (s *Storage) fetchCategoricalResultHistogram(resultURI string, dataset string, resultDataset string, variable *model.Variable) (*model.Histogram, error) {
	targetName := variable.Name

	query := fmt.Sprintf("SELECT base.\"%s\", result.value, COUNT(*) AS count "+
		"FROM %s AS result INNER JOIN %s AS base ON result.index = base.\"d3mIndex\" "+
		"WHERE result.result_id = $1 and result.target = $2 "+
		"GROUP BY result.value, base.\"%s\" "+
		"ORDER BY count desc;", targetName, resultDataset, dataset, targetName)

	// execute the postgres query
	res, err := s.client.Query(query, resultURI, targetName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch histograms for result summaries from postgres")
	}
	defer res.Close()

	return s.parseCategoricalHistogram(res, variable)
}

// FetchResultsSummary gets the summary data about a target variable from the
// results table.
func (s *Storage) FetchResultsSummary(dataset string, resultURI string, index string) (*model.Histogram, error) {
	datasetResult := s.getResultTable(dataset)
	targetName, err := s.getResultTargetName(datasetResult, resultURI, index)
	if err != nil {
		return nil, err
	}

	variable, err := s.getResultTargetVariable(dataset, index, targetName)
	if err != nil {
		return nil, err
	}

	if model.IsNumerical(variable.Type) {
		// fetch numeric histograms
		numeric, err := s.fetchNumericalResultHistogram(resultURI, datasetResult, variable)
		if err != nil {
			return nil, err
		}
		return numeric, nil
	} else if model.IsCategorical(variable.Type) {
		// fetch categorical histograms
		categorical, err := s.fetchCategoricalResultHistogram(resultURI, dataset, datasetResult, variable)
		if err != nil {
			return nil, err
		}
		return categorical, nil
	} else if model.IsText(variable.Type) {
		// fetch text analysis
		return nil, nil
	}

	return nil, errors.Errorf("variable %s of type %s does not support summary", variable.Name, variable.Type)
}

func toFloat(value interface{}) (float64, error) {
	switch t := value.(type) {
	case int:
		return float64(t), nil
	case int8:
		return float64(t), nil
	case int16:
		return float64(t), nil
	case int32:
		return float64(t), nil
	case int64:
		return float64(t), nil
	case float32:
		return float64(t), nil
	case float64:
		return float64(t), nil
	default:
		return math.NaN(), errors.Errorf("unhandled type %T for %v in conversion to float64", t, value)
	}
}
