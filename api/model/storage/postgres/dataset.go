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
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/model"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/postgres"
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

func getBaseTableName(storageName string) string {
	return fmt.Sprintf("%s_base", storageName)
}

// CloneDataset clones an existing dataset.
func (s *Storage) CloneDataset(dataset string, storageName string, datasetNew string, storageNameNew string) error {
	// need to clone base, result, and weight tables
	err := s.cloneTable(getBaseTableName(storageName), getBaseTableName(storageNameNew), true)
	if err != nil {
		return err
	}

	err = s.cloneTable(s.getResultTable(storageName), s.getResultTable(storageNameNew), false)
	if err != nil {
		return err
	}

	err = s.cloneTable(s.getSolutionFeatureWeightTable(storageName), s.getSolutionFeatureWeightTable(storageNameNew), false)
	if err != nil {
		return err
	}

	// need to create the view for the cloned dataset
	fields, err := s.getExistingFields(dataset)
	if err != nil {
		return err
	}

	err = s.createView(storageNameNew, fields, true)
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) cloneTable(existingTable string, newTable string, copyData bool) error {
	sql := fmt.Sprintf("CREATE TABLE %s AS TABLE %s", newTable, existingTable)
	if !copyData {
		sql = fmt.Sprintf("%s WITH NO DATA", sql)
	}
	sql = fmt.Sprintf("%s;", sql)

	_, err := s.client.Exec(sql)
	if err != nil {
		return errors.Wrapf(err, "unable to clone table")
	}

	return nil
}

