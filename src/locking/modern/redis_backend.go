package modern

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/runatlantis/atlantis/server/core/locking/enhanced"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
)

// ModernRedisBackend implements enhanced locking with modern Redis features
type ModernRedisBackend struct {
	client      redis.UniversalClient
	config      *ModernLockingConfig
	log         logging.SimpleLogging

	// Clustering and HA
	isClusteredMode bool
	isSentinelMode  bool

	// Metrics and monitoring
	metrics        *BackendMetrics
	healthChecker  *HealthChecker

	// Connection pooling and pipelining
	pipeline       redis.Pipeliner
	pipelineMutex  sync.Mutex

	// Lua scripts for atomic operations
	luaScripts     map[string]*redis.Script

	// Event handling
	eventPublisher *EventPublisher
}

// BackendMetrics tracks Redis backend performance
type BackendMetrics struct {
	// Connection metrics
	ActiveConnections   int64         `json:"active_connections"`
	IdleConnections     int64         `json:"idle_connections"`
	ConnectionErrors    int64         `json:"connection_errors"`

	// Operation metrics
	LockOperations      int64         `json:"lock_operations"`
	UnlockOperations    int64         `json:"unlock_operations"`
	SuccessfulOps       int64         `json:"successful_ops"`
	FailedOps          int64         `json:"failed_ops"`
	AverageLatency      time.Duration `json:"average_latency"`

	// Redis-specific metrics
	MemoryUsage        int64         `json:"memory_usage"`
	KeyspaceHits       int64         `json:"keyspace_hits"`
	KeyspaceMisses     int64         `json:"keyspace_misses"`
	ReplicationLag     time.Duration `json:"replication_lag"`

	// Clustering metrics
	ClusterNodes       int           `json:"cluster_nodes"`
	ClusterSlotsCovered int          `json:"cluster_slots_covered"`
	ClusterState       string        `json:"cluster_state"`

	LastUpdated        time.Time     `json:"last_updated"`
}

// HealthChecker monitors Redis health
type HealthChecker struct {
	backend       *ModernRedisBackend
	stopChan      chan struct{}
	healthStatus  *HealthStatus
	mutex         sync.RWMutex
}

// HealthStatus represents current health status
type HealthStatus struct {
	IsHealthy     bool          `json:"is_healthy"`
	LastCheck     time.Time     `json:"last_check"`
	CheckDuration time.Duration `json:"check_duration"`
	Errors        []string      `json:"errors,omitempty"`
	Details       map[string]interface{} `json:"details,omitempty"`
}

// EventPublisher handles event publishing with improved reliability
type EventPublisher struct {
	client       redis.UniversalClient
	config       *ModernLockingConfig
	eventBuffer  chan *enhanced.LockEvent
	stopChan     chan struct{}
	wg           sync.WaitGroup
}

// NewModernRedisBackend creates a new modern Redis backend
func NewModernRedisBackend(config *ModernLockingConfig, log logging.SimpleLogging) (*ModernRedisBackend, error) {
	if config == nil {
		config = DefaultModernConfig()
	}

	// Create Redis client based on configuration
	client, err := createRedisClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Redis client: %w", err)
	}

	backend := &ModernRedisBackend{
		client:         client,
		config:         config,
		log:            log,
		isClusteredMode: config.Redis.EnableClustering,
		isSentinelMode:  config.Redis.EnableHA,
		metrics: &BackendMetrics{
			LastUpdated: time.Now(),
		},
		luaScripts: make(map[string]*redis.Script),
	}

	// Initialize Lua scripts
	backend.initLuaScripts()

	// Initialize health checker
	backend.healthChecker = &HealthChecker{
		backend:  backend,
		stopChan: make(chan struct{}),
		healthStatus: &HealthStatus{
			IsHealthy: true,
			LastCheck: time.Now(),
		},
	}

	// Initialize event publisher
	if config.EnableEvents {
		backend.eventPublisher = &EventPublisher{
			client:      client,
			config:      config,
			eventBuffer: make(chan *enhanced.LockEvent, config.EventBufferSize),
			stopChan:    make(chan struct{}),
		}
	}

	// Start background processes
	backend.start()

	log.Info("Modern Redis backend initialized with clustering: %v, HA: %v",
		config.Redis.EnableClustering, config.Redis.EnableHA)

	return backend, nil
}

