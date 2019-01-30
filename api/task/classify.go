package task

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strconv"

	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil-ingest/rest"
	log "github.com/unchartedsoftware/plog"

	"github.com/unchartedsoftware/distil-compute/model"
	"github.com/unchartedsoftware/distil-compute/primitive/compute/description"
	"github.com/unchartedsoftware/distil-compute/primitive/compute/result"
	"github.com/unchartedsoftware/distil/api/util"
)

const (
	defaultEmptyType = model.TextType
	defaultEmptyProb = "1.0"
)

func castTypeArray(in []interface{}) ([]string, error) {
	strArr := make([]string, 0)
	for _, v := range in {
		str, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("arg is not a string, %v", v)
		}
		if str == "" || str == "[]" || str == "()" {
			str = defaultEmptyType
		}
		strArr = append(strArr, str)
	}
	return strArr, nil
}

func castProbabilityArray(in []interface{}) ([]float64, error) {
	fltArr := make([]float64, 0)
	for _, v := range in {
		str, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("arg is not a string, %v", v)
		}
		if str == "" || str == "[]" || str == "()" {
			str = defaultEmptyProb
		}

		flt, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert interface array to float array")
		}
		fltArr = append(fltArr, flt)
	}
	return fltArr, nil
}

// Classify will classify the dataset using a primitive.
func Classify(schemaPath string, index string, dataset string, config *IngestTaskConfig) error {
	schemaDoc := path.Dir(schemaPath)

	// create & submit the solution request
	pip, err := description.CreateSimonPipeline("says", "")
	if err != nil {
		return errors.Wrap(err, "unable to create Simon pipeline")
	}

	datasetURI, err := submitPipeline([]string{schemaDoc}, pip)
	if err != nil {
		return errors.Wrap(err, "unable to run Simon pipeline")
	}

	// parse primitive response
	res, err := result.ParseResultCSV(datasetURI)
	if err != nil {
		return errors.Wrap(err, "unable to parse Simon pipeline result")
	}

	// First row is header, then all other rows are col index, types, probabilities.
	probabilities := make([][]float64, len(res)-1)
	labels := make([][]string, len(res)-1)
	for i, v := range res {
		if i > 0 {
			colIndexString, ok := v[0].(string)
			if !ok {
				return fmt.Errorf("first column returned is not of expected type `string`, %v", v[0])
			}

			typesArray, ok := v[1].([]interface{})
			if !ok {
				vs, ok := v[1].(interface{})
				if !ok {
					return fmt.Errorf("second column returned is not of type `[]interface{}` %v", v[1])
				}
				typesArray = []interface{}{vs}
			}

			probabilitiesArray, ok := v[2].([]interface{})
			if !ok {
				vs, ok := v[2].(interface{})
				if !ok {
					return fmt.Errorf("third column returned is not of type `[]interface{}` %v", v[2])
				}
				probabilitiesArray = []interface{}{vs}
			}

			colIndex, err := strconv.ParseInt(colIndexString, 10, 64)
			if err != nil {
				return err
			}

			fieldLabels, err := castTypeArray(typesArray)
			if err != nil {
				return err
			}
			probs, err := castProbabilityArray(probabilitiesArray)
			if err != nil {
				return err
			}
			labels[colIndex] = mapClassifiedTypes(fieldLabels)
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
	outputPath := path.Join(schemaDoc, config.ClassificationOutputPathRelative)
	log.Debugf("writing classification output to %s", outputPath)
	err = util.WriteFileWithDirs(outputPath, bytes, os.ModePerm)
	if err != nil {
		return errors.Wrap(err, "unable to store classification result")
	}

	return nil
}

func mapClassifiedTypes(types []string) []string {
	for i, typ := range types {
		types[i] = model.MapSimonType(typ)
	}

	return types
}
