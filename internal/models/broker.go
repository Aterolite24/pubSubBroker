package models

import (
	"encoding/json"
	"sync"
	"time"

	"pubSubBroker/internal/utils"
)

type Broker struct {
	mu          sync.RWMutex
	subs        map[string][]chan json.RawMessage
	config      Config
	pendingAcks map[string]pendingMessage
}

func NewBroker(config Config) *Broker {
	return &Broker{
		subs:        make(map[string][]chan json.RawMessage),
		config:      config,
		pendingAcks: make(map[string]pendingMessage),
	}
}

func (b *Broker) Publish(topic string, msg json.RawMessage) error {
	b.mu.RLock()
	defer b.mu.RUnlock()
	subs, ok := b.subs[topic]

	if !ok || len(subs) == 0 {
		return &BrokerError{
			Code:    404,
			Message: "No subscription channels present for topic: " + topic,
		}
	}

	if b.config.AtLeastOnce {
		for _, ch := range subs {
			msgID := utils.GenerateMessageID()
			envelope := Envelope{
				ID:      msgID,
				Payload: msg,
			}
			envelopedData, err := json.Marshal(envelope)
			if err != nil {
				return &BrokerError{
					Code:    422,
					Message: err.Error(),
				}
			}

			select {
			case ch <- envelopedData:
				b.startAckTimer(topic, msgID, envelope, ch)
			default:
				return &BrokerError{
					Code:    503,
					Message: "Subscription channel for topic " + topic + " is full",
				}
			}
		}
	} else {
		for _, ch := range subs {
			select {
			case ch <- msg:
			default:
				return &BrokerError{
					Code:    503,
					Message: "Subscription channel for topic " + topic + " is full",
				}
			}
		}
	}
	return nil
}

func (b *Broker) startAckTimer(topic, msgID string, envelope Envelope, ch chan json.RawMessage) {
	timer := time.AfterFunc(b.config.AckTimeout, func() {
		b.mu.Lock()
		defer b.mu.Unlock()

		if pending, exists := b.pendingAcks[msgID]; exists {
			envelopedData, err := json.Marshal(envelope)
			if err != nil {
				// In production, you may log this error.
				return
			}
			select {
			case pending.subscriber <- envelopedData:
				pending.tries++
				pending.timer.Reset(b.config.AckTimeout)
				b.pendingAcks[msgID] = pending
			default:
				// Optionally log that the channel is still full.
			}
		}
	})

	b.mu.Lock()
	defer b.mu.Unlock()

	b.pendingAcks[msgID] = pendingMessage{
		topic:      topic,
		envelope:   envelope,
		subscriber: ch,
		tries:      1,
		timer:      timer,
	}
	
}

func (b *Broker) Acknowledge(topic, messageID string) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	
	pending, exists := b.pendingAcks[messageID]
	if !exists {
		return &BrokerError{
			Code:    404,
			Message: "No pending acknowledgment for messageID: " + messageID,
		}
	}

	pending.timer.Stop()
	delete(b.pendingAcks, messageID)
	return nil
}


func (b *Broker) Subscribe(topic string) chan json.RawMessage {
	ch := make(chan json.RawMessage, 10)
	b.mu.Lock()
	defer b.mu.Unlock()

	b.subs[topic] = append(b.subs[topic], ch)
	return ch
}

func (b *Broker) Unsubscribe(topic string, ch chan json.RawMessage) {
	b.mu.Lock()
	defer b.mu.Unlock()

	subs := b.subs[topic]
	for i, sub := range subs {
		if sub == ch {
			b.subs[topic] = append(subs[:i], subs[i+1:]...)
			close(ch)
			break
		}
	}
}
