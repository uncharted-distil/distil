package postgres

import (
	"fmt"
	"strings"

	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil/api/model"
)

const (
	// CorrectCategory identifies the correct result meta-category.
	CorrectCategory = "correct"

	// IncorrectCategory identifies the incorrect result meta-category.
	IncorrectCategory = "incorrect"
)

func (s *Storage) parseFilteredData(dataset string, numRows int, rows *pgx.Rows) (*model.FilteredData, error) {
	result := &model.FilteredData{
		NumRows: numRows,
		Values:  make([][]interface{}, 0),
	}

	// Parse the columns.
	if rows != nil {
		fields := rows.FieldDescriptions()
		columns := make([]model.Column, len(fields))
		for i := 0; i < len(fields); i++ {
			columns[i] = model.Column{
				Key:   fields[i].Name,
				Label: fields[i].Name,
				Type:  fields[i].DataTypeName,
			}
		}
		result.Columns = columns

		// Parse the row data.
		for rows.Next() {
			columnValues, err := rows.Values()
			if err != nil {
				return nil, err
			}
			result.Values = append(result.Values, columnValues)
		}
	} else {
		result.Columns = make([]model.Column, 0)
	}

	return result, nil
}

func (s *Storage) formatFilterKey(key string) string {
	if model.IsResultKey(key) {
		return "result.value"
	}
	return fmt.Sprintf("\"%s\"", key)
}

func (s *Storage) buildIncludeFilter(wheres []string, params []interface{}, filter *model.Filter) ([]string, []interface{}) {

	name := s.formatFilterKey(filter.Key)

	switch filter.Type {
	case model.NumericalFilter:
		// numerical
		// cast to double precision in case of string based representation
		where := fmt.Sprintf("cast(%s as double precision) >= $%d AND cast(%s as double precision) <= $%d", name, len(params)+1, name, len(params)+2)
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
	case model.RowFilter:
		// row
		indices := make([]string, 0)
		offset := len(params) + 1
		for i, d3mIndex := range filter.D3mIndices {
			indices = append(indices, fmt.Sprintf("$%d", offset+i))
			params = append(params, d3mIndex)
		}
		where := fmt.Sprintf("\"%s\" IN (%s)", model.D3MIndexFieldName, strings.Join(indices, ", "))
		wheres = append(wheres, where)
	case model.FeatureFilter:
		// feature
		offset := len(params) + 1
		for i, category := range filter.Categories {
			where := fmt.Sprintf("%s ~ (%s)", name, fmt.Sprintf("$%d", offset+i))
			params = append(params, category)
			wheres = append(wheres, where)
		}
	}
	return wheres, params
}

func (s *Storage) buildExcludeFilter(wheres []string, params []interface{}, filter *model.Filter) ([]string, []interface{}) {

	name := s.formatFilterKey(filter.Key)

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
	case model.RowFilter:
		// row
		indices := make([]string, 0)
		offset := len(params) + 1
		for i, d3mIndex := range filter.D3mIndices {
			indices = append(indices, fmt.Sprintf("$%d", offset+i))
			params = append(params, d3mIndex)
		}
		where := fmt.Sprintf("\"%s\" NOT IN (%s)", model.D3MIndexFieldName, strings.Join(indices, ", "))
		wheres = append(wheres, where)
	}
	return wheres, params
}

func (s *Storage) buildFilteredQueryWhere(wheres []string, params []interface{}, dataset string, filters []*model.Filter) ([]string, []interface{}) {
	for _, filter := range filters {
		switch filter.Mode {
		case model.IncludeFilter:
			wheres, params = s.buildIncludeFilter(wheres, params, filter)
		case model.ExcludeFilter:
			wheres, params = s.buildExcludeFilter(wheres, params, filter)
		}
	}
	return wheres, params
}

func (s *Storage) buildFilteredQueryField(dataset string, variables []*model.Variable, filterVariables []string) (string, error) {
	fields := make([]string, 0)
	indexIncluded := false
	for _, variable := range model.GetFilterVariables(filterVariables, variables) {
		fields = append(fields, fmt.Sprintf("\"%s\"", variable.Key))
		if variable.Key == model.D3MIndexFieldName {
			indexIncluded = true
		}
	}
	// if the index is not already in the field list, then append it
	if !indexIncluded {
		fields = append(fields, fmt.Sprintf("\"%s\"", model.D3MIndexFieldName))
	}
	return strings.Join(fields, ","), nil
}

func (s *Storage) buildFilteredResultQueryField(dataset string, variables []*model.Variable, targetVariable *model.Variable, filterVariables []string) (string, error) {
	fields := make([]string, 0)
	for _, variable := range model.GetFilterVariables(filterVariables, variables) {
		if strings.Compare(targetVariable.Key, variable.Key) != 0 {
			fields = append(fields, fmt.Sprintf("\"%s\"", variable.Key))
		}
	}
	fields = append(fields, fmt.Sprintf("\"%s\"", model.D3MIndexFieldName))
	return strings.Join(fields, ","), nil
}

