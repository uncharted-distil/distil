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
	case model.IndexType:
		return dataTypeInteger
	case model.IntegerType, model.FloatType, model.LongitudeType, model.LatitudeType, model.RealType:
		return dataTypeFloat
	case model.OrdinalType, model.CategoricalType, model.TextType, model.DateTimeType, model.StringType:
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
	case dataTypeInteger:
		return int(0)
	default:
		return "''"
	}
}

func (s *Storage) getDatabaseFields(tableName string) ([]string, error) {
	sql := fmt.Sprintf("SELECT column_name FROM information_schema.columns WHERE table_schema = 'public' AND table_name = $1;")

	res, err := s.client.Query(sql, tableName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch database column names from postgres")
	}
	if res != nil {
		defer res.Close()
	}

	cols := make([]string, 0)
	for res.Next() {
		var colName string
		err := res.Scan(&colName)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to parse column name")
		}
		cols = append(cols, colName)
	}

	return cols, nil
}

func (s *Storage) getExistingFields(dataset string) (map[string]*model.Variable, error) {
	// Read the existing fields from the database.
	vars, err := s.metadata.FetchVariablesDisplay(dataset)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to get existing fields")
	}

	// Add the d3m index variable.
	varIndex, err := s.metadata.FetchVariable(dataset, model.D3MIndexFieldName)
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
		fieldList = append(fieldList, s.getViewField(v.Key, v.OriginalVariable, v.Type, s.defaultValue(v.Type)))
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

// SetDataType updates the data type of the specified variable.
// Multiple simultaneous calls to the function can result in discarded changes.
func (s *Storage) SetDataType(dataset string, varName string, varType string) error {
	// get all existing fields to rebuild the view.
	fields, err := s.getExistingFields(dataset)
	if err != nil {
		return errors.Wrap(err, "Unable to read existing fields")
	}

	// update field type in lookup.
	if fields[varName] == nil {
		return fmt.Errorf("field '%s' not found in existing fields", varName)
	}
	fields[varName].Type = varType

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

func (s *Storage) createViewFromMetadataFields(dataset string, fields map[string]*model.Variable) error {
	dbFields := make(map[string]*model.Variable)

	// map the types to db types.
	for field, v := range fields {
		dbFields[field] = &model.Variable{
			Key:              v.Key,
			OriginalVariable: v.OriginalVariable,
			Type:             s.mapType(v.Type),
		}
	}

	err := s.createView(dataset, dbFields)
	if err != nil {
		return errors.Wrap(err, "Unable to create the new view")
	}

	return nil
}

// AddVariable adds a new variable to the dataset.
func (s *Storage) AddVariable(dataset string, varName string, varType string) error {
	// check to make sure the column doesnt exist already
	dbFields, err := s.getDatabaseFields(fmt.Sprintf("%s_base", dataset))
	if err != nil {
		return errors.Wrap(err, "unable to read database fields")
	}

	found := false
	for _, v := range dbFields {
		if v == varName {
			found = true
			break
		}
	}
	if found {
		return errors.Errorf("dataset %s already has variable '%s' in postgres", dataset, varName)
	}

	// add the empty column
	sql := fmt.Sprintf("ALTER TABLE %s_base ADD COLUMN \"%s\" TEXT;", dataset, varName)
	_, err = s.client.Exec(sql)
	if err != nil {
		return errors.Wrap(err, "Unable to add new column to database table")
	}

	// recreate the view with the new column
	fields, err := s.getExistingFields(dataset)
	if err != nil {
		return errors.Wrap(err, "Unable to read existing fields")
	}

	if fields[varName] == nil {
		// need to add the field to the view
		fields[varName] = &model.Variable{
			Key:              varName,
			OriginalVariable: varName,
			Type:             varType,
		}
	}

	err = s.createViewFromMetadataFields(dataset, fields)
	if err != nil {
		return errors.Wrap(err, "Unable to create the new view")
	}

	return nil
}

// DeleteVariable flags a variable as deleted.
func (s *Storage) DeleteVariable(dataset string, varName string) error {
	// check if the variable is in the view
	dbFields, err := s.getDatabaseFields(dataset)
	if err != nil {
		return errors.Wrap(err, "unable to read database fields")
	}

	found := false
	for _, v := range dbFields {
		if v == varName {
			found = true
			break
		}
	}
	if !found {
		return nil
	}

	// recreate the view without the field if it is in it
	fields, err := s.getExistingFields(dataset)
	if err != nil {
		return errors.Wrap(err, "Unable to read existing fields")
	}

	if fields[varName] != nil {
		delete(fields, varName)
	}

	err = s.createViewFromMetadataFields(dataset, fields)
	if err != nil {
		return errors.Wrap(err, "Unable to create the new view")
	}

	return nil
}

// UpdateVariable updates the value of a variable stored in the database.
func (s *Storage) UpdateVariable(dataset string, varName string, d3mIndex string, value string) error {
	sql := fmt.Sprintf("UPDATE %s_base SET \"%s\" = $1 WHERE \"%s\" = $2", dataset, varName, model.D3MIndexFieldName)
	_, err := s.client.Exec(sql, value, d3mIndex)
	if err != nil {
		return errors.Wrap(err, "Unable to update value stored in the database")
	}

	return nil
}
