package middleware

import (
	"net/http"
	"net/http/httputil"

	log "github.com/unchartedsoftware/plog"
)

// LoggingClient is an http.Client that logs *outoing* REST requests.
type LoggingClient struct {
	http.Client
}

// Do wraps the basic http.Client.Do call to log requests and responses.
func (c *LoggingClient) Do(req *http.Request) (*http.Response, error) {

	reqDump, _ := httputil.DumpRequestOut(req, true)
	log.Infof("%s", reqDump)

	// execute the request
	// t1 := time.Now()
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	// t2 := time.Now()

	respDump, _ := httputil.DumpResponse(resp, false)
	log.Infof("%s", respDump)

	// newRequestLogger().
	// 	requestType(fmt.Sprintf("REST CLIENT %s", req.Method)).
	// 	request(req.URL.String()).
	// 	params(req.URL.String()).
	// 	status(resp.StatusCode).
	// 	duration(t2.Sub(t1)).
	// log(resp.StatusCode < 500)

	return resp, err
}
