package common

import "encoding/json"

type Message struct {
	Type    string          `json:"type"`
	Content string          `json:"content"`
	Data    json.RawMessage `json:data,omitempty`
}
