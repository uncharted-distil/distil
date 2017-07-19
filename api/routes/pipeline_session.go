package routes

import (
	"net/http"

	"github.com/pkg/errors"
	"golang.org/x/net/context"

	"github.com/unchartedsoftware/distil/api/pipeline"
)

// PipelineSessionHandler creates a route to handle pipeline create requests.
// ** This is just an example impl for reference and should be replaced.
func PipelineSessionHandler(pipelineService *pipeline.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		// generate the request
		req := pipeline.GenerateStartSessionRequest()

		// gets an existing request or dispatchs a new one
		proxy, err := pipelineService.GetOrDispatch(context.Background(), req)
		if err != nil {
			handleError(w, errors.Wrap(err, "failed to fetch session ID"))
		}

		// process the result proxy, which is replicated for completed, pending requests
		for {
			select {
			case result := <-proxy.Results:
				// A session requests is single call/response, so if we already
				// have a result, process it and we're done.
				res := result.(*pipeline.Response)
				w.Write([]byte(res.GetContext().GetSessionId()))

			case err := <-proxy.Errors:
				// handle error
				handleError(w, errors.Wrap(err, "failed to fetch session ID"))
				return

			case <-proxy.Done:
				// finished
				return
			}
		}
	}
}
