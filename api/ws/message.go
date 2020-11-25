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

package ws

import (
	"encoding/json"
	"time"
)

// Message represents a websocket message.
type Message struct {
	Type      string          `json:"type"`
	ID        string          `json:"id"`
	Timestamp time.Time       `json:"-"`
	Body      json.RawMessage `json:"body"`
}

// NewMessage parses and instantiates a new message struct.
func NewMessage(bytes []byte) (*Message, error) {
	var msg *Message
	err := json.Unmarshal(bytes, &msg)
	if err != nil {
		return nil, err
	}
	msg.Timestamp = time.Now()
	return msg, nil
}
