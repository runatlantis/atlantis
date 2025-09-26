// Copyright 2017 HootSuite Media Inc.
//
// Licensed under the Apache License, Version 2.0 (the License);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an AS IS BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// Modified hereafter by contributors to runatlantis/atlantis.

// Package queue provides priority-based queueing for the enhanced locking system.
package queue

import (
	"container/heap"
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
)

// Priority levels for lock requests
type Priority int

const (
	PriorityLow Priority = iota
	PriorityNormal
	PriorityHigh
	PriorityCritical
	PriorityEmergency
)

// String returns a string representation of the priority
func (p Priority) String() string {
	switch p {
	case PriorityLow:
		return "low"
	case PriorityNormal:
		return "normal"
	case PriorityHigh:
		return "high"
	case PriorityCritical:
		return "critical"
	case PriorityEmergency:
		return "emergency"
	default:
		return "unknown"
	}
}

// QueueItem represents a lock request in the queue
type QueueItem struct {
	// Core lock information
	Project   models.Project
	Workspace string
	Pull      models.PullRequest
	User      models.User

	// Queue metadata
	Priority  Priority
	Timestamp time.Time
	RequestID string

	// Channel for notifications
	ResultChan chan QueueResult

	// Context for cancellation
	Ctx context.Context

	// Metrics
	QueueEntryTime time.Time
	RetryCount     int
}

// QueueResult represents the result of a queue operation
type QueueResult struct {
	Success      bool
	Lock         *models.ProjectLock
	Error        error
	Position     int
	WaitTime     time.Duration
	TotalWaiters int
}

// PriorityQueue implements a priority-based queue for lock requests
type PriorityQueue struct {
	mu       sync.RWMutex
	items    []*QueueItem
	lookup   map[string]*QueueItem // requestID -> item mapping
	metrics  *QueueMetrics
	logger   logging.SimpleLogging
	config   QueueConfig
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}

// QueueConfig configures the priority queue behavior
type QueueConfig struct {
	MaxQueueSize        int
	BatchProcessingSize int
	ProcessingInterval  time.Duration
	MaxWaitTime         time.Duration
	MaxRetries          int
	PriorityWeights     map[Priority]int
	EnableBatching      bool
	EnableMetrics       bool
}

// QueueMetrics tracks queue performance
type QueueMetrics struct {
	mu                    sync.RWMutex
	TotalEnqueued         int64
	TotalProcessed        int64
	TotalTimeouts         int64
	TotalErrors           int64
	CurrentQueueSize      int64
	AverageWaitTime       time.Duration
	PriorityDistribution  map[Priority]int64
	BatchesProcessed      int64
	LastProcessingTime    time.Time
	ProcessingDuration    time.Duration
}

// NewPriorityQueue creates a new priority queue
func NewPriorityQueue(config QueueConfig, logger logging.SimpleLogging) *PriorityQueue {
	ctx, cancel := context.WithCancel(context.Background())

	// Set default weights if not provided
	if config.PriorityWeights == nil {
		config.PriorityWeights = map[Priority]int{
			PriorityLow:       1,
			PriorityNormal:    5,
			PriorityHigh:      10,
			PriorityCritical:  20,
			PriorityEmergency: 50,
		}
	}

	// Set default values
	if config.MaxQueueSize == 0 {
		config.MaxQueueSize = 1000
	}
	if config.BatchProcessingSize == 0 {
		config.BatchProcessingSize = 10
	}
	if config.ProcessingInterval == 0 {
		config.ProcessingInterval = 100 * time.Millisecond
	}
	if config.MaxWaitTime == 0 {
		config.MaxWaitTime = 10 * time.Minute
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}

	pq := &PriorityQueue{
		items:   make([]*QueueItem, 0),
		lookup:  make(map[string]*QueueItem),
		logger:  logger,
		config:  config,
		ctx:     ctx,
		cancel:  cancel,
		metrics: &QueueMetrics{
			PriorityDistribution: make(map[Priority]int64),
		},
	}

	// Initialize heap
	heap.Init(pq)

	// Start background processing
	pq.wg.Add(1)
	go pq.processQueue()

	return pq
}

// Implement heap.Interface
func (pq *PriorityQueue) Len() int {
	return len(pq.items)
}

func (pq *PriorityQueue) Less(i, j int) bool {
	// Higher priority first, then older timestamp (FIFO within priority)
	if pq.items[i].Priority == pq.items[j].Priority {
		return pq.items[i].Timestamp.Before(pq.items[j].Timestamp)
	}
	return pq.items[i].Priority > pq.items[j].Priority
}

