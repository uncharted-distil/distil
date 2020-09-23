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
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/xitongsys/parquet-go-source/local"
	"github.com/xitongsys/parquet-go/parquet"
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

type converter func(val interface{}) string

// NewParquet creates a new parquet backed storage.
func NewParquet() *Parquet {
	return &Parquet{}
}

// ReadDataset reads a raw dataset from the file system, loading the parquet
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
// the data to a parquet file.
func (d *Parquet) WriteDataset(uri string, data *api.RawDataset) error {
	err := d.WriteData(uri, data.Data)
	if err != nil {
		return err
	}

	err = d.WriteMetadata(uri, data.Metadata, true, true)
	if err != nil {
		return err
	}

	return nil
}

// ReadData reads the data from a parquet file.
func (d *Parquet) ReadData(uri string) ([][]string, error) {
	fr, err := local.NewLocalFileReader(uri)
	if err != nil {
		return nil, errors.Wrap(err, "unable to open parquet file")
	}
	defer fr.Close()

	pr, err := reader.NewParquetColumnReader(fr, 1)
	if err != nil {
		return nil, errors.Wrap(err, "unable to open parquet file reader")
	}
	defer pr.ReadStop()

	colCount := pr.SchemaHandler.GetColumnNum()
	rowCount := pr.GetNumRows()
	dataByCol := make([][]string, colCount)
	for i := int64(0); i < colCount; i++ {
		colRaw, err := d.readColumn(pr, i, rowCount)
		if err != nil {
			return nil, err
		}
		dataByCol[i] = d.columnToString(colRaw, *pr.SchemaHandler.SchemaElements[i+1].Type)
	}

	// header is the expected first row of the output
	header, err := d.ReadRawVariables(uri)
	if err != nil {
		return nil, err
	}

	output := make([][]string, rowCount+1)
	output[0] = header
	for rowIndex := 0; rowIndex < int(rowCount); rowIndex++ {
		outputRowIndex := rowIndex + 1
		output[outputRowIndex] = make([]string, colCount)
		for colIndex := 0; colIndex < int(colCount); colIndex++ {
			output[outputRowIndex][colIndex] = dataByCol[colIndex][rowIndex]
		}
	}

	return output, nil
}

func (d *Parquet) readColumn(pr *reader.ParquetReader, index int64, rows int64) ([]interface{}, error) {
	data, _, _, err := pr.ReadColumnByIndex(index, rows)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to read parquet file column index %d", index)
	}

	return data, nil
}

// WriteData writes data to a parquet file.
func (d *Parquet) WriteData(uri string, data [][]string) error {
	// create the containing folder
	// (ignore the error since the write failure will pick it up regardless)
	folder := path.Dir(uri)
	os.MkdirAll(folder, os.ModePerm)

	md := make([]string, len(data[0]))
	for i, c := range data[0] {
		md[i] = fmt.Sprintf("name=%s, type=UTF8", c)
	}

	fw, err := local.NewLocalFileWriter(uri)
	if err != nil {
		return errors.Wrapf(err, "unable to create parquet file '%s'", uri)
	}

	pw, err := writer.NewCSVWriter(md, fw, 1)
	if err != nil {
		return errors.Wrap(err, "unable to create parquet writer")
	}
	pw.CompressionType = parquet.CompressionCodec_UNCOMPRESSED

	//pw, err := writer.NewParquetWriter(fw, nil, 1)
	//pw.SchemaHandler = schema.NewSchemaHandlerFromMetadata(md)
	//pw.Footer.Schema = pw.SchemaHandler.SchemaElements
	//log.Infof("SCHEMA: %v", pw.SchemaHandler)
	//for i, se := range pw.SchemaHandler.SchemaElements {
	//	log.Infof("SCHEMA ELEMENT %d: %v", i, se)
	//}

	//if err != nil {
	//	return errors.Wrap(err, "unable to create parquet writer")
	//}

	for i := 1; i < len(data); i++ {
		rowData := data[i]
		row := make([]interface{}, len(rowData))
		for ic, c := range rowData {
			row[ic] = c
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

	defer fw.Close()

	return nil
}

// ReadRawVariables reads the metadata and extracts the field names.
func (d *Parquet) ReadRawVariables(uri string) ([]string, error) {
	fr, err := local.NewLocalFileReader(uri)
	if err != nil {
		return nil, errors.Wrap(err, "unable to open parquet file")
	}
	defer fr.Close()

	pr, err := reader.NewParquetColumnReader(fr, 1)
	if err != nil {
		return nil, errors.Wrap(err, "unable to open parquet file reader")
	}
	defer pr.ReadStop()

	// the first field info is the root field which is not a part of the dataset
	fields := make([]string, len(pr.SchemaHandler.Infos)-1)
	for i := 1; i < len(pr.SchemaHandler.Infos); i++ {
		fields[i-1] = pr.SchemaHandler.Infos[i].ExName
	}

	return fields, nil
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
func (d *Parquet) WriteMetadata(uri string, meta *model.Metadata, extended bool, update bool) error {
	dataResources := make([]interface{}, 0)

	// make sure the resource format and path match expected parquet types
	mainDR := meta.GetMainDataResource()
	if mainDR.ResFormat["application/parquet"] == nil {
		if !update {
			return errors.Errorf("main data resource not set to parquet format")
		} else {
			mainDR.ResFormat = map[string][]string{"application/parquet": {"parquet"}}
			mainDR.ResPath = fmt.Sprintf("%s.parquet", strings.TrimSuffix(mainDR.ResPath, path.Ext(mainDR.ResPath)))
		}
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

func (d *Parquet) columnToString(colData []interface{}, colType parquet.Type) []string {
	var converterFunc converter
	switch colType {
	case parquet.Type_FLOAT, parquet.Type_INT96, parquet.Type_DOUBLE:
		converterFunc = floatToString
	case parquet.Type_BOOLEAN:
		converterFunc = boolToString
	case parquet.Type_INT64:
		converterFunc = int64ToString
	case parquet.Type_INT32:
		converterFunc = int32ToString
	case parquet.Type_BYTE_ARRAY, parquet.Type_FIXED_LEN_BYTE_ARRAY:
		converterFunc = stringToString
	}

	output := make([]string, len(colData))
	for i, c := range colData {
		output[i] = converterFunc(c)
	}

	return output
}

func floatToString(val interface{}) string {
	return fmt.Sprintf("%f", val)
}

func stringToString(val interface{}) string {
	return val.(string)
}

func boolToString(val interface{}) string {
	return strconv.FormatBool(val.(bool))
}

func int32ToString(val interface{}) string {
	return strconv.Itoa(int(val.(int32)))
}

func int64ToString(val interface{}) string {
	return strconv.Itoa(int(val.(int64)))
}
