package models

import "encoding/json"

type Envelope struct {
	ID      string          `json:"id"`
	Topic   string          `json:"topic"`
	Payload json.RawMessage `json:"payload"`
}
