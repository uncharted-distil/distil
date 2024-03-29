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
	"github.com/uncharted-distil/distil-compute/primitive/compute/result"
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
func (d *CSV) ReadDataset(schemaFile string) (*RawDataset, error) {
	meta, err := d.ReadMetadata(schemaFile)
	if err != nil {
		return nil, err
	}

	data, err := d.ReadData(model.GetResourcePath(schemaFile, meta.GetMainDataResource()))
	if err != nil {
		return nil, err
	}

	return &RawDataset{
		ID:       meta.ID,
		Name:     meta.Name,
		Data:     data,
		Metadata: meta,
	}, nil
}

// WriteDataset writes the raw dataset to the file system, writing out
// the data to a csv file.
func (d *CSV) WriteDataset(uri string, data *RawDataset) error {

	dataFilename := path.Join(uri, compute.D3MDataFolder, compute.D3MLearningData)
	err := d.WriteData(dataFilename, data.Data)
	if err != nil {
		return err
	}

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
	log.Infof("writing csv data to '%s'", uri)
	var outputBuffer bytes.Buffer
	csvWriter := csv.NewWriter(&outputBuffer)
	err := csvWriter.WriteAll(data)
	if err != nil {
		return errors.Wrap(err, "unable to write csv data to buffer")
	}
	csvWriter.Flush()

	err = util.WriteFileWithDirs(uri, outputBuffer.Bytes(), os.ModePerm)
	if err != nil {
		return errors.Wrap(err, "unable to write csv data to disk")
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
		mapped := d.writeDataResource(dr, extended)
		if dr == mainDR {
			mapped["resPath"] = path.Join(path.Dir(uri), compute.D3MDataFolder, compute.D3MLearningData)
		}
		dataResources = append(dataResources, mapped)
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
		model.VarNameField:  variable.HeaderName,
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
		output[model.VarKeyField] = variable.Key
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

// ResultToInputCSV takes a result produced by a TA2 pipeline run ensures that it is in a format
// suitable for storage as a D3M dataset.
func ResultToInputCSV(resultURI string) ([][]string, error) {
	// Parse the result CSV
	result, err := result.ParseResultCSV(resultURI)
	if err != nil {
		return nil, err
	}

	transformedInput := make([][]string, len(result))

	// Loop over the result structure and save each field out
	for rowIdx, row := range result {
		record := make([]string, len(row))
		for colIdx, v := range row {
			if v != nil {
				if arr, ok := v.([]interface{}); ok {
					// If this field was parsed as an array, we will override its string conversion
					// to ensure that we write out in D3M format, which is a quoted comma separated
					// list.  If the array is nested, the nested data will be blindly converted
					// to a string.
					strArr := make([]string, len(arr))
					for i, s := range arr {
						strArr[i] = fmt.Sprintf("%v", s)
					}
					record[colIdx] = strings.Join(strArr, ",")
				} else {
					record[colIdx] = fmt.Sprintf("%v", v)
				}
			}
		}
		transformedInput[rowIdx] = record
	}
	return transformedInput, nil
}
