// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.

package events

import (
	"testing"

	"github.com/runatlantis/atlantis/server/events/models"
)

func TestWorkingDirLockMetadata(t *testing.T) {
	const sha = "0123456789abcdef0123456789abcdef01234567"
	tests := []struct {
		name     string
		hostType models.VCSHostType
		cloneURL string
		wantURL  string
	}{
		{
			name:     "GitHub strips URL-escaped credentials and git suffix",
			hostType: models.Github,
			cloneURL: "https://user:p%40ss%2Fw%3Ford%23@github.com/owner/repo.git",
			wantURL:  "https://github.com/owner/repo/commit/" + sha,
		},
		{
			name:     "GitLab uses canonical commit path and preserves host subpath",
			hostType: models.Gitlab,
			cloneURL: "https://gitlab.example.com/gitlab/group/repo.git",
			wantURL:  "https://gitlab.example.com/gitlab/group/repo/-/commit/" + sha,
		},
		{
			name:     "Gitea",
			hostType: models.Gitea,
			cloneURL: "https://gitea.example.com/owner/repo.git",
			wantURL:  "https://gitea.example.com/owner/repo/commit/" + sha,
		},
		{
			name:     "unsupported provider falls back to SHA",
			hostType: models.BitbucketCloud,
			cloneURL: "https://bitbucket.org/owner/repo.git",
		},
		{
			name:     "non-HTTP clone URL falls back to SHA",
			hostType: models.Github,
			cloneURL: "git@github.com:owner/repo.git",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata := WorkingDirLockMetadataForPull(models.PullRequest{
				HeadCommit: sha,
				BaseRepo: models.Repo{
					CloneURL: tt.cloneURL,
					VCSHost:  models.VCSHost{Type: tt.hostType},
				},
			})
			if metadata.HeadCommit != sha {
				t.Fatalf("expected head commit %q, got %q", sha, metadata.HeadCommit)
			}
			if metadata.CommitURL != tt.wantURL {
				t.Fatalf("expected commit URL %q, got %q", tt.wantURL, metadata.CommitURL)
			}
		})
	}
}

func TestWorkingDirLockMetadataWithoutHeadCommit(t *testing.T) {
	metadata := WorkingDirLockMetadataForPull(models.PullRequest{
		BaseRepo: models.Repo{
			CloneURL: "https://github.com/owner/repo.git",
			VCSHost:  models.VCSHost{Type: models.Github},
		},
	})
	if metadata != (WorkingDirLockMetadata{}) {
		t.Fatalf("expected empty metadata, got %#v", metadata)
	}
}
