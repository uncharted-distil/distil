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
	"path"
	"strings"

	"github.com/uncharted-distil/distil-compute/metadata"
	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil-compute/primitive/compute"
	"github.com/uncharted-distil/distil/api/env"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/serialization"
)

type explainDataset struct {
	explainURI string
}

// CreateDataset creates the raw dataset structure for the SHAP dataset.
func (e explainDataset) CreateDataset(rootDataPath string, datasetName string, config *env.Config) (*api.RawDataset, error) {
	explainStorage := serialization.GetStorage(e.explainURI)
	explainedOutput, err := explainStorage.ReadData(e.explainURI)
	if err != nil {
		return nil, err
	}

	outputDataURI := strings.Replace(compute.D3MLearningData, path.Ext(compute.D3MLearningData), path.Ext(e.explainURI), 1)
	outputDataURI = path.Join(rootDataPath, compute.D3MDataFolder, outputDataURI)

	// write out the data to the dataset folder
	err = explainStorage.WriteData(outputDataURI, explainedOutput)
	if err != nil {
		return nil, err
	}

	// every field except d3m index is a float
	datasetID := model.NormalizeDatasetID(datasetName)
	meta := model.NewMetadata(datasetName, datasetName, "", datasetID)
	dr := model.NewDataResource(compute.DefaultResourceID, model.ResTypeRaw, map[string][]string{compute.D3MResourceFormat: {"csv"}})
	dr.ResPath = outputDataURI
	for i, field := range explainedOutput[0] {
		dr.Variables = append(dr.Variables,
			model.NewVariable(i, field, field, field, model.RealType, model.RealType,
				"", []string{model.RoleAttribute}, model.VarDistilRoleData, nil, dr.Variables, true))
	}

	meta.DataResources = []*model.DataResource{dr}
	return &api.RawDataset{
		ID:       e.explainURI,
		Name:     e.explainURI,
		Data:     explainedOutput,
		Metadata: meta,
	}, nil
}

// ClusterExplainOutput clusters the explained output from a model.
func ClusterExplainOutput(variable string, resultID string, explainURI string, config *env.Config) error {
	// create the SHAP values dataset
	ds := explainDataset{
		explainURI: explainURI,
	}
	_, dsPath, err := CreateDataset(explainURI, ds, "", config)
	if err != nil {
		return err
	}

	// read the metadata from the just created dataset
	explainStorage := serialization.GetStorage(explainURI)
	meta, err := explainStorage.ReadMetadata(dsPath)
	if err != nil {
		return err
	}

	// cluster the SHAP values dataset as any other dataset
	dsCreated := &api.Dataset{
		ID:              meta.ID,
		Name:            meta.Name,
		StorageName:     meta.StorageName,
		Folder:          meta.DatasetFolder,
		Description:     meta.Description,
		Summary:         meta.Summary,
		SummaryML:       meta.SummaryMachine,
		Variables:       meta.GetMainDataResource().Variables,
		NumRows:         meta.NumRows,
		NumBytes:        meta.NumBytes,
		Provenance:      meta.SearchProvenance,
		Source:          metadata.DatasetSource(meta.SourceDataset),
		Type:            api.DatasetType(meta.Type),
		LearningDataset: meta.LearningDataset,
	}

	_, _, err = Cluster(dsCreated, variable, true)
	if err != nil {
		return err
	}

	return nil
}
