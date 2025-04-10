package models

import (
	"time"
)

type ClientData struct {
	TimeStamp time.Time `json:"time"`
	Topic     string    `json:"topic"`
}


