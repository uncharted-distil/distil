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
	"strings"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/model"
	api "github.com/uncharted-distil/distil/api/model"
)

// BoundsField defines behaviour for the remote sensing field type.
type BoundsField struct {
	BasicField
	CoordinatesCol string
	PolygonCol     string
}

type geometryBucket struct {
	bounds *model.Bounds
	x      int
	y      int
}

func (g *geometryBucket) toGeometryString() string {
	return buildBoundsGeometryString(g.bounds)
}

// NewBoundsField creates a new field for remote sensing types.
func NewBoundsField(storage *Storage, datasetName string, datasetStorageName string,
	coordinatesCol string, polygonCol string, key string, label string, typ string, count string) *BoundsField {
	count = getCountSQL(count)

	field := &BoundsField{
		BasicField: BasicField{
			Key:                key,
			Storage:            storage,
			DatasetName:        datasetName,
			DatasetStorageName: datasetStorageName,
			Label:              label,
			Type:               typ,
			Count:              count,
		},
		CoordinatesCol: coordinatesCol,
		PolygonCol:     polygonCol,
	}

	return field
}

// FetchSummaryData pulls summary data from the database and builds a histogram.
func (f *BoundsField) FetchSummaryData(resultURI string, filterParams *api.FilterParams, extrema *api.Extrema, invert bool, mode api.SummaryMode) (*api.VariableSummary, error) {
	var baseline *api.Histogram
	var filtered *api.Histogram
	var err error

	// update the highlight key to use the cluster if necessary
	if err = f.updateClusterHighlight(filterParams, mode); err != nil {
		return nil, err
	}

	if resultURI == "" {
		baseline, err = f.fetchHistogram(api.GetBaselineFilter(filterParams), invert, coordinateBuckets)
		if err != nil {
			return nil, err
		}
		if !filterParams.Empty(true) {
			filtered, err = f.fetchHistogram(filterParams, invert, coordinateBuckets)
			if err != nil {
				return nil, err
			}
		}
	} else {
		baseline, err = f.fetchHistogramByResult(resultURI, api.GetBaselineFilter(filterParams), coordinateBuckets)
		if err != nil {
			return nil, err
		}
		if !filterParams.Empty(true) {
			filtered, err = f.fetchHistogramByResult(resultURI, filterParams, coordinateBuckets)
			if err != nil {
				return nil, err
			}
		}
	}

	return &api.VariableSummary{
		Key:      f.Key,
		Label:    f.Label,
		Type:     model.GeoBoundsType,
		VarType:  f.Type,
		Baseline: baseline,
		Filtered: filtered,
		Timeline: nil,
	}, nil
}

func (f *BoundsField) fetchHistogram(filterParams *api.FilterParams, invert bool, numBuckets int) (*api.Histogram, error) {
	// create the filter for the query.
	wheres := make([]string, 0)
	params := make([]interface{}, 0)
	wheres, params = f.Storage.buildFilteredQueryWhere(f.GetDatasetName(), wheres, params, "", filterParams, invert)

	where := ""
	if len(wheres) > 0 {
		where = fmt.Sprintf("WHERE %s", strings.Join(wheres, " AND "))
	}

	// get the extrema for each axis
	xExtrema, yExtrema, err := f.fetchCombinedExtrema()
	if err != nil {
		return nil, err
	}

	xNumBuckets, yNumBuckets := getEqualBivariateBuckets(numBuckets, xExtrema, yExtrema)
	buckets := getGeoBoundsBuckets(xExtrema, yExtrema, xNumBuckets, yNumBuckets)

	// insert all the buckets into a temp table
	tx, err := f.Storage.client.Begin()
	if err != nil {
		return nil, errors.Wrapf(err, "unable to begin transaction")
	}
	tmpTableName := "tmp_" + f.DatasetStorageName
	err = f.prepareBucketsForQuery(tx, tmpTableName, buckets)
	if err != nil {
		return nil, err
	}

	// join the temp table to the main table to get bucketed data (use base table for index performance)
	queryTableName := getBaseTableName(f.DatasetStorageName)

	//TODO: TEST WITH ST_INTERSECTS
	//	ST_WITHIN WILL UNDERCOUNT THOSE THAT CROSS BOUNDARIES
	//	ST_INTERSECTS WILL OVERCOUNT THOSE THAT CROSS BOUNDARIES
	query := fmt.Sprintf(`
		SELECT b.xbuckets, b.xcoord, b.ybuckets, b.ycoord, COUNT(%s)
		FROM %s AS d inner join %s AS b ON ST_WITHIN(d."%s", b.coordinates) %s
		GROUP BY b.xbuckets, b.xcoord, b.ybuckets, b.ycoord
		ORDER BY b.xbuckets, b.ybuckets;`, f.Count, queryTableName, tmpTableName, f.PolygonCol, where)
	res, err := tx.Query(context.Background(), query, params...)
	if err != nil {
		_ = tx.Rollback(context.Background())
		return nil, errors.Wrapf(err, "unable to query for histogram")
	}

	histogram, err := f.parseHistogram(res, xExtrema, yExtrema, xNumBuckets, yNumBuckets)
	if err != nil {
		_ = tx.Rollback(context.Background())
		return nil, err
	}
	err = tx.Commit(context.Background())
	if err != nil {
		return nil, errors.Wrapf(err, "unable to commit transaction")
	}

	return histogram, nil
}

