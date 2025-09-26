package backends

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/runatlantis/atlantis/server/core/locking/enhanced"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
)

// RedisBackend implements enhanced locking using Redis
type RedisBackend struct {
	client      redis.UniversalClient
	keyPrefix   string
	defaultTTL  time.Duration
	log         logging.SimpleLogging

	// Enhanced components
	scriptMgr     *ScriptManager
	healthMonitor *HealthMonitor
	clusterMgr    *ClusterManager

	// Metrics
	stats       *enhanced.BackendStats
	lastUpdated time.Time

	// Configuration
	config *enhanced.EnhancedConfig
}

// NewRedisBackend creates a new Redis backend for enhanced locking
func NewRedisBackend(client redis.UniversalClient, config *enhanced.EnhancedConfig, log logging.SimpleLogging) *RedisBackend {
	rb := &RedisBackend{
		client:      client,
		keyPrefix:   config.RedisKeyPrefix,
		defaultTTL:  config.RedisLockTTL,
		log:         log,
		config:      config,
		stats: &enhanced.BackendStats{
			HealthScore: 100,
			LastUpdated: time.Now(),
		},
		lastUpdated: time.Now(),
	}

	// Initialize enhanced components
	rb.scriptMgr = NewScriptManager(client)

	// Initialize health monitoring
	healthConfig := DefaultHealthConfig()
	rb.healthMonitor = NewHealthMonitor(client, healthConfig, log)

	// Initialize cluster management if enabled
	if config.RedisClusterMode {
		clusterConfig := DefaultClusterConfig()
		rb.clusterMgr = NewClusterManager(client, clusterConfig, log)
	}

	return rb
}

// AcquireLock attempts to acquire a lock with enhanced capabilities
func (r *RedisBackend) AcquireLock(ctx context.Context, request *enhanced.EnhancedLockRequest) (*enhanced.EnhancedLock, error) {
	lockID := r.generateLockID()
	lockKey := r.getLockKey(request.Resource)
	queueKey := r.getQueueKey(request.Resource)

	// Create the lock object
	lock := &enhanced.EnhancedLock{
		ID:         lockID,
		Resource:   request.Resource,
		State:      enhanced.LockStateAcquired,
		Priority:   request.Priority,
		Owner:      request.User.Username,
		AcquiredAt: time.Now(),
		Metadata:   request.Metadata,
		Version:    1,
	}

	// Set expiration if timeout is specified
	if request.Timeout > 0 {
		expiresAt := time.Now().Add(request.Timeout)
		lock.ExpiresAt = &expiresAt
	}

	// Create backward compatibility lock
	lock.OriginalLock = &models.ProjectLock{
		Project:   request.Project,
		Workspace: request.Workspace,
		User:      request.User,
		Time:      lock.AcquiredAt,
	}

	// Use enhanced script manager for atomic operations
	lockData, err := json.Marshal(lock)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal lock data: %w", err)
	}

	ttlSeconds := int64(0)
	if lock.ExpiresAt != nil {
		ttlSeconds = int64(lock.ExpiresAt.Sub(time.Now()).Seconds())
	}

	// Enhanced arguments for clustering support
	clusterMode := r.config.RedisClusterMode
	nodeID := ""
	if r.clusterMgr != nil {
		nodeID = r.clusterMgr.nodeID
	}

	result, err := r.scriptMgr.Execute(ctx, "acquire_lock",
		[]string{lockKey, queueKey, fmt.Sprintf("atlantis:node:%s", nodeID)},
		string(lockData), ttlSeconds, int(request.Priority),
		r.config.EnablePriorityQueue, clusterMode, nodeID, 1000)
	if err != nil {
		r.updateStats(false)
		return nil, fmt.Errorf("failed to execute lock acquisition script: %w", err)
	}

	resultSlice := result.([]interface{})
	acquired := resultSlice[0].(int64) == 1
	status := resultSlice[1].(string)

	if !acquired {
		if status == "queued" {
			lock.State = enhanced.LockStatePending
			r.log.Info("Lock request queued: %s", lockID)
			return lock, nil
		}
		r.updateStats(false)
		return nil, enhanced.NewLockExistsError(lockKey)
	}

	r.updateStats(true)
	r.log.Info("Lock acquired: %s for resource: %s", lockID, lockKey)
	return lock, nil
}

