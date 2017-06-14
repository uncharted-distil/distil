package env

import (
	"os"

	"github.com/unchartedsoftware/plog"
)

var (
	loaded map[string]string
)

func init() {
	loaded = make(map[string]string)
}

// Load will read a value from an environment variable defaulting to the
// provided fallback if necessary.
func Load(key, fallback string) string {
	// check if already loaded
	val, ok := loaded[key]
	if ok {
		return val
	}
	// load from env var
	val = os.Getenv(key)
	if len(val) == 0 {
		val = fallback
		log.Infof("%s = %s, defaulted", key, val)
	} else {
		log.Infof("%s = %s, from env var", key, val)
	}
	loaded[key] = val
	return val
}
