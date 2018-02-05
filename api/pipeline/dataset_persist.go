package pipeline

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/mitchellh/hashstructure"
	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil/api/model"
	"github.com/unchartedsoftware/plog"
)

const (
	// D3MLearningData provides the name of the training csv file as defined in the D3M schema
	D3MLearningData = "learningData.csv"
	// D3MDataFolder provides the name of the directory containing the dataset
	D3MDataFolder = "tables"
	// D3MDataSchema provides the name of the D3M data schema file
	D3MDataSchema = "datasetDoc.json"
	// D3MDatasetSchemaVersion is the current version supported when persisting
	D3MDatasetSchemaVersion = "3.0"
	// D3MResourceType is the resource type of persisted datasets
	D3MResourceType = "table"
	// D3MResourceFormat is the resource format of persisted dataset
	D3MResourceFormat = "text/csv"
)

// FilteredDataProvider defines a function that will fetch data from a back end source given
// a set of filter parameters.
type FilteredDataProvider func(dataset string, index string, filters *model.FilterParams) (*model.FilteredData, error)

// VariablesProvider defines a function that will get the variables for a dataset.
type VariablesProvider func(dataset string, index string) ([]*model.Variable, error)

// DataSchema encapsulates the data schema json structure.
type DataSchema struct {
	Properties    *DataSchemaProperties `json:"about"`
	DataResources []*DataResource       `json:"dataResources"`
}

// DataSchemaProperties contains the basic properties of a dataset.
type DataSchemaProperties struct {
	DatasetID     string `json:"datasetID"`
	Redacted      bool   `json:"redacted"`
	SchemaVersion string `json:"datasetSchemaVersion"`
}

// DataResource represents a set of variables.
type DataResource struct {
	ResID        string          `json:"resID"`
	ResPath      string          `json:"resPath"`
	ResType      string          `json:"resType"`
	ResFormat    []string        `json:"resFormat"`
	IsCollection bool            `json:"isCollection"`
	Variables    []*DataVariable `json:"columns"`
}

// DataVariable captures the data schema representation of a variable.
type DataVariable struct {
	ColName  string   `json:"colName"`
	Role     []string `json:"role"`
	ColType  string   `json:"colType"`
	ColIndex int      `json:"colIndex"`
}

// Hash the filter set
func getFilteredDatasetHash(dataset string, target string, filterParams *model.FilterParams) (uint64, error) {
	hash, err := hashstructure.Hash([]interface{}{dataset, target, *filterParams}, nil)
	if err != nil {
		return 0, errors.Wrapf(err, "failed to generate hashcode for %s", dataset)
	}
	return hash, nil
}

// PersistFilteredData creates a hash code from the combination of the dataset name, the target name, and its filter
// state, and saves the filtered data and target data to disk if they haven't been previously.  The path to the data
// is returned.
func PersistFilteredData(fetchData FilteredDataProvider, fetchVariables VariablesProvider, datasetDir string, dataset string, index string, target string, filters *model.FilterParams) (string, error) {
	// parse the dataset and its filter state and generate a hashcode from both
	hash, err := getFilteredDatasetHash(dataset, target, filters)
	if err != nil {
		return "", err
	}

	// check to see if we already have this filtered dataset saved - return the path
	// if so
	path := path.Join(datasetDir, strconv.FormatUint(hash, 10))
	if dirExists(path) {
		log.Infof("Found cached data for %s with hash %d", dataset, hash)
		return path, nil
	}

	// get the filtered dataset from elastic search
	start := time.Now()
	filteredData, err := fetchData(dataset, index, filters)
	if err != nil {
		return "", err
	}
	if len(filteredData.Values) <= 0 {
		log.Infof("No data available for %s after filter application", dataset)
		return "", nil
	}

	// find the index of the target variable
	targetIdx := -1
	for idx, column := range filteredData.Columns {
		if column == target {
			targetIdx = idx
			break
		}
	}
	if targetIdx < 0 {
		return "", errors.Errorf("could not find target %s in filtered data", target)
	}

	// create the path for the data and target csvs
	if err := os.MkdirAll(path, 0777); err != nil && !os.IsExist(err) {
		return "", errors.Wrapf(err, "unable to create dataset dir %s", datasetDir)
	}

	// write the filtered data (minus the target field) to csv file
	err = writeData(path, datasetDir, filteredData, targetIdx)
	if err != nil {
		return "", err
	}

	// write the data schema
	variables, err := fetchVariables(dataset, index)
	if err != nil {
		return "", err
	}

	err = writeDataSchema(path, dataset, filteredData, targetIdx, variables)
	if err != nil {
		return "", err
	}

	log.Infof("Persisted data for %s to %s in %v", dataset, path, time.Since(start))
	return path, nil
}

