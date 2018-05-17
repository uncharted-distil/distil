package routes

import (
	"net/http"

	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil/api/compute"
	"github.com/unchartedsoftware/distil/api/model"
	"github.com/unchartedsoftware/distil/api/util/json"
	"goji.io/pat"
)

// ProblemDiscoveryHandler creates a route that saves a discovered problem.
func ProblemDiscoveryHandler(ctorData model.DataStorageCtor, ctorMeta model.MetadataStorageCtor, datasetDir string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		dataset := pat.Param(r, "dataset")
		esIndex := pat.Param(r, "index")
		target := pat.Param(r, "target")

		// parse POST params
		params, err := getPostParameters(r)
		if err != nil {
			handleError(w, errors.Wrap(err, "Unable to parse post parameters"))
			return
		}

		// get variable names and ranges out of the params
		filterParams, err := model.ParseFilterParamsFromJSON(params)
		if err != nil {
			handleError(w, err)
			return
		}
		filterParams.Size = -1

		// NOTE: D3M index field is needed in the persisted data.
		filterParams.Variables = append(filterParams.Variables, "d3mIndex")

		// get storages
		dataStorage, err := ctorData()
		if err != nil {
			handleError(w, err)
			return
		}

		metadataStorage, err := ctorMeta()
		if err != nil {
			handleError(w, err)
			return
		}

		targetVar, err := metadataStorage.FetchVariable(dataset, target)
		if err != nil {
			handleError(w, err)
			return
		}

		ds, err := model.FetchDataset(dataset, esIndex, true, filterParams, metadataStorage, dataStorage)
		if err != nil {
			handleError(w, err)
			return
		}

		path, _, err := compute.PersistFilteredData(datasetDir, target, ds)
		if err != nil {
			handleError(w, err)
			return
		}

		pathProblem, err := compute.PersistProblem(datasetDir, dataset, targetVar, filterParams)
		if err != nil {
			handleError(w, err)
			return
		}

		// marshall output into JSON
		bytes, err := json.Marshal(map[string]interface{}{"result": "discovered", "datasetPath": path, "problemPath": pathProblem})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal filtered data result into JSON"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(bytes)
	}
}
