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
	"net/http"
	"sync"
	"time"

	"github.com/olivere/elastic/v7"
	"github.com/pkg/errors"
	"github.com/unchartedsoftware/plog"
)

const (
	defaultTimeout = time.Second * 30
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
					elastic.SetHttpClient(&http.Client{Timeout: defaultTimeout}),
					elastic.SetMaxRetries(10),
					elastic.SetSniff(false),
					elastic.SetGzip(false),
					elastic.SetTraceLog(&elasticPlogAdapter{}))
			} else {
				client, err = elastic.NewClient(
					elastic.SetURL(endpoint),
					elastic.SetHttpClient(&http.Client{Timeout: defaultTimeout}),
					elastic.SetMaxRetries(10),
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
