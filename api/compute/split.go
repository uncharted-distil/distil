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
	"math"
	"os"
	"path"
	"sort"

	"github.com/mitchellh/hashstructure"
	"github.com/pkg/errors"
	log "github.com/unchartedsoftware/plog"

	"github.com/uncharted-distil/distil-compute/metadata"
	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil-compute/primitive/compute"
	"github.com/uncharted-distil/distil/api/env"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/serialization"
	"github.com/uncharted-distil/distil/api/util"
)

var (
	splitterBasic      *basicSplitter
	splitterTimeseries *timeseriesSplitter
)

type datasetSplitter interface {
	split(data [][]string) ([][]string, [][]string, error)
	hash(schemaFile string) (uint64, error)
}

type timeseriesSplitter struct {
	timeseriesCol  int
	trainTestSplit float64
}

type basicSplitter struct {
	stratify       bool
	rowLimits      rowLimits
	targetCol      int
	groupingCol    int
	trainTestSplit float64
}

func (t *timeseriesSplitter) hash(schemaFile string) (uint64, error) {
	// generate the hash from the params
	hashStruct := struct {
		Schema         string
		TimeseriesCol  int
		TrainTestSplit float64
	}{
		Schema:         schemaFile,
		TimeseriesCol:  t.timeseriesCol,
		TrainTestSplit: t.trainTestSplit,
	}
	hash, err := hashstructure.Hash(hashStruct, nil)
	if err != nil {
		return 0, errors.Wrap(err, "failed to generate persisted data hash")
	}
	return hash, nil
}

func (t *timeseriesSplitter) split(data [][]string) ([][]string, [][]string, error) {
	// training data
	outputTrain := [][]string{}

	// test data
	outputTest := [][]string{}

	// handle the header
	inputData, outputTrain, outputTest := splitTrainTestHeader(data, outputTrain, outputTest, true)

	// find the desired timeseries threshold
	// load the parsed timestamp into a list and read all raw data in memory
	timestamps := make([]float64, 0)
	for _, line := range inputData {
		// attempt to parse as float
		t, err := parseTimeColValue(line[t.timeseriesCol])
		if err != nil {
			return nil, nil, err
		}
		timestamps = append(timestamps, t)
	}

	// find the time threshold by sorting and taking the value that gives
	// the right split (ie value where we would send roughly 90% of
	// the data to train and 10% to test)
	sort.Slice(timestamps, func(i int, j int) bool {
		return timestamps[i] <= timestamps[j]
	})
	timestampSplit := SplitTimeStamps(timestamps, t.trainTestSplit)

	// output the values based on if before threshold or after threshold
	for _, line := range inputData {
		// since we parsed it above, then the parsing here should succeed
		// TODO: the timestamps list is already sorted but we really should
		// reuse it to not double parse things
		t, _ := parseTimeColValue(line[t.timeseriesCol])

		// !After == Before || Equal
		if t <= timestampSplit.SplitValue {
			outputTrain = append(outputTrain, line)
		} else {
			outputTest = append(outputTest, line)
		}
	}

	return outputTrain, outputTest, nil
}

func (b *basicSplitter) hash(schemaFile string) (uint64, error) {
	// generate the hash from the params
	hashStruct := struct {
		Schema         string
		Stratify       bool
		RowLimits      rowLimits
		TargetCol      int
		GroupingCol    int
		TrainTestSplit float64
	}{
		Schema:         schemaFile,
		Stratify:       b.stratify,
		RowLimits:      b.rowLimits,
		TargetCol:      b.targetCol,
		GroupingCol:    b.groupingCol,
		TrainTestSplit: b.trainTestSplit,
	}
	hash, err := hashstructure.Hash(hashStruct, nil)
	if err != nil {
		return 0, errors.Wrap(err, "failed to generate persisted data hash")
	}
	return hash, nil
}

func (b *basicSplitter) split(data [][]string) ([][]string, [][]string, error) {
	// create the output
	outputTrain := [][]string{}
	outputTest := [][]string{}

	// handle the header
	inputData, outputTrain, outputTest := splitTrainTestHeader(data, outputTrain, outputTest, true)

	numDatasetRows := len(inputData)
	numTrainingRows := b.rowLimits.trainingRows(numDatasetRows)
	numTestRows := b.rowLimits.testRows(numDatasetRows)
	if b.stratify {
		// For statification we use a proportionate allocation strategy, dividing the dataset up into
		// subsets by category, and then sampling the subsets using the supplied train/test ratio.

		// first pass - create subsets by category
		categoryRowData := map[string][][]string{}
		for _, row := range inputData {
			key := row[b.targetCol]
			if _, ok := categoryRowData[key]; !ok {
				categoryRowData[key] = [][]string{}
			}
			categoryRowData[key] = append(categoryRowData[key], row)
		}

		// second pass - randomly sample each category to generate train/test split
		for _, data := range categoryRowData {
			maxCategoryTrainingRows := int(math.Max(1, float64(len(data))/float64(len(inputData))*float64(numTrainingRows)))
			maxCategoryTestRows := int(math.Max(1, float64(len(data))/float64(len(inputData))*float64(numTestRows)))
			outputTrain, outputTest = shuffleAndWrite(data, b.groupingCol, maxCategoryTrainingRows, maxCategoryTestRows, true, outputTrain, outputTest, b.trainTestSplit)
		}
	} else {
		// randomly select from dataset based on row limits
		outputTrain, outputTest = shuffleAndWrite(inputData, b.groupingCol, numTrainingRows, numTestRows, true, outputTrain, outputTest, b.trainTestSplit)
	}

	return outputTrain, outputTest, nil
}

