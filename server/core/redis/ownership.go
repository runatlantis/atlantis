// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package redis

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	redislib "github.com/redis/go-redis/v9"
	"github.com/runatlantis/atlantis/server/core/ownership"
	"github.com/runatlantis/atlantis/server/logging"
)

const ownerRecordSchemaVersion = 1

const claimOwnerScript = "" +
	"local current = redis.call(\"GET\", KEYS[1])\n" +
	"if current then\n" +
	"  return current\n" +
	"end\n" +
	"redis.call(\"SET\", KEYS[1], ARGV[1], \"NX\", \"PX\", ARGV[2])\n" +
	"return redis.call(\"GET\", KEYS[1])\n"

const renewOwnerScript = "" +
	"local current = redis.call(\"GET\", KEYS[1])\n" +
	"if current ~= ARGV[1] then\n" +
	"  return 0\n" +
	"end\n" +
	"redis.call(\"PEXPIRE\", KEYS[1], ARGV[2])\n" +
	"return 1\n"

const releaseOwnerScript = "" +
	"local current = redis.call(\"GET\", KEYS[1])\n" +
	"if current ~= ARGV[1] then\n" +
	"  return 0\n" +
	"end\n" +
	"redis.call(\"DEL\", KEYS[1])\n" +
	"return 1\n"

// OwnerStoreConfig identifies a replica and configures its ownership lease.
type OwnerStoreConfig struct {
	ReplicaID    string
	AdvertiseURL string
	TTL          time.Duration
}

type ownedRecord struct {
	key        ownership.Key
	record     ownership.Record
	serialized string
}

// OwnerStore stores renewable pull request ownership records in Redis.
type OwnerStore struct {
	client     redislib.Cmdable
	config     OwnerStoreConfig
	logger     logging.SimpleLogging
	instanceID string
	draining   atomic.Bool

	mu    sync.RWMutex
	owned map[string]ownedRecord

	healthMu            sync.RWMutex
	renewalFailureSince time.Time
	lastRenewalErr      error

	cancel context.CancelFunc
	done   chan struct{}

	closeOnce sync.Once
	closeErr  error
}

// NewOwnerStore creates an ownership store using an existing Redis database client.
func NewOwnerStore(database *RedisDB, config OwnerStoreConfig, logger logging.SimpleLogging) (*OwnerStore, error) {
	if database == nil {
		return nil, errors.New("redis database is required")
	}
	return NewOwnerStoreWithClient(database.client, config, logger)
}

// NewOwnerStoreWithClient creates an ownership store with an injected Redis client.
func NewOwnerStoreWithClient(client redislib.Cmdable, config OwnerStoreConfig, logger logging.SimpleLogging) (*OwnerStore, error) {
	if client == nil {
		return nil, errors.New("redis client is required")
	}
	if config.ReplicaID == "" {
		return nil, errors.New("replica ID is required")
	}
	if config.AdvertiseURL == "" {
		return nil, errors.New("replica advertise URL is required")
	}
	if config.TTL <= 0 {
		return nil, errors.New("ownership TTL must be positive")
	}

	loopCtx, cancel := context.WithCancel(context.Background())
	store := &OwnerStore{
		client:     client,
		config:     config,
		logger:     logger,
		instanceID: uuid.NewString(),
		owned:      make(map[string]ownedRecord),
		cancel:     cancel,
		done:       make(chan struct{}),
	}
	go store.renewLoop(loopCtx)
	return store, nil
}

// Claim creates a lease when no owner exists and otherwise returns the current owner.
func (s *OwnerStore) Claim(ctx context.Context, key ownership.Key) (ownership.Record, error) {
	if s.draining.Load() {
		return ownership.Record{}, ownership.ErrDraining
	}
	redisKey, err := redisOwnershipKey(key)
	if err != nil {
		return ownership.Record{}, err
	}
	record := ownership.Record{
		SchemaVersion: ownerRecordSchemaVersion,
		ReplicaID:     s.config.ReplicaID,
		InstanceID:    s.instanceID,
		AdvertiseURL:  s.config.AdvertiseURL,
		ClaimID:       uuid.NewString(),
		ClaimedAt:     time.Now().UTC(),
	}
	serialized, err := json.Marshal(record)
	if err != nil {
		return ownership.Record{}, fmt.Errorf("serializing ownership claim: %w", err)
	}

	value, err := s.client.Eval(ctx, claimOwnerScript, []string{redisKey}, serialized, s.config.TTL.Milliseconds()).Result()
	if err != nil {
		return ownership.Record{}, fmt.Errorf("claiming PR ownership: %w", err)
	}
	currentSerialized, ok := value.(string)
	if !ok {
		return ownership.Record{}, fmt.Errorf("unexpected ownership claim result type %T", value)
	}
	current, err := decodeOwnerRecord(currentSerialized)
	if err != nil {
		return ownership.Record{}, err
	}
	if current.InstanceID == s.instanceID {
		s.mu.Lock()
		s.owned[redisKey] = ownedRecord{key: key, record: current, serialized: currentSerialized}
		s.mu.Unlock()
	}
	return current, nil
}

