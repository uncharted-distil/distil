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
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/metadata"
	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil-compute/primitive/compute/description"
	"github.com/uncharted-distil/distil-compute/primitive/compute/result"
	"github.com/uncharted-distil/distil/api/env"
	log "github.com/unchartedsoftware/plog"
)

var excludedTypes = map[string]bool{
	model.MultiBandImageType: true,
	model.ImageType:          true,
	model.TimeSeriesType:     true,
	model.RealVectorType:     true,
	model.GeoBoundsType:      true,
	model.GeoCoordinateType:  true,
}

var excludedRoles = map[string]bool{
	model.VarDistilRoleMetadata:   true,
	model.VarDistilRoleSystemData: true,
}

// TargetRank will rank the dataset relative to a target variable using
// a primitive.
func TargetRank(dataset string, target string, features []*model.Variable, source metadata.DatasetSource) (map[string]float64, error) {
	// Some feature types cannot be / should not be ranked - we should skip ranking if there
	// aren't at least 3 valid (target + 2 features)

	// find target feature
	var targetFeature *model.Variable
	for _, feature := range features {
		if strings.EqualFold(feature.Key, target) {
			targetFeature = feature
			break
		}
	}

	// filter features by type / role and skip if we don't have at least 2 features that are valid
	filteredFeatures := filterFeatures(features, target)
	if len(filteredFeatures) <= 2 || excludedTypes[targetFeature.Type] {
		return map[string]float64{}, nil
	}
	filteredFeatures[target] = true

	// create & submit the solution request - we send the list of filtered features to ensure that
	pip, err := description.CreateTargetRankingPipeline("target_rank", "feature ranking relative to the target", targetFeature, features, filteredFeatures)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create ranking pipeline")
	}

	datasetInputDir := env.ResolvePath(source, dataset)
	datasetInputDir, err = filepath.Abs(datasetInputDir)
	if err != nil {
		return nil, errors.Errorf("path \"%s\" cannot be made absolute", datasetInputDir)
	}

	datasetURI, err := submitPipeline([]string{datasetInputDir}, pip, true)
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

func filterFeatures(features []*model.Variable, target string) map[string]bool {
	filteredFeatures := map[string]*model.Variable{}

	groupedVars := []*model.Variable{}
	for _, feature := range features {
		// if this is a grouped var save for follow on processing
		if feature.IsGrouping() {
			groupedVars = append(groupedVars, feature)
		}

		// check if this is a feature that we've marked as elligible for ranking and save it if so
		if !excludedTypes[feature.Type] && !excludedRoles[feature.DistilRole] && feature.Key != target {
			filteredFeatures[feature.Key] = feature
		}
	}

	// remove any variables that are rankable, but are marked as hidden by their parent grouping
	for _, groupedVar := range groupedVars {
		hidden := groupedVar.Grouping.GetHidden()
		for _, hiddenVar := range hidden {
			delete(filteredFeatures, hiddenVar)
		}
	}

	featureList := map[string]bool{}
	for _, feature := range filteredFeatures {
		featureList[strings.ToLower(feature.Key)] = true
	}
	return featureList
}
