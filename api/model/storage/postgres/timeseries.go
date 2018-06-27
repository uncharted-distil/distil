package postgres

import (
	"fmt"
	"math"
	"strings"

	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil/api/model"
)

// TimeSeriesField defines behaviour for the timeseries field type.
type TimeSeriesField struct {
	Storage *Storage
}

// NewTimeSeriesField creates a new field for timeseries types.
func NewTimeSeriesField(storage *Storage) *TimeSeriesField {
	field := &TimeSeriesField{
		Storage: storage,
	}

	return field
}

// FetchSummaryData pulls summary data from the database and builds a histogram.
func (f *TimeSeriesField) FetchSummaryData(dataset string, variable *model.Variable, resultURI string, filterParams *model.FilterParams, extrema *model.Extrema) (*model.Histogram, error) {
	var histogram *model.Histogram
	var err error
	if resultURI == "" {
		histogram, err = f.fetchHistogram(dataset, variable, filterParams)
	} else {
		histogram, err = f.fetchHistogramByResult(dataset, variable, resultURI, filterParams)
	}

	return histogram, err
}

func (f *TimeSeriesField) metadataVarName(varName string) string {
	return fmt.Sprintf("%s%s", model.MetadataVarPrefix, varName)
}

func (f *TimeSeriesField) fetchRepresentationTimeSeriess(dataset string, variable *model.Variable, categoryBuckets []*model.Bucket) ([]string, error) {

	var timeseriesFiles []string

	for _, bucket := range categoryBuckets {

		prefixedVarName := f.metadataVarName(variable.Name)

		// pull sample row containing bucket
		query := fmt.Sprintf("SELECT \"%s\" FROM %s WHERE \"%s\" = $1 LIMIT 1;", variable.Name, dataset, prefixedVarName)

		// execute the postgres query
		rows, err := f.Storage.client.Query(query, bucket.Key)
		if err != nil {
			return nil, errors.Wrap(err, "failed to fetch histograms for variable summaries from postgres")
		}

		if rows.Next() {
			var timeseriesFile string
			err = rows.Scan(&timeseriesFile)
			if err != nil {
				return nil, errors.Wrap(err, "Unable to parse solution from Postgres")
			}
			timeseriesFiles = append(timeseriesFiles, timeseriesFile)
		}
		rows.Close()
	}

	return timeseriesFiles, nil
}

func (f *TimeSeriesField) fetchHistogram(dataset string, variable *model.Variable, filterParams *model.FilterParams) (*model.Histogram, error) {
	// create the filter for the query.
	wheres := make([]string, 0)
	params := make([]interface{}, 0)
	wheres, params = f.Storage.buildFilteredQueryWhere(wheres, params, dataset, filterParams.Filters)

	prefixedVarName := f.metadataVarName(variable.Name)

	where := ""
	if len(wheres) > 0 {
		where = fmt.Sprintf("WHERE %s", strings.Join(wheres, " AND "))
	}

	// Get count by category.
	query := fmt.Sprintf("SELECT \"%s\", COUNT(*) AS count FROM %s %s GROUP BY \"%s\" ORDER BY count desc, \"%s\" LIMIT %d;",
		prefixedVarName, dataset, where, prefixedVarName, prefixedVarName, catResultLimit)

	// execute the postgres query
	res, err := f.Storage.client.Query(query, params...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch histograms for variable summaries from postgres")
	}
	if res != nil {
		defer res.Close()
	}

	histogram, err := f.parseHistogram(res, variable)
	if err != nil {
		return nil, err
	}

	files, err := f.fetchRepresentationTimeSeriess(dataset, variable, histogram.Buckets)
	if err != nil {
		return nil, err
	}
	histogram.Files = files
	return histogram, nil
}

