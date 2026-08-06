// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package backends_test

import (
	"testing"

	"github.com/runatlantis/atlantis/server/core/backends"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/stretchr/testify/assert"
)

func TestResolve_Defaults(t *testing.T) {
	cfg, b, err := backends.Resolve(valid.Backends{}, valid.Stores{}, backends.Legacy{LockingDBType: "boltdb"})
	assert.NoError(t, err)
	assert.Equal(t, backends.BoltDB, b.Coordination)
	assert.Equal(t, backends.Filesystem, b.Plans)
	assert.Nil(t, cfg.Redis)
}

func TestResolve_LegacyRedisFlags(t *testing.T) {
	legacy := backends.Legacy{
		LockingDBType: "redis",
		Redis:         valid.RedisBackend{Host: "flag-host", Port: 6379},
	}
	cfg, b, err := backends.Resolve(valid.Backends{}, valid.Stores{}, legacy)
	assert.NoError(t, err)
	assert.Equal(t, backends.Redis, b.Coordination)
	assert.Equal(t, "flag-host", cfg.Redis.Host)
}

func TestResolve_ConfigWinsOverFlags(t *testing.T) {
	yamlRedis := &valid.RedisBackend{Host: "yaml-host", Port: 6379}
	legacy := backends.Legacy{
		LockingDBType: "redis",
		Redis:         valid.RedisBackend{Host: "flag-host", Port: 6379},
	}

	// The backends block overrides flag-derived connection settings.
	cfg, b, err := backends.Resolve(valid.Backends{Redis: yamlRedis}, valid.Stores{}, legacy)
	assert.NoError(t, err)
	assert.Equal(t, backends.Redis, b.Coordination)
	assert.Equal(t, "yaml-host", cfg.Redis.Host)

	// An explicit stores binding overrides --locking-db-type.
	stores := valid.Stores{Coordination: &valid.CoordinationStore{Backend: "boltdb"}}
	_, b, err = backends.Resolve(valid.Backends{}, stores, legacy)
	assert.NoError(t, err)
	assert.Equal(t, backends.BoltDB, b.Coordination)
}

func TestResolve_LegacyRedisWithoutHost(t *testing.T) {
	_, _, err := backends.Resolve(valid.Backends{}, valid.Stores{}, backends.Legacy{LockingDBType: "redis"})
	assert.ErrorContains(t, err, "requires --redis-host or --redis-cluster-addresses")
}

func TestResolve_UnknownLockingDBType(t *testing.T) {
	_, _, err := backends.Resolve(valid.Backends{}, valid.Stores{}, backends.Legacy{LockingDBType: "etcd"})
	assert.ErrorContains(t, err, "unsupported --locking-db-type")
}

func TestResolve_PlansBinding(t *testing.T) {
	stores := valid.Stores{
		Plans: &valid.PlansStore{Backend: "s3"},
	}
	_, b, err := backends.Resolve(valid.Backends{S3: &valid.S3Backend{Bucket: "b", Region: "us-east-1"}}, stores, backends.Legacy{})
	assert.NoError(t, err)
	assert.Equal(t, backends.S3, b.Plans)
}
