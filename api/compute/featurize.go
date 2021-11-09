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

package compute

import (
	"fmt"

	uuid "github.com/gofrs/uuid"
	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil-compute/pipeline"
	"github.com/uncharted-distil/distil-compute/primitive/compute/description"

	api "github.com/uncharted-distil/distil/api/model"
)

// FeaturizeDataset creates feature outputs that can then be used directly when
// modelling instead of needing to run the complete pipeline.
func FeaturizeDataset(dataset string, target string) ([]string, error) {
	// build the normal solution search

	// start a solution searchID
	//requestID, err := client.StartSearch(context.Background(), searchRequest)
	//if err != nil {
	//		return nil, err
	//}

	return nil, nil
}

// createPreFeaturizedPipeline creates pipeline prepend to process a featurized dataset.
func (s *SolutionRequest) createPreFeaturizedPipeline(learningDataset string,
	sourceVariables []*model.Variable, featurizedVariables []*model.Variable,
	metaStorage api.MetadataStorage, targetIndex int) (*pipeline.PipelineDescription, error) {
	uuid, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	name := fmt.Sprintf("prefeaturized-%s-%s-%s", s.Dataset, learningDataset, uuid.String())
	desc := fmt.Sprintf("Prefeaturized pipeline capturing user feature selection and type information. Dataset: `%s` ID: `%s`", s.Dataset, uuid.String())

	expandedFilters, err := api.ExpandFilterParams(s.Dataset, s.Filters, true, metaStorage)
	if err != nil {
		return nil, err
	}

	// Ensure we remove multiband image data (replaced by feature vectors) and geo coordinates polygon
	// string.  These are in the pre-featurized learning data but are not needed.
	toRemove := map[string]bool{}
	for _, variable := range sourceVariables {
		switch v := variable.Grouping.(type) {
		case *model.MultiBandImageGrouping:
			toRemove[v.BandCol] = true
			toRemove[v.ImageCol] = true
			toRemove[v.IDCol] = true
		case *model.GeoBoundsGrouping:
			toRemove[v.PolygonCol] = true
		default:
			continue
		}
	}
	selectedVariables := []string{}
	for _, v := range expandedFilters.Variables {
		if _, ok := toRemove[v]; !ok {
			selectedVariables = append(selectedVariables, v)
		}
	}

	// image feature variables are m system data role and should be included in our selected set
	for _, featurizedVariable := range featurizedVariables {
		if featurizedVariable.HasRole(model.VarDistilRoleSystemData) {
			selectedVariables = append(selectedVariables, featurizedVariable.Key)
		}
	}

	expandedFilters.Variables = selectedVariables

	prefeaturizedPipeline, err := description.CreatePreFeaturizedDatasetPipeline(name, desc,
		&description.UserDatasetDescription{
			AllFeatures:      featurizedVariables,
			TargetFeature:    featurizedVariables[targetIndex],
			SelectedFeatures: expandedFilters.Variables,
			Filters:          s.Filters.Filters,
		}, nil)
	if err != nil {
		return nil, err
	}

	return prefeaturizedPipeline, nil
}
