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
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil/api/env"
)

type event struct {
	Feature     string           `json:"feature"`
	Activity    string           `json:"activity"`
	SubActivity string           `json:"subActivity"`
	Details     *json.RawMessage `json:"details"`
}

// UserEventHandler logs UI events to the discovery logger
func UserEventHandler(logger *env.DiscoveryLogger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// parse the request
		params, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			handleError(w, errors.Wrap(err, "Unable to parse POST request"))
			return
		}

		var evt event
		if err = json.Unmarshal([]byte(params), &evt); err != nil {
			handleError(w, errors.Wrap(err, "Unable to unmarshal post parameters"))
			return
		}

		logger.LogSystemAction(evt.Feature, evt.Activity, evt.SubActivity, string(*evt.Details))
	}
}
