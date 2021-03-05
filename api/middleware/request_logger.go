//
//   Copyright © 2021 Uncharted Software Inc.
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

package middleware

import (
	"bytes"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/golang/protobuf/proto" //nolint need to update to new protobuf api
	"github.com/mattn/go-isatty"
	"github.com/mgutz/ansi"
	log "github.com/unchartedsoftware/plog"
	"github.com/vova616/xxhash"
)

type requestLogger struct {
	buf      *bytes.Buffer
	colorTTY bool
}

func newRequestLogger() *requestLogger {
	return &requestLogger{
		buf:      &bytes.Buffer{},
		colorTTY: isatty.IsTerminal(os.Stdout.Fd()) && (runtime.GOOS != "windows"),
	}
}

func (r *requestLogger) write(color string, format string, args ...interface{}) {
	if r.colorTTY {
		fmt.Fprint(r.buf, color)
	}
	fmt.Fprintf(r.buf, format, args...)
	if r.colorTTY {
		fmt.Fprint(r.buf, ansi.Reset)
	}
}

func (r *requestLogger) requestType(reqType string) *requestLogger {
	r.write(ansi.Magenta, "%s ", reqType)
	return r
}

func (r *requestLogger) request(request string) *requestLogger {
	urlsplit := strings.Split(request, "?")
	url := urlsplit[0]
	cs := strings.Split(url, "/")
	// write out base URL
	if len(cs) == 2 && cs[0] == "" && cs[1] == "" {
		r.write(ansi.DefaultFG, "/")
	} else {
		for _, c := range cs {
			if c != "" {
				r.write(ansi.DefaultFG, "/")
				r.write(ansi.Blue, c)
			}
		}
	}
	return r
}

func (r *requestLogger) message(request proto.Message) *requestLogger {
	protoString := proto.MarshalTextString(request)
	r.write(ansi.Green, "\n"+protoString)
	return r
}

func (r *requestLogger) params(request string) *requestLogger {
	urlsplit := strings.Split(request, "?")
	if len(urlsplit) > 1 {
		// hash query params
		r.write(ansi.DefaultFG, "?")
		hash := xxhash.Checksum32([]byte(urlsplit[1]))
		r.write(ansi.Green, "%#x ", hash)
	} else {
		r.buf.WriteString(" ")
	}
	return r
}

func (r *requestLogger) status(status int) *requestLogger {
	if status < 200 {
		r.write(ansi.Blue, "%03d", status)
	} else if status < 300 {
		r.write(ansi.Green, "%03d", status)
	} else if status < 400 {
		r.write(ansi.Cyan, "%03d", status)
	} else if status < 500 {
		r.write(ansi.Yellow, "%03d", status)
	} else {
		r.write(ansi.Red, "%03d", status)
	}
	return r
}

func (r *requestLogger) duration(duration time.Duration) *requestLogger {
	r.buf.WriteString(" in ")
	if duration < 200*time.Millisecond {
		r.write(ansi.Blue, "%.2fms", duration.Seconds()*1000)
	} else if duration < 500*time.Millisecond {
		r.write(ansi.Green, "%.2fms", duration.Seconds()*1000)
	} else if duration < 2*time.Second {
		r.write(ansi.Yellow, "%.2fms", duration.Seconds()*1000)
	} else {
		r.write(ansi.Red, "%.2fms", duration.Seconds()*1000)
	}
	return r
}

func (r *requestLogger) log(success bool) {
	if success {
		log.Info(r.buf.String())
	} else {
		log.Warn(r.buf.String())
	}
}
