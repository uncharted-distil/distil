//
//   Copyright © 2021 Uncharted Software Inc.
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
	"encoding/json"
	"os"
	"path"
	"strconv"

	"github.com/pkg/errors"
	log "github.com/unchartedsoftware/plog"

	"github.com/uncharted-distil/distil-compute/primitive/compute/description"
	"github.com/uncharted-distil/distil-compute/primitive/compute/result"
	"github.com/uncharted-distil/distil/api/util"
)

// ImportanceResult is the result from a ranking operation.
type ImportanceResult struct {
	Path     string    `json:"path"`
	Features []float64 `json:"features"`
}

// Rank will rank the dataset using a primitive.
func Rank(schemaPath string, dataset string, config *IngestTaskConfig) (string, error) {
	schemaDoc := path.Dir(schemaPath)

	// create & submit the solution request
	pip, err := description.CreatePCAFeaturesPipeline("harry", "")
	if err != nil {
		return "", errors.Wrap(err, "unable to create PCA pipeline")
	}

	datasetURI, err := submitPipeline([]string{schemaDoc}, pip, true)
	if err != nil {
		return "", errors.Wrap(err, "unable to run PCA pipeline")
	}

	// parse primitive response (col index,importance)
	res, err := result.ParseResultCSV(datasetURI)
	if err != nil {
		return "", errors.Wrap(err, "unable to parse PCA pipeline result")
	}

	ranks := make([]float64, len(res)-1)
	for i, v := range res {
		if i > 0 {
			colIndex, err := strconv.ParseInt(v[0].(string), 10, 64)
			if err != nil {
				return "", errors.Wrap(err, "unable to parse PCA col index")
			}
			vInt, err := strconv.ParseFloat(v[1].(string), 64)
			if err != nil {
				return "", errors.Wrap(err, "unable to parse PCA rank value")
			}
			ranks[colIndex] = vInt
		}
	}

	importance := &ImportanceResult{
		Path:     datasetURI,
		Features: ranks,
	}

	// output the importance in the expected JSON format
	bytes, err := json.MarshalIndent(importance, "", "    ")
	if err != nil {
		return "", errors.Wrap(err, "unable to serialize ranking result")
	}

	// write to file
	outputPath := path.Join(schemaDoc, config.RankingOutputPathRelative)
	log.Debugf("writing ranking output to %s", outputPath)
	err = util.WriteFileWithDirs(outputPath, bytes, os.ModePerm)
	if err != nil {
		return "", errors.Wrap(err, "unable to store ranking result")
	}

	return outputPath, nil
}
