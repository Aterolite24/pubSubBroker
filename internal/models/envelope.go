package models

import "encoding/json"

type Envelope struct {
	ID      string          `json:"id"`
	Payload json.RawMessage `json:"payload"`
}
