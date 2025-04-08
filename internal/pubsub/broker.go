package pubsub

import "slices"

import "sync"

type Broker struct {
	mu    sync.RWMutex
	subs  map[string][]chan string
}

func NewBroker() *Broker {
	return &Broker{
		subs: make(map[string][]chan string),
	}
}

func (b *Broker) Publish(topic, msg string) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, ch := range b.subs[topic] {
		select {
		case ch <- msg:
		default:
		}
	}
}

func (b *Broker) Subscribe(topic string) chan string {
	ch := make(chan string, 10)
	b.mu.Lock()
	b.subs[topic] = append(b.subs[topic], ch)
	b.mu.Unlock()
	return ch
}

func (b *Broker) Unsubscribe(topic string, ch chan string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	subs := b.subs[topic]
	for i, sub := range subs {
		if sub == ch {
			b.subs[topic] = slices.Delete(subs, i, i+1)
			close(ch)
			break
		}
	}
}
