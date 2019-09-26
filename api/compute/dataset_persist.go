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

	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil-compute/primitive/compute"
	"github.com/uncharted-distil/distil-ingest/metadata"
	"github.com/uncharted-distil/distil/api/env"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/util"
	log "github.com/unchartedsoftware/plog"
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

func splitTrainTest(sourceFile string, trainFile string, testFile string, hasHeader bool, targetCol int, maxTrainingCount int, stratify bool) error {
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

	if stratify {
		// For statification we use a proportionate allocation strategy, dividing the dataset up into
		// subsets by category, and then sampling the subsets using the supplied train/test ratio.

		// first pass - create subsets by category
		categoryRowData := map[string][][]string{}
		for _, row := range rowData {
			key := row[targetCol]
			if _, ok := categoryRowData[key]; !ok {
				categoryRowData[key] = [][]string{}
			}
			categoryRowData[key] = append(categoryRowData[key], row)
		}

		// second pass - randomly sample each category to generate train/test split
		for _, data := range categoryRowData {
			maxCategoryRows := int(float64(len(data)) / float64(len(rowData)) * float64(maxTrainingCount))
			err := shuffleAndWrite(data, targetCol, maxCategoryRows, writerTrain, writerTest)
			if err != nil {
				return err
			}
		}
	} else {
		// randomly select from entire dataset
		err := shuffleAndWrite(rowData, targetCol, maxTrainingCount, writerTrain, writerTest)
		if err != nil {
			return err
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

func shuffleAndWrite(rowData [][]string, targetCol int, maxTrainingCount int, writerTrain *csv.Writer, writerTest *csv.Writer) error {
	// shuffle array
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(rowData), func(i, j int) { rowData[i], rowData[j] = rowData[j], rowData[i] })

	// Figure out the number of train and test rows to use capping on the limit supplied by the caller.
	numTest := int(math.Ceil((float64(len(rowData)) * (1.0 - trainTestSplitThreshold))))
	numTrain := min(maxTrainingCount, int(math.Floor(float64(len(rowData))*trainTestSplitThreshold)))
	if maxTrainingCount <= 0 {
		maxTrainingCount = math.MaxInt64
	}

	// Write out to train test
	testCount := 0
	trainCount := 0
	for _, row := range rowData {
		if row[targetCol] != "" && testCount < numTest {
			testCount++
			err := writerTest.Write(row)
			if err != nil {
				return errors.Wrap(err, "unable to write data to train output")
			}
		} else if trainCount < numTrain {
			trainCount++
			err := writerTrain.Write(row)
			if err != nil {
				return errors.Wrap(err, "unable to write data to test output")
			}
		}

		// if we've hit our train and test targets, bail out
		if testCount >= numTest && trainCount >= numTrain {
			break
		}
	}
	return nil
}

// check if the data has already been split using the existing context
type persistedDataParams struct {
	DatasetName          string
	SchemaFile           string
	SourceDataFolder     string
	TmpDataFolder        string
	TaskType             string
	TimeseriesFieldIndex int
	TargetFieldIndex     int
	Stratify             bool
}

// persistOriginalData copies the original data and splits it into a train &
// test subset to be used as needed.
func persistOriginalData(params *persistedDataParams) (string, string, error) {

	splitDatasetName, err := generateSplitDatasetName(params)
	if err != nil {
		return "", "", err
	}

	// The complete data is copied into separate train & test folders.
	// The main data is then split randomly.
	trainFolder := path.Join(params.TmpDataFolder, splitDatasetName, trainFilenamePrefix)
	testFolder := path.Join(params.TmpDataFolder, splitDatasetName, testFilenamePrefix)
	trainSchemaFile := path.Join(trainFolder, params.SchemaFile)
	testSchemaFile := path.Join(testFolder, params.SchemaFile)

	log.Infof("checking folders `%s` & `%s` to see if the dataset has been previously split", trainFolder, testFolder)
	if fileExists(trainSchemaFile) && fileExists(testSchemaFile) {
		log.Infof("dataset '%s' already split", params.DatasetName)
		return trainSchemaFile, testSchemaFile, nil
	}

	// clean out existing data
	if pathExists(trainFolder) {
		err := os.RemoveAll(trainFolder)
		if err != nil {
			return "", "", errors.Wrap(err, "unable to remove train folder from previous split attempt")
		}
	}

	if pathExists(testFolder) {
		err := os.RemoveAll(testFolder)
		if err != nil {
			return "", "", errors.Wrap(err, "unable to remove test folder from previous split attempt")
		}
	}

	// copy the data over
	err = copy.Copy(params.SourceDataFolder, trainFolder)
	if err != nil {
		return "", "", errors.Wrap(err, "unable to copy dataset folder to train")
	}

	err = copy.Copy(params.SourceDataFolder, testFolder)
	if err != nil {
		return "", "", errors.Wrap(err, "unable to copy dataset folder to test")
	}

	// read the dataset document
	schemaFilename := path.Join(params.SourceDataFolder, params.SchemaFile)
	meta, err := metadata.LoadMetadataFromOriginalSchema(schemaFilename)
	if err != nil {
		return "", "", err
	}

	// determine where the d3m index would be
	mainDR := meta.GetMainDataResource()

	// split the source data into train & test
	var config env.Config
	dataPath := path.Join(params.SourceDataFolder, mainDR.ResPath)
	trainDataFile := path.Join(trainFolder, mainDR.ResPath)
	testDataFile := path.Join(testFolder, mainDR.ResPath)
	if params.TaskType == compute.TimeseriesForecastingTask {
		err = splitTrainTestTimeseries(dataPath, trainDataFile, testDataFile, true, params.TimeseriesFieldIndex)
	} else {
		config, err = env.LoadConfig()
		if err != nil {
			return "", "", errors.Wrap(err, "unable to load config")
		}
		err = splitTrainTest(dataPath, trainDataFile, testDataFile, true, params.TargetFieldIndex, config.MaxTrainingRows, params.Stratify)
	}
	if err != nil {
		return "", "", err
	}

	return trainSchemaFile, testSchemaFile, nil
}

func generateSplitDatasetName(params *persistedDataParams) (string, error) {
	// generate the hash from the params
	hash, err := hashstructure.Hash(params, nil)
	if err != nil {
		return "", errors.Wrap(err, "failed to generate persisted data hash")
	}
	hashFileName := fmt.Sprintf("%s-%0x", params.DatasetName, hash)
	return hashFileName, nil
}

func pathExists(path string) bool {
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
	return variable.Type == model.ImageType
}

func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}
