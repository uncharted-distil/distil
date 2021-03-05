//
//   Copyright © 2021 Uncharted Software Inc.
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
	"net/http"
	"net/url"

	"github.com/pkg/errors"
	"goji.io/v3/pat"

	"github.com/uncharted-distil/distil/api/model"
)

// ModelHandler generates a route handler that returns a specified model summary.
func ModelHandler(ctor model.ExportedModelStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		// get dataset name
		model := pat.Param(r, "model")

		// get metadata client
		storage, err := ctor()
		if err != nil {
			handleError(w, err)
			return
		}

		// get model summary
		res, err := storage.FetchModel(model)
		if err != nil {
			handleError(w, err)
			return
		}

		// marshal data
		err = handleJSON(w, res)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to marshal model result into JSON"))
			return
		}
	}
}

// ModelsHandler generates a route handler that facilitates a search of
// model & dataset descriptions, and variable names, returning a name, description and
// variable list for any model that matches. The search parameter is optional
// it contains the search terms if set, and if unset, flags that a list of all
// models should be returned.
func ModelsHandler(modelCtor model.ExportedModelStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// check for search terms
		terms, err := url.QueryUnescape(r.URL.Query().Get("search"))
		if err != nil {
			handleError(w, errors.Wrap(err, "malformed models query"))
			return
		}

		storage, err := modelCtor()
		if err != nil {
			handleError(w, err)
			return
		}

		var models []*model.ExportedModel
		if terms != "" {
			models, err = storage.SearchModels(terms)
		} else {
			models, err = storage.FetchModels()
		}
		if err != nil {
			handleError(w, err)
			return
		}

		// marshal data
		err = handleJSON(w, models)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to marshal model results into JSON"))
			return
		}
	}
}
