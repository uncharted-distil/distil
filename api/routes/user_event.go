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
	"net/http"

	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil/api/env"
	"github.com/uncharted-distil/distil/api/util/json"
)

// UserEventHandler logs UI events to the discovery logger
func UserEventHandler(logger *env.DiscoveryLogger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// parse the request
		params, err := getPostParameters(r)
		if err != nil {
			handleError(w, errors.Wrap(err, "Unable to parse post parameters"))
			return
		}
		feature, ok := json.String(params, "feature")
		if !ok {
			handleError(w, errors.Wrap(err, "Unable to parse `feature` parameter"))
			return
		}
		activity, ok := json.String(params, "activity")
		if !ok {
			handleError(w, errors.Wrap(err, "Unable to parse `activity` parameter"))
			return
		}

		logger.LogSystemAction(feature, activity, "")
	}
}
