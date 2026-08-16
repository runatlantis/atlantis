// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package raw_test

import (
	"testing"

	"github.com/runatlantis/atlantis/server/core/config/raw"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/stretchr/testify/assert"
)

func TestBackendsStores_Unmarshal(t *testing.T) {
	rawYaml := `
backends:
  redis:
    host: redis.internal
    port: 6380
    tls_enabled: true
  s3:
    bucket: my-bucket
    region: us-east-1
    prefix: some/prefix
stores:
  coordination:
    backend: redis
  plans:
    backend: s3
`

	var result raw.GlobalCfg
	err := unmarshalString(rawYaml, &result)
	assert.NoError(t, err)
	assert.NoError(t, result.Validate())

	v := result.ToValid(valid.NewGlobalCfgFromArgs(valid.GlobalCfgArgs{}))
	assert.Equal(t, "redis.internal", v.Backends.Redis.Host)
	assert.Equal(t, 6380, v.Backends.Redis.Port)
	assert.True(t, v.Backends.Redis.TLSEnabled)
	assert.Equal(t, "us-east-1", v.Backends.S3.Region)
	assert.Equal(t, "my-bucket", v.Backends.S3.Bucket)
	assert.Equal(t, "some/prefix", v.Backends.S3.Prefix)
	assert.Equal(t, "redis", v.Stores.Coordination.Backend)
	assert.Equal(t, "s3", v.Stores.Plans.Backend)
}

func TestBackends_RedisPortDefault(t *testing.T) {
	b := raw.Backends{Redis: &raw.RedisBackend{Host: "h"}}
	assert.NoError(t, b.Validate())
	assert.Equal(t, 6379, b.ToValid().Redis.Port)
}

func TestBackends_Validate(t *testing.T) {
	cases := []struct {
		description string
		input       raw.Backends
		errContains string
	}{
		{
			description: "empty is valid",
			input:       raw.Backends{},
		},
		{
			description: "redis needs host or cluster addresses",
			input:       raw.Backends{Redis: &raw.RedisBackend{}},
			errContains: "host or cluster_addresses",
		},
		{
			description: "redis host and cluster addresses conflict",
			input:       raw.Backends{Redis: &raw.RedisBackend{Host: "h", ClusterAddresses: []string{"a:1"}}},
			errContains: "mutually exclusive",
		},
		{
			description: "s3 needs bucket",
			input:       raw.Backends{S3: &raw.S3Backend{Region: "us-east-1"}},
			errContains: "bucket is required",
		},
		{
			description: "s3 needs region",
			input:       raw.Backends{S3: &raw.S3Backend{Bucket: "b"}},
			errContains: "region is required",
		},
	}
	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			err := c.input.Validate()
			if c.errContains == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, c.errContains)
			}
		})
	}
}

func TestStores_RejectUnknownKeys(t *testing.T) {
	// Neither store binding has a prefix; it lives on backends.s3, and
	// strict decoding rejects it on a binding.
	for _, input := range []string{`
stores:
  coordination:
    backend: redis
    prefix: locks/
`, `
stores:
  plans:
    backend: s3
    prefix: p
`} {
		var cfg raw.GlobalCfg
		err := unmarshalString(input, &cfg)
		assert.ErrorContains(t, err, "prefix")
	}
}

func TestStores_Validate(t *testing.T) {
	redisBackends := `
backends:
  redis:
    host: h
`
	cases := []struct {
		description string
		input       string
		errContains string
	}{
		{
			description: "coordination on configured redis",
			input: redisBackends + `
stores:
  coordination:
    backend: redis
`,
		},
		{
			description: "coordination rejects s3",
			input: `
stores:
  coordination:
    backend: s3
`,
			errContains: "stores.coordination: unsupported backend \"s3\"",
		},
		{
			description: "binding requires the backend block",
			input: `
stores:
  coordination:
    backend: redis
`,
			errContains: "backends.redis is not configured",
		},
	}
	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			var cfg raw.GlobalCfg
			assert.NoError(t, unmarshalString(c.input, &cfg))
			err := cfg.Validate()
			if c.errContains == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, c.errContains)
			}
		})
	}
}
