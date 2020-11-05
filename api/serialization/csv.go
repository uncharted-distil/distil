//
//   Copyright © 2019 Uncharted Software Inc.
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

package serialization

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/metadata"
	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil-compute/primitive/compute"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/util"
	log "github.com/unchartedsoftware/plog"
)

// CSV represents a dataset storage backed with csv data and json schema doc.
type CSV struct {
}

// NewCSV creates a new csv backed storage.
func NewCSV() *CSV {
	return &CSV{}
}

// ReadDataset reads a raw dataset from the file system, loading the csv
// data into memory.
func (d *CSV) ReadDataset(uri string) (*api.RawDataset, error) {
	data, err := d.ReadData(uri)
	if err != nil {
		return nil, err
	}

	meta, err := d.ReadMetadata(uri)
	if err != nil {
		return nil, err
	}

	return &api.RawDataset{
		Data:     data,
		Metadata: meta,
	}, nil
}

// WriteDataset writes the raw dataset to the file system, writing out
// the data to a csv file.
func (d *CSV) WriteDataset(uri string, data *api.RawDataset) error {

	dataFilename := path.Join(uri, compute.D3MDataFolder, compute.D3MLearningData)
	err := d.WriteData(dataFilename, data.Data)
	if err != nil {
		return err
	}

	data.Metadata.GetMainDataResource().ResPath = dataFilename
	metaFilename := path.Join(uri, compute.D3MDataSchema)
	err = d.WriteMetadata(metaFilename, data.Metadata, true, true)
	if err != nil {
		return err
	}

	return nil
}

// ReadData reads the data from a csv file.
func (d *CSV) ReadData(uri string) ([][]string, error) {
	return util.ReadCSVFile(uri, false)
}

// WriteData writes data to a csv file.
func (d *CSV) WriteData(uri string, data [][]string) error {
	log.Infof("writing data to '%s'", uri)
	var outputBuffer bytes.Buffer
	csvWriter := csv.NewWriter(&outputBuffer)
	err := csvWriter.WriteAll(data)
	if err != nil {
		return errors.Wrap(err, "unable to write csv data to buffer")
	}
	csvWriter.Flush()

	err = util.WriteFileWithDirs(uri, outputBuffer.Bytes(), os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

// ReadRawVariables reads the csv header file to get a list of variables in the file.
func (d *CSV) ReadRawVariables(uri string) ([]string, error) {
	return util.ReadCSVHeader(uri)
}

// ReadMetadata reads the dataset doc from disk.
func (d *CSV) ReadMetadata(uri string) (*model.Metadata, error) {
	meta, err := metadata.LoadMetadataFromOriginalSchema(uri, true)
	if err != nil {
		return nil, err
	}

	return meta, nil
}

// WriteMetadata writes the dataset doc to disk.
func (d *CSV) WriteMetadata(uri string, meta *model.Metadata, extended bool, update bool) error {
	dataResources := make([]interface{}, 0)

	// make sure the resource format and path match expected csv types
	mainDR := meta.GetMainDataResource()
	if mainDR.ResFormat[compute.D3MResourceFormat] == nil {
		if !update {
			return errors.Errorf("main data resource not set to csv format")
		}
		mainDR.ResFormat = map[string][]string{compute.D3MResourceFormat: {"csv"}}
		mainDR.ResPath = fmt.Sprintf("%s.csv", strings.TrimSuffix(mainDR.ResPath, path.Ext(mainDR.ResPath)))
	}
	for _, dr := range meta.DataResources {
		dataResources = append(dataResources, d.writeDataResource(dr, extended))
	}

	about := map[string]interface{}{
		"datasetID":            meta.ID,
		"datasetName":          meta.Name,
		"description":          meta.Description,
		"datasetSchemaVersion": schemaVersion,
		"license":              license,
		"redacted":             meta.Redacted,
		"digest":               meta.Digest,
	}

	if extended {
		about["parentDatasetIDs"] = meta.ParentDatasetIDs
		about["storageName"] = meta.StorageName
		about["rawData"] = meta.Raw
		about["mergedSchema"] = "false"
	}

	output := map[string]interface{}{
		"about":         about,
		"dataResources": dataResources,
	}

	bytes, err := json.MarshalIndent(output, "", "	")
	if err != nil {
		return errors.Wrap(err, "failed to marshal merged schema file output")
	}
	// write copy to disk
	err = ioutil.WriteFile(uri, bytes, 0644)
	if err != nil {
		return errors.Wrap(err, "failed to write metadata to disk")
	}

	return nil
}

func (d *CSV) writeDataResource(resource *model.DataResource, extended bool) map[string]interface{} {
	vars := make([]interface{}, 0)

	for _, v := range resource.Variables {
		vars = append(vars, d.writeVariable(v, extended))
	}

	output := map[string]interface{}{
		"resID":        resource.ResID,
		"resPath":      resource.ResPath,
		"resType":      resource.ResType,
		"resFormat":    resource.ResFormat,
		"isCollection": resource.IsCollection,
	}

	if len(vars) > 0 {
		output["columns"] = vars
	}

	return output
}

func (d *CSV) writeVariable(variable *model.Variable, extended bool) interface{} {
	// col type index doesn't exist for TA2
	colType := model.MapSchemaType(variable.Type)

	output := map[string]interface{}{
		model.VarIndexField: variable.Index,
		model.VarNameField:  variable.DisplayName,
		model.VarTypeField:  colType,
		model.VarRoleField:  variable.Role,
		"refersTo":          variable.RefersTo,
	}

	if extended {
		output[model.VarDescriptionField] = variable.Description
		output[model.VarOriginalTypeField] = variable.OriginalType
		output[model.VarSelectedRoleField] = variable.SelectedRole
		output[model.VarDistilRole] = variable.DistilRole
		output[model.VarOriginalVariableField] = variable.OriginalVariable
		output[model.VarNameField] = variable.Name
		output[model.VarDisplayVariableField] = variable.DisplayName
		output[model.VarImportanceField] = variable.Importance
		output[model.VarSuggestedTypesField] = variable.SuggestedTypes
		output[model.VarDeleted] = variable.Deleted
		output[model.VarGroupingField] = variable.Grouping
		output[model.VarMinField] = variable.Min
		output[model.VarMaxField] = variable.Max
	}

	return output
}
