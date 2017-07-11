package middleware

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/mattn/go-isatty"
	"github.com/mgutz/ansi"
	"github.com/unchartedsoftware/plog"
	"github.com/vova616/xxhash"
)

var (
	colorTTY = isatty.IsTerminal(os.Stdout.Fd()) && (runtime.GOOS != "windows")
)

// Log is a middleware that logs each request. Inspired from the logging
// middleware from zenazn/goji:
// https://github.com/zenazn/goji/blob/master/web/middleware/logger.go
func Log(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if isWebsocketUpgrade(r) {
			// do not log websocket connections
			// TODO: intercept and log the beginning and the end of the
			// connection.
			h.ServeHTTP(w, r)
			return
		}
		lw := wrapWriter(w)
		t1 := time.Now()
		h.ServeHTTP(lw, r)
		if lw.Status() == 0 {
			lw.WriteHeader(http.StatusOK)
		}
		t2 := time.Now()
		logResponse(r, lw, t2.Sub(t1))
	}
	return http.HandlerFunc(fn)
}

func write(b *bytes.Buffer, color string, format string, args ...interface{}) {
	if colorTTY {
		fmt.Fprintf(b, color)
	}
	fmt.Fprintf(b, format, args...)
	if colorTTY {
		fmt.Fprintf(b, ansi.Reset)
	}
}

func logResponse(r *http.Request, w writerProxy, dt time.Duration) {
	var buf bytes.Buffer
	// write HTTP method
	write(&buf, ansi.Magenta, "%s ", r.Method)
	// split URL based on query params
	urlsplit := strings.Split(r.URL.String(), "?")
	url := urlsplit[0]
	cs := strings.Split(url, "/")
	// write out base URL
	if len(cs) == 2 && cs[0] == "" && cs[1] == "" {
		write(&buf, ansi.Black, "/")
	} else {
		for _, c := range cs {
			if c != "" {
				write(&buf, ansi.Black, "/")
				write(&buf, ansi.Blue, c)
			}
		}
	}
	// if query params, hash them and write it out
	if len(urlsplit) > 1 {
		// hash query params
		write(&buf, ansi.Black, "?")
		hash := xxhash.Checksum32([]byte(urlsplit[1]))
		write(&buf, ansi.Green, "%#x ", hash)
	} else {
		buf.WriteString(" ")
	}
	// write out response status
	status := w.Status()
	if status < 200 {
		write(&buf, ansi.Blue, "%03d", status)
	} else if status < 300 {
		write(&buf, ansi.Green, "%03d", status)
	} else if status < 400 {
		write(&buf, ansi.Cyan, "%03d", status)
	} else if status < 500 {
		write(&buf, ansi.Yellow, "%03d", status)
	} else {
		write(&buf, ansi.Red, "%03d", status)
	}
	// write out duration in milliseconds
	buf.WriteString(" in ")
	if dt < 200*time.Millisecond {
		write(&buf, ansi.Blue, "%.2fms", dt.Seconds()*1000)
	} else if dt < 500*time.Millisecond {
		write(&buf, ansi.Green, "%.2fms", dt.Seconds()*1000)
	} else if dt < 2*time.Second {
		write(&buf, ansi.Yellow, "%.2fms", dt.Seconds()*1000)
	} else {
		write(&buf, ansi.Red, "%.2fms", dt.Seconds()*1000)
	}
	// log the output according to status code
	if status < 500 {
		log.Info(buf.String())
	} else {
		log.Warn(buf.String())
	}
}
