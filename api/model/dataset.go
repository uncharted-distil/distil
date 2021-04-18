//
//   Copyright © 2021 Uncharted Software Inc.
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package model

import (
	"fmt"
	"path"

	"github.com/pkg/errors"
	log "github.com/unchartedsoftware/plog"

	"github.com/uncharted-distil/distil-compute/metadata"
	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil-compute/primitive/compute"
	"github.com/uncharted-distil/distil/api/env"
	"github.com/uncharted-distil/distil/api/serialization"
	"github.com/uncharted-distil/distil/api/util/json"
)

// DatasetType is used to identify the type of dataset ingested.
type DatasetType string

const (
	// DatasetTypeModelling is a dataset used to build models.
	DatasetTypeModelling DatasetType = "modelling"
	// DatasetTypeInference is a dataset consumed by a model to infer predictions.
	DatasetTypeInference DatasetType = "inference"
)

// Dataset represents a decsription of a dataset.
type Dataset struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	StorageName     string                 `json:"storageName"`
	Folder          string                 `json:"datasetFolder"`
	Description     string                 `json:"description"`
	Summary         string                 `json:"summary"`
	SummaryML       string                 `json:"summaryML"`
	Variables       []*model.Variable      `json:"variables"`
	NumRows         int64                  `json:"numRows"`
	NumBytes        int64                  `json:"numBytes"`
	Provenance      string                 `json:"provenance"`
	Source          metadata.DatasetSource `json:"source"`
	JoinSuggestions []*JoinSuggestion      `json:"joinSuggestion"`
	JoinScore       float64                `json:"joinScore"`
	Type            DatasetType            `json:"type"`
	LearningDataset string                 `json:"learningDataset"`
	Clone           bool                   `json:"clone"`
	Immutable       bool                   `json:"immutable"`
	ParentDataset   string                 `json:"parentDataset"`
}

// JoinSuggestion specifies potential joins between datasets.
type JoinSuggestion struct {
	BaseDataset   string               `json:"baseDataset"`
	BaseColumns   []string             `json:"baseColumns"`
	JoinDataset   string               `json:"joinDataset"`
	JoinColumns   []string             `json:"joinColumns"`
	JoinScore     float64              `json:"joinScore"`
	DatasetOrigin *model.DatasetOrigin `json:"datasetOrigin"`
	Index         int                  `json:"index"`
}