func (pq *PriorityQueue) Swap(i, j int) {
	pq.items[i], pq.items[j] = pq.items[j], pq.items[i]
}

func (pq *PriorityQueue) Push(x interface{}) {
	item := x.(*QueueItem)
	pq.items = append(pq.items, item)
	pq.lookup[item.RequestID] = item
}

func (pq *PriorityQueue) Pop() interface{} {
	old := pq.items
	n := len(old)
	if n == 0 {
		return nil
	}
	item := old[n-1]
	pq.items = old[0 : n-1]
	delete(pq.lookup, item.RequestID)
	return item
}

// Enqueue adds a lock request to the queue
func (pq *PriorityQueue) Enqueue(ctx context.Context, project models.Project, workspace string,
	pull models.PullRequest, user models.User, priority Priority) (*QueueItem, error) {

	pq.mu.Lock()
	defer pq.mu.Unlock()

	// Check queue size limit
	if len(pq.items) >= pq.config.MaxQueueSize {
		return nil, fmt.Errorf("queue is full (max size: %d)", pq.config.MaxQueueSize)
	}

	// Generate unique request ID
	requestID := fmt.Sprintf("%s-%s-%d-%d",
		models.GenerateLockKey(project, workspace),
		user.Username,
		pull.Num,
		time.Now().UnixNano())

	// Check for duplicate requests
	if _, exists := pq.lookup[requestID]; exists {
		return nil, fmt.Errorf("duplicate request: %s", requestID)
	}

	// Create queue item
	item := &QueueItem{
		Project:        project,
		Workspace:      workspace,
		Pull:           pull,
		User:           user,
		Priority:       priority,
		Timestamp:      time.Now(),
		RequestID:      requestID,
		ResultChan:     make(chan QueueResult, 1),
		Ctx:            ctx,
		QueueEntryTime: time.Now(),
		RetryCount:     0,
	}

	// Add to heap
	heap.Push(pq, item)

	// Update metrics
	if pq.config.EnableMetrics {
		pq.updateMetrics(func(m *QueueMetrics) {
			m.TotalEnqueued++
			m.CurrentQueueSize = int64(len(pq.items))
			m.PriorityDistribution[priority]++
		})
	}

	pq.logger.Info("enqueued lock request: id=%s, priority=%s, project=%s, workspace=%s, position=%d",
		requestID, priority.String(), project.String(), workspace, len(pq.items))

	return item, nil
}

// Dequeue removes the highest priority item from the queue
func (pq *PriorityQueue) Dequeue() *QueueItem {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	if len(pq.items) == 0 {
		return nil
	}

	item := heap.Pop(pq).(*QueueItem)

	// Update metrics
	if pq.config.EnableMetrics {
		pq.updateMetrics(func(m *QueueMetrics) {
			m.CurrentQueueSize = int64(len(pq.items))
		})
	}

	return item
}

// Remove removes a specific item from the queue
func (pq *PriorityQueue) Remove(requestID string) bool {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	item, exists := pq.lookup[requestID]
	if !exists {
		return false
	}

	// Find and remove the item
	for i, queueItem := range pq.items {
		if queueItem.RequestID == requestID {
			heap.Remove(pq, i)
			break
		}
	}

	// Close the result channel
	close(item.ResultChan)

	// Update metrics
	if pq.config.EnableMetrics {
		pq.updateMetrics(func(m *QueueMetrics) {
			m.CurrentQueueSize = int64(len(pq.items))
		})
	}

	pq.logger.Info("removed queue item: id=%s", requestID)
	return true
}

// GetPosition returns the position of a request in the queue
func (pq *PriorityQueue) GetPosition(requestID string) (int, bool) {
	pq.mu.RLock()
	defer pq.mu.RUnlock()

	item, exists := pq.lookup[requestID]
	if !exists {
		return -1, false
	}

	// Count items with higher priority or same priority but earlier timestamp
	position := 1
	for _, other := range pq.items {
		if other.RequestID == requestID {
			continue
		}
		if other.Priority > item.Priority ||
		   (other.Priority == item.Priority && other.Timestamp.Before(item.Timestamp)) {
			position++
		}
	}

	return position, true
}

// Size returns the current queue size
func (pq *PriorityQueue) Size() int {
	pq.mu.RLock()
	defer pq.mu.RUnlock()
	return len(pq.items)
}

