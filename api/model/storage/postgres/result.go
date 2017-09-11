package postgres

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"strconv"

	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil/api/model"
	log "github.com/unchartedsoftware/plog"
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
	variable, err := model.FetchVariable(s.clientES, index, dataset, targetName)
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

		// handle the parsed result/error
		if err != nil {
			return errors.Wrap(err, "failed csv value parsing")
		}
		parsedVal, err := strconv.ParseInt(records[i][0], 10, 64)
		if err != nil {
			return errors.Wrap(err, "failed csv index parsing")
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

func (s *Storage) parseResults(dataset string, rows *pgx.Rows, variable *model.Variable) (*model.FilteredData, error) {
	// Scan the rows. Each row has only the value as a string.
	values := [][]interface{}{}
	for rows.Next() {
		var value string
		err := rows.Scan(&value)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to parse result row")
		}
		var val interface{}
		err = nil

		switch variable.Type {
		case model.IntegerType:
			val, err = strconv.ParseInt(value, 10, 64)
		case model.FloatType:
			val, err = strconv.ParseFloat(value, 64)
		case model.CategoricalType:
			fallthrough
		case model.TextType:
			fallthrough
		case model.DateTimeType:
			fallthrough
		case model.OrdinalType:
			val = value
		case model.BoolType:
			val, err = strconv.ParseBool(value)
		default:
			val = value
		}
		// handle the parsed result/error
		if err != nil {
			return nil, errors.Wrap(err, "failed string value parsing")
		}
		values = append(values, []interface{}{val})
	}
	// Build the filtered data.
	return &model.FilteredData{
		Name: dataset,
		Metadata: []*model.Variable{
			{
				Name: variable.Name,
				Type: variable.Type,
			},
		},
		Values: values,
	}, nil
}

// FetchResults pulls the results from the Postgres database.
func (s *Storage) FetchResults(dataset string, resultURI string, index string) (*model.FilteredData, error) {
	datasetResult := s.getResultTable(dataset)
	targetName, err := s.getResultTargetName(datasetResult, resultURI, index)
	// fetch the variable info to resolve its type - skip the first column since that will be the d3m_index value
	variable, err := s.getResultTargetVariable(dataset, index, targetName)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf("SELECT value FROM %s WHERE result_id = $1 AND target = $2;", datasetResult)

	rows, err := s.client.Query(sql, resultURI, targetName)
	if err != nil {
		return nil, errors.Wrap(err, "Error querying results")
	}

	return s.parseResults(dataset, rows, variable)
}

func (s *Storage) getResultMinMaxAggsQuery(variable *model.Variable, resultVariable *model.Variable) string {
	// get min / max agg names
	minAggName := model.MinAggPrefix + resultVariable.Name
	maxAggName := model.MaxAggPrefix + resultVariable.Name

	// Only numeric types should occur.
	var fieldTyped string
	switch variable.Type {
	case model.IntegerType:
		fieldTyped = fmt.Sprintf("cast(\"%s\" as double precision)", resultVariable.Name)
	case model.FloatType:
		fieldTyped = fmt.Sprintf("cast(\"%s\" as double precision)", resultVariable.Name)
	default:
		fieldTyped = "error type"
	}

	// create aggregations
	queryPart := fmt.Sprintf("MIN(%s) AS \"%s\", MAX(%s) AS \"%s\"", fieldTyped, minAggName, fieldTyped, maxAggName)
	// add aggregations
	return queryPart
}

func (s *Storage) getResultHistogramAggQuery(extrema *model.Extrema, variable *model.Variable, resultVariable *model.Variable) (string, string) {
	// compute the bucket interval for the histogram
	interval := (extrema.Max - extrema.Min) / model.MaxNumBuckets
	if extrema.Type != model.FloatType {
		interval = math.Floor(interval)
		interval = math.Max(1, interval)
	}

	// Only numeric types should occur.
	var fieldTyped string
	switch variable.Type {
	case model.IntegerType:
		fieldTyped = fmt.Sprintf("cast(\"%s\" as double precision)", resultVariable.Name)
	case model.FloatType:
		fieldTyped = fmt.Sprintf("cast(\"%s\" as double precision)", resultVariable.Name)
	default:
		fieldTyped = "error type"
	}

	// get histogram agg name & query string.
	histogramAggName := fmt.Sprintf("\"%s%s\"", model.HistogramAggPrefix, extrema.Name)
	histogramQueryString := fmt.Sprintf("(%s / %f) * %f", fieldTyped, interval, interval)

	return histogramAggName, histogramQueryString
}

func (s *Storage) fetchResultExtrema(resultURI string, dataset string, variable *model.Variable, resultVariable *model.Variable) (*model.Extrema, error) {
	// add min / max aggregation
	aggQuery := s.getResultMinMaxAggsQuery(variable, resultVariable)

	// create a query that does min and max aggregations for each variable
	queryString := fmt.Sprintf("SELECT %s FROM %s WHERE result_id = $1 AND target = $2;", aggQuery, dataset)

	// execute the postgres query
	// NOTE: We may want to use the refular Query operation since QueryRow
	// hides db exceptions.
	res := s.client.QueryRow(queryString, resultURI, variable.Name)

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
	histogramName, histogramQuery := s.getResultHistogramAggQuery(extrema, variable, resultVariable)

	// Create the complete query string.
	query := fmt.Sprintf(`
		SELECT (%s) AS %s, COUNT(*) AS count FROM %s
		WHERE result_id = $1 AND target = $2
		GROUP BY %s ORDER BY %s;`, histogramQuery, histogramName, dataset, histogramQuery, histogramName)

	// execute the postgres query
	res, err := s.client.Query(query, resultURI, variable.Name)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch histograms for result variable summaries from postgres")
	}
	return s.parseNumericHistogram(res, extrema)
}

func (s *Storage) fetchCategoricalResultHistogram(resultURI string, dataset string, variable *model.Variable) (*model.Histogram, error) {
	// Get count by category.
	query := fmt.Sprintf("SELECT value, COUNT(*) AS count FROM %s WHERE result_id = $1 and target = $2 GROUP BY value ORDER BY count desc, value LIMIT %d;", dataset, catResultLimit)

	// execute the postgres query
	res, err := s.client.Query(query, resultURI, variable.Name)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch histograms for result summaries from postgres")
	}

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
		categorical, err := s.fetchCategoricalResultHistogram(resultURI, datasetResult, variable)
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
