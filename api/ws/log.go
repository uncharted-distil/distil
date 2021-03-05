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

package ws

import (
	"bytes"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/mattn/go-isatty"
	"github.com/mgutz/ansi"
	log "github.com/unchartedsoftware/plog"
	"github.com/vova616/xxhash"
)

type messageLogger struct {
	buf      *bytes.Buffer
	colorTTY bool
}

func newMessageLogger() *messageLogger {
	return &messageLogger{
		buf:      &bytes.Buffer{},
		colorTTY: isatty.IsTerminal(os.Stdout.Fd()) && (runtime.GOOS != "windows"),
	}
}

func (l *messageLogger) write(color string, format string, args ...interface{}) {
	if l.colorTTY {
		fmt.Fprint(l.buf, color)
	}
	fmt.Fprintf(l.buf, format, args...)
	if l.colorTTY {
		fmt.Fprintf(l.buf, ansi.Reset)
	}
}

func (l *messageLogger) messageType(msgType string) *messageLogger {
	l.write(ansi.Magenta, "WS %s ", msgType)
	return l
}

func (l *messageLogger) message(msg []byte) *messageLogger {
	hash := xxhash.Checksum32(msg)
	l.write(ansi.Green, "%#x ", hash)
	return l
}

func (l *messageLogger) duration(duration time.Duration) *messageLogger {
	l.buf.WriteString(" in ")
	if duration < 200*time.Millisecond {
		l.write(ansi.Blue, "%.2fms", duration.Seconds()*1000)
	} else if duration < 500*time.Millisecond {
		l.write(ansi.Green, "%.2fms", duration.Seconds()*1000)
	} else if duration < 2*time.Second {
		l.write(ansi.Yellow, "%.2fms", duration.Seconds()*1000)
	} else {
		l.write(ansi.Red, "%.2fms", duration.Seconds()*1000)
	}
	return l
}

func (l *messageLogger) log(success bool) {
	if success {
		log.Info(l.buf.String())
	} else {
		log.Warn(l.buf.String())
	}
}
