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
	"path"
	"strings"

	"github.com/uncharted-distil/distil-compute/metadata"
	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil-compute/primitive/compute"
	apicomp "github.com/uncharted-distil/distil/api/compute"
	"github.com/uncharted-distil/distil/api/env"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/serialization"
)

type explainDataset struct {
	explainURI string
	resultURI  string
	target     string
}

// CreateDataset creates the raw dataset structure for the SHAP dataset.
func (e explainDataset) CreateDataset(rootDataPath string, datasetName string, config *env.Config) (*serialization.RawDataset, error) {
	explainStorage := serialization.GetStorage(e.explainURI)
	explainedOutput, err := apicomp.ExplainFeatureOutput(e.resultURI, e.explainURI)
	if err != nil {
		return nil, err
	}

	outputDataURI := strings.Replace(compute.D3MLearningData, path.Ext(compute.D3MLearningData), path.Ext(e.explainURI), 1)
	outputDataURI = path.Join(rootDataPath, compute.D3MDataFolder, outputDataURI)

	// add the target variable placeholder
	explainedData := explainedOutput.Values
	explainedData[0] = append(explainedData[0], e.target)
	for i := 1; i < len(explainedData); i++ {
		// need multiple classes for it to work
		// TODO: GET TARGETS FROM INPUT DATA!
		explainedData[i] = append(explainedData[i], fmt.Sprintf("placeholder_%d", i%5))
	}

	// write out the data to the dataset folder
	err = explainStorage.WriteData(outputDataURI, explainedData)
	if err != nil {
		return nil, err
	}

	// every field except d3m index is a float
	datasetID := model.NormalizeDatasetID(datasetName)
	meta := model.NewMetadata(datasetName, datasetName, "", datasetID)
	dr := model.NewDataResource(compute.DefaultResourceID, model.ResTypeTable, map[string][]string{compute.D3MResourceFormat: {"csv"}})
	dr.ResPath = outputDataURI
	for i, field := range explainedData[0] {
		typ := model.RealType
		role := model.RoleAttribute
		if field == e.target {
			typ = model.StringType
		} else if field == model.D3MIndexFieldName {
			typ = model.IndexType
			role = model.RoleIndex
		}
		dr.Variables = append(dr.Variables,
			model.NewVariable(i, field, field, field, field, typ, typ,
				"", []string{role}, model.VarDistilRoleData, nil, dr.Variables, true))
	}

	meta.DataResources = []*model.DataResource{dr}
	return &serialization.RawDataset{
		ID:       e.explainURI,
		Name:     e.explainURI,
		Data:     explainedData,
		Metadata: meta,
	}, nil
}

// ClusterExplainOutput clusters the explained output from a model.
func ClusterExplainOutput(variable string, resultURI string, explainURI string, config *env.Config) (bool, []*ClusterPoint, error) {
	// create the SHAP values dataset
	ds := explainDataset{
		explainURI: explainURI,
		resultURI:  resultURI,
		target:     variable,
	}
	outputPath := path.Join(config.D3MOutputDir, config.AugmentedSubFolder)
	datasetName := strings.TrimSuffix(path.Base(explainURI), path.Ext(explainURI))
	_, dsPath, err := CreateDataset(datasetName, ds, outputPath, config)
	if err != nil {
		return false, nil, err
	}

	// read the metadata from the just created dataset
	explainStorage := serialization.GetStorage(explainURI)
	meta, err := explainStorage.ReadMetadata(path.Join(dsPath, compute.D3MDataSchema))
	if err != nil {
		return false, nil, err
	}

	// cluster the SHAP values dataset as any other dataset
	dsCreated := &api.Dataset{
		ID:              meta.ID,
		Name:            meta.Name,
		StorageName:     meta.StorageName,
		Folder:          meta.ID,
		Description:     meta.Description,
		Summary:         meta.Summary,
		SummaryML:       meta.SummaryMachine,
		Variables:       meta.GetMainDataResource().Variables,
		NumRows:         meta.NumRows,
		NumBytes:        meta.NumBytes,
		Provenance:      meta.SearchProvenance,
		Source:          metadata.Augmented,
		Type:            api.DatasetType(meta.Type),
		LearningDataset: meta.LearningDataset,
	}

	addMeta, clustered, err := Cluster(dsCreated, variable, true)
	if err != nil {
		return false, nil, err
	}

	return addMeta, clustered, nil
}