// TryAcquireLock attempts to acquire a lock without blocking
func (r *RedisBackend) TryAcquireLock(ctx context.Context, request *enhanced.EnhancedLockRequest) (*enhanced.EnhancedLock, bool, error) {
	lock, err := r.AcquireLock(ctx, request)
	if err != nil {
		if strings.Contains(err.Error(), "LOCK_EXISTS") {
			return nil, false, nil
		}
		return nil, false, err
	}

	acquired := lock.State == enhanced.LockStateAcquired
	return lock, acquired, nil
}

// ReleaseLock releases an acquired lock
func (r *RedisBackend) ReleaseLock(ctx context.Context, lockID string) error {
	// First find the lock to get its key
	lock, err := r.GetLock(ctx, lockID)
	if err != nil {
		return err
	}

	lockKey := r.getLockKey(lock.Resource)
	queueKey := r.getQueueKey(lock.Resource)

	// Lua script for atomic lock release and queue processing
	luaScript := `
		local lockKey = KEYS[1]
		local queueKey = KEYS[2]
		local lockID = ARGV[1]

		-- Get current lock
		local currentLock = redis.call('GET', lockKey)
		if not currentLock then
			return {false, "not_found"}
		end

		-- Parse lock data to verify ownership
		local lockData = cjson.decode(currentLock)
		if lockData.id ~= lockID then
			return {false, "not_owner"}
		end

		-- Release the lock
		redis.call('DEL', lockKey)

		-- Check if anyone is queued
		local queued = redis.call('ZRANGE', queueKey, 0, 0)
		if #queued > 0 then
			-- Get next in queue
			local nextLockData = queued[1]
			redis.call('ZREM', queueKey, nextLockData)

			-- Parse and update the queued lock
			local nextLock = cjson.decode(nextLockData)
			nextLock.state = "acquired"
			nextLock.acquired_at = redis.call('TIME')[1]

			-- Acquire lock for queued request
			redis.call('SET', lockKey, cjson.encode(nextLock))

			-- Publish lock transferred event
			redis.call('PUBLISH', 'atlantis:lock:transferred', lockKey)

			return {true, "transferred"}
		end

		-- Publish lock released event
		redis.call('PUBLISH', 'atlantis:lock:released', lockKey)

		return {true, "released"}
	`

	result, err := r.client.Eval(ctx, luaScript, []string{lockKey, queueKey}, lockID).Result()
	if err != nil {
		return fmt.Errorf("failed to execute lock release script: %w", err)
	}

	resultSlice := result.([]interface{})
	success := resultSlice[0].(int64) == 1
	status := resultSlice[1].(string)

	if !success {
		if status == "not_found" {
			return enhanced.NewLockNotFoundError(lockID)
		}
		return fmt.Errorf("failed to release lock: %s", status)
	}

	r.log.Info("Lock released: %s, status: %s", lockID, status)
	return nil
}

// GetLock retrieves a lock by ID
func (r *RedisBackend) GetLock(ctx context.Context, lockID string) (*enhanced.EnhancedLock, error) {
	// Since we don't store locks by ID directly, we need to scan all locks
	// This is not ideal for performance, but maintains compatibility
	locks, err := r.ListLocks(ctx)
	if err != nil {
		return nil, err
	}

	for _, lock := range locks {
		if lock.ID == lockID {
			return lock, nil
		}
	}

	return nil, enhanced.NewLockNotFoundError(lockID)
}

