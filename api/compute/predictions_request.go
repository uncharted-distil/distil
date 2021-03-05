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
	"github.com/uncharted-distil/distil/api/util/json"
)

// PredictRequest defines a request to generate new predictions from a fitted model and input data.
type PredictRequest struct {
	DatasetID        string
	DatasetPath      string
	FittedSolutionID string
	TimestampField   string
	MaxTime          int
	IntervalCount    int
	IntervalLength   float64
	ExistingDataset  bool

	requestChannel chan PredictStatus
	finished       chan error
}

// PredictStatus defines a prediction status update from a downstream autoML system
type PredictStatus struct {
	Progress  string    `json:"progress"`
	RequestID string    `json:"requestId"`
	ResultID  string    `json:"resultId"`
	Error     error     `json:"error"`
	Timestamp time.Time `json:"timestamp"`
}

// NewPredictRequest instantiates a predict request from a raw byte stream.
func NewPredictRequest(data []byte) (*PredictRequest, error) {
	req := &PredictRequest{
		finished:       make(chan error),
		requestChannel: make(chan PredictStatus),
	}

	jsonMap, err := json.Unmarshal(data)
	if err != nil {
		return nil, err
	}

	var ok bool

	// the fitted pipeline to use for the predictions
	req.FittedSolutionID, ok = json.String(jsonMap, "fittedSolutionId")
	if !ok {
		return nil, errors.Errorf("no `fittedSolutionId` in predict request")
	}

	// the name of the input prediction dataset
	req.DatasetID, ok = json.String(jsonMap, "datasetId")
	if !ok {
		return nil, errors.Errorf("no `datasetId` in predict request")
	}

	// the dataset contents as a base 64 encded string
	req.DatasetPath, ok = json.String(jsonMap, "datasetPath")
	if !ok {
		req.DatasetPath = ""
	}

	// timeseries prediction fields
	req.IntervalCount, ok = json.Int(jsonMap, "intervalCount")
	if !ok {
		req.IntervalCount = 0
	}

	req.IntervalLength, ok = json.Float(jsonMap, "intervalLength")
	if !ok {
		req.IntervalLength = 0
	}

	req.ExistingDataset, ok = json.Bool(jsonMap, "existingDataset")
	if !ok {
		req.ExistingDataset = false
	}

	return req, nil
}
