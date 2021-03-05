//
//    Copyright © 2021 Uncharted Software Inc.
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package compute

import (
	"time"

	"github.com/pkg/errors"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/util/json"
)

// QueryRequest defines a request to query a dataset for similar images to labelled observations.
type QueryRequest struct {
	DatasetID string
	Target    string
	Filters   *api.FilterParams

	requestChannel chan QueryStatus
	finished       chan error
}

// QueryStatus defines a query status update from a downstream autoML system
type QueryStatus struct {
	Progress  string    `json:"progress"`
	RequestID string    `json:"requestId"`
	ResultID  string    `json:"resultId"`
	Error     error     `json:"error"`
	Timestamp time.Time `json:"timestamp"`
}

// NewQueryRequest instantiates a query request from a raw byte stream.
func NewQueryRequest(data []byte) (*QueryRequest, error) {
	req := &QueryRequest{
		finished:       make(chan error),
		requestChannel: make(chan QueryStatus),
	}

	jsonMap, err := json.Unmarshal(data)
	if err != nil {
		return nil, err
	}

	var ok bool

	// the name of the input prediction dataset
	req.DatasetID, ok = json.String(jsonMap, "datasetId")
	if !ok {
		return nil, errors.Errorf("no `datasetId` in predict request")
	}

	// the target is the name of the label on which to base the query.
	req.Target = json.StringDefault(jsonMap, "", "target")

	filters, ok := json.Get(jsonMap, "filters")
	if ok {
		req.Filters, err = api.ParseFilterParamsFromJSON(filters)
		if err != nil {
			return nil, err
		}
	}

	return req, nil
}
