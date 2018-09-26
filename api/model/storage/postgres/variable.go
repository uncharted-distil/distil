package postgres

import (
	"fmt"

	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil/api/model"
)

const (
	catResultLimit = 100
)

func (s *Storage) parseExtrema(row *pgx.Rows, variable *model.Variable) (*model.Extrema, error) {
	var minValue *float64
	var maxValue *float64
	if row != nil {
		// Expect one row of data.
		row.Next()
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
	return &model.Extrema{
		Key:  variable.Key,
		Type: variable.Type,
		Min:  *minValue,
		Max:  *maxValue,
	}, nil
}

func (s *Storage) getMinMaxAggsQuery(variable *model.Variable) string {
	// get min / max agg names
	minAggName := model.MinAggPrefix + variable.Key
	maxAggName := model.MaxAggPrefix + variable.Key

	// create aggregations
	queryPart := fmt.Sprintf("MIN(\"%s\") AS \"%s\", MAX(\"%s\") AS \"%s\"", variable.Key, minAggName, variable.Key, maxAggName)
	// add aggregations
	return queryPart
}

func (s *Storage) fetchExtrema(dataset string, variable *model.Variable) (*model.Extrema, error) {
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

func (s *Storage) fetchExtremaByURI(dataset string, resultURI string, variable *model.Variable) (*model.Extrema, error) {
	// add min / max aggregation
	aggQuery := s.getMinMaxAggsQuery(variable)

	// create a query that does min and max aggregations for each variable
	queryString := fmt.Sprintf("SELECT %s FROM %s data INNER JOIN %s result ON data.\"%s\" = result.index WHERE result.result_id = $1;",
		aggQuery, dataset, s.getResultTable(dataset), model.D3MIndexFieldName)

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
func (s *Storage) FetchExtremaByURI(dataset string, resultURI string, varName string) (*model.Extrema, error) {

	variable, err := s.metadata.FetchVariable(dataset, varName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch variable description for summary")
	}
	return s.fetchExtremaByURI(dataset, resultURI, variable)
}

func (s *Storage) fetchSummaryData(dataset string, varName string, resultURI string, filterParams *model.FilterParams, extrema *model.Extrema) (*model.Histogram, error) {
	// need description of the variables to request aggregation against.
	variable, err := s.metadata.FetchVariable(dataset, varName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch variable description for summary")
	}

	// get the histogram by using the variable type.

	var field Field
	if model.IsNumerical(variable.Type) {
		field = NewNumericalField(s)
	} else if model.IsCategorical(variable.Type) {
		field = NewCategoricalField(s)
	} else if model.IsVector(variable.Type) {
		field = NewVectorField(s)
	} else if model.IsText(variable.Type) {
		field = NewTextField(s)
	} else if model.IsImage(variable.Type) {
		field = NewImageField(s)
	} else {
		/*else if model.IsTimeSeries(variable.Type) {
			field = NewTimeSeries(s)
		}*/
		return nil, errors.Errorf("variable %s of type %s does not support summary", variable.Key, variable.Type)
	}

	histogram, err := field.FetchSummaryData(dataset, variable, resultURI, filterParams, extrema)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch summary data")
	}

	// get number of rows
	numRows, err := s.FetchNumRows(dataset, nil)
	if err != nil {
		return nil, err
	}
	histogram.NumRows = numRows

	// add dataset
	histogram.Dataset = dataset

	return histogram, err
}

// FetchSummary returns the summary for the provided dataset and variable.
func (s *Storage) FetchSummary(dataset string, varName string, filterParams *model.FilterParams) (*model.Histogram, error) {
	return s.fetchSummaryData(dataset, varName, "", filterParams, nil)
}

// FetchSummaryByResult returns the summary for the provided dataset
// and variable for data that is part of the result set.
func (s *Storage) FetchSummaryByResult(dataset string, varName string, resultURI string, filterParams *model.FilterParams, extrema *model.Extrema) (*model.Histogram, error) {
	return s.fetchSummaryData(dataset, varName, resultURI, filterParams, extrema)
}