// createRedisClient creates appropriate Redis client based on configuration
func createRedisClient(config *ModernLockingConfig) (redis.UniversalClient, error) {
	if config.Redis.EnableClustering {
		return redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:        config.Redis.ClusterNodes,
			PoolSize:     config.Redis.MaxActiveConns,
			MinIdleConns: config.Redis.MaxIdleConns,
			DialTimeout:  config.Redis.ConnTimeout,
			ReadTimeout:  config.Redis.ReadTimeout,
			WriteTimeout: config.Redis.WriteTimeout,
		}), nil
	}

	if config.Redis.EnableHA {
		return redis.NewFailoverClient(&redis.FailoverOptions{
			MasterName:    config.Redis.SentinelMasterName,
			SentinelAddrs: config.Redis.SentinelNodes,
			PoolSize:      config.Redis.MaxActiveConns,
			MinIdleConns:  config.Redis.MaxIdleConns,
			DialTimeout:   config.Redis.ConnTimeout,
			ReadTimeout:   config.Redis.ReadTimeout,
			WriteTimeout:  config.Redis.WriteTimeout,
		}), nil
	}

	// Single instance configuration
	return redis.NewClient(&redis.Options{
		Addr:         "localhost:6379", // Default - should be configurable
		PoolSize:     config.Redis.MaxActiveConns,
		MinIdleConns: config.Redis.MaxIdleConns,
		DialTimeout:  config.Redis.ConnTimeout,
		ReadTimeout:  config.Redis.ReadTimeout,
		WriteTimeout: config.Redis.WriteTimeout,
	}), nil
}

