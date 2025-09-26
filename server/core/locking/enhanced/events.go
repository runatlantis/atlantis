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
	EventLockRequested  EventType = "lock_requested"
	EventLockAcquired   EventType = "lock_acquired"
	EventLockReleased   EventType = "lock_released"
	EventLockTimeout    EventType = "lock_timeout"
	EventLockTransferred EventType = "lock_transferred"
	EventLockRefreshed  EventType = "lock_refreshed"

	// Queue events
	EventQueuedRequest    EventType = "queued_request"
	EventQueueProcessed   EventType = "queue_processed"
	EventQueueTimeout     EventType = "queue_timeout"
	EventQueueCancelled   EventType = "queue_cancelled"

	// System events
	EventDeadlockDetected EventType = "deadlock_detected"
	EventDeadlockResolved EventType = "deadlock_resolved"
	EventSystemHealthCheck EventType = "system_health_check"
	EventMetricsUpdated   EventType = "metrics_updated"
	EventConfigUpdated    EventType = "config_updated"

	// Error events
	EventBackendError     EventType = "backend_error"
	EventTimeoutError     EventType = "timeout_error"
	EventValidationError  EventType = "validation_error"
)

// LockEventData contains detailed information about a lock event
type LockEventData struct {
	Type         EventType              `json:"type"`
	EventID      string                 `json:"event_id"`
	LockID       string                 `json:"lock_id,omitempty"`
	RequestID    string                 `json:"request_id,omitempty"`
	Resource     ResourceIdentifier     `json:"resource"`
	Owner        string                 `json:"owner,omitempty"`
	Priority     Priority               `json:"priority,omitempty"`
	Timestamp    time.Time              `json:"timestamp"`
	Duration     time.Duration          `json:"duration,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	Error        string                 `json:"error,omitempty"`
	PreviousOwner string                `json:"previous_owner,omitempty"`
	QueuePosition int                   `json:"queue_position,omitempty"`
}

// EventHandler defines the interface for event handlers
type EventHandler interface {
	HandleEvent(ctx context.Context, event *LockEventData) error
	GetEventTypes() []EventType
	GetName() string
}

// EventSubscription represents a subscription to events
type EventSubscription struct {
	ID          string
	EventTypes  []EventType
	Handler     EventHandler
	Filter      EventFilter
	Created     time.Time
	LastEvent   *time.Time
	EventCount  int64
	Active      bool
}

// EventFilter allows filtering events before delivery
type EventFilter func(event *LockEventData) bool

// EventBus manages event publishing and subscription
type EventBus struct {
	subscriptions map[string]*EventSubscription
	mutex         sync.RWMutex
	bufferSize    int
	eventBuffer   chan *LockEventData
	log           logging.SimpleLogging
	metrics       *EventMetrics
	running       bool
	stopChan      chan struct{}
	wg            sync.WaitGroup
}

// EventMetrics tracks event system performance
type EventMetrics struct {
	TotalEvents         int64                    `json:"total_events"`
	EventsByType        map[EventType]int64      `json:"events_by_type"`
	SubscriptionCount   int                      `json:"subscription_count"`
	ActiveSubscriptions int                      `json:"active_subscriptions"`
	FailedDeliveries    int64                    `json:"failed_deliveries"`
	AverageDeliveryTime time.Duration            `json:"average_delivery_time"`
	BufferUtilization   float64                  `json:"buffer_utilization"`
	LastUpdated         time.Time                `json:"last_updated"`
	mutex               sync.RWMutex
}

// NewEventBus creates a new event bus for the locking system
func NewEventBus(bufferSize int, log logging.SimpleLogging) *EventBus {
	return &EventBus{
		subscriptions: make(map[string]*EventSubscription),
		bufferSize:    bufferSize,
		eventBuffer:   make(chan *LockEventData, bufferSize),
		log:           log,
		metrics: &EventMetrics{
			EventsByType: make(map[EventType]int64),
			LastUpdated:  time.Now(),
		},
		stopChan: make(chan struct{}),
	}
}

// Start begins event processing
func (eb *EventBus) Start(ctx context.Context) error {
	eb.mutex.Lock()
	defer eb.mutex.Unlock()

	if eb.running {
		return fmt.Errorf("event bus is already running")
	}

	eb.running = true
	eb.wg.Add(1)
	go eb.processEvents(ctx)

	eb.log.Info("Event bus started with buffer size: %d", eb.bufferSize)
	return nil
}

// Stop gracefully shuts down the event bus
func (eb *EventBus) Stop() error {
	eb.mutex.Lock()
	defer eb.mutex.Unlock()

	if !eb.running {
		return nil
	}

	eb.log.Info("Stopping event bus...")
	close(eb.stopChan)
	eb.running = false

	eb.wg.Wait()
	close(eb.eventBuffer)

	eb.log.Info("Event bus stopped")
	return nil
}

// Publish publishes an event to all subscribed handlers
func (eb *EventBus) Publish(ctx context.Context, event *LockEventData) error {
	if event == nil {
		return fmt.Errorf("event cannot be nil")
	}

	// Set event ID if not provided
	if event.EventID == "" {
		event.EventID = eb.generateEventID()
	}

	// Set timestamp if not provided
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	select {
	case eb.eventBuffer <- event:
		eb.updateMetrics(event)
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		// Buffer full - drop event or handle overflow
		eb.metrics.mutex.Lock()
		eb.metrics.FailedDeliveries++
		eb.metrics.mutex.Unlock()

		eb.log.Warn("Event buffer full, dropping event: %s", event.Type)
		return fmt.Errorf("event buffer full")
	}
}

// Subscribe creates a subscription to events
func (eb *EventBus) Subscribe(handler EventHandler, filter EventFilter) *EventSubscription {
	eb.mutex.Lock()
	defer eb.mutex.Unlock()

	subscription := &EventSubscription{
		ID:         eb.generateSubscriptionID(),
		EventTypes: handler.GetEventTypes(),
		Handler:    handler,
		Filter:     filter,
		Created:    time.Now(),
		Active:     true,
	}

	eb.subscriptions[subscription.ID] = subscription

	eb.metrics.mutex.Lock()
	eb.metrics.SubscriptionCount++
	eb.metrics.ActiveSubscriptions++
	eb.metrics.mutex.Unlock()

	eb.log.Info("New event subscription created: %s for handler: %s",
		subscription.ID, handler.GetName())

	return subscription
}

// Unsubscribe removes a subscription
func (eb *EventBus) Unsubscribe(subscriptionID string) error {
	eb.mutex.Lock()
	defer eb.mutex.Unlock()

	subscription, exists := eb.subscriptions[subscriptionID]
	if !exists {
		return fmt.Errorf("subscription not found: %s", subscriptionID)
	}

	subscription.Active = false
	delete(eb.subscriptions, subscriptionID)

	eb.metrics.mutex.Lock()
	eb.metrics.ActiveSubscriptions--
	eb.metrics.mutex.Unlock()

	eb.log.Info("Event subscription removed: %s", subscriptionID)
	return nil
}

// GetSubscriptions returns all active subscriptions
func (eb *EventBus) GetSubscriptions() []*EventSubscription {
	eb.mutex.RLock()
	defer eb.mutex.RUnlock()

	var subscriptions []*EventSubscription
	for _, sub := range eb.subscriptions {
		if sub.Active {
			subscriptions = append(subscriptions, sub)
		}
	}

	return subscriptions
}

// GetMetrics returns event bus metrics
func (eb *EventBus) GetMetrics() *EventMetrics {
	eb.metrics.mutex.RLock()
	defer eb.metrics.mutex.RUnlock()

	// Create a copy to avoid concurrent access issues
	metrics := &EventMetrics{
		TotalEvents:         eb.metrics.TotalEvents,
		EventsByType:        make(map[EventType]int64),
		SubscriptionCount:   eb.metrics.SubscriptionCount,
		ActiveSubscriptions: eb.metrics.ActiveSubscriptions,
		FailedDeliveries:    eb.metrics.FailedDeliveries,
		AverageDeliveryTime: eb.metrics.AverageDeliveryTime,
		BufferUtilization:   float64(len(eb.eventBuffer)) / float64(eb.bufferSize),
		LastUpdated:         time.Now(),
	}

	for eventType, count := range eb.metrics.EventsByType {
		metrics.EventsByType[eventType] = count
	}

	return metrics
}

// processEvents processes events from the buffer
func (eb *EventBus) processEvents(ctx context.Context) {
	defer eb.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case <-eb.stopChan:
			return
		case event := <-eb.eventBuffer:
			if event == nil {
				continue
			}
			eb.deliverEvent(ctx, event)
		}
	}
}

// deliverEvent delivers an event to all matching subscriptions
func (eb *EventBus) deliverEvent(ctx context.Context, event *LockEventData) {
	startTime := time.Now()

	eb.mutex.RLock()
	subscriptions := make([]*EventSubscription, 0, len(eb.subscriptions))
	for _, sub := range eb.subscriptions {
		if sub.Active && eb.matchesSubscription(event, sub) {
			subscriptions = append(subscriptions, sub)
		}
	}
	eb.mutex.RUnlock()

	// Deliver to matching subscriptions
	for _, subscription := range subscriptions {
		go eb.deliverToSubscription(ctx, event, subscription)
	}

	// Update delivery time metrics
	deliveryTime := time.Since(startTime)
	eb.metrics.mutex.Lock()
	if eb.metrics.AverageDeliveryTime == 0 {
		eb.metrics.AverageDeliveryTime = deliveryTime
	} else {
		eb.metrics.AverageDeliveryTime = (eb.metrics.AverageDeliveryTime + deliveryTime) / 2
	}
	eb.metrics.mutex.Unlock()
}

// deliverToSubscription delivers an event to a specific subscription
func (eb *EventBus) deliverToSubscription(ctx context.Context, event *LockEventData, subscription *EventSubscription) {
	defer func() {
		if r := recover(); r != nil {
			eb.log.Error("Event handler panicked for subscription %s: %v", subscription.ID, r)
			eb.metrics.mutex.Lock()
			eb.metrics.FailedDeliveries++
			eb.metrics.mutex.Unlock()
		}
	}()

	// Apply filter if present
	if subscription.Filter != nil && !subscription.Filter(event) {
		return
	}

	// Deliver event with timeout
	deliveryCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := subscription.Handler.HandleEvent(deliveryCtx, event); err != nil {
		eb.log.Error("Event delivery failed for subscription %s: %v", subscription.ID, err)
		eb.metrics.mutex.Lock()
		eb.metrics.FailedDeliveries++
		eb.metrics.mutex.Unlock()
		return
	}

	// Update subscription metrics
	now := time.Now()
	subscription.LastEvent = &now
	subscription.EventCount++
}

// matchesSubscription checks if an event matches a subscription
func (eb *EventBus) matchesSubscription(event *LockEventData, subscription *EventSubscription) bool {
	if len(subscription.EventTypes) == 0 {
		return true // Subscribe to all events
	}

	for _, eventType := range subscription.EventTypes {
		if event.Type == eventType {
			return true
		}
	}

	return false
}

// updateMetrics updates event metrics
func (eb *EventBus) updateMetrics(event *LockEventData) {
	eb.metrics.mutex.Lock()
	defer eb.metrics.mutex.Unlock()

	eb.metrics.TotalEvents++
	eb.metrics.EventsByType[event.Type]++
	eb.metrics.LastUpdated = time.Now()
}

// generateEventID generates a unique event ID
func (eb *EventBus) generateEventID() string {
	return fmt.Sprintf("evt_%d_%d", time.Now().UnixNano(), eb.metrics.TotalEvents)
}

// generateSubscriptionID generates a unique subscription ID
func (eb *EventBus) generateSubscriptionID() string {
	return fmt.Sprintf("sub_%d_%d", time.Now().UnixNano(), eb.metrics.SubscriptionCount)
}

// Built-in event handlers

// LoggingEventHandler logs all events
type LoggingEventHandler struct {
	log        logging.SimpleLogging
	eventTypes []EventType
	name       string
}

// NewLoggingEventHandler creates a new logging event handler
func NewLoggingEventHandler(log logging.SimpleLogging, eventTypes []EventType) *LoggingEventHandler {
	if len(eventTypes) == 0 {
		// Subscribe to all events
		eventTypes = []EventType{
			EventLockRequested, EventLockAcquired, EventLockReleased,
			EventLockTimeout, EventLockTransferred, EventLockRefreshed,
			EventQueuedRequest, EventQueueProcessed, EventQueueTimeout,
			EventDeadlockDetected, EventDeadlockResolved,
		}
	}

	return &LoggingEventHandler{
		log:        log,
		eventTypes: eventTypes,
		name:       "LoggingEventHandler",
	}
}

// HandleEvent implements EventHandler
func (leh *LoggingEventHandler) HandleEvent(ctx context.Context, event *LockEventData) error {
	level := "INFO"
	if event.Error != "" {
		level = "ERROR"
	}

	message := fmt.Sprintf("[%s] Lock event: %s - Resource: %s/%s",
		level, event.Type, event.Resource.Namespace, event.Resource.Name)

	if event.Owner != "" {
		message += fmt.Sprintf(" Owner: %s", event.Owner)
	}

	if event.Duration > 0 {
		message += fmt.Sprintf(" Duration: %v", event.Duration)
	}

	if event.Error != "" {
		message += fmt.Sprintf(" Error: %s", event.Error)
	}

	if event.Error != "" {
		leh.log.Error(message)
	} else {
		leh.log.Info(message)
	}

	return nil
}

// GetEventTypes implements EventHandler
func (leh *LoggingEventHandler) GetEventTypes() []EventType {
	return leh.eventTypes
}

// GetName implements EventHandler
func (leh *LoggingEventHandler) GetName() string {
	return leh.name
}

// MetricsEventHandler updates metrics based on events
type MetricsEventHandler struct {
	metrics    *ManagerMetrics
	eventTypes []EventType
	name       string
}

// NewMetricsEventHandler creates a new metrics event handler
func NewMetricsEventHandler(metrics *ManagerMetrics) *MetricsEventHandler {
	return &MetricsEventHandler{
		metrics: metrics,
		eventTypes: []EventType{
			EventLockAcquired, EventLockReleased, EventLockTimeout,
			EventQueuedRequest, EventDeadlockDetected, EventDeadlockResolved,
		},
		name: "MetricsEventHandler",
	}
}

// HandleEvent implements EventHandler
func (meh *MetricsEventHandler) HandleEvent(ctx context.Context, event *LockEventData) error {
	switch event.Type {
	case EventLockAcquired:
		meh.metrics.SuccessfulLocks++
		meh.metrics.ActiveLocks++
	case EventLockReleased, EventLockTimeout:
		if meh.metrics.ActiveLocks > 0 {
			meh.metrics.ActiveLocks--
		}
		if event.Duration > 0 {
			// Update average hold time
			if meh.metrics.AverageHoldTime == 0 {
				meh.metrics.AverageHoldTime = event.Duration
			} else {
				meh.metrics.AverageHoldTime = (meh.metrics.AverageHoldTime + event.Duration) / 2
			}
		}
	case EventQueuedRequest:
		meh.metrics.QueuedRequests++
	case EventDeadlockDetected:
		meh.metrics.DeadlocksDetected++
	case EventDeadlockResolved:
		meh.metrics.DeadlocksResolved++
	}

	meh.metrics.LastUpdated = time.Now()
	return nil
}

// GetEventTypes implements EventHandler
func (meh *MetricsEventHandler) GetEventTypes() []EventType {
	return meh.eventTypes
}

// GetName implements EventHandler
func (meh *MetricsEventHandler) GetName() string {
	return meh.name
}

// Event convenience functions

// CreateLockEvent creates a lock-related event
func CreateLockEvent(eventType EventType, lockID string, resource ResourceIdentifier, owner string, duration time.Duration) *LockEventData {
	return &LockEventData{
		Type:      eventType,
		EventID:   "",
		LockID:    lockID,
		Resource:  resource,
		Owner:     owner,
		Timestamp: time.Now(),
		Duration:  duration,
		Metadata:  make(map[string]interface{}),
	}
}

// CreateQueueEvent creates a queue-related event
func CreateQueueEvent(eventType EventType, requestID string, resource ResourceIdentifier, priority Priority, position int) *LockEventData {
	return &LockEventData{
		Type:          eventType,
		EventID:       "",
		RequestID:     requestID,
		Resource:      resource,
		Priority:      priority,
		Timestamp:     time.Now(),
		QueuePosition: position,
		Metadata:      make(map[string]interface{}),
	}
}

// CreateErrorEvent creates an error event
func CreateErrorEvent(eventType EventType, resource ResourceIdentifier, error string) *LockEventData {
	return &LockEventData{
		Type:      eventType,
		EventID:   "",
		Resource:  resource,
		Timestamp: time.Now(),
		Error:     error,
		Metadata:  make(map[string]interface{}),
	}
}