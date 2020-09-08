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
	"encoding/csv"
	"io"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	ingestMetadata "github.com/uncharted-distil/distil-compute/metadata"
	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil-compute/primitive/compute"
	"github.com/uncharted-distil/distil-compute/primitive/compute/description"
	"github.com/uncharted-distil/distil/api/env"
	apiModel "github.com/uncharted-distil/distil/api/model"
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
	DatasetID     string
	DatasetFolder string
	DatasetSource ingestMetadata.DatasetSource
}

// Join will make all your dreams come true.
func Join(joinLeft *JoinSpec, joinRight *JoinSpec, varsLeft []*model.Variable, varsRight []*model.Variable, rightOrigin *model.DatasetOrigin) (*apiModel.FilteredData, error) {
	cfg, err := env.LoadConfig()
	if err != nil {
		return nil, err
	}
	return join(joinLeft, joinRight, varsLeft, varsRight, rightOrigin, defaultSubmitter{}, &cfg)
}

func join(joinLeft *JoinSpec, joinRight *JoinSpec, varsLeft []*model.Variable,
	varsRight []*model.Variable, rightOrigin *model.DatasetOrigin, submitter primitiveSubmitter,
	config *env.Config) (*apiModel.FilteredData, error) {
	// put the vars into a map for quick lookup
	leftVarsMap := createVarMap(varsLeft, true, true)
	rightVarsMap := createVarMap(varsRight, true, true)
	searchResult := ""
	provenance := ""
	if rightOrigin != nil {
		provenance = rightOrigin.Provenance
		searchResult = rightOrigin.SearchResult
	}

	// create & submit the solution request
	pipelineDesc, err := description.CreateDatamartAugmentPipeline("Join Preview",
		"Join to be reviewed by user", searchResult, provenance)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create join pipeline")
	}

	datasetLeftURI := env.ResolvePath(joinLeft.DatasetSource, joinLeft.DatasetFolder)

	// returns a URI pointing to the merged CSV file
	resultURI, err := submitter.submit([]string{datasetLeftURI}, pipelineDesc)
	if err != nil {
		return nil, errors.Wrap(err, "unable to run join pipeline")
	}

	csvFile, err := os.Open(strings.TrimPrefix(resultURI, "file://"))
	if err != nil {
		return nil, errors.Wrap(err, "failed to open raw data file")
	}
	defer csvFile.Close()

	// create a new dataset from the merged CSV file
	leftName := joinLeft.DatasetID
	rightName := joinRight.DatasetID
	datasetName := strings.Join([]string{leftName, rightName}, "-")
	storageName := model.NormalizeDatasetID(datasetName)
	mergedVariables, err := createDatasetFromCSV(config, csvFile, datasetName, storageName, leftVarsMap, rightVarsMap)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create dataset from result CSV")
	}

	// return some of the data for the client to preview
	data, err := createFilteredData(csvFile, mergedVariables, lineCount)
	if err != nil {
		return nil, err
	}

	return data, nil
}

type defaultSubmitter struct{}

func (defaultSubmitter) submit(datasetURIs []string, pipelineDesc *description.FullySpecifiedPipeline) (string, error) {
	return submitPipeline(datasetURIs, pipelineDesc)
}

func createVarMap(vars []*model.Variable, useDisplayName bool, keepOnlyDataVars bool) map[string]*model.Variable {
	varsMap := map[string]*model.Variable{}
	for _, v := range vars {
		if !model.IsTA2Field(v.DistilRole, v.SelectedRole) && keepOnlyDataVars {
			continue
		}
		name := v.Name
		if useDisplayName {
			name = v.DisplayName
		}
		varsMap[name] = v
	}
	return varsMap
}

func createMergedVariables(varNames []string, leftVarsMap map[string]*model.Variable, rightVarsMap map[string]*model.Variable) ([]*model.Variable, error) {
	mergedVariables := []*model.Variable{}
	for i, varName := range varNames {
		v, ok := leftVarsMap[varName]
		if !ok {
			v, ok = rightVarsMap[varName]
			if !ok {
				// variable is probably an aggregation
				// create a new variable and default type to string
				// ingest process should be able to provide better info
				v = model.NewVariable(i, varName, varName, varName, model.UnknownType,
					model.UnknownType, "", []string{"attribute"}, "data", nil, mergedVariables, false)
			} else {
				// map any distil types (country, city, etc.) back to LL schema types since we are
				// persisting as an LL dataset
				if v.OriginalType != "" {
					v.Type = v.OriginalType
				}
				v.Name = v.DisplayName
				v.OriginalVariable = v.DisplayName
			}
		}

		v.Index = i
		mergedVariables = append(mergedVariables, v)
	}
	return mergedVariables, nil
}

