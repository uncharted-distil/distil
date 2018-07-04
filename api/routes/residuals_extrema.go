package routes

import (
	"math"
	"net/http"

	"github.com/pkg/errors"
	"goji.io/pat"

	"github.com/unchartedsoftware/distil/api/model"
)

// ResidualsExtrema contains a residual extrema response.
type ResidualsExtrema struct {
	Extrema *model.Extrema `json:"extrema"`
}

func fetchSolutionResidualExtrema(meta model.MetadataStorage, data model.DataStorage, solution model.SolutionStorage, dataset string, target string, solutionID string) (*model.Extrema, error) {
	// check target var type
	variable, err := meta.FetchVariable(dataset, target)
	if err != nil {
		return nil, err
	}

	if !model.IsNumerical(variable.Type) {
		return nil, nil
	}

	// we need to get extrema min and max for all solutions sharing dataset and target
	requests, err := solution.FetchRequestByDatasetTarget(dataset, target, solutionID)
	if err != nil {
		return nil, err
	}

	// get extrema
	min := math.MaxFloat64
	max := -math.MaxFloat64
	for _, req := range requests {
		for _, sol := range req.Solutions {
			if sol.Result != nil {
				// result uri
				resultURI := sol.Result.ResultURI
				// predicted extrema
				residualExtrema, err := data.FetchResidualsExtremaByURI(dataset, resultURI)
				if err != nil {
					return nil, err
				}
				max = math.Max(max, residualExtrema.Max)
				min = math.Min(min, residualExtrema.Min)
			}
		}
	}

	// make symmetrical
	extremum := math.Max(math.Abs(min), math.Abs(max))

	return model.NewExtrema(-extremum, extremum)
}

// ResidualsExtremaHandler returns the extremas for a residual summary.
func ResidualsExtremaHandler(metaCtor model.MetadataStorageCtor, solutionCtor model.SolutionStorageCtor, dataCtor model.DataStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		dataset := pat.Param(r, "dataset")
		target := pat.Param(r, "target")

		meta, err := metaCtor()
		if err != nil {
			handleError(w, err)
			return
		}

		solution, err := solutionCtor()
		if err != nil {
			handleError(w, err)
			return
		}

		data, err := dataCtor()
		if err != nil {
			handleError(w, err)
			return
		}

		// extract extrema for solution
		extrema, err := fetchSolutionResidualExtrema(meta, data, solution, dataset, target, "")
		if err != nil {
			handleError(w, err)
			return
		}

		// marshall data and sent the response back
		err = handleJSON(w, ResidualsExtrema{
			Extrema: extrema,
		})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal result histogram into JSON"))
			return
		}
	}
}
