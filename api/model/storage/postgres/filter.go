package postgres

import (
	"fmt"
	"strings"

	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil/api/model"
)

func (s *Storage) parseResults(dataset string, rows *pgx.Rows) (*model.FilteredData, error) {
	result := &model.FilteredData{
		Name:   dataset,
		Values: make([][]interface{}, 0),
	}

	// Parse the metadata.
	if rows != nil {
		fields := rows.FieldDescriptions()
		metadata := make([]*model.Variable, len(fields))
		for i := 0; i < len(fields); i++ {
			metadata[i] = &model.Variable{
				Name: fields[i].Name,
				Type: fields[i].DataTypeName,
			}
		}
		result.Metadata = metadata

		// Parse the row data.
		for rows.Next() {
			columnValues, err := rows.Values()
			if err != nil {
				return nil, err
			}
			result.Values = append(result.Values, columnValues)
		}
	} else {
		result.Metadata = make([]*model.Variable, 0)
	}

	return result, nil
}

// FetchData creates a postgres query to fetch a set of rows.  Applies filters to restrict the
// results to a user selected set of fields, with rows further filtered based on allowed ranges and
// categories.
func (s *Storage) FetchData(dataset string, filterParams *model.FilterParams) (*model.FilteredData, error) {
	// construct a Postgres query that fetches documents from the dataset with the supplied variable filters applied
	query := fmt.Sprintf("SELECT * FROM %s", dataset)

	// Build where clauses using the filter parameters.
	// param identifiers in the query are 1-based $x.
	params := make([]interface{}, 0)
	wheres := make([]string, len(filterParams.Ranged))
	for i, variable := range filterParams.Ranged {
		wheres[i] = fmt.Sprintf("\"%s\" >= $%d AND \"%s\" <= $%d", variable.Name, i*2+1, variable.Name, i*2+2)
		params = append(params, variable.Min)
		params = append(params, variable.Max)
	}

	for _, variable := range filterParams.Categorical {
		// this is imposed by go's language design - []string needs explicit conversion to []interface{} before
		// passing to interface{} ...
		categories := make([]string, len(variable.Categories))
		baseParam := len(params) + 1
		for i := range variable.Categories {
			categories[i] = fmt.Sprintf("$%d", baseParam+i)
			params = append(params, variable.Categories[i])
		}
		wheres = append(wheres, fmt.Sprintf("\"%s\" IN (%s)", variable.Name, strings.Join(categories, ", ")))
	}
	//for _, variableName := range filterParams.None {
	//    excludes = append(excludes, variableName)
	//}

	if len(wheres) > 0 {
		query = fmt.Sprintf("%s WHERE %s;", query, strings.Join(wheres, " AND "))
	}

	// execute the postgres query
	res, err := s.client.Query(query, params...)
	if err != nil {
		return nil, errors.Wrap(err, "postgres filtered data query failed")
	}

	// parse the result
	return s.parseResults(dataset, res)
}
