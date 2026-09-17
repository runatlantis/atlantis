// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package redis

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"github.com/runatlantis/atlantis/server/core/db"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
)

// Sharing the status key's hash slot permits atomic status+lease transactions
// on Redis Cluster as well as standalone Redis.
func publicationLeaseKey(statusKey string) (string, error) {
	tag := statusKey
	if start := strings.IndexByte(statusKey, '{'); start >= 0 {
		end := strings.IndexByte(statusKey[start+1:], '}')
		if end <= 0 {
			// Redis hashes the complete key when its first brace pair is empty
			// or unclosed. Find an equivalent binary tag instead of rejecting a
			// repository identity accepted by the existing status store.
			tag = publicationSlotTag(statusKey)
		} else {
			tag = statusKey[start+1 : start+1+end]
		}
	} else if strings.ContainsRune(statusKey, '}') {
		tag = publicationSlotTag(statusKey)
	}
	digest := sha256.Sum256([]byte(statusKey))
	return "publication:{" + tag + "}:" + hex.EncodeToString(digest[:]), nil
}

// Redis Cluster uses CRC16/XMODEM modulo 16384. Every slot has a two-byte
// preimage without brace delimiters, so this bounded search handles unusual
// legacy keys without changing either their status key or cluster slot.
func publicationSlotTag(key string) string {
	slot := publicationCRC16(key) & 16383
	for candidate := range 1 << 16 {
		tag := [2]byte{byte((candidate >> 8) & 255), byte(candidate & 255)}
		if tag[0] == '{' || tag[0] == '}' || tag[1] == '{' || tag[1] == '}' {
			continue
		}
		if publicationCRC16(string(tag[:]))&16383 == slot {
			return string(tag[:])
		}
	}
	panic("all Redis slots must have a brace-free two-byte tag")
}

func publicationCRC16(value string) uint16 {
	var crc uint16
	for i := range len(value) {
		crc ^= uint16(value[i]) << 8
		for range 8 {
			if crc&0x8000 != 0 {
				crc = crc<<1 ^ 0x1021
			} else {
				crc <<= 1
			}
		}
	}
	return crc
}

const compareAndSwapPublicationLease = `
local current = redis.call("GET", KEYS[1])
if ARGV[1] == "0" then
 if current then return 0 end
elseif not current or current ~= ARGV[2] then
 return 0
end
if ARGV[4] ~= "" then
 if not current then return -1 end
 local lease = cjson.decode(current)
 local now = redis.call("TIME")
 local millis = tonumber(now[1]) * 1000 + math.floor(tonumber(now[2]) / 1000)
 if lease.Owner ~= ARGV[4] or lease.DeadlineUnixMilli <= millis then return -1 end
end
if ARGV[3] == "" then redis.call("DEL", KEYS[1])
else redis.call("SET", KEYS[1], ARGV[3]) end
return 1
`

func (r *RedisDB) mutatePublicationLease(ctx context.Context, pull models.PullRequest, activeOwner string, transition func(*db.PublicationLease, time.Time) (*db.PublicationLease, error)) (db.PublicationLease, error) {
	statusKey, err := r.pullKey(pull)
	if err != nil {
		return db.PublicationLease{}, err
	}
	key, err := publicationLeaseKey(statusKey)
	if err != nil {
		return db.PublicationLease{}, err
	}
	for range 8 {
		if err := ctx.Err(); err != nil {
			return db.PublicationLease{}, err
		}
		raw, err := r.client.Get(ctx, key).Result()
		if err != nil && !errors.Is(err, goredis.Nil) {
			return db.PublicationLease{}, err
		}
		exists := "0"
		var current *db.PublicationLease
		if err == nil {
			exists = "1"
			current = &db.PublicationLease{}
			if err := json.Unmarshal([]byte(raw), current); err != nil {
				return db.PublicationLease{}, err
			}
		}
		// Deadlines use the Redis clock, never a replica's possibly skewed clock.
		now, err := r.client.Time(ctx).Result()
		if err != nil {
			return db.PublicationLease{}, err
		}
		next, err := transition(current, now)
		if err != nil {
			return db.PublicationLease{}, err
		}
		encoded := ""
		if next != nil {
			body, err := json.Marshal(next)
			if err != nil {
				return db.PublicationLease{}, err
			}
			encoded = string(body)
		}
		changed, err := r.client.Eval(ctx, compareAndSwapPublicationLease, []string{key}, exists, raw, encoded, activeOwner).Int()
		if err != nil {
			return db.PublicationLease{}, err
		}
		if changed == -1 {
			return db.PublicationLease{}, db.ErrPublicationOwnerLost
		}
		if changed == 1 {
			if next == nil {
				return db.PublicationLease{}, nil
			}
			return *next, nil
		}
	}
	return db.PublicationLease{}, db.ErrPublicationBusy
}

