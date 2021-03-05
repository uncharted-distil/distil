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

package middleware

import (
	"fmt"
	"net/http"
	"time"
)

// LoggingClient is an http.Client that logs *outoing* REST requests.
type LoggingClient struct {
	http.Client
}

// Do wraps the basic http.Client.Do call to log requests and responses.
func (c *LoggingClient) Do(req *http.Request) (*http.Response, error) {
	// execute the request
	t1 := time.Now()
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	t2 := time.Now()

	newRequestLogger().
		requestType(fmt.Sprintf("REST CLIENT %s", req.Method)).
		request(req.URL.String()).
		params(req.URL.String()).
		status(resp.StatusCode).
		duration(t2.Sub(t1)).
		log(resp.StatusCode < 500)

	return resp, err
}
