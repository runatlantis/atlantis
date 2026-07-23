// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package planstore

import (
	"testing"

	"github.com/runatlantis/atlantis/server/core/backends"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/stretchr/testify/assert"
)

func TestNew_FilesystemBackend(t *testing.T) {
	store, err := New(logging.NewNoopLogger(t), &backends.FilesystemBackend{})
	assert.NoError(t, err)
	assert.IsType(t, &LocalPlanStore{}, store)
}

func TestNew_UnsupportedBackend(t *testing.T) {
	_, err := New(logging.NewNoopLogger(t), &backends.RedisBackend{})
	assert.ErrorContains(t, err, "the plan store has no redis driver")
}

func TestS3PlanPrefix(t *testing.T) {
	assert.Equal(t, "plans", s3PlanPrefix(""))
	assert.Equal(t, "some/team/plans", s3PlanPrefix("some/team"))
	assert.Equal(t, "some/team/plans", s3PlanPrefix("/some/team/"))
}
