//
//   Copyright © 2021 Uncharted Software Inc.
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
	log "github.com/unchartedsoftware/plog"
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

func getVariableTableName(storageName string) string {
	return fmt.Sprintf("%s_variable", storageName)
}

// SaveDataset is used for dropping the unused values based on filter param. (Only used in save_dataset route)
func (s *Storage) SaveDataset(dataset string, storageName string, filterParams *api.FilterParams) error {
	// due to values being dropped from base table result table is invalid
	err := s.deleteRows(dataset, s.getResultTable(storageName), nil)
	if err != nil {
		return err
	}
	// due to values being dropped from base table explanation table is invalid
	err = s.deleteRows(dataset, s.getSolutionFeatureWeightTable(storageName), nil)
	if err != nil {
		return err
	}

	if !filterParams.IsEmpty(false) {
		err = s.deleteRows(dataset, getBaseTableName(storageName), filterParams)
		if err != nil {
			return err
		}
	}

	return nil
}

// deleteRows deletes rows based on filterParams
func (s *Storage) deleteRows(dataset string, storageName string, filterParams *api.FilterParams) error {
	wheres := []string{}
	paramsFilter := make([]interface{}, 0)
	wheres, paramsFilter = s.buildFilteredQueryWhere(dataset, wheres, paramsFilter, "", filterParams)
	where := ""
	if len(wheres) > 0 {
		where = "WHERE " + strings.Join(wheres, " AND ")
	}
	sql := fmt.Sprintf("DELETE FROM %s %s;", storageName, where)
	_, err := s.client.Exec(sql, paramsFilter...)
	if err != nil {
		return errors.Wrapf(err, "unable execute query to delete rows")
	}
	return nil
}

// CloneDataset clones an existing dataset.
func (s *Storage) CloneDataset(dataset string, storageName string, datasetNew string, storageNameNew string) error {
	// need to clone base, variable, result, and weight tables
	err := s.cloneTable(getBaseTableName(storageName), getBaseTableName(storageNameNew), true)
	if err != nil {
		return err
	}

	err = s.cloneTable(getVariableTableName(storageName), getVariableTableName(storageNameNew), true)
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
	fields, err := s.getExistingFields(dataset, getBaseTableName(storageNameNew))
	if err != nil {
		return err
	}

	err = s.createView(storageNameNew, fields, true)
	if err != nil {
		return err
	}

	return nil
}

// DeleteDataset drops all tables associated to the dataset
func (s *Storage) DeleteDataset(storageName string) error {
	// drop view
	err := s.dropView(storageName)
	if err != nil {
		return err
	}
	// drop base table
	err = s.dropTable(getBaseTableName(storageName))
	if err != nil {
		return err
	}
	// drop variable table
	err = s.dropTable(getVariableTableName(storageName))
	if err != nil {
		return err
	}
	// drop result table
	err = s.dropTable(s.getResultTable(storageName))
	if err != nil {
		return err
	}
	// drop explanation table
	err = s.dropTable(s.getSolutionFeatureWeightTable(storageName))
	if err != nil {
		return err
	}
	return nil
}
func (s *Storage) dropView(view string) error {
	sql := fmt.Sprintf("DROP VIEW IF EXISTS %s", view)
	_, err := s.client.Exec(sql)
	if err != nil {
		return errors.Wrapf(err, "unable to drop table")
	}
	return nil
}
func (s *Storage) dropTable(table string) error {
	sql := fmt.Sprintf("DROP TABLE IF EXISTS %s", table)
	_, err := s.client.Exec(sql)
	if err != nil {
		return errors.Wrapf(err, "unable to drop table")
	}
	return nil
}
func (s *Storage) cloneTable(existingTable string, newTable string, copyData bool) error {
	// copy indices and columns (this does not copy data need separate query for that)
	sql := fmt.Sprintf("CREATE TABLE %s (LIKE %s INCLUDING ALL);", newTable, existingTable)
	_, err := s.client.Exec(sql)
	if err != nil {
		return errors.Wrapf(err, "unable to clone table")
	}
	// if copy data insert data from other table
	if copyData {
		sql = fmt.Sprintf("INSERT INTO %s SELECT * FROM %s;", newTable, existingTable)
		_, err := s.client.Exec(sql)
		if err != nil {
			return errors.Wrapf(err, "unable to clone table")
		}
	}

	return nil
}

