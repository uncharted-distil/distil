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

package task

import (
	"bytes"
	"encoding/csv"
	"os"
	"path"

	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/metadata"
	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil-compute/primitive/compute/description"
	"github.com/uncharted-distil/distil-compute/primitive/compute/result"

	"github.com/uncharted-distil/distil/api/util"
)

const (
	unicornResultFieldName = "label"
	slothResultFieldName   = "cluster_labels"
)

// ClusterPoint contains data that has been clustered.
type ClusterPoint struct {
	D3MIndex    string
	SourceField string
	Label       string
}

// ClusterDataset will cluster the dataset fields using a primitive.
func ClusterDataset(datasetSource metadata.DatasetSource, schemaFile string, index string, dataset string, config *IngestTaskConfig) (string, error) {
	outputPath, err := initializeDatasetCopy(schemaFile, dataset, config.ClusteringOutputSchemaRelative, config.ClusteringOutputDataRelative)
	if err != nil {
		return "", errors.Wrap(err, "unable to copy source data folder")
	}

	// load metadata from original schema
	meta, err := metadata.LoadMetadataFromOriginalSchema(schemaFile)
	if err != nil {
		return "", errors.Wrap(err, "unable to load original schema file")
	}
	mainDR := meta.GetMainDataResource()

	// add feature variables
	features, err := getClusterVariables(meta, model.ClusterVarPrefix)
	if err != nil {
		return "", errors.Wrap(err, "unable to get cluster variables")
	}

	d3mIndexField := getD3MIndexField(mainDR)

	// open the input file
	dataPath := path.Join(outputPath.sourceFolder, mainDR.ResPath)
	lines, err := util.ReadCSVFile(dataPath, config.HasHeader)
	if err != nil {
		return "", errors.Wrap(err, "error reading raw data")
	}

	// add the cluster data to the raw data
	for _, f := range features {
		mainDR.Variables = append(mainDR.Variables, f.Variable)

		// header already removed, lines does not have a header
		lines, err = appendFeature(outputPath.outputFolder, d3mIndexField, false, f, lines)
		if err != nil {
			return "", errors.Wrap(err, "error appending clustered data")
		}
	}

	// initialize csv writer
	output := &bytes.Buffer{}
	writer := csv.NewWriter(output)

	// output the header
	header := make([]string, len(mainDR.Variables))
	for _, v := range mainDR.Variables {
		header[v.Index] = v.Name
	}
	err = writer.Write(header)
	if err != nil {
		return "", errors.Wrap(err, "error storing clustered header")
	}

	for _, line := range lines {
		err = writer.Write(line)
		if err != nil {
			return "", errors.Wrap(err, "error storing clustered output")
		}
	}

	// output the data with the new feature
	writer.Flush()

	err = util.WriteFileWithDirs(outputPath.outputData, output.Bytes(), os.ModePerm)
	if err != nil {
		return "", errors.Wrap(err, "error writing clustered output")
	}

	relativePath := getRelativePath(path.Dir(outputPath.outputSchema), outputPath.outputData)
	mainDR.ResPath = relativePath

	// write the new schema to file
	err = metadata.WriteSchema(meta, outputPath.outputSchema, true)
	if err != nil {
		return "", errors.Wrap(err, "unable to store cluster schema")
	}

	return outputPath.outputSchema, nil
}

// Cluster will cluster the dataset fields using a primitive.
func Cluster(datasetInputDir string, dataset string, variable string) ([]*ClusterPoint, error) {
	step, err := description.CreateSlothPipeline("time series clustering",
		"k-means time series clustering", "", "", nil)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create sloth pipeline")
	}

	datasetURI, err := submitPipeline([]string{datasetInputDir}, step)
	if err != nil {
		return nil, errors.Wrap(err, "unable to run pipeline primitive")
	}

	// parse primitive response (new field contains output)
	res, err := result.ParseResultCSV(datasetURI)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse pipeline primitive result")
	}
	header, err := castTypeArray(res[0])
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse Sloth pipeline header")
	}

	// find the field with the feature output
	clusterIndex := getFieldIndex(header, "__cluster")
	d3mIndexIndex := getFieldIndex(header, model.D3MIndexName)

	// build the output (skipping the header)
	clusteredData := make([]*ClusterPoint, len(res)-1)
	for i, v := range res[1:] {
		label := createFriendlyLabel(v[clusterIndex].(string))
		d3mIndex := v[d3mIndexIndex].(string)

		clusteredData[i] = &ClusterPoint{
			D3MIndex:    d3mIndex,
			SourceField: variable,
			Label:       label,
		}
	}

	return clusteredData, nil
}
