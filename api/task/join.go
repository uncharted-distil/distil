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
	"strconv"
	"strings"

	"github.com/pkg/errors"
	ingestMetadata "github.com/uncharted-distil/distil-compute/metadata"
	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil-compute/primitive/compute"
	"github.com/uncharted-distil/distil-compute/primitive/compute/description"
	"github.com/uncharted-distil/distil/api/env"
	apiModel "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/serialization"
	log "github.com/unchartedsoftware/plog"
)

const (
	lineCount         = 100
	maxReportedErrors = 50
)

type primitiveSubmitter interface {
	submit(datasetURIs []string, pipelineDesc *description.FullySpecifiedPipeline) (string, error)
}

// JoinSpec stores information for one side of a join operation.
type JoinSpec struct {
	DatasetID        string
	DatasetFolder    string
	DatasetSource    ingestMetadata.DatasetSource
	ExistingMetadata *model.Metadata
	UpdatedVariables []*model.Variable
}

// JoinDatamart will make all your dreams come true.
func JoinDatamart(joinLeft *JoinSpec, joinRight *JoinSpec, rightOrigin *model.DatasetOrigin) (string, *apiModel.FilteredData, error) {
	cfg, err := env.LoadConfig()
	if err != nil {
		return "", nil, err
	}
	pipelineDesc, err := description.CreateDatamartAugmentPipeline("Join Preview",
		"Join to be reviewed by user", rightOrigin.SearchResult, rightOrigin.Provenance)
	if err != nil {
		return "", nil, err
	}
	datasetLeftURI := env.ResolvePath(joinLeft.DatasetSource, joinLeft.DatasetFolder)

	return join(joinLeft, joinRight, pipelineDesc, []string{datasetLeftURI}, defaultSubmitter{}, &cfg)
}

// JoinDistil will bring misery.
func JoinDistil(joinLeft *JoinSpec, joinRight *JoinSpec, leftCols []string, rightCols []string, accuracy float32) (string, *apiModel.FilteredData, error) {
	cfg, err := env.LoadConfig()
	if err != nil {
		return "", nil, err
	}

	varsLeftMapUpdated := apiModel.MapVariables(joinLeft.UpdatedVariables, func(variable *model.Variable) string { return variable.Key })
	varsRightMapUpdated := apiModel.MapVariables(joinRight.UpdatedVariables, func(variable *model.Variable) string { return variable.Key })
	leftVars := make([]*model.Variable, len(leftCols))
	rightVars := make([]*model.Variable, len(rightCols))
	for i, v := range leftCols {
		leftVars[i] = varsLeftMapUpdated[v]
	}
	for i, v := range rightCols {
		rightVars[i] = varsRightMapUpdated[v]
	}

	pipelineDesc, err := description.CreateJoinPipeline("Joiner", "Join existing data", leftVars, rightVars, accuracy)
	if err != nil {
		return "", nil, err
	}
	datasetLeftURI := env.ResolvePath(joinLeft.DatasetSource, joinLeft.DatasetFolder)
	datasetRightURI := env.ResolvePath(joinRight.DatasetSource, joinRight.DatasetFolder)

	return join(joinLeft, joinRight, pipelineDesc, []string{datasetLeftURI, datasetRightURI}, defaultSubmitter{}, &cfg)
}

func join(joinLeft *JoinSpec, joinRight *JoinSpec, pipelineDesc *description.FullySpecifiedPipeline,
	datasetURIs []string, submitter primitiveSubmitter, config *env.Config) (string, *apiModel.FilteredData, error) {
	// returns a URI pointing to the merged CSV file
	resultURI, err := submitter.submit(datasetURIs, pipelineDesc)
	if err != nil {
		return "", nil, errors.Wrap(err, "unable to run join pipeline")
	}

	// create a new dataset from the merged CSV file
	csvFilename := strings.TrimPrefix(resultURI, "file://")
	leftName := joinLeft.DatasetID
	rightName := joinRight.DatasetID
	datasetName := strings.Join([]string{leftName, rightName}, "-")
	storageName := model.NormalizeDatasetID(datasetName)
	mergedVariables, err := createDatasetFromCSV(config, csvFilename, datasetName, storageName, joinLeft, joinRight)
	if err != nil {
		return "", nil, errors.Wrap(err, "unable to create dataset from result CSV")
	}

	// return some of the data for the client to preview
	data, err := createFilteredData(csvFilename, mergedVariables, lineCount)
	if err != nil {
		return "", nil, err
	}

	return env.ResolvePath(ingestMetadata.Augmented, datasetName), data, nil
}

type defaultSubmitter struct{}

func (defaultSubmitter) submit(datasetURIs []string, pipelineDesc *description.FullySpecifiedPipeline) (string, error) {
	return submitPipeline(datasetURIs, pipelineDesc, true)
}

