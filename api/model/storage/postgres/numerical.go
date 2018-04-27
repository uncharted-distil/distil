package postgres

import (
	"fmt"
	"strconv"

	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil/api/model"
)

// NumericalField defines behaviour for the numerical field type.
type NumericalField struct {
	Storage *Storage
}

// NewNumericalField creates a new field for numerical types.
func NewNumericalField(storage *Storage) *NumericalField {
	field := &NumericalField{
		Storage: storage,
	}

	return field
}

// FetchSummaryData pulls summary data from the database and builds a histogram.
func (f *NumericalField) FetchSummaryData(dataset string, variable *model.Variable, resultURI string, filterParams *model.FilterParams, extrema *model.Extrema) (*model.Histogram, error) {
	var histogram *model.Histogram
	var err error
	if resultURI == "" {
		histogram, err = f.fetchHistogram(dataset, variable, filterParams)
	} else {
		histogram, err = f.fetchHistogramByResult(dataset, variable, resultURI, filterParams, extrema)
	}

	return histogram, err
}

func (f *NumericalField) fetchHistogram(dataset string, variable *model.Variable, filterParams *model.FilterParams) (*model.Histogram, error) {
	// create the filter for the query.
	where, params := f.Storage.buildFilteredQueryWhere(dataset, filterParams)
	if len(where) > 0 {
		where = fmt.Sprintf(" WHERE %s", where)
	}

	// need the extrema to calculate the histogram interval
	extrema, err := f.fetchExtrema(dataset, variable)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch variable extrema for summary")
	}
	// for each returned aggregation, create a histogram aggregation. Bucket
	// size is derived from the min/max and desired bucket count.
	histogramName, bucketQuery, histogramQuery := f.getHistogramAggQuery(extrema)

	// Create the complete query string.
	query := fmt.Sprintf("SELECT %s as bucket, CAST(%s as double precision) AS %s, COUNT(*) AS count FROM %s%s GROUP BY %s ORDER BY %s;",
		bucketQuery, histogramQuery, histogramName, dataset, where, bucketQuery, histogramName)

	// execute the postgres query
	res, err := f.Storage.client.Query(query, params...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch histograms for variable summaries from postgres")
	}
	if res != nil {
		defer res.Close()
	}

	return f.parseHistogram(variable.Type, res, extrema)
}

func (f *NumericalField) fetchHistogramByResult(dataset string, variable *model.Variable, resultURI string, filterParams *model.FilterParams, extrema *model.Extrema) (*model.Histogram, error) {
	// create the filter for the query.
	where, params := f.Storage.buildFilteredQueryWhere(dataset, filterParams)
	if len(where) > 0 {
		where = fmt.Sprintf(" AND %s", where)
	}
	params = append(params, resultURI)

	// need the extrema to calculate the histogram interval
	var err error
	if extrema == nil {
		extrema, err = f.fetchExtremaByURI(dataset, resultURI, variable)
		if err != nil {
			return nil, errors.Wrap(err, "failed to fetch variable extrema for summary")
		}
	} else {
		extrema.Name = variable.Name
		extrema.Type = variable.Type
	}
	// for each returned aggregation, create a histogram aggregation. Bucket
	// size is derived from the min/max and desired bucket count.
	histogramName, bucketQuery, histogramQuery := f.getHistogramAggQuery(extrema)

	// Create the complete query string.
	query := fmt.Sprintf("SELECT %s as bucket, CAST(%s as double precision) AS %s, COUNT(*) AS count FROM %s data INNER JOIN %s result ON data.\"%s\" = result.index WHERE result.result_id = $%d%s GROUP BY %s ORDER BY %s;",
		bucketQuery, histogramQuery, histogramName, dataset,
		f.Storage.getResultTable(dataset), model.D3MIndexFieldName, len(params), where, bucketQuery, histogramName)

	// execute the postgres query
	res, err := f.Storage.client.Query(query, params...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch histograms for variable summaries from postgres")
	}
	if res != nil {
		defer res.Close()
	}

	return f.parseHistogram(variable.Type, res, extrema)
}

