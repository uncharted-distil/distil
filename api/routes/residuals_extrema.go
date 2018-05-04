package routes

import (
	"math"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
	"goji.io/pat"

	"github.com/unchartedsoftware/distil/api/model"
)

// ResidualsExtremaHandler returns the extremas for a residual summary.
func ResidualsExtremaHandler(solutionCtor model.SolutionStorageCtor, dataCtor model.DataStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// extract route parameters
		index := pat.Param(r, "index")
		dataset := pat.Param(r, "dataset")
		resultUUID, err := url.PathUnescape(pat.Param(r, "results-uuid"))
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to unescape results uuid"))
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

		// get the result URI. Error ignored to make it ES compatible.
		res, err := solution.FetchSolutionResultByUUID(resultUUID)
		if err != nil {
			handleError(w, err)
			return
		}

		extrema, err := data.FetchResidualsExtremaByURI(dataset, res.ResultURI, index)
		if err != nil {
			handleError(w, err)
			return
		}

		extremum := math.Max(math.Abs(extrema.Min), math.Abs(extrema.Max))

		// marshall data and sent the response back
		err = handleJSON(w, map[string]interface{}{
			"extrema": model.Extrema{
				Min: -extremum,
				Max: extremum,
			},
		})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal result histogram into JSON"))
			return
		}
	}
}
