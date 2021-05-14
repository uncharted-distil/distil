//
//   Copyright © 2021 Uncharted Software Inc.
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
		Key:  variable.Key,
		Type: variable.Type,
		Min:  *minValue,
		Max:  *maxValue,
	}, nil
}

// IsKey verifies if a given set of variables is unique across the dataset.
func (s *Storage) IsKey(dataset string, storageName string, variables []*model.Variable) (bool, error) {
	// use a group by on the variables and retain only rows with more than a single result
	columnNames := []string{}
	for _, v := range variables {
		columnNames = append(columnNames, fmt.Sprintf("\"%s\"", v.Key))
	}
	grouping := strings.Join(columnNames, ",")

	sql := fmt.Sprintf("select count(*) from (select %s, count(*) from %s group by %s having count(*) > 1) as d;",
		grouping, storageName, grouping)
	rows, err := s.client.Query(sql)
	if err != nil {
		return false, errors.Wrapf(err, "unable execute query to verify key validity")
	}
	defer rows.Close()

	// the column combination is not a key if there are any rows in the result rows
	badKeyCount := 0
	if rows.Next() {
		err = rows.Scan(&badKeyCount)
		if err != nil {
			return false, errors.Wrap(err, "unable to read bad key count")
		}
	}

	return badKeyCount == 0, nil
}

// FetchExtrema return extrema of a variable in a result set.
func (s *Storage) FetchExtrema(dataset string, storageName string, variable *model.Variable) (*api.Extrema, error) {
	field, err := s.createField(dataset, storageName, variable, api.DefaultMode)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch extrema for variable summaries from postgres")
	}

	return field.fetchExtremaStorage()
}

// FetchExtremaByURI return extrema of a variable in a result set.
func (s *Storage) FetchExtremaByURI(dataset string, storageName string, resultURI string, varName string) (*api.Extrema, error) {
	variable, err := s.metadata.FetchVariable(dataset, varName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch variable description for summary")
	}

	field, err := s.createField(dataset, storageName, variable, api.DefaultMode)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch extrema for variable summaries from postgres")
	}

	return field.fetchExtremaByURI(resultURI)
}

func (s *Storage) fetchSummaryData(dataset string, storageName string, varName string, resultURI string, filterParams *api.FilterParams, extrema *api.Extrema, mode api.SummaryMode) (*api.VariableSummary, error) {
	// need description of the variables to request aggregation against.
	variable, err := s.metadata.FetchVariable(dataset, varName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch variable description for summary")
	}

	// get the histogram by using the variable type.
	field, err := s.createField(dataset, storageName, variable, mode)
	if err != nil {
		return nil, err
	}

	summary, err := field.FetchSummaryData(resultURI, filterParams, extrema, mode)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch summary data")
	}

	// add dataset
	summary.Dataset = dataset

	// add description
	summary.Description = variable.Description

	// if there are no filters, and we are returning the exclude set, we expect
	// no results in the filtered set
	if filterParams.Invert && filterParams.Filters == nil {
		summary.EmptyFilteredHistogram()
	}

	return summary, err
}

// FetchSummary returns the summary for the provided dataset and variable.
func (s *Storage) FetchSummary(dataset string, storageName string, varName string, filterParams *api.FilterParams, mode api.SummaryMode) (*api.VariableSummary, error) {
	return s.fetchSummaryData(dataset, storageName, varName, "", filterParams, nil, mode)
}

// FetchSummaryByResult returns the summary for the provided dataset
// and variable for data that is part of the result set.
func (s *Storage) FetchSummaryByResult(dataset string, storageName string, varName string, resultURI string, filterParams *api.FilterParams, extrema *api.Extrema, mode api.SummaryMode) (*api.VariableSummary, error) {
	return s.fetchSummaryData(dataset, storageName, varName, resultURI, filterParams, extrema, mode)
}

// FetchCategoryCounts fetches the count of each label that occurs for the supplied categorical variable.
func (s *Storage) FetchCategoryCounts(storageName string, variable *model.Variable) (map[string]int, error) {
	if !model.IsCategorical(variable.Type) {
		return nil, errors.Errorf("supplied variable %s is of type %s", variable.Key, variable.Type)
	}

	// Run a query to count the categories in the given row
	query := fmt.Sprintf("SELECT \"%s\", COUNT(\"%s\") FROM %s GROUP BY \"%s\"", variable.Key, variable.Key, storageName, variable.Key)
	rows, err := s.client.Query(query)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to count categories for %s", variable.Key)
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

func (s *Storage) createField(dataset string, storageName string, variable *model.Variable, mode api.SummaryMode) (Field, error) {
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

			field = NewTimeSeriesField(s, dataset, storageName, tsg.ClusterCol, variable.Key, variable.DisplayName, variable.Type,
				variable.Grouping.GetIDCol(), timeColVar.Key, timeColVar.Type, valueColVar.Key, valueColVar.Type)
		} else if model.IsGeoCoordinate(variable.Grouping.GetType()) {
			gcg := variable.Grouping.(*model.GeoCoordinateGrouping)
			field = NewCoordinateField(variable.Key, s, dataset, storageName, gcg.XCol, gcg.YCol, variable.DisplayName, variable.Grouping.GetType(), "")
		} else if model.IsMultiBandImage(variable.Grouping.GetType()) {
			rsg := variable.Grouping.(*model.MultiBandImageGrouping)
			field = NewMultiBandImageField(s, dataset, storageName, rsg.ClusterCol, variable.Key, variable.DisplayName, variable.Grouping.GetType(), rsg.IDCol, rsg.BandCol)
		} else if model.IsGeoBounds(variable.Type) {
			gbg := variable.Grouping.(*model.GeoBoundsGrouping)
			field = NewBoundsField(s, dataset, storageName, gbg.CoordinatesCol, gbg.PolygonCol, variable.Key, variable.DisplayName, variable.Grouping.GetType(), "")
		} else {
			return nil, errors.Errorf("variable grouping `%s` of type `%s` does not support summary", variable.Grouping.GetIDCol(), variable.Grouping.GetType())
		}
	} else {
		// if timeseries mode, get the grouping field and use that for counts
		countCol := ""
		if mode == api.TimeseriesMode || mode == api.MultiBandImageMode {
			vars, err := s.metadata.FetchVariables(dataset, false, true, false)
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
			field = NewNumericalField(s, dataset, storageName, variable.Key, variable.DisplayName, variable.Type, countCol)
		} else if model.IsCategorical(variable.Type) {
			field = NewCategoricalField(s, dataset, storageName, variable.Key, variable.DisplayName, variable.Type, countCol)
		} else if model.IsVector(variable.Type) || model.IsList(variable.Type) {
			field = NewVectorField(s, dataset, storageName, variable.Key, variable.DisplayName, variable.Type)
		} else if model.IsText(variable.Type) {
			field = NewTextField(s, dataset, storageName, variable.Key, variable.DisplayName, variable.Type, countCol)
		} else if model.IsImage(variable.Type) {
			field = NewImageField(s, dataset, storageName, variable.Key, variable.DisplayName, variable.Type, countCol)
		} else if model.IsDateTime(variable.Type) {
			field = NewDateTimeField(s, dataset, storageName, variable.Key, variable.DisplayName, variable.Type, countCol)
		} else {
			return nil, errors.Errorf("variable `%s` of type `%s` does not support summary", variable.Key, variable.Type)
		}
	}

	return field, nil
}
