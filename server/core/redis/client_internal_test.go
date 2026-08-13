// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package redis

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
	redislib "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestNewWithConfig_EnablesContextTimeouts(t *testing.T) {
	mr := miniredis.RunT(t)
	database, err := NewWithConfig(Config{
		Hostname: mr.Host(),
		Port:     mr.Server().Addr().Port,
	})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, database.Close()) })

	client, ok := database.client.(*redislib.Client)
	require.True(t, ok)
	require.True(t, client.Options().ContextTimeoutEnabled)
}