func (f *TimeSeriesField) fetchHistogramByResult(dataset string, variable *model.Variable, resultURI string, filterParams *model.FilterParams) (*model.Histogram, error) {

	// pull filters generated against the result facet out for special handling
	filters := f.Storage.splitFilters(filterParams)

	// create the filter for the query.
	wheres := make([]string, 0)
	params := make([]interface{}, 0)
	wheres, params = f.Storage.buildFilteredQueryWhere(wheres, params, dataset, filters.genericFilters)

	var err error
	// apply the predicted result filter
	if filters.predictedFilter != nil {
		wheres, params, err = f.Storage.buildPredictedResultWhere(wheres, params, dataset, resultURI, filters.predictedFilter)
		if err != nil {
			return nil, err
		}
	} else if filters.correctnessFilter != nil {
		wheres, params, err = f.Storage.buildCorrectnessResultWhere(wheres, params, dataset, resultURI, filters.correctnessFilter)
		if err != nil {
			return nil, err
		}
	} else if filters.errorFilter != nil {
		wheres, params, err = f.Storage.buildErrorResultWhere(wheres, params, filters.errorFilter)
		if err != nil {
			return nil, err
		}
	}

	params = append(params, resultURI)

	where := ""
	if len(wheres) > 0 {
		where = fmt.Sprintf("AND %s", strings.Join(wheres, " AND "))
	}

	prefixedVarName := f.metadataVarName(variable.Name)

	// Get count by category.
	query := fmt.Sprintf(
		`SELECT data."%s", COUNT(*) AS count
		 FROM %s data INNER JOIN %s result ON data."%s" = result.index
		 WHERE result.result_id = $%d %s
		 GROUP BY "%s"
		 ORDER BY count desc, "%s" LIMIT %d;`,
		prefixedVarName, dataset, f.Storage.getResultTable(dataset),
		model.D3MIndexFieldName, len(params), where, prefixedVarName,
		prefixedVarName, catResultLimit)

	// execute the postgres query
	res, err := f.Storage.client.Query(query, params...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch histograms for variable summaries from postgres")
	}
	if res != nil {
		defer res.Close()
	}

	histogram, err := f.parseHistogram(res, variable)
	if err != nil {
		return nil, err
	}

	files, err := f.fetchRepresentationTimeSeriess(dataset, variable, histogram.Buckets)
	if err != nil {
		return nil, err
	}
	histogram.Files = files
	return histogram, nil
}

func (f *TimeSeriesField) parseHistogram(rows *pgx.Rows, variable *model.Variable) (*model.Histogram, error) {
	prefixedVarName := f.metadataVarName(variable.Name)

	termsAggName := model.TermsAggPrefix + prefixedVarName

	// parse as either one dimension or two dimension category histogram.  This could be collapsed down into a
	// single function.
	dimension := len(rows.FieldDescriptions()) - 1

	if dimension == 1 {
		return f.parseUnivariateHistogram(rows, variable, termsAggName)
	} else if dimension == 2 {
		return f.parseBivariateHistogram(rows, variable, termsAggName)
	} else {
		return nil, errors.Errorf("Unhandled dimension of %d for histogram %s", dimension, termsAggName)
	}
}

func (f *TimeSeriesField) parseUnivariateHistogram(rows *pgx.Rows, variable *model.Variable, termsAggName string) (*model.Histogram, error) {
	// Parse bucket results.
	buckets := make([]*model.Bucket, 0)
	min := int64(math.MaxInt32)
	max := int64(-math.MaxInt32)

	if rows != nil {
		for rows.Next() {
			var term string
			var bucketCount int64
			err := rows.Scan(&term, &bucketCount)
			if err != nil {
				return nil, errors.Wrap(err, fmt.Sprintf("no %s histogram aggregation found", termsAggName))
			}

			buckets = append(buckets, &model.Bucket{
				Key:   term,
				Count: bucketCount,
			})
			if bucketCount < min {
				min = bucketCount
			}
			if bucketCount > max {
				max = bucketCount
			}
		}
	}

	// assign histogram attributes
	return &model.Histogram{
		Name:    variable.Name,
		Type:    model.CategoricalType,
		VarType: variable.Type,
		Buckets: buckets,
		Extrema: &model.Extrema{
			Min: float64(min),
			Max: float64(max),
		},
	}, nil
}