func (s *Storage) getViewField(fieldSelect string, displayName string, typ string, defaultValue interface{}) string {
	viewField := fmt.Sprintf("COALESCE(CAST(%s AS %s), %v)", fieldSelect, typ, defaultValue)
	if postgres.IsDatabaseFloatingPoint(typ) {
		// handle missing values
		viewField = fmt.Sprintf("CASE WHEN %s = '' THEN %v ELSE %s END", fieldSelect, defaultValue, viewField)
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
	vars, err := api.FetchDatasetVariables(dataset, s.metadata)
	if err != nil {
		return nil, err
	}

	fields := make(map[string]*model.Variable)
	for _, v := range vars {
		fields[v.Name] = v
	}

	return fields, nil
}

func (s *Storage) createView(storageName string, fields map[string]*model.Variable, overwrite bool) error {
	// CREATE OR REPLACE VIEW requires the same column names and order (with additions at the end being allowed).
	sql := "CREATE VIEW %s_tmp AS SELECT %s FROM %s_base;"

	// Build the select statement of the query.
	fieldList := make([]string, 0)
	for _, v := range fields {
		fieldList = append(fieldList, s.getViewField(postgres.ValueForFieldType(v.Type, v.Name),
			v.Name, postgres.MapD3MTypeToPostgresType(v.Type), postgres.DefaultPostgresValueFromD3MType(v.Type)))
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
		return false, errors.Errorf("field '%s' not found in existing fields", varName)
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
	_, _ = s.client.Exec(fmt.Sprintf("DROP VIEW %s_tmp;", storageName))
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
		return errors.Errorf("field '%s' not found in existing fields", varName)
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

// AddField adds a new field to the data storage. This only adds a new column.
// It does not add the column to other tables nor does it rebuild a view.
func (s *Storage) AddField(dataset string, storageName string, varName string, varType string) error {
	// check to make sure the column doesnt exist already
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

	if found {
		return nil
	}

	sql := fmt.Sprintf("ALTER TABLE %s ADD COLUMN \"%s\" %s;", storageName, varName, postgres.MapD3MTypeToPostgresType(varType))
	_, err = s.client.Exec(sql)
	if err != nil {
		return errors.Wrap(err, "unable to add new column to database explain table")
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

func (s *Storage) insertBulkCopy(storageName string, varNames []string, inserts [][]interface{}) error {
	rowsCopied, err := s.batchClient.CopyFrom(storageName, varNames, inserts)
	if err != nil {
		return errors.Wrapf(err, "unable to bulk copy data to postgres")
	}

	if rowsCopied != int64(len(inserts)) {
		return errors.Errorf("only bulk copied %d of %d rows to postgres", rowsCopied, len(inserts))
	}

	// update the stats to make sure postgres runs the best queries possible
	s.updateStats(storageName)

	return nil
}

func (s *Storage) insertBulkCopyTransaction(tx pgx.Tx, storageName string, varNames []string, inserts [][]interface{}) error {
	sourceValues := pgx.CopyFromRows(inserts)
	rowsCopied, err := tx.CopyFrom(context.Background(), pgx.Identifier{storageName}, varNames, sourceValues)
	if err != nil {
		return errors.Wrapf(err, "unable to bulk copy data to postgres")
	}

	if rowsCopied != int64(len(inserts)) {
		return errors.Errorf("only bulk copied %d of %d rows to postgres", rowsCopied, len(inserts))
	}

	return nil
}

// InsertBatch batches the data to insert for increased performance.
func (s *Storage) InsertBatch(storageName string, varNames []string, inserts [][]interface{}) error {

	err := s.insertBatchData(storageName, varNames, inserts)
	if err != nil {
		return errors.Wrap(err, "unable to insert batches")
	}

	return nil
}

func (s *Storage) insertBatchData(storageName string, varNames []string, inserts [][]interface{}) error {
	// get the boiler plate of the query
	fieldCount := len(varNames)
	paramList := ""
	for i := 0; i < fieldCount; i++ {
		paramList = fmt.Sprintf("%s, $%d", paramList, i+1)
	}
	paramList = paramList[2:]

	// need to quote the fields
	// after joining, the first and last fields are missing a quote
	fieldList := strings.Join(varNames, "\", \"")
	fieldList = fmt.Sprintf("\"%s\"", fieldList)

	batchSQL := fmt.Sprintf("INSERT INTO \"%s\" (%s) VALUES (%s);", storageName, fieldList, paramList)

	// build the batches and run the queries
	batch := &pgx.Batch{}
	for i := 0; i < len(inserts); i++ {
		params := make([]interface{}, 0)
		for j := 0; j < len(inserts[i]); j++ {
			params = append(params, inserts[i][j])
		}
		batch.Queue(batchSQL, params...)

		// append nil for remaining fields
		for j := len(inserts[i]); j < fieldCount; j++ {
			params = append(params, nil)
		}

		if batch.Len() > maxBatchSize {
			// submit the batch
			resBatch := s.batchClient.SendBatch(batch)
			for i := 0; i < maxBatchSize; i++ {
				_, err := resBatch.Exec()
				if err != nil {
					resBatch.Close()
					return errors.Wrapf(err, "unable to insert batch")
				}
			}
			resBatch.Close()

			// reset the batch
			batch = &pgx.Batch{}
		}
	}

	// submit remaining rows
	count := batch.Len()
	if count > 0 {
		resBatch := s.batchClient.SendBatch(batch)
		for i := 0; i < count; i++ {
			_, err := resBatch.Exec()
			if err != nil {
				resBatch.Close()
				return errors.Wrapf(err, "unable to insert final batch")
			}
		}
		resBatch.Close()
	}

	// update the stats to make sure postgres runs the best queries possible
	s.updateStats(storageName)

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
	tableNameTmp := fmt.Sprintf("%s_utmp", storageName)
	dataSQL := fmt.Sprintf("CREATE TEMP TABLE \"%s\" (\"%s\" TEXT NOT NULL, \"%s\" TEXT);",
		tableNameTmp, model.D3MIndexName, varName)
	_, err := s.batchClient.Exec(dataSQL)
	if err != nil {
		return errors.Wrap(err, "unable to create temp table")
	}

	err = s.insertBulkCopy(tableNameTmp, []string{model.D3MIndexName, varName}, params)
	if err != nil {
		return errors.Wrap(err, "unable to insert into temp table")
	}

	// run the update
	updateSQL := fmt.Sprintf("UPDATE %s.%s.\"%s_base\" AS b SET \"%s\" = t.\"%s\" FROM \"%s\" AS t WHERE t.\"%s\" = b.\"%s\";",
		"distil", "public", storageName, varName, varName, tableNameTmp, model.D3MIndexName, model.D3MIndexName)
	_, err = s.batchClient.Exec(updateSQL)
	if err != nil {
		return errors.Wrap(err, "unable to update base data")
	}
	s.batchClient.Exec(fmt.Sprintf("DROP TABLE \"%s\"", tableNameTmp))

	return nil
}

// UpdateData updates data stored using the d3m index as key, but also allows
// for filtering in cases where the d3m index is not unique.
func (s *Storage) UpdateData(dataset string, storageName string, varName string, updates map[string]string, filterParams *api.FilterParams) error {
	// build params
	params := make([][]interface{}, 0)
	for i, v := range updates {
		params = append(params, []interface{}{i, v})
	}

	// loop through the updates, building batches to minimize overhead
	tx, err := s.batchClient.Begin()
	if err != nil {
		tx.Rollback(context.Background())
		return errors.Wrap(err, "unable to create transaction")
	}

	tableNameTmp := fmt.Sprintf("%s_utmp", storageName)
	dataSQL := fmt.Sprintf("CREATE TEMP TABLE \"%s\" (\"%s\" TEXT NOT NULL, \"%s\" TEXT);",
		tableNameTmp, model.D3MIndexName, varName)
	_, err = tx.Exec(context.Background(), dataSQL)
	if err != nil {
		tx.Rollback(context.Background())
		return errors.Wrap(err, "unable to create temp table")
	}

	err = s.insertBulkCopyTransaction(tx, tableNameTmp, []string{model.D3MIndexName, varName}, params)
	if err != nil {
		tx.Rollback(context.Background())
		return errors.Wrap(err, "unable to insert into temp table")
	}

	// build the filter structure
	wheres := []string{fmt.Sprintf("t.\"%s\" = b.\"%s\"::text", model.D3MIndexName, model.D3MIndexName)}
	paramsFilter := make([]interface{}, 0)
	wheres, paramsFilter = s.buildFilteredQueryWhere(dataset, wheres, paramsFilter, "b", filterParams, false)

	// run the update
	updateSQL := fmt.Sprintf("UPDATE %s.%s.\"%s\" AS b SET \"%s\" = t.\"%s\" FROM \"%s\" AS t WHERE %s",
		"distil", "public", storageName, varName, varName, tableNameTmp, strings.Join(wheres, " AND "))
	_, err = tx.Exec(context.Background(), updateSQL, paramsFilter...)
	if err != nil {
		tx.Rollback(context.Background())
		return errors.Wrap(err, "unable to update base data")
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return errors.Wrap(err, "unable to commit bulk update")
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