// SplitDataset splits a dataset into train and test, using an approach to splitting
// suitable to the task performed.
func SplitDataset(schemaFile string, splitter datasetSplitter) (string, string, error) {
	// load the metadata to get the data resource
	meta, err := metadata.LoadMetadataFromOriginalSchema(schemaFile, true)
	if err != nil {
		return "", "", err
	}

	// check if already split
	splitDatasetName, err := generateSplitDatasetName(meta.Name, schemaFile, splitter)
	if err != nil {
		return "", "", err
	}
	trainFolder := path.Join(env.GetTmpPath(), splitDatasetName, trainFilenamePrefix)
	testFolder := path.Join(env.GetTmpPath(), splitDatasetName, testFilenamePrefix)
	trainSchemaFile := path.Join(trainFolder, compute.D3MDataSchema)
	testSchemaFile := path.Join(testFolder, compute.D3MDataSchema)

	if alreadySplit(meta.Name, trainSchemaFile, testSchemaFile) {
		return trainSchemaFile, testSchemaFile, nil
	}

	// delete existing folders
	err = deleteIfExists(trainFolder)
	if err != nil {
		return "", "", nil
	}
	err = deleteIfExists(testFolder)
	if err != nil {
		return "", "", nil
	}

	// load data to split
	data, err := loadData(path.Dir(schemaFile), meta)
	if err != nil {
		return "", "", err
	}

	// split the data
	trainData, testData, err := splitter.split(data)
	if err != nil {
		return "", "", err
	}

	// determine output location
	outputFilename := path.Base(meta.GetMainDataResource().ResPath)
	trainOutput := path.Join(trainFolder, compute.D3MDataFolder, outputFilename)
	testOutput := path.Join(testFolder, compute.D3MDataFolder, outputFilename)

	// output the train and test data
	mainDR := meta.GetMainDataResource()
	outputStore := serialization.GetStorage(mainDR.ResPath)
	mainDR.ResPath = trainOutput
	outputTrain := &api.RawDataset{
		ID:       meta.ID,
		Name:     meta.Name,
		Metadata: meta,
		Data:     trainData,
	}
	err = outputStore.WriteDataset(trainFolder, outputTrain)
	if err != nil {
		return "", "", err
	}

	mainDR.ResPath = testOutput
	outputTest := &api.RawDataset{
		ID:       meta.ID,
		Name:     meta.Name,
		Metadata: meta,
		Data:     testData,
	}
	err = outputStore.WriteDataset(testFolder, outputTest)
	if err != nil {
		return "", "", err
	}

	return trainSchemaFile, testSchemaFile, nil
}

func loadData(sourceFolder string, meta *model.Metadata) ([][]string, error) {
	// use the path to determine how to load the data
	mainDR := meta.GetMainDataResource()
	filename := model.GetResourcePathFromFolder(sourceFolder, mainDR)

	storage := serialization.GetStorage(filename)
	data, err := storage.ReadData(filename)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func createSplitter(taskType []string, targetFieldIndex int, groupingFieldIndex int, stratify bool, quality string, trainTestSplit float64) datasetSplitter {
	// build row limits
	config, _ := env.LoadConfig()
	limits := rowLimits{
		MinTrainingRows: config.MinTrainingRows,
		MinTestRows:     config.MinTestRows,
		MaxTrainingRows: config.MaxTrainingRows,
		MaxTestRows:     config.MaxTestRows,
		Sample:          config.FastDataPercentage,
		Quality:         quality,
	}

	for _, task := range taskType {
		if task == compute.ForecastingTask {
			return &timeseriesSplitter{
				timeseriesCol:  groupingFieldIndex,
				trainTestSplit: trainTestSplit,
			}
		}
	}

	return &basicSplitter{
		stratify:       stratify,
		rowLimits:      limits,
		targetCol:      targetFieldIndex,
		groupingCol:    groupingFieldIndex,
		trainTestSplit: trainTestSplit,
	}
}

func alreadySplit(name string, trainFilename string, testFilename string) bool {
	exists := false
	log.Infof("checking folders `%s` & `%s` to see if the dataset has been previously split", trainFilename, testFilename)
	if util.FileExists(trainFilename) && util.FileExists(testFilename) {
		log.Infof("dataset '%s' already split", name)
		exists = true
	}

	return exists
}

func deleteIfExists(folderName string) error {
	if pathExists(folderName) {
		err := os.RemoveAll(folderName)
		if err != nil {
			return errors.Wrapf(err, "unable to remove folder '%s' from previous split attempt", folderName)
		}
	}

	return nil
}
