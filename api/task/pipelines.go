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
	"path"
	"strings"

	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil-compute/primitive/compute"
	"github.com/uncharted-distil/distil-compute/primitive/compute/description"
	"github.com/uncharted-distil/distil-compute/primitive/compute/result"

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
	return sr.SubmitPipeline(client, datasets, nil, nil, step, shouldCache)
}

func appendFeature(dataset string, d3mIndexField int, hasHeader bool, feature *FeatureRequest, lines [][]string) ([][]string, error) {
	datasetURI, err := submitPipeline([]string{dataset}, feature.Step, false)
	if err != nil {
		return nil, errors.Wrap(err, "unable to run pipeline primitive")
	}

	// parse primitive response (new field contains output)
	res, err := result.ParseResultCSV(datasetURI)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse pipeline primitive result")
	}

	// find the field with the feature output
	labelIndex := 1
	for i, f := range res[0] {
		if f == feature.OutputVariableName {
			labelIndex = i
		}
	}

	// build the lookup for the new field
	features := make(map[string]string)
	for i, v := range res {
		// skip header
		if i > 0 {
			d3mIndex := v[0].(string)
			label := v[labelIndex].(string)
			if feature.Clustering {
				label = createFriendlyLabel(label)
			}
			labels := label
			features[d3mIndex] = labels
		}
	}

	// add the new feature to the raw data
	for i, line := range lines {
		if i > 0 || !hasHeader {
			d3mIndex := line[d3mIndexField]
			feature := features[d3mIndex]
			line = append(line, feature)
			lines[i] = line
		}
	}

	return lines, nil
}

func getClusterVariables(meta *model.Metadata, prefix string) ([]*FeatureRequest, error) {
	mainDR := meta.GetMainDataResource()
	features := make([]*FeatureRequest, 0)
	for _, v := range mainDR.Variables {
		if v.RefersTo != nil && v.RefersTo["resID"] != nil {
			// get the refered DR
			resID := v.RefersTo["resID"].(string)

			res := getDataResource(meta, resID)

			// check if needs to be featurized
			if res.ResType == "timeseries" {
				// create the new resource to hold the featured output
				indexName := fmt.Sprintf("%s%s", prefix, v.StorageName)

				// add the feature variable
				v := model.NewVariable(len(mainDR.Variables), indexName, "group", v.StorageName, v.StorageName, model.CategoricalType,
					model.CategoricalType, "", []string{"attribute"}, model.VarDistilRoleMetadata, nil, mainDR.Variables, false)

				// create the required pipeline
				var step *description.FullySpecifiedPipeline
				var err error
				outputName := ""
				if colNames, ok := getTimeValueCols(res); ok {
					step, err = description.CreateSlothPipeline("time series clustering",
						"k-means time series clustering", colNames.timeCol, colNames.valueCol, nil, res.Variables)
					outputName = slothResultFieldName
				}

				if err != nil {
					return nil, errors.Wrap(err, "unable to create step pipeline")
				}

				features = append(features, &FeatureRequest{
					SourceVariableName:  denormFieldName,
					FeatureVariableName: indexName,
					OutputVariableName:  outputName,
					Variable:            v,
					Step:                step,
					Clustering:          true,
				})
			}
		}
	}

	return features, nil
}

func getD3MIndexField(dr *model.DataResource) int {
	d3mIndexField := -1
	for _, v := range dr.Variables {
		if v.StorageName == model.D3MIndexName {
			d3mIndexField = v.Index
		}
	}

	return d3mIndexField
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

type timeValueCols struct {
	timeCol  string
	valueCol string
}

func getTimeValueCols(dr *model.DataResource) (*timeValueCols, bool) {
	// find the first column marked as a time and the first that is an
	// attribute and use those as series values
	var timeCol string
	var valueCol string
	if dr.ResType == "timeseries" {
		// find a suitable time column and value column - we take the first that works in each
		// case
		for _, v := range dr.Variables {
			for _, r := range v.Role {
				if r == "timeIndicator" && timeCol == "" {
					timeCol = v.StorageName
				}
				if r == "attribute" && valueCol == "" {
					valueCol = v.StorageName
				}
			}
		}
		if timeCol != "" && valueCol != "" {
			return &timeValueCols{
				timeCol:  timeCol,
				valueCol: valueCol,
			}, true
		}
	}
	return nil, false
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

func getRelativePath(rootPath string, filePath string) string {
	relativePath := strings.TrimPrefix(filePath, rootPath)
	relativePath = strings.TrimPrefix(relativePath, "/")

	return relativePath
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

func getFieldIndex(header []string, fieldName string) int {
	for i, f := range header {
		if f == fieldName {
			return i
		}
	}

	return -1
}