func (f *TimeSeriesField) parseBivariateHistogram(rows *pgx.Rows, variable *model.Variable, termsAggName string) (*model.Histogram, error) {
	// extract the counts
	countMap := map[string]map[string]int64{}
	if rows != nil {
		for rows.Next() {
			var predictedTerm string
			var targetTerm string
			var bucketCount int64
			err := rows.Scan(&targetTerm, &predictedTerm, &bucketCount)
			if err != nil {
				return nil, errors.Wrap(err, fmt.Sprintf("no %s histogram aggregation found", termsAggName))
			}
			if len(countMap[predictedTerm]) == 0 {
				countMap[predictedTerm] = map[string]int64{}
			}
			countMap[predictedTerm][targetTerm] = bucketCount
		}
	}

	// convert the extracted counts into buckets suitable for serialization
	buckets := make([]*model.Bucket, 0)
	min := int64(math.MaxInt32)
	max := int64(-math.MaxInt32)

	for predictedKey, targetCounts := range countMap {
		bucket := model.Bucket{
			Key:     predictedKey,
			Count:   0,
			Buckets: []*model.Bucket{},
		}
		for targetKey, count := range targetCounts {
			targetBucket := model.Bucket{
				Key:   targetKey,
				Count: count,
			}
			bucket.Count = bucket.Count + count
			bucket.Buckets = append(bucket.Buckets, &targetBucket)
		}
		buckets = append(buckets, &bucket)
		if bucket.Count < min {
			min = bucket.Count
		}
		if bucket.Count > max {
			max = bucket.Count
		}
	}
	// assign histogram attributes
	return &model.Histogram{
		Name:    variable.Name,
		VarType: variable.Type,
		Type:    model.CategoricalType,
		Buckets: buckets,
		Extrema: &model.Extrema{
			Min: float64(min),
			Max: float64(max),
		},
	}, nil
}

// FetchPredictedSummaryData pulls predicted data from the result table and builds
// the timeseries histogram for the field.
func (f *TimeSeriesField) FetchPredictedSummaryData(resultURI string, dataset string, datasetResult string, variable *model.Variable, filterParams *model.FilterParams, extrema *model.Extrema) (*model.Histogram, error) {
	targetName := f.metadataVarName(variable.Name)

	// pull filters generated against the result facet out for special handling
	filters := f.Storage.splitFilters(filterParams)

	// create the filter for the query.
	wheres := make([]string, 0)
	params := make([]interface{}, 0)
	wheres, params = f.Storage.buildFilteredQueryWhere(wheres, params, dataset, filters.genericFilters)

	var err error
	// apply the predicted result filter
	if filters.predictedFilter != nil {
		wheres, params, err = f.Storage.buildPredictedResultWhere(wheres, params, dataset, resultURI, filters.predictedFilter)
		if err != nil {
			return nil, err
		}
	} else if filters.correctnessFilter != nil {
		wheres, params, err = f.Storage.buildCorrectnessResultWhere(wheres, params, dataset, resultURI, filters.correctnessFilter)
		if err != nil {
			return nil, err
		}
	} else if filters.errorFilter != nil {
		wheres, params, err = f.Storage.buildErrorResultWhere(wheres, params, filters.errorFilter)
		if err != nil {
			return nil, err
		}
	}

	wheres = append(wheres, fmt.Sprintf("result.result_id = $%d AND result.target = $%d ", len(params)+1, len(params)+2))
	params = append(params, resultURI, targetName)

	query := fmt.Sprintf(
		`SELECT data."%s", result.value, COUNT(*) AS count
		 FROM %s AS result INNER JOIN %s AS data ON result.index = data."%s"
		 WHERE %s
		 GROUP BY result.value, data."%s"
		 ORDER BY count desc;`,
		targetName, datasetResult, dataset, model.D3MIndexFieldName, strings.Join(wheres, " AND "), targetName)

	// execute the postgres query
	res, err := f.Storage.client.Query(query, params...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch histograms for result summaries from postgres")
	}
	defer res.Close()

	histogram, err := f.parseHistogram(res, variable)
	if err != nil {
		return nil, err
	}

	files, err := f.fetchRepresentationTimeSeriess(dataset, variable, histogram.Buckets)
	if err != nil {
		return nil, err
	}
	histogram.Files = files
	return histogram, nil
}
