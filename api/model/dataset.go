package model

const (
	// DatasetSuffix is the suffix for the dataset entry when stored in
	// elasticsearch.
	DatasetSuffix = "_dataset"
	metadataType  = "metadata"
)

// Dataset represents a decsription of a dataset.
type Dataset struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Summary     string      `json:"summary"`
	SummaryML   string      `json:"summaryML"`
	Variables   []*Variable `json:"variables"`
	NumRows     int64       `json:"numRows"`
	NumBytes    int64       `json:"numBytes"`
}
