package common

import "encoding/json"

type Message struct {
	Type    string          `json:"type"`
	Content string          `json:"content"`
	RoomID  string          `json:"room,omitempty"`
	Sender  string          `json:"sender,omitempty"`
	Data    json.RawMessage `json:"data,omitempty"`
}
