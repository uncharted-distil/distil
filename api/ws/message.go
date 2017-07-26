package ws

import (
	"encoding/json"
	"time"
)

// Message represents a websocket message.
type Message struct {
	Type      string    `json:"type"`
	ID        string    `json:"id"`
	Session   string    `json:"session"`
	Timestamp time.Time `json:"-"`
	Raw       []byte    `json:"-"`
}

// NewMessage parses and instantiates a new message struct.
func NewMessage(bytes []byte) (*Message, error) {
	var msg *Message
	err := json.Unmarshal(bytes, &msg)
	if err != nil {
		return nil, err
	}
	msg.Timestamp = time.Now()
	msg.Raw = bytes
	return msg, nil
}
