package postgres

import (
	"fmt"
	"strings"

	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil/api/model"
)

const (
	filterLimit = 100
)

func (s *Storage) parseFilteredData(dataset string, rows *pgx.Rows) (*model.FilteredData, error) {
	result := &model.FilteredData{
		Name:   dataset,
		Values: make([][]interface{}, 0),
	}

	// Parse the columns.
	if rows != nil {
		fields := rows.FieldDescriptions()
		columns := make([]string, len(fields))
		types := make([]string, len(fields))
		for i := 0; i < len(fields); i++ {
			columns[i] = fields[i].Name
			types[i] = fields[i].DataTypeName
		}
		result.Columns = columns
		result.Types = types

		// Parse the row data.
		for rows.Next() {
			columnValues, err := rows.Values()
			if err != nil {
				return nil, err
			}
			result.Values = append(result.Values, columnValues)
		}
	} else {
		result.Columns = make([]string, 0)
		result.Types = make([]string, 0)
	}

	return result, nil
}

func (s *Storage) buildFilteredQueryWhere(dataset string, filterParams *model.FilterParams) (string, []interface{}, error) {
	// Build where clauses using the filter parameters.
	// param identifiers in the query are 1-based $x.
	params := make([]interface{}, 0)
	wheres := make([]string, 0)

	for _, filter := range filterParams.Filters {
		switch filter.Type {
		case model.NumericalFilter:
			// numerical
			where := fmt.Sprintf("\"%s\" >= $%d AND \"%s\" <= $%d", filter.Name, len(params)+1, filter.Name, len(params)+2)
			wheres = append(wheres, where)
			params = append(params, *filter.Min)
			params = append(params, *filter.Max)
		case model.CategoricalFilter:
			// categorical
			categories := make([]string, 0)
			offset := len(params) + 1
			for i, category := range filter.Categories {
				categories = append(categories, fmt.Sprintf("$%d", offset+i))
				params = append(params, category)
			}
			where := fmt.Sprintf("\"%s\" IN (%s)", filter.Name, strings.Join(categories, ", "))
			wheres = append(wheres, where)
		}
	}

	return strings.Join(wheres, " AND "), params, nil
}

func (s *Storage) buildFilteredQueryField(dataset string, variables []*model.Variable, filterParams *model.FilterParams, inclusive bool) (string, error) {
	fields := make([]string, 0)
	for _, variable := range model.GetFilterVariables(filterParams, variables, inclusive) {
		fields = append(fields, fmt.Sprintf("\"%s\"", variable.Name))
	}
	return strings.Join(fields, ","), nil
}

// FetchData creates a postgres query to fetch a set of rows.  Applies filters to restrict the
// results to a user selected set of fields, with rows further filtered based on allowed ranges and
// categories.
func (s *Storage) FetchData(dataset string, index string, filterParams *model.FilterParams, inclusive bool) (*model.FilteredData, error) {
	variables, err := s.metadata.FetchVariables(dataset, index, false)
	if err != nil {
		return nil, errors.Wrap(err, "Could not pull variables from ES")
	}

	fields, err := s.buildFilteredQueryField(dataset, variables, filterParams, inclusive)
	if err != nil {
		return nil, errors.Wrap(err, "Could not build field list")
	}

	// construct a Postgres query that fetches documents from the dataset with the supplied variable filters applied
	query := fmt.Sprintf("SELECT %s FROM %s", fields, dataset)

	where, params, err := s.buildFilteredQueryWhere(dataset, filterParams)
	if err != nil {
		return nil, errors.Wrap(err, "Could not build where clause")
	}

	if len(where) > 0 {
		query = fmt.Sprintf("%s WHERE %s", query, where)
	}

	// order & limit the filtered data.
	query = fmt.Sprintf("%s ORDER BY \"%s\"", query, d3mIndexFieldName)
	if filterParams.Size > 0 {
		query = fmt.Sprintf("%s LIMIT %d", query, filterParams.Size)
	}
	query = query + ";"

	// execute the postgres query
	res, err := s.client.Query(query, params...)
	if err != nil {
		return nil, errors.Wrap(err, "postgres filtered data query failed")
	}
	if res != nil {
		defer res.Close()
	}

	// parse the result
	return s.parseFilteredData(dataset, res)
}
