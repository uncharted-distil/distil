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

var (
	client *compute.Client
)

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

	// First row is header, then all other rows are field name, types, probabilities.
	probabilities := make([][]float64, len(res)-1)
	labels := make([][]string, len(res)-1)
	for i, v := range res {
		if i > 0 {
			labels[i-1] = toStringArray(v[1].([]interface{}))
			res, err := toFloat64Array(v[2].([]interface{}))
			if err != nil {
				return err
			}
			probabilities[i-1] = res
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

	// parse primitive response (variable,importance)
	res, err := result.ParseResultCSV(datasetURI)
	if err != nil {
		return errors.Wrap(err, "unable to parse PCA pipeline result")
	}

	ranks := make([]float64, len(res)-1)
	for i, v := range res {
		if i > 0 {
			vInt, err := strconv.ParseFloat(v[1].(string), 64)
			if err != nil {
				return errors.Wrap(err, "unable to parse PCA rank value")
			}
			ranks[i-1] = vInt
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
	inputFeatureNames := make([]string, 0)
	outputFeatureNames := make([]string, 0)
	for _, v := range mainDR.Variables {
		if v.RefersTo != nil && v.RefersTo["resID"] != nil {
			// get the refered DR
			resID := v.RefersTo["resID"].(string)

			res := getDataResource(meta, resID)

			// check if needs to be featurized
			if res.CanBeFeaturized() {
				// create the new resource to hold the featured output
				indexName := fmt.Sprintf("_feature_%s", v.Name)

				// add the feature variable
				mainDR.AddVariable(indexName, v.Name, "string", []string{"attribute"}, metadata.VarRoleMetadata)

				inputFeatureNames = append(inputFeatureNames, v.Name)
				outputFeatureNames = append(outputFeatureNames, indexName)
			}
		}
	}

	// create & submit the solution request
	if len(inputFeatureNames) > 0 {
		pip, err := description.CreateCrocPipeline("leather", "", inputFeatureNames, outputFeatureNames)
		if err != nil {
			return errors.Wrap(err, "unable to create Croc pipeline")
		}

		datasetURI, err := submitPrimitive(dataset, pip)
		if err != nil {
			return errors.Wrap(err, "unable to run Croc pipeline")
		}

		// parse primitive response (d3mIndex,labels,probabilities)
		res, err := result.ParseResultCSV(datasetURI)
		if err != nil {
			return errors.Wrap(err, "unable to parse Croc pipeline result")
		}

		// build the lookup for the new field
		features := make(map[string]string)
		for i, v := range res {
			// skip header
			if i > 0 {
				d3mIndex := v[0].(string)
				labels := v[1].(string)
				features[d3mIndex] = labels
			}
		}

		dataPath := path.Join(config.ContainerDataPath, mainDR.ResPath)
		csvFile, err := os.Open(dataPath)
		if err != nil {
			return errors.Wrap(err, "failed to open data file")
		}
		defer csvFile.Close()
		reader := csv.NewReader(csvFile)

		// initialize csv writer
		output := &bytes.Buffer{}
		writer := csv.NewWriter(output)

		// write the header as needed
		if config.HasHeader {
			header := make([]string, len(mainDR.Variables))
			for _, v := range mainDR.Variables {
				header[v.Index] = v.Name
			}
			err = writer.Write(header)
			if err != nil {
				return errors.Wrap(err, "error writing header to output")
			}
		}

		// skip header
		if config.HasHeader {
			_, err = reader.Read()
			if err != nil {
				return errors.Wrap(err, "failed to read header from file")
			}
		}

		d3mIndexField := -1
		for _, v := range mainDR.Variables {
			if v.Name == metadata.D3MIndexName {
				d3mIndexField = v.Index
			}
		}

		// read the raw data and add the features column to the output
		for {
			line, err := reader.Read()
			if err == io.EOF {
				break
			} else if err != nil {
				return errors.Wrap(err, "failed to read line from file")
			}

			d3mIndex := line[d3mIndexField]
			feature := features[d3mIndex]
			line = append(line, feature)

			writer.Write(line)
			if err != nil {
				return errors.Wrap(err, "error storing featured output")
			}
		}

		// output the data with the new feature
		writer.Flush()
		err = ioutil.WriteFile(config.getTmpAbsolutePath(config.FeaturizationOutputDataRelative), output.Bytes(), 0644)
		if err != nil {
			return errors.Wrap(err, "error writing feature output")
		}
	} else {
		// copy input to make merging happy
		copyFileContents(path.Join(config.ContainerDataPath, mainDR.ResPath), config.getTmpAbsolutePath(config.FeaturizationOutputDataRelative))
	}

	mainDR.ResPath = config.FeaturizationOutputDataRelative

	// write the new schema to file
	err = meta.WriteSchema(config.getTmpAbsolutePath(config.FeaturizationOutputSchemaRelative))
	if err != nil {
		return errors.Wrap(err, "unable to store feature schema")
	}

	return nil
}

func toStringArray(in []interface{}) []string {
	strArr := make([]string, len(in))
	for _, v := range in {
		strArr = append(strArr, v.(string))
	}
	return strArr
}

func toFloat64Array(in []interface{}) ([]float64, error) {
	strArr := make([]float64, len(in))
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

func copyFileContents(source string, destination string) error {
	in, err := os.Open(source)
	if err != nil {
		return errors.Wrap(err, "unable to open source")
	}
	defer in.Close()
	out, err := os.Create(destination)
	if err != nil {
		return errors.Wrap(err, "unable to open destination")
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()

	if _, err = io.Copy(out, in); err != nil {
		return errors.Wrap(err, "unable to copy data")
	}
	err = out.Sync()
	if err != nil {
		return errors.Wrap(err, "unable to finalize copy")
	}

	return nil
}
