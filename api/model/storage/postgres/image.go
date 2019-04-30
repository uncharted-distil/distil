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
	"math"
	"strings"

	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/model"
	api "github.com/uncharted-distil/distil/api/model"
)

// ImageField defines behaviour for the image field type.
type ImageField struct {
	Storage     *Storage
	StorageName string
	Key         string
	Label       string
	Type        string
}

// NewImageField creates a new field for image types.
func NewImageField(storage *Storage, storageName string, key string, label string, typ string) *ImageField {
	field := &ImageField{
		Storage:     storage,
		StorageName: storageName,
		Key:         key,
		Label:       label,
		Type:        typ,
	}

	return field
}

// FetchSummaryData pulls summary data from the database and builds a histogram.
func (f *ImageField) FetchSummaryData(resultURI string, filterParams *api.FilterParams, extrema *api.Extrema) (*api.Histogram, error) {
	var histogram *api.Histogram
	var err error
	if resultURI == "" {
		histogram, err = f.fetchHistogram(filterParams)
	} else {
		histogram, err = f.fetchHistogramByResult(resultURI, filterParams)
	}

	return histogram, err
}

// FetchTimeseriesSummaryData pulls summary data from the database and builds a histogram.
func (f *ImageField) FetchTimeseriesSummaryData(timeVar *model.Variable, interval int, resultURI string, filterParams *api.FilterParams) (*api.Histogram, error) {
	return nil, fmt.Errorf("not implemented")
}

func (f *ImageField) featureVarName(varName string) string {
	return fmt.Sprintf("%s%s", model.FeatureVarPrefix, varName)
}

func (f *ImageField) fetchRepresentationImages(categoryBuckets []*api.Bucket) ([]string, error) {

	var imageFiles []string

	for _, bucket := range categoryBuckets {

		prefixedVarName := f.featureVarName(f.Key)

		// pull sample row containing bucket
		query := fmt.Sprintf("SELECT \"%s\" FROM %s WHERE \"%s\" ~ $1 LIMIT 1;",
			f.Key, f.StorageName, prefixedVarName)

		// execute the postgres query
		rows, err := f.Storage.client.Query(query, bucket.Key)
		if err != nil {
			return nil, errors.Wrap(err, "failed to fetch histograms for variable summaries from postgres")
		}

		if rows.Next() {
			var imageFile string
			err = rows.Scan(&imageFile)
			if err != nil {
				return nil, errors.Wrap(err, "Unable to parse solution from Postgres")
			}
			imageFiles = append(imageFiles, imageFile)
		}
		rows.Close()
	}
	return imageFiles, nil
}

func (f *ImageField) fetchHistogram(filterParams *api.FilterParams) (*api.Histogram, error) {
	// create the filter for the query.
	wheres := make([]string, 0)
	params := make([]interface{}, 0)
	wheres, params = f.Storage.buildFilteredQueryWhere(wheres, params, filterParams.Filters)

	prefixedVarName := f.featureVarName(f.Key)
	fieldSelect := fmt.Sprintf("unnest(string_to_array(\"%s\", ','))", prefixedVarName)

	where := ""
	if len(wheres) > 0 {
		where = fmt.Sprintf("WHERE %s", strings.Join(wheres, " AND "))
	}

	// Get count by category.
	query := fmt.Sprintf("SELECT %s AS \"%s\", COUNT(*) AS count FROM %s %s GROUP BY %s ORDER BY count desc, %s LIMIT %d;",
		fieldSelect, prefixedVarName, f.StorageName, where, fieldSelect, fieldSelect, catResultLimit)

	// execute the postgres query
	res, err := f.Storage.client.Query(query, params...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch histograms for variable summaries from postgres")
	}
	if res != nil {
		defer res.Close()
	}

	histogram, err := f.parseHistogram(res)
	if err != nil {
		return nil, err
	}

	files, err := f.fetchRepresentationImages(histogram.Buckets)
	if err != nil {
		return nil, err
	}
	histogram.Exemplars = files
	return histogram, nil
}

func (f *ImageField) fetchHistogramByResult(resultURI string, filterParams *api.FilterParams) (*api.Histogram, error) {

	// get filter where / params
	wheres, params, err := f.Storage.buildResultQueryFilters(f.StorageName, resultURI, filterParams)
	if err != nil {
		return nil, err
	}

	params = append(params, resultURI)

	where := ""
	if len(wheres) > 0 {
		where = fmt.Sprintf("AND %s", strings.Join(wheres, " AND "))
	}

	prefixedVarName := f.featureVarName(f.Key)

	// Get count by category.
	query := fmt.Sprintf(
		`SELECT data."%s", COUNT(*) AS count
		 FROM %s data INNER JOIN %s result ON data."%s" = result.index
		 WHERE result.result_id = $%d %s
		 GROUP BY "%s"
		 ORDER BY count desc, "%s" LIMIT %d;`,
		prefixedVarName, f.StorageName, f.Storage.getResultTable(f.StorageName),
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

	histogram, err := f.parseHistogram(res)
	if err != nil {
		return nil, err
	}

	files, err := f.fetchRepresentationImages(histogram.Buckets)
	if err != nil {
		return nil, err
	}
	histogram.Exemplars = files
	return histogram, nil
}

func (f *ImageField) parseHistogram(rows *pgx.Rows) (*api.Histogram, error) {
	prefixedVarName := f.featureVarName(f.Key)

	termsAggName := api.TermsAggPrefix + prefixedVarName

	buckets := make([]*api.Bucket, 0)
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

			buckets = append(buckets, &api.Bucket{
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
	return &api.Histogram{
		Label:   f.Label,
		Key:     f.Key,
		Type:    model.CategoricalType,
		VarType: f.Type,
		Buckets: buckets,
		Extrema: &api.Extrema{
			Min: float64(min),
			Max: float64(max),
		},
	}, nil
}

// FetchPredictedSummaryData pulls predicted data from the result table and builds
// the image histogram for the field.
func (f *ImageField) FetchPredictedSummaryData(resultURI string, datasetResult string, filterParams *api.FilterParams, extrema *api.Extrema) (*api.Histogram, error) {
	targetName := f.featureVarName(f.Key)

	// get filter where / params
	wheres, params, err := f.Storage.buildResultQueryFilters(f.StorageName, resultURI, filterParams)
	if err != nil {
		return nil, err
	}

	wheres = append(wheres, fmt.Sprintf("result.result_id = $%d AND result.target = $%d ", len(params)+1, len(params)+2))
	params = append(params, resultURI, targetName)

	query := fmt.Sprintf(
		`SELECT data."%s", result.value, COUNT(*) AS count
		 FROM %s AS result INNER JOIN %s AS data ON result.index = data."%s"
		 WHERE %s
		 GROUP BY result.value, data."%s"
		 ORDER BY count desc;`,
		targetName, datasetResult, f.StorageName, model.D3MIndexFieldName, strings.Join(wheres, " AND "), targetName)

	// execute the postgres query
	res, err := f.Storage.client.Query(query, params...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch histograms for result summaries from postgres")
	}
	defer res.Close()

	histogram, err := f.parseHistogram(res)
	if err != nil {
		return nil, err
	}

	files, err := f.fetchRepresentationImages(histogram.Buckets)
	if err != nil {
		return nil, err
	}
	histogram.Exemplars = files
	return histogram, nil
}
