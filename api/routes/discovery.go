package routes

import (
	"fmt"
	"net/http"
	"os"
	"path"

	"github.com/pkg/errors"
	"github.com/unchartedsoftware/plog"
	"goji.io/pat"

	"github.com/unchartedsoftware/distil/api/compute"
	"github.com/unchartedsoftware/distil/api/model"
	"github.com/unchartedsoftware/distil/api/util"
	"github.com/unchartedsoftware/distil/api/util/json"
)

const (
	apiExportFile     = "ssapi.json"
	problemSchemaFile = "schema.json"

	// ProblemLabelFile is the file listing the exported problems.
	ProblemLabelFile = "labels.csv"
)

// ProblemDiscoveryHandler creates a route that saves a discovered problem.
func ProblemDiscoveryHandler(ctorData model.DataStorageCtor, ctorMeta model.MetadataStorageCtor, problemDir string, userAgent string, skipPrepends bool) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		dataset := pat.Param(r, "dataset")
		target := pat.Param(r, "target")

		// parse POST params
		params, err := getPostParameters(r)
		if err != nil {
			handleError(w, errors.Wrap(err, "Unable to parse post parameters"))
			return
		}
		meaningful, ok := params["meaningful"].(string)
		if !ok {
			meaningful = "no"
		}

		// get variable names and ranges out of the params
		filterParams, err := model.ParseFilterParamsFromJSON(params["filterParams"].(map[string]interface{}))
		if err != nil {
			handleError(w, err)
			return
		}
		filterParams.Size = -1

		// NOTE: D3M index field is needed in the persisted data.
		filterParams.Variables = append(filterParams.Variables, model.D3MIndexFieldName)

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

		ds, err := model.FetchDataset(dataset, true, true, filterParams, metadataStorage, dataStorage)
		if err != nil {
			handleError(w, err)
			return
		}

		problem, problemID, err := compute.CreateProblemSchema(problemDir, dataset, targetVar, filterParams)
		if err != nil {
			handleError(w, err)
			return
		}

		problemOutputDirectory := path.Join(problemDir, problemID)
		err = os.MkdirAll(problemOutputDirectory, 0777)
		if err != nil {
			handleError(w, err)
			return
		}
		log.Infof("Writing problem information to %s", problemOutputDirectory)

		problemJSON, err := json.Marshal(problem)
		if err != nil {
			handleError(w, err)
			return
		}

		problemSchemaOutputFile := path.Join(problemOutputDirectory, problemSchemaFile)
		err = util.WriteFileWithDirs(problemSchemaOutputFile, problemJSON, os.ModePerm)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to write problem schema"))
			return
		}

		// store the search solution request for this problem
		request, err := compute.CreateSearchSolutionRequest(ds.Metadata.Variables, filterParams.Variables, target, problemDir, dataset, userAgent, skipPrepends)
		if err != nil {
			handleError(w, err)
			return
		}

		requestJSON, err := json.Marshal(request)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to marshal search solution request into JSON"))
			return
		}

		problemAPIExportFile := path.Join(problemOutputDirectory, apiExportFile)
		err = util.WriteFileWithDirs(problemAPIExportFile, requestJSON, os.ModePerm)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to write search solution request"))
			return
		}

		// update the problem listing
		// the listing is shared between all problems
		// need to append a row to the listing
		problemListingFile := path.Join(problemDir, ProblemLabelFile)
		problemLabel := fmt.Sprintf("%s,\"user\",\"%s\"\n", problemID, meaningful)
		f, err := os.OpenFile(problemListingFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModePerm)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to open problem listing"))
			return
		}
		_, err = f.Write([]byte(problemLabel))
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to write new problem to listing"))
			return
		}
		err = f.Close()
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to close problem listing"))
			return
		}

		// marshall output into JSON
		bytes, err := json.Marshal(map[string]interface{}{"result": "discovered", "problemPath": problemSchemaOutputFile, "apiPath": problemAPIExportFile})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal result into JSON"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(bytes)
	}
}
