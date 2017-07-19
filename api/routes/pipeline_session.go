package routes

import (
	"net/http"

	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil/api/pipeline"
	context "golang.org/x/net/context"
)

// PipelineSessionHandler creates a route to handle pipeline create requests.
// ** This is just an example impl for reference and should be replaced.
func PipelineSessionHandler(pipelineService *pipeline.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		requestInfo := pipeline.GenerateStartSessionRequest()

		// Just an example - a session requests hash will always be different so the case where we can
		// actually attach to a running request won't ever get hit
		var results *pipeline.ResultProxy
		var bufferedResults []interface{}
		if attachID, ok := pipelineService.IsRequestAttachable(requestInfo); ok {
			// there was a running request we could attach to, so we'll re-use that
			results, bufferedResults = pipelineService.Attach(attachID)
		} else {
			// no request we could re-use, dispatch a new one and attach
			requestID := pipelineService.Dispatch(context.Background(), requestInfo.RequestFunc)
			results, bufferedResults = pipelineService.Attach(requestID)
		}

		if len(bufferedResults) > 0 {
			// A session requests is single call/response, so if we already have a result, process it and
			// we're done.
			response := bufferedResults[0].(*pipeline.Response)
			w.Write([]byte(response.GetContext().GetSessionId()))
		} else {
			for {
				select {
				case r := <-results.Results:
					response := r.(*pipeline.Response)
					w.Write([]byte(response.GetContext().GetSessionId()))
				case err := <-results.Errors:
					handleError(w, errors.Wrap(err, "failed to fetch session ID"))
				case <-results.Done:
					return
				}
			}
		}
	}
}
