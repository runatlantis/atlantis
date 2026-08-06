// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package raw

import (
	"fmt"
	"slices"
	"strings"

	"github.com/runatlantis/atlantis/server/core/config/valid"
)

// Backends is the raw schema for backend connection config, one optional
// block per backend type. The boltdb and filesystem backends work
// without configuration.
type Backends struct {
	BoltDB *BoltDBBackend `yaml:"boltdb,omitempty" json:"boltdb,omitempty"`
	Redis  *RedisBackend  `yaml:"redis,omitempty" json:"redis,omitempty"`
	S3     *S3Backend     `yaml:"s3,omitempty" json:"s3,omitempty"`
}

// BoltDBBackend is the raw schema for BoltDB location config.
type BoltDBBackend struct {
	DataDir string `yaml:"data_dir,omitempty" json:"data_dir,omitempty"`
}

// RedisBackend is the raw schema for Redis connection config.
type RedisBackend struct {
	Host               string   `yaml:"host" json:"host"`
	Port               int      `yaml:"port" json:"port"`
	Password           string   `yaml:"password" json:"password"`
	Username           string   `yaml:"username" json:"username"`
	TLSEnabled         bool     `yaml:"tls_enabled" json:"tls_enabled"`
	InsecureSkipVerify bool     `yaml:"insecure_skip_verify" json:"insecure_skip_verify"`
	DB                 int      `yaml:"db" json:"db"`
	ClusterAddresses   []string `yaml:"cluster_addresses" json:"cluster_addresses"`
}

// S3Backend is the raw schema for S3 connection and location config.
type S3Backend struct {
	Bucket         string `yaml:"bucket" json:"bucket"`
	Region         string `yaml:"region" json:"region"`
	Prefix         string `yaml:"prefix,omitempty" json:"prefix,omitempty"`
	Endpoint       string `yaml:"endpoint" json:"endpoint"`
	ForcePathStyle bool   `yaml:"force_path_style" json:"force_path_style"`
	Profile        string `yaml:"profile" json:"profile"`
}

// Stores is the raw schema for binding stores to backends. Each store
// has its own schema because stores don't share the same options.
type Stores struct {
	Coordination *CoordinationStore `yaml:"coordination,omitempty" json:"coordination,omitempty"`
	Plans        *PlansStore        `yaml:"plans,omitempty" json:"plans,omitempty"`
}

// CoordinationStore is the raw schema for the coordination store binding.
type CoordinationStore struct {
	Backend string `yaml:"backend" json:"backend"`
}

// PlansStore is the raw schema for the plans store binding.
type PlansStore struct {
	Backend string `yaml:"backend" json:"backend"`
}

// storeBackendTypes is the set of backend types each store has a driver for.
var storeBackendTypes = map[string][]string{
	"coordination": {valid.BackendTypeBoltDB, valid.BackendTypeRedis},
	"plans":        {valid.BackendTypeFilesystem, valid.BackendTypeS3},
}

func (b Backends) Validate() error {
	if b.Redis != nil {
		if b.Redis.Host == "" && len(b.Redis.ClusterAddresses) == 0 {
			return fmt.Errorf("backends.redis: host or cluster_addresses is required")
		}
		if b.Redis.Host != "" && len(b.Redis.ClusterAddresses) > 0 {
			return fmt.Errorf("backends.redis: host and cluster_addresses are mutually exclusive")
		}
	}
	if b.S3 != nil {
		if b.S3.Bucket == "" {
			return fmt.Errorf("backends.s3: bucket is required")
		}
		if b.S3.Region == "" {
			return fmt.Errorf("backends.s3: region is required")
		}
	}
	return nil
}

// validate checks each binding against the configured backends and the
// per-store driver support matrix.
func (s Stores) validate(backends Backends) error {
	checkBackend := func(store, backend string) error {
		supported := storeBackendTypes[store]
		if !slices.Contains(supported, backend) {
			return fmt.Errorf("stores.%s: unsupported backend %q, supported backends are [%s]",
				store, backend, strings.Join(supported, ", "))
		}
		switch backend {
		case valid.BackendTypeRedis:
			if backends.Redis == nil {
				return fmt.Errorf("stores.%s: backend 'redis' selected but backends.redis is not configured", store)
			}
		case valid.BackendTypeS3:
			if backends.S3 == nil {
				return fmt.Errorf("stores.%s: backend 's3' selected but backends.s3 is not configured", store)
			}
		}
		return nil
	}
	if s.Coordination != nil {
		if err := checkBackend("coordination", s.Coordination.Backend); err != nil {
			return err
		}
	}
	if s.Plans != nil {
		if err := checkBackend("plans", s.Plans.Backend); err != nil {
			return err
		}
	}
	return nil
}

func (b Backends) ToValid() valid.Backends {
	var v valid.Backends
	if b.BoltDB != nil {
		v.BoltDB = &valid.BoltDBBackend{DataDir: b.BoltDB.DataDir}
	}
	if b.Redis != nil {
		port := b.Redis.Port
		if port == 0 {
			port = 6379
		}
		v.Redis = &valid.RedisBackend{
			Host:               b.Redis.Host,
			Port:               port,
			Password:           b.Redis.Password,
			Username:           b.Redis.Username,
			TLSEnabled:         b.Redis.TLSEnabled,
			InsecureSkipVerify: b.Redis.InsecureSkipVerify,
			DB:                 b.Redis.DB,
			ClusterAddresses:   b.Redis.ClusterAddresses,
		}
	}
	if b.S3 != nil {
		v.S3 = &valid.S3Backend{
			Bucket:         b.S3.Bucket,
			Region:         b.S3.Region,
			Prefix:         b.S3.Prefix,
			Endpoint:       b.S3.Endpoint,
			ForcePathStyle: b.S3.ForcePathStyle,
			Profile:        b.S3.Profile,
		}
	}
	return v
}

func (s Stores) ToValid() valid.Stores {
	var v valid.Stores
	if s.Coordination != nil {
		v.Coordination = &valid.CoordinationStore{Backend: s.Coordination.Backend}
	}
	if s.Plans != nil {
		v.Plans = &valid.PlansStore{Backend: s.Plans.Backend}
	}
	return v
}
