package service

import (
	"sync"
	"time"
)

// Event types
const (
	EventSyncStarted     = "sync.started"
	EventSyncCompleted   = "sync.completed"
	EventSyncFailed      = "sync.failed"
	EventPeerDiscovered  = "peer.discovered"
	EventPeerLost        = "peer.lost"
	EventConflictNew     = "conflict.new"
	EventConflictResolved = "conflict.resolved"
)

// Event represents a sync system event.
type Event struct {
	Type      string
	Timestamp time.Time
	Payload   map[string]string
}

// Subscriber receives events.
type Subscriber struct {
	Ch      chan Event
	Filters []string // empty = all events
}

// EventBus provides pub/sub for sync events.
type EventBus struct {
	mu          sync.RWMutex
	subscribers []*Subscriber
	bufferSize  int
}

// NewEventBus creates a new event bus.
func NewEventBus(bufferSize int) *EventBus {
	if bufferSize <= 0 {
		bufferSize = 100
	}
	return &EventBus{bufferSize: bufferSize}
}

// Subscribe creates a new subscriber with optional type filtering.
func (eb *EventBus) Subscribe(filters []string) *Subscriber {
	sub := &Subscriber{
		Ch:      make(chan Event, eb.bufferSize),
		Filters: filters,
	}
	eb.mu.Lock()
	eb.subscribers = append(eb.subscribers, sub)
	eb.mu.Unlock()
	return sub
}

// Unsubscribe removes a subscriber.
func (eb *EventBus) Unsubscribe(sub *Subscriber) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	for i, s := range eb.subscribers {
		if s == sub {
			eb.subscribers = append(eb.subscribers[:i], eb.subscribers[i+1:]...)
			close(s.Ch)
			return
		}
	}
}

// Publish sends an event to all matching subscribers.
func (eb *EventBus) Publish(event Event) {
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	eb.mu.RLock()
	defer eb.mu.RUnlock()

	for _, sub := range eb.subscribers {
		if matchesFilter(event.Type, sub.Filters) {
			select {
			case sub.Ch <- event:
			default:
				// Drop event if subscriber buffer is full
			}
		}
	}
}

func matchesFilter(eventType string, filters []string) bool {
	if len(filters) == 0 {
		return true
	}
	for _, f := range filters {
		if f == eventType {
			return true
		}
	}
	return false
}