func createDatasetFromCSV(config *env.Config, csvFile string, datasetName string, storageName string, joinLeft *JoinSpec, joinRight *JoinSpec) ([]*model.Variable, error) {
	inputData, err := serialization.ResultToInputCSV(csvFile)
	if err != nil {
		return nil, err
	}

	metadata := model.NewMetadata(datasetName, datasetName, datasetName, storageName)
	dataResource := model.NewDataResource(compute.DefaultResourceID, compute.D3MResourceType, map[string][]string{compute.D3MResourceFormat: {"csv"}})

	mergedVariables, referencedResources := joinMetadataVariables(inputData[0], joinLeft.ExistingMetadata, joinRight.ExistingMetadata)
	dataResource.Variables = mergedVariables
	inputData[0] = dataResource.GenerateHeader()

	metadata.DataResources = append(referencedResources, dataResource)

	outputPath := env.ResolvePath(ingestMetadata.Augmented, datasetName)

	rawDataset := &serialization.RawDataset{
		Name:     metadata.Name,
		ID:       metadata.ID,
		Metadata: metadata,
		Data:     inputData,
	}

	err = serialization.WriteDataset(outputPath, rawDataset)
	if err != nil {
		return nil, err
	}

	return mergedVariables, nil
}

func createFilteredData(csvFile string, variables []*model.Variable, lineCount int) (*apiModel.FilteredData, error) {
	datasetStorage := serialization.GetStorage(csvFile)
	inputData, err := datasetStorage.ReadData(csvFile)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read joined data")
	}

	data := &apiModel.FilteredData{}

	data.Columns = map[string]*apiModel.Column{}
	for _, variable := range variables {
		data.Columns[variable.Key] = &apiModel.Column{
			Label: variable.DisplayName,
			Key:   variable.Key,
			Type:  variable.Type,
			Index: len(data.Columns),
		}
	}

	data.Values = [][]*apiModel.FilteredDataValue{}

	// discard header
	inputData = inputData[1:]

	errorCount := 0
	discardCount := 0
	for i := 0; i < lineCount && i < len(inputData); i++ {
		row := inputData[i]

		// convert row values to schema type
		// rows that are malformed are discarded
		typedRow := make([]*apiModel.FilteredDataValue, len(row))
		var rowError error
		for j := 0; j < len(row); j++ {
			varType := variables[j].Type
			typedRow[j] = &apiModel.FilteredDataValue{}
			if model.IsNumerical(varType) && row[j] != "" {
				if model.IsFloatingPoint(varType) {
					typedRow[j].Value, err = strconv.ParseFloat(row[j], 64)
					if err != nil {
						rowError = errors.Wrapf(err, "failed conversion for row %d", i)
						errorCount++
						break
					}
				} else {
					typedRow[j].Value, err = strconv.ParseInt(row[j], 10, 64)
					if err != nil {
						flt, err := strconv.ParseFloat(row[j], 64)
						if err != nil {
							rowError = errors.Wrapf(err, "failed conversion for row %d", i)
							errorCount++
							break
						}
						typedRow[j].Value = int64(flt)
					}
				}
			} else {
				typedRow[j].Value = row[j]
			}
		}
		if rowError != nil {
			discardCount++
			if errorCount < maxReportedErrors {
				log.Warn(rowError)
			} else if errorCount == maxReportedErrors {
				log.Warn("too many errors - logging of remainder surpressed")
			}
			continue
		}
		data.Values = append(data.Values, typedRow)
	}

	if discardCount > 0 {
		log.Warnf("discarded %d rows due to parsing parsing errors", discardCount)
	}

	data.NumRows = len(data.Values)

	return data, nil
}

func denormVariableName(variable *model.Variable) string {
	// The denorm primitive renames columns that refer to resources to "filename".
	// The maps should account for this.
	if variable.RefersTo != nil {
		return "filename"
	}

	return variable.HeaderName
}

func joinMetadataVariables(headerNames []string, leftMetadata *model.Metadata, rightMetadata *model.Metadata) ([]*model.Variable, []*model.DataResource) {
	// map the variables using the header names
	leftMap := apiModel.MapVariables(leftMetadata.GetMainDataResource().Variables, denormVariableName)
	rightMap := apiModel.MapVariables(rightMetadata.GetMainDataResource().Variables, denormVariableName)

	// left dataset takes priority in case of conflict
	mergedVariables := make([]*model.Variable, len(headerNames))
	mergedResources := []*model.DataResource{}
	for i, varName := range headerNames {
		v, ok := leftMap[varName]
		if ok && v.RefersTo != nil {
			mergedResources = append(mergedResources, getDataResource(leftMetadata, v.RefersTo["resID"].(string)))
		} else if !ok {
			v, ok = rightMap[varName]
			if ok && v.RefersTo != nil {
				mergedResources = append(mergedResources, getDataResource(rightMetadata, v.RefersTo["resID"].(string)))
			} else if !ok {
				// variable is probably an aggregation
				// create a new variable and default type to string
				// ingest process should be able to provide better info
				v = model.NewVariable(i, varName, varName, varName, varName, model.UnknownType,
					model.UnknownType, "", []string{"attribute"}, "data", nil, mergedVariables, false)
			}
		}

		v.Index = i
		mergedVariables[i] = v
	}

	return mergedVariables, mergedResources
}