// ListLocks returns all active locks
func (r *RedisBackend) ListLocks(ctx context.Context) ([]*enhanced.EnhancedLock, error) {
	pattern := r.keyPrefix + "*"
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to scan lock keys: %w", err)
	}

	var locks []*enhanced.EnhancedLock
	for _, key := range keys {
		// Skip queue keys
		if strings.Contains(key, ":queue:") {
			continue
		}

		lockData, err := r.client.Get(ctx, key).Result()
		if err != nil {
			if err == redis.Nil {
				continue // Lock was deleted between scan and get
			}
			r.log.Warn("Failed to get lock data for key %s: %v", key, err)
			continue
		}

		var lock enhanced.EnhancedLock
		if err := json.Unmarshal([]byte(lockData), &lock); err != nil {
			r.log.Warn("Failed to unmarshal lock data for key %s: %v", key, err)
			continue
		}

		// Check if lock is expired
		if lock.ExpiresAt != nil && time.Now().After(*lock.ExpiresAt) {
			lock.State = enhanced.LockStateExpired
		}

		locks = append(locks, &lock)
	}

	return locks, nil
}

// RefreshLock extends the expiration time of a lock
func (r *RedisBackend) RefreshLock(ctx context.Context, lockID string, extension time.Duration) error {
	lock, err := r.GetLock(ctx, lockID)
	if err != nil {
		return err
	}

	lockKey := r.getLockKey(lock.Resource)

	// Update expiration time
	if lock.ExpiresAt != nil {
		newExpiration := lock.ExpiresAt.Add(extension)
		lock.ExpiresAt = &newExpiration
	} else {
		newExpiration := time.Now().Add(extension)
		lock.ExpiresAt = &newExpiration
	}

	lockData, err := json.Marshal(lock)
	if err != nil {
		return fmt.Errorf("failed to marshal updated lock: %w", err)
	}

	ttl := int64(lock.ExpiresAt.Sub(time.Now()).Seconds())
	err = r.client.SetEx(ctx, lockKey, string(lockData), time.Duration(ttl)*time.Second).Err()
	if err != nil {
		return fmt.Errorf("failed to refresh lock: %w", err)
	}

	return nil
}

// TransferLock transfers ownership of a lock to another user
func (r *RedisBackend) TransferLock(ctx context.Context, lockID string, newOwner string) error {
	lock, err := r.GetLock(ctx, lockID)
	if err != nil {
		return err
	}

	lockKey := r.getLockKey(lock.Resource)

	// Update owner
	lock.Owner = newOwner
	lock.Version++

	lockData, err := json.Marshal(lock)
	if err != nil {
		return fmt.Errorf("failed to marshal updated lock: %w", err)
	}

	// Update the lock with new owner
	var ttl time.Duration
	if lock.ExpiresAt != nil {
		ttl = lock.ExpiresAt.Sub(time.Now())
	}

	if ttl > 0 {
		err = r.client.SetEx(ctx, lockKey, string(lockData), ttl).Err()
	} else {
		err = r.client.Set(ctx, lockKey, string(lockData), 0).Err()
	}

	if err != nil {
		return fmt.Errorf("failed to transfer lock: %w", err)
	}

	return nil
}

// EnqueueLockRequest adds a lock request to the priority queue
func (r *RedisBackend) EnqueueLockRequest(ctx context.Context, request *enhanced.EnhancedLockRequest) error {
	if !r.config.EnablePriorityQueue {
		return fmt.Errorf("priority queue is not enabled")
	}

	queueKey := r.getQueueKey(request.Resource)

	requestData, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal queue request: %w", err)
	}

	// Priority score: higher priority = lower score (for min priority queue)
	// Add timestamp to ensure FIFO within same priority
	score := float64((4-int(request.Priority))*1000000) + float64(time.Now().Unix())

	err = r.client.ZAdd(ctx, queueKey, redis.Z{
		Score:  score,
		Member: string(requestData),
	}).Err()

	if err != nil {
		return fmt.Errorf("failed to enqueue lock request: %w", err)
	}

	return nil
}

