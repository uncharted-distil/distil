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
	"os"
	"path/filepath"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/uncharted-distil/distil-compute/primitive/compute"
	"github.com/uncharted-distil/distil/api/env"
	"github.com/uncharted-distil/distil/api/serialization"
	"github.com/uncharted-distil/distil/api/util"
)

func TestPersistOriginalDataUnstratified(t *testing.T) {
	assert.NoError(t, removeTestFiles())
	params := createTestParams(false, ModelQualityHigh)
	limits, err := createLimits(ModelQualityHigh)
	assert.NoError(t, err)

	splitDatasetName, err := generateSplitDatasetName(params, limits)
	assert.NoError(t, err)
	trainPath := fmt.Sprintf("test/tmp_data/%s/train/datasetDoc.json", splitDatasetName)
	testPath := fmt.Sprintf("test/tmp_data/%s/test/datasetDoc.json", splitDatasetName)

	trainFolder, testFolder, err := persistOriginalData(params)
	assert.NoError(t, err)
	assert.Equal(t, trainPath, trainFolder)
	assert.Equal(t, testPath, testFolder)

	trainDataPath := fmt.Sprintf("test/tmp_data/%s/train/tables/learningData.csv", splitDatasetName)
	testDataPath := fmt.Sprintf("test/tmp_data/%s/test/tables/learningData.csv", splitDatasetName)
	lines, err := util.ReadCSVFile(trainDataPath, true)
	assert.NoError(t, err)
	assert.Equal(t, 29, len(lines))

	lines, err = util.ReadCSVFile(testDataPath, true)
	assert.NoError(t, err)
	assert.Equal(t, 4, len(lines))
}

func TestPersistOriginalDataStratified(t *testing.T) {
	assert.NoError(t, removeTestFiles())
	params := createTestParams(true, ModelQualityHigh)
	limits, err := createLimits(ModelQualityHigh)
	assert.NoError(t, err)

	splitDatasetName, err := generateSplitDatasetName(params, limits)
	assert.NoError(t, err)
	trainPath := fmt.Sprintf("test/tmp_data/%s/train/datasetDoc.json", splitDatasetName)
	testPath := fmt.Sprintf("test/tmp_data/%s/test/datasetDoc.json", splitDatasetName)

	trainFolder, testFolder, err := persistOriginalData(params)
	assert.NoError(t, err)
	assert.Equal(t, trainPath, trainFolder)
	assert.Equal(t, testPath, testFolder)

	trainDataPath := fmt.Sprintf("test/tmp_data/%s/train/tables/learningData.csv", splitDatasetName)
	testDataPath := fmt.Sprintf("test/tmp_data/%s/test/tables/learningData.csv", splitDatasetName)
	lines, err := util.ReadCSVFile(trainDataPath, true)
	assert.NoError(t, err)
	assert.Equal(t, 28, len(lines))

	lines, err = util.ReadCSVFile(testDataPath, true)
	assert.NoError(t, err)
	assert.Equal(t, 5, len(lines))

	categoricalValues := map[string]int{}
	for _, rowData := range lines {
		categoricalValues[rowData[2]]++
	}
	assert.Equal(t, 2, categoricalValues["a"])
	assert.Equal(t, 1, categoricalValues["b"])
	assert.Equal(t, 1, categoricalValues["c"])
	assert.Equal(t, 1, categoricalValues["d"])
}

func TestParamChange(t *testing.T) {
	assert.NoError(t, removeTestFiles())

	params := createTestParams(false, ModelQualityHigh)
	limits, err := createLimits(ModelQualityHigh)
	assert.NoError(t, err)
	splitDatasetName0, err := generateSplitDatasetName(params, limits)
	assert.NoError(t, err)

	_, _, err = persistOriginalData(params)
	assert.NoError(t, err)
	trainPath := fmt.Sprintf("test/tmp_data/%s/train/datasetDoc.json", splitDatasetName0)
	assert.FileExists(t, trainPath)
	testPath := fmt.Sprintf("test/tmp_data/%s/test/datasetDoc.json", splitDatasetName0)
	assert.FileExists(t, testPath)

	params = createTestParams(true, ModelQualityHigh)
	splitDatasetName1, err := generateSplitDatasetName(params, limits)
	assert.NoError(t, err)
	assert.NotEqual(t, splitDatasetName0, splitDatasetName1)

	_, _, err = persistOriginalData(params)
	assert.NoError(t, err)
	trainPath = fmt.Sprintf("test/tmp_data/%s/train/datasetDoc.json", splitDatasetName1)
	assert.FileExists(t, trainPath)
	testPath = fmt.Sprintf("test/tmp_data/%s/test/datasetDoc.json", splitDatasetName1)
	assert.FileExists(t, testPath)
}

func TestLimitChange(t *testing.T) {
	assert.NoError(t, removeTestFiles())

	params := createTestParams(false, ModelQualityHigh)
	limits, err := createLimits(ModelQualityHigh)
	assert.NoError(t, err)
	splitDatasetName0, err := generateSplitDatasetName(params, limits)
	assert.NoError(t, err)

	_, _, err = persistOriginalData(params)
	assert.NoError(t, err)
	trainPath := fmt.Sprintf("test/tmp_data/%s/train/datasetDoc.json", splitDatasetName0)
	assert.FileExists(t, trainPath)
	testPath := fmt.Sprintf("test/tmp_data/%s/test/datasetDoc.json", splitDatasetName0)
	assert.FileExists(t, testPath)

	params = createTestParams(false, ModelQualityFast)
	limits, err = createLimits(ModelQualityFast)
	assert.NoError(t, err)
	splitDatasetName1, err := generateSplitDatasetName(params, limits)
	assert.NoError(t, err)
	assert.NotEqual(t, splitDatasetName0, splitDatasetName1)

	_, _, err = persistOriginalData(params)
	assert.NoError(t, err)
	trainPath = fmt.Sprintf("test/tmp_data/%s/train/datasetDoc.json", splitDatasetName1)
	assert.FileExists(t, trainPath)
	testPath = fmt.Sprintf("test/tmp_data/%s/test/datasetDoc.json", splitDatasetName1)
	assert.FileExists(t, testPath)
}

func createTestParams(stratify bool, quality string) *persistedDataParams {
	datasetStorage = serialization.NewCSV()
	return &persistedDataParams{
		DatasetName:        "test_dataset",
		SchemaFile:         compute.D3MDataSchema,
		SourceDataFolder:   "./test/test_dataset",
		TmpDataFolder:      "./test/tmp_data",
		TaskType:           []string{"classification"},
		GroupingFieldIndex: -1,
		TargetFieldIndex:   2,
		Stratify:           stratify,
		Quality:            quality,
	}
}

func createLimits(quality string) (*rowLimits, error) {
	var config env.Config
	config, err := env.LoadConfig()
	if err != nil {
		return nil, err
	}
	return &rowLimits{
		MinTrainingRows: config.MinTrainingRows,
		MinTestRows:     config.MinTestRows,
		MaxTrainingRows: config.MaxTestRows,
		MaxTestRows:     config.MaxTestRows,
		Sample:          config.FastDataPercentage,
		Quality:         quality,
	}, nil
}

func removeTestFiles() error {
	files, err := filepath.Glob("./test/tmp_data/test_dataset*")
	if err != nil {
		return errors.Wrap(err, "temp file glob failed")
	}
	for _, f := range files {
		if pathExists(f) {
			if err := os.RemoveAll(f); err != nil {
				return errors.Wrap(err, "temp file remove failed")
			}
		}
	}
	return nil
}
