package routes

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"goji.io/pat"

	api "github.com/uncharted-distil/distil-ingest/metadata"
	"github.com/uncharted-distil/distil/api/env"
	"github.com/uncharted-distil/distil/api/model"
)

const (
	timeseriesFolder = "timeseries"
)

// TimeseriesResult represents the result of a timeseries request.
type TimeseriesResult struct {
	Timeseries [][]float64 `json:"timeseries"`
}

// TimeseriesHandler provides a static file lookup route using simple directory mapping.
func TimeseriesHandler(ctor model.MetadataStorageCtor, resourceDir string, config *env.Config) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// resources can either be local or remote
		dataset := pat.Param(r, "dataset")
		source := pat.Param(r, "source")
		file := pat.Param(r, "file")
		path := path.Join(timeseriesFolder, file)

		// get metadata client
		storage, err := ctor()
		if err != nil {
			handleError(w, err)
			return
		}

		res, err := storage.FetchDataset(dataset, false, false)
		if err != nil {
			handleError(w, err)
			return
		}

		sourcePath := env.ResolvePath(api.DatasetSource(source), res.Folder)

		bytes, err := fetchResourceBytes(sourcePath, dataset, path)
		if err != nil {
			handleError(w, err)
			return
		}

		timeseriesCSV := string(bytes)

		reader := csv.NewReader(strings.NewReader(timeseriesCSV))

		// discard header
		if _, err := reader.Read(); err != nil {
			handleError(w, err)
			return
		}

		var points [][]float64

		for {
			record, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				handleError(w, err)
				return
			}
			if len(record) != 2 {
				handleError(w, fmt.Errorf("bad line in timeseries csv file %v", record))
				return
			}
			x, err := strconv.ParseFloat(record[0], 64)
			if err != nil {
				handleError(w, err)
				return
			}
			y, err := strconv.ParseFloat(record[1], 64)
			if err != nil {
				handleError(w, err)
				return
			}
			points = append(points, []float64{x, y})
		}

		err = handleJSON(w, TimeseriesResult{
			Timeseries: points,
		})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal dataset result into JSON"))
			return
		}
	}
}
