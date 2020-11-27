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
	"strings"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/model"
	api "github.com/uncharted-distil/distil/api/model"
)

const (
	catResultLimit           = 100
	timeSeriesCatResultLimit = 10
)

func (s *Storage) parseExtrema(row pgx.Rows, variable *model.Variable) (*api.Extrema, error) {
	var minValue *float64
	var maxValue *float64
	if row != nil {
		// Expect one row of data.
		exists := row.Next()
		if !exists {
			return nil, fmt.Errorf("no row found")
		}
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
	return &api.Extrema{
		Key:  variable.StorageName,
		Type: variable.Type,
		Min:  *minValue,
		Max:  *maxValue,
	}, nil
}

func (s *Storage) parseDateExtrema(row pgx.Rows, variable *model.Variable) (*api.Extrema, error) {
	var minValue *int64
	var maxValue *int64
	if row != nil {
		// Expect one row of data.
		exists := row.Next()
		if !exists {
			return nil, fmt.Errorf("no row found")
		}
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
	return &api.Extrema{
		Key:  variable.StorageName,
		Type: variable.Type,
		Min:  float64(*minValue),
		Max:  float64(*maxValue),
	}, nil
}

func (s *Storage) getMinMaxAggsQuery(variableName string, variableType string) string {
	// get min / max agg names
	minAggName := api.MinAggPrefix + variableName
	maxAggName := api.MaxAggPrefix + variableName

	vName := fmt.Sprintf("\"%s\"", variableName)
	if variableType == model.DateTimeType {
		vName = fmt.Sprintf("CAST(extract(epoch from \"%s\") AS INTEGER)", variableName)
	}

	// create aggregations
	queryPart := fmt.Sprintf("MIN(%s) AS \"%s\", MAX(%s) AS \"%s\"", vName, minAggName, vName, maxAggName)
	// add aggregations
	return queryPart
}

// FetchExtrema return extrema of a variable in a result set.
func (s *Storage) FetchExtrema(storageName string, variable *model.Variable) (*api.Extrema, error) {
	// add min / max aggregation
	aggQuery := s.getMinMaxAggsQuery(variable.StorageName, variable.Type)

	// numerical columns need to filter NaN out
	filter := ""
	if model.IsNumerical(variable.Type) {
		filter = fmt.Sprintf("WHERE \"%s\" != 'NaN'", variable.StorageName)
	}

	// create a query that does min and max aggregations for each variable
	queryString := fmt.Sprintf("SELECT %s FROM %s %s;", aggQuery, storageName, filter)

	// execute the postgres query
	// NOTE: We may want to use the regular Query operation since QueryRow
	// hides db exceptions.
	res, err := s.client.Query(queryString)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch extrema for variable summaries from postgres")
	}
	if res != nil {
		defer res.Close()
	}

	if variable.Type == model.DateTimeType {
		return s.parseDateExtrema(res, variable)
	}
	return s.parseExtrema(res, variable)
}

func (s *Storage) fetchExtremaByURI(storageName string, resultURI string, variable *model.Variable) (*api.Extrema, error) {
	varName := variable.StorageName
	if variable.IsGrouping() && model.IsTimeSeries(variable.Grouping.GetType()) {
		tsg := variable.Grouping.(*model.TimeseriesGrouping)
		varName = tsg.YCol
	} else if variable.IsGrouping() && model.IsGeoCoordinate(variable.Grouping.GetType()) {
		gcg := variable.Grouping.(*model.GeoCoordinateGrouping)
		varName = gcg.YCol
	}

	// add min / max aggregation
	aggQuery := s.getMinMaxAggsQuery(varName, variable.Type)

	// create a query that does min and max aggregations for each variable
	queryString := fmt.Sprintf("SELECT %s FROM %s data INNER JOIN %s result ON data.\"%s\" = result.index WHERE result.result_id = $1;",
		aggQuery, storageName, s.getResultTable(storageName), model.D3MIndexFieldName)

	// execute the postgres query
	// NOTE: We may want to use the regular Query operation since QueryRow
	// hides db exceptions.
	res, err := s.client.Query(queryString, resultURI)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch extrema for variable summaries from postgres")
	}
	if res != nil {
		defer res.Close()
	}

	return s.parseExtrema(res, variable)
}

