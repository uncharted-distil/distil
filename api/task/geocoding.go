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

package task

import (
	"fmt"
	"path"

	"github.com/pkg/errors"

	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil-compute/primitive/compute"
	"github.com/uncharted-distil/distil-compute/primitive/compute/description"
	"github.com/uncharted-distil/distil-compute/primitive/compute/result"

	"github.com/uncharted-distil/distil-compute/metadata"

	"github.com/uncharted-distil/distil/api/serialization"
	"github.com/uncharted-distil/distil/api/util"
)

// GeocodedPoint contains data that has been geocoded.
type GeocodedPoint struct {
	D3MIndex    string
	SourceField string
	Latitude    string
	Longitude   string
}

// GeocodeForwardDataset geocodes fields that are types of locations.
// The results are append to the dataset and the whole is output to disk.
func GeocodeForwardDataset(schemaFile string, dataset string, config *IngestTaskConfig) (string, error) {
	outputPath := createDatasetPaths(schemaFile, dataset, compute.D3MLearningData)

	// load metadata from original schema
	meta, err := metadata.LoadMetadataFromClassification(schemaFile, path.Join(path.Dir(schemaFile), config.ClassificationOutputPathRelative), false, true)
	if err != nil {
		return "", errors.Wrap(err, "unable to load original schema file")
	}
	mainDR := meta.GetMainDataResource()
	d3mIndexVariable := getD3MIndexField(mainDR)

	// read raw data
	dataPath := path.Join(outputPath.sourceFolder, mainDR.ResPath)
	lines, err := util.ReadCSVFile(dataPath, config.HasHeader)
	if err != nil {
		return "", errors.Wrap(err, "error reading raw data")
	}

	// index d3m indices by row since primitive returns row numbers.
	// header row already skipped in the readCSVFile call.
	rowIndex := make(map[int]string)
	for i, line := range lines {
		rowIndex[i] = line[d3mIndexVariable]
	}

	// Geocode location fields
	datasetInputDir := outputPath.sourceFolder
	colsToGeocode := geocodeColumns(meta)
	geocodedData := make([][]*GeocodedPoint, 0)
	for _, col := range colsToGeocode {
		geocoded, err := GeocodeForward(datasetInputDir, dataset, col)
		if err != nil {
			return "", err
		}
		geocodedData = append(geocodedData, geocoded)
	}

	// map geocoded data by d3m index
	indexedData := make(map[string][]*GeocodedPoint)
	fields := make(map[string][]*model.Variable)
	for _, field := range geocodedData {
		latName, lonName := getLatLonVariableNames(field[0].SourceField)
		latDesc := fmt.Sprintf("latitude obtained from field %s", field[0].SourceField)
		lonDesc := fmt.Sprintf("longitude obtained from field %s", field[0].SourceField)
		fields[field[0].SourceField] = []*model.Variable{
			model.NewVariable(len(mainDR.Variables), latName, "label", latName, latName, model.LatitudeType, model.LatitudeType, latDesc, []string{"attribute"}, model.VarDistilRoleMetadata, nil, mainDR.Variables, false),
			model.NewVariable(len(mainDR.Variables)+1, lonName, "label", latName, lonName, model.LongitudeType, model.LongitudeType, lonDesc, []string{"attribute"}, model.VarDistilRoleMetadata, nil, mainDR.Variables, false),
		}
		mainDR.Variables = append(mainDR.Variables, fields[field[0].SourceField]...)
		for _, gc := range field {
			if indexedData[gc.D3MIndex] == nil {
				indexedData[gc.D3MIndex] = make([]*GeocodedPoint, 0)
			}
			indexedData[gc.D3MIndex] = append(indexedData[gc.D3MIndex], gc)
		}
	}

	// add the geocoded data to the raw data
	for i, line := range lines {
		geocodedFields := indexedData[line[d3mIndexVariable]]
		for _, geo := range geocodedFields {
			line = append(line, geo.Latitude)
			line = append(line, geo.Longitude)
		}
		lines[i] = line
	}

	// output the header
	header := make([]string, len(mainDR.Variables))
	for _, v := range mainDR.Variables {
		header[v.Index] = v.HeaderName
	}
	output := [][]string{header}
	output = append(output, lines...)

	// output the data with the new feature
	datasetStorage := serialization.GetStorage(outputPath.outputData)
	err = datasetStorage.WriteData(outputPath.outputData, output)
	if err != nil {
		return "", errors.Wrap(err, "error writing feature output")
	}
	mainDR.ResPath = path.Dir(outputPath.outputData)

	// write the new schema to file
	err = datasetStorage.WriteMetadata(outputPath.outputSchema, meta, true, false)
	if err != nil {
		return "", errors.Wrap(err, "unable to store feature schema")
	}

	return outputPath.outputSchema, nil
}

// GeocodeForward will geocode a column into lat & lon values.
func GeocodeForward(datasetInputDir string, dataset string, variable *model.Variable) ([]*GeocodedPoint, error) {

	// create & submit the solution request
	pip, err := description.CreateGoatForwardPipeline("mountain", "", variable)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create Goat pipeline")
	}

	datasetURI, err := submitPipeline([]string{datasetInputDir}, pip, true)
	if err != nil {
		return nil, errors.Wrap(err, "unable to run Goat pipeline")
	}

	// parse primitive response (col index,importance)
	res, err := result.ParseResultCSV(datasetURI)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse Goat pipeline result")
	}

	// result should be d3m index, lat, lon
	header, err := castTypeArray(res[0])
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse Goat pipeline header")
	}

	// skip the header
	res = res[1:]
	geocodedData := make([]*GeocodedPoint, len(res))

	latIndex := getFieldIndex(header, fmt.Sprintf("%s_latitude", variable.HeaderName))
	lonIndex := getFieldIndex(header, fmt.Sprintf("%s_longitude", variable.HeaderName))
	d3mIndexIndex := getFieldIndex(header, model.D3MIndexFieldName)
	for i, v := range res {
		lat := v[latIndex].(string)
		lon := v[lonIndex].(string)

		d3mIndex := v[d3mIndexIndex].(string)

		geocodedData[i] = &GeocodedPoint{
			D3MIndex:    d3mIndex,
			SourceField: variable.Key,
			Latitude:    lat,
			Longitude:   lon,
		}
	}

	return geocodedData, nil
}

func getLatLonVariableNames(variableName string) (string, string) {
	lat := fmt.Sprintf("_lat_%s", variableName)
	lon := fmt.Sprintf("_lon_%s", variableName)

	return lat, lon
}

func geocodeColumns(meta *model.Metadata) []*model.Variable {
	// cycle throught types to determine columns to geocode.
	colsToGeocode := make([]*model.Variable, 0)
	for _, v := range meta.DataResources[0].Variables {
		for _, t := range v.SuggestedTypes {
			if isLocationType(t.Type) {
				colsToGeocode = append(colsToGeocode, v)
			}
		}
	}

	return colsToGeocode
}

func isLocationType(typ string) bool {
	return typ == model.AddressType || typ == model.CityType || typ == model.CountryType ||
		typ == model.PostalCodeType || typ == model.StateType || typ == model.TA2LocationType
}
