package compute

import (
	"context"
	"encoding/json"
)

// StopSolutionSearchRequest represents a request to stop any pending siolution searches.
type StopSolutionSearchRequest struct {
	RequestID string `json:"requestId"`
}

// NewStopSolutionSearchRequest instantiates a new StopSolutionSearchRequest.
func NewStopSolutionSearchRequest(data []byte) (*StopSolutionSearchRequest, error) {
	req := &StopSolutionSearchRequest{}
	err := json.Unmarshal(data, &req)
	if err != nil {
		return nil, err
	}
	return req, nil
}

// Dispatch dispatches the stop search request.
func (s *StopSolutionSearchRequest) Dispatch(client *Client) error {
	return client.StopSearch(context.Background(), s.RequestID)
}