func (f *NumericalField) fetchExtrema(dataset string, variable *model.Variable) (*model.Extrema, error) {
	// add min / max aggregation
	aggQuery := f.getMinMaxAggsQuery(variable)

	// create a query that does min and max aggregations for each variable
	queryString := fmt.Sprintf("SELECT %s FROM %s;", aggQuery, dataset)

	// execute the postgres query
	// NOTE: We may want to use the regular Query operation since QueryRow
	// hides db exceptions.
	res, err := f.Storage.client.Query(queryString)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch extrema for variable summaries from postgres")
	}
	if res != nil {
		defer res.Close()
	}

	return f.parseExtrema(res, variable)
}

func (f *NumericalField) getHistogramAggQuery(extrema *model.Extrema) (string, string, string) {
	interval := extrema.GetBucketInterval()

	// get histogram agg name & query string.
	histogramAggName := fmt.Sprintf("\"%s%s\"", model.HistogramAggPrefix, extrema.Name)
	bucketQueryString := fmt.Sprintf("width_bucket(\"%s\", %g, %g, %d) - 1",
		extrema.Name, extrema.Min, extrema.Max, extrema.GetBucketCount())
	histogramQueryString := fmt.Sprintf("(%s) * %g + %g", bucketQueryString, interval, extrema.Min)

	return histogramAggName, bucketQueryString, histogramQueryString
}

func (f *NumericalField) parseHistogram(varType string, rows *pgx.Rows, extrema *model.Extrema) (*model.Histogram, error) {
	// get histogram agg name
	histogramAggName := model.HistogramAggPrefix + extrema.Name

	// Parse bucket results.
	interval := extrema.GetBucketInterval()

	buckets := make([]*model.Bucket, extrema.GetBucketCount())
	key := extrema.Min
	for i := 0; i < len(buckets); i++ {
		keyString := ""
		if model.IsFloatingPoint(extrema.Type) {
			keyString = fmt.Sprintf("%f", key)
		} else {
			keyString = strconv.Itoa(int(key))
		}

		buckets[i] = &model.Bucket{
			Key:   keyString,
			Count: 0,
		}

		key = key + interval
	}

	for rows.Next() {
		var bucketValue float64
		var bucketCount int64
		var bucket int64
		err := rows.Scan(&bucket, &bucketValue, &bucketCount)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("no %s histogram aggregation found", histogramAggName))
		}

		// Since the max can match the limit, an extra bucket may exist.
		// Add the value to the second to last bucket.
		if bucket < int64(len(buckets)) {
			buckets[bucket].Count = bucketCount
		} else {
			buckets[len(buckets)-1].Count += bucketCount

		}
	}
	// assign histogram attributes
	return &model.Histogram{
		Name:    extrema.Name,
		Type:    model.NumericalType,
		VarType: varType,
		Extrema: extrema,
		Buckets: buckets,
	}, nil
}

func (f *NumericalField) parseExtrema(row *pgx.Rows, variable *model.Variable) (*model.Extrema, error) {
	var minValue *float64
	var maxValue *float64
	if row != nil {
		// Expect one row of data.
		row.Next()
		err := row.Scan(&minValue, &maxValue)
		if err != nil {
			return nil, errors.Wrap(err, "no min / max aggregation found")
		}
	}
	// check values exist
	if minValue == nil || maxValue == nil {
		return nil, errors.Errorf("no min / max aggregation values found")
	}
	// assign attributes
	return &model.Extrema{
		Name: variable.Name,
		Type: variable.Type,
		Min:  *minValue,
		Max:  *maxValue,
	}, nil
}

func (f *NumericalField) getMinMaxAggsQuery(variable *model.Variable) string {
	// get min / max agg names
	minAggName := model.MinAggPrefix + variable.Name
	maxAggName := model.MaxAggPrefix + variable.Name

	// create aggregations
	queryPart := fmt.Sprintf("MIN(\"%s\") AS \"%s\", MAX(\"%s\") AS \"%s\"", variable.Name, minAggName, variable.Name, maxAggName)
	// add aggregations
	return queryPart
}

func (f *NumericalField) fetchExtremaByURI(dataset string, resultURI string, variable *model.Variable) (*model.Extrema, error) {
	// add min / max aggregation
	aggQuery := f.getMinMaxAggsQuery(variable)

	// create a query that does min and max aggregations for each variable
	queryString := fmt.Sprintf("SELECT %s FROM %s data INNER JOIN %s result ON data.\"%s\" = result.index WHERE result.result_id = $1;",
		aggQuery, dataset, f.Storage.getResultTable(dataset), model.D3MIndexFieldName)

	// execute the postgres query
	// NOTE: We may want to use the regular Query operation since QueryRow
	// hides db exceptions.
	res, err := f.Storage.client.Query(queryString, resultURI)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch extrema for variable summaries from postgres")
	}
	if res != nil {
		defer res.Close()
	}

	return f.parseExtrema(res, variable)
}

