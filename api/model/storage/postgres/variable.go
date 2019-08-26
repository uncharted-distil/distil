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

	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/model"
	api "github.com/uncharted-distil/distil/api/model"
)

const (
	catResultLimit           = 100
	timeSeriesCatResultLimit = 10
)

func (s *Storage) parseExtrema(row *pgx.Rows, variable *model.Variable) (*api.Extrema, error) {
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
		Key:  variable.Name,
		Type: variable.Type,
		Min:  *minValue,
		Max:  *maxValue,
	}, nil
}

func (s *Storage) parseDateExtrema(row *pgx.Rows, variable *model.Variable) (*api.Extrema, error) {
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
		Key:  variable.Name,
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
	aggQuery := s.getMinMaxAggsQuery(variable.Name, variable.Type)

	// numerical columns need to filter NaN out
	filter := ""
	if model.IsNumerical(variable.Type) {
		filter = fmt.Sprintf("WHERE \"%s\" != 'NaN'", variable.Name)
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
	varName := variable.Name
	if variable.Grouping != nil {
		varName = variable.Grouping.Properties.YCol
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

func (s *Storage) fetchSummaryData(dataset string, storageName string, varName string, resultURI string, filterParams *api.FilterParams, extrema *api.Extrema, invert bool) (*api.VariableSummary, error) {
	// need description of the variables to request aggregation against.
	variable, err := s.metadata.FetchVariable(dataset, varName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch variable description for summary")
	}

	// get the histogram by using the variable type.
	var field Field

	if variable.Grouping != nil {

		if model.IsTimeSeries(variable.Grouping.Type) {

			timeColVar, err := s.metadata.FetchVariable(dataset, variable.Grouping.Properties.XCol)
			if err != nil {
				return nil, errors.Wrap(err, "failed to fetch variable description for summary")
			}

			valueColVar, err := s.metadata.FetchVariable(dataset, variable.Grouping.Properties.YCol)
			if err != nil {
				return nil, errors.Wrap(err, "failed to fetch variable description for summary")
			}

			field = NewTimeSeriesField(s, storageName, variable.Grouping.Properties.ClusterCol, variable.Grouping.IDCol, variable.Grouping.IDCol, variable.Grouping.Type,
				timeColVar.Name, timeColVar.Type, valueColVar.Name, valueColVar.Type)

		} else {
			return nil, errors.Errorf("variable grouping `%s` of type `%s` does not support summary", variable.Grouping.IDCol, variable.Grouping.Type)
		}

	} else {

		if model.IsNumerical(variable.Type) || model.IsTimestamp(variable.Type) {
			field = NewNumericalField(s, storageName, variable.Name, variable.DisplayName, variable.Type)
		} else if model.IsCategorical(variable.Type) {
			field = NewCategoricalField(s, storageName, variable.Name, variable.DisplayName, variable.Type)
		} else if model.IsVector(variable.Type) {
			field = NewVectorField(s, storageName, variable.Name, variable.DisplayName, variable.Type)
		} else if model.IsText(variable.Type) {
			field = NewTextField(s, storageName, variable.Name, variable.DisplayName, variable.Type)
		} else if model.IsImage(variable.Type) {
			field = NewImageField(s, storageName, variable.Name, variable.DisplayName, variable.Type)
		} else if model.IsDateTime(variable.Type) {
			field = NewDateTimeField(s, storageName, variable.Name, variable.DisplayName, variable.Type)
		} else {
			return nil, errors.Errorf("variable `%s` of type `%s` does not support summary", variable.Name, variable.Type)
		}

	}

	summary, err := field.FetchSummaryData(resultURI, filterParams, extrema, invert)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch summary data")
	}

	// add dataset
	summary.Dataset = dataset

	if variable.Grouping != nil {
		if model.IsTimeSeries(variable.Grouping.Type) {
			summary.Label = variable.Grouping.Properties.YCol
		}
	}

	// if there are no filters, and we are returning the exclude set, we expect
	// no results in the filtered set
	if invert && filterParams.Filters == nil {
		summary.EmptyFilteredHistogram()
	}

	return summary, err
}

// FetchSummary returns the summary for the provided dataset and variable.
func (s *Storage) FetchSummary(dataset string, storageName string, varName string, filterParams *api.FilterParams, invert bool) (*api.VariableSummary, error) {
	return s.fetchSummaryData(dataset, storageName, varName, "", filterParams, nil, invert)
}

// FetchSummaryByResult returns the summary for the provided dataset
// and variable for data that is part of the result set.
func (s *Storage) FetchSummaryByResult(dataset string, storageName string, varName string, resultURI string, filterParams *api.FilterParams, extrema *api.Extrema) (*api.VariableSummary, error) {
	return s.fetchSummaryData(dataset, storageName, varName, resultURI, filterParams, extrema, false)
}

func (s *Storage) fetchTimeseriesSummary(dataset string, storageName string, xColName string, yColName string, resultURI string, interval int, filterParams *api.FilterParams, invert bool) (*api.VariableSummary, error) {

	// need description of the variables to request aggregation against.
	timeColVar, err := s.metadata.FetchVariable(dataset, xColName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch variable description for summary")
	}
	variable, err := s.metadata.FetchVariable(dataset, yColName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch variable description for summary")
	}

	// get the histogram by using the variable type.
	var field Field

	if variable.Grouping != nil {
		return nil, errors.Errorf("not implemented")
	}

	if model.IsNumerical(variable.Type) {
		field = NewNumericalField(s, storageName, variable.Name, variable.DisplayName, variable.Type)
	} else if model.IsCategorical(variable.Type) {
		field = NewCategoricalField(s, storageName, variable.Name, variable.DisplayName, variable.Type)
	} else if model.IsVector(variable.Type) {
		field = NewVectorField(s, storageName, variable.Name, variable.DisplayName, variable.Type)
	} else if model.IsText(variable.Type) {
		field = NewTextField(s, storageName, variable.Name, variable.DisplayName, variable.Type)
	} else if model.IsImage(variable.Type) {
		field = NewImageField(s, storageName, variable.Name, variable.DisplayName, variable.Type)
	} else if model.IsDateTime(variable.Type) {
		field = NewDateTimeField(s, storageName, variable.Name, variable.DisplayName, variable.Type)
	} else {
		return nil, errors.Errorf("variable `%s` of type `%s` does not support summary", variable.Name, variable.Type)
	}

	timeseries, err := field.FetchTimeseriesSummaryData(timeColVar, interval, resultURI, filterParams, invert)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch summary data")
	}

	// get number of rows
	timeseries.Type = "timeseries"

	// add dataset
	timeseries.Dataset = dataset

	return timeseries, err
}

// FetchTimeseriesSummary fetches a timeseries.
func (s *Storage) FetchTimeseriesSummary(dataset string, storageName string, xColName string, yColName string, interval int, filterParams *api.FilterParams, invert bool) (*api.VariableSummary, error) {
	return s.fetchTimeseriesSummary(dataset, storageName, xColName, yColName, "", interval, filterParams, invert)
}

// FetchTimeseriesSummaryByResult fetches a timeseries for a given result.
func (s *Storage) FetchTimeseriesSummaryByResult(dataset string, storageName string, xColName string, yColName string, interval int, resultURI string, filterParams *api.FilterParams) (*api.VariableSummary, error) {
	return s.fetchTimeseriesSummary(dataset, storageName, xColName, yColName, resultURI, interval, filterParams, false)
}
