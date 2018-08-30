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
)

const (
	timeseriesFolder = "timeseries"
)

// TimeseriesResult represents the result of a timeseries request.
type TimeseriesResult struct {
	Timeseries [][]float64 `json:"timeseries"`
}

// TimeseriesHandler provides a static file lookup route using simple directory mapping.
func TimeseriesHandler(resourceDir string, proxyServer string, proxy map[string]bool) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// resources can either be local or remote
		dataset := pat.Param(r, "dataset")
		file := pat.Param(r, "file")
		path := path.Join(timeseriesFolder, file)

		bytes, err := fetchResourceBytes(resourceDir, proxyServer, proxy, dataset, path)
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
