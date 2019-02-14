package task

import (
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil-compute/model"
	"github.com/unchartedsoftware/distil-compute/primitive/compute/description"
	"github.com/unchartedsoftware/distil-compute/primitive/compute/result"
	"github.com/unchartedsoftware/distil-ingest/metadata"
	"github.com/unchartedsoftware/distil/api/env"
)

// TargetRank will rank the dataset relative to a target variable using
// a primitive.
func TargetRank(dataset string, target string, features []*model.Variable, source metadata.DatasetSource) (map[string]float64, error) {
	// create & submit the solution request
	pip, err := description.CreateTargetRankingPipeline("roger", "", target, features)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create ranking pipeline")
	}

	datasetInputDir, err := env.ResolvePath(source, dataset)
	if err != nil {
		return nil, errors.Wrap(err, "unable to resolve path")
	}

	datasetURI, err := submitPipeline([]string{datasetInputDir}, pip)
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