// initLuaScripts initializes all Lua scripts for atomic operations
func (mrb *ModernRedisBackend) initLuaScripts() {
	// Enhanced lock acquisition script with better fairness
	mrb.luaScripts["acquire_lock"] = redis.NewScript(`
		local lockKey = KEYS[1]
		local queueKey = KEYS[2]
		local lockData = ARGV[1]
		local ttl = tonumber(ARGV[2])
		local priority = tonumber(ARGV[3])
		local fairness = ARGV[4] == "true"
		local timestamp = tonumber(ARGV[5])

		-- Check if lock exists
		local existing = redis.call('GET', lockKey)
		if existing then
			-- Add to queue with fair scheduling
			local score = priority * 1000000
			if fairness then
				-- Add randomness for fairness within same priority
				score = score + (timestamp % 1000)
			else
				score = score + timestamp
			end
			redis.call('ZADD', queueKey, score, lockData)
			redis.call('PUBLISH', 'atlantis:lock:queued', lockKey)
			return {false, "queued", score}
		end

		-- Acquire the lock
		if ttl > 0 then
			redis.call('SETEX', lockKey, ttl, lockData)
		else
			redis.call('SET', lockKey, lockData)
		end

		-- Set additional metadata
		local metaKey = lockKey .. ':meta'
		redis.call('HMSET', metaKey,
			'priority', priority,
			'acquired_at', timestamp,
			'node', ARGV[6] or 'unknown')

		if ttl > 0 then
			redis.call('EXPIRE', metaKey, ttl)
		end

		redis.call('PUBLISH', 'atlantis:lock:acquired', lockKey)
		return {true, "acquired", 0}
	`)

	// Enhanced lock release script with queue processing
	mrb.luaScripts["release_lock"] = redis.NewScript(`
		local lockKey = KEYS[1]
		local queueKey = KEYS[2]
		local lockID = ARGV[1]
		local enableQueue = ARGV[2] == "true"

		-- Get current lock
		local currentLock = redis.call('GET', lockKey)
		if not currentLock then
			return {false, "not_found", nil}
		end

		-- Parse and verify lock ownership
		local lockData = cjson.decode(currentLock)
		if lockData.id ~= lockID then
			return {false, "not_owner", nil}
		end

		-- Release the lock and metadata
		redis.call('DEL', lockKey)
		redis.call('DEL', lockKey .. ':meta')

		local nextLock = nil
		if enableQueue then
			-- Process queue fairly
			local queued = redis.call('ZRANGE', queueKey, 0, 0, 'WITHSCORES')
			if #queued > 0 then
				local nextLockData = queued[1]
				local score = queued[2]

				-- Remove from queue
				redis.call('ZREM', queueKey, nextLockData)

				-- Parse and update next lock
				nextLock = cjson.decode(nextLockData)
				nextLock.state = "acquired"
				nextLock.acquired_at = redis.call('TIME')[1]

				-- Acquire lock for queued request
				local nextLockEncoded = cjson.encode(nextLock)
				redis.call('SET', lockKey, nextLockEncoded)

				-- Set metadata for next lock
				local metaKey = lockKey .. ':meta'
				redis.call('HMSET', metaKey,
					'priority', nextLock.priority,
					'acquired_at', nextLock.acquired_at,
					'transferred_from', lockID)

				redis.call('PUBLISH', 'atlantis:lock:transferred', lockKey)
				return {true, "transferred", nextLock}
			end
		end

		redis.call('PUBLISH', 'atlantis:lock:released', lockKey)
		return {true, "released", nil}
	`)

	// Queue management script
	mrb.luaScripts["manage_queue"] = redis.NewScript(`
		local queueKey = KEYS[1]
		local action = ARGV[1]

		if action == "peek" then
			local items = redis.call('ZRANGE', queueKey, 0, 0, 'WITHSCORES')
			return items
		elseif action == "size" then
			return redis.call('ZCARD', queueKey)
		elseif action == "clear" then
			return redis.call('DEL', queueKey)
		elseif action == "remove" then
			local itemToRemove = ARGV[2]
			return redis.call('ZREM', queueKey, itemToRemove)
		end

		return nil
	`)

	// Batch operations script for better performance
	mrb.luaScripts["batch_operations"] = redis.NewScript(`
		local operations = cjson.decode(ARGV[1])
		local results = {}

		for i, op in ipairs(operations) do
			if op.type == "get" then
				results[i] = redis.call('GET', op.key)
			elseif op.type == "set" then
				results[i] = redis.call('SET', op.key, op.value)
			elseif op.type == "del" then
				results[i] = redis.call('DEL', op.key)
			elseif op.type == "exists" then
				results[i] = redis.call('EXISTS', op.key)
			end
		end

		return results
	`)
}

// AcquireLock attempts to acquire a lock with enhanced features
func (mrb *ModernRedisBackend) AcquireLock(ctx context.Context, request *enhanced.EnhancedLockRequest) (*enhanced.EnhancedLock, error) {
	startTime := time.Now()
	defer func() {
		mrb.updateMetrics("acquire", time.Since(startTime), nil)
	}()

	lockID := mrb.generateLockID()
	lockKey := mrb.getLockKey(request.Resource)
	queueKey := mrb.getQueueKey(request.Resource)

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

	lockData, err := json.Marshal(lock)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal lock data: %w", err)
	}

	ttlSeconds := int64(0)
	if lock.ExpiresAt != nil {
		ttlSeconds = int64(lock.ExpiresAt.Sub(time.Now()).Seconds())
	}

	// Use enhanced Lua script
	result, err := mrb.luaScripts["acquire_lock"].Run(ctx, mrb.client,
		[]string{lockKey, queueKey},
		string(lockData),
		ttlSeconds,
		int(request.Priority),
		mrb.config.FairScheduling.Algorithm != "priority",
		time.Now().UnixNano(),
		mrb.getNodeID()).Result()

	if err != nil {
		mrb.updateMetrics("acquire", time.Since(startTime), err)
		return nil, fmt.Errorf("failed to execute lock acquisition script: %w", err)
	}

	resultSlice := result.([]interface{})
	acquired := resultSlice[0].(int64) == 1
	status := resultSlice[1].(string)

	if !acquired {
		if status == "queued" {
			lock.State = enhanced.LockStatePending
			score := resultSlice[2].(int64)
			lock.Metadata["queue_score"] = fmt.Sprintf("%d", score)
			mrb.log.Info("Lock request queued with score %d: %s", score, lockID)

			// Emit queued event
			if mrb.eventPublisher != nil {
				mrb.eventPublisher.Publish(&enhanced.LockEvent{
					Type:      "lock_queued",
					LockID:    lockID,
					Resource:  request.Resource,
					Owner:     request.User.Username,
					Timestamp: time.Now(),
					Metadata:  map[string]string{"queue_score": fmt.Sprintf("%d", score)},
				})
			}

			return lock, nil
		}
		return nil, enhanced.NewLockExistsError(lockKey)
	}

	mrb.log.Info("Lock acquired: %s for resource: %s", lockID, lockKey)

	// Emit acquired event
	if mrb.eventPublisher != nil {
		mrb.eventPublisher.Publish(&enhanced.LockEvent{
			Type:      "lock_acquired",
			LockID:    lockID,
			Resource:  lock.Resource,
			Owner:     lock.Owner,
			Timestamp: time.Now(),
		})
	}

	return lock, nil
}

