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
	catResultLimit = 100
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

func (s *Storage) getMinMaxAggsQuery(variable *model.Variable) string {
	// get min / max agg names
	minAggName := api.MinAggPrefix + variable.Name
	maxAggName := api.MaxAggPrefix + variable.Name

	// create aggregations
	queryPart := fmt.Sprintf("MIN(\"%s\") AS \"%s\", MAX(\"%s\") AS \"%s\"", variable.Name, minAggName, variable.Name, maxAggName)
	// add aggregations
	return queryPart
}

func (s *Storage) fetchExtrema(dataset string, variable *model.Variable) (*api.Extrema, error) {
	// add min / max aggregation
	aggQuery := s.getMinMaxAggsQuery(variable)

	// create a query that does min and max aggregations for each variable
	queryString := fmt.Sprintf("SELECT %s FROM %s;", aggQuery, dataset)

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

	return s.parseExtrema(res, variable)
}

func (s *Storage) fetchExtremaByURI(storageName string, resultURI string, variable *model.Variable) (*api.Extrema, error) {
	// add min / max aggregation
	aggQuery := s.getMinMaxAggsQuery(variable)

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

func (s *Storage) fetchSummaryData(dataset string, storageName string, varName string, resultURI string, filterParams *api.FilterParams, extrema *api.Extrema) (*api.Histogram, error) {
	// need description of the variables to request aggregation against.
	variable, err := s.metadata.FetchVariable(dataset, varName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch variable description for summary")
	}

	// get the histogram by using the variable type.

	var field Field
	if model.IsNumerical(variable.Type) {
		field = NewNumericalField(s, storageName, variable)
	} else if model.IsCategorical(variable.Type) {
		field = NewCategoricalField(s, storageName, variable)
	} else if model.IsVector(variable.Type) {
		field = NewVectorField(s, storageName, variable)
	} else if model.IsText(variable.Type) {
		field = NewTextField(s, storageName, variable)
	} else if model.IsImage(variable.Type) {
		field = NewImageField(s, storageName, variable)
	} else if model.IsDateTime(variable.Type) {
		field = NewDateTimeField(s, storageName, variable)
	} else if model.IsTimeSeries(variable.Type) {
		field = NewTimeSeriesField(s, storageName, variable)
	} else {
		return nil, errors.Errorf("variable `%s` of type `%s` does not support summary", variable.Name, variable.Type)
	}

	histogram, err := field.FetchSummaryData(resultURI, filterParams, extrema)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch summary data")
	}

	// get number of rows
	numRows, err := s.FetchNumRows(storageName, nil)
	if err != nil {
		return nil, err
	}
	histogram.NumRows = numRows

	// add dataset
	histogram.Dataset = dataset

	return histogram, err
}

// FetchSummary returns the summary for the provided dataset and variable.
func (s *Storage) FetchSummary(dataset string, storageName string, varName string, filterParams *api.FilterParams) (*api.Histogram, error) {
	return s.fetchSummaryData(dataset, storageName, varName, "", filterParams, nil)
}

// FetchSummaryByResult returns the summary for the provided dataset
// and variable for data that is part of the result set.
func (s *Storage) FetchSummaryByResult(dataset string, storageName string, varName string, resultURI string, filterParams *api.FilterParams, extrema *api.Extrema) (*api.Histogram, error) {
	return s.fetchSummaryData(dataset, storageName, varName, resultURI, filterParams, extrema)
}