// Current returns the current live ownership record, if one exists.
func (s *OwnerStore) Current(ctx context.Context, key ownership.Key) (ownership.Record, bool, error) {
	redisKey, err := redisOwnershipKey(key)
	if err != nil {
		return ownership.Record{}, false, err
	}
	serialized, err := s.client.Get(ctx, redisKey).Result()
	if err == redislib.Nil {
		return ownership.Record{}, false, nil
	}
	if err != nil {
		return ownership.Record{}, false, fmt.Errorf("reading PR ownership: %w", err)
	}
	record, err := decodeOwnerRecord(serialized)
	if err != nil {
		return ownership.Record{}, false, err
	}
	return record, true, nil
}

// Owns reports whether this process holds the exact claim in its local lease set.
func (s *OwnerStore) Owns(key ownership.Key, claimID string) bool {
	redisKey, err := redisOwnershipKey(key)
	if err != nil {
		return false
	}
	s.mu.RLock()
	owned, ok := s.owned[redisKey]
	s.mu.RUnlock()
	return ok && owned.record.InstanceID == s.instanceID && owned.record.ClaimID == claimID
}

// Release removes a claim only when this process still owns the exact record.
func (s *OwnerStore) Release(ctx context.Context, key ownership.Key, claimID string) error {
	redisKey, err := redisOwnershipKey(key)
	if err != nil {
		return err
	}
	s.mu.RLock()
	owned, ok := s.owned[redisKey]
	s.mu.RUnlock()
	if !ok || owned.record.ClaimID != claimID {
		return nil
	}

	_, err = s.releaseOwned(ctx, redisKey, owned)
	return err
}

// BeginDrain prevents new claims while allowing existing claims to renew.
func (s *OwnerStore) BeginDrain() {
	s.draining.Store(true)
}

// Ready reports whether the store can safely accept routed work.
func (s *OwnerStore) Ready(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if s.draining.Load() {
		return ownership.ErrDraining
	}

	s.healthMu.RLock()
	failureSince := s.renewalFailureSince
	lastErr := s.lastRenewalErr
	s.healthMu.RUnlock()
	if !failureSince.IsZero() && time.Since(failureSince) >= s.config.TTL/2 {
		return fmt.Errorf("ownership renewal unhealthy: %w", lastErr)
	}
	return nil
}

// Close stops renewal and releases claims still owned by this process.
func (s *OwnerStore) Close() error {
	s.closeOnce.Do(func() {
		s.BeginDrain()
		s.cancel()
		<-s.done

		s.mu.RLock()
		owned := make(map[string]ownedRecord, len(s.owned))
		for redisKey, record := range s.owned {
			owned[redisKey] = record
		}
		s.mu.RUnlock()

		var releaseErrors []error
		for redisKey, record := range owned {
			if _, err := s.releaseOwned(context.Background(), redisKey, record); err != nil {
				releaseErrors = append(releaseErrors, err)
			}
		}
		s.closeErr = errors.Join(releaseErrors...)
	})
	return s.closeErr
}

func (s *OwnerStore) renewLoop(ctx context.Context) {
	defer close(s.done)
	interval := s.config.TTL / 3
	if interval < time.Millisecond {
		interval = time.Millisecond
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.renewOwned(ctx)
		}
	}
}

