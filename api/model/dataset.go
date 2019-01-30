package model

import (
	"github.com/pkg/errors"

	"github.com/unchartedsoftware/distil-compute/model"
	"github.com/unchartedsoftware/distil-ingest/metadata"
)

const (
	metadataType = "metadata"
)

// Dataset represents a decsription of a dataset.
type Dataset struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	StorageName string                 `json:"storageName"`
	Folder      string                 `json:"folder"`
	Description string                 `json:"description"`
	Summary     string                 `json:"summary"`
	SummaryML   string                 `json:"summaryML"`
	Variables   []*model.Variable      `json:"variables"`
	NumRows     int64                  `json:"numRows"`
	NumBytes    int64                  `json:"numBytes"`
	Provenance  string                 `json:"provenance"`
	Source      metadata.DatasetSource `json:"source"`
}

// QueriedDataset wraps dataset querying components into a single entity.
type QueriedDataset struct {
	Metadata *Dataset
	Data     *FilteredData
	Filters  *FilterParams
	IsTrain  bool
}

// FetchDataset builds a QueriedDataset from the needed parameters.
func FetchDataset(dataset string, includeIndex bool, includeMeta bool, filterParams *FilterParams, storageMeta MetadataStorage, storageData DataStorage) (*QueriedDataset, error) {
	datasets, err := storageMeta.FetchDatasets(includeIndex, includeMeta)
	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch variables")
	}

	// TODO: Add FetchDataset function to metadata storage.
	var metadata *Dataset
	for _, ds := range datasets {
		if ds.ID == dataset {
			metadata = ds
		}
	}
	if metadata == nil {
		return nil, errors.Wrap(err, "unable to fetch metadata")
	}

	data, err := storageData.FetchData(dataset, metadata.StorageName, filterParams, false)
	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch data")
	}

	return &QueriedDataset{
		Metadata: metadata,
		Data:     data,
		Filters:  filterParams,
	}, nil
}

// GetD3MIndexVariable returns the D3M index variable.
func (d *Dataset) GetD3MIndexVariable() *model.Variable {
	for _, v := range d.Variables {
		if v.Name == model.D3MIndexName {
			return v
		}
	}

	return nil
}
