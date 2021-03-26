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
	"regexp"
	"strings"
	"time"

	"github.com/pkg/errors"
	log "github.com/unchartedsoftware/plog"

	"github.com/uncharted-distil/distil-compute/model"
)

const (
	// Database data types
	dataTypeText     = "TEXT"
	dataTypeDouble   = "double precision"
	dataTypeFloat    = "FLOAT8"
	dataTypeVector   = "FLOAT8[]"
	dataTypeGeometry = "geometry"
	dataTypeInteger  = "INTEGER"
	dataTypeDate     = "TIMESTAMP"
	dataTypeBool     = "BOOLEAN"
	dataTypeImageExt = "IMAGE_EXT"
	dataTypeEmail    = "EMAIL"
	dataTypeLat      = "LATITUDE"
	dataTypeLon      = "LONGITUDE"
	dataTypeCoord    = "SPECIAL_COORD"
	dateFormat       = "2006-01-02T15:04:05Z"

	metadataTableCreationSQL = `CREATE TABLE %s (
			name	text	NOT NULL,
			role	varchar(100),
			type	varchar(100)
		);`
	resultTableCreationSQL = `CREATE TABLE %s (
			result_id	text	NOT NULL,
			index		BIGINT,
			target		text,
			value		text,
			explain_values jsonb
		);`

	// PredictionTableName is the name of the table for prediction requests.
	PredictionTableName = "prediction"
	// RequestTableName is the name of the table for solution requests.
	RequestTableName = "request"

	// SolutionTableName is the name of the table for solutions.
	SolutionTableName = "solution"
	// SolutionFeatureWeightTableName is the name of the table for solution feature weights.
	SolutionFeatureWeightTableName = "solution_weight"
	// SolutionStateTableName is the name of the table for solution state.
	SolutionStateTableName = "solution_state"
	// SolutionResultTableName is the name of the table for the result.
	SolutionResultTableName = "solution_result"
	// SolutionResultExplainOutputTableName is the name of the table for the result explain output.
	SolutionResultExplainOutputTableName = "solution_result_explain"
	// SolutionScoreTableName is the name of the table for the score.
	SolutionScoreTableName = "solution_score"
	// RequestFeatureTableName is the name of the table for the request features.
	RequestFeatureTableName = "request_feature"
	// RequestFilterTableName is the name of the table for the request filters.
	RequestFilterTableName = "request_filter"
	// WordStemTableName is the name of the table for the word stems.
	WordStemTableName = "word_stem"

	requestTableCreationSQL = `CREATE TABLE %s (
			request_id			text,
			dataset				varchar(200),
			progress			varchar(40),
			created_time		timestamp,
			last_updated_time	timestamp
		);`
	predictionTableCreationSQL = `CREATE TABLE %s (
				request_id			text,
				dataset				varchar(200),
				target		text,
				fitted_solution_id	text,
				progress			varchar(40),
				created_time		timestamp,
				last_updated_time	timestamp
			);`
	solutionTableCreationSQL = `CREATE TABLE %s (
			request_id		text,
			solution_id		text,
			explained_solution_id text,
			created_time	timestamp,
			deleted         boolean
		);`
	solutionFeatureWeightTableCreationSQL = `CREATE TABLE %s (
			solution_id	text,
			feature_name	text,
			feature_index int,
			weight		double precision
		);`
	solutionStateTableCreationSQL = `CREATE TABLE %s (
			solution_id		text,
			progress		varchar(40),
			created_time	timestamp
		);`
	requestFeatureTableCreationSQL = `CREATE TABLE %s (
			request_id		text,
			feature_name	text,
			feature_type	varchar(20)
		);`
	requestFilterTableCreationSQL = `CREATE TABLE %s (
			request_id			text,
			feature_name		text,
			filter_type			varchar(40),
			filter_mode			varchar(40),
			filter_min			double precision,
			filter_max			double precision,
			filter_min_x		double precision,
			filter_max_x		double precision,
			filter_min_y		double precision,
			filter_max_y		double precision,
			filter_categories	varchar(200),
			filter_indices		varchar(200)
		);`
	modelFeatureWeightTableCreationSQL = `CREATE TABLE %s (
			result_id	text	NOT NULL,
			%s
		);`
	solutionScoreTableCreationSQL = `CREATE TABLE %s (
			solution_id	text,
			metric		varchar(40),
			score		double precision
		);`
	solutionResultTableCreationSQL = `CREATE TABLE %s (
			solution_id			text,
			fitted_solution_id	text,
			produce_request_id text,
			result_type varchar(20),
			result_uuid			text,
			result_uri			text,
			progress			varchar(40),
			created_time		timestamp
		);`
	solutionResultExplainTableCreationSQL = `CREATE TABLE %s (
			result_id	text,
			explain_uri	text,
			explain_type	text
		);`
	wordStemsTableCreationSQL = `CREATE TABLE %s (
			stem		text PRIMARY KEY,
			word		text
		);`

	resultTableSuffix   = "_result"
	variableTableSuffix = "_variable"
	explainTableSuffix  = "_explain"
)

