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
	"path"

	"github.com/uncharted-distil/distil-compute/metadata"
	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil-compute/primitive/compute"
	"github.com/uncharted-distil/distil-compute/primitive/compute/description"
	"github.com/uncharted-distil/distil-compute/primitive/compute/result"

	"github.com/uncharted-distil/distil/api/env"
	api "github.com/uncharted-distil/distil/api/model"
)

// ClusterPoint contains data that has been clustered.
type ClusterPoint struct {
	D3MIndex    string
	SourceField string
	Label       string
}

// Cluster will cluster the dataset fields using a primitive.
func Cluster(dataset *api.Dataset, variable string, useKMeans bool, clusterCount int) (bool, []*ClusterPoint, error) {

	datasetInputDir := env.ResolvePath(dataset.Source, dataset.Folder)
	features := dataset.Variables

	// find the particular clustering variable - relevant for images and remote sensing tile sets, not
	// needed for full set clustering
	var clusteringVar *model.Variable
	for _, v := range features {
		if v.Key == variable {
			clusteringVar = v
		}
	}

	envConfig, err := env.LoadConfig()
	if err != nil {
		return false, nil, err
	}

	clusterParams := &description.ClusterParams{
		UseKMeans:    useKMeans,
		ClusterCount: clusterCount,
		PoolFeatures: envConfig.PoolFeatures,
	}

	var step *description.FullySpecifiedPipeline
	clusterGroup := getClusterGroup(clusteringVar.Key, features)
	if model.IsImage(clusteringVar.Type) {
		step, err = description.CreateImageClusteringPipeline("image_cluster", "basic image clustering", []*model.Variable{clusteringVar}, clusterParams)
	} else if clusterGroup != nil && model.IsMultiBandImage(clusterGroup.GetType()) {
		// Check to see if this dataset redirects to a different dataset for actual learning / analytic tasks.
		if dataset.LearningDataset != "" {
			// get the pre-featurized dataset location and its metadata
			datasetInputDir = dataset.LearningDataset
			var meta *model.Metadata
			meta, err = metadata.LoadMetadataFromOriginalSchema(path.Join(datasetInputDir, compute.D3MDataSchema), false)
			if err == nil {
				// the pre-featurized dataset does not have remote sensing image file names - they have instead been replaced
				// with 2048 columns of float values generated by the pre-featurization step.  We need to use this variable list
				// for clustering.
				variables := meta.GetMainDataResource().Variables
				step, err = description.CreatePreFeaturizedMultiBandImageClusteringPipeline(
					"remote_sensing_cluster", "k-means pre-featurized remote sensing clustering", variables, clusterParams)
			}
		} else {
			rsg := clusterGroup.(*model.MultiBandImageGrouping)
			step, err = description.CreateMultiBandImageClusteringPipeline("remote_sensing_cluster", "multiband image clustering",
				rsg, features, clusterParams, envConfig.RemoteSensingGPUBatchSize, envConfig.RemoteSensingNumJobs)
		}
	} else if clusteringVar.HasRole(model.VarDistilRoleGrouping) {
		// assume timeseries for now if distil role is grouping
		step, err = description.CreateSlothPipeline("timeseries_cluster", "k-means time series clustering",
			"", "", clusterGroup.(*model.TimeseriesGrouping), features)
	} else {
		// general clustering pipeline
		selectedFeatures := make([]string, len(features))
		for i, f := range features {
			selectedFeatures[i] = f.Key
		}
		datasetDescription := &description.UserDatasetDescription{
			AllFeatures:      features,
			TargetFeature:    clusteringVar,
			SelectedFeatures: selectedFeatures,
		}
		step, err = description.CreateGeneralClusteringPipeline("tabular_cluster",
			"k-means tabular clustering", datasetDescription, nil, clusterParams)
	}
	if err != nil {
		return false, nil, err
	}

	datasetURI, err := submitPipeline([]string{datasetInputDir}, step, true)
	if err != nil {
		return false, nil, err
	}

	// parse primitive response (new field contains output)
	res, err := result.ParseResultCSV(datasetURI)
	if err != nil {
		return false, nil, err
	}
	header, err := castTypeArray(res[0])
	if err != nil {
		return false, nil, err
	}

	// find the field with the feature output
	clusterIndex := getFieldIndex(header, "__cluster")
	if clusterIndex == -1 {
		// cluster label may be returned with target name
		clusterIndex = getFieldIndex(header, variable)
	}
	d3mIndexIndex := getFieldIndex(header, model.D3MIndexFieldName)
	if clusterIndex == -1 && len(header) == 2 {
		// default to second column
		clusterIndex = (d3mIndexIndex + 1) % 2
	}

	// build the output (skipping the header)
	clusteredData := make([]*ClusterPoint, len(res)-1)
	for i, v := range res[1:] {
		label := createFriendlyLabel(v[clusterIndex].(string))
		d3mIndex := v[d3mIndexIndex].(string)

		clusteredData[i] = &ClusterPoint{
			D3MIndex:    d3mIndex,
			SourceField: variable,
			Label:       label,
		}
	}

	return true, clusteredData, nil
}

func getClusterGroup(clusterVar string, features []*model.Variable) model.BaseGrouping {
	for _, feature := range features {
		if feature.IsGrouping() && feature.Grouping.GetIDCol() == clusterVar {
			return feature.Grouping
		}
	}
	return nil
}
