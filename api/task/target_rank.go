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
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/metadata"
	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil-compute/primitive/compute/description"
	"github.com/uncharted-distil/distil-compute/primitive/compute/result"
	"github.com/uncharted-distil/distil/api/env"
	log "github.com/unchartedsoftware/plog"
)

// TargetRank will rank the dataset relative to a target variable using
// a primitive.
func TargetRank(dataset string, target string, features []*model.Variable, source metadata.DatasetSource) (map[string]float64, error) {
	// create & submit the solution request
	pip, err := description.CreateTargetRankingPipeline("roger", "", target, features)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create ranking pipeline")
	}

	datasetInputDir := env.ResolvePath(source, dataset)
	datasetInputDir, err = filepath.Abs(datasetInputDir)
	if err != nil {
		return nil, errors.Errorf("path \"%s\" cannot be made absolute", datasetInputDir)
	}

	datasetURI, err := submitPipeline([]string{datasetInputDir}, pip)
	if err != nil {
		return nil, errors.Wrap(err, "unable to run ranking pipeline")
	}

	// parse primitive response (col index,importance)
	res, err := result.ParseResultCSV(datasetURI)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse ranking pipeline result")
	}

	ranks := make(map[string]float64)
	for i, v := range res {
		if i > 0 {
			key, ok := v[1].(string)
			if !ok {
				return nil, fmt.Errorf("unable to parse rank key")
			}

			// default to 0 for rank if error parsing (empty value most likely)
			rank, err := strconv.ParseFloat(v[2].(string), 64)
			if err != nil {
				log.Warnf("defaulting target rank to 0 due to error parsing rank value value: %+v", err)
				rank = 0
			}
			ranks[key] = rank
		}
	}

	return ranks, nil
}
