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

	log "github.com/unchartedsoftware/plog"
)

var (
	verboseError = false
)

// SetVerboseError sets the flag determining if the client should receive
// error details
func SetVerboseError(verbose bool) {
	verboseError = verbose
}

func handleError(w http.ResponseWriter, err error) {
	handleErrorType(w, err, http.StatusInternalServerError)
}

func handleErrorType(w http.ResponseWriter, err error, code int) {
	log.Errorf("%+v", err)
	errMessage := "An error occured on the server while processing the request"
	if verboseError {
		errMessage = err.Error()
	}
	http.Error(w, errMessage, code)
}
