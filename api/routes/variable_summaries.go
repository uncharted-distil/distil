package routes

import (
	"net/http"

	"github.com/pkg/errors"
	"goji.io/pat"
	"gopkg.in/olivere/elastic.v3"

	"github.com/unchartedsoftware/distil/api/model"
)

// SummaryResult represents a summary response for a variable.
type SummaryResult struct {
	Histograms []model.Histogram `json:"histograms"`
}

// VariableSummariesHandler generates a route handler that facilitates the
// creation and retrieval of summary information about the variables in a
// dataset.  Currently this consists of a histogram for each variable, but can
// be extended to support avg, std dev, percentiles etc.  in th future.
func VariableSummariesHandler(client *elastic.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// get index name
		index := pat.Param(r, "index")
		// get dataset name
		dataset := pat.Param(r, "dataset")
		// fetch summary histogram
		histograms, err := model.FetchSummaries(client, index, dataset)
		if err != nil {
			handleError(w, err)
			return
		}
		// marshall output into JSON
		err = handleJSON(w, SummaryResult{
			Histograms: histograms,
		})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal summary result into JSON"))
			return
		}
	}
}
