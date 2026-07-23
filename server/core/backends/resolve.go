// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package backends

import (
	"fmt"

	"github.com/runatlantis/atlantis/server/core/config/valid"
)

// Legacy carries the flag-based backend configuration that predates the
// backends/stores config.
//
// Deprecated: remove together with --locking-db-type and --redis-*.
type Legacy struct {
	LockingDBType string
	Redis         valid.RedisBackend
}

// Bindings is the resolved backend selection for each store.
type Bindings struct {
	Coordination Kind
	Plans        Kind
}

// Resolve merges the backends/stores config with legacy flags. An
// explicit stores binding wins; legacy flags fill in only where no
// binding is set, and a backends block from config wins over flag-derived
// connection settings.
func Resolve(cfg valid.Backends, stores valid.Stores, legacy Legacy) (valid.Backends, Bindings, error) {
	var b Bindings

	switch {
	case stores.Coordination != nil:
		b.Coordination = Kind(stores.Coordination.Backend)
	case legacy.LockingDBType == "redis":
		if cfg.Redis == nil {
			if legacy.Redis.Host == "" && len(legacy.Redis.ClusterAddresses) == 0 {
				return cfg, b, fmt.Errorf("--locking-db-type=redis requires --redis-host or --redis-cluster-addresses")
			}
			rc := legacy.Redis
			cfg.Redis = &rc
		}
		b.Coordination = Redis
	case legacy.LockingDBType == "boltdb" || legacy.LockingDBType == "":
		b.Coordination = BoltDB
	default:
		return cfg, b, fmt.Errorf("unsupported --locking-db-type %q", legacy.LockingDBType)
	}

	if stores.Plans != nil {
		b.Plans = Kind(stores.Plans.Backend)
	} else {
		b.Plans = Filesystem
	}

	return cfg, b, nil
}
