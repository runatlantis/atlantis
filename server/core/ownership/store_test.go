// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package ownership_test

import (
	"testing"

	"github.com/runatlantis/atlantis/server/core/ownership"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/stretchr/testify/require"
)

func TestKey_CanonicalNormalizesHostname(t *testing.T) {
	key := ownership.NewKey(models.Repo{
		FullName: "Owner/Repo",
		VCSHost:  models.VCSHost{Hostname: " GitHub.COM "},
	}, 42)

	canonical, err := key.Canonical()
	require.NoError(t, err)
	require.Equal(t, "github.com\x00Owner/Repo\x0042", canonical)
}

func TestKey_CanonicalSeparatesRepositoriesAndPulls(t *testing.T) {
	keys := []ownership.Key{
		{VCSHostname: "github.com", RepoFullName: "owner/repo", PullNum: 1},
		{VCSHostname: "github.com", RepoFullName: "owner/other", PullNum: 1},
		{VCSHostname: "github.com", RepoFullName: "owner/repo", PullNum: 2},
	}

	canonical := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		value, err := key.Canonical()
		require.NoError(t, err)
		canonical[value] = struct{}{}
	}
	require.Len(t, canonical, len(keys))
}

func TestKey_CanonicalRejectsIncompleteIdentity(t *testing.T) {
	tests := []ownership.Key{
		{RepoFullName: "owner/repo", PullNum: 1},
		{VCSHostname: "github.com", PullNum: 1},
		{VCSHostname: "github.com", RepoFullName: "owner/repo"},
		{VCSHostname: "github.com", RepoFullName: "owner/repo", PullNum: -1},
	}

	for _, key := range tests {
		_, err := key.Canonical()
		require.Error(t, err)
	}
}