func (s *Storage) getDatabaseFields(tableName string) ([]string, error) {
	sql := "SELECT column_name FROM information_schema.columns WHERE table_schema = 'public' AND table_name = $1;"

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

func (s *Storage) getExistingFields(dataset string, storageName string) (map[string]*model.Variable, error) {
	vars, err := api.FetchDatasetVariables(dataset, s.metadata)
	if err != nil {
		return nil, err
	}

	// make sure they exist in the underlying database already
	fields := make(map[string]*model.Variable)
	for _, v := range vars {
		exists, _ := s.DoesVariableExist(dataset, storageName, v.Key)
		if exists {
			fields[v.Key] = v
		}
	}

	return fields, nil
}

func (s *Storage) createView(storageName string, fields map[string]*model.Variable, overwrite bool) error {
	// CREATE OR REPLACE VIEW requires the same column names and order (with additions at the end being allowed).
	sql := "CREATE VIEW %s_tmp AS SELECT %s FROM %s_base;"

	// Build the select statement of the query.
	fieldList := make([]string, 0)
	for _, v := range fields {
		fieldList = append(fieldList, postgres.GetViewField(v))
	}
	sql = fmt.Sprintf(sql, storageName, strings.Join(fieldList, ","), storageName)

	// Create the temporary view with the new column type.
	_, err := s.client.Exec(sql)
	if err != nil {
		return errors.Wrap(err, "Unable to create new temp view")
	}

	if overwrite {
		// Drop the existing view.
		_, err = s.client.Exec(fmt.Sprintf("DROP VIEW IF EXISTS %s;", storageName))
		if err != nil {
			return errors.Wrap(err, "Unable to drop existing view")
		}

		// Rename the temporary view to the actual view name.
		_, err = s.client.Exec(fmt.Sprintf("ALTER VIEW %s_tmp RENAME TO %s;", storageName, storageName))
	}

	return err
}

func (s *Storage) parseData(rows pgx.Rows) ([][]string, error) {
	output := [][]string{}
	if rows != nil {
		// fields not populated until at least one row has been pulled
		for rows.Next() {
			columnValues, err := rows.Values()
			if err != nil {
				return nil, err
			}
			if len(output) == 0 {
				// parse columns
				fields := rows.FieldDescriptions()
				columns := []string{}
				for i := 0; i < len(fields); i++ {
					switch columnValues[i].(type) {
					case string:
						columns = append(columns, string(fields[i].Name))
					default:
						// handle nested data as separate columns
						nested := unnestStringJSON(columnValues[i])
						for k := range nested {
							columns = append(columns, fmt.Sprintf("%s_%s", fields[i].Name, k))
						}
					}
				}
				output = append(output, columns)
			}

			// read data
			row := make([]string, len(output[0]))
			nestedAdjustment := 0
			for i, cv := range columnValues {
				switch t := cv.(type) {
				case string:
					row[i+nestedAdjustment] = t
				default:
					// assume a map[string]interface{} (explanations)
					nested := unnestStringJSON(cv)
					for _, v := range nested {
						row[i+nestedAdjustment] = v
						nestedAdjustment = nestedAdjustment + 1
					}

					// the nested column itself was already counted
					nestedAdjustment = nestedAdjustment - 1
				}

			}

			output = append(output, row)
		}
		err := rows.Err()
		if err != nil {
			return nil, errors.Wrapf(err, "error reading data from postgres")
		}
	}

	return output, nil
}

func unnestStringJSON(raw interface{}) map[string]string {
	result := map[string]string{}
	if raw == nil {
		return result
	}

	cast := raw.(map[string]interface{})
	for k, v := range cast {
		switch t := v.(type) {
		case float64:
			result[k] = fmt.Sprintf("%f", t)
		case string:
			result[k] = t
		}
	}

	return result
}

// FetchDataset extracts the complete raw data from the database.
func (s *Storage) FetchDataset(dataset string, storageName string,
	includeMetadata bool, limitSelectedFields bool, filterParams *api.FilterParams) ([][]string, error) {
	// get data variables (to exclude metadata variables)
	vars, err := s.metadata.FetchVariables(dataset, true, includeMetadata, false)
	if err != nil {
		return nil, err
	}
	filteredVars := []*model.Variable{}

	selectedVars := map[string]bool{}
	if limitSelectedFields && filterParams != nil {
		for _, v := range filterParams.Variables {
			selectedVars[v] = true
		}
	}

	// only include data with distilrole data and index
	// limit vars to only those selected (if applicable)
	for _, v := range vars {
		if (v.IsTA2Field() ||
			(v.HasRole(model.VarDistilRoleMetadata) && includeMetadata)) &&
			(!limitSelectedFields || selectedVars[v.Key]) {
			filteredVars = append(filteredVars, v)
		}
	}
	varNames := []string{}
	for _, v := range filteredVars {
		fieldSelect := "COALESCE(CAST(\"%s\" as text), '') AS \"%s\""
		if model.IsVector(v.Type) {
			fieldSelect = "COALESCE(TRANSLATE(CAST(\"%s\" as text), '{}', ''), '') AS \"%s\""
		}
		varNames = append(varNames, fmt.Sprintf(fieldSelect, v.Key, v.Key))
	}
	wheres := []string{}
	paramsFilter := make([]interface{}, 0)
	wheres, paramsFilter = s.buildFilteredQueryWhere(dataset, wheres, paramsFilter, "", filterParams)
	where := ""
	if len(wheres) > 0 {
		where = "WHERE " + strings.Join(wheres, " AND ")
	}
	sql := fmt.Sprintf("SELECT %s FROM %s %s;", strings.Join(varNames, ", "), getBaseTableName(storageName), where)
	res, err := s.client.Query(sql, paramsFilter...)
	if err != nil {
		return nil, errors.Wrapf(err, "unable execute query to extract dataset")
	}

	return s.parseData(res)
}
func (s *Storage) createIndex(storageName string, colName string, colType string) error {
	sql := postgres.GetIndexStatement(storageName, colName, colType)

	_, err := s.client.Exec(sql)
	if err != nil {
		return errors.Wrapf(err, "unable to create postgres index")
	}

	return nil
}

// CreateIndices generates indices for the suppled fields on the "dataset"_base table
func (s *Storage) CreateIndices(dataset string, indexFields []string) error {
	variables, err := s.metadata.FetchVariables(dataset, true, true, true)
	if err != nil {
		return err
	}
	ds, err := s.metadata.FetchDataset(dataset, false, false, false)
	if err != nil {
		return err
	}
	mappedVariables := map[string]*model.Variable{}
	for _, v := range variables {
		mappedVariables[v.Key] = v
	}
	for _, fieldName := range indexFields {
		field := mappedVariables[fieldName]
		log.Infof("creating index on %s", field.Key)
		err := s.createIndex(getBaseTableName(ds.StorageName), field.Key, field.Type)
		if err != nil {
			return err
		}
	}
	return nil
}

// IsValidDataType checks to see if a specified type is valid for a variable.
// Multiple simultaneous calls to the function can result in inaccurate.
func (s *Storage) IsValidDataType(dataset string, storageName string, varName string, varType string) (bool, error) {
	// get all existing fields to rebuild the view.
	fields, err := s.getExistingFields(dataset, storageName)
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
	verificationSQL := fmt.Sprintf("SELECT COUNT(\"%s\") FROM %s_tmp WHERE \"%s\" != %v;",
		varName, storageName, varName, postgres.DefaultPostgresValueFromD3MType(varType))
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
	// geometry types need special handling to make sure the backing field exists properly
	if model.IsGeoBounds(varType) {
		err := s.createGeometryField(dataset, storageName, varName)
		if err != nil {
			return err
		}
	}

	// get all existing fields to rebuild the view.
	fields, err := s.getExistingFields(dataset, getBaseTableName(storageName))
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
			Key:              v.Key,
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
func (s *Storage) AddVariable(dataset string, storageName string, key string, varType string, defaultVal string) error {
	// check to make sure the column doesnt exist already
	dbFields, err := s.getDatabaseFields(fmt.Sprintf("%s_base", storageName))
	if err != nil {
		return errors.Wrap(err, "unable to read database fields")
	}

	found := false
	for _, v := range dbFields {
		if v == key {
			found = true
			break
		}
	}

	if !found {
		defaultClause := ""
		if len(defaultVal) > 0 {
			defaultClause = fmt.Sprintf(" Default '%s'", defaultVal)
		}

		// geometry is not stored as a text field
		fieldType := "TEXT"
		if model.IsGeoBounds(varType) {
			fieldType = postgres.MapD3MTypeToPostgresType(varType)
		}

		// add the empty column to the base table and the explain table
		sql := fmt.Sprintf("ALTER TABLE %s_base ADD COLUMN \"%s\" %s%s;", storageName, key, fieldType, defaultClause)

		_, err = s.client.Exec(sql)
		if err != nil {
			return errors.Wrap(err, "unable to add new column to database table")
		}

		sql = fmt.Sprintf("ALTER TABLE %s_explain ADD COLUMN \"%s\" DOUBLE PRECISION;", storageName, key)
		_, err = s.client.Exec(sql)
		if err != nil {
			return errors.Wrap(err, "unable to add new column to database explain table")
		}
	}

	// recreate the view with the new column
	fields, err := s.getExistingFields(dataset, storageName)
	if err != nil {
		return errors.Wrap(err, "unable to read existing fields")
	}

	// need to add the field to the view
	fields[key] = &model.Variable{
		Key:              key,
		OriginalVariable: key,
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
func (s *Storage) AddField(dataset string, storageName string, varName string, varType string, defaultVal string) error {
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

	defaultClause := ""
	if len(defaultVal) > 0 {
		defaultClause = fmt.Sprintf(" Default '%s'", defaultVal)
	}
	sql := fmt.Sprintf("ALTER TABLE %s ADD COLUMN \"%s\" %s%s;", storageName, varName, postgres.MapD3MTypeToPostgresType(varType), defaultClause)
	_, err = s.client.Exec(sql)
	if err != nil {
		return errors.Wrap(err, "unable to add new column to database explain table")
	}

	return nil
}

// DeleteVariable flags a variable as deleted.
func (s *Storage) DeleteVariable(dataset string, storageName string, varName string) error {
	// check if the variable is in the view
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
		return nil
	}

	// recreate the view without the field if it is in it
	fields, err := s.getExistingFields(dataset, storageName)
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

// SetVariableValue updates an entire column to specified value
func (s *Storage) SetVariableValue(dataset string, storageName string, varName string, value string, filterParams *api.FilterParams) error {
	wheres := []string{}
	params := []interface{}{value}
	wheres, params = s.buildFilteredQueryWhere(dataset, wheres, params, "", filterParams)
	whereClause := ""
	if len(wheres) > 0 {
		whereClause = fmt.Sprintf("WHERE %s", strings.Join(wheres, " AND "))
	}
	sql := fmt.Sprintf("UPDATE %s_base SET \"%s\" = $1 %s;", storageName, varName, whereClause)
	_, err := s.client.Exec(sql, params...)
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

	tx, err := s.batchClient.Begin()
	if err != nil {
		_ = tx.Rollback(context.Background())
		return errors.Wrap(err, "unable to create transaction")
	}

	// loop through the updates, building batches to minimize overhead
	tableNameTmp := fmt.Sprintf("%s_utmp", storageName)
	dataSQL := fmt.Sprintf("CREATE TEMP TABLE \"%s\" (\"%s\" TEXT NOT NULL, \"%s\" TEXT) ON COMMIT DROP;",
		tableNameTmp, model.D3MIndexFieldName, varName)
	_, err = tx.Exec(context.Background(), dataSQL)
	if err != nil {
		_ = tx.Rollback(context.Background())
		return errors.Wrap(err, "unable to create temp table")
	}

	err = s.insertBulkCopyTransaction(tx, tableNameTmp, []string{model.D3MIndexFieldName, varName}, params)
	if err != nil {
		_ = tx.Rollback(context.Background())
		return errors.Wrap(err, "unable to insert into temp table")
	}

	// run the update
	updateSQL := fmt.Sprintf("UPDATE %s.%s.\"%s_base\" AS b SET \"%s\" = t.\"%s\" FROM \"%s\" AS t WHERE t.\"%s\" = b.\"%s\";",
		"distil", "public", storageName, varName, varName, tableNameTmp, model.D3MIndexFieldName, model.D3MIndexFieldName)
	_, err = tx.Exec(context.Background(), updateSQL)
	if err != nil {
		_ = tx.Rollback(context.Background())
		return errors.Wrap(err, "unable to update base data")
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return errors.Wrap(err, "unable to commit bulk update")
	}

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
		_ = tx.Rollback(context.Background())
		return errors.Wrap(err, "unable to create transaction")
	}

	tableNameTmp := fmt.Sprintf("%s_utmp", storageName)
	dataSQL := fmt.Sprintf("CREATE TEMP TABLE \"%s\" (\"%s\" TEXT NOT NULL, \"%s\" TEXT) ON COMMIT DROP;",
		tableNameTmp, model.D3MIndexFieldName, varName)
	_, err = tx.Exec(context.Background(), dataSQL)
	if err != nil {
		_ = tx.Rollback(context.Background())
		return errors.Wrap(err, "unable to create temp table")
	}

	err = s.insertBulkCopyTransaction(tx, tableNameTmp, []string{model.D3MIndexFieldName, varName}, params)
	if err != nil {
		_ = tx.Rollback(context.Background())
		return errors.Wrap(err, "unable to insert into temp table")
	}

	// build the filter structure
	wheres := []string{fmt.Sprintf("t.\"%s\" = b.\"%s\"::text", model.D3MIndexFieldName, model.D3MIndexFieldName)}
	paramsFilter := make([]interface{}, 0)
	wheres, paramsFilter = s.buildFilteredQueryWhere(dataset, wheres, paramsFilter, "b", filterParams)

	// geometries should be updated slightly differently
	// they should be reduced to their centroid
	varMeta, err := s.metadata.FetchVariable(dataset, varName)
	if err != nil {
		_ = tx.Rollback(context.Background())
		return err
	}

	updateValue := fmt.Sprintf("\"%s\" = t.\"%s\"", varName, varName)
	if model.IsGeoBounds(varMeta.Type) {
		updateValue = fmt.Sprintf("\"%s\" = ST_CENTROID(t.\"%s\"::geometry)", varName, varName)
	}

	// run the update
	updateSQL := fmt.Sprintf("UPDATE %s.%s.\"%s\" AS b SET %s FROM \"%s\" AS t WHERE %s",
		"distil", "public", storageName, updateValue, tableNameTmp, strings.Join(wheres, " AND "))
	_, err = tx.Exec(context.Background(), updateSQL, paramsFilter...)
	if err != nil {
		if rbErr := tx.Rollback(context.Background()); rbErr != nil {
			log.Error("rollback failed")
		}
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

func (s *Storage) createGeometryField(dataset string, storageName string, varName string) error {
	postgisFieldName := fmt.Sprintf("__geo_%s", varName)
	exists, _ := s.DoesVariableExist(dataset, storageName, postgisFieldName)
	baseTable := getBaseTableName(storageName)
	if !exists {
		err := s.AddField(dataset, baseTable, postgisFieldName, model.GeoBoundsType, "")
		if err != nil {
			return err
		}

		// create index on the field
		// less efficient than doing it after ingest, but a bit cleaner codewise
		err = s.createIndex(baseTable, postgisFieldName, model.GeoBoundsType)
		if err != nil {
			return err
		}
	}

	// query the vector field to get the data for the geometry field
	querySQL := fmt.Sprintf("SELECT \"%s\", concat('{', \"%s\", '}')::double precision[] FROM %s;", model.D3MIndexFieldName, varName, baseTable)
	rows, err := s.client.Query(querySQL)
	if err != nil {
		return errors.Wrapf(err, "unable to query for geobounds")
	}

	updates := map[string]string{}
	for rows.Next() {
		var d3mIndex string
		var geometry []float64
		err := rows.Scan(&d3mIndex, &geometry)
		if err != nil {
			return errors.Wrapf(err, "unable to read geometry field from postgres")
		}
		if len(geometry) != 8 {
			return errors.Errorf("field '%s' is not a vector of 4 points", varName)
		}

		// add the link back to the first point since a polygon must be closed
		geometry = append(geometry, geometry[0], geometry[1])

		// geometry string should have the coordinates of a point separate by a space
		// and the points separated by commas
		geometryString := ""
		for i := 0; i < len(geometry); i += 2 {
			geometryString = fmt.Sprintf("%s,%f %f", geometryString, geometry[i], geometry[i+1])
		}
		updates[d3mIndex] = fmt.Sprintf("POLYGON((%s))", geometryString[1:])
	}
	err = rows.Err()
	if err != nil {
		return errors.Wrapf(err, "error reading geometry data from postgres")
	}

	err = s.UpdateData(dataset, baseTable, postgisFieldName, updates, nil)
	if err != nil {
		return err
	}

	return nil
}
