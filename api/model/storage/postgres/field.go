package postgres

import (
	"github.com/unchartedsoftware/distil/api/model"
)

// Field defines behaviour for a database field type.
type Field interface {
	FetchSummaryData(dataset string, variable *model.Variable, resultURI string, filterParams *model.FilterParams, extrema *model.Extrema) (*model.Histogram, error)
	FetchPredictedSummaryData(resultURI string, dataset string, datasetResult string, variable *model.Variable, filterParams *model.FilterParams, extrema *model.Extrema) (*model.Histogram, error)
}
