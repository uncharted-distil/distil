package routes

import (
	"net/http"

	"github.com/pkg/errors"
	"goji.io/pat"

	"github.com/unchartedsoftware/distil/api/elastic"
	"github.com/unchartedsoftware/distil/api/model"
)

// SummaryResult represents a summary response for a variable.
type SummaryResult struct {
	Histograms []*model.Histogram `json:"histograms"`
}

// VariableSummaryHandler generates a route handler that facilitates the
// creation and retrieval of summary information about the specified variable.
func VariableSummaryHandler(ctor elastic.ClientCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// get index name
		index := pat.Param(r, "index")
		// get dataset name
		dataset := pat.Param(r, "dataset")
		// get variabloe name
		variable := pat.Param(r, "variable")
		// get elasticsearch client
		client, err := ctor()
		if err != nil {
			handleError(w, err)
			return
		}
		// fetch summary histogram
		histogram, err := model.FetchSummary(client, index, dataset, variable)
		if err != nil {
			handleError(w, err)
			return
		}
		// marshall output into JSON
		err = handleJSON(w, SummaryResult{
			Histograms: []*model.Histogram{histogram},
		})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal summary result into JSON"))
			return
		}
	}
}
