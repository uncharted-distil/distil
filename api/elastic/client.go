//
//   Copyright © 2019 Uncharted Software Inc.
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

package elastic

import (
	"context"
	"net/http"
	"sync"
	"syscall"
	"time"

	"github.com/olivere/elastic/v7"
	"github.com/pkg/errors"
	log "github.com/unchartedsoftware/plog"
)

const (
	defaultHTTPTimeoutSec = 30
	defaultIntervalMs     = 5000
	defaultRetries        = 50
)

var (
	mu      = &sync.Mutex{}
	clients map[string]*elastic.Client
)

func init() {
	clients = make(map[string]*elastic.Client)
}

// ClientCtor repressents a client constructor to instantiate an elasticsearch
// client.
type ClientCtor func() (*elastic.Client, error)

// Retrier defined a constant backoff retry strategy for the elastic connection.
type Retrier struct {
	backoff elastic.ConstantBackoff
	retries int
}

// NewRetrier creates a the elastic client connection retries.  It will attempt to connect
// every intervalMs milliseconds up to a maximum of retries attempts.
func NewRetrier(retries int, intervalMs time.Duration) *Retrier {
	return &Retrier{
		backoff: *elastic.NewConstantBackoff(intervalMs * time.Millisecond),
		retries: retries,
	}
}

// Retry is called when an attempted connection to the elastic client has failed.
func (r *Retrier) Retry(ctx context.Context, retry int, req *http.Request, resp *http.Response, err error) (time.Duration, bool, error) {
	// Fail hard on a specific error
	if err == syscall.ECONNREFUSED {
		return 0, false, errors.New("elasticsearch or network down")
	}

	// Stop after allowed number of retries surpassed
	if retry >= r.retries {
		return 0, false, nil
	}

	// Let the backoff strategy decide how long to wait and whether to stop
	wait, stop := r.backoff.Next(retry)
	return wait, stop, nil
}

// NewClient instantiates and returns a new elasticsearch client constructor.
func NewClient(endpoint string, debug bool) ClientCtor {
	return func() (*elastic.Client, error) {
		mu.Lock()
		defer mu.Unlock()

		// see if we have an existing connection
		client, ok := clients[endpoint]
		if !ok {
			var err error
			// NOTE: have to break this into two cases, since declaring an adapter
			// and leaving it nil will cause a panic as typed interfaces fail equality
			// comparisons with `nil`
			if debug {
				// turn on trace logs if necessary
				client, err = elastic.NewClient(
					elastic.SetURL(endpoint),
					elastic.SetHttpClient(&http.Client{Timeout: defaultHTTPTimeoutSec * time.Second}),
					elastic.SetRetrier(NewRetrier(defaultRetries, defaultIntervalMs*time.Millisecond)),
					elastic.SetSniff(false),
					elastic.SetGzip(false),
					elastic.SetTraceLog(&elasticPlogAdapter{}))
			} else {
				client, err = elastic.NewClient(
					elastic.SetURL(endpoint),
					elastic.SetHttpClient(&http.Client{Timeout: defaultHTTPTimeoutSec * time.Second}),
					elastic.SetRetrier(NewRetrier(defaultRetries, defaultIntervalMs*time.Millisecond)),
					elastic.SetSniff(false),
					elastic.SetGzip(false))
			}
			if err != nil {
				return nil, errors.Wrap(err, "ES client init failed")
			}
			log.Infof("Elasticsearch connection established to endpoint %s", endpoint)
			clients[endpoint] = client
		}
		return client, nil
	}
}
