package task

import (
	"encoding/csv"
	"io"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil-compute/model"
	"github.com/unchartedsoftware/distil-compute/pipeline"
	"github.com/unchartedsoftware/distil-compute/primitive/compute"
	"github.com/unchartedsoftware/distil-compute/primitive/compute/description"
	ingestMetadata "github.com/unchartedsoftware/distil-ingest/metadata"
	"github.com/unchartedsoftware/distil/api/env"
	apiModel "github.com/unchartedsoftware/distil/api/model"
)

const lineCount = 100

type primitiveSubmitter interface {
	submit(datasetURIs []string, pipelineDesc *pipeline.PipelineDescription) (string, error)
}

// Join will make all your dreams come true.
func Join(datasetLeft string, datasetRight string, colLeft string, colRight string,
	varsLeft []*model.Variable, varsRight []*model.Variable) (*apiModel.FilteredData, error) {
	cfg, err := env.LoadConfig()
	if err != nil {
		return nil, err
	}
	return join(datasetLeft, datasetRight, colLeft, colRight, varsLeft, varsRight, defaultSubmitter{}, &cfg)
}

func join(datasetLeft string, datasetRight string, colLeft string, colRight string,
	varsLeft []*model.Variable, varsRight []*model.Variable, submitter primitiveSubmitter,
	config *env.Config) (*apiModel.FilteredData, error) {

	// create & submit the solution request
	pipelineDesc, err := description.CreateJoinPipeline("Join Preview", "Join to be reviewed by user", colLeft, colRight)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create join pipeline")
	}

	// returns a URI pointing to the merged CSV file
	resultURI, err := submitter.submit([]string{datasetLeft, datasetRight}, pipelineDesc)
	if err != nil {
		return nil, errors.Wrap(err, "unable to run join pipeline")
	}

	csvFile, err := os.Open(strings.TrimPrefix(resultURI, "file://"))
	if err != nil {
		return nil, errors.Wrap(err, "failed to open raw data file")
	}
	defer csvFile.Close()

	// create a new dataset from the merged CSV file
	leftName := path.Base(path.Dir(strings.TrimPrefix(datasetLeft, "file://")))
	rightName := path.Base(path.Dir(strings.TrimPrefix(datasetRight, "file://")))
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
	for _, varName := range varNames {
		v, ok := leftVarsMap[varName]
		if !ok {
			v, ok = rightVarsMap[varName]
			if !ok {
				return nil, errors.Errorf("can't find data for result var %s", varName)
			}
		}
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

	metadata := model.NewMetadata(datasetName, datasetName, datasetName)
	dataResource := model.NewDataResource("0", compute.D3MResourceType, []string{compute.D3MResourceFormat})

	mergedVariables, err := createMergedVariables(fields, varsLeft, varsRight)
	dataResource.Variables = mergedVariables

	metadata.DataResources = []*model.DataResource{dataResource}

	// save the metadata to the output dataset path
	outputPath := path.Join(config.TmpDataPath, datasetName, compute.D3MDataFolder)
	err = os.MkdirAll(outputPath, 0774)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create join dataset dir structure")
	}

	outputFilePath := path.Join(outputPath, compute.D3MDataSchema)
	err = ingestMetadata.WriteSchema(metadata, outputFilePath) // may write out augmented data structure
	if err != nil {
		return nil, errors.Wrap(err, "failed to write schema")
	}

	// copy the csv data to the output dataset path
	csvDestPath := path.Join(outputPath, compute.D3MLearningData)
	out, err := os.Create(csvDestPath)
	if err != nil {
		return nil, errors.Wrap(err, "unable to open destination")
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()

	csvFile.Seek(0, 0)
	if _, err = io.Copy(out, csvFile); err != nil {
		return nil, errors.Wrap(err, "unable to copy data")
	}
	err = out.Sync()
	if err != nil {
		return nil, errors.Wrap(err, "unable to finalize copy")
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

	csvFile.Seek(0, 0)
	reader := csv.NewReader(csvFile)
	_, err := reader.Read()
	if err != nil {
		return nil, errors.Wrap(err, "failed to read header line")
	}

	data.Values = [][]interface{}{}
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
		typedRow := make([]interface{}, len(row))
		rowErrored := false
		for i, value := range row {
			varType := variables[i].Type
			if model.IsNumerical(varType) {
				if model.IsFloatingPoint(varType) {
					typedRow[i], err = strconv.ParseFloat(value, 64)
					if err != nil {
						errors.Wrapf(err, "failed conversion for row %d", i)
						rowErrored = true
						break
					}
				} else {
					typedRow[i], err = strconv.ParseInt(value, 10, 64)
					if err != nil {
						errors.Wrapf(err, "failed conversion for row %d", i)
						rowErrored = true
						break
					}
				}
			} else {
				typedRow[i] = value
			}
		}
		if rowErrored {
			continue
		}

		data.Values = append(data.Values, typedRow)
	}

	data.NumRows = len(data.Values)

	return data, nil
}
