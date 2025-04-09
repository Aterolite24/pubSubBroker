package models

import (
	"encoding/json"
	"time"
)

type pendingMessage struct {
	topic      string
	envelope   Envelope
	subscriber chan json.RawMessage
	tries      int
	timer      *time.Timer
}