package routes

import (
	"fmt"
	"net/http"
	"path"

	"github.com/pkg/errors"
	"goji.io/pat"

	"github.com/unchartedsoftware/distil-compute/model"
	api "github.com/unchartedsoftware/distil/api/model"
	"github.com/unchartedsoftware/distil/api/task"
)

// GeocodingResult represents a geocoding response for a variable.
type GeocodingResult struct {
	LatitudeField  string `json:"latitude"`
	LongitudeField string `json:"longitude"`
}

// GeocodingHandler generates a route handler that enables geocoding
// of a variable and the creation of two new columns to hold the lat and lon.
func GeocodingHandler(metaCtor api.MetadataStorageCtor, dataCtor api.DataStorageCtor, sourceFolder string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// get dataset name
		dataset := pat.Param(r, "dataset")
		// get variable name
		variable := pat.Param(r, "variable")

		// get storage clients
		metaStorage, err := metaCtor()
		if err != nil {
			handleError(w, err)
			return
		}
		dataStorage, err := dataCtor()
		if err != nil {
			handleError(w, err)
			return
		}
		latVarName := fmt.Sprintf("_lat_%s", variable)
		lonVarName := fmt.Sprintf("_lon_%s", variable)

		// check if the lat and lon variables exist
		// NOTE: ignore the errors,
		latVarExist, err := metaStorage.DoesVariableExist(dataset, latVarName)
		if err != nil {
			handleError(w, err)
			return
		}
		lonVarExist, err := metaStorage.DoesVariableExist(dataset, lonVarName)
		if err != nil {
			handleError(w, err)
			return
		}

		// create the new metadata and database variables
		if !latVarExist {
			err = metaStorage.AddVariable(dataset, latVarName, model.LatitudeType, "geocoding")
			if err != nil {
				handleError(w, err)
				return
			}
			err = dataStorage.AddVariable(dataset, latVarName, model.LatitudeType)
			if err != nil {
				handleError(w, err)
				return
			}
		}
		if !lonVarExist {
			err = metaStorage.AddVariable(dataset, lonVarName, model.LongitudeType, "geocoding")
			if err != nil {
				handleError(w, err)
				return
			}
			err = dataStorage.AddVariable(dataset, lonVarName, model.LongitudeType)
			if err != nil {
				handleError(w, err)
				return
			}
		}

		// build the row index since geocoding does not return the d3m index
		lines, err := task.ReadCSVFile(path.Join(sourceFolder, "tables", "learningData.csv"), true)
		if err != nil {
			handleError(w, err)
			return
		}

		d3mIndex, _ := metaStorage.FetchVariable(dataset, model.D3MIndexName)
		rowIndex := make(map[int]string)
		for i, line := range lines {
			rowIndex[i] = line[d3mIndex.Index]
		}

		// geocode data
		geocoded, err := task.GeocodeForward(sourceFolder, dataset, variable, rowIndex)
		if err != nil {
			handleError(w, err)
			return
		}

		for _, point := range geocoded {
			err = dataStorage.UpdateVariable(dataset, latVarName, point.D3MIndex, fmt.Sprintf("%f", point.Latitude))
			if err != nil {
				handleError(w, err)
				return
			}
			err = dataStorage.UpdateVariable(dataset, lonVarName, point.D3MIndex, fmt.Sprintf("%f", point.Longitude))
			if err != nil {
				handleError(w, err)
				return
			}
		}

		// marshal output into JSON
		err = handleJSON(w, GeocodingResult{
			LatitudeField:  latVarName,
			LongitudeField: lonVarName,
		})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal geocoded result into JSON"))
			return
		}
	}
}
