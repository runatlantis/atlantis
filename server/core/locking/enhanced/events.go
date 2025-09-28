package enhanced

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/runatlantis/atlantis/server/logging"
)

// EventManager provides comprehensive event tracking and subscription capabilities
// This is a key component of PR #4 - Enhanced Manager and Events
type EventManager struct {
	logger      logging.SimpleLogging
	bufferSize  int
	eventChan   chan *LockEvent
	subscribers map[string][]EventSubscriber
	history     []*EventHistoryEntry

	// Manager state
	mu       sync.RWMutex
	started  bool
	stopCh   chan struct{}
	workerWg sync.WaitGroup

	// Statistics
	stats *EventStats
}

// EventSubscriber interface for event consumers
type EventSubscriber interface {
	OnEvent(event *LockEvent) error
	GetID() string
	GetEventTypes() []string
}

// EventHistoryEntry represents a historical event with metadata
type EventHistoryEntry struct {
	Event       *LockEvent `json:"event"`
	ProcessedAt time.Time  `json:"processed_at"`
	Subscribers []string   `json:"subscribers"`
	Errors      []string   `json:"errors,omitempty"`
}

// EventStats provides statistics about event processing
type EventStats struct {
	TotalEvents       int64            `json:"total_events"`
	EventsByType      map[string]int64 `json:"events_by_type"`
	ActiveSubscribers int              `json:"active_subscribers"`
	BufferUtilization float64          `json:"buffer_utilization"`
	ProcessingRate    float64          `json:"processing_rate"` // events per second
	LastEventTime     *time.Time       `json:"last_event_time,omitempty"`
	ErrorRate         float64          `json:"error_rate"`
	HistorySize       int              `json:"history_size"`
}

// NewEventManager creates a new event manager
func NewEventManager(bufferSize int, logger logging.SimpleLogging) *EventManager {
	if bufferSize <= 0 {
		bufferSize = 1000 // Default buffer size
	}

	return &EventManager{
		logger:      logger,
		bufferSize:  bufferSize,
		eventChan:   make(chan *LockEvent, bufferSize),
		subscribers: make(map[string][]EventSubscriber),
		history:     make([]*EventHistoryEntry, 0),
		stopCh:      make(chan struct{}),
		stats: &EventStats{
			EventsByType: make(map[string]int64),
		},
	}
}

// Start starts the event manager
func (em *EventManager) Start(ctx context.Context) error {
	em.mu.Lock()
	defer em.mu.Unlock()

	if em.started {
		return fmt.Errorf("event manager already started")
	}

	em.logger.Info("Starting Enhanced Event Manager (PR #4)")

	// Start event processing worker
	em.workerWg.Add(1)
	go em.eventProcessor(ctx)

	em.started = true
	em.logger.Info("Enhanced Event Manager started")
	return nil
}

// Stop stops the event manager
func (em *EventManager) Stop(ctx context.Context) error {
	em.mu.Lock()
	defer em.mu.Unlock()

	if !em.started {
		return nil
	}

	em.logger.Info("Stopping Enhanced Event Manager (PR #4)")

	// Signal stop
	close(em.stopCh)

	// Wait for worker to finish
	done := make(chan struct{})
	go func() {
		em.workerWg.Wait()
		close(done)
	}()

	select {
	case <-done:
		em.logger.Info("Event manager stopped")
	case <-ctx.Done():
		em.logger.Warn("Context cancelled while stopping event manager")
	}

	em.started = false
	return nil
}

// Emit emits an event to all subscribers
func (em *EventManager) Emit(event *LockEvent) {
	if event == nil {
		return
	}

	select {
	case em.eventChan <- event:
		// Event queued successfully
	default:
		// Buffer full, log warning
		em.logger.Warn("Event buffer full, dropping event: %s", event.Type)
	}
}

// Subscribe adds an event subscriber
func (em *EventManager) Subscribe(subscriber EventSubscriber) error {
	if subscriber == nil {
		return fmt.Errorf("subscriber cannot be nil")
	}

	em.mu.Lock()
	defer em.mu.Unlock()

	subscriberID := subscriber.GetID()
	eventTypes := subscriber.GetEventTypes()

	if len(eventTypes) == 0 {
		eventTypes = []string{"*"} // Subscribe to all events
	}

	for _, eventType := range eventTypes {
		em.subscribers[eventType] = append(em.subscribers[eventType], subscriber)
	}

	em.logger.Info("Registered event subscriber: %s for types: %v", subscriberID, eventTypes)
	return nil
}

// Unsubscribe removes an event subscriber
func (em *EventManager) Unsubscribe(subscriberID string) error {
	em.mu.Lock()
	defer em.mu.Unlock()

	removed := 0
	for eventType, subs := range em.subscribers {
		newSubs := make([]EventSubscriber, 0, len(subs))
		for _, sub := range subs {
			if sub.GetID() != subscriberID {
				newSubs = append(newSubs, sub)
			} else {
				removed++
			}
		}
		em.subscribers[eventType] = newSubs
	}

	if removed > 0 {
		em.logger.Info("Unregistered event subscriber: %s (removed %d subscriptions)", subscriberID, removed)
	}

	return nil
}

// GetStats returns event processing statistics
func (em *EventManager) GetStats() *EventStats {
	em.mu.RLock()
	defer em.mu.RUnlock()

	// Calculate buffer utilization
	bufferUtilization := float64(len(em.eventChan)) / float64(em.bufferSize) * 100

	// Count active subscribers
	activeSubscribers := 0
	for _, subs := range em.subscribers {
		activeSubscribers += len(subs)
	}

	// Create a copy of stats with current values
	statsCopy := *em.stats
	statsCopy.BufferUtilization = bufferUtilization
	statsCopy.ActiveSubscribers = activeSubscribers
	statsCopy.HistorySize = len(em.history)

	return &statsCopy
}

