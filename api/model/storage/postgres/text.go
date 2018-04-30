package postgres

import (
	"fmt"
	"math"

	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil/api/model"
)

// TextField defines behaviour for the text field type.
type TextField struct {
	Storage *Storage
}

// NewTextField creates a new field for text types.
func NewTextField(storage *Storage) *TextField {
	field := &TextField{
		Storage: storage,
	}

	return field
}

// FetchSummaryData pulls summary data from the database and builds a histogram.
func (f *TextField) FetchSummaryData(dataset string, variable *model.Variable, resultURI string, filterParams *model.FilterParams, extrema *model.Extrema) (*model.Histogram, error) {
	var histogram *model.Histogram
	var err error
	if resultURI == "" {
		histogram, err = f.fetchHistogram(dataset, variable, filterParams)
	} else {
		histogram, err = f.fetchHistogramByResult(dataset, variable, resultURI, filterParams)
	}

	return histogram, err
}

func (f *TextField) fetchHistogram(dataset string, variable *model.Variable, filterParams *model.FilterParams) (*model.Histogram, error) {
	// create the filter for the query.
	where, params := f.Storage.buildFilteredQueryWhere(dataset, filterParams.Filters)
	if len(where) > 0 {
		where = fmt.Sprintf(" WHERE %s", where)
	}

	// Get count by category.
	query := fmt.Sprintf("SELECT w.word as %s, COUNT(*) as count "+
		"FROM (SELECT unnest(tsvector_to_array(to_tsvector(\"%s\"))) as stem FROM %s%s) as r "+
		"INNER JOIN %s as w on r.stem = w.stem "+
		"GROUP BY w.word ORDER BY count desc, w.word LIMIT %d;",
		variable.Name, variable.Name, dataset, where, wordStemTableName, catResultLimit)

	// execute the postgres query
	res, err := f.Storage.client.Query(query, params...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch text histogram for variable summaries from postgres")
	}
	if res != nil {
		defer res.Close()
	}

	return f.parseHistogram(res, variable)
}

func (f *TextField) fetchHistogramByResult(dataset string, variable *model.Variable, resultURI string, filterParams *model.FilterParams) (*model.Histogram, error) {
	// create the filter for the query.
	where, params := f.Storage.buildFilteredQueryWhere(dataset, filterParams.Filters)
	if len(where) > 0 {
		where = fmt.Sprintf(" AND %s", where)
	}
	params = append(params, resultURI)

	// Get count by category.
	query := fmt.Sprintf("SELECT w.word as \"%s\", COUNT(*) as count "+
		"FROM (SELECT unnest(tsvector_to_array(to_tsvector(\"%s\"))) as stem "+
		"FROM %s data INNER JOIN %s result ON data.\"%s\" = result.index WHERE result.result_id = $%d%s) as r "+
		"INNER JOIN %s as w on r.stem = w.stem "+
		"GROUP BY w.word ORDER BY count desc, w.word LIMIT %d;",
		variable.Name, variable.Name, dataset, f.Storage.getResultTable(dataset),
		model.D3MIndexFieldName, len(params), where, wordStemTableName, catResultLimit)

	// execute the postgres query
	res, err := f.Storage.client.Query(query, params...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch text histogram for variable summaries from postgres")
	}
	if res != nil {
		defer res.Close()
	}

	return f.parseHistogram(res, variable)
}

func (f *TextField) parseHistogram(rows *pgx.Rows, variable *model.Variable) (*model.Histogram, error) {
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

func (f *TextField) parseUnivariateHistogram(rows *pgx.Rows, variable *model.Variable, termsAggName string) (*model.Histogram, error) {
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

func (f *TextField) parseBivariateHistogram(rows *pgx.Rows, variable *model.Variable, termsAggName string) (*model.Histogram, error) {
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

// FetchResultSummaryData pulls data from the result table and builds
// the categorical histogram for the field.
func (f *TextField) FetchResultSummaryData(resultURI string, dataset string, datasetResult string, variable *model.Variable, filterParams *model.FilterParams, extrema *model.Extrema) (*model.Histogram, error) {
	targetName := variable.Name

	where, params := f.Storage.buildFilteredQueryWhere(dataset, filterParams.Filters)
	if len(where) > 0 {
		where = fmt.Sprintf(" WHERE %s AND result.result_id = $%d and result.target = $%d", where, len(params)+1, len(params)+2)
	} else {
		where = " WHERE result.result_id = $1 and result.target = $2"
	}
	params = append(params, resultURI, targetName)
	query := fmt.Sprintf("SELECT word_b.word as \"%s\", word_v.word as value, COUNT(*) as count "+
		"FROM (SELECT unnest(tsvector_to_array(to_tsvector(base.\"%s\"))) as stem_b, "+
		"unnest(tsvector_to_array(to_tsvector(result.value))) as stem_v "+
		"FROM %s AS result INNER JOIN %s AS base ON result.index = base.\"d3mIndex\" "+
		"%s) r INNER JOIN %s word_b ON r.stem_b = word_b.stem INNER JOIN %s word_v ON r.stem_v = word_v.stem "+
		"GROUP BY word_v.word, word_b.word "+
		"ORDER BY count desc;", targetName, targetName, datasetResult, dataset, where, wordStemTableName, wordStemTableName)

	// execute the postgres query
	res, err := f.Storage.client.Query(query, params...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch histograms for result summaries from postgres")
	}
	defer res.Close()

	return f.parseHistogram(res, variable)
}
