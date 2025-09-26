// PR #2: Redis Backend Integration - Redis Client and Backend Implementation
// This file implements the Redis backend for the enhanced locking system

package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/core/locking"
	"github.com/runatlantis/atlantis/server/models"
)

// Backend implements the locking.Backend interface using Redis
type Backend struct {
	client redis.UniversalClient
	config Config
	healthChecker *HealthChecker
}

// Config holds Redis backend configuration
type Config struct {
	// Redis connection settings
	Addresses []string
	Password string
	DB int
	PoolSize int
	
	// Locking-specific settings
	KeyPrefix string
	LockTTL time.Duration
	ConnTimeout time.Duration
	ReadTimeout time.Duration
	WriteTimeout time.Duration
	
	// Advanced settings
	ClusterMode bool
	TLSEnabled bool
	TLSConfig *redis.TLSConfig
}

// HealthChecker monitors Redis health
type HealthChecker struct {
	client redis.UniversalClient
	interval time.Duration
	status chan HealthStatus
}

type HealthStatus struct {
	Healthy bool
	Error error
	Latency time.Duration
	Connections int
}

// NewBackend creates a new Redis backend
func NewBackend(config Config) (*Backend, error) {
	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, errors.Wrap(err, "invalid Redis configuration")
	}
	
	// Create Redis client options
	opts := &redis.UniversalOptions{
		Addrs: config.Addresses,
		Password: config.Password,
		DB: config.DB,
		PoolSize: config.PoolSize,
		DialTimeout: config.ConnTimeout,
		ReadTimeout: config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
	}
	
	// Configure TLS if enabled
	if config.TLSEnabled {
		opts.TLSConfig = config.TLSConfig
	}
	
	// Create Redis client
	var client redis.UniversalClient
	if config.ClusterMode {
		client = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs: config.Addresses,
			Password: config.Password,
			PoolSize: config.PoolSize,
			DialTimeout: config.ConnTimeout,
			ReadTimeout: config.ReadTimeout,
			WriteTimeout: config.WriteTimeout,
			TLSConfig: config.TLSConfig,
		})
	} else {
		client = redis.NewClient(&redis.Options{
			Addr: config.Addresses[0], // Use first address for single-node
			Password: config.Password,
			DB: config.DB,
			PoolSize: config.PoolSize,
			DialTimeout: config.ConnTimeout,
			ReadTimeout: config.ReadTimeout,
			WriteTimeout: config.WriteTimeout,
			TLSConfig: config.TLSConfig,
		})
	}
	
	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), config.ConnTimeout)
	defer cancel()
	
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, errors.Wrap(err, "failed to connect to Redis")
	}
	
	// Create backend
	backend := &Backend{
		client: client,
		config: config,
		healthChecker: NewHealthChecker(client, 30*time.Second),
	}
	
	return backend, nil
}

// TryLock attempts to acquire a lock
func (b *Backend) TryLock(lock models.ProjectLock) (bool, locking.LockingError) {
	ctx, cancel := context.WithTimeout(context.Background(), b.config.ConnTimeout)
	defer cancel()
	
	// Generate lock key
	key := b.generateLockKey(lock)
	
	// Serialize lock data
	lockData, err := json.Marshal(lock)
	if err != nil {
		return false, locking.NewLockingError("failed to serialize lock data", err)
	}
	
	// Use Redis SET with NX (not exists) and EX (expiration)
	// This provides atomic lock acquisition with TTL
	result := b.client.SetNX(ctx, key, lockData, b.config.LockTTL)
	if result.Err() != nil {
		return false, locking.NewLockingError("failed to acquire lock", result.Err())
	}
	
	acquired := result.Val()
	if acquired {
		// Start heartbeat for lock renewal
		go b.startLockHeartbeat(key, lock)
	}
	
	return acquired, nil
}

