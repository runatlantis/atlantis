package enhanced

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/runatlantis/atlantis/server/logging"
)

// EventType defines the types of events in the enhanced locking system
type EventType string

const (
	// Lock lifecycle events
	EventLockRequested EventType = "lock_requested"
	EventLockAcquired  EventType = "lock_acquired"
	EventLockReleased  EventType = "lock_released"
	EventLockExpired   EventType = "lock_expired"
	EventLockFailed    EventType = "lock_failed"

	// Queue events
	EventLockQueued        EventType = "lock_queued"
	EventQueuedLockAcquired EventType = "queued_lock_acquired"
	EventQueueTimeout      EventType = "queue_timeout"

	// System events
	EventDeadlockDetected  EventType = "deadlock_detected"
	EventDeadlockResolved  EventType = "deadlock_resolved"
	EventSystemMaintenance EventType = "system_maintenance"
	EventHealthChange      EventType = "health_change"
)

// EventManager handles enhanced locking events for PR #4
type EventManager struct {
	bufferSize   int
	eventChan    chan *LockEvent
	subscribers  map[string][]EventSubscriber
	mu           sync.RWMutex
	logger       logging.SimpleLogging
	running      bool
	stopChan     chan struct{}

	// Event history for debugging
	eventHistory []EventHistoryEntry
	historyLimit int
}

// EventSubscriber defines event handling interface
type EventSubscriber interface {
	HandleEvent(ctx context.Context, event *LockEvent) error
	GetID() string
	GetEventTypes() []EventType
}

// EventHistoryEntry stores events for analysis
type EventHistoryEntry struct {
	Event     *LockEvent `json:"event"`
	Timestamp time.Time  `json:"timestamp"`
	Source    string     `json:"source"`
}

// NewEventManager creates a new event manager for enhanced locking
func NewEventManager(bufferSize int, logger logging.SimpleLogging) *EventManager {
	return &EventManager{
		bufferSize:   bufferSize,
		eventChan:    make(chan *LockEvent, bufferSize),
		subscribers:  make(map[string][]EventSubscriber),
		logger:       logger,
		stopChan:     make(chan struct{}),
		eventHistory: make([]EventHistoryEntry, 0),
		historyLimit: 1000,
	}
}

// Start begins event processing
func (em *EventManager) Start(ctx context.Context) error {
	em.mu.Lock()
	defer em.mu.Unlock()

	if em.running {
		return fmt.Errorf("event manager already running")
	}

	em.running = true
	go em.eventProcessor(ctx)

	em.logger.Info("Enhanced locking event manager started (PR #4)")
	return nil
}

// Stop gracefully shuts down event processing
func (em *EventManager) Stop(ctx context.Context) error {
	em.mu.Lock()
	defer em.mu.Unlock()

	if !em.running {
		return nil
	}

	em.logger.Info("Stopping enhanced locking event manager...")
	close(em.stopChan)
	em.running = false

	return nil
}

// EmitEvent sends an event for processing
func (em *EventManager) EmitEvent(event *LockEvent) {
	if !em.IsHealthy() {
		em.logger.Warn("Event manager not healthy, dropping event: %s", event.Type)
		return
	}

	select {
	case em.eventChan <- event:
		// Event queued successfully
	default:
		em.logger.Warn("Event buffer full, dropping event: %s for lock %s", event.Type, event.LockID)
	}
}

// Subscribe adds an event subscriber
func (em *EventManager) Subscribe(subscriber EventSubscriber) error {
	em.mu.Lock()
	defer em.mu.Unlock()

	subscriberID := subscriber.GetID()
	eventTypes := subscriber.GetEventTypes()

	for _, eventType := range eventTypes {
		typeStr := string(eventType)
		em.subscribers[typeStr] = append(em.subscribers[typeStr], subscriber)
	}

	em.logger.Info("Event subscriber %s registered for %d event types", subscriberID, len(eventTypes))
	return nil
}

// IsHealthy checks event manager health
func (em *EventManager) IsHealthy() bool {
	em.mu.RLock()
	defer em.mu.RUnlock()

	if !em.running {
		return false
	}

	// Check buffer usage
	bufferUsage := float64(len(em.eventChan)) / float64(em.bufferSize)
	return bufferUsage < 0.95
}

// eventProcessor handles events in background
func (em *EventManager) eventProcessor(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-em.stopChan:
			return
		case event := <-em.eventChan:
			em.processEvent(ctx, event)
		}
	}
}

// processEvent handles a single event
func (em *EventManager) processEvent(ctx context.Context, event *LockEvent) {
	// Add to history
	em.addToHistory(event)

	// Get subscribers
	em.mu.RLock()
	subscribers := em.subscribers[event.Type]
	subscribersCopy := make([]EventSubscriber, len(subscribers))
	copy(subscribersCopy, subscribers)
	em.mu.RUnlock()

	// Process subscribers
	for _, subscriber := range subscribersCopy {
		em.processSubscriber(ctx, event, subscriber)
	}
}

// processSubscriber handles event for specific subscriber
func (em *EventManager) processSubscriber(ctx context.Context, event *LockEvent, subscriber EventSubscriber) {
	defer func() {
		if r := recover(); r != nil {
			em.logger.Error("Event subscriber %s panicked: %v", subscriber.GetID(), r)
		}
	}()

	subCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := subscriber.HandleEvent(subCtx, event); err != nil {
		em.logger.Error("Subscriber %s failed to handle event %s: %v",
			subscriber.GetID(), event.Type, err)
	}
}

// addToHistory stores event in history buffer
func (em *EventManager) addToHistory(event *LockEvent) {
	em.mu.Lock()
	defer em.mu.Unlock()

	entry := EventHistoryEntry{
		Event:     event,
		Timestamp: time.Now(),
		Source:    "enhanced_manager_pr4",
	}

	em.eventHistory = append(em.eventHistory, entry)

	// Trim history if needed
	if len(em.eventHistory) > em.historyLimit {
		copy(em.eventHistory, em.eventHistory[len(em.eventHistory)-em.historyLimit:])
		em.eventHistory = em.eventHistory[:em.historyLimit]
	}
}

// LoggingEventSubscriber logs all events
type LoggingEventSubscriber struct {
	id     string
	logger logging.SimpleLogging
}

// NewLoggingEventSubscriber creates logging subscriber
func NewLoggingEventSubscriber(id string, logger logging.SimpleLogging) *LoggingEventSubscriber {
	return &LoggingEventSubscriber{
		id:     id,
		logger: logger,
	}
}

func (les *LoggingEventSubscriber) HandleEvent(ctx context.Context, event *LockEvent) error {
	les.logger.Info("Enhanced lock event: %s for lock %s (resource: %s/%s, owner: %s)",
		event.Type, event.LockID, event.Resource.Namespace, event.Resource.Workspace, event.Owner)
	return nil
}

func (les *LoggingEventSubscriber) GetID() string {
	return les.id
}

func (les *LoggingEventSubscriber) GetEventTypes() []EventType {
	return []EventType{
		EventLockRequested, EventLockAcquired, EventLockReleased,
		EventLockFailed, EventLockQueued, EventQueuedLockAcquired,
	}
}