func (s *Storage) buildCorrectnessResultWhere(wheres []string, params []interface{}, dataset string, resultURI string, resultFilter *model.Filter) ([]string, []interface{}, error) {
	// get the target variable name
	datasetResult := s.getResultTable(dataset)
	targetName, err := s.getResultTargetName(datasetResult, resultURI)
	if err != nil {
		return nil, nil, err
	}

	// correct/incorrect are well known categories that require the predicted category to be compared
	// to the target category
	op := ""
	for _, category := range resultFilter.Categories {
		if strings.EqualFold(category, CorrectCategory) {
			op = "="
			break
		} else if strings.EqualFold(category, IncorrectCategory) {
			op = "!="
			break
		}
	}
	if op == "" {
		return nil, nil, err
	}
	where := fmt.Sprintf("result.value %s data.\"%s\"", op, targetName)
	wheres = append(wheres, where)
	return wheres, params, nil
}

func (s *Storage) buildErrorResultWhere(wheres []string, params []interface{}, residualFilter *model.Filter) ([]string, []interface{}, error) {
	// Add a clause to filter residuals to the existing where
	nameWithoutSuffix := model.StripKeySuffix(residualFilter.Key)
	typedError := getErrorTyped(nameWithoutSuffix)
	where := fmt.Sprintf("(%s >= $%d AND %s <= $%d)", typedError, len(params)+1, typedError, len(params)+2)
	params = append(params, *residualFilter.Min)
	params = append(params, *residualFilter.Max)

	// Append the AND clause
	wheres = append(wheres, where)
	return wheres, params, nil
}

func (s *Storage) buildPredictedResultWhere(wheres []string, params []interface{}, dataset string, resultURI string, resultFilter *model.Filter) ([]string, []interface{}, error) {
	// handle the general category case
	wheres, params = s.buildFilteredQueryWhere(wheres, params, dataset, []*model.Filter{resultFilter})
	return wheres, params, nil
}

type filters struct {
	genericFilters    []*model.Filter
	predictedFilter   *model.Filter
	residualFilter    *model.Filter
	correctnessFilter *model.Filter
}

func (s *Storage) splitFilters(filterParams *model.FilterParams) *filters {
	// Groups filters for handling downstream
	var predictedFilter *model.Filter
	var residualFilter *model.Filter
	var correctnessFilter *model.Filter
	var remaining []*model.Filter
	for _, filter := range filterParams.Filters {
		if model.IsPredictedKey(filter.Key) {
			predictedFilter = filter
		} else if model.IsErrorKey(filter.Key) {
			if filter.Type == model.NumericalFilter {
				residualFilter = filter
			} else if filter.Type == model.CategoricalFilter {
				correctnessFilter = filter
			}
		} else {
			remaining = append(remaining, filter)
		}
	}
	return &filters{
		genericFilters:    remaining,
		predictedFilter:   predictedFilter,
		residualFilter:    residualFilter,
		correctnessFilter: correctnessFilter,
	}
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
		if v.Key == model.D3MIndexFieldName {
			return true
		}
	}

	return false
}

// FetchData creates a postgres query to fetch a set of rows.  Applies filters to restrict the
// results to a user selected set of fields, with rows further filtered based on allowed ranges and
// categories.
func (s *Storage) FetchData(dataset string, filterParams *model.FilterParams, invert bool) (*model.FilteredData, error) {
	variables, err := s.metadata.FetchVariables(dataset, true, true)
	if err != nil {
		return nil, errors.Wrap(err, "Could not pull variables from ES")
	}

	numRows, err := s.FetchNumRows(dataset, nil)
	if err != nil {
		return nil, errors.Wrap(err, "Could not pull num rows")
	}

	fields, err := s.buildFilteredQueryField(dataset, variables, filterParams.Variables)
	if err != nil {
		return nil, errors.Wrap(err, "Could not build field list")
	}

	// construct a Postgres query that fetches documents from the dataset with the supplied variable filters applied
	query := fmt.Sprintf("SELECT %s FROM %s", fields, dataset)

	wheres := make([]string, 0)
	params := make([]interface{}, 0)
	wheres, params = s.buildFilteredQueryWhere(wheres, params, dataset, filterParams.Filters)

	if len(wheres) > 0 {
		if invert {
			query = fmt.Sprintf("%s WHERE NOT(%s)", query, strings.Join(wheres, " AND "))
		} else {
			query = fmt.Sprintf("%s WHERE %s", query, strings.Join(wheres, " AND "))
		}
	} else {
		// if there are not WHERE's and we are inverting, that means we expect
		// no results.
		if invert {
			return &model.FilteredData{
				NumRows: numRows,
				Columns: make([]model.Column, 0),
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
