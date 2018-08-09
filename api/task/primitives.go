package task

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/unchartedsoftware/distil-ingest/metadata"
	"github.com/unchartedsoftware/distil-ingest/rest"
	"github.com/unchartedsoftware/distil/api/compute"
	"github.com/unchartedsoftware/distil/api/compute/description"
	"github.com/unchartedsoftware/distil/api/compute/result"
	"github.com/unchartedsoftware/distil/api/pipeline"
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
	Variable            *metadata.Variable
	Step                *pipeline.PipelineDescription
}

// SetClient sets the compute client to use when invoking primitives.
func SetClient(computeClient *compute.Client) {
	client = computeClient
}

func submitPrimitive(dataset string, step *pipeline.PipelineDescription) (string, error) {
	request := compute.NewExecPipelineRequest(dataset, step)

	err := request.Dispatch(client)
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

// ClassifyPrimitive will classify the dataset using a primitive.
func ClassifyPrimitive(index string, dataset string, config *IngestTaskConfig) error {
	// create & submit the solution request
	pip, err := description.CreateSimonPipeline("says", "")
	if err != nil {
		return errors.Wrap(err, "unable to create Simon pipeline")
	}

	datasetURI, err := submitPrimitive(dataset, pip)
	if err != nil {
		return errors.Wrap(err, "unable to run Simon pipeline")
	}

	// parse primitive response (variable,probabilities,labels)
	res, err := result.ParseResultCSV(datasetURI)
	if err != nil {
		return errors.Wrap(err, "unable to parse Simon pipeline result")
	}

	// First row is header, then all other rows are col index, types, probabilities.
	probabilities := make([][]float64, len(res)-1)
	labels := make([][]string, len(res)-1)
	for i, v := range res {
		if i > 0 {
			colIndex, err := strconv.ParseInt(v[0].(string), 10, 64)
			if err != nil {
				return err
			}
			labels[colIndex] = toStringArray(v[1].([]interface{}))
			probs, err := toFloat64Array(v[2].([]interface{}))
			if err != nil {
				return err
			}
			probabilities[colIndex] = probs
		}
	}
	classification := &rest.ClassificationResult{
		Path:          datasetURI,
		Labels:        labels,
		Probabilities: probabilities,
	}

	// output the classification in the expected JSON format
	bytes, err := json.MarshalIndent(classification, "", "    ")
	if err != nil {
		return errors.Wrap(err, "unable to serialize classification result")
	}
	// write to file
	err = ioutil.WriteFile(config.getTmpAbsolutePath(config.ClassificationOutputPathRelative), bytes, 0644)
	if err != nil {
		return errors.Wrap(err, "unable to store classification result")
	}

	return nil
}

// RankPrimmitive will rank the dataset using a primitive.
func RankPrimmitive(index string, dataset string, config *IngestTaskConfig) error {
	// create & submit the solution request
	pip, err := description.CreatePCAFeaturesPipeline("harry", "")
	if err != nil {
		return errors.Wrap(err, "unable to create PCA pipeline")
	}

	datasetURI, err := submitPrimitive(dataset, pip)
	if err != nil {
		return errors.Wrap(err, "unable to run PCA pipeline")
	}

	// parse primitive response (col index,importance)
	res, err := result.ParseResultCSV(datasetURI)
	if err != nil {
		return errors.Wrap(err, "unable to parse PCA pipeline result")
	}

	ranks := make([]float64, len(res)-1)
	for i, v := range res {
		if i > 0 {
			colIndex, err := strconv.ParseInt(v[0].(string), 10, 64)
			if err != nil {
				return errors.Wrap(err, "unable to parse PCA col index")
			}
			vInt, err := strconv.ParseFloat(v[1].(string), 64)
			if err != nil {
				return errors.Wrap(err, "unable to parse PCA rank value")
			}
			ranks[colIndex] = vInt
		}
	}

	importance := &rest.ImportanceResult{
		Path:     datasetURI,
		Features: ranks,
	}

	// output the classification in the expected JSON format
	bytes, err := json.MarshalIndent(importance, "", "    ")
	if err != nil {
		return errors.Wrap(err, "unable to serialize ranking result")
	}
	// write to file
	err = ioutil.WriteFile(config.getTmpAbsolutePath(config.RankingOutputPathRelative), bytes, 0644)
	if err != nil {
		return errors.Wrap(err, "unable to store ranking result")
	}

	return nil
}

// SummarizePrimitive will summarize the dataset using a primitive.
func SummarizePrimitive(index string, dataset string, config *IngestTaskConfig) error {
	// create & submit the solution request
	pip, err := description.CreateDukePipeline("wellington", "")
	if err != nil {
		return errors.Wrap(err, "unable to create Duke pipeline")
	}

	datasetURI, err := submitPrimitive(dataset, pip)
	if err != nil {
		return errors.Wrap(err, "unable to run Duke pipeline")
	}

	// parse primitive response (token,probability)
	res, err := result.ParseResultCSV(datasetURI)
	if err != nil {
		return errors.Wrap(err, "unable to parse Duke pipeline result")
	}

	tokens := make([]string, len(res)-1)
	for i, v := range res {
		// skip the header
		if i > 0 {
			token, ok := v[0].(string)
			if !ok {
				return errors.Wrap(err, "unable to parse Duke token")
			}
			tokens[i-1] = token
		}
	}

	sum := &rest.SummaryResult{
		Summary: strings.Join(tokens, ", "),
	}

	// output the classification in the expected JSON format
	bytes, err := json.MarshalIndent(sum, "", "    ")
	if err != nil {
		return errors.Wrap(err, "unable to serialize summary result")
	}
	// write to file
	err = ioutil.WriteFile(config.getTmpAbsolutePath(config.SummaryOutputPathRelative), bytes, 0644)
	if err != nil {
		return errors.Wrap(err, "unable to store summary result")
	}

	return nil
}

// FeaturizePrimitive will featurize the dataset fields using a primitive.
func FeaturizePrimitive(index string, dataset string, config *IngestTaskConfig) error {
	// create required folders for outputPath
	createContainingDirs(config.getTmpAbsolutePath(config.FeaturizationOutputDataRelative))
	createContainingDirs(config.getTmpAbsolutePath(config.FeaturizationOutputSchemaRelative))

	// load metadata from original schema
	meta, err := metadata.LoadMetadataFromOriginalSchema(config.getAbsolutePath(config.SchemaPathRelative))
	if err != nil {
		return errors.Wrap(err, "unable to load original schema file")
	}
	mainDR := meta.GetMainDataResource()

	// add feature variables
	features, err := getClusterVariables(meta, "_feature_")
	if err != nil {
		return errors.Wrap(err, "unable to get feature variables")
	}

	d3mIndexField := getD3MIndexField(mainDR)

	// open the input file
	dataPath := path.Join(config.ContainerDataPath, mainDR.ResPath)
	lines, err := readCSVFile(dataPath, config.HasHeader)
	if err != nil {
		return errors.Wrap(err, "error reading raw data")
	}

	// add the cluster data to the raw data
	for _, f := range features {
		mainDR.Variables = append(mainDR.Variables, f.Variable)

		lines, err = appendFeature(dataset, d3mIndexField, config.HasHeader, f, lines)
		if err != nil {
			return errors.Wrap(err, "error appending feature data")
		}
	}

	// initialize csv writer
	output := &bytes.Buffer{}
	writer := csv.NewWriter(output)

	// output the header
	header := make([]string, len(mainDR.Variables))
	for _, v := range mainDR.Variables {
		header[v.Index] = v.Name
	}
	err = writer.Write(header)
	if err != nil {
		return errors.Wrap(err, "error storing feature header")
	}

	for _, line := range lines {
		err = writer.Write(line)
		if err != nil {
			return errors.Wrap(err, "error storing feature output")
		}
	}

	// output the data with the new feature
	writer.Flush()
	err = ioutil.WriteFile(config.getTmpAbsolutePath(config.FeaturizationOutputDataRelative), output.Bytes(), 0644)
	if err != nil {
		return errors.Wrap(err, "error writing feature output")
	}

	mainDR.ResPath = config.FeaturizationOutputDataRelative

	// write the new schema to file
	err = meta.WriteSchema(config.getTmpAbsolutePath(config.FeaturizationOutputSchemaRelative))
	if err != nil {
		return errors.Wrap(err, "unable to store feature schema")
	}

	return nil
}

// ClusterPrimitive will cluster the dataset fields using a primitive.
func ClusterPrimitive(index string, dataset string, config *IngestTaskConfig) error {
	// create required folders for outputPath
	createContainingDirs(config.getTmpAbsolutePath(config.ClusteringOutputDataRelative))
	createContainingDirs(config.getTmpAbsolutePath(config.ClusteringOutputSchemaRelative))

	// load metadata from original schema
	meta, err := metadata.LoadMetadataFromOriginalSchema(config.getAbsolutePath(config.SchemaPathRelative))
	if err != nil {
		return errors.Wrap(err, "unable to load original schema file")
	}
	mainDR := meta.GetMainDataResource()

	// add feature variables
	features, err := getFeatureVariables(meta, "_cluster_")
	if err != nil {
		return errors.Wrap(err, "unable to get cluster variables")
	}

	d3mIndexField := getD3MIndexField(mainDR)

	// open the input file
	dataPath := path.Join(config.ContainerDataPath, mainDR.ResPath)
	lines, err := readCSVFile(dataPath, config.HasHeader)
	if err != nil {
		return errors.Wrap(err, "error reading raw data")
	}

	// add the cluster data to the raw data
	for _, f := range features {
		mainDR.Variables = append(mainDR.Variables, f.Variable)

		lines, err = appendFeature(dataset, d3mIndexField, config.HasHeader, f, lines)
		if err != nil {
			return errors.Wrap(err, "error appending clustered data")
		}
	}

	// initialize csv writer
	output := &bytes.Buffer{}
	writer := csv.NewWriter(output)

	// output the header
	header := make([]string, len(mainDR.Variables))
	for _, v := range mainDR.Variables {
		header[v.Index] = v.Name
	}
	err = writer.Write(header)
	if err != nil {
		return errors.Wrap(err, "error storing clustered header")
	}

	for _, line := range lines {
		err = writer.Write(line)
		if err != nil {
			return errors.Wrap(err, "error storing clustered output")
		}
	}

	// output the data with the new feature
	writer.Flush()
	err = ioutil.WriteFile(config.getTmpAbsolutePath(config.ClusteringOutputDataRelative), output.Bytes(), 0644)
	if err != nil {
		return errors.Wrap(err, "error writing clustered output")
	}

	mainDR.ResPath = config.ClusteringOutputDataRelative

	// write the new schema to file
	err = meta.WriteSchema(config.getTmpAbsolutePath(config.ClusteringOutputSchemaRelative))
	if err != nil {
		return errors.Wrap(err, "unable to store cluster schema")
	}

	return nil
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
		if f == feature.FeatureVariableName {
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
		}
	}

	return lines, nil
}

