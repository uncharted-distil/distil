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
	"os"
	"path/filepath"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/uncharted-distil/distil-compute/primitive/compute"
	"github.com/uncharted-distil/distil/api/task"
)

func TestPersistOriginalDataUnstratified(t *testing.T) {
	assert.NoError(t, removeTestFiles())
	params := createTestParams(false)
	trainFolder, testFolder, err := persistOriginalData(params)

	assert.NoError(t, err)
	assert.Equal(t, "test/tmp_data/test_dataset/train/datasetDoc.json", trainFolder)
	assert.Equal(t, "test/tmp_data/test_dataset/test/datasetDoc.json", testFolder)

	lines, err := task.ReadCSVFile("test/tmp_data/test_dataset/train/tables/learningData.csv", true)
	assert.NoError(t, err)
	assert.Equal(t, 29, len(lines))

	lines, err = task.ReadCSVFile("test/tmp_data/test_dataset/test/tables/learningData.csv", true)
	assert.NoError(t, err)
	assert.Equal(t, 4, len(lines))
}

func TestPersistOriginalDataStratified(t *testing.T) {
	assert.NoError(t, removeTestFiles())
	params := createTestParams(true)
	trainFolder, testFolder, err := persistOriginalData(params)

	assert.NoError(t, err)
	assert.Equal(t, "test/tmp_data/test_dataset/train/datasetDoc.json", trainFolder)
	assert.Equal(t, "test/tmp_data/test_dataset/test/datasetDoc.json", testFolder)

	lines, err := task.ReadCSVFile("test/tmp_data/test_dataset/train/tables/learningData.csv", true)
	assert.NoError(t, err)
	assert.Equal(t, 28, len(lines))

	lines, err = task.ReadCSVFile("test/tmp_data/test_dataset/test/tables/learningData.csv", true)
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

func TestCacheHit(t *testing.T) {
	assert.NoError(t, removeTestFiles())
	params := createTestParams(true)
	keyMatch, err := checkAndUpdateCacheKey(params)
	assert.NoError(t, err)
	params = createTestParams(true)
	keyMatch, err = checkAndUpdateCacheKey(params)
	assert.NoError(t, err)
	assert.True(t, keyMatch)
}

func TestCacheMiss(t *testing.T) {
	assert.NoError(t, removeTestFiles())
	params := createTestParams(true)
	keyMatch, err := checkAndUpdateCacheKey(params)
	params = createTestParams(false)
	keyMatch, err = checkAndUpdateCacheKey(params)
	assert.NoError(t, err)
	assert.False(t, keyMatch)
}

func createTestParams(stratify bool) *persistedDataParams {
	return &persistedDataParams{
		DatasetName:          "test_dataset",
		SchemaFile:           compute.D3MDataSchema,
		SourceDataFolder:     "./test/test_dataset",
		TmpDataFolder:        "./test/tmp_data",
		TaskType:             "classification",
		TimeseriesFieldIndex: -1,
		TargetFieldIndex:     2,
		Stratify:             stratify,
	}
}

func removeTestFiles() error {
	files, err := filepath.Glob("./test/tmp_data/.split*")
	if err != nil {
		return errors.Wrap(err, "temp file glob failed")
	}
	files = append(files, "./test/tmp_data/test_dataset/train/test_dataset", "./test/tmp_data/test_dataset/test/test_dataset")
	for _, f := range files {
		if pathExists(f) {
			if err := os.RemoveAll(f); err != nil {
				return errors.Wrap(err, "temp file remove failed")
			}
		}
	}
	return nil
}
