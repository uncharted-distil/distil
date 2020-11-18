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
	"fmt"
	"math"
	"math/rand"
	"os"
	"path"
	"sort"
	"strconv"
	"time"

	"github.com/araddon/dateparse"
	"github.com/mitchellh/hashstructure"
	"github.com/otiai10/copy"
	"github.com/pkg/errors"

	"github.com/uncharted-distil/distil-compute/metadata"
	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil-compute/primitive/compute"
	"github.com/uncharted-distil/distil/api/env"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/serialization"
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
	Digest          string `json:"digest"`
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

func splitTrainTestHeader(data [][]string, outputTrain [][]string, outputTest [][]string, hasHeader bool) ([][]string, [][]string, [][]string) {
	// write header to both outputs
	if hasHeader {
		header := data[0]
		data = data[1:]
		outputTrain = append(outputTrain, header)
		outputTest = append(outputTest, header)
	}

	return data, outputTrain, outputTest
}

func splitTrainTestTimeseries(sourceFile string, trainFile string, testFile string, hasHeader bool, timeseriesCol int, trainTestSplit float64) error {
	// create the output containers

	// training data
	outputTrain := [][]string{}

	// test data
	outputTest := [][]string{}

	// read the data
	datasetStorage := serialization.GetStorage(sourceFile)
	inputData, err := datasetStorage.ReadData(sourceFile)
	if err != nil {
		return errors.Wrap(err, "failed to open source file")
	}

	// handle the header
	inputData, outputTrain, outputTest = splitTrainTestHeader(inputData, outputTrain, outputTest, hasHeader)

	// find the desired timeseries threshold
	// load the parsed timestamp into a list and read all raw data in memory
	timestamps := make([]float64, 0)
	for _, line := range inputData {
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
	timestampSplit := SplitTimeStamps(timestamps, trainTestSplit)

	// output the values based on if before threshold or after threshold
	for _, line := range inputData {
		// since we parsed it above, then the parsing here should succeed
		// TODO: the timestamps list is already sorted but we really should
		// reuse it to not double parse things
		t, _ := parseTimeColValue(line[timeseriesCol])

		// !After == Before || Equal
		if t <= timestampSplit.SplitValue {
			outputTrain = append(outputTrain, line)
		} else {
			outputTest = append(outputTest, line)
		}
	}

	err = datasetStorage.WriteData(trainFile, outputTrain)
	if err != nil {
		return errors.Wrap(err, "unable to output train data")
	}

	err = datasetStorage.WriteData(testFile, outputTest)
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

func splitTrainTest(sourceFile string, trainFile string, testFile string, hasHeader bool, targetCol int, groupingCol int,
	stratify bool, rowLimits rowLimits, trainTestSplit float64) error {
	// create the output
	outputTrain := [][]string{}
	outputTest := [][]string{}

	// read the data
	datasetStorage := serialization.GetStorage(sourceFile)
	inputData, err := datasetStorage.ReadData(sourceFile)
	if err != nil {
		return errors.Wrap(err, "failed to read source file")
	}

	// handle the header
	inputData, outputTrain, outputTest = splitTrainTestHeader(inputData, outputTrain, outputTest, hasHeader)

	numDatasetRows := len(inputData)
	numTrainingRows := rowLimits.trainingRows(numDatasetRows)
	numTestRows := rowLimits.testRows(numDatasetRows)
	if stratify {
		// For statification we use a proportionate allocation strategy, dividing the dataset up into
		// subsets by category, and then sampling the subsets using the supplied train/test ratio.

		// first pass - create subsets by category
		categoryRowData := map[string][][]string{}
		for _, row := range inputData {
			key := row[targetCol]
			if _, ok := categoryRowData[key]; !ok {
				categoryRowData[key] = [][]string{}
			}
			categoryRowData[key] = append(categoryRowData[key], row)
		}

		// second pass - randomly sample each category to generate train/test split
		for _, data := range categoryRowData {
			maxCategoryTrainingRows := int(math.Max(1, float64(len(data))/float64(len(inputData))*float64(numTrainingRows)))
			maxCategoryTestRows := int(math.Max(1, float64(len(data))/float64(len(inputData))*float64(numTestRows)))
			outputTrain, outputTest = shuffleAndWrite(data, groupingCol, maxCategoryTrainingRows, maxCategoryTestRows, true, outputTrain, outputTest, trainTestSplit)
		}
	} else {
		// randomly select from dataset based on row limits
		outputTrain, outputTest = shuffleAndWrite(inputData, groupingCol, numTrainingRows, numTestRows, true, outputTrain, outputTest, trainTestSplit)
	}

	err = datasetStorage.WriteData(trainFile, outputTrain)
	if err != nil {
		return errors.Wrap(err, "unable to output train data")
	}

	err = datasetStorage.WriteData(testFile, outputTest)
	if err != nil {
		return errors.Wrap(err, "unable to output test data")
	}

	return nil
}

type shuffleTracker struct {
	output [][]string
	count  int
	max    int
}

func (s *shuffleTracker) lessThanMax() bool {
	return s.count < s.max
}

func shuffleAndWrite(rowData [][]string, groupCol int, maxTrainingCount int,
	maxTestCount int, adjustCount bool, outputTrain [][]string, outputTest [][]string,
	trainTestSplit float64) ([][]string, [][]string) {
	if maxTrainingCount <= 0 {
		maxTrainingCount = math.MaxInt64
	}

	// shuffle array
	rand.Seed(time.Now().UnixNano())

	// Figure out the number of train and test rows to use capping on the limit supplied by the caller.
	numTrain := maxTrainingCount
	numTest := maxTestCount
	if adjustCount {
		numTrain = min(maxTrainingCount, int(math.Floor(float64(len(rowData))*trainTestSplit)))
		numTest = min(maxTestCount, int(math.Ceil(float64(len(rowData))*(1.0-trainTestSplit))))
	}

	// structures for tracking test and train counts
	shuffleTest := &shuffleTracker{
		output: outputTest,
		count:  0,
		max:    numTest,
	}
	shuffleTrain := &shuffleTracker{
		output: outputTrain,
		count:  0,
		max:    numTrain,
	}

	if groupCol < 0 {
		// Shuffle the list of unique group keys to randomize their order.
		rand.Shuffle(len(rowData), func(i, j int) { rowData[i], rowData[j] = rowData[j], rowData[i] })

		// write out training data until we we reach the max training count, then write out the
		// test data
		tracker := shuffleTest
		for _, data := range rowData {
			// could want to write 0 rows of data (ex: sampling)
			if !tracker.lessThanMax() {
				if tracker == shuffleTest {
					tracker = shuffleTrain
				} else {
					break
				}
			}
			tracker.output = append(tracker.output, data)
			tracker.count++

		}
	} else {
		groupData := map[string][][]string{}
		groupKeys := []string{}

		// Break the rows out by group key, and build a list of the unique group keys for later use
		for _, row := range rowData {
			groupKey := row[groupCol]
			if _, ok := groupData[groupKey]; !ok {
				groupData[groupKey] = [][]string{}
				groupKeys = append(groupKeys, groupKey)
			}
			groupData[groupKey] = append(groupData[groupKey], row)
		}

		// Shuffle the list of unique group keys to randomize their order.
		rand.Shuffle(len(groupKeys), func(i, j int) { groupKeys[i], groupKeys[j] = groupKeys[j], groupKeys[i] })

		// Iterate over the randomized list of group keys, looking up the associated rows for each.  Write out
		// the train rows, then the test rows.
		tracker := shuffleTest
		for _, groupKey := range groupKeys {
			for _, row := range groupData[groupKey] {
				tracker.output = append(tracker.output, row)
				tracker.count++
			}
			if !tracker.lessThanMax() {
				if tracker == shuffleTest {
					tracker = shuffleTrain
				} else {
					break
				}
			}
		}
	}
	return shuffleTrain.output, shuffleTest.output
}

// check if the data has already been split using the existing context
type persistedDataParams struct {
	DatasetName        string
	GroupingFieldIndex int
	SchemaFile         string
	SourceDataFolder   string
	Stratify           bool
	TargetFieldIndex   int
	TaskType           []string
	TmpDataFolder      string
	Quality            string
	TrainTestSplit     float64
}

type rowLimits struct {
	MinTrainingRows int
	MinTestRows     int
	MaxTrainingRows int
	MaxTestRows     int
	Sample          float64
	Quality         string
}

func (s rowLimits) trainingRows(rows int) int {
	// determine whether or not we need to sample
	sampledRows := rows
	if s.Quality == ModelQualityFast {
		sampledRows = int(float64(rows) * s.Sample)
	} else if s.Quality != ModelQualityHigh {
		log.Warnf("model quality '%s' unsupported - defaulting to %s", s.Quality, ModelQualityHigh)
	}

	// limit samples by configured bounds
	if sampledRows < s.MinTrainingRows {
		sampledRows = s.MinTrainingRows
	} else if sampledRows > s.MaxTrainingRows {
		sampledRows = s.MaxTrainingRows
	}
	return sampledRows
}

func (s rowLimits) testRows(rows int) int {
	// determine whether or not we need to sample
	sampledRows := rows
	if s.Quality == ModelQualityFast {
		sampledRows = int(float64(rows) * s.Sample)
	} else if s.Quality != ModelQualityHigh {
		log.Warnf("model quality '%s' unsupported - defaulting to '%s'", s.Quality, ModelQualityHigh)
	}

	// limit samples by configured bounds
	if sampledRows < s.MinTestRows {
		sampledRows = s.MinTestRows
	} else if sampledRows > s.MaxTestRows {
		sampledRows = s.MaxTestRows
	}
	return sampledRows
}

// persistOriginalData copies the original data and splits it into a train &
// test subset to be used as needed.
func persistOriginalData(params *persistedDataParams) (string, string, error) {
	var config env.Config
	config, err := env.LoadConfig()
	if err != nil {
		return "", "", errors.Wrap(err, "unable to load config")
	}

	// configure a struct to help calculate row limits based on quality during the train / test split
	limits := rowLimits{
		MinTrainingRows: config.MinTrainingRows,
		MinTestRows:     config.MinTestRows,
		MaxTrainingRows: config.MaxTrainingRows,
		MaxTestRows:     config.MaxTestRows,
		Sample:          config.FastDataPercentage,
		Quality:         params.Quality,
	}

	// generate a unique dataset name from the params and the limit values
	splitDatasetName, err := generateSplitDatasetName(params.DatasetName, params.SourceDataFolder, &basicSplitter{
		stratify:       params.Stratify,
		rowLimits:      limits,
		targetCol:      params.TargetFieldIndex,
		groupingCol:    params.GroupingFieldIndex,
		trainTestSplit: params.TrainTestSplit,
	})
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
	if util.FileExists(trainSchemaFile) && util.FileExists(testSchemaFile) {
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
	meta, err := metadata.LoadMetadataFromOriginalSchema(schemaFilename, true)
	if err != nil {
		return "", "", err
	}

	// determine where the d3m index would be
	mainDR := meta.GetMainDataResource()

	// split the source data into train & test
	dataPath := model.GetResourcePath(schemaFilename, mainDR)

	// set the data file paths
	// paths in main data resource can be absolute OR relative
	trainDataFile := path.Join(trainFolder, path.Base(path.Dir(mainDR.ResPath)), path.Base(mainDR.ResPath))
	testDataFile := path.Join(testFolder, path.Base(path.Dir(mainDR.ResPath)), path.Base(mainDR.ResPath))

	// Check to see if the task keyword list contains forecasting
	hasForecasting := false
	for _, task := range params.TaskType {
		if task == compute.ForecastingTask {
			hasForecasting = true
			break
		}
	}

	var trainTestSplit float64

	if params.TrainTestSplit != 0 {
		trainTestSplit = params.TrainTestSplit
	} else {
		if hasForecasting {
			trainTestSplit = config.TrainTestSplitTimeSeries
		} else {
			trainTestSplit = config.TrainTestSplit
		}
	}

	if hasForecasting {
		err = splitTrainTestTimeseries(dataPath, trainDataFile, testDataFile, true,
			params.GroupingFieldIndex, trainTestSplit)
	} else {
		err = splitTrainTest(dataPath, trainDataFile, testDataFile, true, params.TargetFieldIndex,
			params.GroupingFieldIndex, params.Stratify, limits, trainTestSplit)
	}
	if err != nil {
		return "", "", err
	}

	// update metadata to point to correct data for train & test
	dataSerialization := serialization.GetStorage(trainDataFile)
	mainDR.ResPath = trainDataFile
	err = dataSerialization.WriteMetadata(trainSchemaFile, meta, true, true)
	if err != nil {
		return "", "", err
	}
	mainDR.ResPath = testDataFile
	err = dataSerialization.WriteMetadata(testSchemaFile, meta, true, true)
	if err != nil {
		return "", "", err
	}

	return trainSchemaFile, testSchemaFile, nil
}

func generateSplitDatasetName(datasetName string, schemaFilename string, splitter datasetSplitter) (string, error) {
	hash, err := splitter.hash(schemaFilename)
	if err != nil {
		return "", errors.Wrap(err, "failed to generate persisted data hash")
	}
	hashFileName := fmt.Sprintf("%s-%0x", datasetName, hash)
	return hashFileName, nil
}

func pathExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

// TimeStampSplit defines a train/test split in a timeseries based on time values.
type TimeStampSplit struct {
	StartValue float64
	SplitValue float64
	EndValue   float64
}

// SplitTimeStamps splits a set of time stamps such that `trainPercentage` *data points* are less than or equal
// to the split value, and the remaining data points are greater than the split value.  The timestamps are assumed
// to be ordered.
func SplitTimeStamps(timestamps []float64, trainPercentage float64) TimeStampSplit {
	splitIndex := int(trainPercentage * float64(len(timestamps)-1))
	return TimeStampSplit{
		StartValue: timestamps[0],
		SplitValue: timestamps[splitIndex],
		EndValue:   timestamps[len(timestamps)-1],
	}
}

// SplitTimeSeries splits a set of (timestamps, value) tuples such that `trainPercentage` *data points* are less than or equal
// to the split value, and the remaining data points are greater than the split value.  The timestamps are assumed
// to be ordered.
func SplitTimeSeries(timeseries []*api.TimeseriesObservation, trainPercentage float64) TimeStampSplit {
	timestamps := make([]float64, len(timeseries))
	for i, v := range timeseries {
		timestamps[i] = v.Time
	}
	return SplitTimeStamps(timestamps, trainPercentage)
}

// SampleDataset shuffles a dataset's rows and takes a subsample, returning
// the raw byte data of the sampled dataset.
func SampleDataset(rawData [][]string, maxRows int, hasHeader bool) [][]string {
	// initialize csv writer
	output := [][]string{}

	if hasHeader {
		output = append(output, rawData[0])
		rawData = rawData[1:]
	}

	trainTestSplit := float64(1) // keep all rows in train
	output, _ = shuffleAndWrite(rawData, -1, maxRows, 0, false, output, nil, trainTestSplit)

	return output
}

// CreateBatches splits the dataset into batches of at most maxBatchSize rows,
// returning paths to the schema files for all resulting batches.
func CreateBatches(schemaFile string, maxBatchSize int) ([]string, error) {
	log.Infof("splitting dataset '%s' into batches of %d rows", schemaFile, maxBatchSize)
	// load the metadata of the source dataset
	meta, err := metadata.LoadMetadataFromOriginalSchema(schemaFile, false)
	if err != nil {
		return nil, err
	}
	rootPath := env.ResolvePath(metadata.Batch, meta.ID)

	// load the main data
	dataStorage := serialization.GetStorage(meta.GetMainDataResource().ResPath)
	data, err := dataStorage.ReadData(meta.GetMainDataResource().ResPath)
	if err != nil {
		return nil, err
	}

	// get grouping column, defaulting to d3m index
	groupColIndex := -1
	for _, v := range meta.GetMainDataResource().Variables {
		if v.DistilRole == model.VarDistilRoleGrouping {
			groupColIndex = v.Index
		} else if v.Name == model.D3MIndexFieldName && groupColIndex == -1 {
			groupColIndex = v.Index
		}
	}

	// group the data to prepare for batching
	groups := map[string][][]string{}
	for _, row := range data[1:] {
		key := row[groupColIndex]
		if groups[key] == nil {
			groups[key] = [][]string{}
		}
		groups[key] = append(groups[key], row)
	}

	// iterate over the groups to build batches, allowing for overflow to put a group together
	batch := [][]string{data[0]}
	outputURIs := []string{}
	for _, g := range groups {
		batch = append(batch, g...)
		if len(batch) >= maxBatchSize {
			batchURI, err := processBatch(batch, rootPath, len(outputURIs)+1, meta, dataStorage)
			if err != nil {
				return nil, err
			}
			outputURIs = append(outputURIs, batchURI)
			batch = [][]string{data[0]}
		}
	}

	if len(batch) > 0 {
		batchURI, err := processBatch(batch, rootPath, len(outputURIs)+1, meta, dataStorage)
		if err != nil {
			return nil, err
		}
		outputURIs = append(outputURIs, batchURI)
	}

	log.Infof("dataset '%s' split into %d batches", schemaFile, len(outputURIs))

	return outputURIs, nil
}

func processBatch(batch [][]string, rootPath string, batchIndex int,
	meta *model.Metadata, dataStorage serialization.Storage) (string, error) {
	batchURI := path.Join(rootPath, fmt.Sprintf("batch-%d", batchIndex))
	batchDataURI := path.Join(batchURI, compute.D3MDataFolder, compute.D3MLearningData)
	err := dataStorage.WriteData(batchDataURI, batch)
	if err != nil {
		return "", err
	}

	// write out the metadata as well (updating the dataset id to reflect the batch nature)
	meta.GetMainDataResource().ResPath = batchDataURI
	meta.ID = fmt.Sprintf("%s-batch-%d", meta.ID, batchIndex)
	err = dataStorage.WriteMetadata(path.Join(batchURI, compute.D3MDataSchema), meta, true, false)
	if err != nil {
		return "", err
	}

	return batchURI, nil
}
