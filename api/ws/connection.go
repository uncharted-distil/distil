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
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

const (
	writeWait      = 10 * time.Second
	maxMessageSize = 256 * 256
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  maxMessageSize,
	WriteBufferSize: maxMessageSize,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// requestHandler represents a handler for the ws request.
type requestHandler func(*Connection, []byte)

// Connection represents a single clients tile dispatcher.
type Connection struct {
	conn    *websocket.Conn
	mu      *sync.Mutex
	handler requestHandler
}

// NewConnection returns a pointer to a new tile dispatcher object.
func NewConnection(w http.ResponseWriter, r *http.Request, handler requestHandler) (*Connection, error) {
	// open a websocket connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}
	return &Connection{
		conn:    conn,
		handler: handler,
		mu:      &sync.Mutex{},
	}, nil
}

// ListenAndRespond waits on both tile request and responses and handles each
// until the websocket connection dies.
func (c *Connection) ListenAndRespond() error {
	for {
		// wait on read
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			return err
		}
		// handle the message
		go c.handler(c, msg)
	}
}

// SendResponse will send a json response in a thread safe manner.
func (c *Connection) SendResponse(res interface{}) error {
	// writes are not thread safe
	c.mu.Lock()
	defer c.mu.Unlock()
	// write response to websocket
	err := c.conn.SetWriteDeadline(time.Now().Add(writeWait))
	if err != nil {
		return errors.Wrap(err, "failed to set socket write deadline")
	}
	return c.conn.WriteJSON(res)
}

// Close closes the dispatchers websocket connection.
func (c *Connection) Close() {
	// ensure we aren't closing during a write
	c.mu.Lock()
	defer c.mu.Unlock()
	// close websocket connection
	c.conn.Close()
}
