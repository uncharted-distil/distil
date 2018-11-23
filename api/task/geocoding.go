package task

import (
	"path"
	"strconv"

	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil-compute/model"
	"github.com/unchartedsoftware/distil-compute/primitive/compute/description"
	"github.com/unchartedsoftware/distil-compute/primitive/compute/result"

	"github.com/unchartedsoftware/distil/api/env"
)

// GeocodedPoint contains data that has been geocoded.
type GeocodedPoint struct {
	D3MIndex  string
	Latitude  float64
	Longitude float64
}

// GeocodeForward will geocode a column into lat & lon values.
func GeocodeForward(dataset string, variable string, features []*model.Variable) ([]*GeocodedPoint, error) {
	// create a reference to the original data path
	config, err := env.LoadConfig()
	if err != nil {
		return nil, errors.Wrap(err, "unable to load config")
	}
	datasetInputDir := path.Join(config.D3MInputDirRoot, dataset, "TRAIN", "dataset_TRAIN")

	// create & submit the solution request
	pip, err := description.CreateGoatForwardPipeline("mountain", "", variable, features)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create Goat pipeline")
	}

	datasetURI, err := submitPrimitive(datasetInputDir, pip)
	if err != nil {
		return nil, errors.Wrap(err, "unable to run Goat pipeline")
	}

	// parse primitive response (col index,importance)
	res, err := result.ParseResultCSV(datasetURI)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse Goat pipeline result")
	}

	// result should be d3m index, lat, lon
	geocodedData := make([]*GeocodedPoint, len(res)-1)
	for i, v := range res {
		if i > 0 {
			d3mIndex, ok := v[0].(string)
			if !ok {
				return nil, errors.Errorf("unable to parse d3m index from result")
			}
			lat, err := strconv.ParseFloat(v[1].(string), 64)
			if err != nil {
				return nil, errors.Wrap(err, "unable to parse latitude from result")
			}
			lon, err := strconv.ParseFloat(v[2].(string), 64)
			if err != nil {
				return nil, errors.Wrap(err, "unable to parse longitude from result")
			}

			geocodedData = append(geocodedData, &GeocodedPoint{
				D3MIndex:  d3mIndex,
				Latitude:  lat,
				Longitude: lon,
			})
		}
	}

	return geocodedData, nil
}