// GetHistory returns recent event history
func (em *EventManager) GetHistory(limit int) []*EventHistoryEntry {
	em.mu.RLock()
	defer em.mu.RUnlock()

	if limit <= 0 || limit > len(em.history) {
		limit = len(em.history)
	}

	// Return the most recent events
	start := len(em.history) - limit
	if start < 0 {
		start = 0
	}

	history := make([]*EventHistoryEntry, limit)
	copy(history, em.history[start:])
	return history
}

// eventProcessor processes events from the channel and distributes to subscribers
func (em *EventManager) eventProcessor(ctx context.Context) {
	defer em.workerWg.Done()

	em.logger.Debug("Event processor started")
	defer em.logger.Debug("Event processor stopped")

	ticker := time.NewTicker(1 * time.Second) // For rate calculation
	defer ticker.Stop()

	lastEventCount := int64(0)

	for {
		select {
		case <-em.stopCh:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Update processing rate
			em.mu.Lock()
			currentCount := em.stats.TotalEvents
			em.stats.ProcessingRate = float64(currentCount - lastEventCount)
			lastEventCount = currentCount
			em.mu.Unlock()

		case event := <-em.eventChan:
			em.processEvent(event)
		}
	}
}

// processEvent processes a single event and distributes to subscribers
func (em *EventManager) processEvent(event *LockEvent) {
	startTime := time.Now()

	em.mu.Lock()

	// Update statistics
	em.stats.TotalEvents++
	em.stats.EventsByType[event.Type]++
	now := time.Now()
	em.stats.LastEventTime = &now

	// Find subscribers for this event type
	subscribers := make([]EventSubscriber, 0)

	// Add subscribers for specific event type
	if subs, exists := em.subscribers[event.Type]; exists {
		subscribers = append(subscribers, subs...)
	}

	// Add subscribers for all events (*)
	if subs, exists := em.subscribers["*"]; exists {
		subscribers = append(subscribers, subs...)
	}

	em.mu.Unlock()

	// Process subscribers
	historyEntry := &EventHistoryEntry{
		Event:       event,
		ProcessedAt: startTime,
		Subscribers: make([]string, len(subscribers)),
		Errors:      make([]string, 0),
	}

	for i, subscriber := range subscribers {
		historyEntry.Subscribers[i] = subscriber.GetID()

		if err := subscriber.OnEvent(event); err != nil {
			errorMsg := fmt.Sprintf("Subscriber %s failed to process event %s: %v",
				subscriber.GetID(), event.Type, err)
			em.logger.Err(errorMsg)
			historyEntry.Errors = append(historyEntry.Errors, errorMsg)
		}
	}

	// Update error rate
	em.mu.Lock()
	if len(historyEntry.Errors) > 0 {
		// Simple error rate calculation (could be more sophisticated)
		errorCount := int64(len(historyEntry.Errors))
		totalEvents := em.stats.TotalEvents
		em.stats.ErrorRate = float64(errorCount) / float64(totalEvents) * 100
	}

	// Add to history (with size limit)
	em.history = append(em.history, historyEntry)
	if len(em.history) > 1000 { // Keep last 1000 events
		em.history = em.history[1:]
	}

	em.mu.Unlock()

	em.logger.Debug("Processed event %s for %d subscribers in %v",
		event.Type, len(subscribers), time.Since(startTime))
}

// LoggingEventSubscriber is a simple subscriber that logs events
type LoggingEventSubscriber struct {
	id         string
	eventTypes []string
	logger     logging.SimpleLogging
}

// NewLoggingEventSubscriber creates a new logging event subscriber
func NewLoggingEventSubscriber(id string, logger logging.SimpleLogging, eventTypes ...string) *LoggingEventSubscriber {
	if len(eventTypes) == 0 {
		eventTypes = []string{"*"} // Subscribe to all events by default
	}

	return &LoggingEventSubscriber{
		id:         id,
		eventTypes: eventTypes,
		logger:     logger,
	}
}

// OnEvent implements EventSubscriber interface
func (les *LoggingEventSubscriber) OnEvent(event *LockEvent) error {
	les.logger.Info("Event [%s]: %s for %s (owner: %s)",
		event.Type, event.LockID, event.Resource.Name, event.Owner)
	return nil
}

// GetID implements EventSubscriber interface
func (les *LoggingEventSubscriber) GetID() string {
	return les.id
}

// GetEventTypes implements EventSubscriber interface
func (les *LoggingEventSubscriber) GetEventTypes() []string {
	return les.eventTypes
}

// MetricsEventSubscriber forwards events to metrics collector
type MetricsEventSubscriber struct {
	id               string
	metricsCollector *MetricsCollector
}

// NewMetricsEventSubscriber creates a new metrics event subscriber
func NewMetricsEventSubscriber(metricsCollector *MetricsCollector) *MetricsEventSubscriber {
	return &MetricsEventSubscriber{
		id:               "metrics_collector",
		metricsCollector: metricsCollector,
	}
}

// OnEvent implements EventSubscriber interface
func (mes *MetricsEventSubscriber) OnEvent(event *LockEvent) error {
	if mes.metricsCollector != nil {
		mes.metricsCollector.RecordEvent(event)
	}
	return nil
}

// GetID implements EventSubscriber interface
func (mes *MetricsEventSubscriber) GetID() string {
	return mes.id
}

// GetEventTypes implements EventSubscriber interface
func (mes *MetricsEventSubscriber) GetEventTypes() []string {
	return []string{"*"} // Subscribe to all events
}