// DequeueNextRequest gets the next highest priority request from the queue
func (r *RedisBackend) DequeueNextRequest(ctx context.Context) (*enhanced.EnhancedLockRequest, error) {
	// This is a simplified implementation - in practice you'd specify which resource queue
	// For now, we'll scan all queues
	pattern := r.keyPrefix + "*:queue"
	queueKeys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to scan queue keys: %w", err)
	}

	for _, queueKey := range queueKeys {
		// Get highest priority item (lowest score)
		items, err := r.client.ZRangeWithScores(ctx, queueKey, 0, 0).Result()
		if err != nil || len(items) == 0 {
			continue
		}

		// Remove from queue
		_, err = r.client.ZRem(ctx, queueKey, items[0].Member).Result()
		if err != nil {
			continue
		}

		// Parse request
		var request enhanced.EnhancedLockRequest
		if err := json.Unmarshal([]byte(items[0].Member.(string)), &request); err != nil {
			continue
		}

		return &request, nil
	}

	return nil, nil // No requests in any queue
}

// GetQueueStatus returns the current status of queues
func (r *RedisBackend) GetQueueStatus(ctx context.Context) (*enhanced.QueueStatus, error) {
	pattern := r.keyPrefix + "*:queue"
	queueKeys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to scan queue keys: %w", err)
	}

	status := &enhanced.QueueStatus{
		QueuesByPriority: make(map[enhanced.Priority]int),
	}

	var allRequests []*enhanced.EnhancedLockRequest
	var oldestTime *time.Time

	for _, queueKey := range queueKeys {
		items, err := r.client.ZRangeWithScores(ctx, queueKey, 0, -1).Result()
		if err != nil {
			continue
		}

		status.Size += len(items)

		for _, item := range items {
			var request enhanced.EnhancedLockRequest
			if err := json.Unmarshal([]byte(item.Member.(string)), &request); err != nil {
				continue
			}

			allRequests = append(allRequests, &request)
			status.QueuesByPriority[request.Priority]++

			if oldestTime == nil || request.RequestedAt.Before(*oldestTime) {
				oldestTime = &request.RequestedAt
			}
		}
	}

	status.PendingRequests = allRequests
	status.OldestRequest = oldestTime

	return status, nil
}

// HealthCheck verifies the backend is operational
func (r *RedisBackend) HealthCheck(ctx context.Context) error {
	_, err := r.client.Ping(ctx).Result()
	if err != nil {
		r.stats.HealthScore = 0
		return fmt.Errorf("redis health check failed: %w", err)
	}

	r.stats.HealthScore = 100
	r.stats.LastUpdated = time.Now()
	return nil
}

// GetStats returns backend performance statistics
func (r *RedisBackend) GetStats(ctx context.Context) (*enhanced.BackendStats, error) {
	r.stats.LastUpdated = time.Now()

	// Update active locks count
	locks, err := r.ListLocks(ctx)
	if err == nil {
		r.stats.ActiveLocks = int64(len(locks))
	}

	// Update queue depth
	queueStatus, err := r.GetQueueStatus(ctx)
	if err == nil {
		r.stats.QueueDepth = queueStatus.Size
	}

	return r.stats, nil
}

// Subscribe creates a subscription to lock events
func (r *RedisBackend) Subscribe(ctx context.Context, eventTypes []string) (<-chan *enhanced.LockEvent, error) {
	// Subscribe to Redis pub/sub channels
	var channels []string
	for _, eventType := range eventTypes {
		channels = append(channels, fmt.Sprintf("atlantis:lock:%s", eventType))
	}

	if len(channels) == 0 {
		channels = []string{"atlantis:lock:*"}
	}

	pubsub := r.client.PSubscribe(ctx, channels...)
	eventChan := make(chan *enhanced.LockEvent, r.config.EventBufferSize)

	go func() {
		defer close(eventChan)
		defer pubsub.Close()

		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-pubsub.Channel():
				// Parse the event from the message
				parts := strings.Split(msg.Channel, ":")
				if len(parts) < 3 {
					continue
				}

				eventType := parts[2]
				event := &enhanced.LockEvent{
					Type:      eventType,
					LockID:    msg.Payload, // This would need to be parsed properly
					Timestamp: time.Now(),
				}

				select {
				case eventChan <- event:
				case <-ctx.Done():
					return
				default:
					// Channel full, drop event
				}
			}
		}
	}()

	return eventChan, nil
}

