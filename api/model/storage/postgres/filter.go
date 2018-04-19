package postgres

import (
	"fmt"
	"strings"

	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil/api/model"
)

func (s *Storage) parseFilteredData(dataset string, numRows int, rows *pgx.Rows) (*model.FilteredData, error) {
	result := &model.FilteredData{
		Name:    dataset,
		NumRows: numRows,
		Values:  make([][]interface{}, 0),
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

func (s *Storage) formatFilterName(name string) string {
	if strings.HasSuffix(name, predictedSuffix) {
		//name = "value"
		return "CAST(\"value\" as double precision)"
	}
	return fmt.Sprintf("\"%s\"", name)
}

func (s *Storage) buildIncludeFilter(wheres []string, params []interface{}, filter *model.Filter) ([]string, []interface{}) {

	name := s.formatFilterName(filter.Name)

	switch filter.Type {
	case model.NumericalFilter:
		// numerical
		where := fmt.Sprintf("%s >= $%d AND %s <= $%d", name, len(params)+1, name, len(params)+2)
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
		where := fmt.Sprintf("%s IN (%s)", name, strings.Join(categories, ", "))
		wheres = append(wheres, where)
	}
	return wheres, params
}

func (s *Storage) buildExcludeFilter(wheres []string, params []interface{}, filter *model.Filter) ([]string, []interface{}) {

	name := s.formatFilterName(filter.Name)

	switch filter.Type {
	case model.NumericalFilter:
		// numerical
		where := fmt.Sprintf("(%s < $%d OR %s > $%d)", name, len(params)+1, name, len(params)+2)
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
		where := fmt.Sprintf("%s NOT IN (%s)", name, strings.Join(categories, ", "))
		wheres = append(wheres, where)
	}
	return wheres, params
}

func (s *Storage) buildFilteredQueryWhere(dataset string, filterParams *model.FilterParams) (string, []interface{}) {

	wheres := make([]string, 0)
	params := make([]interface{}, 0)

	for _, filter := range filterParams.Filters {
		switch filter.Mode {
		case model.IncludeFilter:
			wheres, params = s.buildIncludeFilter(wheres, params, filter)
		case model.ExcludeFilter:
			wheres, params = s.buildExcludeFilter(wheres, params, filter)
		}
	}

	return strings.Join(wheres, " AND "), params
}

func (s *Storage) buildFilteredQueryField(dataset string, variables []*model.Variable, filterParams *model.FilterParams) (string, error) {
	fields := make([]string, 0)
	for _, variable := range model.GetFilterVariables(filterParams, variables) {
		fields = append(fields, fmt.Sprintf("\"%s\"", variable.Name))
	}
	return strings.Join(fields, ","), nil
}

func (s *Storage) buildFilteredResultQueryField(dataset string, variables []*model.Variable, targetVariable *model.Variable, filterParams *model.FilterParams) (string, error) {
	fields := make([]string, 0)
	for _, variable := range model.GetFilterVariables(filterParams, variables) {
		if strings.Compare(targetVariable.Name, variable.Name) != 0 {
			fields = append(fields, fmt.Sprintf("\"%s\"", variable.Name))
		}
	}
	return strings.Join(fields, ","), nil
}

// FetchNumRows pulls the number of rows in the table.
func (s *Storage) FetchNumRows(dataset string, filters map[string]interface{}) (int, error) {
	query := fmt.Sprintf("SELECT count(*) FROM %s", dataset)
	params := make([]interface{}, 0)
	if filters != nil && len(filters) > 0 {
		clauses := make([]string, 0)
		for field, value := range filters {
			clauses = append(clauses, fmt.Sprintf("%s = $%d", field, len(clauses)+1))
			params = append(params, value)
		}
		query = fmt.Sprintf("%s WHERE %s", query, strings.Join(clauses, " AND "))
	}
	var numRows int
	err := s.client.QueryRow(query, params...).Scan(&numRows)
	if err != nil {
		return -1, errors.Wrap(err, "postgres row query failed")
	}
	return numRows, nil
}

func (s *Storage) filterIncludesIndex(filterParams *model.FilterParams) bool {
	for _, v := range filterParams.Filters {
		if v.Name == model.D3MIndexFieldName {
			return true
		}
	}

	return false
}

// FetchData creates a postgres query to fetch a set of rows.  Applies filters to restrict the
// results to a user selected set of fields, with rows further filtered based on allowed ranges and
// categories.
func (s *Storage) FetchData(dataset string, index string, filterParams *model.FilterParams, invert bool) (*model.FilteredData, error) {
	variables, err := s.metadata.FetchVariables(dataset, index, true)
	if err != nil {
		return nil, errors.Wrap(err, "Could not pull variables from ES")
	}

	numRows, err := s.FetchNumRows(dataset, nil)
	if err != nil {
		return nil, errors.Wrap(err, "Could not pull num rows")
	}

	fields, err := s.buildFilteredQueryField(dataset, variables, filterParams)
	if err != nil {
		return nil, errors.Wrap(err, "Could not build field list")
	}

	// construct a Postgres query that fetches documents from the dataset with the supplied variable filters applied
	query := fmt.Sprintf("SELECT %s FROM %s", fields, dataset)

	where, params := s.buildFilteredQueryWhere(dataset, filterParams)

	if len(where) > 0 {
		if invert {
			query = fmt.Sprintf("%s WHERE NOT(%s)", query, where)
		} else {
			query = fmt.Sprintf("%s WHERE %s", query, where)
		}
	} else {
		// if there are not WHERE's and we are inverting, that means we expect
		// no results.
		if invert {
			return &model.FilteredData{
				Name:    dataset,
				NumRows: numRows,
				Columns: make([]string, 0),
				Types:   make([]string, 0),
				Values:  make([][]interface{}, 0),
			}, nil
		}
	}

	// order & limit the filtered data.
	query = fmt.Sprintf("%s ORDER BY \"%s\"", query, model.D3MIndexFieldName)
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
	return s.parseFilteredData(dataset, numRows, res)
}
