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

	api "github.com/unchartedsoftware/distil-ingest/metadata"
	"github.com/unchartedsoftware/distil/api/env"
	"github.com/unchartedsoftware/distil/api/model"
	"github.com/unchartedsoftware/distil/api/util"
)

const (
	timeseriesFolder = "timeseries"
)

// TimeseriesResult represents the result of a timeseries request.
type TimeseriesResult struct {
	Timeseries [][]float64 `json:"timeseries"`
}

// TimeseriesHandler provides a static file lookup route using simple directory mapping.
func TimeseriesHandler(ctor model.MetadataStorageCtor, resourceDir string, proxyServer string, proxy map[string]bool, config *env.Config) func(http.ResponseWriter, *http.Request) {
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

		resolver := createResolverForResource(api.DatasetSource(source), res.Folder, config)

		bytes, err := fetchResourceBytes(resolver.ResolveInputAbsolute(""), proxyServer, proxy, dataset, path)
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

func createResolverForResource(datasetSource api.DatasetSource, datasetFolder string, config *env.Config) *util.PathResolver {
	if datasetSource == api.Contrib {
		return util.NewPathResolver(&util.PathConfig{
			InputFolder:  path.Join(config.DatamartImportFolder, datasetFolder),
			OutputFolder: path.Join(config.DatamartImportFolder, datasetFolder),
		})
	}
	if datasetSource == api.Seed {
		return util.NewPathResolver(&util.PathConfig{
			InputFolder:     path.Join(config.D3MInputDir, datasetFolder),
			InputSubFolders: "TRAIN/dataset_TRAIN",
			OutputFolder:    config.D3MInputDir,
		})
	}
	if datasetSource == api.Augmented {
		return util.NewPathResolver(&util.PathConfig{
			InputFolder:  path.Join(config.TmpDataPath, config.AugmentedSubFolder, datasetFolder),
			OutputFolder: path.Join(config.TmpDataPath, config.AugmentedSubFolder, datasetFolder),
		})
	}
	return util.NewPathResolver(&util.PathConfig{
		InputFolder:     config.D3MInputDir,
		InputSubFolders: "TRAIN/dataset_TRAIN",
		OutputFolder:    config.D3MInputDir,
	})
}
