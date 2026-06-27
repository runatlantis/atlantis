// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import "testing"

func TestIsUnsafeNonPRRef(t *testing.T) {
	for _, tc := range []struct {
		name string
		ref  string
		want bool
	}{
		{name: "GitHub pull head", ref: "pull/123/head", want: true},
		{name: "GitHub refs pull merge", ref: "refs/pull/123/merge", want: true},
		{name: "GitLab merge request head", ref: "merge-requests/123/head", want: true},
		{name: "GitLab refs merge request merge", ref: "refs/merge-requests/123/merge", want: true},
		{name: "GitHub pull refspec", ref: "refs/pull/123/head:refs/tmp/x", want: true},
		{name: "GitHub short pull refspec", ref: "pull/123/head:refs/tmp/x", want: true},
		{name: "GitLab merge request refspec", ref: "refs/merge-requests/123/head:refs/tmp/x", want: true},
		{name: "branch", ref: "main", want: false},
		{name: "tag", ref: "refs/tags/v1.2.3", want: false},
		{name: "sha", ref: "0123456789abcdef0123456789abcdef01234567", want: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := isUnsafeNonPRRef(tc.ref); got != tc.want {
				t.Fatalf("isUnsafeNonPRRef(%q) = %t, want %t", tc.ref, got, tc.want)
			}
		})
	}
}
