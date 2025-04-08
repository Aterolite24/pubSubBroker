package models

import "encoding/json"

type PublishData struct {
	Topic   string          `json:"topic"`
	Message json.RawMessage `json:"message"`
}
