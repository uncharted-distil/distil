package routes

import (
	"net/http"

	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil/api/pipeline"
	context "golang.org/x/net/context"
)

// PipelineSessionHandler creates a route to handle pipeline create requests.
func PipelineSessionHandler(pipelineService *pipeline.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := pipelineService.Dispatch(context.Background(), pipeline.GenerateStartSessionRequest())
		resultChannels, bufferedResults := pipelineService.Attach(requestID)

		if len(bufferedResults) > 0 {
			response := bufferedResults[0].(*pipeline.Response)
			w.Write([]byte(response.GetContext().GetSessionId()))
		} else {
			select {
			case results := <-resultChannels.Results:
				response := results.(*pipeline.Response)
				w.Write([]byte(response.GetContext().GetSessionId()))
			case err := <-resultChannels.Errors:
				handleError(w, errors.Wrap(err, "failed to fetch session ID"))
			}
		}
	}
}