func (f *BoundsField) prepareBucketsForQuery(tx pgx.Tx, tmpTableName string, buckets []*geometryBucket) error {
	tmpTableCreateSQL := fmt.Sprintf(`
		CREATE TEMPORARY TABLE "%s" (
			xbuckets INT,
			xcoord DOUBLE PRECISION,
			ybuckets INT,
			ycoord DOUBLE PRECISION,
			coordinates geometry
		) ON COMMIT DROP`, tmpTableName)
	_, err := tx.Exec(context.Background(), tmpTableCreateSQL)
	if err != nil {
		_ = tx.Rollback(context.Background())
		return errors.Wrapf(err, "unable to create temp table")
	}

	bucketSQL := `INSERT INTO "%s" (xbuckets, xcoord, ybuckets, ycoord, coordinates) VALUES (%d, %g, %d, %g, '%s'::geometry)`
	insertSQLs := []string{}
	for _, b := range buckets {
		insertSQLs = append(insertSQLs, fmt.Sprintf(bucketSQL, tmpTableName, b.x, b.bounds.MinX, b.y, b.bounds.MinY, b.toGeometryString()))
	}
	_, err = tx.Exec(context.Background(), strings.Join(insertSQLs, "; "))
	if err != nil {
		_ = tx.Rollback(context.Background())
		return errors.Wrapf(err, "unable to insert to temp table")
	}

	return nil
}

func (f *BoundsField) fetchCombinedExtrema() (*api.Extrema, *api.Extrema, error) {
	// add min / max aggregation
	aggQueryX := f.getMinMaxAggsQuery(f.PolygonCol, "x", "X")
	aggQueryY := f.getMinMaxAggsQuery(f.PolygonCol, "y", "Y")

	// create a query that does min and max aggregations for each variable
	queryString := fmt.Sprintf("SELECT %s, %s FROM %s;", aggQueryX, aggQueryY, f.DatasetStorageName)

	// execute the postgres query
	res, err := f.Storage.client.Query(queryString)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to fetch extrema for variable summaries from postgres")
	}
	if res != nil {
		defer res.Close()
	}

	return f.parseExtrema(res)
}

func (f *BoundsField) getMinMaxAggsQuery(key string, label string, axisLabel string) string {
	// get min / max agg names
	minAggName := api.MinAggPrefix + label
	maxAggName := api.MaxAggPrefix + label

	// create aggregations
	queryPart := fmt.Sprintf("MIN(ST_%sMin(\"%s\")) AS \"%s\", MAX(ST_%sMax(\"%s\")) AS \"%s\"",
		axisLabel, key, minAggName, axisLabel, key, maxAggName)

	return queryPart
}

