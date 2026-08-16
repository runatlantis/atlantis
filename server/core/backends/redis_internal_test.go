// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package backends

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/stretchr/testify/assert"
)

func TestNewRedisClient_SingleNode(t *testing.T) {
	s := miniredis.RunT(t)
	client, err := newRedisClient(valid.RedisBackend{
		Host: s.Host(),
		Port: s.Server().Addr().Port,
	}, logging.NewNoopLogger(t))
	assert.NoError(t, err)
	assert.NotNil(t, client)
}

func TestNewRedisClient_WithUsername(t *testing.T) {
	s := miniredis.RunT(t)
	client, err := newRedisClient(valid.RedisBackend{
		Host:     s.Host(),
		Port:     s.Server().Addr().Port,
		Username: "testuser",
	}, logging.NewNoopLogger(t))
	assert.NoError(t, err)
	assert.NotNil(t, client)
}

func TestNewRedisClient_ClusterEmptyAddresses(t *testing.T) {
	_, err := newRedisClient(valid.RedisBackend{
		ClusterAddresses: []string{"", ""},
	}, logging.NewNoopLogger(t))
	assert.ErrorContains(t, err, "redis cluster addresses provided but all are empty")
}
