package queue

import (
	"container/heap"
	"context"
	"sync"
	"time"

	"github.com/runatlantis/atlantis/server/core/locking/types"
)

// PriorityQueue implements a priority-based lock request queue
type PriorityQueue struct {
	items    []*QueueItem
	mutex    sync.RWMutex
	maxSize  int
	notEmpty chan struct{}
}

// QueueItem represents an item in the priority queue
type QueueItem struct {
	Request   *types.EnhancedLockRequest
	Priority  types.Priority
	Timestamp time.Time
	Index     int // heap index
}

// NewPriorityQueue creates a new priority queue
func NewPriorityQueue(maxSize int) *PriorityQueue {
	return &PriorityQueue{
		items:    make([]*QueueItem, 0),
		maxSize:  maxSize,
		notEmpty: make(chan struct{}, 1),
	}
}

// Enqueue adds a request to the priority queue
func (pq *PriorityQueue) Enqueue(ctx context.Context, request *types.EnhancedLockRequest) error {
	pq.mutex.Lock()
	defer pq.mutex.Unlock()

	if len(pq.items) >= pq.maxSize {
		return &types.LockError{
			Type:    "QueueFull",
			Message: "priority queue is full",
			Code:    types.ErrCodeQueueFull,
		}
	}

	item := &QueueItem{
		Request:   request,
		Priority:  request.GetPriority(),
		Timestamp: time.Now(),
	}

	heap.Push(pq, item)

	// Signal that queue is not empty
	select {
	case pq.notEmpty <- struct{}{}:
	default:
	}

	return nil
}

// Dequeue removes and returns the highest priority request
func (pq *PriorityQueue) Dequeue(ctx context.Context) (*types.EnhancedLockRequest, error) {
	pq.mutex.Lock()
	defer pq.mutex.Unlock()

	if len(pq.items) == 0 {
		return nil, nil
	}

	item := heap.Pop(pq).(*QueueItem)
	return item.Request, nil
}

// DequeueWithTimeout waits for an item or timeout
func (pq *PriorityQueue) DequeueWithTimeout(ctx context.Context, timeout time.Duration) (*types.EnhancedLockRequest, error) {
	// Fast path: check if item is immediately available
	if item, err := pq.Dequeue(ctx); err != nil || item != nil {
		return item, err
	}

	// Wait for item or timeout
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-timer.C:
		return nil, types.NewTimeoutError(timeout)
	case <-pq.notEmpty:
		return pq.Dequeue(ctx)
	}
}

// Peek returns the highest priority request without removing it
func (pq *PriorityQueue) Peek() *types.EnhancedLockRequest {
	pq.mutex.RLock()
	defer pq.mutex.RUnlock()

	if len(pq.items) == 0 {
		return nil
	}

	return pq.items[0].Request
}

// Size returns the current queue size
func (pq *PriorityQueue) Size() int {
	pq.mutex.RLock()
	defer pq.mutex.RUnlock()
	return len(pq.items)
}

// IsEmpty checks if the queue is empty
func (pq *PriorityQueue) IsEmpty() bool {
	return pq.Size() == 0
}

// Clear removes all items from the queue
func (pq *PriorityQueue) Clear() {
	pq.mutex.Lock()
	defer pq.mutex.Unlock()

	pq.items = pq.items[:0]
}

// GetStats returns queue statistics
func (pq *PriorityQueue) GetStats() *QueueStats {
	pq.mutex.RLock()
	defer pq.mutex.RUnlock()

	stats := &QueueStats{
		Size:        len(pq.items),
		MaxSize:     pq.maxSize,
		ByPriority:  make(map[types.Priority]int),
		AverageWait: 0,
		OldestItem:  nil,
	}

	if len(pq.items) == 0 {
		return stats
	}

	var totalWait time.Duration
	var oldest *time.Time

	for _, item := range pq.items {
		stats.ByPriority[item.Priority]++

		wait := time.Since(item.Timestamp)
		totalWait += wait

		if oldest == nil || item.Timestamp.Before(*oldest) {
			oldest = &item.Timestamp
		}
	}

	stats.AverageWait = totalWait / time.Duration(len(pq.items))
	stats.OldestItem = oldest

	return stats
}

// Remove removes a specific request from the queue
func (pq *PriorityQueue) Remove(requestID string) bool {
	pq.mutex.Lock()
	defer pq.mutex.Unlock()

	for i, item := range pq.items {
		if item.Request.ID == requestID {
			heap.Remove(pq, i)
			return true
		}
	}

	return false
}

