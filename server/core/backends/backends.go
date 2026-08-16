// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

// Package backends constructs storage backend clients independently of
// the stores (locks, plans, ...) that use them. A backend is
// connection-level: it knows how to dial, not what data it holds.
package backends

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	goredis "github.com/redis/go-redis/v9"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/logging"
	bolt "go.etcd.io/bbolt"
)

// Kind identifies a backend type.
type Kind string

const (
	BoltDB     Kind = valid.BackendTypeBoltDB
	Filesystem Kind = valid.BackendTypeFilesystem
	Redis      Kind = valid.BackendTypeRedis
	S3         Kind = valid.BackendTypeS3
)

// Backend is a constructed backend handle: the client for its type,
// ready for a store driver to layer its data layout on. Store factories
// type-switch on the concrete types below.
type Backend interface {
	Kind() Kind
}

// BoltDBBackend holds an opened bolt database file.
type BoltDBBackend struct {
	DB *bolt.DB
}

func (*BoltDBBackend) Kind() Kind { return BoltDB }

// FilesystemBackend is the server's working directory on local disk; it
// needs no client.
type FilesystemBackend struct{}

func (*FilesystemBackend) Kind() Kind { return Filesystem }

// RedisBackend holds a connected Redis client (single-node or cluster).
type RedisBackend struct {
	Client goredis.Cmdable
}

func (*RedisBackend) Kind() Kind { return Redis }

// S3Backend holds a constructed S3 client and the bucket and prefix all
// stores on this backend live under.
type S3Backend struct {
	Client *s3.Client
	Bucket string
	Prefix string
}

func (*S3Backend) Kind() Kind { return S3 }

// Registry constructs backends on first use and caches them, so a
// configured-but-unbound backend is never dialed.
type Registry struct {
	logger  logging.SimpleLogging
	dataDir string
	cfg     valid.Backends
	cache   map[Kind]Backend
}

func NewRegistry(logger logging.SimpleLogging, dataDir string, cfg valid.Backends) *Registry {
	return &Registry{
		logger:  logger,
		dataDir: dataDir,
		cfg:     cfg,
		cache:   map[Kind]Backend{},
	}
}

func (r *Registry) Get(kind Kind) (Backend, error) {
	if b, ok := r.cache[kind]; ok {
		return b, nil
	}
	b, err := r.build(kind)
	if err != nil {
		return nil, fmt.Errorf("initializing %s backend: %w", kind, err)
	}
	r.cache[kind] = b
	return b, nil
}

func (r *Registry) build(kind Kind) (Backend, error) {
	switch kind {
	case BoltDB:
		dataDir := r.dataDir
		if r.cfg.BoltDB != nil && r.cfg.BoltDB.DataDir != "" {
			dataDir = r.cfg.BoltDB.DataDir
		}
		db, err := openBoltDB(dataDir)
		if err != nil {
			return nil, err
		}
		return &BoltDBBackend{DB: db}, nil
	case Filesystem:
		return &FilesystemBackend{}, nil
	case Redis:
		if r.cfg.Redis == nil {
			return nil, fmt.Errorf("backends.redis is not configured in the server-side repo config")
		}
		client, err := newRedisClient(*r.cfg.Redis, r.logger)
		if err != nil {
			return nil, err
		}
		return &RedisBackend{Client: client}, nil
	case S3:
		if r.cfg.S3 == nil {
			return nil, fmt.Errorf("backends.s3 is not configured in the server-side repo config")
		}
		client, err := newS3Client(*r.cfg.S3)
		if err != nil {
			return nil, err
		}
		return &S3Backend{Client: client, Bucket: r.cfg.S3.Bucket, Prefix: r.cfg.S3.Prefix}, nil
	default:
		return nil, fmt.Errorf("unknown backend type %q", kind)
	}
}