var (
	nonNullableTypes = map[string]bool{
		"index":   true,
		"integer": true,
		"float":   true,
		"real":    true,
	}
	wordRegex     = regexp.MustCompile("[^a-zA-Z]")
	resultIndices = []string{
		"result_id",
	}
)

// Database is a struct representing a full logical database.
type Database struct {
	Client            DatabaseDriver
	Tables            map[string]*Dataset
	BatchSize         int
	BatchSizeStemWord int
	wordStemCache     map[string]bool
}

// WordStem contains the pairing of a word and its stemmed version.
type WordStem struct {
	Word string
	Stem string
}

// Config has the necessary configuration values for a postgres connection.
type Config struct {
	Database         string
	Table            string
	User             string
	Password         string
	Port             int
	Host             string
	BatchSize        int
	PostgresLogLevel string
}

// NewDatabase creates a new database instance.
func NewDatabase(config *Config, batch bool) (*Database, error) {
	client, err := NewClient(config.Host, config.Port, config.User, config.Password, config.Database, config.PostgresLogLevel, batch)()
	if err != nil {
		return nil, err
	}

	database := &Database{
		Client:            client,
		Tables:            make(map[string]*Dataset),
		BatchSize:         config.BatchSize,
		BatchSizeStemWord: config.BatchSize * 10,
		wordStemCache:     make(map[string]bool),
	}

	database.Tables[WordStemTableName] = NewDataset(WordStemTableName, WordStemTableName, "",
		[]*model.Variable{{Key: "stem"}, {Key: "word"}}, "stem")
	database.Tables[WordStemTableName].insertFunction = insertFromSourceUnique

	return database, nil
}

