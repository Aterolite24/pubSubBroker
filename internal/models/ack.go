package models

import "time"

type Ack struct {
	MessageID string    `json:"messageId"`
	Topic     string    `json:"topic"`
	Timestamp time.Time `json:"timestamp"`
}
