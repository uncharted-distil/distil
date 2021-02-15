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

package task

import (
	"encoding/csv"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil-compute/primitive/compute/description"
	"github.com/uncharted-distil/distil/api/env"
	apiModel "github.com/uncharted-distil/distil/api/model"
)

type testSubmitter struct{}

func (testSubmitter) submit(datasetURIs []string, pipelineDesc *description.FullySpecifiedPipeline) (string, error) {
	return "file://test_data/result.csv", nil
}

func TestJoin(t *testing.T) {

	varsLeft := []*model.Variable{
		{
			Key:         "d3mIndex",
			DisplayName: "D3M Index",
			Type:        model.IntegerType,
			DistilRole:  "data",
		},
		{
			Key:         "alpha",
			DisplayName: "Alpha",
			Type:        model.RealType,
			DistilRole:  "data",
		},
		{
			Key:         "bravo",
			DisplayName: "Bravo",
			Type:        model.IntegerType,
			DistilRole:  "data",
		},
	}

	varsRight := []*model.Variable{
		{
			Key:         "d3mIndex",
			DisplayName: "D3M Index",
			Type:        model.IntegerType,
			DistilRole:  "data",
		},
		{
			Key:          "charlie",
			DisplayName:  "Charlie",
			Type:         model.CountryType,
			OriginalType: model.CategoricalFilter,
			DistilRole:   "data",
		},
		{
			Key:         "delta",
			DisplayName: "Delta",
			Type:        model.IntegerType,
			DistilRole:  "data",
		},
	}

	cfg, err := env.LoadConfig()
	assert.NoError(t, err)

	cfg.D3MOutputDir = "test_data"
	cfg.D3MInputDir = "test_data"
	cfg.DatamartImportFolder = "test_data"
	assert.NoError(t, env.Initialize(&cfg))

	leftJoin := &JoinSpec{
		DatasetID:     "test_1",
		DatasetFolder: "test_1_TRAIN",
		DatasetSource: "contrib",
	}

	rightJoin := &JoinSpec{
		DatasetID:     "test_2",
		DatasetFolder: "test_2_TRAIN",
		DatasetSource: "contrib",
	}

	rightOrigin := &model.DatasetOrigin{
		SearchResult: "{}",
		Provenance:   "NYU",
	}

	pipelineDesc, err := description.CreateDatamartAugmentPipeline("Join Preview",
		"Join to be reviewed by user", rightOrigin.SearchResult, rightOrigin.Provenance)
	assert.NoError(t, err)
	datasetLeftURI := env.ResolvePath(leftJoin.DatasetSource, leftJoin.DatasetFolder)
	_, result, err := join(leftJoin, rightJoin, varsLeft, varsRight, pipelineDesc, []string{datasetLeftURI}, testSubmitter{}, &cfg)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	expected := [][]string{
		{"D3M Index", "Alpha", "Charlie"},
		{"0", "1.0", "a"},
		{"1", "2.0", "b"},
		{"2", "3.0", "c"},
		{"3", "4.0", "d"},
	}

	csvFile, err := os.Open("test_data/augmented/test_1-test_2/tables/learningData.csv")
	assert.NoError(t, err)
	defer csvFile.Close()

	csvReader := csv.NewReader(csvFile)
	records, err := csvReader.ReadAll()
	assert.NoError(t, err)
	assert.Equal(t, len(expected), len(records))
	for i := 0; i < len(records); i++ {
		assert.ElementsMatch(t, records[i], expected[i])
	}

	assert.ElementsMatch(t, result.Columns, []*apiModel.Column{
		{
			Label:  "D3M Index",
			Key:    "d3mIndex",
			Type:   model.IntegerType,
			Weight: float64(0),
		},
		{
			Label:  "Alpha",
			Key:    "alpha",
			Type:   model.RealType,
			Weight: float64(0),
		},
		{
			Label:  "Charlie",
			Key:    "Charlie",
			Type:   model.CategoricalType,
			Weight: float64(0),
		},
	})

	expectedTyped := [][]interface{}{
		{int64(0), 1.0, "a"},
		{int64(1), 2.0, "b"},
		{int64(2), 3.0, "c"},
		{int64(3), 4.0, "d"},
	}

	assert.Equal(t, len(expectedTyped), len(records)-1)
	assert.Equal(t, result.NumRows, 4)
	for i := 0; i < len(expectedTyped); i++ {
		for j := 0; j < len(expectedTyped[i]); j++ {
			assert.Equal(t, result.Values[i][j].Value, expectedTyped[i][j])
		}
	}
}
