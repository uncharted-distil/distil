package compute

import (
	"encoding/base64"
	"time"

	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil/api/util/json"
)

// PredictRequest defines a request to generate new predictions from a fitted model and input data.
type PredictRequest struct {
	Dataset          string
	TargetType       string
	FittedSolutionID string
	TimestampField   string
	MaxTime          int

	requestChannel chan PredictStatus
	listener       PredictStatusListener
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

// NewPredictRequest instantiates a predict request from a raw byte stream, as would
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
	req.Dataset, ok = json.String(jsonMap, "dataset")
	if !ok {
		return nil, errors.Errorf("no `dataset` in predict request")
	}

	// the
	req.TargetType, ok = json.String(jsonMap, "targetType")
	if !ok {
		return nil, errors.Errorf("no `target` in predict request")
	}

	return req, nil
}

// PredictStatusListener executes whenever prediction status is returned by the downstream autoML system.
type PredictStatusListener func(status PredictStatus)

// ExtractDatasetEncodedFromRawRequest extracts the dataset name from the raw message.
func ExtractDatasetEncodedFromRawRequest(data []byte) (string, error) {
	jsonMap, err := json.Unmarshal(data)
	if err != nil {
		return "", err
	}

	var ok bool

	encoded, ok := json.String(jsonMap, "dataset")
	if !ok {
		return "", errors.New("no `dataset` in predict request")
	}

	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", errors.Wrap(err, "could not decoded `dataset`")
	}

	return string(decoded), nil
}
