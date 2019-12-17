//
//   Copyright © 2019 Uncharted Software Inc.
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package postgres

import (
	"fmt"
	"strings"

	"github.com/go-pg/pg"
	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/model"
)

const (
	maxBatchSize = 250
)

type joinDefinition struct {
	baseAlias     string
	baseColumn    string
	joinTableName string
	joinAlias     string
	joinColumn    string
}

func (s *Storage) getViewField(name string, displayName string, typ string, defaultValue interface{}) string {
	viewField := fmt.Sprintf("COALESCE(CAST(\"%s\" AS %s), %v)", name, typ, defaultValue)
	if model.IsDatabaseFloatingPoint(typ) {
		// handle missing values
		viewField = fmt.Sprintf("CASE WHEN \"%s\" = '' THEN %v ELSE %s END", name, defaultValue, viewField)
	}
	viewField = fmt.Sprintf("%s AS \"%s\"", viewField, displayName)
	return viewField
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
	vars, err := s.metadata.FetchVariables(dataset, false, false, true)
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

func (s *Storage) createView(storageName string, fields map[string]*model.Variable, overwrite bool) error {
	// CREATE OR REPLACE VIEW requires the same column names and order (with additions at the end being allowed).
	sql := "CREATE VIEW %s_tmp AS SELECT %s FROM %s_base;"

	// Build the select statement of the query.
	fieldList := make([]string, 0)
	for _, v := range fields {
		fieldList = append(fieldList, s.getViewField(v.Name, v.OriginalVariable, model.MapD3MTypeToPostgresType(v.Type), model.DefaultPostgresValueFromD3MType(v.Type)))
	}
	sql = fmt.Sprintf(sql, storageName, strings.Join(fieldList, ","), storageName)

	// Create the temporary view with the new column type.
	_, err := s.client.Exec(sql)
	if err != nil {
		return errors.Wrap(err, "Unable to create new temp view")
	}

	if overwrite {
		// Drop the existing view.
		_, err = s.client.Exec(fmt.Sprintf("DROP VIEW %s;", storageName))
		if err != nil {
			return errors.Wrap(err, "Unable to drop existing view")
		}

		// Rename the temporary view to the actual view name.
		_, err = s.client.Exec(fmt.Sprintf("ALTER VIEW %s_tmp RENAME TO %s;", storageName, storageName))
	}

	return err
}

// IsValidDataType checks to see if a specified type is valid for a variable.
// Multiple simultaneous calls to the function can result in inaccurate.
func (s *Storage) IsValidDataType(dataset string, storageName string, varName string, varType string) (bool, error) {
	// get all existing fields to rebuild the view.
	fields, err := s.getExistingFields(dataset)
	if err != nil {
		return false, errors.Wrap(err, "Unable to read existing fields")
	}

	// update field type in lookup.
	if fields[varName] == nil {
		return false, fmt.Errorf("field '%s' not found in existing fields", varName)
	}
	fields[varName].Type = varType

	// create view based on field lookup.
	err = s.createView(storageName, fields, false)
	if err != nil {
		return false, errors.Wrap(err, "Unable to create the new view")
	}

	// check if the new field type works with the data
	// a count on the field with the updated type should error if invalid
	verificationSQL := fmt.Sprintf("SELECT COUNT(\"%s\") FROM %s_tmp;", varName, storageName)
	_, err = s.client.Exec(verificationSQL)
	s.client.Exec(fmt.Sprintf("DROP VIEW %s_tmp;", storageName))
	if err != nil {
		return false, nil
	}

	return true, nil
}

// SetDataType updates the data type of the specified variable.
// Multiple simultaneous calls to the function can result in discarded changes.
func (s *Storage) SetDataType(dataset string, storageName string, varName string, varType string) error {
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

	// create view based on field lookup.
	err = s.createView(storageName, fields, true)
	if err != nil {
		return errors.Wrap(err, "Unable to create the new view")
	}

	return nil
}

func (s *Storage) createViewFromMetadataFields(storageName string, fields map[string]*model.Variable) error {
	dbFields := make(map[string]*model.Variable)

	// map the types to db types.
	for field, v := range fields {
		dbFields[field] = &model.Variable{
			Name:             v.Name,
			OriginalVariable: v.OriginalVariable,
			Type:             v.Type,
		}
	}

	err := s.createView(storageName, dbFields, true)
	if err != nil {
		return errors.Wrap(err, "Unable to create the new view")
	}

	return nil
}

// AddVariable adds a new variable to the dataset.
func (s *Storage) AddVariable(dataset string, storageName string, varName string, varType string) error {
	// check to make sure the column doesnt exist already
	dbFields, err := s.getDatabaseFields(fmt.Sprintf("%s_base", storageName))
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
		// add the empty column to the base table and the explain table
		sql := fmt.Sprintf("ALTER TABLE %s_base ADD COLUMN \"%s\" TEXT;", storageName, varName)
		_, err = s.client.Exec(sql)
		if err != nil {
			return errors.Wrap(err, "unable to add new column to database table")
		}

		sql = fmt.Sprintf("ALTER TABLE %s_explain ADD COLUMN \"%s\" DOUBLE PRECISION;", storageName, varName)
		_, err = s.client.Exec(sql)
		if err != nil {
			return errors.Wrap(err, "unable to add new column to database explain table")
		}
	}

	// recreate the view with the new column
	fields, err := s.getExistingFields(dataset)
	if err != nil {
		return errors.Wrap(err, "unable to read existing fields")
	}

	// need to add the field to the view
	fields[varName] = &model.Variable{
		Name:             varName,
		OriginalVariable: varName,
		Type:             varType,
	}

	err = s.createViewFromMetadataFields(storageName, fields)
	if err != nil {
		return errors.Wrap(err, "unable to create the new view")
	}

	return nil
}