func getFeatureVariables(meta *metadata.Metadata, prefix string) ([]*FeatureRequest, error) {
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
				v := metadata.NewVariable(len(mainDR.Variables), indexName, "label", v.Name, "string", "string", "", "", []string{"attribute"}, metadata.VarRoleMetadata, nil, mainDR.Variables, false)

				// create the required pipeline
				step, err := description.CreateCrocPipeline("leather", "", []string{v.Name}, []string{indexName})
				if err != nil {
					return nil, errors.Wrap(err, "unable to create step pipeline")
				}

				features = append(features, &FeatureRequest{
					SourceVariableName:  denormFieldName,
					FeatureVariableName: indexName,
					Variable:            v,
					Step:                step,
				})
			}
		}
	}

	return features, nil
}

func getClusterVariables(meta *metadata.Metadata, prefix string) ([]*FeatureRequest, error) {
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
				v := metadata.NewVariable(len(mainDR.Variables), indexName, "group", v.Name, "string", "string", "", "", []string{"attribute"}, metadata.VarRoleMetadata, nil, mainDR.Variables, false)

				// create the required pipeline
				var step *pipeline.PipelineDescription
				var err error
				if res.CanBeFeaturized() {
					step, err = description.CreateUnicornPipeline("horned", "", []string{v.Name}, []string{indexName})
				} else {
					step, err = description.CreateSlothPipeline("leaf", "", []string{v.Name}, []string{indexName})
				}
				if err != nil {
					return nil, errors.Wrap(err, "unable to create step pipeline")
				}

				features = append(features, &FeatureRequest{
					SourceVariableName:  denormFieldName,
					FeatureVariableName: indexName,
					Variable:            v,
					Step:                step,
				})
			}
		}
	}

	return features, nil
}

func getD3MIndexField(dr *metadata.DataResource) int {
	d3mIndexField := -1
	for _, v := range dr.Variables {
		if v.Name == metadata.D3MIndexName {
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

func getDataResource(meta *metadata.Metadata, resID string) *metadata.DataResource {
	// main data resource has d3m index variable
	for _, dr := range meta.DataResources {
		if dr.ResID == resID {
			return dr
		}
	}

	return nil
}
