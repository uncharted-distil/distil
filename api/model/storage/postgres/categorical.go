package postgres

import (
	"fmt"
	"math"

	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil/api/model"
)

// CategoricalField defines behaviour for the categorical field type.
type CategoricalField struct {
	Storage *Storage
}

// NewCategoricalField creates a new field for categorical types.
func NewCategoricalField(storage *Storage) *CategoricalField {
	field := &CategoricalField{
		Storage: storage,
	}

	return field
}

// FetchSummaryData pulls summary data from the database and builds a histogram.
func (f *CategoricalField) FetchSummaryData(dataset string, variable *model.Variable, resultURI string, filterParams *model.FilterParams, extrema *model.Extrema) (*model.Histogram, error) {
	var histogram *model.Histogram
	var err error
	if resultURI == "" {
		histogram, err = f.fetchHistogram(dataset, variable, filterParams)
	} else {
		histogram, err = f.fetchHistogramByResult(dataset, variable, resultURI, filterParams)
	}

	return histogram, err
}

func (f *CategoricalField) fetchHistogram(dataset string, variable *model.Variable, filterParams *model.FilterParams) (*model.Histogram, error) {
	// create the filter for the query.
	where, params := f.Storage.buildFilteredQueryWhere(dataset, filterParams.Filters)
	if len(where) > 0 {
		where = fmt.Sprintf(" WHERE %s", where)
	}

	// Get count by category.
	query := fmt.Sprintf("SELECT \"%s\", COUNT(*) AS count FROM %s%s GROUP BY \"%s\" ORDER BY count desc, \"%s\" LIMIT %d;", variable.Name, dataset, where, variable.Name, variable.Name, catResultLimit)

	// execute the postgres query
	res, err := f.Storage.client.Query(query, params...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch histograms for variable summaries from postgres")
	}
	if res != nil {
		defer res.Close()
	}

	return f.parseHistogram(res, variable)
}

func (f *CategoricalField) fetchHistogramByResult(dataset string, variable *model.Variable, resultURI string, filterParams *model.FilterParams) (*model.Histogram, error) {

	// pull filters generated against the result facet out for special handling
	filters := f.Storage.splitFilters(filterParams)

	// create the filter for the query.
	where, params := f.Storage.buildFilteredQueryWhere(dataset, filters.genericFilters)

	// apply the predicted result filter
	if filters.predictedFilter != nil {
		resultWhere, predictedParams, err := f.Storage.buildPredictedResultWhere(dataset, resultURI, filters.predictedFilter)
		if err != nil {
			return nil, err
		}
		where = appendAndClause(where, resultWhere)
		params = append(params, predictedParams...)
	} else if filters.correctnessFilter != nil {
		resultWhere, err := f.Storage.buildCorrectnessResultWhere(dataset, resultURI, filters.correctnessFilter)
		if err != nil {
			return nil, err
		}
		where = appendAndClause(where, resultWhere)
	}

	if where != "" {
		where = "AND " + where
	}

	params = append(params, resultURI)

	// Get count by category.
	query := fmt.Sprintf(
		`SELECT data."%s", COUNT(*) AS count
		 FROM %s data INNER JOIN %s result ON data."%s" = result.index
		 WHERE result.result_id = $%d %s
		 GROUP BY "%s"
		 ORDER BY count desc, "%s" LIMIT %d;`,
		variable.Name, dataset, f.Storage.getResultTable(dataset),
		model.D3MIndexFieldName, len(params), where, variable.Name,
		variable.Name, catResultLimit)

	// execute the postgres query
	res, err := f.Storage.client.Query(query, params...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch histograms for variable summaries from postgres")
	}
	if res != nil {
		defer res.Close()
	}

	return f.parseHistogram(res, variable)
}

func (f *CategoricalField) parseHistogram(rows *pgx.Rows, variable *model.Variable) (*model.Histogram, error) {
	termsAggName := model.TermsAggPrefix + variable.Name

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

func (f *CategoricalField) parseUnivariateHistogram(rows *pgx.Rows, variable *model.Variable, termsAggName string) (*model.Histogram, error) {
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

func (f *CategoricalField) parseBivariateHistogram(rows *pgx.Rows, variable *model.Variable, termsAggName string) (*model.Histogram, error) {
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

// FetchResultSummaryData pulls predicted data from the result table and builds
// the categorical histogram for the field.
func (f *CategoricalField) FetchResultSummaryData(resultURI string, dataset string, datasetResult string, variable *model.Variable, filterParams *model.FilterParams, extrema *model.Extrema) (*model.Histogram, error) {
	targetName := variable.Name

	// pull filters generated against the result facet out for special handling
	filters := f.Storage.splitFilters(filterParams)

	// create the filter for the query.
	where, params := f.Storage.buildFilteredQueryWhere(dataset, filters.genericFilters)

	// apply the predicted result filter
	if filters.predictedFilter != nil {
		resultWhere, predictedParams, err := f.Storage.buildPredictedResultWhere(dataset, resultURI, filters.predictedFilter)
		if err != nil {
			return nil, err
		}
		where = appendAndClause(where, resultWhere)
		params = append(params, predictedParams...)
	} else if filters.correctnessFilter != nil {
		resultWhere, err := f.Storage.buildCorrectnessResultWhere(dataset, resultURI, filters.correctnessFilter)
		if err != nil {
			return nil, err
		}
		where = appendAndClause(where, resultWhere)
	}

	if len(where) > 0 {
		where = fmt.Sprintf(" %s AND result.result_id = $%d and result.target = $%d", where, len(params)+1, len(params)+2)
	} else {
		where = " result.result_id = $1 and result.target = $2"
	}
	params = append(params, resultURI, targetName)

	query := fmt.Sprintf(
		`SELECT data."%s", result.value, COUNT(*) AS count
		 FROM %s AS result INNER JOIN %s AS data ON result.index = data."d3mIndex"
		 WHERE %s
		 GROUP BY result.value, data."%s"
		 ORDER BY count desc;`,
		targetName, datasetResult, dataset, where, targetName)

	// execute the postgres query
	res, err := f.Storage.client.Query(query, params...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch histograms for result summaries from postgres")
	}
	defer res.Close()

	return f.parseHistogram(res, variable)
}
