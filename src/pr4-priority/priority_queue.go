// PR #4: Priority Queue Implementation - Queue Management and Processing
// This file implements the priority queue system for enhanced locking

package queue

import (
	"container/heap"
	"context"
	"sync"
	"time"

	"github.com/runatlantis/atlantis/server/models"
	"go.uber.org/zap"
)

// PriorityQueue manages lock acquisition requests with priority ordering
type PriorityQueue struct {
	mu sync.RWMutex
	
	// Queue data structures
	queues map[string]*ProjectQueue // keyed by project+workspace
	globalQueue *RequestHeap
	
	// Processing configuration
	config QueueConfig
	processor *QueueProcessor
	
	// Monitoring and metrics
	metrics *QueueMetrics
	logger *zap.Logger
	
	// Control channels
	stopCh chan struct{}
	done chan struct{}
}

// QueueConfig controls queue behavior
type QueueConfig struct {
	// Processing settings
	BatchSize int
	ProcessingInterval time.Duration
	MaxQueueSize int
	
	// Timeout settings
	DefaultTimeout time.Duration
	MaxTimeout time.Duration
	
	// Priority settings
	EnablePriority bool
	DefaultPriority int
	MaxPriority int
	
	// Fairness settings
	EnableFairness bool
	FairnessWindow time.Duration
}

// LockRequest represents a queued lock acquisition request
type LockRequest struct {
	ID string
	Lock models.ProjectLock
	Priority int
	Timeout time.Duration
	SubmittedAt time.Time
	Context context.Context
	ResponseCh chan LockResponse
	
	// Queue management
	Index int // heap index
	RetryCount int
	LastAttempt time.Time
}

// LockResponse contains the result of a lock acquisition attempt
type LockResponse struct {
	Acquired bool
	Error error
	WaitTime time.Duration
	QueuePosition int
}

// ProjectQueue manages requests for a specific project/workspace
type ProjectQueue struct {
	mu sync.RWMutex
	ProjectKey string
	Requests []*LockRequest
	CurrentHolder *LockRequest
	ProcessingStarted time.Time
}

// RequestHeap implements a priority heap for lock requests
type RequestHeap []*LockRequest

// QueueProcessor handles the actual lock acquisition processing
type QueueProcessor struct {
	backend EnhancedLockingBackend
	queues *PriorityQueue
	config QueueConfig
	logger *zap.Logger
}

// QueueMetrics tracks queue performance
type QueueMetrics struct {
	mu sync.RWMutex
	
	// Request statistics
	TotalRequests int64
	ProcessedRequests int64
	FailedRequests int64
	TimeoutRequests int64
	
	// Timing statistics
	AverageWaitTime time.Duration
	MaxWaitTime time.Duration
	AverageProcessingTime time.Duration
	
	// Queue statistics
	CurrentQueueDepth int
	MaxQueueDepth int
	ActiveQueues int
}

// EnhancedLockingBackend interface for queue integration
type EnhancedLockingBackend interface {
	TryLock(lock models.ProjectLock) (bool, error)
	Unlock(lock models.ProjectLock) error
	IsLocked(project models.Project, workspace string) (bool, error)
}

// NewPriorityQueue creates a new priority queue
func NewPriorityQueue(config QueueConfig, backend EnhancedLockingBackend, logger *zap.Logger) *PriorityQueue {
	pq := &PriorityQueue{
		queues: make(map[string]*ProjectQueue),
		globalQueue: &RequestHeap{},
		config: config,
		metrics: &QueueMetrics{},
		logger: logger,
		stopCh: make(chan struct{}),
		done: make(chan struct{}),
	}
	
	heap.Init(pq.globalQueue)
	
	pq.processor = &QueueProcessor{
		backend: backend,
		queues: pq,
		config: config,
		logger: logger,
	}
	
	return pq
}

// Start begins queue processing
func (pq *PriorityQueue) Start(ctx context.Context) error {
	go pq.processQueue(ctx)
	go pq.collectMetrics(ctx)
	return nil
}

// Stop gracefully stops queue processing
func (pq *PriorityQueue) Stop() error {
	close(pq.stopCh)
	<-pq.done
	return nil
}

