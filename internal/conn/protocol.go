package conn

import (
	"encoding/json"
	"fmt"
)

// Message represents a structured protocol message.
type Message struct {
	Action  string `json:"action"`  // e.g. "SUB", "PUB", "ACK"
	Topic   string `json:"topic"`   // e.g. "sports", "updates"
	Payload string `json:"payload"` // e.g. message content
}

// ParseMessage tries to parse raw string into Message struct
func ParseMessage(raw string) (*Message, error) {
	var msg Message
	err := json.Unmarshal([]byte(raw), &msg)
	if err != nil {
		return nil, fmt.Errorf("invalid message format: %w", err)
	}
	if msg.Action == "" {
		return nil, fmt.Errorf("missing action field")
	}
	return &msg, nil
}

// EncodeMessage encodes a Message struct to a JSON string.
func EncodeMessage(msg *Message) (string, error) {
	bytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