// VariableUpdate captures the information to update the dataset data.
type VariableUpdate struct {
	Index string `json:"index"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

// DiskDataset represents a dataset stored on disk.
type DiskDataset struct {
	Dataset           *serialization.RawDataset
	FeaturizedDataset *DiskDataset
	schemaPath        string
}

// MapVariables creates a variable map using the mapper function to create the key.
func MapVariables(variables []*model.Variable, mapper func(variable *model.Variable) string) map[string]*model.Variable {
	mapped := map[string]*model.Variable{}
	for _, d := range variables {
		mapped[mapper(d)] = d
	}

	return mapped
}

// LoadDiskDataset loads a dataset from disk. It will load the learning dataset
// as the featurized dataset.
func LoadDiskDataset(ds *Dataset) (*DiskDataset, error) {
	folder := env.ResolvePath(ds.Source, ds.Folder)
	output, err := LoadDiskDatasetFromFolder(folder)
	if err != nil {
		return nil, err
	}

	if ds.LearningDataset != "" {
		pre, err := LoadDiskDatasetFromFolder(ds.LearningDataset)
		if err != nil {
			return nil, err
		}
		output.FeaturizedDataset = pre
	}

	return output, nil
}

// LoadDiskDatasetFromFolder loads a dataset from disk at the specified location.
func LoadDiskDatasetFromFolder(folder string) (*DiskDataset, error) {
	schemaPath := path.Join(folder, compute.D3MDataSchema)
	dsDisk, err := serialization.ReadDataset(schemaPath)
	if err != nil {
		return nil, err
	}

	return &DiskDataset{
		Dataset:    dsDisk,
		schemaPath: schemaPath,
	}, nil
}

// UpdateDiskDataset updates a disk dataset to have the new and updated data.
func UpdateDiskDataset(ds *Dataset, data [][]string) error {
	// read the dataset from disk
	dsDisk, err := LoadDiskDataset(ds)
	if err != nil {
		return err
	}
	return dsDisk.UpdateOnDisk(ds, data, false)
}

// UpdateOnDisk updates a disk dataset to have the new and updated data.
func (d *DiskDataset) UpdateOnDisk(ds *Dataset, data [][]string, updateImmutable bool) error {
	// use the header row to determine the variables to update
	varMap := MapVariables(ds.Variables, func(variable *model.Variable) string { return variable.HeaderName })
	d3mIndexIndex := -1
	updates := map[string]map[string]string{}
	headerMap := map[string]int{}
	for i, c := range data[0] {
		if c == model.D3MIndexFieldName {
			d3mIndexIndex = i
		} else {
			sourceVar := varMap[c]

			// if source var doesnt exist, then no update possible
			// (prefeaturized datasets have all the featurized values that dont exist outside of the disk version)
			if sourceVar == nil || (!updateImmutable && sourceVar.Immutable) {
				continue
			}
			headerMap[sourceVar.HeaderName] = i
			updates[sourceVar.HeaderName] = map[string]string{}

			// add missing fields
			if !d.FieldExists(sourceVar) {
				err := d.AddField(sourceVar)
				if err != nil {
					return err
				}
			}
		}
	}

	// create the update maps (d3m index -> new value)
	for _, row := range data[1:] {
		for headerName, colIndex := range headerMap {
			updates[headerName][row[d3mIndexIndex]] = row[colIndex]
		}
	}

	err := d.Update(updates, true)
	if err != nil {
		return err
	}

	err = d.SaveDataset()
	if err != nil {
		return err
	}

	return nil
}

// SaveDataset saves a dataset to disk.
func (d *DiskDataset) SaveDataset() error {
	err := serialization.WriteDataset(path.Dir(d.schemaPath), d.Dataset)
	if err != nil {
		return err
	}

	if d.FeaturizedDataset != nil {
		err = d.FeaturizedDataset.SaveDataset()
		if err != nil {
			return err
		}
	}

	return nil
}

// AddField adds a field to the dataset.
func (d *DiskDataset) AddField(variable *model.Variable) error {
	err := d.Dataset.AddField(variable)
	if err != nil {
		return err
	}

	if d.FeaturizedDataset != nil {
		err = d.FeaturizedDataset.AddField(variable)
		if err != nil {
			return err
		}
	}

	return nil
}

// FieldExists returns true if a field exists in the dataset.
func (d *DiskDataset) FieldExists(variable *model.Variable) bool {
	return d.Dataset.FieldExists(variable)
}

// Clone clones a dataset on disk.
func (d *DiskDataset) Clone(targetFolder string, cloneDatasetID string, cloneStorageName string) (*DiskDataset, error) {
	// easiest clone is to write the dataset to the new location then read it
	err := serialization.WriteDataset(targetFolder, d.Dataset)
	if err != nil {
		return nil, err
	}

	schemaPath := path.Join(targetFolder, compute.D3MDataSchema)
	cloned, err := serialization.ReadDataset(schemaPath)
	if err != nil {
		return nil, err
	}

	// update the metadata
	cloned.Metadata.ID = cloneDatasetID
	cloned.Metadata.StorageName = cloneStorageName
	err = serialization.WriteMetadata(schemaPath, cloned.Metadata)
	if err != nil {
		return nil, err
	}

	output := &DiskDataset{
		Dataset:    cloned,
		schemaPath: schemaPath,
	}

	if d.FeaturizedDataset != nil {
		clonedPrefeaturized, err := d.FeaturizedDataset.Clone(fmt.Sprintf("%s-featurized", targetFolder), cloneDatasetID, cloneStorageName)
		if err != nil {
			return nil, err
		}
		output.FeaturizedDataset = clonedPrefeaturized
	}

	return output, nil
}

// GetLearningFolder returns the folder containing the learning dataset.
func (d *DiskDataset) GetLearningFolder() string {
	if d.FeaturizedDataset != nil {
		return d.FeaturizedDataset.GetLearningFolder()
	}

	return path.Dir(d.schemaPath)
}

// Update updates a dataset, optionally filtering rows not being updated.
func (d *DiskDataset) Update(updates map[string]map[string]string, filterNotFound bool) error {
	if filterNotFound {
		// do an initial filter pass to keep only the rows found in the updates
		filterMap := map[string]bool{}
		for _, colUpdates := range updates {
			for key := range colUpdates {
				filterMap[key] = true
			}
		}
		d.Dataset.FilterDataset(filterMap)
	}

	// translate column names to indices
	updatesMapped := map[int]map[string]string{}
	for colName, colUpdates := range updates {
		index := d.Dataset.GetVariableIndex(colName)

		// could add missing columns!
		if index == -1 {
			return errors.Errorf("column %s not in dataset for updates", colName)
		}
		updatesMapped[index] = colUpdates
	}
	d.Dataset.UpdateDataset(updatesMapped)

	if d.FeaturizedDataset != nil {
		err := d.FeaturizedDataset.Update(updates, filterNotFound)
		if err != nil {
			return err
		}
	}

	return nil
}

// ParseVariableUpdateList returns a list of parsed variable updates.
func ParseVariableUpdateList(data map[string]interface{}) ([]*VariableUpdate, error) {
	updatesRaw, ok := json.Array(data, "updates")
	if !ok {
		log.Infof("no variable updates to parse")
		return nil, nil
	}

	updatesParsed := make([]*VariableUpdate, len(updatesRaw))
	for i, update := range updatesRaw {
		index, ok := json.String(update, "index")
		if !ok {
			return nil, errors.Errorf("no index provided for variable update")
		}
		name, ok := json.String(update, "name")
		if !ok {
			return nil, errors.Errorf("no feature name provided for variable update")
		}
		value, ok := json.String(update, "value")
		if !ok {
			return nil, errors.Errorf("no feature value provided for variable update")
		}
		updatesParsed[i] = &VariableUpdate{
			Index: index,
			Name:  name,
			Value: value,
		}
	}

	return updatesParsed, nil
}

// GetD3MIndexVariable returns the D3M index variable.
func (d *Dataset) GetD3MIndexVariable() *model.Variable {
	for _, v := range d.Variables {
		if v.Key == model.D3MIndexFieldName {
			return v
		}
	}

	return nil
}

// ToMetadata capture the dataset metadata in a d3m metadata struct.
func (d *Dataset) ToMetadata() *model.Metadata {
	// create the data resource
	dr := model.NewDataResource(compute.DefaultResourceID, model.ResTypeTable, map[string][]string{compute.D3MResourceFormat: {"csv"}})
	dr.Variables = d.Variables

	// create the necessary data structures for the mapping
	origins := make([]*model.DatasetOrigin, len(d.JoinSuggestions))
	for i, js := range d.JoinSuggestions {
		origins[i] = js.DatasetOrigin
	}

	// create the metadata
	meta := model.NewMetadata(d.ID, d.Name, d.Description, d.StorageName)
	meta.DatasetFolder = d.Folder
	meta.DatasetOrigins = origins
	meta.LearningDataset = d.LearningDataset
	meta.NumBytes = d.NumBytes
	meta.NumRows = d.NumRows
	meta.SchemaSource = string(d.Source)
	meta.SearchProvenance = d.Provenance
	meta.Summary = d.Summary
	meta.SummaryMachine = d.SummaryML
	meta.Redacted = false
	meta.Raw = false
	meta.Clone = d.Clone
	meta.Immutable = d.Immutable
	meta.DataResources = []*model.DataResource{dr}

	return meta
}

// GetLearningFolder returns the folder on disk that has the data for learning.
func (d *Dataset) GetLearningFolder() string {
	if d.LearningDataset != "" {
		return d.LearningDataset
	}
	return env.ResolvePath(d.Source, d.Folder)
}

// UpdateExtremas updates the variable extremas based on the data stored.
func UpdateExtremas(dataset string, varName string, storageMeta MetadataStorage, storageData DataStorage) error {
	// get the metadata and then query the data storage for the latest values
	d, err := storageMeta.FetchDataset(dataset, false, false, false)
	if err != nil {
		return err
	}

	// find the variable
	var v *model.Variable
	for _, variable := range d.Variables {
		if variable.Key == varName {
			v = variable
			break
		}
	}

	// only care about datetime, categorical and numerical
	// may want to consider building a map containing the types we care about
	if model.IsDateTime(v.Type) || model.IsNumerical(v.Type) || model.IsCategorical(v.Type) {
		// get the extrema
		extrema, err := storageData.FetchExtrema(d.ID, d.StorageName, v)
		if err != nil {
			return err
		}

		// store the extrema to ES
		err = storageMeta.SetExtrema(dataset, varName, extrema)
		if err != nil {
			return err
		}
	}

	return nil
}

// ParseDatasetOriginsFromJSON parses dataset origins from string maps.
func ParseDatasetOriginsFromJSON(originsJSON []map[string]interface{}) []*model.DatasetOrigin {

	origins := make([]*model.DatasetOrigin, len(originsJSON))

	for i, originJSON := range originsJSON {
		origins[i] = parseDatasetOriginFromJSON(originJSON)
	}

	return origins
}

func parseDatasetOriginFromJSON(originJSON map[string]interface{}) *model.DatasetOrigin {
	searchResult, ok := json.String(originJSON, "searchResult")
	if !ok {
		searchResult = ""
	}
	provenance, ok := json.String(originJSON, "provenance")
	if !ok {
		provenance = ""
	}
	sourceDataset, ok := json.String(originJSON, "sourceDataset")
	if !ok {
		sourceDataset = ""
	}

	return &model.DatasetOrigin{
		SearchResult:  searchResult,
		Provenance:    provenance,
		SourceDataset: sourceDataset,
	}
}
