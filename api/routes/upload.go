package routes

import (
	"bytes"
	"io"
	"net/http"

	"github.com/pkg/errors"
	log "github.com/unchartedsoftware/plog"
	"goji.io/pat"

	"github.com/uncharted-distil/distil/api/task"
)

// UploadHandler uploads a file to the local file system and then imports it.
func UploadHandler(outputPath string, config *task.IngestTaskConfig) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		dataset := pat.Param(r, "dataset")

		// read the file from the request
		bytes, err := receiveFile(r)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to receive file from request"))
			return
		}

		// create the raw dataset schema doc
		formattedPath, err := task.CreateDataset(dataset, bytes, outputPath, config)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to create dataset"))
			return
		}
		log.Infof("uploaded new dataset %s at %s", dataset, formattedPath)
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
