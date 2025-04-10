package models

import (
	"encoding/json"
	"sync"
	"time"

	"context"
	"fmt"

	"pubSubBroker/internal/utils"

	"go.mongodb.org/mongo-driver/v2/bson"
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

func (b *Broker) insertDatabase(topic string, msg json.RawMessage) {
	connection, err1 := utils.GetMongoClient()
	if err1 != nil {
		fmt.Println("Error : ", err1)
		return
	}

	message := DbMessageData{
		Topic:     topic,
		Message:   msg,
		TimeStamp: time.Now(),
	}

	_, err := connection.InsertOne(context.TODO(), message)

	if err != nil {
		fmt.Println("Error : ", err)
		return
	}
}

func (b *Broker) Publish(topic string, msg json.RawMessage) error {
	b.mu.RLock()
	defer b.mu.RUnlock()
	subs := b.subs[topic]

	go b.insertDatabase(topic, msg)

	if b.config.AtLeastOnce {
		for _, ch := range subs {
			msgID := utils.GenerateMessageID()
			envelope := Envelope{
				ID:      msgID,
				Topic:   topic,
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
				go b.startAckTimer(topic, msgID, envelope, ch)
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
				fmt.Println("Error")
				return
			}

			ok := true
			select {
			case _, ok = <-pending.subscriber:
			default:
			}

			if ok {
				select {

				case pending.subscriber <- envelopedData:
					pending.tries++
					pending.timer.Reset(b.config.AckTimeout)
					b.pendingAcks[msgID] = pending
				default:
				}
			} else {
				pending.timer.Stop()
				delete(b.pendingAcks, msgID)
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

func (b *Broker) Subscribe(topic string, timeStamp time.Time) chan json.RawMessage {
	ch := make(chan json.RawMessage, 10)
	b.mu.Lock()
	defer b.mu.Unlock()

	b.subs[topic] = append(b.subs[topic], ch)

	connection, err := utils.GetMongoClient()

	if err != nil {
		fmt.Println("Cant connect to db..")
	}

	res, err1 := connection.Find(context.TODO(),
		bson.D{
			{Key: "timestamp", Value: bson.D{{Key: "$gt", Value: timeStamp}}},
			{Key: "topic", Value: topic},
		})

	if err1 != nil {
		fmt.Println("Cant connect to db..")
	}
	defer res.Close(context.TODO())

	if b.config.AtLeastOnce {
		for res.Next(context.TODO()) {
			var result PublishData
			if err := res.Decode(&result); err != nil {
				fmt.Println("error")
			}

			msgID := utils.GenerateMessageID()
			envelope := Envelope{
				ID:      msgID,
				Topic:   topic,
				Payload: result.Message,
			}
			envelopedData, err := json.Marshal(envelope)
			if err != nil {
				return ch
			}
			select {
			case ch <- envelopedData:
				go b.startAckTimer(topic, msgID, envelope, ch)
			default:
				return nil
			}
		}
	} else {
		for res.Next(context.TODO()) {

			var result PublishData
			if err := res.Decode(&result); err != nil {
				fmt.Println("error")
			}

			finalResult, err := json.Marshal(result.Message)
			if err != nil {
				return ch
			}

			select {
			case ch <- finalResult:
			default:
				return nil
			}
		}
	}

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