// QueueStats provides statistics about the queue
type QueueStats struct {
	Size        int                    `json:"size"`
	MaxSize     int                    `json:"max_size"`
	ByPriority  map[types.Priority]int `json:"by_priority"`
	AverageWait time.Duration          `json:"average_wait"`
	OldestItem  *time.Time             `json:"oldest_item,omitempty"`
}

// Heap interface implementation for priority queue

func (pq *PriorityQueue) Len() int {
	return len(pq.items)
}

func (pq *PriorityQueue) Less(i, j int) bool {
	// Higher priority comes first (Critical = 3, High = 2, Normal = 1, Low = 0)
	if pq.items[i].Priority != pq.items[j].Priority {
		return pq.items[i].Priority > pq.items[j].Priority
	}
	// If same priority, FIFO (earlier timestamp comes first)
	return pq.items[i].Timestamp.Before(pq.items[j].Timestamp)
}

func (pq *PriorityQueue) Swap(i, j int) {
	pq.items[i], pq.items[j] = pq.items[j], pq.items[i]
	pq.items[i].Index = i
	pq.items[j].Index = j
}

func (pq *PriorityQueue) Push(x interface{}) {
	item := x.(*QueueItem)
	item.Index = len(pq.items)
	pq.items = append(pq.items, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := pq.items
	n := len(old)
	item := old[n-1]
	item.Index = -1
	pq.items = old[0 : n-1]
	return item
}

// ResourceBasedQueue manages separate queues per resource to prevent head-of-line blocking
type ResourceBasedQueue struct {
	queues  map[string]*PriorityQueue
	mutex   sync.RWMutex
	maxSize int
}

// NewResourceBasedQueue creates a new resource-based queue system
func NewResourceBasedQueue(maxSizePerResource int) *ResourceBasedQueue {
	return &ResourceBasedQueue{
		queues:  make(map[string]*PriorityQueue),
		maxSize: maxSizePerResource,
	}
}

// Push adds a request to the appropriate resource queue
func (rbq *ResourceBasedQueue) Push(ctx context.Context, request *types.EnhancedLockRequest) error {
	resourceKey := rbq.getResourceKey(request.Resource)

	rbq.mutex.Lock()
	queue, exists := rbq.queues[resourceKey]
	if !exists {
		queue = NewPriorityQueue(rbq.maxSize)
		rbq.queues[resourceKey] = queue
	}
	rbq.mutex.Unlock()

	return queue.Enqueue(ctx, request)
}

// PopForResource removes the highest priority request for a specific resource
func (rbq *ResourceBasedQueue) PopForResource(ctx context.Context, resource types.ResourceIdentifier) (*types.EnhancedLockRequest, error) {
	resourceKey := rbq.getResourceKey(resource)

	rbq.mutex.RLock()
	queue, exists := rbq.queues[resourceKey]
	rbq.mutex.RUnlock()

	if !exists {
		return nil, nil
	}

	return queue.Dequeue(ctx)
}

// PopForResourceWithTimeout waits for a request for a specific resource
func (rbq *ResourceBasedQueue) PopForResourceWithTimeout(ctx context.Context, resource types.ResourceIdentifier, timeout time.Duration) (*types.EnhancedLockRequest, error) {
	resourceKey := rbq.getResourceKey(resource)

	rbq.mutex.RLock()
	queue, exists := rbq.queues[resourceKey]
	rbq.mutex.RUnlock()

	if !exists {
		// Wait for queue to be created or timeout
		startTime := time.Now()
		timer := time.NewTimer(timeout)
		defer timer.Stop()

		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-timer.C:
				return nil, types.NewTimeoutError(timeout)
			case <-ticker.C:
				rbq.mutex.RLock()
				queue, exists = rbq.queues[resourceKey]
				rbq.mutex.RUnlock()
				if exists {
					elapsed := time.Since(startTime)
					remaining := timeout - elapsed
					if remaining <= 0 {
						return nil, types.NewTimeoutError(timeout)
					}
					return queue.DequeueWithTimeout(ctx, remaining)
				}
			}
		}
	}

	return queue.DequeueWithTimeout(ctx, timeout)
}

