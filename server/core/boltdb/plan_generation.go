// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package boltdb

import (
	"errors"
	"fmt"

	"github.com/runatlantis/atlantis/server/core/db"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	bolt "go.etcd.io/bbolt"
)

func (b *BoltDB) BeginPlanGeneration(pull models.PullRequest, generation string, projects []command.ProjectContext, replace bool) (db.PlanGenerationBeginResult, error) {
	var result db.PlanGenerationBeginResult
	_, err := b.mutatePullStatus(pull, true, func(current *models.PullStatus) (models.PullStatus, error) {
		var transitionErr error
		result, transitionErr = db.BeginPlanGeneration(current, pull, generation, projects, replace)
		return result.PullStatus, transitionErr
	})
	return result, err
}

func (b *BoltDB) mutatePullStatus(pull models.PullRequest, allowUnreadable bool, transition func(*models.PullStatus) (models.PullStatus, error)) (models.PullStatus, error) {
	key, err := b.pullKey(pull)
	if err != nil {
		return models.PullStatus{}, err
	}
	var next models.PullStatus
	var transitionErr error
	err = b.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(b.pullsBucketName)
		current, readErr := b.getPullFromBucket(bucket, key)
		if readErr != nil {
			if !allowUnreadable && !db.LegacyPullStatus(bucket.Get(key)) {
				return fmt.Errorf("reading plan status; run `atlantis plan` to recover: %w", readErr)
			}
			current = nil
		}
		next, transitionErr = transition(current)
		if transitionErr != nil && !errors.Is(transitionErr, db.ErrApplyExecutionAmbiguous) {
			return transitionErr
		}
		if current == nil && transitionErr != nil {
			return nil
		}
		return b.writePullToBucket(bucket, key, next)
	})
	if err != nil {
		return models.PullStatus{}, fmt.Errorf("updating pull status: %w", err)
	}
	return next, transitionErr
}

func (b *BoltDB) DiscardPlanStatus(pull models.PullRequest, expected models.ProjectStatus) (bool, error) {
	var discarded bool
	_, err := b.mutatePullStatus(pull, false, func(current *models.PullStatus) (models.PullStatus, error) {
		next, changed, transitionErr := db.DiscardPlanStatus(current, pull, expected)
		discarded = changed
		return next, transitionErr
	})
	return discarded, err
}