// FetchResultSummaryData pulls data from the result table and builds
// the numerical histogram for the field.
func (f *NumericalField) FetchResultSummaryData(resultURI string, dataset string, datasetResult string, variable *model.Variable, filterParams *model.FilterParams, extrema *model.Extrema) (*model.Histogram, error) {
	resultVariable := &model.Variable{
		Name: "value",
		Type: model.TextType,
	}

	// need the extrema to calculate the histogram interval
	var err error
	if extrema == nil {
		extrema, err = f.fetchResultsExtrema(resultURI, dataset, variable, resultVariable)
		if err != nil {
			return nil, errors.Wrap(err, "failed to fetch result variable extrema for summary")
		}
	} else {
		extrema.Name = variable.Name
		extrema.Type = variable.Type
	}
	// for each returned aggregation, create a histogram aggregation. Bucket
	// size is derived from the min/max and desired bucket count.
	histogramName, bucketQuery, histogramQuery := f.getResultHistogramAggQuery(extrema, variable, resultVariable)

	// create the filter for the query.
	where, params := f.Storage.buildFilteredQueryWhere(dataset, filterParams)
	if len(where) > 0 {
		where = fmt.Sprintf("WHERE %s AND result.result_id = $%d AND result.target = $%d", where, len(params)+1, len(params)+2)
	} else {
		where = "WHERE result.result_id = $1 AND result.target = $2"
	}
	params = append(params, resultURI, variable.Name)

	// Create the complete query string.
	query := fmt.Sprintf("SELECT %s as bucket, CAST(%s as double precision) AS %s, COUNT(*) AS count "+
		"FROM %s data INNER JOIN %s result ON data.\"%s\" = result.index %s "+
		"GROUP BY %s ORDER BY %s;",
		bucketQuery, histogramQuery, histogramName, dataset, datasetResult,
		model.D3MIndexFieldName, where, bucketQuery, histogramName)

	// execute the postgres query
	res, err := f.Storage.client.Query(query, params...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch histograms for result variable summaries from postgres")
	}
	defer res.Close()

	return f.parseHistogram(variable.Type, res, extrema)
}

func (f *NumericalField) getResultMinMaxAggsQuery(variable *model.Variable, resultVariable *model.Variable) string {
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

func (f *NumericalField) getResultHistogramAggQuery(extrema *model.Extrema, variable *model.Variable, resultVariable *model.Variable) (string, string, string) {
	// compute the bucket interval for the histogram
	interval := extrema.GetBucketInterval()

	// Only numeric types should occur.
	fieldTyped := fmt.Sprintf("cast(\"%s\" as double precision)", resultVariable.Name)

	// get histogram agg name & query string.
	histogramAggName := fmt.Sprintf("\"%s%s\"", model.HistogramAggPrefix, extrema.Name)
	bucketQueryString := fmt.Sprintf("width_bucket(%s, %g, %g, %d) - 1",
		fieldTyped, extrema.Min, extrema.Max, extrema.GetBucketCount())
	histogramQueryString := fmt.Sprintf("(%s) * %g + %g", bucketQueryString, interval, extrema.Min)

	return histogramAggName, bucketQueryString, histogramQueryString
}

func (f *NumericalField) fetchResultsExtrema(resultURI string, dataset string, variable *model.Variable, resultVariable *model.Variable) (*model.Extrema, error) {
	// add min / max aggregation
	aggQuery := f.getResultMinMaxAggsQuery(variable, resultVariable)

	// create a query that does min and max aggregations for each variable
	queryString := fmt.Sprintf("SELECT %s FROM %s WHERE result_id = $1 AND target = $2;", aggQuery, dataset)

	// execute the postgres query
	res, err := f.Storage.client.Query(queryString, resultURI, variable.Name)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch extrema for result from postgres")
	}
	defer res.Close()

	return f.parseExtrema(res, variable)
}
