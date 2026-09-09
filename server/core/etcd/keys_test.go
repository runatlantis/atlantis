// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.

package etcd_test

import (
	"strings"
	"testing"

	"github.com/runatlantis/atlantis/server/core/etcd"
	. "github.com/runatlantis/atlantis/testing"
)

func TestNewKeyspace_Normalization(t *testing.T) {
	cases := map[string]string{
		"/atlantis":         "/atlantis/v1",
		"atlantis":          "/atlantis/v1",
		"/atlantis/":        "/atlantis/v1",
		"//atlantis//prod/": "/atlantis/prod/v1",
		"":                  "/atlantis/v1",
	}
	for in, want := range cases {
		Equals(t, want, etcd.NewKeyspace(in).Root())
	}
}

func TestKeys_UnderRoot(t *testing.T) {
	k := etcd.NewKeyspace("/atlantis")
	ps := etcd.PullScope{VCSHostname: "github.com", Repository: "o/r", PullNum: 7}
	proj := etcd.ProjectScope{VCSHostname: "github.com", Repository: "o/r", Path: ".", Workspace: "default"}
	keys := []string{
		k.SchemaKey(), k.DeploymentKey(), k.ProjectLockKey(proj),
		k.PullStatusKey(ps), k.OwnershipKey(ps), k.GlobalLockKey("apply"),
		k.CommandKey("epoch1", "delivery1"), k.ExecutionBarrierKey(ps, "gen1", "exec1"),
	}
	for _, key := range keys {
		Assert(t, strings.HasPrefix(key, k.Root()+"/"), "key %q not under root", key)
	}
	Assert(t, strings.HasPrefix(k.ProjectLockKey(proj), k.ProjectLockPrefix()), "project lock not under its prefix")
}

// TestEncode_HostnameDisambiguation proves two VCS hosts with identical
// repository names and pull numbers produce distinct keys (design §968), the
// gap the legacy models.GenerateLockKey leaves open.
func TestEncode_HostnameDisambiguation(t *testing.T) {
	k := etcd.NewKeyspace("/atlantis")
	a := etcd.PullScope{VCSHostname: "github.com", Repository: "o/r", PullNum: 1}
	b := etcd.PullScope{VCSHostname: "gitlab.com", Repository: "o/r", PullNum: 1}
	Assert(t, k.OwnershipKey(a) != k.OwnershipKey(b), "distinct hosts must not collide")

	pa := etcd.ProjectScope{VCSHostname: "github.com", Repository: "o/r", Path: ".", Workspace: "default"}
	pb := etcd.ProjectScope{VCSHostname: "gitlab.com", Repository: "o/r", Path: ".", Workspace: "default"}
	Assert(t, k.ProjectLockKey(pa) != k.ProjectLockKey(pb), "distinct hosts must not collide")
}

// TestEncode_Injective proves the canonical encoding is delimiter-safe: fields
// whose naive concatenation would collide still produce distinct keys.
func TestEncode_Injective(t *testing.T) {
	k := etcd.NewKeyspace("/atlantis")
	// "a/b" workspace "" vs "a" workspace "b" would collide under naive "/" join.
	x := etcd.ProjectScope{Repository: "a/b", Workspace: ""}
	y := etcd.ProjectScope{Repository: "a", Workspace: "b"}
	Assert(t, k.ProjectLockKey(x) != k.ProjectLockKey(y), "length-prefixed encoding must be injective")

	same := etcd.ProjectScope{Repository: "a/b", Workspace: ""}
	Equals(t, k.ProjectLockKey(x), k.ProjectLockKey(same))
}
