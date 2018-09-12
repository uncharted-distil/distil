package compute

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/mitchellh/hashstructure"
	"github.com/otiai10/copy"
	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil-ingest/metadata"
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

	trainFilenamePrefix = "train"
	testFilenamePrefix  = "test"
)

// FilteredDataProvider defines a function that will fetch data from a back end source given
// a set of filter parameters.
type FilteredDataProvider func(dataset string, index string, filters *model.FilterParams) (*model.FilteredData, error)

// VariablesProvider defines a function that will get the variables for a dataset.
type VariablesProvider func(dataset string, index string) ([]*model.Variable, error)

// DataSchema encapsulates the data schema json structure.
type DataSchema struct {
	About         *DataSchemaAbout `json:"about"`
	DataResources []*DataResource  `json:"dataResources"`
}

// DataSchemaAbout contains the basic properties of a dataset.
type DataSchemaAbout struct {
	DatasetID       string `json:"datasetID"`
	DatasetName     string `json:"datasetName"`
	Description     string `json:"description"`
	Citation        string `json:"citation"`
	License         string `json:"license"`
	Source          string `json:"source"`
	SourceURI       string `json:"sourceURI"`
	ApproximateSize string `json:"approximateSize"`
	Redacted        bool   `json:"redacted"`
	SchemaVersion   string `json:"datasetSchemaVersion"`
}

// DataResource represents a set of variables.
type DataResource struct {
	ResID        string          `json:"resID"`
	ResPath      string          `json:"resPath"`
	ResType      string          `json:"resType"`
	ResFormat    []string        `json:"resFormat"`
	IsCollection bool            `json:"isCollection"`
	Variables    []*DataVariable `json:"columns,omitempty"`
}

// DataVariable captures the data schema representation of a variable.
type DataVariable struct {
	ColName  string         `json:"colName"`
	Role     []string       `json:"role"`
	ColType  string         `json:"colType"`
	ColIndex int            `json:"colIndex"`
	RefersTo *DataReference `json:"refersTo,omitempty"`
}

// DataReference captures the data schema representation of a resource reference.
type DataReference struct {
	ResID     string `json:"resID"`
	ResObject string `json:"resObject"`
}

// Hash the filter set
func getFilteredDatasetHash(dataset string, target string, filterParams *model.FilterParams, isTrain bool) (uint64, error) {
	hash, err := hashstructure.Hash([]interface{}{dataset, target, *filterParams, isTrain}, nil)
	if err != nil {
		return 0, errors.Wrapf(err, "failed to generate hashcode for %s", dataset)
	}
	return hash, nil
}

func updateSchemaReferenceFile(schema string, prevReferenceFile string, newReferenceFile string) string {
	return strings.Replace(schema, fmt.Sprintf("\"resPath\": \"%s\"", prevReferenceFile), fmt.Sprintf("\"resPath\": \"%s\"", newReferenceFile), 1)
}

func splitTrainTest(sourceFile string, trainFile string, testFile string, hasHeader bool) error {
	// create the writers
	outputTrain := &bytes.Buffer{}
	writerTrain := csv.NewWriter(outputTrain)
	outputTest := &bytes.Buffer{}
	writerTest := csv.NewWriter(outputTest)

	// open the file
	file, err := os.Open(sourceFile)
	if err != nil {
		return errors.Wrap(err, "failed to open source file")
	}
	reader := csv.NewReader(file)

	// write header to both outputs
	if hasHeader {
		header, err := reader.Read()
		if err != nil {
			return errors.Wrap(err, "unable to read header row")
		}
		err = writerTrain.Write(header)
		if err != nil {
			return errors.Wrap(err, "unable to write header to train output")
		}
		err = writerTest.Write(header)
		if err != nil {
			return errors.Wrap(err, "unable to write header to test output")
		}
	}

	// randomly assign rows to either train or test
	for {
		line, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return errors.Wrap(err, "failed to read line from file")
		}
		if rand.Float64() < trainTestSplitThreshold {
			err = writerTrain.Write(line)
			if err != nil {
				return errors.Wrap(err, "unable to write data to train output")
			}
		} else {
			err = writerTest.Write(line)
			if err != nil {
				return errors.Wrap(err, "unable to write data to test output")
			}
		}
	}
	writerTrain.Flush()
	writerTest.Flush()

	err = ioutil.WriteFile(trainFile, outputTrain.Bytes(), 0644)
	if err != nil {
		return errors.Wrap(err, "unable to output train data")
	}

	err = ioutil.WriteFile(testFile, outputTest.Bytes(), 0644)
	if err != nil {
		return errors.Wrap(err, "unable to output test data")
	}

	return nil
}

