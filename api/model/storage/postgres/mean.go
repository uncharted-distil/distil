package postgres

import (
	"fmt"
	"strings"

	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil/api/model"
)

// FetchMean returns the mean for a given dataset and variable.
func (s *Storage) FetchMean(dataset string, variable *model.Variable, filterParams *model.FilterParams) (float64, error) {
	// create the filter for the query.
	wheres := make([]string, 0)
	params := make([]interface{}, 0)
	wheres, params = s.buildFilteredQueryWhere(wheres, params, dataset, filterParams.Filters)

	where := ""
	if len(wheres) > 0 {
		where = fmt.Sprintf("WHERE %s", strings.Join(wheres, " AND "))
	}

	// Create the complete query string.
	query := fmt.Sprintf("SELECT avg(\"%s\") FROM %s %s;", variable.Key, dataset, where)

	// execute the postgres query
	res, err := s.client.Query(query, params...)
	if err != nil {
		return 0, errors.Wrap(err, "failed to fetch histograms for variable summaries from postgres")
	}
	if res != nil {
		defer res.Close()
	}

	return s.parseMean(res)
}

// FetchMeanByResult returns the mean for a given dataset, variable, and result.
func (s *Storage) FetchMeanByResult(dataset string, variable *model.Variable, resultURI string, filterParams *model.FilterParams) (float64, error) {
	// get filter where / params
	wheres, params, err := s.buildResultQueryFilters(dataset, resultURI, filterParams)
	if err != nil {
		return 0, err
	}

	params = append(params, resultURI)

	where := ""
	if len(wheres) > 0 {
		where = fmt.Sprintf("AND %s", strings.Join(wheres, " AND "))
	}

	// Create the complete query string.
	query := fmt.Sprintf("SELECT avg(\"%s\") FROM %s data INNER JOIN %s result ON data.\"%s\" = result.index WHERE result.result_id = $%d %s;",
		variable.Key, dataset, s.getResultTable(dataset), model.D3MIndexFieldName, len(params), where)

	// execute the postgres query
	res, err := s.client.Query(query, params...)
	if err != nil {
		return 0, errors.Wrap(err, "failed to fetch histograms for variable summaries from postgres")
	}
	if res != nil {
		defer res.Close()
	}

	return s.parseMean(res)
}

func (s *Storage) parseMean(row *pgx.Rows) (float64, error) {
	var mean *float64
	if row != nil {
		// Expect one row of data.
		row.Next()
		err := row.Scan(&mean)
		if err != nil {
			return 0, errors.Wrap(err, "no mean found")
		}
	}
	// check values exist
	if mean == nil {
		return 0, errors.Errorf("no mean found")
	}
	return *mean, nil
}
