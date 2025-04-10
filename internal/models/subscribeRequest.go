package models

import "time"

type SubscribeRequest struct {
	Topic string    `json:"topic"`
	Time  time.Time `json:"time"`
}
