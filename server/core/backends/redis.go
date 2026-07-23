// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package backends

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"strings"

	goredis "github.com/redis/go-redis/v9"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/logging"
)

// newRedisClient dials Redis and verifies connectivity with a ping. It
// automatically selects the client type: cluster mode when
// ClusterAddresses is set, single-node mode otherwise.
func newRedisClient(cfg valid.RedisBackend, logger logging.SimpleLogging) (goredis.Cmdable, error) {
	var tlsConfig *tls.Config
	if cfg.TLSEnabled {
		tlsConfig = &tls.Config{
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: cfg.InsecureSkipVerify, //nolint:gosec // In some cases, users may want to use this at their own caution
		}
	}

	var rdb goredis.Cmdable
	var connDesc string
	switch {
	case len(cfg.ClusterAddresses) > 0:
		// Filter out empty addresses
		var addrs []string
		for _, addr := range cfg.ClusterAddresses {
			trimmed := strings.TrimSpace(addr)
			if trimmed != "" {
				addrs = append(addrs, trimmed)
			}
		}
		if len(addrs) == 0 {
			return nil, errors.New("redis cluster addresses provided but all are empty")
		}
		logger.Info("connecting to Redis in cluster mode, addresses: %s", strings.Join(addrs, ", "))
		rdb = goredis.NewClusterClient(&goredis.ClusterOptions{
			Addrs:     addrs,
			Username:  cfg.Username,
			Password:  cfg.Password,
			TLSConfig: tlsConfig,
		})
		connDesc = fmt.Sprintf("cluster nodes %s", strings.Join(addrs, ", "))
	default:
		address := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
		logger.Info("connecting to Redis in single-node mode, host: %s, port: %d", cfg.Host, cfg.Port)
		rdb = goredis.NewClient(&goredis.Options{
			Addr:      address,
			Username:  cfg.Username,
			Password:  cfg.Password,
			DB:        cfg.DB,
			TLSConfig: tlsConfig,
		})
		connDesc = address
	}

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis at %s: %w", connDesc, err)
	}

	return rdb, nil
}