// GetQueueForResource returns the queue for a specific resource
func (rbq *ResourceBasedQueue) GetQueueForResource(resource types.ResourceIdentifier) *PriorityQueue {
	resourceKey := rbq.getResourceKey(resource)

	rbq.mutex.RLock()
	defer rbq.mutex.RUnlock()

	return rbq.queues[resourceKey]
}

// GetAllStats returns statistics for all resource queues
func (rbq *ResourceBasedQueue) GetAllStats() map[string]*QueueStats {
	rbq.mutex.RLock()
	defer rbq.mutex.RUnlock()

	stats := make(map[string]*QueueStats)
	for resourceKey, queue := range rbq.queues {
		stats[resourceKey] = queue.GetStats()
	}

	return stats
}

// GetTotalSize returns the total size across all queues
func (rbq *ResourceBasedQueue) GetTotalSize() int {
	rbq.mutex.RLock()
	defer rbq.mutex.RUnlock()

	total := 0
	for _, queue := range rbq.queues {
		total += queue.Size()
	}

	return total
}

// CleanupEmptyQueues removes empty queues to prevent memory leaks
func (rbq *ResourceBasedQueue) CleanupEmptyQueues() {
	rbq.mutex.Lock()
	defer rbq.mutex.Unlock()

	for resourceKey, queue := range rbq.queues {
		if queue.IsEmpty() {
			delete(rbq.queues, resourceKey)
		}
	}
}

// Remove removes a specific request from all queues
func (rbq *ResourceBasedQueue) Remove(requestID string) bool {
	rbq.mutex.RLock()
	defer rbq.mutex.RUnlock()

	for _, queue := range rbq.queues {
		if queue.Remove(requestID) {
			return true
		}
	}

	return false
}

func (rbq *ResourceBasedQueue) getResourceKey(resource types.ResourceIdentifier) string {
	return resource.Namespace + "/" + resource.Name + "/" + resource.Workspace
}

// MemoryQueue is a simple FIFO queue implementation for basic scenarios
type MemoryQueue struct {
	items    []*types.EnhancedLockRequest
	mutex    sync.RWMutex
	maxSize  int
	notEmpty chan struct{}
}

// NewMemoryQueue creates a new simple memory queue
func NewMemoryQueue(maxSize int) *MemoryQueue {
	return &MemoryQueue{
		items:    make([]*types.EnhancedLockRequest, 0),
		maxSize:  maxSize,
		notEmpty: make(chan struct{}, 1),
	}
}

// Enqueue adds a request to the end of the queue
func (mq *MemoryQueue) Enqueue(ctx context.Context, request *types.EnhancedLockRequest) error {
	mq.mutex.Lock()
	defer mq.mutex.Unlock()

	if len(mq.items) >= mq.maxSize {
		return &types.LockError{
			Type:    "QueueFull",
			Message: "memory queue is full",
			Code:    types.ErrCodeQueueFull,
		}
	}

	mq.items = append(mq.items, request)

	// Signal that queue is not empty
	select {
	case mq.notEmpty <- struct{}{}:
	default:
	}

	return nil
}

// Dequeue removes and returns the first request (FIFO)
func (mq *MemoryQueue) Dequeue(ctx context.Context) (*types.EnhancedLockRequest, error) {
	mq.mutex.Lock()
	defer mq.mutex.Unlock()

	if len(mq.items) == 0 {
		return nil, nil
	}

	request := mq.items[0]
	mq.items = mq.items[1:]

	return request, nil
}

// DequeueWithTimeout waits for an item or timeout
func (mq *MemoryQueue) DequeueWithTimeout(ctx context.Context, timeout time.Duration) (*types.EnhancedLockRequest, error) {
	// Fast path: check if item is immediately available
	if item, err := mq.Dequeue(ctx); err != nil || item != nil {
		return item, err
	}

	// Wait for item or timeout
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-timer.C:
		return nil, types.NewTimeoutError(timeout)
	case <-mq.notEmpty:
		return mq.Dequeue(ctx)
	}
}

// Size returns the current queue size
func (mq *MemoryQueue) Size() int {
	mq.mutex.RLock()
	defer mq.mutex.RUnlock()
	return len(mq.items)
}

// IsEmpty checks if the queue is empty
func (mq *MemoryQueue) IsEmpty() bool {
	return mq.Size() == 0
}

// Clear removes all items from the queue
func (mq *MemoryQueue) Clear() {
	mq.mutex.Lock()
	defer mq.mutex.Unlock()

	mq.items = mq.items[:0]
}
