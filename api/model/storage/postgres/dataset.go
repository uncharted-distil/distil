package postgres

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil/api/model"
)

func (s *Storage) getViewField(name string, displayName string, typ string, defaultValue interface{}) string {
	return fmt.Sprintf("COALESCE(CAST(\"%s\" AS %s), %v) AS \"%s\"",
		name, typ, defaultValue, displayName)
}

func (s *Storage) mapType(typ string) string {
	// Integer types can be returned as floats.
	switch typ {
	case model.IntegerType:
		fallthrough
	case model.IntType:
		fallthrough
	case model.FloatType:
		return dataTypeFloat
	case model.CategoricalType:
		fallthrough
	case model.TextType:
		fallthrough
	case model.DateTimeType:
		fallthrough
	case model.OrdinalType:
		return dataTypeText
	default:
		return dataTypeText
	}
}

func (s *Storage) defaultValue(typ string) interface{} {
	switch typ {
	case dataTypeDouble:
		return float64(0)
	case dataTypeFloat:
		return float64(0)
	default:
		return "''"
	}
}

func (s *Storage) getExistingFields(dataset string, index string) (map[string]*model.Variable, error) {
	// Read the existing fields from the database.
	vars, err := s.metadata.FetchVariablesDisplay(dataset, index)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to get existing fields")
	}

	// Add the d3m index variable.
	varIndex, err := s.metadata.FetchVariable(dataset, index, d3mIndexFieldName)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to get d3m index variable")
	}
	vars = append(vars, varIndex)

	fields := make(map[string]*model.Variable)
	for _, v := range vars {
		fields[v.OriginalVariable] = v
	}

	return fields, nil
}

func (s *Storage) createView(dataset string, fields map[string]*model.Variable) error {
	// CREATE OR REPLACE VIEW requires the same column names and order (with additions at the end being allowed).
	sql := "CREATE VIEW %s_tmp AS SELECT %s FROM %s_base;"

	// Build the select statement of the query.
	fieldList := make([]string, 0)
	for _, v := range fields {
		fieldList = append(fieldList, s.getViewField(v.Name, v.OriginalVariable, v.Type, s.defaultValue(v.Type)))
	}
	sql = fmt.Sprintf(sql, dataset, strings.Join(fieldList, ","), dataset)

	// Create the temporary view with the new column type.
	_, err := s.client.Exec(sql)
	if err != nil {
		return errors.Wrap(err, "Unable to create new temp view")
	}

	// Drop the existing view.
	_, err = s.client.Exec(fmt.Sprintf("DROP VIEW %s;", dataset))
	if err != nil {
		return errors.Wrap(err, "Unable to drop existing view")
	}

	// Rename the temporary view to the actual view name.
	_, err = s.client.Exec(fmt.Sprintf("ALTER VIEW %s_tmp RENAME TO %s;", dataset, dataset))

	return err
}

// SetDataType updates the data type of the specified field.
// Multiple simultaneous calls to the function can result in discarded changes.
func (s *Storage) SetDataType(dataset string, index string, field string, fieldType string) error {
	// get all existing fields to rebuild the view.
	fields, err := s.getExistingFields(dataset, index)
	if err != nil {
		return errors.Wrap(err, "Unable to read existing fields")
	}

	// update field type in lookup.
	if fields[field] == nil {
		return fmt.Errorf("field '%s' not found in existing fields", field)
	}
	fields[field].Type = fieldType

	// map the types to db types.
	for field, v := range fields {
		fields[field].Type = s.mapType(v.Type)
	}

	// create view based on field lookup.
	err = s.createView(dataset, fields)
	if err != nil {
		return errors.Wrap(err, "Unable to create the new view")
	}

	return nil
}