// CreateSolutionMetadataTables creates an empty table for the solution results.
func (d *Database) CreateSolutionMetadataTables() error {
	// Create the solution tables.
	log.Infof("Creating solution metadata tables.")

	_ = d.DropTable(PredictionTableName)
	_, err := d.Client.Exec(fmt.Sprintf(predictionTableCreationSQL, PredictionTableName))
	if err != nil {
		return errors.Wrap(err, "failed to drop table")
	}

	_ = d.DropTable(RequestTableName)
	_, err = d.Client.Exec(fmt.Sprintf(requestTableCreationSQL, RequestTableName))
	if err != nil {
		return errors.Wrap(err, "failed to drop table")
	}

	_ = d.DropTable(RequestFeatureTableName)
	_, err = d.Client.Exec(fmt.Sprintf(requestFeatureTableCreationSQL, RequestFeatureTableName))
	if err != nil {
		return errors.Wrap(err, "failed to drop table")
	}

	_ = d.DropTable(RequestFilterTableName)
	_, err = d.Client.Exec(fmt.Sprintf(requestFilterTableCreationSQL, RequestFilterTableName))
	if err != nil {
		return errors.Wrap(err, "failed to drop table")
	}

	_ = d.DropTable(SolutionTableName)
	_, err = d.Client.Exec(fmt.Sprintf(solutionTableCreationSQL, SolutionTableName))
	if err != nil {
		return errors.Wrap(err, "failed to drop table")
	}

	_ = d.DropTable(SolutionFeatureWeightTableName)
	_, err = d.Client.Exec(fmt.Sprintf(solutionFeatureWeightTableCreationSQL, SolutionFeatureWeightTableName))
	if err != nil {
		return errors.Wrap(err, "failed to drop table")
	}

	_ = d.DropTable(SolutionStateTableName)
	_, err = d.Client.Exec(fmt.Sprintf(solutionStateTableCreationSQL, SolutionStateTableName))
	if err != nil {
		return errors.Wrap(err, "failed to drop table")
	}

	_ = d.DropTable(SolutionResultTableName)
	_, err = d.Client.Exec(fmt.Sprintf(solutionResultTableCreationSQL, SolutionResultTableName))
	if err != nil {
		return errors.Wrap(err, "failed to drop table")
	}

	_ = d.DropTable(SolutionResultExplainOutputTableName)
	_, err = d.Client.Exec(fmt.Sprintf(solutionResultExplainTableCreationSQL, SolutionResultExplainOutputTableName))
	if err != nil {
		return errors.Wrap(err, "failed to drop table")
	}

	_ = d.DropTable(SolutionScoreTableName)
	_, err = d.Client.Exec(fmt.Sprintf(solutionScoreTableCreationSQL, SolutionScoreTableName))
	if err != nil {
		return errors.Wrap(err, "failed to drop table")
	}

	// do not drop the word stem table as we want it to include all words.
	_, _ = d.Client.Exec(fmt.Sprintf(wordStemsTableCreationSQL, WordStemTableName))
	// ignore the error in the word stem creation.
	// Almost certainly due to the table already existing.

	return nil
}

// insertFromSourceUnique function only inserting unique rows.
func insertFromSourceUnique(d *Database, tableName string, ds *Dataset) error {
	// first ingest to a temporary table
	tmpTableName := fmt.Sprintf("tmp_%s", tableName)
	createSQL := ds.createTableSQL(tmpTableName, true, true, nil)
	_, err := d.Client.Exec(createSQL)
	if err != nil {
		return errors.Wrapf(err, "unable to create tmp table for inserts")
	}
	// drop the temp table
	defer func() {
		_ = d.DropTable(tmpTableName)
	}()

	tmpInsertCount, err := d.Client.CopyFrom(tmpTableName, ds.GetColumns(), ds.GetInsertSource())
	if err != nil {
		return errors.Wrapf(err, "unable to insert batch to postgres")
	}
	if tmpInsertCount != int64(ds.GetInsertSourceLength()) {
		return errors.Errorf("batch insert only copied %d rows from source out of %d", tmpInsertCount, ds.GetInsertSourceLength())
	}

	// then copy from the temp table to the real table all new rows
	updateSQL := fmt.Sprintf("INSERT INTO \"%s\" SELECT d.* FROM \"%s\" AS d WHERE NOT EXISTS (SELECT 1 FROM \"%s\" AS d2 WHERE d.\"%s\" = d2.\"%s\");",
		tableName, tmpTableName, tableName, ds.GetPrimaryKey(), ds.GetPrimaryKey())
	_, err = d.Client.Exec(updateSQL)
	if err != nil {
		return errors.Wrapf(err, "unable to insert from tmp table to base table")
	}

	return nil
}

// insertFromSourceBase function implementing a basic approach to using copy from.
func insertFromSourceBase(d *Database, tableName string, ds *Dataset) error {
	insertCount, err := d.Client.CopyFrom(fmt.Sprintf("%s_base", tableName), ds.GetColumns(), ds.GetInsertSource())
	if err != nil {
		return errors.Wrapf(err, "unable to insert batch to postgres")
	}
	if insertCount != int64(ds.GetInsertSourceLength()) {
		return errors.Errorf("batch insert only copied %d rows from source out of %d", insertCount, ds.GetInsertSourceLength())
	}

	return nil
}

