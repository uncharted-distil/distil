package routes

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"path"

	"github.com/pkg/errors"
	"goji.io/pat"

	"github.com/unchartedsoftware/distil-compute/model"
	"github.com/unchartedsoftware/distil-compute/primitive/compute"

	"github.com/unchartedsoftware/distil-ingest/metadata"
	"github.com/unchartedsoftware/distil/api/task"
	"github.com/unchartedsoftware/distil/api/util"
)

// UploadHandler uploads a file to the local file system and then imports it.
func UploadHandler(outputPath string, config *task.IngestTaskConfig) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		dataset := pat.Param(r, "dataset")
		outputDatasetPath := path.Join(outputPath, dataset)

		// read the file from the request
		bytes, err := receiveFile(r)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to receive file from request"))
			return
		}

		// save the csv file in the file system datasets folder
		dataFilePath := path.Join(compute.D3MDataFolder, compute.D3MLearningData)
		dataPath := path.Join(outputDatasetPath, dataFilePath)
		err = util.WriteFileWithDirs(dataPath, bytes, os.ModePerm)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to write raw data file"))
			return
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
			handleError(w, errors.Wrap(err, "unable to output schema"))
			return
		}

		// format the dataset into a D3M format
		formattedPath, err := task.Format(schemaPath, config)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to format datamart dataset"))
			return
		}

		// copy to the original output location for consistency
		if formattedPath != outputDatasetPath {
			err = os.RemoveAll(outputDatasetPath)
			if err != nil {
				handleError(w, errors.Wrap(err, "unable to delete raw upload dataset"))
				return
			}

			err = util.Copy(formattedPath, path.Dir(schemaPath))
			if err != nil {
				handleError(w, errors.Wrap(err, "unable to copy formatted datamart dataset"))
				return
			}
		}

		// marshal data and sent the response back
		err = handleJSON(w, map[string]interface{}{"result": "uploaded"})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal result histogram into JSON"))
			return
		}
	}
}

func receiveFile(r *http.Request) ([]byte, error) {
	file, _, err := r.FormFile("file")
	if err != nil {
		return nil, errors.Wrap(err, "unable to get file from request")
	}
	defer file.Close()

	// Copy the file data to the buffer
	var buf bytes.Buffer
	_, err = io.Copy(&buf, file)
	if err != nil {
		return nil, errors.Wrap(err, "unable to copy file")
	}

	return buf.Bytes(), nil
}
