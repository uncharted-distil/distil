package postgres

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil/api/model"
)

func (s *Storage) getViewField(name string, typ string, defaultValue interface{}) string {
	return fmt.Sprintf("COALESCE(CAST(\"%s\" AS %s), %v) AS \"%s\"",
		name, typ, defaultValue, name)
}

func (s *Storage) mapType(typ string) string {
	// Integer types can be returned as floats.
	switch typ {
	case model.IntegerType:
		fallthrough
	case model.FloatType:
		return "FLOAT8"
	case model.CategoricalType:
		fallthrough
	case model.TextType:
		fallthrough
	case model.DateTimeType:
		fallthrough
	case model.OrdinalType:
		return "TEXT"
	default:
		return "TEXT"
	}
}

func (s *Storage) defaultValue(typ string) interface{} {
	switch typ {
	case "double precision":
		return float64(0)
	case "FLOAT8":
		return float64(0)
	default:
		return "''"
	}
}

func (s *Storage) getExistingFields(dataset string) (map[string]string, error) {
	// Read the existing fields from the database.
	sql := "SELECT column_name, data_type FROM information_schema.columns WHERE table_name = $1;"

	rows, err := s.client.Query(sql, dataset)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to get existing fields")
	}
	defer rows.Close()

	fields := make(map[string]string)
	for rows.Next() {
		var columnName string
		var dataType string
		err = rows.Scan(&columnName, &dataType)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to parse existing fields")
		}

		fields[columnName] = dataType
	}

	return fields, nil
}

func (s *Storage) createView(dataset string, fields map[string]string) error {
	// CREATE OR REPLACE VIEW requires the same column names and order (with additions at the end being allowed).
	sql := "CREATE VIEW %s AS SELECT %s FROM %s_base;"

	// Build the select statement of the query.
	fieldList := make([]string, 0)
	for field, typ := range fields {
		fieldList = append(fieldList, s.getViewField(field, typ, s.defaultValue(typ)))
	}

	// Drop the existing view.
	_, err := s.client.Exec(fmt.Sprintf("DROP VIEW %s;", dataset))
	if err != nil {
		return errors.Wrap(err, "Unable to drop existing view")
	}

	// Create the view with the new column type.
	sql = fmt.Sprintf(sql, dataset, strings.Join(fieldList, ","), dataset)
	_, err = s.client.Exec(sql)

	return err
}

// SetDataType updates the data type of the specified field.
// Multiple simultaneous calls to the function can result in discarded changes.
func (s *Storage) SetDataType(dataset string, index string, field string, fieldType string) error {
	// get all existing fields to rebuild the view.
	fields, err := s.getExistingFields(dataset)
	if err != nil {
		return errors.Wrap(err, "Unable to read existing fields")
	}

	// create field lookup to map fields to types.
	dbType := s.mapType(fieldType)

	// update field type in lookup.
	fields[field] = dbType

	// create view based on field lookup.
	err = s.createView(dataset, fields)
	if err != nil {
		return errors.Wrap(err, "Unable to create the new view")
	}

	return nil
}
