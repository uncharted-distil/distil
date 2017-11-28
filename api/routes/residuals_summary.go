package routes

import (
	"net/http"
	"net/url"

	"github.com/pkg/errors"
	"goji.io/pat"

	"github.com/unchartedsoftware/distil/api/model"
)

// ResidualsSummary contains a fetch result histogram.
type ResidualsSummary struct {
	ResidualsSummary *model.Histogram `json:"histogram"`
}

// ResidualsSummaryHandler bins predicted result data for consumption in a downstream summary view.
func ResidualsSummaryHandler(ctor model.PipelineStorageCtor, ctorData model.DataStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// extract route parameters
		index := pat.Param(r, "index")
		resultsDataset := pat.Param(r, "results-dataset")
		dataset := pat.Param(r, "dataset")
		resultUUID, err := url.PathUnescape(pat.Param(r, "results-uuid"))
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to unescape results uuid"))
			return
		}

		client, err := ctor()
		if err != nil {
			handleError(w, err)
			return
		}

		clientData, err := ctorData()
		if err != nil {
			handleError(w, err)
			return
		}

		// get the result URI. Error ignored to make it ES compatible.
		res, err := client.FetchResultMetadataByUUID(resultUUID)
		resultURI := resultUUID
		if res != nil {
			resultURI = res.ResultURI
		}

		histogram, err := clientData.FetchResidualsSummary(resultsDataset, dataset, resultURI, index)
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
