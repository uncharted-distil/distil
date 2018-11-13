package task

import (
	"context"
	"encoding/csv"
	"fmt"
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
	"github.com/unchartedsoftware/distil-compute/primitive/compute/result"

	"github.com/unchartedsoftware/distil/api/env"
	"github.com/unchartedsoftware/distil/api/model"
	"github.com/unchartedsoftware/distil/api/pipeline"
)

const (
	denormFieldName  = "filename"
	useMockTA2System = true
)

var (
	client       *compute.Client
	inputRootDir string
)

// FeatureRequest captures the properties of a request to a primitive.
type FeatureRequest struct {
	SourceVariableName  string
	FeatureVariableName string
	OutputVariableName  string
	Variable            *model.Variable
	Step                *pipeline.PipelineDescription
}

// SetClient sets the compute client to use when invoking primitives.
func SetClient(computeClient *compute.Client) {
	client = computeClient
}

func submitPrimitive(dataset string, step *pipeline.PipelineDescription) (string, error) {

	config, err := env.LoadConfig()
	if err != nil {
		return "", errors.Wrap(err, "unable to load config")
	}

	// create a reference to the original data path
	datasetInputDir := path.Join(config.D3MInputDirRoot, dataset, "TRAIN", "dataset_TRAIN")

	if config.UseTA2Runner {
		res, err := client.ExecutePipeline(context.Background(), datasetInputDir, step)
		if err != nil {
			return "", errors.Wrap(err, "unable to dispatch mocked pipeline")
		}
		resultURI := strings.Replace(res.ResultURI, "file://", "", -1)
		return resultURI, nil
	}

	request := compute.NewExecPipelineRequest(dataset, step)

	err = request.Dispatch(client)
	if err != nil {
		return "", errors.Wrap(err, "unable to dispatch pipeline")
	}

	// listen for completion
	var errPipeline error
	var datasetURI string
	err = request.Listen(func(status compute.ExecPipelineStatus) {
		// check for error
		if status.Error != nil {
			errPipeline = status.Error
		}

		if status.Progress == compute.RequestCompletedStatus {
			datasetURI = status.ResultURI
		}
	})
	if err != nil {
		return "", errors.Wrap(err, "unable to listen to pipeline")
	}

	if errPipeline != nil {
		return "", errors.Wrap(errPipeline, "error executing pipeline")
	}

	datasetURI = strings.Replace(datasetURI, "file://", "", -1)

	return datasetURI, nil
}

// TargetRankPrimitive will rank the dataset relative to a target variable using
// a primitive.
func TargetRankPrimitive(dataset string, target string, features []*model.Variable) (map[string]float64, error) {
	// create & submit the solution request
	pip, err := description.CreateTargetRankingPipeline("roger", "", target, features)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create ranking pipeline")
	}

	datasetURI, err := submitPrimitive(dataset, pip)
	if err != nil {
		return nil, errors.Wrap(err, "unable to run ranking pipeline")
	}

	// parse primitive response (col index,importance)
	res, err := result.ParseResultCSV(datasetURI)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse ranking pipeline result")
	}

	ranks := make(map[string]float64)
	for i, v := range res {
		if i > 0 {
			key, ok := v[2].(string)
			if !ok {
				return nil, fmt.Errorf("unable to parse rank key")
			}
			rank, err := strconv.ParseFloat(v[3].(string), 64)
			if err != nil {
				return nil, errors.Wrap(err, "unable to parse rank value")
			}
			ranks[key] = rank
		}
	}

	return ranks, nil
}

func readCSVFile(filename string, hasHeader bool) ([][]string, error) {
	// open the file
	csvFile, err := os.Open(filename)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open data file")
	}
	defer csvFile.Close()
	reader := csv.NewReader(csvFile)

	lines := make([][]string, 0)

	// skip the header as needed
	if hasHeader {
		_, err = reader.Read()
		if err != nil {
			return nil, errors.Wrap(err, "failed to read header from file")
		}
	}

	// read the raw data
	for {
		line, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, errors.Wrap(err, "failed to read line from file")
		}

		lines = append(lines, line)
	}

	return lines, nil
}

