// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"testing"

	"github.com/runatlantis/atlantis/server/events/models"
)

func TestIsUnsafeNonPRRef(t *testing.T) {
	for _, tc := range []struct {
		name string
		ref  string
		want bool
	}{
		{name: "GitHub pull head", ref: "pull/123/head", want: true},
		{name: "GitHub refs pull merge", ref: "refs/pull/123/merge", want: true},
		{name: "GitHub plus refs pull head", ref: "+refs/pull/123/head", want: true},
		{name: "GitLab merge request head", ref: "merge-requests/123/head", want: true},
		{name: "GitLab refs merge request merge", ref: "refs/merge-requests/123/merge", want: true},
		{name: "GitLab plus refs merge request merge", ref: "+refs/merge-requests/123/merge", want: true},
		{name: "GitHub pull refspec", ref: "refs/pull/123/head:refs/tmp/x", want: true},
		{name: "GitHub short pull refspec", ref: "pull/123/head:refs/tmp/x", want: true},
		{name: "GitLab merge request refspec", ref: "refs/merge-requests/123/head:refs/tmp/x", want: true},
		{name: "branch", ref: "main", want: false},
		{name: "tag", ref: "refs/tags/v1.2.3", want: false},
		{name: "sha", ref: "0123456789abcdef0123456789abcdef01234567", want: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := models.IsUnsafeAPIRef(tc.ref); got != tc.want {
				t.Fatalf("models.IsUnsafeAPIRef(%q) = %t, want %t", tc.ref, got, tc.want)
			}
		})
	}
}

func TestUsesPRSourceRemote(t *testing.T) {
	github := models.Repo{VCSHost: models.VCSHost{Type: models.Github}}
	gitlab := models.Repo{VCSHost: models.VCSHost{Type: models.Gitlab}}
	for _, tc := range []struct {
		name          string
		head          models.Repo
		num           int
		appEnabled    bool
		checkoutMerge bool
		want          bool
	}{
		// GitHub PR: the source (fork) remote is skipped in favour of
		// pull/<n>/head from origin when either the App path or the merge
		// checkout strategy is used.
		{name: "github token+branch uses source remote", head: github, num: 1, want: true},
		{name: "github app skips source remote", head: github, num: 1, appEnabled: true, want: false},
		{name: "github merge skips source remote", head: github, num: 1, checkoutMerge: true, want: false},
		{name: "github app+merge skips source remote", head: github, num: 1, appEnabled: true, checkoutMerge: true, want: false},
		// Non-GitHub always uses the source remote (no pull ref on origin).
		{name: "gitlab merge still uses source remote", head: gitlab, num: 1, checkoutMerge: true, want: true},
		// API-driven runs (num <= 0) have no pull ref, so keep the source remote.
		{name: "github merge but non-PR uses source remote", head: github, num: -1, checkoutMerge: true, appEnabled: true, want: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := &FileWorkspace{GithubAppEnabled: tc.appEnabled, CheckoutMerge: tc.checkoutMerge}
			if got := w.usesPRSourceRemote(tc.head, models.PullRequest{Num: tc.num}); got != tc.want {
				t.Fatalf("usesPRSourceRemote() = %t, want %t", got, tc.want)
			}
		})
	}
}