// EnqueueLockRequest adds a lock request to the queue
func (pq *PriorityQueue) EnqueueLockRequest(ctx context.Context, lock models.ProjectLock, priority int, timeout time.Duration) (*LockRequest, error) {
	pq.mu.Lock()
	defer pq.mu.Unlock()
	
	// Check queue size limits
	if len(*pq.globalQueue) >= pq.config.MaxQueueSize {
		return nil, ErrQueueFull
	}
	
	// Create lock request
	request := &LockRequest{
		ID: generateRequestID(),
		Lock: lock,
		Priority: priority,
		Timeout: timeout,
		SubmittedAt: time.Now(),
		Context: ctx,
		ResponseCh: make(chan LockResponse, 1),
	}
	
	// Add to global queue
	heap.Push(pq.globalQueue, request)
	
	// Add to project-specific queue
	projectKey := generateProjectKey(lock.Project, lock.Workspace)
	if pq.queues[projectKey] == nil {
		pq.queues[projectKey] = &ProjectQueue{
			ProjectKey: projectKey,
			Requests: make([]*LockRequest, 0),
		}
	}
	pq.queues[projectKey].addRequest(request)
	
	// Update metrics
	pq.metrics.mu.Lock()
	pq.metrics.TotalRequests++
	pq.metrics.CurrentQueueDepth = len(*pq.globalQueue)
	if pq.metrics.CurrentQueueDepth > pq.metrics.MaxQueueDepth {
		pq.metrics.MaxQueueDepth = pq.metrics.CurrentQueueDepth
	}
	pq.metrics.mu.Unlock()
	
	pq.logger.Debug("Lock request enqueued",
		zap.String("request_id", request.ID),
		zap.String("project", projectKey),
		zap.Int("priority", priority),
		zap.Int("queue_depth", len(*pq.globalQueue)))
	
	return request, nil
}

// GetQueueStatus returns status for a specific project/workspace
func (pq *PriorityQueue) GetQueueStatus(project models.Project, workspace string) *QueueStatus {
	pq.mu.RLock()
	defer pq.mu.RUnlock()
	
	projectKey := generateProjectKey(project, workspace)
	projQueue := pq.queues[projectKey]
	
	if projQueue == nil {
		return &QueueStatus{
			QueueDepth: 0,
			EstimatedWait: 0,
			CurrentPosition: 0,
		}
	}
	
	return &QueueStatus{
		QueueDepth: len(projQueue.Requests),
		EstimatedWait: pq.estimateWaitTime(projQueue),
		CurrentPosition: pq.getQueuePosition(projectKey),
		AverageProcessingTime: pq.metrics.AverageProcessingTime,
	}
}

// processQueue is the main processing loop
func (pq *PriorityQueue) processQueue(ctx context.Context) {
	defer close(pq.done)
	
	ticker := time.NewTicker(pq.config.ProcessingInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-pq.stopCh:
			return
		case <-ticker.C:
			pq.processBatch()
		}
	}
}

// processBatch processes a batch of requests
func (pq *PriorityQueue) processBatch() {
	pq.mu.Lock()
	batch := pq.getNextBatch()
	pq.mu.Unlock()
	
	if len(batch) == 0 {
		return
	}
	
	pq.logger.Debug("Processing batch", zap.Int("batch_size", len(batch)))
	
	// Process requests concurrently within the batch
	var wg sync.WaitGroup
	for _, request := range batch {
		wg.Add(1)
		go func(req *LockRequest) {
			defer wg.Done()
			pq.processRequest(req)
		}(request)
	}
	wg.Wait()
}

// getNextBatch selects requests for processing
func (pq *PriorityQueue) getNextBatch() []*LockRequest {
	var batch []*LockRequest
	batchSize := pq.config.BatchSize
	
	for len(batch) < batchSize && pq.globalQueue.Len() > 0 {
		request := heap.Pop(pq.globalQueue).(*LockRequest)
		
		// Check if request is still valid
		if pq.isRequestValid(request) {
			batch = append(batch, request)
		} else {
			// Clean up invalid request
			pq.cleanupRequest(request)
		}
	}
	
	return batch
}

// processRequest handles individual request processing
func (pq *PriorityQueue) processRequest(request *LockRequest) {
	start := time.Now()
	defer func() {
		processingTime := time.Since(start)
		pq.updateProcessingMetrics(processingTime)
	}()
	
	// Check if request has timed out
	if time.Since(request.SubmittedAt) > request.Timeout {
		pq.sendTimeoutResponse(request)
		return
	}
	
	// Check if context is cancelled
	select {
	case <-request.Context.Done():
		pq.sendCancelledResponse(request)
		return
	default:
	}
	
	// Attempt lock acquisition
	acquired, err := pq.processor.backend.TryLock(request.Lock)
	
	// Calculate wait time
	waitTime := time.Since(request.SubmittedAt)
	
	// Send response
	response := LockResponse{
		Acquired: acquired,
		Error: err,
		WaitTime: waitTime,
		QueuePosition: pq.getRequestPosition(request),
	}
	
	select {
	case request.ResponseCh <- response:
		// Response sent successfully
	case <-time.After(1 * time.Second):
		// Response channel blocked or closed
		pq.logger.Warn("Failed to send response", zap.String("request_id", request.ID))
	}
	
	// Update metrics
	pq.updateRequestMetrics(request, response)
	
	// Clean up request
	pq.cleanupRequest(request)
	
	pq.logger.Debug("Request processed",
		zap.String("request_id", request.ID),
		zap.Bool("acquired", acquired),
		zap.Duration("wait_time", waitTime),
		zap.Error(err))
}

// Helper methods