// insertFromSourceGeometry function for inserting geometries using copy from.
func insertFromSourceGeometry(d *Database, tableName string, ds *Dataset) error {
	// first ingest to a temporary table
	tmpTableName := fmt.Sprintf("tmp_%s", tableName)
	createSQL := ds.createTableSQL(tmpTableName, true, false, nil)
	_, err := d.Client.Exec(createSQL)
	if err != nil {
		return errors.Wrapf(err, "unable to create tmp table for inserts")
	}
	// drop the temp table
	defer func() {
		_ = d.DropTable(tmpTableName)
	}()

	tmpInsertCount, err := d.Client.CopyFrom(tmpTableName, ds.GetColumns(), ds.GetInsertSource())
	if err != nil {
		return errors.Wrapf(err, "unable to insert batch to postgres")
	}
	if tmpInsertCount != int64(ds.GetInsertSourceLength()) {
		return errors.Errorf("batch insert only copied %d rows from source out of %d", tmpInsertCount, ds.GetInsertSourceLength())
	}

	// then copy from the temp table to the real table, casting geometry fields
	fields := []string{}
	for _, v := range ds.Variables {
		typ := dataTypeText
		if v.Key == model.D3MIndexFieldName || v.Type == model.GeoBoundsType {
			typ = MapD3MTypeToPostgresType(v.Type)
		}
		fields = append(fields, fmt.Sprintf("\"%s\"::%s", v.Key, typ))
	}
	fieldSQL := strings.Join(fields, ",")
	updateSQL := fmt.Sprintf("INSERT INTO \"%s_base\" SELECT %s FROM \"%s\";", tableName, fieldSQL, tmpTableName)
	_, err = d.Client.Exec(updateSQL)
	if err != nil {
		return errors.Wrapf(err, "unable to copy geometry data from temp table")
	}

	return nil
}

func (d *Database) executeInserts(tableName string) error {
	ds := d.Tables[tableName]
	if ds.GetInsertSourceLength() > 0 {
		err := ds.insertFunction(d, tableName, ds)
		if err != nil {
			return errors.Wrapf(err, "unable to copy from source to postgres")
		}
	}

	batchCount := ds.GetBatchSize()
	if batchCount > 0 {
		batch := ds.GetBatch()
		res := d.Client.SendBatch(batch)
		defer res.Close()
		for i := 0; i < batchCount; i++ {
			_, err := res.Exec()
			if err != nil {
				return errors.Wrapf(err, "unable to insert batch to postgres")
			}
		}
	}

	return nil
}

// CreateResultTable creates an empty table for the solution results.
func (d *Database) CreateResultTable(tableName string) error {
	resultTableName := fmt.Sprintf("%s%s", tableName, resultTableSuffix)

	// Make sure the table is clear. If the table did not previously exist,
	// an error is returned. May as well ignore it since a serious problem
	// will cause errors on the other statements as well.
	err := d.DropTable(resultTableName)
	if err != nil {
		return err
	}

	// Create the variable table.
	log.Infof("Creating result table %s", resultTableName)
	createStatement := fmt.Sprintf(resultTableCreationSQL, resultTableName)
	_, err = d.Client.Exec(createStatement)
	if err != nil {
		return err
	}

	log.Infof("creating indices on result table %s", resultTableName)
	for i, index := range resultIndices {
		indexStatement := fmt.Sprintf("CREATE INDEX idx%d_%s ON %s (%s)", i+1, resultTableName, resultTableName, index)
		_, err = d.Client.Exec(indexStatement)
		if err != nil {
			return err
		}
	}

	return nil
}

