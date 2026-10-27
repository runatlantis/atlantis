// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.

package events_test

import (
	"testing"

	"github.com/runatlantis/atlantis/server/events"
	. "github.com/runatlantis/atlantis/testing"
)

func TestRepoAllowlistChecker_IsAllowlisted(t *testing.T) {
	cases := []struct {
		Description  string
		Allowlist    string
		RepoFullName string
		Hostname     string
		Exp          bool
	}{
		{
			"exact match",
			"github.com/owner/repo",
			"owner/repo",
			"github.com",
			true,
		},
		{
			"exact match shouldn't match anything else",
			"github.com/owner/repo",
			"owner/rep",
			"github.com",
			false,
		},
		{
			"* should match anything",
			"*",
			"owner/repo",
			"github.com",
			true,
		},
		{
			"github.com* should match anything github",
			"github.com*",
			"owner/repo",
			"github.com",
			true,
		},
		{
			"github.com* should not match gitlab",
			"github.com*",
			"owner/repo",
			"gitlab.com",
			false,
		},
		{
			"github.com/o* should match",
			"github.com/o*",
			"owner/repo",
			"github.com",
			true,
		},
		{
			"github.com/owner/rep* should not match",
			"github.com/owner/rep*",
			"owner/re",
			"github.com",
			false,
		},
		{
			"github.com/owner/rep* should match",
			"github.com/owner/rep*",
			"owner/rep",
			"github.com",
			true,
		},
		{
			"github.com/o* should not match",
			"github.com/o*",
			"somethingelse/repo",
			"github.com",
			false,
		},
		{
			"github.com/owner/repo* should match exactly",
			"github.com/owner/repo*",
			"owner/repo",
			"github.com",
			true,
		},
		{
			"github.com/owner/* should match anything in org",
			"github.com/owner/*",
			"owner/repo",
			"github.com",
			true,
		},
		{
			"github.com/owner/* should not match anything not in org",
			"github.com/owner/*",
			"otherorg/repo",
			"github.com",
			false,
		},
		{
			"if there's any * it should match",
			"github.com/owner/repo,*",
			"otherorg/repo",
			"github.com",
			true,
		},
		{
			"any exact match should match",
			"github.com/owner/repo,github.com/otherorg/repo",
			"otherorg/repo",
			"github.com",
			true,
		},
		{
			"longer shouldn't match on exact",
			"github.com/owner/repo",
			"owner/repo-longer",
			"github.com",
			false,
		},
		{
			"should be case insensitive",
			"github.com/owner/repo",
			"OwNeR/rEpO",
			"github.com",
			true,
		},
		{
			"should be case insensitive for wildcards",
			"github.com/owner/*",
			"OwNeR/rEpO",
			"github.com",
			true,
		},
		{
			"should match if wildcard is not last character",
			"github.com/owner/*-repo",
			"owner/prefix-repo",
			"github.com",
			true,
		},
		{
			"should match if wildcard is first character within owner name",
			"github.com/*-owner/repo",
			"prefix-owner/repo",
			"github.com",
			true,
		},
		{
			"should match if wildcard is at beginning",
			"*-owner/repo",
			"prefix-owner/repo",
			"github.com",
			true,
		},
		{
			"should match with duplicate",
			"*runatlantis",
			"runatlantis/runatlantis",
			"github.com",
			true,
		},
		{
			// A rule with a wildcard in the middle must match the part before
			// the wildcard as well. Matching only the suffix allows any
			// repository, in any organisation, whose name ends the right way.
			"wildcard in the middle must still anchor the prefix",
			"github.com/myorg/*-prod",
			"evil/anything-prod",
			"github.com",
			false,
		},
		{
			"wildcard in the middle matches the intended org",
			"github.com/myorg/*-prod",
			"myorg/app-prod",
			"github.com",
			true,
		},
		{
			"prefix before a trailing wildcard must still anchor",
			"github.com/myorg/app*",
			"evil/app-clone",
			"github.com",
			false,
		},
		{
			"multiple wildcards are matched in order",
			"github.com/*/team-*",
			"someorg/team-alpha",
			"github.com",
			true,
		},
		{
			"multiple wildcards do not match when a literal segment differs",
			"github.com/*/team-*",
			"someorg/group-alpha",
			"github.com",
			false,
		},
		{
			// The host is part of the anchored prefix, so a rule written for
			// one VCS host must not admit the same owner and name on another.
			"wildcard in the middle does not match a different host",
			"github.com/myorg/*-prod",
			"myorg/app-prod",
			"gitlab.com",
			false,
		},
		{
			"trailing wildcard does not match a different host",
			"github.com/myorg/app*",
			"myorg/app-clone",
			"gitlab.com",
			false,
		},
		{
			// A negated rule goes through the same matcher, so a wildcard in
			// the middle of it must anchor the prefix in the same way.
			"negated rule with a wildcard in the middle excludes the intended repo",
			"github.com/myorg/*,!github.com/myorg/*-prod",
			"myorg/app-prod",
			"github.com",
			false,
		},
		{
			"negated rule with a wildcard in the middle does not exclude other repos",
			"github.com/myorg/*,!github.com/myorg/*-prod",
			"myorg/app-dev",
			"github.com",
			true,
		},
		{
			"negated rule with a wildcard in the middle only excludes its own org",
			"github.com/*/*,!github.com/myorg/*-prod",
			"otherorg/app-prod",
			"github.com",
			true,
		},
		{
			"should exclude with negative match",
			"github.com/owner/*,!github.com/owner/badrepo",
			"owner/badrepo",
			"github.com",
			false,
		},
		{
			"should match if with negative rule doesn't match",
			"github.com/owner/*,!github.com/owner/badrepo",
			"owner/otherrepo",
			"github.com",
			true,
		},
	}

	for _, c := range cases {
		t.Run(c.Description, func(t *testing.T) {
			w, err := events.NewRepoAllowlistChecker(c.Allowlist)
			Ok(t, err)
			Equals(t, c.Exp, w.IsAllowlisted(c.RepoFullName, c.Hostname))
		})
	}
}

// If the allowlist contains a schema then we should get an error.
func TestRepoAllowlistChecker_ContainsSchema(t *testing.T) {
	cases := []struct {
		allowlist string
		expErr    string
	}{
		{
			"://",
			`allowlist "://" contained ://`,
		},
		{
			"valid/*,https://bitbucket.org/*",
			`allowlist "https://bitbucket.org/*" contained ://`,
		},
	}

	for _, c := range cases {
		t.Run(c.allowlist, func(t *testing.T) {
			_, err := events.NewRepoAllowlistChecker(c.allowlist)
			ErrEquals(t, c.expErr, err)
		})
	}
}
