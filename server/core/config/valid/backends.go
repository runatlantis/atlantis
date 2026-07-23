// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package valid

// Backend type names accepted in the server-side repo config. "boltdb"
// and "filesystem" live under the server's data-dir and work without
// configuration.
const (
	BackendTypeBoltDB     = "boltdb"
	BackendTypeFilesystem = "filesystem"
	BackendTypeRedis      = "redis"
	BackendTypeS3         = "s3"
)

// Backends holds connection-level configuration for each backend type,
// one instance per type, with no knowledge of which store uses it.
type Backends struct {
	BoltDB *BoltDBBackend
	Redis  *RedisBackend
	S3     *S3Backend
}

// BoltDBBackend holds location config for the BoltDB backend.
type BoltDBBackend struct {
	// DataDir overrides the server data-dir as the location of the
	// atlantis.db file.
	DataDir string
}

// RedisBackend holds connection config for the Redis backend.
type RedisBackend struct {
	Host               string
	Port               int
	Password           string
	Username           string
	TLSEnabled         bool
	InsecureSkipVerify bool
	DB                 int
	// ClusterAddresses switches the client to cluster mode when set.
	ClusterAddresses []string
}

// S3Backend holds connection and location config for the S3 backend.
// Prefix namespaces all Atlantis objects within the bucket; each store
// additionally appends its own segment (e.g. "plans/"), so stores
// sharing the backend cannot collide.
type S3Backend struct {
	Bucket         string
	Region         string
	Prefix         string
	Endpoint       string
	ForcePathStyle bool
	Profile        string
}

// Stores binds each store (function) to a backend type. Each store has
// its own config struct because stores don't share the same options.
type Stores struct {
	Coordination *CoordinationStore
	Plans        *PlansStore
}

// CoordinationStore selects the backend for locks, the global apply
// lock, and pull status.
type CoordinationStore struct {
	Backend string
}

// PlansStore selects the backend for plan artifacts.
type PlansStore struct {
	Backend string
}
