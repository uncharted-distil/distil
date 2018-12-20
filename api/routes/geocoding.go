package routes

import (
	"fmt"
	"net/http"

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
func GeocodingHandler(metaCtor api.MetadataStorageCtor, dataCtor api.DataStorageCtor) func(http.ResponseWriter, *http.Request) {
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

		// create the new metadata variables
		err = metaStorage.AddVariable(dataset, latVarName, model.LatitudeType, "geocoding")
		if err != nil {
			handleError(w, err)
			return
		}
		err = metaStorage.AddVariable(dataset, latVarName, model.LongitudeType, "geocoding")
		if err != nil {
			handleError(w, err)
			return
		}

		// create the database variables
		err = dataStorage.AddVariable(dataset, latVarName, model.LatitudeType)
		if err != nil {
			handleError(w, err)
			return
		}
		err = dataStorage.AddVariable(dataset, latVarName, model.LongitudeType)
		if err != nil {
			handleError(w, err)
			return
		}

		d, err := metaStorage.FetchDataset(dataset, false, false)
		if err != nil {
			handleError(w, err)
			return
		}

		// build the row index since geocoding does not return the d3m index
		lines, err := task.ReadCSVFile(d.Folder, true)
		if err != nil {
			handleError(w, err)
			return
		}

		d3mIndex := d.GetD3MIndexVariable()
		rowIndex := make(map[int]string)
		for i, line := range lines {
			rowIndex[i] = line[d3mIndex.Index]
		}

		// geocode data
		geocoded, err := task.GeocodeForward(d.Folder, dataset, variable, rowIndex)
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

		// marshall output into JSON
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
