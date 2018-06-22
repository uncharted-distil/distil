package model

import (
	"github.com/pkg/errors"
)

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

// QueriedDataset wraps dataset querying components into a single entity.
type QueriedDataset struct {
	Metadata *Dataset
	Data     *FilteredData
	Filters  *FilterParams
	IsTrain  bool
}

// FetchDataset builds a QueriedDataset from the needed parameters.
func FetchDataset(dataset string, index string, includeIndex bool, includeMeta bool, filterParams *FilterParams, storageMeta MetadataStorage, storageData DataStorage) (*QueriedDataset, error) {
	datasets, err := storageMeta.FetchDatasets(includeIndex, includeMeta)
	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch variables")
	}

	// TODO: Add FetchDataset function to metadata storage.
	var metadata *Dataset
	for _, ds := range datasets {
		if ds.Name == dataset {
			metadata = ds
		}
	}

	data, err := storageData.FetchData(dataset, filterParams, false)
	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch data")
	}

	return &QueriedDataset{
		Metadata: metadata,
		Data:     data,
		Filters:  filterParams,
	}, nil
}
