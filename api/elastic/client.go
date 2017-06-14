package elastic

import (
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/unchartedsoftware/plog"
	"gopkg.in/olivere/elastic.v2"
)

const (
	defaultTimeout = time.Second * 30
)

// Wraps calls to plog in the elastic.Logger interface
type elasticPlogAdapter struct{}

func (elasticPlogAdapter) Printf(format string, v ...interface{}) {
	log.Infof(format, v)
}

// NewClient instantiates and returns a new elasticsearch client.
func NewClient(endpoint string, debug bool) (*elastic.Client, error) {
	// turn on trace logs if necessary
	var client *elastic.Client
	var err error
	if debug {
		adapter := &elasticPlogAdapter{}
		client, err = elastic.NewClient(
			elastic.SetURL(endpoint),
			elastic.SetHttpClient(&http.Client{Timeout: defaultTimeout}),
			elastic.SetMaxRetries(10),
			elastic.SetSniff(false),
			elastic.SetGzip(false),
			elastic.SetTraceLog(adapter))
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
	log.Infof("Connected to endpoint %s", endpoint)
	return client, nil
}