// FetchExtremaByURI return extrema of a variable in a result set.
func (s *Storage) FetchExtremaByURI(dataset string, storageName string, resultURI string, varName string) (*api.Extrema, error) {
	variable, err := s.metadata.FetchVariable(dataset, varName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch variable description for summary")
	}
	return s.fetchExtremaByURI(storageName, resultURI, variable)
}

func (s *Storage) fetchSummaryData(dataset string, storageName string, varName string, resultURI string, filterParams *api.FilterParams, extrema *api.Extrema, invert bool, mode api.SummaryMode) (*api.VariableSummary, error) {
	// need description of the variables to request aggregation against.
	variable, err := s.metadata.FetchVariable(dataset, varName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch variable description for summary")
	}

	// get the histogram by using the variable type.
	var field Field

	if variable.IsGrouping() {

		if model.IsTimeSeries(variable.Type) {
			tsg := variable.Grouping.(*model.TimeseriesGrouping)

			timeColVar, err := s.metadata.FetchVariable(dataset, tsg.XCol)
			if err != nil {
				return nil, errors.Wrap(err, "failed to fetch variable description for summary")
			}

			valueColVar, err := s.metadata.FetchVariable(dataset, tsg.YCol)
			if err != nil {
				return nil, errors.Wrap(err, "failed to fetch variable description for summary")
			}

			field = NewTimeSeriesField(s, dataset, storageName, tsg.ClusterCol, variable.StorageName, variable.DisplayName, variable.Type,
				variable.Grouping.GetIDCol(), timeColVar.StorageName, timeColVar.Type, valueColVar.StorageName, valueColVar.Type)
		} else if model.IsGeoCoordinate(variable.Grouping.GetType()) {
			gcg := variable.Grouping.(*model.GeoCoordinateGrouping)
			field = NewCoordinateField(variable.StorageName, s, dataset, storageName, gcg.XCol, gcg.YCol, variable.DisplayName, variable.Grouping.GetType(), "")
		} else if model.IsMultiBandImage(variable.Grouping.GetType()) {
			rsg := variable.Grouping.(*model.MultiBandImageGrouping)
			field = NewMultiBandImageField(s, dataset, storageName, rsg.ClusterCol, variable.StorageName, variable.DisplayName, variable.Grouping.GetType(), rsg.IDCol, rsg.BandCol)
		} else if model.IsGeoBounds(variable.Type) {
			gbg := variable.Grouping.(*model.GeoBoundsGrouping)
			field = NewBoundsField(s, dataset, storageName, gbg.CoordinatesCol, gbg.PolygonCol, variable.StorageName, variable.DisplayName, variable.Grouping.GetType(), "")
		} else {
			return nil, errors.Errorf("variable grouping `%s` of type `%s` does not support summary", variable.Grouping.GetIDCol(), variable.Grouping.GetType())
		}
	} else {
		// if timeseries mode, get the grouping field and use that for counts
		countCol := ""
		if mode == api.TimeseriesMode || mode == api.MultiBandImageMode {
			vars, err := s.metadata.FetchVariables(dataset, false, true)
			if err != nil {
				return nil, err
			}
			for _, v := range vars {
				if v.IsGrouping() {
					countCol = v.Grouping.GetIDCol()
				}
			}
		}

		if model.IsNumerical(variable.Type) || model.IsTimestamp(variable.Type) {
			field = NewNumericalField(s, dataset, storageName, variable.StorageName, variable.DisplayName, variable.Type, countCol)
		} else if model.IsCategorical(variable.Type) {
			field = NewCategoricalField(s, dataset, storageName, variable.StorageName, variable.DisplayName, variable.Type, countCol)
		} else if model.IsVector(variable.Type) || model.IsList(variable.Type) {
			field = NewVectorField(s, dataset, storageName, variable.StorageName, variable.DisplayName, variable.Type)
		} else if model.IsText(variable.Type) {
			field = NewTextField(s, dataset, storageName, variable.StorageName, variable.DisplayName, variable.Type, countCol)
		} else if model.IsImage(variable.Type) {
			field = NewImageField(s, dataset, storageName, variable.StorageName, variable.DisplayName, variable.Type, countCol)
		} else if model.IsDateTime(variable.Type) {
			field = NewDateTimeField(s, dataset, storageName, variable.StorageName, variable.DisplayName, variable.Type, countCol)
		} else {
			return nil, errors.Errorf("variable `%s` of type `%s` does not support summary", variable.StorageName, variable.Type)
		}
	}

	summary, err := field.FetchSummaryData(resultURI, filterParams, extrema, invert, mode)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch summary data")
	}

	// add dataset
	summary.Dataset = dataset

	// add description
	summary.Description = variable.Description

	// if there are no filters, and we are returning the exclude set, we expect
	// no results in the filtered set
	if invert && filterParams.Filters == nil {
		summary.EmptyFilteredHistogram()
	}

	return summary, err
}