func (f *BoundsField) parseExtrema(rows pgx.Rows) (*api.Extrema, *api.Extrema, error) {
	var minXValue *float64
	var maxXValue *float64
	var minYValue *float64
	var maxYValue *float64
	if rows != nil {
		// Expect one row of data.
		exists := rows.Next()
		if !exists {
			return nil, nil, fmt.Errorf("no rows in extrema query result")
		}
		err := rows.Scan(&minXValue, &maxXValue, &minYValue, &maxYValue)
		if err != nil {
			return nil, nil, errors.Wrap(err, "no min / max aggregation found")
		}
	}
	// check values exist
	if minXValue == nil || maxXValue == nil || minYValue == nil || maxYValue == nil {
		return nil, nil, errors.Errorf("no min / max aggregation values found")
	}
	// assign attributes
	xExtrema := &api.Extrema{
		Key:  f.Key,
		Type: model.LongitudeType,
		Min:  *minXValue,
		Max:  *maxXValue,
	}
	yExtrema := &api.Extrema{
		Key:  f.Key,
		Type: model.LatitudeType,
		Min:  *minYValue,
		Max:  *maxYValue,
	}

	return xExtrema, yExtrema, nil
}

func (f *BoundsField) fetchHistogramByResult(resultURI string, filterParams *api.FilterParams, numBuckets int) (*api.Histogram, error) {
	// get filter where / params
	wheres, params, err := f.Storage.buildResultQueryFilters(f.GetDatasetName(), f.DatasetStorageName, resultURI, filterParams, baseTableAlias)
	if err != nil {
		return nil, err
	}

	params = append(params, resultURI)

	where := ""
	if len(wheres) > 0 {
		where = fmt.Sprintf("AND %s", strings.Join(wheres, " AND "))
	}

	// get the extrema for each axis
	xExtrema, yExtrema, err := f.fetchCombinedExtrema()
	if err != nil {
		return nil, err
	}

	xNumBuckets, yNumBuckets := getEqualBivariateBuckets(numBuckets, xExtrema, yExtrema)
	buckets := getGeoBoundsBuckets(xExtrema, yExtrema, xNumBuckets, yNumBuckets)

	// insert all the buckets into a temp table
	tx, err := f.Storage.client.Begin()
	if err != nil {
		return nil, errors.Wrapf(err, "unable to begin transaction")
	}
	tmpTableName := "tmp_" + f.DatasetStorageName
	err = f.prepareBucketsForQuery(tx, tmpTableName, buckets)
	if err != nil {
		return nil, err
	}

	// join the temp table to the main table to get bucketed data, using the base table for index performance.
	queryTableName := getBaseTableName(f.DatasetStorageName)
	query := fmt.Sprintf(`
		SELECT b.xbuckets, b.xcoord, b.ybuckets, b.ycoord, COUNT(%s)
		FROM %s AS data inner join %s AS b ON ST_WITHIN(data."%s", b.coordinates)
		INNER JOIN %s result ON cast(data."%s" as double precision) = result.index
		WHERE result.result_id = $%d %s
		GROUP BY b.xbuckets, b.xcoord, b.ybuckets, b.ycoord
		ORDER BY b.xbuckets, b.ybuckets;`, f.Count, queryTableName, tmpTableName, f.PolygonCol,
		f.Storage.getResultTable(f.DatasetStorageName), model.D3MIndexFieldName, len(params), where)
	res, err := tx.Query(context.Background(), query, params...)
	if err != nil {
		_ = tx.Rollback(context.Background())
		return nil, errors.Wrapf(err, "failed to fetch histograms for variable summaries from postgres")
	}
	if res != nil {
		defer res.Close()
	}

	histogram, err := f.parseHistogram(res, xExtrema, yExtrema, xNumBuckets, yNumBuckets)
	if err != nil {
		_ = tx.Rollback(context.Background())
		return nil, err
	}
	err = tx.Commit(context.Background())
	if err != nil {
		return nil, errors.Wrapf(err, "unable to commit transaction")
	}

	return histogram, nil
}

