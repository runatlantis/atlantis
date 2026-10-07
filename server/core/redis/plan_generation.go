// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"github.com/runatlantis/atlantis/server/core/db"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
)

// All pull-status mutations use the same compare-and-swap primitive. The
// existence flag distinguishes a missing record from an existing empty blob.
// This operates on one key and is also safe on Redis Cluster.
const compareAndSwapPullStatus = `
local current = redis.call("GET", KEYS[1])
if ARGV[1] == "0" then
 if current then return 0 end
elseif not current or current ~= ARGV[2] then
 return 0
end
redis.call("SET", KEYS[1], ARGV[3])
return 1
`

func (r *RedisDB) BeginPlanGeneration(pull models.PullRequest, generation string, projects []command.ProjectContext, replace bool) (db.PlanGenerationBeginResult, error) {
	var result db.PlanGenerationBeginResult
	_, err := r.mutatePullStatus(pull, true, func(current *models.PullStatus) (models.PullStatus, error) {
		var transitionErr error
		result, transitionErr = db.BeginPlanGeneration(current, pull, generation, projects, replace)
		return result.PullStatus, transitionErr
	})
	return result, err
}

func (r *RedisDB) mutatePullStatus(pull models.PullRequest, allowUnreadable bool, transition func(*models.PullStatus) (models.PullStatus, error)) (models.PullStatus, error) {
	key, err := r.pullKey(pull)
	if err != nil {
		return models.PullStatus{}, err
	}
	opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	// Bound contention retries; this is not a publication-claim wait loop.
	for range 8 {
		raw, readErr := r.client.Get(opCtx, key).Result()
		if readErr != nil && !errors.Is(readErr, goredis.Nil) {
			return models.PullStatus{}, fmt.Errorf("reading pull status: %w", readErr)
		}
		exists := "0"
		var current *models.PullStatus
		if readErr == nil {
			exists = "1"
			var status models.PullStatus
			if decodeErr := json.Unmarshal([]byte(raw), &status); decodeErr != nil {
				if !allowUnreadable && !db.LegacyPullStatus([]byte(raw)) {
					return models.PullStatus{}, fmt.Errorf("reading plan status; run `atlantis plan` to recover: %w", decodeErr)
				}
			} else {
				current = &status
			}
		}
		next, transitionErr := transition(current)
		if transitionErr != nil && !errors.Is(transitionErr, db.ErrApplyExecutionAmbiguous) {
			return models.PullStatus{}, transitionErr
		}
		if current == nil && transitionErr != nil {
			return next, transitionErr
		}
		encoded, err := json.Marshal(next)
		if err != nil {
			return models.PullStatus{}, fmt.Errorf("encoding pull status: %w", err)
		}
		changed, err := r.client.Eval(opCtx, compareAndSwapPullStatus, []string{key}, exists, raw, encoded).Int()
		if err != nil {
			return models.PullStatus{}, fmt.Errorf("updating pull status: %w", err)
		}
		if changed == 1 {
			return next, transitionErr
		}
	}
	return models.PullStatus{}, fmt.Errorf("pull status changed concurrently; retry the command")
}

func (r *RedisDB) DiscardPlanStatus(pull models.PullRequest, expected models.ProjectStatus) (bool, error) {
	var discarded bool
	_, err := r.mutatePullStatus(pull, false, func(current *models.PullStatus) (models.PullStatus, error) {
		next, changed, transitionErr := db.DiscardPlanStatus(current, pull, expected)
		discarded = changed
		return next, transitionErr
	})
	return discarded, err
}