// StoreMetadata stores the variable information to the specified table.
func (d *Database) StoreMetadata(tableName string) error {
	variableTableName := fmt.Sprintf("%s%s", tableName, variableTableSuffix)

	// Make sure the table is clear. If the table did not previously exist,
	// an error is returned. May as well ignore it since a serious problem
	// will cause errors on the other statements as well.
	err := d.DropTable(variableTableName)
	if err != nil {
		return err
	}

	// Create the variable table.
	log.Infof("Creating variable table %s", variableTableName)
	createStatement := fmt.Sprintf(metadataTableCreationSQL, variableTableName)
	_, err = d.Client.Exec(createStatement)
	if err != nil {
		return err
	}

	// Insert the variable metadata into the new table.
	for _, v := range d.Tables[tableName].Variables {
		insertStatement := fmt.Sprintf("INSERT INTO %s (name, role, type) VALUES ($1, $2, $3);", variableTableName)
		values := []interface{}{v.Key, v.DistilRole, v.Type}
		_, err = d.Client.Exec(insertStatement, values...)
		if err != nil {
			return errors.Wrapf(err, "unable to store variable in postgres")
		}
	}

	return nil
}

// DeleteDataset deletes all tables & views for a dataset.
func (d *Database) DeleteDataset(name string) {
	baseName := fmt.Sprintf("%s_base", name)
	resultName := fmt.Sprintf("%s%s", name, resultTableSuffix)
	variableName := fmt.Sprintf("%s%s", name, variableTableSuffix)
	explainName := fmt.Sprintf("%s%s", name, explainTableSuffix)

	_ = d.DropView(name)
	_ = d.DropTable(baseName)
	_ = d.DropTable(resultName)
	_ = d.DropTable(variableName)
	_ = d.DropTable(explainName)
}

// IngestRow parses the raw csv data and stores it to the table specified.
// The previously parsed metadata is used to map columns.
func (d *Database) IngestRow(tableName string, data []string) error {
	ds := d.Tables[tableName]

	variables := ds.Variables
	values := make([]interface{}, len(variables))
	for i := 0; i < len(variables); i++ {
		// Default columns that have an empty column.
		var val interface{}
		if d.isNullVariable(variables[i].Type, data[i]) {
			val = nil
		} else if d.isArray(variables[i].Type) && !d.dataIsArray(data[i]) {
			val = fmt.Sprintf("{%s}", data[i])
		} else if variables[i].Type == model.GeoBoundsType+"s" {
			val = fmt.Sprintf("%s:geometry", data[i])
		} else {
			val = data[i]
		}
		values[i] = val
	}
	ds.AddInsertFromSource(values)

	if ds.GetInsertSourceLength() >= d.BatchSize {

		err := d.executeInserts(tableName)
		if err != nil {
			return errors.Wrap(err, "unable to insert to table "+tableName)
		}

		ds.ResetBatch()
	}

	return nil
}

// InsertRemainingRows empties all batches and inserts the data to the database.
func (d *Database) InsertRemainingRows() error {
	for tableName, ds := range d.Tables {
		if ds.GetBatchSize() > 0 || ds.GetInsertSourceLength() > 0 {
			err := d.executeInserts(tableName)
			if err != nil {
				return errors.Wrap(err, "unable to insert remaining rows for table "+tableName)
			}

			if tableName != WordStemTableName {
				tableName = fmt.Sprintf("%s_base", tableName)
			}
			_, err = d.Client.Exec(fmt.Sprintf("ANALYZE \"%s\"", tableName))
			if err != nil {
				log.Warnf("error updating stats for %s: %+v", tableName, err)
			}
		}
	}

	return nil
}

// AddWordStems builds the word stemming lookup in the database.
func (d *Database) AddWordStems(data []string) error {
	ds := d.Tables[WordStemTableName]

	for i := 0; i < len(data); i++ {
		// split the field into tokens.
		fields := strings.Fields(data[i])
		for _, f := range fields {
			fieldValue := wordRegex.ReplaceAllString(f, "")
			if fieldValue == "" || d.wordStemCache[fieldValue] {
				continue
			}

			// query for the stemmed version of each word.
			query := fmt.Sprintf("INSERT INTO %s VALUES (unnest(tsvector_to_array(to_tsvector($1))), $2) ON CONFLICT (stem) DO NOTHING;", WordStemTableName)
			ds.AddInsert(query, []interface{}{fieldValue, strings.ToLower(fieldValue)})
			d.wordStemCache[fieldValue] = true
			if ds.GetBatchSize() >= d.BatchSizeStemWord {
				err := d.executeInserts(WordStemTableName)
				if err != nil {
					return errors.Wrap(err, "unable to insert to table "+WordStemTableName)
				}

				ds.ResetBatch()
				d.wordStemCache = make(map[string]bool)
			}
		}
	}

	return nil
}