func appendFeature(dataset string, d3mIndexField int, hasHeader bool, feature *FeatureRequest, lines [][]string) ([][]string, error) {
	datasetURI, err := submitPrimitive(dataset, feature.Step)
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
			labels := v[labelIndex].(string)
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

func getFeatureVariables(meta *model.Metadata, prefix string) ([]*FeatureRequest, error) {
	mainDR := meta.GetMainDataResource()
	features := make([]*FeatureRequest, 0)
	for _, v := range mainDR.Variables {
		if v.RefersTo != nil && v.RefersTo["resID"] != nil {
			// get the refered DR
			resID := v.RefersTo["resID"].(string)

			res := getDataResource(meta, resID)

			// check if needs to be featurized
			if res.CanBeFeaturized() {
				// create the new resource to hold the featured output
				indexName := fmt.Sprintf("%s%s", prefix, v.Name)

				// add the feature variable
				v := model.NewVariable(len(mainDR.Variables), indexName, "label", v.Name, "string", "string", []string{"attribute"}, model.VarRoleMetadata, nil, mainDR.Variables, false)

				// create the required pipeline
				step, err := description.CreateCrocPipeline("leather", "", []string{denormFieldName}, []string{indexName})
				if err != nil {
					return nil, errors.Wrap(err, "unable to create step pipeline")
				}

				features = append(features, &FeatureRequest{
					SourceVariableName:  denormFieldName,
					FeatureVariableName: indexName,
					OutputVariableName:  fmt.Sprintf("%s_object_label", indexName),
					Variable:            v,
					Step:                step,
				})
			}
		}
	}

	return features, nil
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
			if res.CanBeFeaturized() || res.ResType == "timeseries" {
				// create the new resource to hold the featured output
				indexName := fmt.Sprintf("%s%s", prefix, v.Name)

				// add the feature variable
				v := model.NewVariable(len(mainDR.Variables), indexName, "group", v.Name, "string", "string", []string{"attribute"}, model.VarRoleMetadata, nil, mainDR.Variables, false)

				// create the required pipeline
				var step *pipeline.PipelineDescription
				var err error
				outputName := ""
				if res.CanBeFeaturized() {
					step, err = description.CreateUnicornPipeline("horned",
						"clustering based on resnet-50 detected objects", []string{denormFieldName}, []string{indexName})
					outputName = unicornResultFieldName
				} else {
					if colNames, ok := getTimeValueCols(res); ok {
						step, err = description.CreateSlothPipeline("time series clustering",
							"k-means time series clustering", colNames.timeCol, colNames.valueCol, res.Variables)
						outputName = slothResultFieldName
					}
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
				})
			}
		}
	}

	return features, nil
}

func getD3MIndexField(dr *model.DataResource) int {
	d3mIndexField := -1
	for _, v := range dr.Variables {
		if v.Name == model.D3MIndexName {
			d3mIndexField = v.Index
		}
	}

	return d3mIndexField
}

func toStringArray(in []interface{}) []string {
	strArr := make([]string, 0)
	for _, v := range in {
		strArr = append(strArr, v.(string))
	}
	return strArr
}

func toFloat64Array(in []interface{}) ([]float64, error) {
	strArr := make([]float64, 0)
	for _, v := range in {
		strFloat, err := strconv.ParseFloat(v.(string), 64)
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert interface array to float array")
		}
		strArr = append(strArr, strFloat)
	}
	return strArr, nil
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
					timeCol = v.Name
				}
				if r == "attribute" && valueCol == "" {
					valueCol = v.Name
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

func mapFields(meta *model.Metadata) map[string]*model.Variable {
	// cycle through each data resource, mapping field names to variables.
	fields := make(map[string]*model.Variable)
	for _, dr := range meta.DataResources {
		for _, v := range dr.Variables {
			fields[v.Name] = v
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
