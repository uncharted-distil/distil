package task

import (
	"encoding/json"
	"os"
	"path"
	"strconv"

	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil-ingest/rest"

	"github.com/unchartedsoftware/distil-compute/primitive/compute/description"
	"github.com/unchartedsoftware/distil-compute/primitive/compute/result"
	"github.com/unchartedsoftware/distil/api/util"
)

// Rank will rank the dataset using a primitive.
func Rank(index string, dataset string, config *IngestTaskConfig) error {
	schemaDoc := path.Dir(config.GetTmpAbsolutePath(path.Join(dataset, config.MergedOutputSchemaPathRelative)))

	// create & submit the solution request
	pip, err := description.CreatePCAFeaturesPipeline("harry", "")
	if err != nil {
		return errors.Wrap(err, "unable to create PCA pipeline")
	}

	datasetURI, err := submitPipeline([]string{schemaDoc}, pip)
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
	err = util.WriteFileWithDirs(config.GetTmpAbsolutePath(path.Join(dataset, config.RankingOutputPathRelative)), bytes, os.ModePerm)
	if err != nil {
		return errors.Wrap(err, "unable to store ranking result")
	}

	return nil
}
