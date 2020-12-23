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
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uncharted-distil/distil/api/util"
)

func TestSampleStratified(t *testing.T) {
	assert.NoError(t, removeTestFiles())
	params := createTestParams(false, ModelQualityHigh, "test_data")
	initializeTestConfig(t)
	expectedSamplePath := "test/tmp_data/sample-6f5d2a90aff56355/datasetDoc.json"

	sampleURI, err := SampleDataset(path.Join(params.SourceDataFolder, params.SchemaFile), "./test/tmp_data", 10, true, 2, -1)
	assert.NoError(t, err)
	assert.Equal(t, expectedSamplePath, sampleURI)

	lines, err := util.ReadCSVFile(path.Join(path.Dir(sampleURI), "tables", "learningData.csv"), true)
	assert.NoError(t, err)
	assert.Equal(t, 11, len(lines))
	assert.NoError(t, removeTestFiles())

	categoricalValues := map[string]int{}
	for _, rowData := range lines {
		categoricalValues[rowData[2]]++
	}
	assert.Equal(t, 6, categoricalValues["a"])
	assert.Equal(t, 3, categoricalValues["b"])
	assert.Equal(t, 1, categoricalValues["c"])
	assert.Equal(t, 1, categoricalValues["d"])
}

func TestSplitTimestampValue(t *testing.T) {
	assert.NoError(t, removeTestFiles())
	params := createTestParams(false, ModelQualityHigh, "test_data")
	splitter := createTestTimestampSplitter(100)
	initializeTestConfig(t)

	splitDatasetName, err := generateSplitDatasetName("test_data", path.Join("./test/test_dataset", "datasetDoc.json"), splitter)
	assert.NoError(t, err)
	trainPath := fmt.Sprintf("test/tmp_data/%s/train/datasetDoc.json", splitDatasetName)
	testPath := fmt.Sprintf("test/tmp_data/%s/test/datasetDoc.json", splitDatasetName)

	trainFolder, testFolder, err := SplitDataset(path.Join(params.SourceDataFolder, params.SchemaFile), splitter)
	assert.NoError(t, err)
	assert.Equal(t, trainPath, trainFolder)
	assert.Equal(t, testPath, testFolder)

	trainDataPath := fmt.Sprintf("test/tmp_data/%s/train/tables/learningData.csv", splitDatasetName)
	testDataPath := fmt.Sprintf("test/tmp_data/%s/test/tables/learningData.csv", splitDatasetName)
	lines, err := util.ReadCSVFile(trainDataPath, true)
	assert.NoError(t, err)
	assert.Equal(t, 15, len(lines))

	lines, err = util.ReadCSVFile(testDataPath, true)
	assert.NoError(t, err)
	assert.Equal(t, 18, len(lines))

	categoricalValues := map[string]int{}
	for _, rowData := range lines {
		categoricalValues[rowData[2]]++
	}
	assert.Equal(t, 10, categoricalValues["a"])
	assert.Equal(t, 5, categoricalValues["b"])
	assert.Equal(t, 2, categoricalValues["c"])
	assert.Equal(t, 1, categoricalValues["d"])
}

func createTestTimestampSplitter(timestampValue float64) datasetSplitter {
	return &timeseriesSplitter{
		timeseriesCol:       1,
		timestampValueSplit: timestampValue,
	}
}
