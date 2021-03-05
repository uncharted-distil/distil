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
	"fmt"
	"strings"

	"github.com/jackc/pgx/v4"
	"github.com/uncharted-distil/distil-compute/model"
)

// CopyFromSource represents a function that can use the CopyFrom operation to
// bulk load the data to SQL.
type CopyFromSource func(*Database, string, *Dataset) error

// Dataset is a struct containing the metadata of a dataset being processed.
type Dataset struct {
	ID              string
	Name            string
	Description     string
	Variables       []*model.Variable
	variablesLookup map[string]bool
	insertBatch     *pgx.Batch
	fieldSQL        string
	insertValues    [][]interface{}
	primaryKey      string
	insertFunction  CopyFromSource
}

// NewDataset creates a new dataset instance.
func NewDataset(id, name, description string, variables []*model.Variable, primaryKey string) *Dataset {
	ds := &Dataset{
		ID:              id,
		Name:            name,
		Description:     description,
		variablesLookup: make(map[string]bool),
		primaryKey:      primaryKey,
	}
	// NOTE: Can only support data in a single data resource for now.
	if len(variables) > 0 {
		ds.Variables = variables
		fields := []string{}
		for _, v := range ds.Variables {
			fields = append(fields, v.Key)
		}
		ds.fieldSQL = fmt.Sprintf("\"%s\"", strings.Join(fields, "\",\""))
	}

	ds.ResetBatch()

	return ds
}

// ResetBatch clears the batch contents.
func (ds *Dataset) ResetBatch() {
	ds.insertBatch = &pgx.Batch{}
	ds.insertValues = make([][]interface{}, 0)
}

// HasVariable checks to see if a variable is already contained in the dataset.
func (ds *Dataset) HasVariable(variable *model.Variable) bool {
	return ds.variablesLookup[variable.Key]
}

// AddVariable adds a variable to the dataset.
func (ds *Dataset) AddVariable(variable *model.Variable) {
	ds.Variables = append(ds.Variables, variable)
	ds.variablesLookup[variable.Key] = true
}

// AddInsert adds an insert statement and parameters to the batch.
func (ds *Dataset) AddInsert(statement string, args []interface{}) {
	ds.insertBatch.Queue(statement, args...)
}

// AddInsertFromSource adds a row to insert using the copy protocol.
func (ds *Dataset) AddInsertFromSource(values []interface{}) {
	ds.insertValues = append(ds.insertValues, values)
}

// GetColumns builds captures the variable names in a string slice.
func (ds *Dataset) GetColumns() []string {
	columns := make([]string, len(ds.Variables))
	for i, v := range ds.Variables {
		columns[i] = v.Key
	}

	return columns
}

// GetBatch returns the insert statement batch.
func (ds *Dataset) GetBatch() *pgx.Batch {
	return ds.insertBatch
}

// GetInsertSource returns the insert from source batch.
func (ds *Dataset) GetInsertSource() [][]interface{} {
	return ds.insertValues
}

// GetBatchSize gets the insert batch count.
func (ds *Dataset) GetBatchSize() int {
	return ds.insertBatch.Len()
}

// GetInsertSourceLength gets the insert from source size.
func (ds *Dataset) GetInsertSourceLength() int {
	return len(ds.insertValues)
}

func (ds *Dataset) createTableSQL(tableName string, temp bool, typeAll bool, typeMap map[string]bool) string {
	fieldsSQL := []string{}
	for _, v := range ds.Variables {
		typ := "TEXT"
		if typeAll || (typeMap != nil && typeMap[v.Key]) {
			typ = MapD3MTypeToPostgresType(v.Type)
		}
		fieldsSQL = append(fieldsSQL, fmt.Sprintf("\"%s\" %s", v.Key, typ))
	}

	tempString := ""
	if temp {
		tempString = "TEMP"
	}

	return fmt.Sprintf("CREATE %s TABLE \"%s\" (%s);", tempString, tableName, strings.Join(fieldsSQL, ", "))
}

// GetPrimaryKey returns the primary key of the dataset.
func (ds *Dataset) GetPrimaryKey() string {
	return ds.primaryKey
}