// IsEmpty returns true if the queue is empty
func (pq *PriorityQueue) IsEmpty() bool {
	pq.mu.RLock()
	defer pq.mu.RUnlock()
	return len(pq.items) == 0
}

// processQueue handles background processing of queued items
func (pq *PriorityQueue) processQueue() {
	defer pq.wg.Done()

	ticker := time.NewTicker(pq.config.ProcessingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-pq.ctx.Done():
			pq.logger.Info("stopping queue processor")
			return
		case <-ticker.C:
			pq.processBatch()
		}
	}
}

// processBatch processes a batch of queue items
func (pq *PriorityQueue) processBatch() {
	if pq.config.EnableMetrics {
		start := time.Now()
		defer func() {
			pq.updateMetrics(func(m *QueueMetrics) {
				m.BatchesProcessed++
				m.LastProcessingTime = start
				m.ProcessingDuration = time.Since(start)
			})
		}()
	}

	batchSize := pq.config.BatchProcessingSize
	if !pq.config.EnableBatching {
		batchSize = 1
	}

	processed := 0
	for processed < batchSize {
		item := pq.Dequeue()
		if item == nil {
			break // Queue is empty
		}

		// Check if context is cancelled
		select {
		case <-item.Ctx.Done():
			pq.sendResult(item, QueueResult{
				Success: false,
				Error:   item.Ctx.Err(),
			})
			processed++
			continue
		default:
		}

		// Check timeout
		if time.Since(item.QueueEntryTime) > pq.config.MaxWaitTime {
			pq.sendResult(item, QueueResult{
				Success: false,
				Error:   fmt.Errorf("queue timeout exceeded: %v", pq.config.MaxWaitTime),
			})
			if pq.config.EnableMetrics {
				pq.updateMetrics(func(m *QueueMetrics) {
					m.TotalTimeouts++
				})
			}
			processed++
			continue
		}

		// Process the item (this would integrate with the actual locking backend)
		result := pq.processItem(item)
		pq.sendResult(item, result)

		if pq.config.EnableMetrics {
			pq.updateMetrics(func(m *QueueMetrics) {
				m.TotalProcessed++
				if result.Success {
					// Update average wait time
					waitTime := time.Since(item.QueueEntryTime)
					if m.AverageWaitTime == 0 {
						m.AverageWaitTime = waitTime
					} else {
						m.AverageWaitTime = (m.AverageWaitTime + waitTime) / 2
					}
				} else {
					m.TotalErrors++
				}
			})
		}

		processed++
	}
}

// processItem processes a single queue item (placeholder for actual lock logic)
func (pq *PriorityQueue) processItem(item *QueueItem) QueueResult {
	// This is a placeholder - in the real implementation, this would
	// call the actual locking backend to try to acquire the lock

	position, _ := pq.GetPosition(item.RequestID)
	waitTime := time.Since(item.QueueEntryTime)

	return QueueResult{
		Success:      true, // Placeholder - would be actual lock result
		Lock:         nil,  // Placeholder - would be actual lock
		Error:        nil,
		Position:     position,
		WaitTime:     waitTime,
		TotalWaiters: pq.Size(),
	}
}

// sendResult sends a result to the item's result channel
func (pq *PriorityQueue) sendResult(item *QueueItem, result QueueResult) {
	select {
	case item.ResultChan <- result:
		close(item.ResultChan)
	case <-time.After(1 * time.Second):
		pq.logger.Warn("timeout sending result for request %s", item.RequestID)
		close(item.ResultChan)
	}
}

// updateMetrics safely updates queue metrics
func (pq *PriorityQueue) updateMetrics(updateFn func(*QueueMetrics)) {
	pq.metrics.mu.Lock()
	defer pq.metrics.mu.Unlock()
	updateFn(pq.metrics)
}

// GetMetrics returns a copy of the current metrics
func (pq *PriorityQueue) GetMetrics() QueueMetrics {
	pq.metrics.mu.RLock()
	defer pq.metrics.mu.RUnlock()

	// Create a copy to avoid race conditions
	metrics := QueueMetrics{
		TotalEnqueued:        pq.metrics.TotalEnqueued,
		TotalProcessed:       pq.metrics.TotalProcessed,
		TotalTimeouts:        pq.metrics.TotalTimeouts,
		TotalErrors:          pq.metrics.TotalErrors,
		CurrentQueueSize:     pq.metrics.CurrentQueueSize,
		AverageWaitTime:      pq.metrics.AverageWaitTime,
		PriorityDistribution: make(map[Priority]int64),
		BatchesProcessed:     pq.metrics.BatchesProcessed,
		LastProcessingTime:   pq.metrics.LastProcessingTime,
		ProcessingDuration:   pq.metrics.ProcessingDuration,
	}

	for k, v := range pq.metrics.PriorityDistribution {
		metrics.PriorityDistribution[k] = v
	}

	return metrics
}