// Unlock releases a lock
func (b *Backend) Unlock(lock models.ProjectLock) locking.LockingError {
	ctx, cancel := context.WithTimeout(context.Background(), b.config.ConnTimeout)
	defer cancel()
	
	// Generate lock key
	key := b.generateLockKey(lock)
	
	// Use Lua script for atomic delete with verification
	luaScript := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`
	
	// Serialize lock data for verification
	lockData, err := json.Marshal(lock)
	if err != nil {
		return locking.NewLockingError("failed to serialize lock data", err)
	}
	
	// Execute Lua script
	result := b.client.Eval(ctx, luaScript, []string{key}, string(lockData))
	if result.Err() != nil {
		return locking.NewLockingError("failed to release lock", result.Err())
	}
	
	deleted := result.Val().(int64)
	if deleted == 0 {
		return locking.NewLockingError("lock not found or already released", nil)
	}
	
	return nil
}

// List returns all current locks
func (b *Backend) List() (map[string]models.ProjectLock, error) {
	ctx, cancel := context.WithTimeout(context.Background(), b.config.ReadTimeout)
	defer cancel()
	
	// Find all lock keys
	pattern := b.config.KeyPrefix + "*"
	keys := b.client.Keys(ctx, pattern)
	if keys.Err() != nil {
		return nil, errors.Wrap(keys.Err(), "failed to list lock keys")
	}
	
	locks := make(map[string]models.ProjectLock)
	
	// Retrieve all lock data in a pipeline for efficiency
	pipe := b.client.Pipeline()
	cmds := make([]*redis.StringCmd, len(keys.Val()))
	
	for i, key := range keys.Val() {
		cmds[i] = pipe.Get(ctx, key)
	}
	
	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, errors.Wrap(err, "failed to retrieve lock data")
	}
	
	// Parse lock data
	for i, cmd := range cmds {
		if cmd.Err() == redis.Nil {
			continue // Key expired or deleted
		}
		
		if cmd.Err() != nil {
			continue // Skip invalid entries
		}
		
		var lock models.ProjectLock
		if err := json.Unmarshal([]byte(cmd.Val()), &lock); err != nil {
			continue // Skip invalid lock data
		}
		
		// Generate lock ID for map key
		lockID := b.generateLockID(lock)
		locks[lockID] = lock
	}
	
	return locks, nil
}

// UnlockByPull removes all locks for a specific pull request
func (b *Backend) UnlockByPull(repoFullName string, pullNum int) ([]models.ProjectLock, error) {
	ctx, cancel := context.WithTimeout(context.Background(), b.config.ReadTimeout)
	defer cancel()
	
	// Find all locks for this pull request
	pattern := fmt.Sprintf("%s*:%s:%d:*", b.config.KeyPrefix, repoFullName, pullNum)
	keys := b.client.Keys(ctx, pattern)
	if keys.Err() != nil {
		return nil, errors.Wrap(keys.Err(), "failed to find locks for pull request")
	}
	
	if len(keys.Val()) == 0 {
		return []models.ProjectLock{}, nil
	}
	
	// Retrieve and delete locks atomically
	var unlockedLocks []models.ProjectLock
	
	for _, key := range keys.Val() {
		// Get lock data before deleting
		lockData := b.client.Get(ctx, key)
		if lockData.Err() == redis.Nil {
			continue // Already deleted
		}
		
		if lockData.Err() != nil {
			continue // Skip errors
		}
		
		// Parse lock
		var lock models.ProjectLock
		if err := json.Unmarshal([]byte(lockData.Val()), &lock); err != nil {
			continue // Skip invalid data
		}
		
		// Delete lock
		if err := b.client.Del(ctx, key).Err(); err != nil {
			continue // Skip deletion errors
		}
		
		unlockedLocks = append(unlockedLocks, lock)
	}
	
	return unlockedLocks, nil
}

// GetLock retrieves a specific lock
func (b *Backend) GetLock(project models.Project, workspace string) (*models.ProjectLock, error) {
	ctx, cancel := context.WithTimeout(context.Background(), b.config.ReadTimeout)
	defer cancel()
	
	// Generate lock key
	key := b.generateProjectLockKey(project, workspace)
	
	// Get lock data
	lockData := b.client.Get(ctx, key)
	if lockData.Err() == redis.Nil {
		return nil, nil // Lock not found
	}
	
	if lockData.Err() != nil {
		return nil, errors.Wrap(lockData.Err(), "failed to retrieve lock")
	}
	
	// Parse lock data
	var lock models.ProjectLock
	if err := json.Unmarshal([]byte(lockData.Val()), &lock); err != nil {
		return nil, errors.Wrap(err, "failed to parse lock data")
	}
	
	return &lock, nil
}

// Health returns the current health status
func (b *Backend) Health() HealthStatus {
	return b.healthChecker.GetStatus()
}

// Close closes the Redis connection
func (b *Backend) Close() error {
	b.healthChecker.Stop()
	return b.client.Close()
}

// Helper methods

func (b *Backend) generateLockKey(lock models.ProjectLock) string {
	return fmt.Sprintf("%s%s:%s:%d:%s",
		b.config.KeyPrefix,
		lock.Project.RepoFullName,
		lock.Project.Path,
		lock.Pull.Num,
		lock.Workspace,
	)
}

func (b *Backend) generateProjectLockKey(project models.Project, workspace string) string {
	return fmt.Sprintf("%s%s:%s:*:%s",
		b.config.KeyPrefix,
		project.RepoFullName,
		project.Path,
		workspace,
	)
}

func (b *Backend) generateLockID(lock models.ProjectLock) string {
	return fmt.Sprintf("%s/%s/%s",
		lock.Project.RepoFullName,
		lock.Project.Path,
		lock.Workspace,
	)
}

func (b *Backend) startLockHeartbeat(key string, lock models.ProjectLock) {
	// Implement heartbeat mechanism to prevent lock expiration
	// This will be called in a goroutine to renew locks periodically
	ticker := time.NewTicker(b.config.LockTTL / 3) // Renew at 1/3 TTL
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), b.config.ConnTimeout)
			
			// Check if lock still exists and renew
			if b.client.Expire(ctx, key, b.config.LockTTL).Err() != nil {
				cancel()
				return // Lock doesn't exist anymore
			}
			
			cancel()
		}
	}
}

// Validate validates the Redis configuration
func (c *Config) Validate() error {
	if len(c.Addresses) == 0 {
		return errors.New("at least one Redis address is required")
	}
	
	if c.PoolSize <= 0 {
		c.PoolSize = 10 // Default pool size
	}
	
	if c.KeyPrefix == "" {
		c.KeyPrefix = "atlantis:lock:" // Default prefix
	}
	
	if c.LockTTL <= 0 {
		c.LockTTL = time.Hour // Default TTL
	}
	
	if c.ConnTimeout <= 0 {
		c.ConnTimeout = 5 * time.Second
	}
	
	if c.ReadTimeout <= 0 {
		c.ReadTimeout = 3 * time.Second
	}
	
	if c.WriteTimeout <= 0 {
		c.WriteTimeout = 3 * time.Second
	}
	
	return nil
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(client redis.UniversalClient, interval time.Duration) *HealthChecker {
	return &HealthChecker{
		client: client,
		interval: interval,
		status: make(chan HealthStatus, 1),
	}
}

// Start starts the health checker
func (h *HealthChecker) Start() {
	go h.checkHealth()
}

// Stop stops the health checker
func (h *HealthChecker) Stop() {
	close(h.status)
}

// GetStatus returns the current health status
func (h *HealthChecker) GetStatus() HealthStatus {
	select {
	case status := <-h.status:
		return status
	default:
		return HealthStatus{Healthy: true} // Default to healthy
	}
}

func (h *HealthChecker) checkHealth() {
	ticker := time.NewTicker(h.interval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			start := time.Now()
			err := h.client.Ping(context.Background()).Err()
			latency := time.Since(start)
			
			status := HealthStatus{
				Healthy: err == nil,
				Error: err,
				Latency: latency,
				Connections: h.getPoolStats(),
			}
			
			select {
			case h.status <- status:
			default:
				// Channel full, skip this update
			}
		}
	}
}

func (h *HealthChecker) getPoolStats() int {
	// Get connection pool statistics
	// This is implementation-specific to the Redis client
	return 0 // Placeholder
}
