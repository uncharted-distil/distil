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
	initializeTestConfig()
	expectedSamplePath := fmt.Sprintf("test/tmp_data/sample-6f5d2a90aff56355/datasetDoc.json")

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

func createTestSampler(stratify bool) datasetSplitter {

	if stratify {
		return &stratifiedSplitter{
			targetCol:   2,
			groupingCol: -1,
		}
	}
	return &basicSplitter{
		targetCol:   2,
		groupingCol: -1,
	}
}