// Shutdown gracefully shuts down the queue
func (pq *PriorityQueue) Shutdown() {
	pq.logger.Info("shutting down priority queue")

	// Cancel context to stop processing
	pq.cancel()

	// Wait for processing to complete
	pq.wg.Wait()

	// Notify all waiting items
	pq.mu.Lock()
	defer pq.mu.Unlock()

	for _, item := range pq.items {
		pq.sendResult(item, QueueResult{
			Success: false,
			Error:   fmt.Errorf("queue shutdown"),
		})
	}

	// Clear the queue
	pq.items = nil
	pq.lookup = make(map[string]*QueueItem)

	pq.logger.Info("priority queue shutdown complete")
}

// BatchProcessor manages batch processing of multiple queues
type BatchProcessor struct {
	queues     []*PriorityQueue
	config     BatchConfig
	logger     logging.SimpleLogging
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
}

// BatchConfig configures batch processing behavior
type BatchConfig struct {
	MaxBatchSize       int
	ProcessingInterval time.Duration
	LoadBalancing      bool
	RoundRobin         bool
}

// NewBatchProcessor creates a new batch processor
func NewBatchProcessor(queues []*PriorityQueue, config BatchConfig, logger logging.SimpleLogging) *BatchProcessor {
	ctx, cancel := context.WithCancel(context.Background())

	if config.MaxBatchSize == 0 {
		config.MaxBatchSize = 50
	}
	if config.ProcessingInterval == 0 {
		config.ProcessingInterval = 50 * time.Millisecond
	}

	bp := &BatchProcessor{
		queues: queues,
		config: config,
		logger: logger,
		ctx:    ctx,
		cancel: cancel,
	}

	// Start batch processing
	bp.wg.Add(1)
	go bp.processBatches()

	return bp
}

// processBatches handles batch processing across multiple queues
func (bp *BatchProcessor) processBatches() {
	defer bp.wg.Done()

	ticker := time.NewTicker(bp.config.ProcessingInterval)
	defer ticker.Stop()

	queueIndex := 0 // For round-robin processing

	for {
		select {
		case <-bp.ctx.Done():
			bp.logger.Info("stopping batch processor")
			return
		case <-ticker.C:
			if bp.config.RoundRobin {
				bp.processQueueRoundRobin(&queueIndex)
			} else if bp.config.LoadBalancing {
				bp.processQueueLoadBalanced()
			} else {
				bp.processAllQueues()
			}
		}
	}
}

// processQueueRoundRobin processes queues in round-robin fashion
func (bp *BatchProcessor) processQueueRoundRobin(index *int) {
	if len(bp.queues) == 0 {
		return
	}

	queue := bp.queues[*index]
	if !queue.IsEmpty() {
		queue.processBatch()
	}

	*index = (*index + 1) % len(bp.queues)
}

// processQueueLoadBalanced processes the queue with the most items
func (bp *BatchProcessor) processQueueLoadBalanced() {
	var largestQueue *PriorityQueue
	maxSize := 0

	for _, queue := range bp.queues {
		if size := queue.Size(); size > maxSize {
			maxSize = size
			largestQueue = queue
		}
	}

	if largestQueue != nil && maxSize > 0 {
		largestQueue.processBatch()
	}
}

// processAllQueues processes all queues simultaneously
func (bp *BatchProcessor) processAllQueues() {
	var batchWg sync.WaitGroup

	for _, queue := range bp.queues {
		if !queue.IsEmpty() {
			batchWg.Add(1)
			go func(q *PriorityQueue) {
				defer batchWg.Done()
				q.processBatch()
			}(queue)
		}
	}

	batchWg.Wait()
}

// Shutdown gracefully shuts down the batch processor
func (bp *BatchProcessor) Shutdown() {
	bp.logger.Info("shutting down batch processor")
	bp.cancel()
	bp.wg.Wait()
	bp.logger.Info("batch processor shutdown complete")
}