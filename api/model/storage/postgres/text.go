package postgres

import (
	"fmt"
	"math"
	"strings"

	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil/api/model"
)

// TextField defines behaviour for the text field type.
type TextField struct {
	Storage  *Storage
	Dataset  string
	Variable *model.Variable
}

// NewTextField creates a new field for text types.
func NewTextField(storage *Storage, dataset string, variable *model.Variable) *TextField {
	field := &TextField{
		Storage:  storage,
		Dataset:  dataset,
		Variable: variable,
	}

	return field
}

// FetchSummaryData pulls summary data from the database and builds a histogram.
func (f *TextField) FetchSummaryData(resultURI string, filterParams *model.FilterParams, extrema *model.Extrema) (*model.Histogram, error) {
	var histogram *model.Histogram
	var err error
	if resultURI == "" {
		histogram, err = f.fetchHistogram(filterParams)
	} else {
		histogram, err = f.fetchHistogramByResult(resultURI, filterParams)
	}

	return histogram, err
}

func (f *TextField) fetchHistogram(filterParams *model.FilterParams) (*model.Histogram, error) {
	// create the filter for the query.
	wheres := make([]string, 0)
	params := make([]interface{}, 0)
	wheres, params = f.Storage.buildFilteredQueryWhere(wheres, params, f.Dataset, filterParams.Filters)

	where := ""
	if len(wheres) > 0 {
		where = fmt.Sprintf("WHERE %s", strings.Join(wheres, " AND "))
	}

	// Get count by category.
	query := fmt.Sprintf("SELECT w.word as %s, COUNT(*) as count "+
		"FROM (SELECT unnest(tsvector_to_array(to_tsvector(\"%s\"))) as stem FROM %s %s) as r "+
		"INNER JOIN %s as w on r.stem = w.stem "+
		"GROUP BY w.word ORDER BY count desc, w.word LIMIT %d;",
		f.Variable.Key, f.Variable.Key, f.Dataset, where, wordStemTableName, catResultLimit)

	// execute the postgres query
	res, err := f.Storage.client.Query(query, params...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch text histogram for variable summaries from postgres")
	}
	if res != nil {
		defer res.Close()
	}

	return f.parseHistogram(res)
}

func (f *TextField) fetchHistogramByResult(resultURI string, filterParams *model.FilterParams) (*model.Histogram, error) {

	// get filter where / params
	wheres, params, err := f.Storage.buildResultQueryFilters(f.Dataset, resultURI, filterParams)
	if err != nil {
		return nil, err
	}

	params = append(params, resultURI)

	where := ""
	if len(wheres) > 0 {
		where = fmt.Sprintf("AND %s", strings.Join(wheres, " AND "))
	}

	// Get count by category.
	query := fmt.Sprintf("SELECT w.word as \"%s\", COUNT(*) as count "+
		"FROM (SELECT unnest(tsvector_to_array(to_tsvector(\"%s\"))) as stem "+
		"FROM %s data INNER JOIN %s result ON data.\"%s\" = result.index WHERE result.result_id = $%d %s) as r "+
		"INNER JOIN %s as w on r.stem = w.stem "+
		"GROUP BY w.word ORDER BY count desc, w.word LIMIT %d;",
		f.Variable.Key, f.Variable.Key, f.Dataset, f.Storage.getResultTable(f.Dataset),
		model.D3MIndexFieldName, len(params), where, wordStemTableName, catResultLimit)

	// execute the postgres query
	res, err := f.Storage.client.Query(query, params...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch text histogram for variable summaries from postgres")
	}
	if res != nil {
		defer res.Close()
	}

	return f.parseHistogram(res)
}

func (f *TextField) parseHistogram(rows *pgx.Rows) (*model.Histogram, error) {
	termsAggName := model.TermsAggPrefix + f.Variable.Key

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
		Label:   f.Variable.Label,
		Key:     f.Variable.Key,
		Type:    model.CategoricalType,
		VarType: f.Variable.Type,
		Buckets: buckets,
		Extrema: &model.Extrema{
			Min: float64(min),
			Max: float64(max),
		},
	}, nil
}

// FetchPredictedSummaryData pulls data from the result table and builds
// the categorical histogram for the field.
func (f *TextField) FetchPredictedSummaryData(resultURI string, datasetResult string, filterParams *model.FilterParams, extrema *model.Extrema) (*model.Histogram, error) {
	targetName := f.Variable.Key

	// get filter where / params
	wheres, params, err := f.Storage.buildResultQueryFilters(f.Dataset, resultURI, filterParams)
	if err != nil {
		return nil, err
	}

	wheres = append(wheres, fmt.Sprintf("result.result_id = $%d AND result.target = $%d ", len(params)+1, len(params)+2))
	params = append(params, resultURI, targetName)

	query := fmt.Sprintf("SELECT word_b.word as \"%s\", word_v.word as value, COUNT(*) as count "+
		"FROM (SELECT unnest(tsvector_to_array(to_tsvector(base.\"%s\"))) as stem_b, "+
		"unnest(tsvector_to_array(to_tsvector(result.value))) as stem_v "+
		"FROM %s AS result INNER JOIN %s AS base ON result.index = base.\"d3mIndex\" "+
		"WHERE %s) r INNER JOIN %s word_b ON r.stem_b = word_b.stem INNER JOIN %s word_v ON r.stem_v = word_v.stem "+
		"GROUP BY word_v.word, word_b.word "+
		"ORDER BY count desc;", targetName, targetName, datasetResult, f.Dataset, strings.Join(wheres, " AND "), wordStemTableName, wordStemTableName)

	// execute the postgres query
	res, err := f.Storage.client.Query(query, params...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch histograms for result summaries from postgres")
	}
	defer res.Close()

	return f.parseHistogram(res)
}
