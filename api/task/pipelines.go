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

	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil-compute/primitive/compute"
	"github.com/uncharted-distil/distil-compute/primitive/compute/description"

	sr "github.com/uncharted-distil/distil/api/compute"
)

const (
	denormFieldName = "filename"
)

var (
	client *compute.Client
)

// FeatureRequest captures the properties of a request to a primitive.
type FeatureRequest struct {
	SourceVariableName  string
	FeatureVariableName string
	OutputVariableName  string
	Variable            *model.Variable
	Step                *description.FullySpecifiedPipeline
	Clustering          bool
}

type datasetCopyPath struct {
	sourceFolder string
	outputFolder string
	outputSchema string
	outputData   string
}

// SetClient sets the compute client to use when invoking primitives.
func SetClient(computeClient *compute.Client) {
	client = computeClient
}

func submitPipeline(datasets []string, step *description.FullySpecifiedPipeline, shouldCache bool) (string, error) {
	return sr.SubmitPipeline(client, datasets, nil, nil, step, nil, shouldCache)
}

func getD3MIndexField(dr *model.DataResource) int {
	for _, v := range dr.Variables {
		if v.Key == model.D3MIndexFieldName {
			return v.Index
		}
	}

	return -1
}

func getDataResource(meta *model.Metadata, resID string) *model.DataResource {
	// main data resource has d3m index variable
	for _, dr := range meta.DataResources {
		if dr.ResID == resID {
			return dr
		}
	}

	return nil
}

func mapHeaderFields(meta *model.Metadata) map[string]*model.Variable {
	// cycle through each data resource, mapping field names to variables.
	fields := make(map[string]*model.Variable)
	for _, dr := range meta.DataResources {
		for _, v := range dr.Variables {
			fields[v.HeaderName] = v
		}
	}

	return fields
}

func mapDenormFields(mainDR *model.DataResource) map[string]*model.Variable {
	fields := make(map[string]*model.Variable)
	for _, field := range mainDR.Variables {
		if field.IsMediaReference() {
			// DENORM PRIMITIVE RENAMES REFERENCE FIELDS TO `filename`
			fields[denormFieldName] = field
		}
	}
	return fields
}

func createDatasetPaths(schemaFile string, dataset string, dataPathRelative string) *datasetCopyPath {
	sourceFolder := path.Dir(schemaFile)
	outputSchemaPath := schemaFile
	outputDataPath := path.Join(sourceFolder, compute.D3MDataFolder, dataPathRelative)
	outputFolder := sourceFolder

	return &datasetCopyPath{
		sourceFolder: sourceFolder,
		outputFolder: outputFolder,
		outputSchema: outputSchemaPath,
		outputData:   outputDataPath,
	}
}

func createFriendlyLabel(label string) string {
	// label is a char between 1 and cluster max
	if label == "-1" {
		return "Other"
	}
	return fmt.Sprintf("Pattern %s", string('A'-'0'+label[0]))
}

func createFriendlyOutlierLabel(label string) string {
	if label == "-1" {
		return "anomaly"
	}
	return "regular"
}

func getFieldIndex(header []string, fieldName string) int {
	for i, f := range header {
		if f == fieldName {
			return i
		}
	}

	return -1
}