// CleanupExpiredLocks removes expired locks
func (r *RedisBackend) CleanupExpiredLocks(ctx context.Context) (int, error) {
	locks, err := r.ListLocks(ctx)
	if err != nil {
		return 0, err
	}

	cleaned := 0
	for _, lock := range locks {
		if lock.ExpiresAt != nil && time.Now().After(*lock.ExpiresAt) {
			if err := r.ReleaseLock(ctx, lock.ID); err != nil {
				r.log.Warn("Failed to cleanup expired lock %s: %v", lock.ID, err)
				continue
			}
			cleaned++
		}
	}

	return cleaned, nil
}

// Backward compatibility methods

// GetLegacyLock converts an enhanced lock to legacy format
func (r *RedisBackend) GetLegacyLock(project models.Project, workspace string) (*models.ProjectLock, error) {
	resource := enhanced.ResourceIdentifier{
		Type:      enhanced.ResourceTypeProject,
		Namespace: project.RepoFullName,
		Name:      project.Path,
		Workspace: workspace,
		Path:      project.Path,
	}

	lockKey := r.getLockKey(resource)

	lockData, err := r.client.Get(context.Background(), lockKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // No lock exists
		}
		return nil, fmt.Errorf("failed to get legacy lock: %w", err)
	}

	var lock enhanced.EnhancedLock
	if err := json.Unmarshal([]byte(lockData), &lock); err != nil {
		return nil, fmt.Errorf("failed to unmarshal lock data: %w", err)
	}

	return lock.OriginalLock, nil
}

// ConvertToLegacy converts an enhanced lock to legacy format
func (r *RedisBackend) ConvertToLegacy(lock *enhanced.EnhancedLock) *models.ProjectLock {
	if lock.OriginalLock != nil {
		return lock.OriginalLock
	}

	// Create a legacy lock from enhanced lock data
	return &models.ProjectLock{
		Project: models.Project{
			RepoFullName: lock.Resource.Namespace,
			Path:         lock.Resource.Path,
		},
		Workspace: lock.Resource.Workspace,
		User:      models.User{Username: lock.Owner},
		Time:      lock.AcquiredAt,
	}
}

// ConvertFromLegacy converts a legacy lock to enhanced format
func (r *RedisBackend) ConvertFromLegacy(legacyLock *models.ProjectLock) *enhanced.EnhancedLock {
	return &enhanced.EnhancedLock{
		ID: r.generateLockID(),
		Resource: enhanced.ResourceIdentifier{
			Type:      enhanced.ResourceTypeProject,
			Namespace: legacyLock.Project.RepoFullName,
			Name:      legacyLock.Project.Path,
			Workspace: legacyLock.Workspace,
			Path:      legacyLock.Project.Path,
		},
		State:        enhanced.LockStateAcquired,
		Priority:     enhanced.PriorityNormal,
		Owner:        legacyLock.User.Username,
		AcquiredAt:   legacyLock.Time,
		Version:      1,
		OriginalLock: legacyLock,
	}
}

// Helper methods

func (r *RedisBackend) getLockKey(resource enhanced.ResourceIdentifier) string {
	return fmt.Sprintf("%slock:%s:%s:%s", r.keyPrefix, resource.Namespace, resource.Path, resource.Workspace)
}

func (r *RedisBackend) getQueueKey(resource enhanced.ResourceIdentifier) string {
	return fmt.Sprintf("%squeue:%s:%s:%s", r.keyPrefix, resource.Namespace, resource.Path, resource.Workspace)
}

func (r *RedisBackend) generateLockID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func (r *RedisBackend) updateStats(success bool) {
	r.stats.TotalRequests++
	if success {
		r.stats.SuccessfulAcquires++
	} else {
		r.stats.FailedAcquires++
	}
	r.stats.LastUpdated = time.Now()
}