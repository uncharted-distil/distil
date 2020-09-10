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
	"encoding/json"
	"io/ioutil"

	"github.com/xitongsys/parquet-go-source/local"
	"github.com/xitongsys/parquet-go/reader"
	"github.com/xitongsys/parquet-go/writer"

	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/metadata"
	"github.com/uncharted-distil/distil-compute/model"
	api "github.com/uncharted-distil/distil/api/model"
)

// Parquet represents a dataset storage backed with parquet data and json schema doc.
type Parquet struct {
}

type parquetRow struct {
	data []string `parquet:"name=data, type=SLICE, valuetype=UTF8"`
}

// NewParquet creates a new parquet backed storage.
func NewParquet() *Parquet {
	return &Parquet{}
}

// ReadDataset reads a raw dataset from the file system, loading the csv
// data into memory.
func (d *Parquet) ReadDataset(uri string) (*api.RawDataset, error) {
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
func (d *Parquet) WriteDataset(uri string, data *api.RawDataset) error {
	err := d.WriteData(uri, data.Data)
	if err != nil {
		return err
	}

	err = d.WriteMetadata(uri, data.Metadata, true)
	if err != nil {
		return err
	}

	return nil
}

// ReadData reads the data from a csv file.
func (d *Parquet) ReadData(uri string) ([][]string, error) {
	fr, err := local.NewLocalFileReader(uri)
	if err != nil {
		return nil, errors.Wrap(err, "unable to open parquet file")
	}

	pr, err := reader.NewParquetReader(fr, new(parquetRow), 1)
	if err != nil {
		return nil, errors.Wrap(err, "unable to open parquet file reader")
	}

	num := int(pr.GetNumRows())
	rows := make([]parquetRow, num)
	err = pr.Read(&rows)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read parquet data")
	}
	pr.ReadStop()
	fr.Close()

	output := make([][]string, len(rows))
	for i, row := range rows {
		output[i] = row.data
	}

	return output, nil
}

// WriteData writes data to a csv file.
func (d *Parquet) WriteData(uri string, data [][]string) error {
	fw, err := local.NewLocalFileWriter(uri)
	if err != nil {
		return errors.Wrapf(err, "unable to create parquet file '%s'", uri)
	}

	//write
	pw, err := writer.NewParquetWriter(fw, new(parquetRow), 1)
	if err != nil {
		return errors.Wrap(err, "unable to create parquet writer")
	}
	defer fw.Close()

	for i := 0; i < len(data); i++ {
		row := &parquetRow{
			data: data[i],
		}
		err = pw.Write(row)
		if err != nil {
			return errors.Wrap(err, "error writing data to parquet file")
		}
	}
	err = pw.WriteStop()
	if err != nil {
		return errors.Wrap(err, "error ending parquet write")
	}

	return nil
}

// ReadMetadata reads the dataset doc from disk.
func (d *Parquet) ReadMetadata(uri string) (*model.Metadata, error) {
	meta, err := metadata.LoadMetadataFromOriginalSchema(uri, true)
	if err != nil {
		return nil, err
	}

	return meta, nil
}

// WriteMetadata writes the dataset doc to disk.
func (d *Parquet) WriteMetadata(uri string, meta *model.Metadata, extended bool) error {
	dataResources := make([]interface{}, 0)
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

func (d *Parquet) writeDataResource(resource *model.DataResource, extended bool) map[string]interface{} {
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

func (d *Parquet) writeVariable(variable *model.Variable, extended bool) interface{} {
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
