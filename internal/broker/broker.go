package broker

import (
	"log"
	"pubSubBroker/internal/connection"
)

type Broker struct {
	connManager *connection.Manager
}

func NewBroker() *Broker {
	return &Broker{
		connManager: connection.NewManager(),
	}
}

func (b *Broker) Start() {
	log.Println("Starting Connection Manager...")
	b.connManager.Start()
}
