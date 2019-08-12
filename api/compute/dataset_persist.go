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

package compute

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/araddon/dateparse"
	"github.com/mitchellh/hashstructure"
	"github.com/otiai10/copy"
	"github.com/pkg/errors"
	log "github.com/unchartedsoftware/plog"

	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil-compute/primitive/compute"
	"github.com/uncharted-distil/distil-ingest/metadata"
	"github.com/uncharted-distil/distil/api/env"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/util"
)

const (
	trainFilenamePrefix = "train"
	testFilenamePrefix  = "test"
)

// FilteredDataProvider defines a function that will fetch data from a back end source given
// a set of filter parameters.
type FilteredDataProvider func(dataset string, index string, filters *api.FilterParams) (*api.FilteredData, error)

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
func getFilteredDatasetHash(dataset string, target string, filterParams *api.FilterParams, isTrain bool) (uint64, error) {
	hash, err := hashstructure.Hash([]interface{}{dataset, target, *filterParams, isTrain}, nil)
	if err != nil {
		return 0, errors.Wrapf(err, "failed to generate hashcode for %s", dataset)
	}
	return hash, nil
}

func updateSchemaReferenceFile(schema string, prevReferenceFile string, newReferenceFile string) string {
	return strings.Replace(schema, fmt.Sprintf("\"resPath\": \"%s\"", prevReferenceFile), fmt.Sprintf("\"resPath\": \"%s\"", newReferenceFile), 1)
}

func splitTrainTestHeader(reader *csv.Reader, writerTrain *csv.Writer, writerTest *csv.Writer, hasHeader bool) error {
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

	return nil
}

func splitTrainTestTimeseries(sourceFile string, trainFile string, testFile string, hasHeader bool, timeseriesCol int) error {
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

	// handle the header
	err = splitTrainTestHeader(reader, writerTrain, writerTest, hasHeader)
	if err != nil {
		return errors.Wrap(err, "failed to open source file")
	}

	// find the desired timeseries threshold
	// load the parsed timestamp into a list and read all raw data in memory
	timestamps := make([]float64, 0)
	data := make([][]string, 0)
	for {
		line, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return errors.Wrap(err, "failed to read line from file")
		}
		data = append(data, line)
		// attempt to parse as float
		t, err := parseTimeColValue(line[timeseriesCol])
		if err != nil {
			return err
		}
		timestamps = append(timestamps, t)
	}

	// find the time threshold by sorting and taking the value that gives
	// the right split (ie value where we would send roughly 90% of
	// the data to train and 10% to test)
	sort.Slice(timestamps, func(i int, j int) bool {
		return timestamps[i] <= timestamps[j]
	})
	thresholdIndex := int(trainTestSplitThreshold * float64(len(timestamps)-1))
	threshold := timestamps[thresholdIndex]

	// output the values based on if before threshold or after threshold
	for _, line := range data {
		// since we parsed it above, then the parsing here should succeed
		// TODO: the timestamps list is already sorted but we really should
		// reuse it to not double parse things
		t, _ := parseTimeColValue(line[timeseriesCol])

		// !After == Before || Equal
		if t <= threshold {
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

	err = util.WriteFileWithDirs(trainFile, outputTrain.Bytes(), os.ModePerm)
	if err != nil {
		return errors.Wrap(err, "unable to output train data")
	}

	err = util.WriteFileWithDirs(testFile, outputTest.Bytes(), os.ModePerm)
	if err != nil {
		return errors.Wrap(err, "unable to output test data")
	}

	return nil
}

func parseTimeColValue(timeColValue string) (float64, error) {
	f, err := strconv.ParseFloat(timeColValue, 64)
	if err != nil {
		t, err := dateparse.ParseAny(timeColValue)
		if err != nil {
			return math.NaN(), errors.Wrapf(err, "failed to parse timeseries column val [%s]", timeColValue)
		}
		return float64(t.Unix()), nil
	}
	return f, nil
}

func splitTrainTest(sourceFile string, trainFile string, testFile string, hasHeader bool, targetCol int, maxTrainingCount int) error {
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

	// handle the header
	err = splitTrainTestHeader(reader, writerTrain, writerTest, hasHeader)
	if err != nil {
		return errors.Wrap(err, "failed to open source file")
	}

	// load train test
	rowData := [][]string{}
	for {
		line, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return errors.Wrap(err, "failed to read line from file")
		}
		rowData = append(rowData, line)
	}

	// shuffle array
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(rowData), func(i, j int) { rowData[i], rowData[j] = rowData[j], rowData[i] })

	// figure out the numer of train and test rows to use - training rows are capped to avoid excessive
	// fit times for users, although it results in poorer model fidelity
	numTest := int(float32(len(rowData)) * (1.0 - trainTestSplitThreshold))
	if maxTrainingCount <= 0 {
		maxTrainingCount = math.MaxInt64
	}
	numTrain := min(maxTrainingCount, int(float32(len(rowData))*trainTestSplitThreshold))

	// Write out to train test
	testCount := 0
	trainCount := 0
	for _, row := range rowData {
		if row[targetCol] != "" && testCount < numTest {
			testCount++
			err = writerTest.Write(row)
			if err != nil {
				return errors.Wrap(err, "unable to write data to train output")
			}
		} else if trainCount < numTrain {
			trainCount++
			err = writerTrain.Write(row)
			if err != nil {
				return errors.Wrap(err, "unable to write data to test output")
			}
		}

		// if we've hit our train and test targets, bail out
		if testCount >= numTest && trainCount >= numTrain {
			break
		}
	}

	writerTrain.Flush()
	writerTest.Flush()

	err = util.WriteFileWithDirs(trainFile, outputTrain.Bytes(), os.ModePerm)
	if err != nil {
		return errors.Wrap(err, "unable to output train data")
	}

	err = util.WriteFileWithDirs(testFile, outputTest.Bytes(), os.ModePerm)
	if err != nil {
		return errors.Wrap(err, "unable to output test data")
	}

	return nil
}

// PersistOriginalData copies the original data and splits it into a train &
// test subset to be used as needed.
func PersistOriginalData(datasetName string, schemaFile string, sourceDataFolder string, tmpDataFolder string, taskType string,
	timeseriesFieldIndex int, targetFieldIndex int) (string, string, error) {
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
	var config env.Config
	dataPath := path.Join(sourceDataFolder, mainDR.ResPath)
	trainDataFile := path.Join(trainFolder, mainDR.ResPath)
	testDataFile := path.Join(testFolder, mainDR.ResPath)
	if taskType == compute.TaskTypeTimeseries {
		err = splitTrainTestTimeseries(dataPath, trainDataFile, testDataFile, true, timeseriesFieldIndex)
	} else {
		config, err = env.LoadConfig()
		if err != nil {
			return "", "", errors.Wrap(err, "unable to load config")
		}
		err = splitTrainTest(dataPath, trainDataFile, testDataFile, true, targetFieldIndex, config.MaxTrainingRows)
	}
	if err != nil {
		return "", "", err
	}

	return trainSchemaFile, testSchemaFile, nil
}

func dirExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
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

func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}
