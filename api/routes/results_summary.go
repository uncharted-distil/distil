package routes

import (
	"net/http"
	"net/url"

	"github.com/pkg/errors"
	"goji.io/pat"

	"github.com/unchartedsoftware/distil/api/model"
)

// ResultsSummary contains a fetch result histogram.
type ResultsSummary struct {
	ResultsSummary *model.Histogram `json:"histogram"`
}

// ResultsSummaryHandler bins predicted result data for consumption in a downstream summary view.
func ResultsSummaryHandler(ctor model.StorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// extract route parameters
		index := pat.Param(r, "index")
		dataset := pat.Param(r, "dataset")
		resultURI, err := url.PathUnescape(pat.Param(r, "result-uri"))
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to unescape result uri"))
			return
		}

		client, err := ctor()
		if err != nil {
			handleError(w, err)
			return
		}

		histogram, err := model.FetchResultsSummary(client, resultURI, index, dataset)
		if err != nil {
			handleError(w, err)
			return
		}

		// marshall data and sent the response back
		err = handleJSON(w, ResultsSummary{
			ResultsSummary: histogram,
		})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal result histogram into JSON"))
			return
		}
	}
}
