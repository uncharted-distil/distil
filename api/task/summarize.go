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
	"strings"

	"github.com/pkg/errors"
	log "github.com/unchartedsoftware/plog"

	"github.com/uncharted-distil/distil-compute/primitive/compute/description"
	"github.com/uncharted-distil/distil-compute/primitive/compute/result"
	"github.com/uncharted-distil/distil/api/util"
)

// SummaryResult represents a summary result.
type SummaryResult struct {
	Summary string `json:"summary"`
}

// Summarize will summarize the dataset using a primitive.
func Summarize(schemaPath string, dataset string, config *IngestTaskConfig) (string, error) {
	schemaDoc := path.Dir(schemaPath)

	// create & submit the solution request
	pip, err := description.CreateDukePipeline("wellington", "")
	if err != nil {
		return "", errors.Wrap(err, "unable to create Duke pipeline")
	}

	datasetURI, err := submitPipeline([]string{schemaDoc}, pip, true)
	if err != nil {
		return "", errors.Wrap(err, "unable to run Duke pipeline")
	}

	// parse primitive response (token,probability)
	res, err := result.ParseResultCSV(datasetURI)
	if err != nil {
		return "", errors.Wrap(err, "unable to parse Duke pipeline result")
	}

	tokens := make([]string, len(res)-1)
	for i, v := range res {
		// skip the header
		if i > 0 {
			token, ok := v[0].(string)
			if !ok {
				return "", errors.Wrap(err, "unable to parse Duke token")
			}
			tokens[i-1] = token
		}
	}

	sum := &SummaryResult{
		Summary: strings.Join(tokens, ", "),
	}

	// output the classification in the expected JSON format
	bytes, err := json.MarshalIndent(sum, "", "    ")
	if err != nil {
		return "", errors.Wrap(err, "unable to serialize summary result")
	}
	// write to file
	outputPath := path.Join(schemaDoc, config.SummaryMachineOutputPathRelative)
	log.Debugf("writing summary output to %s", outputPath)
	err = util.WriteFileWithDirs(outputPath, bytes, os.ModePerm)
	if err != nil {
		return "", errors.Wrap(err, "unable to store summary result")
	}

	return outputPath, nil
}