func (pq *PriorityQueue) isRequestValid(request *LockRequest) bool {
	// Check if context is still valid
	select {
	case <-request.Context.Done():
		return false
	default:
	}
	
	// Check if request has timed out
	return time.Since(request.SubmittedAt) <= request.Timeout
}

func (pq *PriorityQueue) cleanupRequest(request *LockRequest) {
	// Remove from project queue
	projectKey := generateProjectKey(request.Lock.Project, request.Lock.Workspace)
	if projQueue := pq.queues[projectKey]; projQueue != nil {
		projQueue.removeRequest(request)
		
		// Clean up empty project queue
		if len(projQueue.Requests) == 0 {
			delete(pq.queues, projectKey)
		}
	}
	
	// Update metrics
	pq.metrics.mu.Lock()
	pq.metrics.CurrentQueueDepth = len(*pq.globalQueue)
	pq.metrics.mu.Unlock()
}

func (pq *PriorityQueue) sendTimeoutResponse(request *LockRequest) {
	response := LockResponse{
		Acquired: false,
		Error: ErrRequestTimeout,
		WaitTime: time.Since(request.SubmittedAt),
		QueuePosition: -1,
	}
	
	select {
	case request.ResponseCh <- response:
	case <-time.After(1 * time.Second):
	}
	
	pq.metrics.mu.Lock()
	pq.metrics.TimeoutRequests++
	pq.metrics.mu.Unlock()
}

func (pq *PriorityQueue) sendCancelledResponse(request *LockRequest) {
	response := LockResponse{
		Acquired: false,
		Error: request.Context.Err(),
		WaitTime: time.Since(request.SubmittedAt),
		QueuePosition: -1,
	}
	
	select {
	case request.ResponseCh <- response:
	case <-time.After(1 * time.Second):
	}
}

func (pq *PriorityQueue) updateRequestMetrics(request *LockRequest, response LockResponse) {
	pq.metrics.mu.Lock()
	defer pq.metrics.mu.Unlock()
	
	pq.metrics.ProcessedRequests++
	if response.Error != nil {
		pq.metrics.FailedRequests++
	}
	
	// Update wait time statistics
	if response.WaitTime > pq.metrics.MaxWaitTime {
		pq.metrics.MaxWaitTime = response.WaitTime
	}
	
	// Calculate rolling average (simplified)
	totalProcessed := float64(pq.metrics.ProcessedRequests)
	pq.metrics.AverageWaitTime = time.Duration(
		(float64(pq.metrics.AverageWaitTime)*totalProcessed + float64(response.WaitTime)) / (totalProcessed + 1),
	)
}

func (pq *PriorityQueue) updateProcessingMetrics(processingTime time.Duration) {
	pq.metrics.mu.Lock()
	defer pq.metrics.mu.Unlock()
	
	// Update processing time average (simplified)
	totalProcessed := float64(pq.metrics.ProcessedRequests)
	pq.metrics.AverageProcessingTime = time.Duration(
		(float64(pq.metrics.AverageProcessingTime)*totalProcessed + float64(processingTime)) / (totalProcessed + 1),
	)
}

// Heap interface implementation for RequestHeap

func (h RequestHeap) Len() int { return len(h) }

func (h RequestHeap) Less(i, j int) bool {
	// Higher priority first, then earlier submission time
	if h[i].Priority == h[j].Priority {
		return h[i].SubmittedAt.Before(h[j].SubmittedAt)
	}
	return h[i].Priority > h[j].Priority
}

func (h RequestHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].Index = i
	h[j].Index = j
}

func (h *RequestHeap) Push(x interface{}) {
	n := len(*h)
	request := x.(*LockRequest)
	request.Index = n
	*h = append(*h, request)
}

func (h *RequestHeap) Pop() interface{} {
	old := *h
	n := len(old)
	request := old[n-1]
	request.Index = -1
	*h = old[0 : n-1]
	return request
}

// Project queue methods

func (pq *ProjectQueue) addRequest(request *LockRequest) {
	pq.mu.Lock()
	defer pq.mu.Unlock()
	pq.Requests = append(pq.Requests, request)
}

func (pq *ProjectQueue) removeRequest(request *LockRequest) {
	pq.mu.Lock()
	defer pq.mu.Unlock()
	
	for i, req := range pq.Requests {
		if req.ID == request.ID {
			pq.Requests = append(pq.Requests[:i], pq.Requests[i+1:]...)
			break
		}
	}
}

// Utility functions

func generateRequestID() string {
	return fmt.Sprintf("req-%d-%d", time.Now().UnixNano(), rand.Int63())
}

func generateProjectKey(project models.Project, workspace string) string {
	return fmt.Sprintf("%s/%s/%s", project.RepoFullName, project.Path, workspace)
}

// Error definitions

var (
	ErrQueueFull = errors.New("queue is full")
	ErrRequestTimeout = errors.New("request timed out")
)

// Additional interfaces and types would be defined here...