func (f *BoundsField) parseHistogram(rows pgx.Rows, xExtrema *api.Extrema, yExtrema *api.Extrema, xNumBuckets int, yNumBuckets int) (*api.Histogram, error) {
	// get histogram agg name
	histogramAggName := api.HistogramAggPrefix + f.Key

	// Parse bucket results.
	xInterval := xExtrema.GetBucketInterval(xNumBuckets)
	yInterval := yExtrema.GetBucketInterval(yNumBuckets)
	xRounded := xExtrema.GetBucketMinMax(xNumBuckets)
	yRounded := yExtrema.GetBucketMinMax(yNumBuckets)

	xBucketCount := int64(xExtrema.GetBucketCount(xNumBuckets))
	yBucketCount := int64(yExtrema.GetBucketCount(yNumBuckets))
	xBuckets := make([]*api.Bucket, xBucketCount)

	// initialize empty histogram structure
	i := 0
	for xVal := xRounded.Min; i < int(xBucketCount); xVal += xInterval {
		yBuckets := make([]*api.Bucket, yBucketCount)
		j := 0
		for yVal := yRounded.Min; j < int(yBucketCount); yVal += yInterval {
			yBuckets[j] = &api.Bucket{
				Key:     fmt.Sprintf("%f", yVal),
				Count:   0,
				Buckets: nil,
			}
			j++
		}
		xBuckets[i] = &api.Bucket{
			Key:     fmt.Sprintf("%f", xVal),
			Count:   0,
			Buckets: yBuckets,
		}
		i++
	}

	for rows.Next() {
		var xBucketValue float64
		var yBucketValue float64
		var xBucket int64
		var yBucket int64
		var yRowBucketCount int64
		err := rows.Scan(&xBucket, &xBucketValue, &yBucket, &yBucketValue, &yRowBucketCount)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("no %s histogram aggregation found", histogramAggName))
		}

		// Due to float representation, sometimes the lowest value <
		// first bucket interval and so ends up in bucket -1.
		// Since the max can match the limit, an extra bucket may exist.
		// Add the value to the second to last bucket.
		if xBucket < 0 {
			xBucket = 0
		} else if xBucket >= xBucketCount {
			xBucket = xBucketCount - 1
		}
		if yBucket < 0 {
			yBucket = 0
		} else if yBucket >= yBucketCount {
			yBucket = yBucketCount - 1
		}
		xBuckets[xBucket].Buckets[yBucket].Count += yRowBucketCount
	}
	err := rows.Err()
	if err != nil {
		return nil, errors.Wrapf(err, "error reading data from postgres")
	}

	// assign histogram attributes
	return &api.Histogram{
		Buckets: xBuckets,
	}, nil
}

// FetchPredictedSummaryData pulls predicted data from the result table and builds
// the coordinate histogram for the field.
func (f *BoundsField) FetchPredictedSummaryData(resultURI string, datasetResult string, filterParams *api.FilterParams, extrema *api.Extrema, mode api.SummaryMode) (*api.VariableSummary, error) {
	return nil, fmt.Errorf("not implemented")
}

func buildBoundsGeometryString(bounds *model.Bounds) string {
	coords := []string{
		pointToString(bounds.MinX, bounds.MinY, " "),
		pointToString(bounds.MinX, bounds.MaxY, " "),
		pointToString(bounds.MaxX, bounds.MaxY, " "),
		pointToString(bounds.MaxX, bounds.MinY, " "),
		pointToString(bounds.MinX, bounds.MinY, " "),
	}
	return fmt.Sprintf("POLYGON((%s))", strings.Join(coords, ","))
}

func pointToString(x float64, y float64, separator string) string {
	return fmt.Sprintf("%f%s%f", x, separator, y)
}

func getGeoBoundsBuckets(xExtrema *api.Extrema, yExtrema *api.Extrema,
	xNumBuckets int, yNumBuckets int) []*geometryBucket {
	// build a list of bounds representing the buckets
	xInterval := xExtrema.GetBucketInterval(xNumBuckets)
	yInterval := yExtrema.GetBucketInterval(yNumBuckets)

	buckets := []*geometryBucket{}
	xLeft := xExtrema.Min
	xRight := xLeft + xInterval
	for xCount := 0; xLeft <= xExtrema.Max; xLeft, xRight = xLeft+xInterval, xRight+xInterval {
		yBottom := yExtrema.Min
		yTop := yBottom + yInterval
		for yCount := 0; yBottom <= yExtrema.Max; yBottom, yTop = yBottom+yInterval, yTop+yInterval {
			buckets = append(buckets, &geometryBucket{
				bounds: &model.Bounds{
					MinX: xLeft,
					MaxX: xRight,
					MinY: yBottom,
					MaxY: yTop,
				},
				x: xCount,
				y: yCount,
			})
			yCount++
		}
		xCount++
	}

	return buckets
}