// PersistOriginalData copies the original data and splits it into a train &
// test subset to be used as needed.
func PersistOriginalData(datasetName string, schemaFile string, sourceDataFolder string, tmpDataFolder string) (string, string, error) {
	// The complete data is copied into separate train & test folders.
	// The main data is then split randomly.
	trainFolder := path.Join(tmpDataFolder, datasetName, trainFilenamePrefix)
	testFolder := path.Join(tmpDataFolder, datasetName, testFilenamePrefix)
	trainSchemaFile := path.Join(trainFolder, schemaFile)
	testSchemaFile := path.Join(testFolder, schemaFile)

	// check if the data has already been split
	log.Infof("checking folders `%s` & `%s` to see if the dataset has been previously split", trainFolder, testFolder)
	if fileExists(trainSchemaFile) && fileExists(testSchemaFile) {
		log.Infof("dataset '%s' already split", datasetName)
		return trainSchemaFile, testSchemaFile, nil
	}

	if dirExists(trainFolder) {
		err := os.RemoveAll(trainFolder)
		if err != nil {
			return "", "", errors.Wrap(err, "unable to remove train folder from previous split attempt")
		}
	}

	if dirExists(testFolder) {
		err := os.RemoveAll(testFolder)
		if err != nil {
			return "", "", errors.Wrap(err, "unable to remove test folder from previous split attempt")
		}
	}

	// copy the data over
	err := copy.Copy(sourceDataFolder, trainFolder)
	if err != nil {
		return "", "", errors.Wrap(err, "unable to copy dataset folder to train")
	}

	err = copy.Copy(sourceDataFolder, testFolder)
	if err != nil {
		return "", "", errors.Wrap(err, "unable to copy dataset folder to test")
	}

	// read the dataset document
	schemaFilename := path.Join(sourceDataFolder, schemaFile)
	meta, err := metadata.LoadMetadataFromOriginalSchema(schemaFilename)
	if err != nil {
		return "", "", err
	}

	// determine where the d3m index would be
	mainDR := meta.GetMainDataResource()

	// split the source data into train & test
	dataPath := path.Join(sourceDataFolder, mainDR.ResPath)
	trainDataFile := path.Join(trainFolder, mainDR.ResPath)
	testDataFile := path.Join(testFolder, mainDR.ResPath)
	err = splitTrainTest(dataPath, trainDataFile, testDataFile, true)
	if err != nil {
		return "", "", err
	}

	return trainSchemaFile, testSchemaFile, nil
}