// ReleaseLock releases an acquired lock with enhanced queue processing
func (mrb *ModernRedisBackend) ReleaseLock(ctx context.Context, lockID string) error {
	startTime := time.Now()
	defer func() {
		mrb.updateMetrics("release", time.Since(startTime), nil)
	}()

	// Find the lock to get its key
	lock, err := mrb.GetLock(ctx, lockID)
	if err != nil {
		return err
	}

	lockKey := mrb.getLockKey(lock.Resource)
	queueKey := mrb.getQueueKey(lock.Resource)

	// Use enhanced release script
	result, err := mrb.luaScripts["release_lock"].Run(ctx, mrb.client,
		[]string{lockKey, queueKey},
		lockID,
		mrb.config.EnablePriorityQueue).Result()

	if err != nil {
		mrb.updateMetrics("release", time.Since(startTime), err)
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

	// Handle transferred lock
	if status == "transferred" && resultSlice[2] != nil {
		transferredLock := resultSlice[2]
		mrb.log.Info("Lock transferred from %s to next in queue", lockID)

		if mrb.eventPublisher != nil {
			mrb.eventPublisher.Publish(&enhanced.LockEvent{
				Type:      "lock_transferred",
				LockID:    lockID,
				Resource:  lock.Resource,
				Timestamp: time.Now(),
				Metadata:  map[string]string{"transferred_to": fmt.Sprintf("%v", transferredLock)},
			})
		}
	} else {
		if mrb.eventPublisher != nil {
			mrb.eventPublisher.Publish(&enhanced.LockEvent{
				Type:      "lock_released",
				LockID:    lockID,
				Resource:  lock.Resource,
				Owner:     lock.Owner,
				Timestamp: time.Now(),
			})
		}
	}

	mrb.log.Info("Lock released: %s, status: %s", lockID, status)
	return nil
}

// Helper methods for the modern Redis backend continue...
// (Additional methods would include GetLock, ListLocks, TryAcquireLock, etc.)

// getNodeID returns a unique identifier for this node
func (mrb *ModernRedisBackend) getNodeID() string {
	// In production, this could be hostname, instance ID, etc.
	hash := sha256.Sum256([]byte(fmt.Sprintf("%p", mrb)))
	return hex.EncodeToString(hash[:8])
}

// getLockKey generates a Redis key for a lock
func (mrb *ModernRedisBackend) getLockKey(resource enhanced.ResourceIdentifier) string {
	return fmt.Sprintf("%slock:%s:%s:%s",
		mrb.config.RedisKeyPrefix,
		resource.Namespace,
		resource.Path,
		resource.Workspace)
}

// getQueueKey generates a Redis key for a queue
func (mrb *ModernRedisBackend) getQueueKey(resource enhanced.ResourceIdentifier) string {
	return fmt.Sprintf("%squeue:%s:%s:%s",
		mrb.config.RedisKeyPrefix,
		resource.Namespace,
		resource.Path,
		resource.Workspace)
}

// generateLockID generates a unique lock ID
func (mrb *ModernRedisBackend) generateLockID() string {
	return fmt.Sprintf("modern_%d_%s", time.Now().UnixNano(), mrb.getNodeID()[:8])
}

// updateMetrics updates backend metrics
func (mrb *ModernRedisBackend) updateMetrics(operation string, duration time.Duration, err error) {
	mrb.metrics.LastUpdated = time.Now()

	switch operation {
	case "acquire":
		mrb.metrics.LockOperations++
	case "release":
		mrb.metrics.UnlockOperations++
	}

	if err != nil {
		mrb.metrics.FailedOps++
	} else {
		mrb.metrics.SuccessfulOps++
	}

	// Update average latency (simple moving average)
	if mrb.metrics.AverageLatency == 0 {
		mrb.metrics.AverageLatency = duration
	} else {
		mrb.metrics.AverageLatency = (mrb.metrics.AverageLatency + duration) / 2
	}
}

// start initializes background processes
func (mrb *ModernRedisBackend) start() {
	// Start health checker
	go mrb.healthChecker.start()

	// Start event publisher
	if mrb.eventPublisher != nil {
		go mrb.eventPublisher.start()
	}
}

// Publish publishes an event to the event buffer
func (ep *EventPublisher) Publish(event *enhanced.LockEvent) {
	select {
	case ep.eventBuffer <- event:
	default:
		// Buffer full, drop event (could implement overflow strategies here)
	}
}

// start begins the event publishing loop
func (ep *EventPublisher) start() {
	ep.wg.Add(1)
	defer ep.wg.Done()

	for {
		select {
		case <-ep.stopChan:
			return
		case event := <-ep.eventBuffer:
			ep.publishEvent(event)
		}
	}
}

// publishEvent publishes a single event to Redis
func (ep *EventPublisher) publishEvent(event *enhanced.LockEvent) {
	ctx := context.Background()

	eventData, err := json.Marshal(event)
	if err != nil {
		return
	}

	channel := fmt.Sprintf("atlantis:lock:%s", event.Type)
	ep.client.Publish(ctx, channel, eventData)
}

// start begins the health checking loop
func (hc *HealthChecker) start() {
	ticker := time.NewTicker(hc.backend.config.Observability.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-hc.stopChan:
			return
		case <-ticker.C:
			hc.performHealthCheck()
		}
	}
}