// DropTable drops the specified table from the database.
func (d *Database) DropTable(tableName string) error {
	log.Infof("Dropping table %s", tableName)
	drop := fmt.Sprintf("DROP TABLE IF EXISTS %s;", tableName)
	_, err := d.Client.Exec(drop)
	log.Infof("Dropped table %s", tableName)

	return err
}

// DropView drops the specified view from the database.
func (d *Database) DropView(viewName string) error {
	log.Infof("Dropping view %s", viewName)
	drop := fmt.Sprintf("DROP VIEW IF EXISTS %s;", viewName)
	_, err := d.Client.Exec(drop)
	log.Infof("Dropped view %s", viewName)

	return err
}

// InitializeTable generates and runs a table create statement based on the schema.
func (d *Database) InitializeTable(tableName string, ds *Dataset) error {
	d.Tables[tableName] = ds

	// Create the view and table statements as well as the feature weight table.
	// The table has almost everything stored as a string.
	// The view uses casting to set the types.
	createStatementTable := `CREATE TABLE %s_base (%s);`
	createStatementView := `CREATE VIEW %s AS SELECT %s FROM %s_base;`
	varsTable := ""
	varsView := ""
	varsExplain := ""
	for _, variable := range ds.Variables {
		tableType := dataTypeText
		viewVar := fmt.Sprintf("COALESCE(CAST(%s AS %s), %v) AS \"%s\"", ValueForFieldType(variable.Type, variable.Key),
			MapD3MTypeToPostgresType(variable.Type), DefaultPostgresValueFromD3MType(variable.Type), variable.Key)

		// it needs to be a geometry if it was originally typed as a geobounds
		if variable.Type == model.GeoBoundsType || variable.OriginalType == model.GeoBoundsType {
			tableType = dataTypeGeometry
			viewVar = fmt.Sprintf("\"%s\"", variable.Key)
		}
		varsTable = fmt.Sprintf("%s\n\"%s\" %s,", varsTable, variable.Key, tableType)
		varsExplain = fmt.Sprintf("%s\n\"%s\" DOUBLE PRECISION,", varsExplain, variable.Key)
		varsView = fmt.Sprintf("%s\n%s,", varsView, viewVar)
	}
	if len(varsTable) > 0 {
		varsTable = varsTable[:len(varsTable)-1]
		varsView = varsView[:len(varsView)-1]
		varsExplain = varsExplain[:len(varsExplain)-1]
	}
	createStatementTable = fmt.Sprintf(createStatementTable, tableName, varsTable)
	log.Infof("Creating table %s_base", tableName)

	// Create the table.
	_, err := d.Client.Exec(createStatementTable)
	if err != nil {
		return err
	}

	createStatementView = fmt.Sprintf(createStatementView, tableName, varsView, tableName)
	log.Infof("Creating view %s", tableName)

	// Create the view.
	_, err = d.Client.Exec(createStatementView)
	if err != nil {
		return err
	}

	explainName := fmt.Sprintf("%s%s", tableName, explainTableSuffix)
	createStatementExplain := fmt.Sprintf(modelFeatureWeightTableCreationSQL, explainName, varsExplain)
	log.Infof("Creating table %s", explainName)

	// Create the feature weight table.
	_, err = d.Client.Exec(createStatementExplain)
	if err != nil {
		return err
	}

	return nil
}

// InitializeDataset initializes the dataset with the provided metadata.
func (d *Database) InitializeDataset(meta *model.Metadata) (*Dataset, error) {
	// geobounds data not batched using copy from due to issues with loading of the data
	loadFunction := insertFromSourceBase
	for _, v := range meta.GetMainDataResource().Variables {
		if v.Type == model.GeoBoundsType || v.OriginalType == model.GeoBoundsType {
			loadFunction = insertFromSourceGeometry
			break
		}
	}
	ds := NewDataset(meta.ID, meta.Name, meta.Description, meta.GetMainDataResource().Variables, "")
	ds.insertFunction = loadFunction

	return ds, nil
}

