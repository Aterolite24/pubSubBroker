package models

import (
	"encoding/json"
	"sync"
)

type Broker struct {
	mu   sync.RWMutex
	subs map[string][]chan json.RawMessage
}

func NewBroker() *Broker {
	return &Broker{
		subs: make(map[string][]chan json.RawMessage),
	}
}

func (b *Broker) Publish(topic string, msg json.RawMessage) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, ch := range b.subs[topic] {
		select {
		case ch <- msg:
		default:
			// Optionally, log or handle the case where the channel is full.
		}
	}
}

func (b *Broker) Subscribe(topic string) chan json.RawMessage {
	ch := make(chan json.RawMessage, 10)
	b.mu.Lock()
	b.subs[topic] = append(b.subs[topic], ch)
	b.mu.Unlock()
	return ch
}

func (b *Broker) Unsubscribe(topic string, ch chan json.RawMessage) {
	b.mu.Lock()
	defer b.mu.Unlock()
	subs := b.subs[topic]
	for i, sub := range subs {
		if sub == ch {
			// Remove the subscriber channel from the slice.
			b.subs[topic] = append(subs[:i], subs[i+1:]...)
			close(ch)
			break
		}
	}
}
