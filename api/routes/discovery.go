package routes

import (
	"net/http"

	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil/api/model"
	"github.com/unchartedsoftware/distil/api/pipeline"
	"github.com/unchartedsoftware/distil/api/util/json"
	"goji.io/pat"
)

// ProblemDiscoveryHandler creates a route that saves a discovered problem.
func ProblemDiscoveryHandler(ctorData model.DataStorageCtor, ctorMeta model.MetadataStorageCtor, datasetDir string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		dataset := pat.Param(r, "dataset")
		esIndex := pat.Param(r, "esIndex")
		target := pat.Param(r, "target")

		// get variable names and ranges out of the params
		filterParams, err := model.ParseFilterParamsURL(r.URL.Query())
		if err != nil {
			handleError(w, err)
			return
		}

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

		fetchFilteredData := func(dataset string, index string, filterParams *model.FilterParams) (*model.FilteredData, error) {
			// fetch the whole data and include the target feature
			filterParams.Filters = append(filterParams.Filters, model.NewEmptyFilter(target))
			return dataStorage.FetchData(dataset, index, filterParams, false)
		}
		fetchVariable := func(dataset string, index string) ([]*model.Variable, error) {
			return metadataStorage.FetchVariables(dataset, index, false)
		}

		path, err := pipeline.PersistFilteredData(fetchFilteredData, fetchVariable, datasetDir, dataset, esIndex, target, filterParams)
		if err != nil {
			handleError(w, err)
			return
		}

		// marshall output into JSON
		bytes, err := json.Marshal(map[string]interface{}{"result": "discovered", "path": path})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal filtered data result into JSON"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(bytes)
	}
}
