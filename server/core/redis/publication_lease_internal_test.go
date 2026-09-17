// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package redis

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPublicationLeaseKey_PreservesRedisClusterSlot(t *testing.T) {
	slot := func(key string) uint16 {
		if start := strings.IndexByte(key, '{'); start >= 0 {
			if end := strings.IndexByte(key[start+1:], '}'); end > 0 {
				key = key[start+1 : start+1+end]
			}
		}
		return publicationCRC16(key) & 16383
	}
	require.Equal(t, uint16(0x31c3), publicationCRC16("123456789"), "CRC16/XMODEM check vector")
	for _, key := range []string{"github.com::owner/repo::1", "host::{owner}/repo::1", "host::owner/{}/repo::1", "host::owner/{repo::1", "host::owner/repo}::1", "host::{{owner}}/repo::1"} {
		t.Run(key, func(t *testing.T) {
			leaseKey, err := publicationLeaseKey(key)
			require.NoError(t, err)
			require.NotEqual(t, key, leaseKey)
			require.Equal(t, slot(key), slot(leaseKey))
			again, err := publicationLeaseKey(key)
			require.NoError(t, err)
			require.Equal(t, leaseKey, again)
		})
	}
	// Prove the bounded fallback can represent every slot, without testing
	// time-dependent searches or assuming a repository-name restriction.
	var represented [16384]bool
	for candidate := range 1 << 16 {
		tag := [2]byte{byte((candidate >> 8) & 255), byte(candidate & 255)}
		if tag[0] != '{' && tag[0] != '}' && tag[1] != '{' && tag[1] != '}' {
			represented[publicationCRC16(string(tag[:]))&16383] = true
		}
	}
	for i, found := range represented {
		require.True(t, found, "slot %d must have a tag", i)
	}
}
