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
	// D3MTrainTargets provides the name of the training targets csv file as defined in the D3M schema
	D3MTrainTargets = "trainTargets.csv"
	// D3MTrainData provides the name of the training targets csv file as defined in the D3M schema
	D3MTrainData = "trainData.csv"
	// D3MDataSchema provides the name of the D3M data schema file
	D3MDataSchema = "dataSchema.json"
)

// FilteredDataProvider defines a function that will fetch data from a back end source given
// a set of filter parameters.
type FilteredDataProvider func(dataset string, index string, filters *model.FilterParams) (*model.FilteredData, error)

// VariableProvider defines a function that will get the variables for a dataset.
type VariableProvider func(dataset string, index string) ([]*model.Variable, error)

// DataSchema encapsulates the data schema json structure.
type DataSchema struct {
	DatasetID                            string     `json:"datasetId"`
	RawData                              bool       `json:"rawData"`
	Redacted                             bool       `json:"redacted"`
	TestDataSchemaMirrorsTrainDataSchema bool       `json:"testDataSchemaMirrorsTrainDataSchema"`
	TrainData                            *TrainData `json:"trainData"`
}

// TrainData represents a set of training and target variables.
type TrainData struct {
	TrainData    []*DataVariable `json:"trainData"`
	TrainTargets []*DataVariable `json:"trainTargets"`
}

// DataVariable captures the data schema representation of a variable.
type DataVariable struct {
	VarName string `json:"varName"`
	VarRole string `json:"varRole"`
	VarType string `json:"varType"`
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
func PersistFilteredData(fetchData FilteredDataProvider, fetchVariables VariableProvider, datasetDir string, dataset string, index string, target string, filters *model.FilterParams) (string, error) {
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
	if err := os.MkdirAll(path, 0700); err != nil && !os.IsExist(err) {
		return "", errors.Wrapf(err, "unable to create dataset dir %s", datasetDir)
	}

	// write the filtered data (minus the target field) to csv file
	err = writeTrainData(path, datasetDir, filteredData, targetIdx)
	if err != nil {
		return "", err
	}

	// write the target data to csv file
	err = writeTrainTargets(path, datasetDir, filteredData, targetIdx)
	if err != nil {
		return "", err
	}

	// write the target data to csv file
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

func dirExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func writeTrainData(dataPath string, datasetDir string, filteredData *model.FilteredData, targetIdx int) error {
	file, err := os.Create(path.Join(dataPath, D3MTrainData))
	if err != nil {
		return errors.Wrapf(err, "unable to persist data to %s", datasetDir)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// write out the header, including the d3m_index field
	variableNames := []string{"d3mIndex"}
	for i, column := range filteredData.Columns {
		if i != targetIdx {
			variableNames = append(variableNames, column)
		}
	}
	err = writer.Write(variableNames)
	if err != nil {
		return errors.Wrapf(err, "unable to persist %v", variableNames)
	}

	for rowNum, row := range filteredData.Values {
		// append the index as the d3m_index col
		strVals := []string{strconv.Itoa(rowNum)}

		// convert vals in row to string, ignoring target feature
		for i, value := range row {
			if i != targetIdx {
				strVals = append(strVals, fmt.Sprintf("%v", value))
			}
		}
		err := writer.Write(strVals)
		if err != nil {
			log.Errorf("%v", errors.Wrapf(err, "unable to persist %v", strVals))
		}
	}
	return nil
}

func writeTrainTargets(targetPath string, datasetDir string, filteredData *model.FilteredData, targetIdx int) error {
	file, err := os.Create(path.Join(targetPath, D3MTrainTargets))
	if err != nil {
		return errors.Wrapf(err, "unable to persist data to %s", datasetDir)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// write out the variable names including the d3m_index
	variableNames := []string{"d3mIndex", filteredData.Columns[targetIdx]}
	err = writer.Write(variableNames)
	if err != nil {
		return errors.Wrapf(err, "unable to persist %v", variableNames)
	}

	for rowNum, row := range filteredData.Values {
		// append the index as the d3m_index value
		targetValue := row[targetIdx]
		strVals := []string{strconv.Itoa(rowNum), fmt.Sprintf("%v", targetValue)}
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
	ds := &DataSchema{
		DatasetID: dataset,
		RawData:   false,
		Redacted:  true,
		TestDataSchemaMirrorsTrainDataSchema: true,
		TrainData: &TrainData{
			TrainData:    make([]*DataVariable, 0),
			TrainTargets: make([]*DataVariable, 0),
		},
	}

	// Both outputs have the index.
	ds.TrainData.TrainData = append(ds.TrainData.TrainData, &DataVariable{
		VarName: "d3mIndex",
		VarRole: "index",
		VarType: "integer",
	})
	ds.TrainData.TrainTargets = append(ds.TrainData.TrainTargets, &DataVariable{
		VarName: "d3mIndex",
		VarRole: "index",
		VarType: "integer",
	})

	// Add all other variables.
	for i, c := range filteredData.Columns {
		v := &DataVariable{
			VarName: c,
			VarRole: "attribute",
			VarType: vars[c].Type,
		}

		if i == targetIdx {
			ds.TrainData.TrainTargets = append(ds.TrainData.TrainTargets, v)
		} else {
			ds.TrainData.TrainData = append(ds.TrainData.TrainData, v)
		}
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
