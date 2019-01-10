package task

import (
	"encoding/json"
	"os"
	"path"
	"strings"

	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil-ingest/rest"

	"github.com/unchartedsoftware/distil-compute/primitive/compute/description"
	"github.com/unchartedsoftware/distil-compute/primitive/compute/result"
	"github.com/unchartedsoftware/distil/api/util"
)

// Summarize will summarize the dataset using a primitive.
func Summarize(index string, dataset string, config *IngestTaskConfig) error {
	schemaDoc := path.Dir(config.GetTmpAbsolutePath(config.MergedOutputSchemaPathRelative))

	// create & submit the solution request
	pip, err := description.CreateDukePipeline("wellington", "")
	if err != nil {
		return errors.Wrap(err, "unable to create Duke pipeline")
	}

	datasetURI, err := submitPipeline([]string{schemaDoc}, pip)
	if err != nil {
		return errors.Wrap(err, "unable to run Duke pipeline")
	}

	// parse primitive response (row,token,probability)
	res, err := result.ParseResultCSV(datasetURI)
	if err != nil {
		return errors.Wrap(err, "unable to parse Duke pipeline result")
	}

	tokens := make([]string, len(res)-1)
	for i, v := range res {
		// skip the header
		if i > 0 {
			token, ok := v[1].(string)
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
	err = util.WriteFileWithDirs(config.GetTmpAbsolutePath(config.SummaryMachineOutputPathRelative), bytes, os.ModePerm)
	if err != nil {
		return errors.Wrap(err, "unable to store summary result")
	}

	return nil
}