func (s *OwnerStore) renewOwned(ctx context.Context) {
	renewalStarted := time.Now()
	renewalTimeout := s.config.TTL / 3
	if renewalTimeout < time.Millisecond {
		renewalTimeout = time.Millisecond
	}
	renewalCtx, cancel := context.WithTimeout(ctx, renewalTimeout)
	defer cancel()

	s.mu.RLock()
	owned := make(map[string]ownedRecord, len(s.owned))
	for redisKey, record := range s.owned {
		owned[redisKey] = record
	}
	s.mu.RUnlock()

	commands := make(map[string]*redislib.Cmd, len(owned))
	_, pipelineErr := s.client.Pipelined(renewalCtx, func(pipe redislib.Pipeliner) error {
		for redisKey, record := range owned {
			commands[redisKey] = pipe.Eval(
				renewalCtx,
				renewOwnerScript,
				[]string{redisKey},
				record.serialized,
				s.config.TTL.Milliseconds(),
			)
		}
		return nil
	})

	var renewalErr error
	for redisKey, command := range commands {
		renewed, err := command.Int()
		if err != nil {
			renewalErr = errors.Join(renewalErr, fmt.Errorf("renewing PR ownership: %w", err))
			continue
		}
		if renewed != 1 {
			record := owned[redisKey]
			s.removeOwned(redisKey, record.record.ClaimID)
		}
	}
	if pipelineErr != nil && renewalErr == nil {
		renewalErr = fmt.Errorf("renewing PR ownership batch: %w", pipelineErr)
	}
	if renewalErr != nil {
		s.markRenewalFailure(renewalStarted, renewalErr)
		s.logger.Warn("ownership renewal failed: %s", renewalErr)
		return
	}
	s.clearRenewalFailure()
}

func (s *OwnerStore) renewClaim(ctx context.Context, key ownership.Key, record ownership.Record) (bool, error) {
	redisKey, err := redisOwnershipKey(key)
	if err != nil {
		return false, err
	}
	serialized, err := json.Marshal(record)
	if err != nil {
		return false, fmt.Errorf("serializing ownership claim: %w", err)
	}
	return s.renewSerialized(ctx, redisKey, string(serialized))
}

func (s *OwnerStore) renewSerialized(ctx context.Context, redisKey, serialized string) (bool, error) {
	renewed, err := s.client.Eval(ctx, renewOwnerScript, []string{redisKey}, serialized, s.config.TTL.Milliseconds()).Int()
	if err != nil {
		return false, fmt.Errorf("renewing PR ownership: %w", err)
	}
	return renewed == 1, nil
}

func (s *OwnerStore) releaseOwned(ctx context.Context, redisKey string, owned ownedRecord) (bool, error) {
	released, err := s.client.Eval(ctx, releaseOwnerScript, []string{redisKey}, owned.serialized).Int()
	if err != nil {
		return false, fmt.Errorf("releasing PR ownership: %w", err)
	}
	s.removeOwned(redisKey, owned.record.ClaimID)
	return released == 1, nil
}

func (s *OwnerStore) removeOwned(redisKey, claimID string) {
	s.mu.Lock()
	if current, ok := s.owned[redisKey]; ok && current.record.ClaimID == claimID {
		delete(s.owned, redisKey)
	}
	s.mu.Unlock()
}

func (s *OwnerStore) markRenewalFailure(started time.Time, err error) {
	s.healthMu.Lock()
	if s.renewalFailureSince.IsZero() {
		s.renewalFailureSince = started
	}
	s.lastRenewalErr = err
	s.healthMu.Unlock()
}

func (s *OwnerStore) clearRenewalFailure() {
	s.healthMu.Lock()
	s.renewalFailureSince = time.Time{}
	s.lastRenewalErr = nil
	s.healthMu.Unlock()
}

func redisOwnershipKey(key ownership.Key) (string, error) {
	canonical, err := key.Canonical()
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256([]byte(canonical))
	return "atlantis:ha:ownership:v1:" + hex.EncodeToString(sum[:]), nil
}

func decodeOwnerRecord(serialized string) (ownership.Record, error) {
	var record ownership.Record
	if err := json.Unmarshal([]byte(serialized), &record); err != nil {
		return ownership.Record{}, fmt.Errorf("deserializing PR ownership: %w", err)
	}
	if record.SchemaVersion != ownerRecordSchemaVersion || record.ReplicaID == "" || record.InstanceID == "" || record.AdvertiseURL == "" || record.ClaimID == "" {
		return ownership.Record{}, errors.New("invalid PR ownership record")
	}
	return record, nil
}
