//
//   Copyright © 2019 Uncharted Software Inc.
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package routes

import (
	"io/ioutil"
	"net/http"

	//"path"

	"github.com/pkg/errors"
	//"goji.io/v3/pat"

	"github.com/uncharted-distil/distil/api/util/graph"
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