func (d *Database) isNullVariable(typ, value string) bool {
	return value == "" && nonNullableTypes[typ]
}

func (d *Database) isArray(typ string) bool {
	return strings.HasSuffix(typ, "Vector")
}

func (d *Database) dataIsArray(data string) bool {
	dataLength := len(data)
	if dataLength < 2 {
		return false
	}

	return data[0] == '{' && data[dataLength-1] == '}'
}

// MapD3MTypeToPostgresType generates a postgres type from a d3m type.
func MapD3MTypeToPostgresType(typ string) string {
	// Integer types can be returned as floats.
	switch typ {
	case model.IndexType:
		return dataTypeInteger
	case model.IntegerType, model.LongitudeType, model.LatitudeType, model.RealType, model.TimestampType:
		return dataTypeFloat
	case model.OrdinalType, model.CategoricalType, model.StringType:
		return dataTypeText
	case model.DateTimeType:
		return dataTypeDate
	case model.GeoBoundsType:
		return dataTypeGeometry
	case model.RealVectorType, model.RealListType:
		return dataTypeVector
	default:
		return dataTypeText
	}
}

// MapPostgresTypeToD3MType converts postgres types to D3M types
func MapPostgresTypeToD3MType(pType string) ([]string, error) {
	switch pType {
	case dataTypeDate:
		return []string{model.DateTimeType}, nil
	case dataTypeDouble:
		return []string{model.RealType, model.TimestampType}, nil
	case dataTypeFloat:
		return []string{model.RealType, model.TimestampType}, nil
	case dataTypeInteger:
		return []string{model.TimestampType, model.IntegerType}, nil
	case dataTypeVector, dataTypeCoord:
		return []string{model.RealVectorType, model.RealListType}, nil
	case dataTypeText:
		return []string{model.OrdinalType, model.CategoricalType, model.StringType, model.AddressType, model.CityType, model.CountryType, model.PostalCodeType, model.StateType, model.URIType, model.PhoneType}, nil
	case dataTypeGeometry:
		return []string{model.GeoBoundsType}, nil
	case dataTypeBool:
		return []string{model.BoolType}, nil
	case dataTypeEmail:
		return []string{model.EmailType}, nil
	case dataTypeLat:
		return []string{model.LatitudeType}, nil
	case dataTypeLon:
		return []string{model.LongitudeType}, nil
	case dataTypeImageExt:
		return []string{model.ImageType, model.MultiBandImageType}, nil
	default:
		return []string{}, errors.New("pType is not a supported type")
	}
}

// DefaultPostgresValueFromD3MType generates a default postgres value from a d3m type.
func DefaultPostgresValueFromD3MType(typ string) interface{} {
	switch typ {
	case model.IndexType:
		return float64(0)
	case model.LongitudeType, model.LatitudeType, model.RealType:
		return "'NaN'::double precision"
	case model.IntegerType, model.TimestampType:
		return int(0)
	case model.DateTimeType:
		return fmt.Sprintf("'%s'", time.Time{}.Format(dateFormat))
	case model.GeoBoundsType:
		return "'POLYGON EMPTY'"
	case model.RealVectorType, model.RealListType:
		return "'{}'"
	default:
		return "''"
	}
}

// IsDatabaseFloatingPoint indicates whether or not a database type is a floating point
// value.
func IsDatabaseFloatingPoint(typ string) bool {
	return typ == dataTypeFloat
}

// ValueForFieldType generates the select field value for a given variable type.
func ValueForFieldType(typ string, field string) string {
	fieldQuote := fmt.Sprintf("\"%s\"", field)
	switch typ {
	case model.RealListType:
		return fmt.Sprintf("string_to_array(%s, ',')", fieldQuote)
	case model.DateTimeType:
		// datetime may be only time so need to support both cases
		// times can have first value missing a 0 so want to first get a time value then add it to epoch time 0
		return fmt.Sprintf("CASE WHEN length(%[1]s) IN (4, 5) AND position(':' in %[1]s) > 0 THEN CONCAT('1970-01-01 ', to_char(to_timestamp(%[1]s, 'MI:SS'), 'HH24:MI:SS')) ELSE %[1]s END", fieldQuote)
	default:
		return fmt.Sprintf("\"%s\"", field)
	}
}

