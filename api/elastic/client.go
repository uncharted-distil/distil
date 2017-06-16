package elastic

import (
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/unchartedsoftware/plog"
	"gopkg.in/olivere/elastic.v3"
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
	// NOTE: have to break this into two cases, since declaring an adapter
	// and leaving it nil will cause a panic as typed interfaces fail equality
	// comparisons with `nil`
	if debug {
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
	log.Infof("Connected to endpoint %s", endpoint)
	return client, nil
}
