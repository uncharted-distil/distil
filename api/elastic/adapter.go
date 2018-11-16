package elastic

import (
	"github.com/unchartedsoftware/plog"
)

// Wraps calls to plog in the elastic.Logger interface
type elasticPlogAdapter struct{}

func (elasticPlogAdapter) Printf(format string, v ...interface{}) {
	log.Infof(format, v...)
}