// PersistData writes out the data to the specified file using a csv structure.
func PersistData(dataDir string, filename string, data *model.FilteredData) error {
	filenameFull := path.Join(dataDir, filename)
	file, err := os.Create(filenameFull)
	if err != nil {
		return errors.Wrapf(err, "unable to persist data to %s", filenameFull)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// write out the header, including the d3m_index field
	variableNames := make([]string, len(data.Columns))
	for i, column := range data.Columns {
		variableNames[i] = column
	}
	err = writer.Write(variableNames)
	if err != nil {
		return errors.Wrapf(err, "unable to persist %v", variableNames)
	}

	for _, row := range data.Values {
		strVals := make([]string, len(row))
		// convert vals in row to string
		for i, value := range row {
			strVals[i] = fmt.Sprintf("%v", value)
		}
		err := writer.Write(strVals)
		if err != nil {
			log.Errorf("%v", errors.Wrapf(err, "unable to persist %v", strVals))
		}
	}
	return nil
}

func dirExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func writeData(dataPath string, datasetDir string, filteredData *model.FilteredData, targetIdx int) error {
	// make sure the output folder exists
	dataFolder := path.Join(dataPath, D3MDataFolder)
	err := os.MkdirAll(dataFolder, os.ModePerm)
	if err != nil {
		return errors.Wrapf(err, "unable to create data folder for %s", datasetDir)
	}

	file, err := os.Create(path.Join(dataFolder, D3MLearningData))
	if err != nil {
		return errors.Wrapf(err, "unable to persist data to %s", datasetDir)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// write out the header, including the d3m_index field
	variableNames := []string{"d3mIndex"}
	for _, column := range filteredData.Columns {
		variableNames = append(variableNames, column)
	}
	err = writer.Write(variableNames)
	if err != nil {
		return errors.Wrapf(err, "unable to persist %v", variableNames)
	}

	for rowNum, row := range filteredData.Values {
		// append the index as the d3m_index col
		strVals := []string{strconv.Itoa(rowNum)}

		// convert vals in row to string
		for _, value := range row {
			strVals = append(strVals, fmt.Sprintf("%v", value))
		}
		err := writer.Write(strVals)
		if err != nil {
			log.Errorf("%v", errors.Wrapf(err, "unable to persist %v", strVals))
		}
	}
	return nil
}

func writeDataSchema(schemaPath string, dataset string, filteredData *model.FilteredData, targetIdx int, variables []*model.Variable) error {
	// Build a map of variable name to variable.
	vars := make(map[string]*model.Variable)
	for _, v := range variables {
		vars[v.Name] = v
	}

	// Build the schema data for output.
	drs := make([]*DataResource, 1)
	drs[0] = &DataResource{
		ResID:        "0",
		ResPath:      path.Join(D3MDataFolder, D3MLearningData),
		ResType:      D3MResourceType,
		ResFormat:    []string{D3MResourceFormat},
		IsCollection: false,
		Variables:    make([]*DataVariable, 0),
	}
	dsProperties := &DataSchemaProperties{
		DatasetID:     dataset,
		Redacted:      true,
		SchemaVersion: D3MDatasetSchemaVersion,
	}
	ds := &DataSchema{
		Properties:    dsProperties,
		DataResources: drs,
	}

	// Both outputs have the index.
	ds.DataResources[0].Variables = append(ds.DataResources[0].Variables, &DataVariable{
		ColName:  "d3mIndex",
		Role:     []string{"index"},
		ColType:  "integer",
		ColIndex: 0,
	})

	// Add all other variables.
	// NOTE: the target is identified by the suggested target role.
	for i, c := range filteredData.Columns {
		role := []string{"attribute"}
		if i == targetIdx {
			role[0] = "suggestedTarget"
		}
		v := &DataVariable{
			ColName:  c,
			Role:     role,
			ColType:  vars[c].Type,
			ColIndex: i + 1,
		}
		ds.DataResources[0].Variables = append(ds.DataResources[0].Variables, v)
	}

	dsJSON, err := json.Marshal(ds)
	if err != nil {
		return errors.Wrap(err, "Unable to marshal data schema")
	}

	err = ioutil.WriteFile(path.Join(schemaPath, D3MDataSchema), dsJSON, 0644)
	if err != nil {
		return errors.Wrap(err, "Unable to write data schema")
	}

	return nil
}
