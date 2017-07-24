package routes

import (
	"net/http"

	"goji.io/pat"

	"github.com/pkg/errors"
	"golang.org/x/net/context"

	"github.com/unchartedsoftware/distil/api/pipeline"
)

// PipelineEndSessionHandler creates a route to handle pipeline session end requests.
// ** This is just an example impl for reference and should be replaced.
func PipelineEndSessionHandler(pipelineService *pipeline.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		sessionID := pat.Param(r, "session-id")

		// generate the request
		req := pipeline.GenerateEndSessionRequest(sessionID)

		// gets an existing request or dispatchs a new one
		proxy, err := pipelineService.GetOrDispatch(context.Background(), req)
		if err != nil {
			handleError(w, errors.Wrap(err, "failed to end session"))
		}

		// process the result proxy, which is replicated for completed, pending requests
		for {
			select {
			case result := <-proxy.Results:
				res := (*result).(*pipeline.Response)
				w.Write([]byte(res.String()))
			case err := <-proxy.Errors:
				// handle error
				handleError(w, errors.Wrap(err, "failed to end session"))
			case <-proxy.Done:
				// finished
				return
			}
		}
	}
}
