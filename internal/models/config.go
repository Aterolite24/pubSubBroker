package models

import (
	"time"
)

type Config struct {
	AtLeastOnce        bool
	AtMostOnce         bool
	AckTimeout         time.Duration
	DeletionInterval   time.Duration
	PersistenceTimeout time.Duration
}

func NewConfig(atLeastOnce bool, atMostOnce bool, ackTimeout time.Duration, deletionInterval time.Duration, persistenceTimeout time.Duration) Config {
	return Config{
		AtLeastOnce:        atLeastOnce,
		AtMostOnce:         atMostOnce,
		AckTimeout:         ackTimeout,
		DeletionInterval:   deletionInterval,
		PersistenceTimeout: persistenceTimeout,
	}
}
