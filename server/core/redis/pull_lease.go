// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package redis

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

const acquirePullLeaseScript = "" +
	"local current = redis.call(\"GET\", KEYS[1])\n" +
	"if current then\n" +
	"  return {0, current}\n" +
	"end\n" +
	"redis.call(\"SET\", KEYS[1], ARGV[1], \"PX\", ARGV[2])\n" +
	"return {1, ARGV[1]}\n"

const renewPullLeaseScript = "" +
	"if redis.call(\"GET\", KEYS[1]) ~= ARGV[1] then\n" +
	"  return 0\n" +
	"end\n" +
	"redis.call(\"PEXPIRE\", KEYS[1], ARGV[2])\n" +
	"return 1\n"

const releasePullLeaseScript = "" +
	"if redis.call(\"GET\", KEYS[1]) ~= ARGV[1] then\n" +
	"  return 0\n" +
	"end\n" +
	"redis.call(\"DEL\", KEYS[1])\n" +
	"return 1\n"

// TryAcquirePullLease atomically acquires a renewable lease for a pull request.
// When another process owns the lease, it returns that process's serialized
// value so callers can preserve the existing working-directory lock error.
func (r *RedisDB) TryAcquirePullLease(ctx context.Context, repoFullName string, pullNum int, value string, ttl time.Duration) (bool, string, error) {
	result, err := r.client.Eval(
		ctx,
		acquirePullLeaseScript,
		[]string{pullLeaseKey(repoFullName, pullNum)},
		value,
		ttl.Milliseconds(),
	).Result()
	if err != nil {
		return false, "", fmt.Errorf("acquiring pull lease: %w", err)
	}

	values, ok := result.([]any)
	if !ok || len(values) != 2 {
		return false, "", fmt.Errorf("unexpected pull lease result type %T", result)
	}
	acquired, ok := values[0].(int64)
	if !ok {
		return false, "", fmt.Errorf("unexpected pull lease acquired result type %T", values[0])
	}
	current, ok := values[1].(string)
	if !ok {
		return false, "", fmt.Errorf("unexpected pull lease value result type %T", values[1])
	}
	return acquired == 1, current, nil
}

// RenewPullLease extends a lease only while value remains its exact owner.
func (r *RedisDB) RenewPullLease(ctx context.Context, repoFullName string, pullNum int, value string, ttl time.Duration) (bool, error) {
	result, err := r.client.Eval(
		ctx,
		renewPullLeaseScript,
		[]string{pullLeaseKey(repoFullName, pullNum)},
		value,
		ttl.Milliseconds(),
	).Int64()
	if err != nil {
		return false, fmt.Errorf("renewing pull lease: %w", err)
	}
	return result == 1, nil
}

// ReleasePullLease removes a lease only while value remains its exact owner.
func (r *RedisDB) ReleasePullLease(ctx context.Context, repoFullName string, pullNum int, value string) (bool, error) {
	result, err := r.client.Eval(
		ctx,
		releasePullLeaseScript,
		[]string{pullLeaseKey(repoFullName, pullNum)},
		value,
	).Int64()
	if err != nil {
		return false, fmt.Errorf("releasing pull lease: %w", err)
	}
	return result == 1, nil
}

func pullLeaseKey(repoFullName string, pullNum int) string {
	key := fmt.Sprintf("%s\x00%d", repoFullName, pullNum)
	hash := sha256.Sum256([]byte(key))
	return "working-dir-lock/pull/" + hex.EncodeToString(hash[:])
}