// IsValidType validates the string to make sure it is a valid supported type
func IsValidType(pType string) bool {
	switch pType {
	case dataTypeText:
		return true
	case dataTypeDouble:
		return true
	case dataTypeFloat:
		return true
	case dataTypeVector:
		return true
	case dataTypeGeometry:
		return true
	case dataTypeInteger:
		return true
	case dataTypeDate:
		return true
	case dataTypeLat:
		return true
	case dataTypeLon:
		return true
	case dataTypeImageExt:
		return true
	case dataTypeBool:
		return true
	case dataTypeEmail:
		return true
	default:
		return false
	}
}

// GetValidTypes returns all of the supported types in the DB
func GetValidTypes() []string {
	return []string{dataTypeText,
		dataTypeDouble,
		dataTypeFloat,
		dataTypeVector,
		dataTypeGeometry,
		dataTypeInteger,
		dataTypeDate,
		dataTypeBool,
		dataTypeEmail,
		dataTypeLat,
		dataTypeLon,
		dataTypeImageExt,
		dataTypeCoord}
}

// GetIndexStatement returns the index SQL statement for a field of the provided type.
func GetIndexStatement(tableName string, fieldName string, typ string) string {
	typeSQL := ""
	switch typ {
	case model.GeoBoundsType:
		typeSQL = fmt.Sprintf("USING GIST(\"%s\")", fieldName)
	default:
		typeSQL = fmt.Sprintf("(\"%s\")", fieldName)
	}

	return fmt.Sprintf("CREATE INDEX \"%s_%s\" ON \"%s\" %s;", tableName, fieldName, tableName, typeSQL)
}

// CreateIndex will create an index in postgres on the specified field.
func (d *Database) CreateIndex(tableName string, fieldName string, typ string) error {
	sql := GetIndexStatement(tableName, fieldName, typ)

	_, err := d.Client.Exec(sql)
	if err != nil {
		return errors.Wrapf(err, "unable to create postgres index")
	}

	return nil
}

// IsColumnType can be use to check columns potential types
func IsColumnType(client DatabaseDriver, tableName string, variable *model.Variable, colType string) bool {
	// check colType is valid
	if !IsValidType(colType) {
		return false
	}
	viewSelect := fmt.Sprintf("\"%s\"", variable.Key)
	groupBy := ""
	if colType == dataTypeCoord {
		viewSelect = fmt.Sprintf("concat('{', %s::%s, '}')", viewSelect, dataTypeVector)
		groupBy = fmt.Sprintf("GROUP BY array_length(\"%s\", 1)", variable.Key)
	} else {
		viewSelect = fmt.Sprintf("%s::%s", viewSelect, colType)
	}
	// generate view query
	viewQuery := fmt.Sprintf("CREATE TEMPORARY VIEW temp_view_%[1]s AS SELECT %[3]s AS %[1]s FROM %[2]s", variable.Key, tableName, viewSelect)
	// test query
	testQuery := fmt.Sprintf("SELECT COUNT(\"%[1]s\") FROM temp_view_%[1]s %[2]s", variable.Key, groupBy)

	// create transaction
	tx, err := client.Begin()
	if err != nil {
		if rbErr := tx.Rollback(context.Background()); rbErr != nil {
			log.Error("rollback failed")
		}
		return false
	}
	defer func() {
		if rbErr := tx.Rollback(context.Background()); rbErr != nil {
			log.Error("rollback failed")
		}
	}()
	// create temp view
	_, err = tx.Exec(context.Background(), viewQuery)
	if err != nil {
		return false
	}
	// test to see if the data can fit into the type
	_, err = tx.Exec(context.Background(), testQuery)
	return err == nil
}
