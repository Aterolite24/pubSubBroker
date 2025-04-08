package router

import (
	"log"
	"sync"

	"gobroker/internal/conn"
)

type Router struct {
	mu     sync.RWMutex
	topics map[string]map[*conn.Client]bool
}

func NewRouter() *Router {
	return &Router{
		topics: make(map[string]map[*conn.Client]bool),
	}
}

// Subscribe adds a client to a topic.
func (r *Router) Subscribe(topic string, client *conn.Client) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.topics[topic]; !ok {
		r.topics[topic] = make(map[*conn.Client]bool)
	}
	r.topics[topic][client] = true
	log.Printf("Client %s subscribed to topic %s", client.ID(), topic)
}

// Unsubscribe removes a client from a topic.
func (r *Router) Unsubscribe(topic string, client *conn.Client) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if subs, ok := r.topics[topic]; ok {
		delete(subs, client)
		log.Printf("Client %s unsubscribed from topic %s", client.ID(), topic)
		if len(subs) == 0 {
			delete(r.topics, topic)
		}
	}
}

// Publish sends a message to all clients subscribed to the topic.
func (r *Router) Publish(topic string, msg *conn.Message) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	subscribers, ok := r.topics[topic]
	if !ok {
		log.Printf("No subscribers for topic %s", topic)
		return
	}

	for client := range subscribers {
		client.SendJSON(msg)
	}
}
