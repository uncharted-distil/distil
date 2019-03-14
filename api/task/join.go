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
	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil-compute/pipeline"
	"github.com/uncharted-distil/distil-compute/primitive/compute"
	"github.com/uncharted-distil/distil-compute/primitive/compute/description"
	ingestMetadata "github.com/uncharted-distil/distil-ingest/metadata"
	"github.com/uncharted-distil/distil/api/env"
	apiModel "github.com/uncharted-distil/distil/api/model"
	log "github.com/unchartedsoftware/plog"
)

const (
	lineCount         = 100
	maxReportedErrors = 50
)

type primitiveSubmitter interface {
	submit(datasetURIs []string, pipelineDesc *pipeline.PipelineDescription) (string, error)
}

// JoinSpec stores information for one side of a join operation.
type JoinSpec struct {
	DatasetID     string
	DatasetFolder string
	DatasetSource ingestMetadata.DatasetSource
	Column        string
}

// Join will make all your dreams come true.
func Join(joinLeft *JoinSpec, joinRight *JoinSpec, varsLeft []*model.Variable, varsRight []*model.Variable) (*apiModel.FilteredData, error) {
	cfg, err := env.LoadConfig()
	if err != nil {
		return nil, err
	}
	return join(joinLeft, joinRight, varsLeft, varsRight, defaultSubmitter{}, &cfg)
}

func join(joinLeft *JoinSpec, joinRight *JoinSpec, varsLeft []*model.Variable, varsRight []*model.Variable, submitter primitiveSubmitter,
	config *env.Config) (*apiModel.FilteredData, error) {

	// create & submit the solution request
	pipelineDesc, err := description.CreateJoinPipeline("Join Preview", "Join to be reviewed by user", joinLeft.Column, joinRight.Column, 0.8)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create join pipeline")
	}

	datasetLeftURI := env.ResolvePath(joinLeft.DatasetSource, joinLeft.DatasetFolder)
	datasetRightURI := env.ResolvePath(joinRight.DatasetSource, joinRight.DatasetFolder)

	// returns a URI pointing to the merged CSV file
	resultURI, err := submitter.submit([]string{datasetLeftURI, datasetRightURI}, pipelineDesc)
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
	mergedVariables, err := createDatasetFromCSV(config, csvFile, datasetName, varsLeft, varsRight)
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

func (defaultSubmitter) submit(datasetURIs []string, pipelineDesc *pipeline.PipelineDescription) (string, error) {
	return submitPipeline(datasetURIs, pipelineDesc)
}

func createVarMap(vars []*model.Variable) map[string]*model.Variable {
	varsMap := map[string]*model.Variable{}
	for _, v := range vars {
		varsMap[v.Name] = v
	}
	return varsMap
}

func createMergedVariables(varNames []string, varsLeft []*model.Variable, varsRight []*model.Variable) ([]*model.Variable, error) {
	// put the vars into a map for quick lookup
	leftVarsMap := createVarMap(varsLeft)
	rightVarsMap := createVarMap(varsRight)

	mergedVariables := []*model.Variable{}
	for i, varName := range varNames {
		v, ok := leftVarsMap[varName]
		if !ok {
			v, ok = rightVarsMap[varName]
			if !ok {
				return nil, errors.Errorf("can't find data for result var \"%s\"", varName)
			}
		}
		v.Index = i
		mergedVariables = append(mergedVariables, v)
	}
	return mergedVariables, nil
}

func createDatasetFromCSV(config *env.Config, csvFile *os.File, datasetName string, varsLeft []*model.Variable, varsRight []*model.Variable) ([]*model.Variable, error) {
	reader := csv.NewReader(csvFile)
	fields, err := reader.Read()
	if err != nil {
		return nil, errors.Wrap(err, "failed to read header line")
	}
	// first column will be the dataframe index which we should ignore
	fields = fields[1:]

	metadata := model.NewMetadata(datasetName, datasetName, datasetName, model.NormalizeDatasetID(datasetName))
	dataResource := model.NewDataResource("learningData", compute.D3MResourceType, []string{compute.D3MResourceFormat})

	mergedVariables, err := createMergedVariables(fields, varsLeft, varsRight)
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
	err = ingestMetadata.WriteSchema(metadata, metadataDestPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to write schema")
	}

	// write out csv rows, ignoring the first column (contains dataframe index)
	writer := csv.NewWriter(out)
	defer writer.Flush()

	writer.Write(fields) // header row
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			// skip malformed input for now
			errors.Wrap(err, "failed to parse joined csv row")
			continue
		}
		writer.Write(row[1:])
	}

	return mergedVariables, nil
}

func createFilteredData(csvFile *os.File, variables []*model.Variable, lineCount int) (*apiModel.FilteredData, error) {
	data := &apiModel.FilteredData{}

	data.Columns = []apiModel.Column{}
	for _, variable := range variables {
		data.Columns = append(data.Columns, apiModel.Column{
			Label: variable.DisplayName,
			Key:   variable.Name,
			Type:  variable.Type,
		})
	}

	data.Values = [][]interface{}{}

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

		// convert row values to schema type, ignoring first column (contains dataframe index)
		// rows that are malformed are discarded
		typedRow := make([]interface{}, len(row)-1)
		var rowError error
		for j := 1; j < len(row); j++ {
			varType := variables[j-1].Type
			if model.IsNumerical(varType) {
				if model.IsFloatingPoint(varType) {
					typedRow[j-1], err = strconv.ParseFloat(row[j], 64)
					if err != nil {
						rowError = errors.Wrapf(err, "failed conversion for row %d", i)
						errorCount++
						break
					}
				} else {
					typedRow[j-1], err = strconv.ParseInt(row[j], 10, 64)
					if err != nil {
						flt, err := strconv.ParseFloat(row[j], 64)
						if err != nil {
							rowError = errors.Wrapf(err, "failed conversion for row %d", i)
							errorCount++
							break
						}
						typedRow[j-1] = int64(flt)
					}
				}
			} else {
				typedRow[j-1] = row[j]
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