// FetchSummary returns the summary for the provided dataset and variable.
func (s *Storage) FetchSummary(dataset string, storageName string, varName string, filterParams *api.FilterParams, invert bool, mode api.SummaryMode) (*api.VariableSummary, error) {
	return s.fetchSummaryData(dataset, storageName, varName, "", filterParams, nil, invert, mode)
}

// FetchSummaryByResult returns the summary for the provided dataset
// and variable for data that is part of the result set.
func (s *Storage) FetchSummaryByResult(dataset string, storageName string, varName string, resultURI string, filterParams *api.FilterParams, extrema *api.Extrema, mode api.SummaryMode) (*api.VariableSummary, error) {
	return s.fetchSummaryData(dataset, storageName, varName, resultURI, filterParams, extrema, false, mode)
}

// FetchCategoryCounts fetches the count of each label that occurs for the supplied categorical variable.
func (s *Storage) FetchCategoryCounts(storageName string, variable *model.Variable) (map[string]int, error) {
	if !model.IsCategorical(variable.Type) {
		return nil, errors.Errorf("supplied variable %s is of type %s", variable.StorageName, variable.Type)
	}

	// Run a query to count the categories in the given row
	query := fmt.Sprintf("SELECT \"%s\", COUNT(\"%s\") FROM %s GROUP BY \"%s\"", variable.StorageName, variable.StorageName, storageName, variable.StorageName)
	rows, err := s.client.Query(query)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to count categories for %s", variable.StorageName)
	}

	// Exract into a (category,count) map
	counts := map[string]int{}
	if rows != nil {
		defer rows.Close()

		for rows.Next() {
			var label string
			var count int
			err := rows.Scan(&label, &count)
			if err != nil {
				return nil, err
			}
			counts[label] = count
		}
		err = rows.Err()
		if err != nil {
			return nil, errors.Wrapf(err, "error reading data from postgres")
		}
	}
	return counts, nil
}

// FetchRawDistinctValues fetches the distinct values for a variable from the base table.
func (s *Storage) FetchRawDistinctValues(dataset string, storageName string, varNames []string) ([][]string, error) {
	sql := fmt.Sprintf("SELECT DISTINCT \"%s\" FROM %s_base;", strings.Join(varNames, "\",\""), storageName)
	rows, err := s.client.Query(sql)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read distinct values")
	}
	defer rows.Close()

	// Exract into a (category,count) map
	values := make([][]string, 0)
	for rows.Next() {
		rowValues, err := rows.Values()
		if err != nil {
			return nil, err
		}
		stringVals := make([]string, len(rowValues))
		for i, v := range rowValues {
			stringVals[i] = v.(string)
		}
		values = append(values, stringVals)
	}
	err = rows.Err()
	if err != nil {
		return nil, errors.Wrapf(err, "error reading data from postgres")
	}
	return values, nil
}

// DoesVariableExist returns whether or not a variable exists.
func (s *Storage) DoesVariableExist(dataset string, storageName string, varName string) (bool, error) {
	sql := "SELECT column_name FROM information_schema.columns WHERE table_name = $1 and column_name = $2;"
	rows, err := s.client.Query(sql, storageName, varName)
	if err != nil {
		return false, errors.Wrapf(err, "failed to check if a variable exists in postgres")
	}
	defer rows.Close()

	exists := rows.Next()
	err = rows.Err()
	if err != nil {
		return false, errors.Wrapf(err, "error reading data from postgres")
	}

	return exists, nil
}
