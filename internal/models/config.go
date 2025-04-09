package models

import (
	"time"
)

type Config struct {
	AtLeastOnce bool
	AtMostOnce  bool
	AckTimeout  time.Duration
}

func NewConfig(atLeastOnce bool, atMostOnce bool, ackTimeout time.Duration) Config {
	return Config{
		AtLeastOnce: atLeastOnce,
		AtMostOnce:  atMostOnce,
		AckTimeout:  ackTimeout,
	}
}