// DeleteVariable flags a variable as deleted.
func (s *Storage) DeleteVariable(dataset string, storageName string, varName string) error {
	// check if the variable is in the view
	dbFields, err := s.getDatabaseFields(storageName)
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

	err = s.createViewFromMetadataFields(storageName, fields)
	if err != nil {
		return errors.Wrap(err, "Unable to create the new view")
	}

	return nil
}

// InsertBatch batches the data to insert for increased performance.
func (s *Storage) InsertBatch(storageName string, varNames []string, inserts [][]interface{}) error {
	db := s.client.GetBatchClient()
	defer db.Close()

	err := s.insertBatchData(db, storageName, varNames, inserts)
	if err != nil {
		return errors.Wrap(err, "unable to insert batches")
	}

	return nil
}

func (s *Storage) insertBatchData(db *pg.DB, storageName string, varNames []string, inserts [][]interface{}) error {
	// get the boiler plater of the query
	fieldCount := len(varNames)
	basicInsert := "INSERT INTO \"%s\" (%s) VALUES (%s);"
	paramList := strings.Repeat(", ?", fieldCount)[2:]

	// need to quote the fields
	// after joining, the first and last fields are missing a quote
	fieldList := strings.Join(varNames, "\", \"")
	fieldList = fmt.Sprintf("\"%s\"", fieldList)

	basicInsert = fmt.Sprintf(basicInsert, storageName, fieldList, paramList)

	// build the batches and run the queries
	params := make([]interface{}, 0)
	insertSQL := ""
	count := 0
	for i := 0; i < len(inserts); i++ {
		insertSQL = fmt.Sprintf("%s %s", insertSQL, basicInsert)
		for j := 0; j < fieldCount; j++ {
			params = append(params, inserts[i][j])
		}

		count = count + 1
		if count > maxBatchSize {
			// submit the batch
			_, err := db.Exec(insertSQL, params...)
			if err != nil {
				return errors.Wrap(err, "unable to insert batch")
			}

			// reset the batch
			insertSQL = ""
			count = 0
			params = make([]interface{}, 0)
		}
	}

	// submit remaining rows
	if count > 0 {
		_, err := db.Exec(insertSQL, params...)
		if err != nil {
			return errors.Wrap(err, "unable to insert batch")
		}
	}

	return nil
}

// UpdateVariable updates the value of a variable stored in the database.
func (s *Storage) UpdateVariable(storageName string, varName string, d3mIndex string, value string) error {
	sql := fmt.Sprintf("UPDATE %s_base SET \"%s\" = $1 WHERE \"%s\" = $2", storageName, varName, model.D3MIndexFieldName)
	_, err := s.client.Exec(sql, value, d3mIndex)
	if err != nil {
		return errors.Wrap(err, "Unable to update value stored in the database")
	}

	return nil
}

// UpdateVariableBatch batches updates for a variable to increase performance.
func (s *Storage) UpdateVariableBatch(storageName string, varName string, updates map[string]string) error {
	// A couple of approaches are possible:
	// 1. Batch the updates in a string and send many updates at once to diminish network time.
	// 2. Batch insert the updates to a temp table, send an update command where a join
	//		between the original table and the temp table is done to get the new values
	//		and then delete the temp table.

	// build params
	params := make([][]interface{}, 0)
	for i, v := range updates {
		params = append(params, []interface{}{i, v})
	}

	// loop through the updates, building batches to minimize overhead
	db := s.client.GetBatchClient()
	defer db.Close()
	tableNameTmp := fmt.Sprintf("%s_utmp", storageName)
	dataSQL := fmt.Sprintf("CREATE TEMP TABLE \"%s\" (\"%s\" TEXT NOT NULL, \"%s\" TEXT);",
		tableNameTmp, model.D3MIndexName, varName)
	_, err := db.Exec(dataSQL)
	if err != nil {
		return errors.Wrap(err, "unable to create temp table")
	}

	err = s.insertBatchData(db, tableNameTmp, []string{model.D3MIndexName, varName}, params)
	if err != nil {
		return errors.Wrap(err, "unable to insert into temp table")
	}

	// run the update
	updateSQL := fmt.Sprintf("UPDATE %s.%s.\"%s_base\" AS b SET \"%s\" = t.\"%s\" FROM \"%s\" AS t WHERE t.\"%s\" = b.\"%s\";",
		"distil", "public", storageName, varName, varName, tableNameTmp, model.D3MIndexName, model.D3MIndexName)
	_, err = db.Exec(updateSQL)
	if err != nil {
		return errors.Wrap(err, "unable to update base data")
	}

	return nil
}

func getJoinSQL(join *joinDefinition, inner bool) string {
	joinType := "INNER JOIN"
	if !inner {
		joinType = "LEFT OUTER JOIN"
	}
	return fmt.Sprintf("%s %s AS %s ON %s.\"%s\" = %s.\"%s\"",
		joinType, join.joinTableName, join.joinAlias, join.joinAlias,
		join.joinColumn, join.baseAlias, join.baseColumn)
}
