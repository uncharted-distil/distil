package task

import (
	"fmt"
	"path"
	"strconv"

	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil-compute/model"
	"github.com/unchartedsoftware/distil-compute/primitive/compute/description"
	"github.com/unchartedsoftware/distil-compute/primitive/compute/result"
	"github.com/unchartedsoftware/distil/api/env"
)

// TargetRankPrimitive will rank the dataset relative to a target variable using
// a primitive.
func TargetRankPrimitive(dataset string, target string, features []*model.Variable) (map[string]float64, error) {
	// create & submit the solution request
	pip, err := description.CreateTargetRankingPipeline("roger", "", target, features)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create ranking pipeline")
	}

	// create a reference to the original data path
	config, err := env.LoadConfig()
	if err != nil {
		return nil, errors.Wrap(err, "unable to load config")
	}
	datasetInputDir := path.Join(config.D3MInputDirRoot, dataset, "TRAIN", "dataset_TRAIN")

	datasetURI, err := submitPrimitive([]string{datasetInputDir}, pip)
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
