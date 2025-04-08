package models

import "encoding/json"

type SubscribeData struct {
	Topic   string          `json:"topic"`
	Message json.RawMessage `json:"message"`
}
