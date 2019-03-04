package task

import (
	"os"
	"path"

	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil-compute/primitive/compute"

	"github.com/uncharted-distil/distil-ingest/metadata"
	"github.com/unchartedsoftware/distil/api/util"
)

// CreateDataset structures a raw csv file into a valid D3M dataset.
func CreateDataset(dataset string, csvData []byte, outputPath string, config *IngestTaskConfig) (string, error) {
	// save the csv file in the file system datasets folder
	outputDatasetPath := path.Join(outputPath, dataset)
	dataFilePath := path.Join(compute.D3MDataFolder, compute.D3MLearningData)
	dataPath := path.Join(outputDatasetPath, dataFilePath)
	err := util.WriteFileWithDirs(dataPath, csvData, os.ModePerm)
	if err != nil {
		return "", err
	}

	// create the raw dataset schema doc
	datasetID := model.NormalizeDatasetID(dataset)
	meta := model.NewMetadata(dataset, dataset, "", datasetID)
	dr := model.NewDataResource("0", model.ResTypeRaw, []string{compute.D3MResourceFormat})
	dr.ResPath = dataFilePath
	meta.DataResources = []*model.DataResource{dr}

	schemaPath := path.Join(outputDatasetPath, compute.D3MDataSchema)
	err = metadata.WriteSchema(meta, schemaPath)
	if err != nil {
		return "", err
	}

	// format the dataset into a D3M format
	formattedPath, err := Format(metadata.Contrib, schemaPath, config)
	if err != nil {
		return "", err
	}

	// copy to the original output location for consistency
	if formattedPath != outputDatasetPath {
		err = os.RemoveAll(outputDatasetPath)
		if err != nil {
			return "", err
		}

		err = util.Copy(formattedPath, path.Dir(schemaPath))
		if err != nil {
			return "", err
		}
	}

	return formattedPath, nil
}
