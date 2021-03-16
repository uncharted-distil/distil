package util

import (
	"runtime"
	"time"

	log "github.com/unchartedsoftware/plog"
)

// StartMemLogging starts logging memory usage at the caller specified interval
func StartMemLogging(intervalMs int) {
	go func() {
		for {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			log.Infof("\nAlloc = %v MB\nTotalAlloc = %v MB\nSys = %v MB\nNumGC = %v\n\n", m.Alloc/1024/1024, m.TotalAlloc/1024/1024, m.Sys/1024/1024, m.NumGC)
			time.Sleep(time.Duration(intervalMs) * time.Millisecond)
		}
	}()
}
