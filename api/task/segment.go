package task

import (
	"fmt"
	"os"
	"path"
	"strconv"

	"github.com/pkg/errors"

	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil-compute/primitive/compute/description"
	"github.com/uncharted-distil/distil-compute/primitive/compute/result"
	"github.com/uncharted-distil/distil/api/env"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/util"
	"github.com/uncharted-distil/distil/api/util/imagery"
)

// Segment segments an image into separate parts.
func Segment(ds *api.Dataset, dataStorage api.DataStorage, variableName string) (string, error) {
	envConfig, err := env.LoadConfig()
	if err != nil {
		return "", err
	}

	datasetInputDir := env.ResolvePath(ds.Source, ds.Folder)

	var variable *model.Variable
	for _, v := range ds.Variables {
		if v.Key == variableName {
			variable = v
			break
		}
	}

	step, err := description.CreateRemoteSensingSegmentationPipeline("segmentation", "basic image segmentation", variable, envConfig.RemoteSensingNumJobs)
	if err != nil {
		return "", err
	}

	resultURI, err := submitPipeline([]string{datasetInputDir}, step, true)
	if err != nil {
		return "", err
	}

	// read the file and parse the output mask
	result, err := result.ParseResultCSV(resultURI)
	if err != nil {
		return "", err
	}

	// need to pull the data to properly map d3m index to expected file names
	// filenames should be "groupid-segmentation.png" for now
	// TODO: may need to build the grouping key from multiple fields when moving away from test
	var groupingKey *model.Variable
	for _, v := range ds.Variables {
		if v.HasRole(model.VarDistilRoleGrouping) {
			groupingKey = v
			break
		}
	}
	if groupingKey == nil {
		return "", errors.Errorf("no grouping found to use for output filename")
	}
	mapping, err := getFieldMapping(ds, groupingKey.Key, dataStorage)
	if err != nil {
		return "", err
	}

	// need to output all the masks as images
	imageOutputFolder := path.Join(env.GetResourcePath(), ds.ID, "media")
	for _, r := range result[1:] {
		// create the image that captures the mask
		d3mIndex := r[0].(string)
		rawMask := r[1].([]interface{})
		rawFloats := make([][]float64, len(rawMask))
		for i, f := range rawMask {
			dataF := f.([]interface{})
			nestedFloats := make([]float64, len(dataF))
			for j, nf := range dataF {
				fp, err := strconv.ParseFloat(nf.(string), 64)
				if err != nil {
					return "", errors.Wrapf(err, "unable to parse mask")
				}
				nestedFloats[j] = fp
			}
			rawFloats[i] = nestedFloats
		}

		filter := imagery.ConfidenceMatrixToImage(rawFloats, imagery.MagmaColorScale, uint8(100))
		imageBytes, err := imagery.ImageToPNG(filter)
		if err != nil {
			return "", err
		}

		// write the image to disk using a basic naming convention
		imageFilename := path.Join(imageOutputFolder, fmt.Sprintf("%s-segmentation.png", mapping[d3mIndex]))
		err = util.WriteFileWithDirs(imageFilename, imageBytes, os.ModePerm)
		if err != nil {
			return "", err
		}
	}

	return "", nil
}

func getFieldMapping(ds *api.Dataset, fieldName string, dataStorage api.DataStorage) (map[string]string, error) {
	filter := &api.FilterParams{Variables: []string{model.D3MIndexFieldName, fieldName}}

	// pull back all rows for a group id
	data, err := dataStorage.FetchData(ds.ID, ds.StorageName, filter, true, nil)
	if err != nil {
		return nil, err
	}

	// cycle through results to build the band mapping
	fieldColumn, ok := data.Columns[fieldName]
	if !ok {
		return nil, errors.Errorf("'%s' column not found in stored data", fieldName)
	}
	fieldColumnIndex := fieldColumn.Index
	d3mColumn, ok := data.Columns[model.D3MIndexFieldName]
	if !ok {
		return nil, errors.Errorf("'%s' column not found in stored data", model.D3MIndexFieldName)
	}
	d3mColumnIndex := d3mColumn.Index

	mapping := map[string]string{}
	for _, r := range data.Values {
		d3mIndexData := fmt.Sprintf("%.0f", r[d3mColumnIndex].Value.(float64))
		fieldData := r[fieldColumnIndex].Value.(string)
		mapping[d3mIndexData] = fieldData
	}

	return mapping, nil
}