// performHealthCheck performs a comprehensive health check
func (hc *HealthChecker) performHealthCheck() {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()

	startTime := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(),
		hc.backend.config.Observability.HealthCheckTimeout)
	defer cancel()

	status := &HealthStatus{
		LastCheck: startTime,
		Details:   make(map[string]interface{}),
	}

	// Test basic connectivity
	_, err := hc.backend.client.Ping(ctx).Result()
	if err != nil {
		status.IsHealthy = false
		status.Errors = append(status.Errors, fmt.Sprintf("ping failed: %v", err))
	}

	// Test Lua script functionality
	if status.IsHealthy {
		testResult, err := hc.backend.luaScripts["manage_queue"].Run(ctx, hc.backend.client,
			[]string{"health:test:queue"}, "size").Result()
		if err != nil {
			status.IsHealthy = false
			status.Errors = append(status.Errors, fmt.Sprintf("lua script test failed: %v", err))
		} else {
			status.Details["lua_test_result"] = testResult
		}
	}

	// Get Redis info if healthy
	if status.IsHealthy {
		info, err := hc.backend.client.Info(ctx, "memory", "replication", "cluster").Result()
		if err == nil {
			status.Details["redis_info"] = hc.parseRedisInfo(info)
		}
	}

	status.CheckDuration = time.Since(startTime)
	if status.IsHealthy {
		status.IsHealthy = status.CheckDuration < hc.backend.config.Observability.HealthCheckTimeout
	}

	hc.healthStatus = status
}

// parseRedisInfo parses Redis INFO output
func (hc *HealthChecker) parseRedisInfo(info string) map[string]string {
	result := make(map[string]string)
	lines := strings.Split(info, "\n")

	for _, line := range lines {
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				result[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			}
		}
	}

	return result
}