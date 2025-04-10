package models

import (
	"encoding/json"
	"time"
)

type DbMessageData struct {
	Topic     string          `json:"topic"`
	Message   json.RawMessage `json:"message"`
	TimeStamp time.Time       `json:"time"`
}
