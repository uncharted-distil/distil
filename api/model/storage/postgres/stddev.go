package postgres

import (
	"fmt"
	"strings"

	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil/api/model"
)

// FetchStdDev returns the stddev for a given dataset and variable.
func (s *Storage) FetchStdsDev(dataset string, variable *model.Variable, filterParams *model.FilterParams) (float64, error) {
	// create the filter for the query.
	wheres := make([]string, 0)
	params := make([]interface{}, 0)
	wheres, params = s.buildFilteredQueryWhere(wheres, params, dataset, filterParams.Filters)

	where := ""
	if len(wheres) > 0 {
		where = fmt.Sprintf("WHERE %s", strings.Join(wheres, " AND "))
	}

	// Create the complete query string.
	query := fmt.Sprintf("SELECT stddev(\"%s\") FROM %s %s;", variable.Key, dataset, where)

	// execute the postgres query
	res, err := s.client.Query(query, params...)
	if err != nil {
		return 0, errors.Wrap(err, "failed to fetch histograms for variable summaries from postgres")
	}
	if res != nil {
		defer res.Close()
	}

	return s.parseStdDev(res)
}

// FetchStdDevByResult returns the stddev for a given dataset, variable, and result.
func (s *Storage) FetchStdDevByRessult(dataset string, variable *model.Variable, resultURI string, filterParams *model.FilterParams) (float64, error) {
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
	query := fmt.Sprintf("SELECT stddev(\"%s\") FROM %s data INNER JOIN %s result ON data.\"%s\" = result.index WHERE result.result_id = $%d %s;",
		variable.Key, dataset, s.getResultTable(dataset), model.D3MIndexFieldName, len(params), where)

	// execute the postgres query
	res, err := s.client.Query(query, params...)
	if err != nil {
		return 0, errors.Wrap(err, "failed to fetch histograms for variable summaries from postgres")
	}
	if res != nil {
		defer res.Close()
	}

	return s.parseStdDev(res)
}

func (s *Storage) parseStdDev(row *pgx.Rows) (float64, error) {
	var stddev *float64
	if row != nil {
		// Expect one row of data.
		row.Next()
		err := row.Scan(&stddev)
		if err != nil {
			return 0, errors.Wrap(err, "no std dev found")
		}
	}
	// check values exist
	if stddev == nil {
		return 0, errors.Errorf("no std dev found")
	}
	return *stddev, nil
}
