package routes

import (
	"io/ioutil"
	"net/http"
	//"path"

	"github.com/pkg/errors"
	//"goji.io/pat"

	"github.com/unchartedsoftware/distil/api/util/graph"
)

const (
	graphsFolder = "graphs"
)

// GraphsResult represents the result of a graphs request.
type GraphsResult struct {
	Graphs []*graph.Graph `json:"graphs"`
}

// GraphsHandler provides a static file lookup route using simple directory mapping.
func GraphsHandler(resourceDir string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// resources can either be local or remote
		//dataset := pat.Param(r, "dataset")
		//file := pat.Param(r, "file")
		//path := path.Join(graphsFolder, file)

		// bytes, err := fetchResourceBytes(resourceDir, dataset, path)
		// if err != nil {
		// 	handleError(w, err)
		// 	return
		// }

		// DEBUG: mocked file
		bytes, err := ioutil.ReadFile("dist/graphs/G1.gml")
		if err != nil {
			handleError(w, err)
			return
		}

		graphs, err := graph.ParseGML(string(bytes))
		if err != nil {
			handleError(w, err)
			return
		}

		err = handleJSON(w, GraphsResult{
			Graphs: graphs,
		})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal graphs result into JSON"))
			return
		}
	}
}
