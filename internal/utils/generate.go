package utils

import (
	"github.com/google/uuid"
)

func GenerateMessageID() string {
	return uuid.New().String()
}