// PersistFilteredData creates a hash code from the combination of the dataset name, the target name, and its filter
// state, and saves the filtered data and target data to disk if they haven't been previously.  The path to the data
// is returned.
func PersistFilteredData(inputPath string, datasetDir string, target string, dataset *model.QueriedDataset, variables []*model.Variable) (string, int, error) {
	// parse the dataset and its filter state and generate a hashcode from both
	hash, err := getFilteredDatasetHash(dataset.Metadata.Name, target, dataset.Filters, dataset.IsTrain)
	if err != nil {
		return "", -1, err
	}

	// REMOVED CACHING FOR NOW DUE TO TRAIN / TEST SPLIT
	// check to see if we already have this filtered dataset saved - return the path
	// if so
	path := path.Join(datasetDir, strconv.FormatUint(hash, 10))
	if dirExists(path) {
		log.Infof("Found cached data with hash %d", hash)
		//	return path, -1, nil
		// delete existing data to overwrite with latest
		os.RemoveAll(path)
		log.Infof("Deleted data at %s", path)
	}

	// get the filtered dataset from elastic search
	start := time.Now()
	if len(dataset.Data.Values) <= 0 {
		log.Info("No data available for data after filter application")
		return "", -1, nil
	}

	// find the index of the target variable
	targetIdx := -1
	for _, variable := range variables {
		if variable.Key == target {
			targetIdx = variable.Index
			break
		}
	}
	if targetIdx < 0 {
		return "", -1, errors.Errorf("could not find target %s in filtered data", target)
	}

	// create the path for the data and target csvs
	if err := os.MkdirAll(path, 0777); err != nil && !os.IsExist(err) {
		return "", -1, errors.Wrapf(err, "unable to create dataset dir %s", datasetDir)
	}

	// create a var name lookup table
	variablesByKey := map[string]*model.Variable{}
	for _, variable := range variables {
		variablesByKey[variable.Key] = variable
	}

	// write the filtered data to csv file
	err = writeData(path, datasetDir, dataset.Data, variablesByKey)
	if err != nil {
		return "", -1, err
	}

	err = writeDataSchema(inputPath, path, dataset.Metadata.Name, targetIdx, variablesByKey)
	if err != nil {
		return "", -1, err
	}

	log.Infof("Persisted data to %s in %v", path, time.Since(start))
	return path, targetIdx, nil
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
		variableNames[i] = column.Key
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

func writeData(dataPath string, datasetDir string, filteredData *model.FilteredData, variables map[string]*model.Variable) error {
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

	// create a map of col idx to original idx
	columnToOriginal := make([]int, len(filteredData.Columns))
	for i, column := range filteredData.Columns {
		if columnVar, ok := variables[column.Key]; ok {
			columnToOriginal[i] = columnVar.Index
		} else {
			columnToOriginal[i] = -1
		}
	}

	// map the name to the display name
	variableNamesDisplay := make(map[string]string)
	for _, v := range variables {
		variableNamesDisplay[v.Key] = v.DisplayVariable
	}

	// write out the header, including the d3m_index field - we remap from col index
	// back to the original dataset index to enforce the original column ordering
	variableNames := make([]string, len(variables))
	for i, column := range filteredData.Columns {
		if columnToOriginal[i] >= 0 {
			variableNames[columnToOriginal[i]] = variableNamesDisplay[column.Key]
		}
	}
	err = writer.Write(variableNames)
	if err != nil {
		return errors.Wrapf(err, "unable to persist %v", variableNames)
	}

	for _, row := range filteredData.Values {
		strVals := make([]string, len(variables))

		// convert vals in row to string and reorder to reflect original column ordering
		for i, value := range row {
			if columnToOriginal[i] >= 0 {
				strVals[columnToOriginal[i]] = fmt.Sprintf("%v", value)
			}
		}
		err := writer.Write(strVals)
		if err != nil {
			log.Errorf("%v", errors.Wrapf(err, "unable to persist %v", strVals))
		}
	}
	return nil
}

func writeDataSchema(rootPath string, schemaPath string, dataset string, targetIdx int, variables map[string]*model.Variable) error {

	// Build the schema data for output.
	drs := make([]*DataResource, 1)
	drs[0] = &DataResource{
		ResID:        "0",
		ResPath:      path.Join(D3MDataFolder, D3MLearningData),
		ResType:      D3MResourceType,
		ResFormat:    []string{D3MResourceFormat},
		IsCollection: false,
		Variables:    make([]*DataVariable, len(variables)),
	}
	dsProperties := &DataSchemaAbout{
		DatasetID:     dataset,
		Redacted:      true,
		SchemaVersion: D3MDatasetSchemaVersion,
	}
	ds := &DataSchema{
		About:         dsProperties,
		DataResources: drs,
	}

	// NOTE: the target is identified by the suggested target role.
	for k, v := range variables {
		role := []string{"attribute"}
		if v.Index == targetIdx {
			role[0] = "suggestedTarget"
		} else if k == model.D3MIndexFieldName {
			// Set the specific values for the d3m index.
			role[0] = "index"
		}
		// Write out the original index and type for the variable - column removal and semantic type
		// updates are preprended to all generated pipelines, so we want the data we pass through
		// to be the original version (minus any filtered rows).
		// TODO: Metadata variables are always fetched regardless of filter state, so we do a check to
		// ignore them when persisting.
		dataVariable := &DataVariable{
			ColName:  v.DisplayVariable,
			Role:     role,
			ColType:  v.OriginalType,
			ColIndex: v.Index,
		}

		// if a resource is references (images), add it.
		// TODO: need a general solution for resource collections - image is only one type
		if referencesResource(v) {
			resID := fmt.Sprintf("%d", len(ds.DataResources))
			// manually append separator because path.Join cleans the path, which will remove it
			resPath := fmt.Sprintf("%s%c", path.Join(rootPath, "TRAIN", "dataset_TRAIN", "media"), os.PathSeparator)
			ds.DataResources = append(ds.DataResources, &DataResource{
				ResID:        resID,
				ResPath:      resPath,
				ResType:      v.Type,
				ResFormat:    []string{"image/jpeg"},
				IsCollection: true,
			})
			dataVariable.RefersTo = &DataReference{
				ResID:     resID,
				ResObject: "item",
			}
		}
		ds.DataResources[0].Variables[dataVariable.ColIndex] = dataVariable
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

// LoadDatasetSchemaFromFile loads the dataset schema from file.
func LoadDatasetSchemaFromFile(filename string) (*DataSchema, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	dataDoc := &DataSchema{}
	err = json.Unmarshal(b, dataDoc)
	if err != nil {
		return nil, err
	}
	return dataDoc, nil
}

func referencesResource(variable *model.Variable) bool {
	if variable.Type == model.ImageType {
		return true
	}

	return false
}