func createDatasetFromCSV(config *env.Config, csvFile *os.File, datasetName string, storageName string,
	leftVarsMap map[string]*model.Variable, rightVarsMap map[string]*model.Variable) ([]*model.Variable, error) {

	reader := csv.NewReader(csvFile)
	fields, err := reader.Read()
	if err != nil {
		return nil, errors.Wrap(err, "failed to read header line")
	}

	metadata := model.NewMetadata(datasetName, datasetName, datasetName, storageName)
	dataResource := model.NewDataResource(compute.DefaultResourceID, compute.D3MResourceType, map[string][]string{compute.D3MResourceFormat: {"csv"}})

	mergedVariables, err := createMergedVariables(fields, leftVarsMap, rightVarsMap)
	if err != nil {
		return nil, err
	}
	dataResource.Variables = mergedVariables

	metadata.DataResources = []*model.DataResource{dataResource}

	outputPath := env.ResolvePath(ingestMetadata.Augmented, datasetName)

	// create dest csv file
	csvDestFolder := path.Join(outputPath, compute.D3MDataFolder)
	err = os.MkdirAll(path.Join(outputPath, compute.D3MDataFolder), os.ModePerm)
	if err != nil {
		return nil, errors.Wrapf(err, "unabled to created dir %s", csvDestFolder)
	}
	csvDestPath := path.Join(csvDestFolder, compute.D3MLearningData)
	out, err := os.Create(csvDestPath)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to open destination %s", csvDestPath)
	}
	defer out.Close()

	// save the metadata to the output dataset path
	err = os.MkdirAll(outputPath, os.ModePerm)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create join dataset dir structure")
	}

	// write out the metadata
	metadataDestPath := path.Join(outputPath, compute.D3MDataSchema)
	relativePath := getRelativePath(path.Dir(metadataDestPath), csvDestPath)
	dataResource.ResPath = relativePath
	err = datasetStorage.WriteMetadata(metadataDestPath, metadata, true)
	if err != nil {
		return nil, errors.Wrap(err, "failed to write schema")
	}

	// write out csv rows, ignoring the first column (contains dataframe index)
	writer := csv.NewWriter(out)
	defer writer.Flush()

	err = writer.Write(fields) // header row
	if err != nil {
		return nil, errors.Wrap(err, "failed to write header row")
	}

	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			// skip malformed input for now
			continue
		}
		writer.Write(row)
	}

	return mergedVariables, nil
}

func createFilteredData(csvFile *os.File, variables []*model.Variable, lineCount int) (*apiModel.FilteredData, error) {
	data := &apiModel.FilteredData{}

	data.Columns = []*apiModel.Column{}
	for _, variable := range variables {
		data.Columns = append(data.Columns, &apiModel.Column{
			Label: variable.DisplayName,
			Key:   variable.Name,
			Type:  variable.Type,
		})
	}

	data.Values = [][]*apiModel.FilteredDataValue{}

	_, err := csvFile.Seek(0, 0)
	if err != nil {
		return nil, errors.Wrap(err, "failed to reset file on read")
	}

	// write the header
	reader := csv.NewReader(csvFile)

	// discard header
	_, err = reader.Read()
	if err != nil {
		return nil, errors.Wrap(err, "failed to read header line")
	}

	errorCount := 0
	discardCount := 0
	for i := 0; i < lineCount; i++ {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			// skip malformed input for now
			errors.Wrap(err, "failed to parse joined csv row")
			continue
		}

		// convert row values to schema type
		// rows that are malformed are discarded
		typedRow := make([]*apiModel.FilteredDataValue, len(row))
		var rowError error
		for j := 0; j < len(row); j++ {
			varType := variables[j].Type
			typedRow[j] = &apiModel.FilteredDataValue{}
			if model.IsNumerical(varType) {
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
