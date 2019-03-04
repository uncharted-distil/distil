package postgres

import (
	"github.com/uncharted-distil/distil/api/model"
)

// Field defines behaviour for a database field type.
type Field interface {
	FetchSummaryData(resultURI string, filterParams *model.FilterParams, extrema *model.Extrema) (*model.Histogram, error)
	FetchPredictedSummaryData(resultURI string, datasetResult string, filterParams *model.FilterParams, extrema *model.Extrema) (*model.Histogram, error)
}