func (r *RedisDB) AcquirePublicationLease(ctx context.Context, pull models.PullRequest, owner string, ttl time.Duration) (db.PublicationLease, error) {
	return r.mutatePublicationLease(ctx, pull, "", func(current *db.PublicationLease, now time.Time) (*db.PublicationLease, error) {
		next, err := db.AcquirePublicationLease(current, owner, now, ttl)
		return &next, err
	})
}
func (r *RedisDB) RenewPublicationLease(ctx context.Context, pull models.PullRequest, fence db.PublicationFence, ttl time.Duration) (db.PublicationLease, error) {
	return r.mutatePublicationLease(ctx, pull, fence.Owner, func(current *db.PublicationLease, now time.Time) (*db.PublicationLease, error) {
		next, err := db.RenewPublicationLease(current, fence, now, ttl)
		return &next, err
	})
}
func (r *RedisDB) BeginPublication(ctx context.Context, pull models.PullRequest, fence db.PublicationFence) error {
	_, err := r.mutatePublicationLease(ctx, pull, fence.Owner, func(current *db.PublicationLease, now time.Time) (*db.PublicationLease, error) {
		next, err := db.BeginPublication(current, fence, now)
		return &next, err
	})
	return err
}
func (r *RedisDB) CompletePublication(ctx context.Context, pull models.PullRequest, fence db.PublicationFence) error {
	_, err := r.mutatePublicationLease(ctx, pull, "", func(current *db.PublicationLease, _ time.Time) (*db.PublicationLease, error) {
		next, err := db.CompletePublication(current, fence)
		return &next, err
	})
	return err
}
func (r *RedisDB) ReleasePublicationLease(ctx context.Context, pull models.PullRequest, fence db.PublicationFence) error {
	_, err := r.mutatePublicationLease(ctx, pull, "", func(current *db.PublicationLease, _ time.Time) (*db.PublicationLease, error) {
		return nil, db.CanReleasePublicationLease(current, fence)
	})
	return err
}

var _ db.PublicationLeaseStore = (*RedisDB)(nil)

// This guard and the status mutation run in one Lua invocation. Checking the
// lease with a separate GET would permit an expired owner to commit later.
const publicationWriteGuardLua = `
local rawLease = redis.call("GET", KEYS[2])
local lease = nil
if rawLease then lease = cjson.decode(rawLease) end
local clock = redis.call("TIME")
local millis = tonumber(clock[1]) * 1000 + math.floor(tonumber(clock[2]) / 1000)
if ARGV[4] == "" then
 if lease then
  if lease.Publishing then return -3 end
  if lease.DeadlineUnixMilli > millis then return -1 end
 end
else
 if not lease or lease.Owner ~= ARGV[4] or lease.DeadlineUnixMilli <= millis then return -2 end
 if lease.Publishing then return -3 end
end
`

func publicationWriteOwner(mode command.PublicationWriteMode) (string, error) {
	switch fence := mode.(type) {
	case command.NoClaim:
		return "", nil
	case command.PublicationFence:
		if fence.Owner != "" {
			return fence.Owner, nil
		}
	}
	return "", db.ErrPublicationOwnerLost
}
func publicationWriteError(result int) error {
	switch result {
	case -1:
		return db.ErrPublicationBusy
	case -3:
		return db.ErrPublicationAmbiguous
	default:
		return db.ErrPublicationOwnerLost
	}
}

func (r *RedisDB) GetPublicationLease(ctx context.Context, pull models.PullRequest) (*db.PublicationLease, error) {
	statusKey, err := r.pullKey(pull)
	if err != nil {
		return nil, err
	}
	key, err := publicationLeaseKey(statusKey)
	if err != nil {
		return nil, err
	}
	raw, err := r.client.Get(ctx, key).Bytes()
	if errors.Is(err, goredis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var lease db.PublicationLease
	if err := json.Unmarshal(raw, &lease); err != nil {
		return nil, err
	}
	return &lease, nil
}

func (r *RedisDB) RecoverPublicationLease(ctx context.Context, pull models.PullRequest, fence db.PublicationFence) error {
	_, err := r.mutatePublicationLease(ctx, pull, "", func(current *db.PublicationLease, now time.Time) (*db.PublicationLease, error) {
		return nil, db.CanRecoverPublicationLease(current, fence, now)
	})
	return err
}

var _ db.PublicationRecoveryStore = (*RedisDB)(nil)
