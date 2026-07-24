// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.

package events

import (
	"testing"

	"github.com/runatlantis/atlantis/server/events/command"
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

func TestWorkingDirLockMetadataForProject(t *testing.T) {
	ctx := command.ProjectContext{Pull: models.PullRequest{HeadCommit: "sha"}}
	metadata := WorkingDirLockMetadataForProject(ctx, "https://atlantis.example.com/jobs/job-id")
	if metadata.HeadCommit != "sha" || metadata.JobURL != "https://atlantis.example.com/jobs/job-id" {
		t.Fatalf("unexpected metadata: %#v", metadata)
	}
}

func TestWorkingDirLockErrorUsesOwnerMetadata(t *testing.T) {
	locker := NewDefaultWorkingDirLocker()
	owner := WorkingDirLockMetadata{HeadCommit: "owner-sha", CommitURL: "owner-commit", JobURL: "owner-job"}
	_, err := locker.TryLock("owner/repo", 1, "default", ".", "project", command.Plan, owner)
	if err != nil {
		t.Fatal(err)
	}
	_, err = locker.TryLock("owner/repo", 1, "default", ".", "project", command.Apply, WorkingDirLockMetadata{HeadCommit: "loser-sha", CommitURL: "loser-commit", JobURL: "loser-job"})
	lockErr, ok := err.(*workingDirLockError)
	if !ok {
		t.Fatalf("expected typed lock error, got %T", err)
	}
	if lockErr.metadata != owner {
		t.Fatalf("expected owner metadata %#v, got %#v", owner, lockErr.metadata)
	}
}

func TestTryLockPullUsesBlockingPlanProjectMetadata(t *testing.T) {
	locker := NewDefaultWorkingDirLocker()
	owner := WorkingDirLockMetadata{HeadCommit: "owner-sha", CommitURL: "owner-commit", JobURL: "owner-job"}
	_, err := locker.TryLock("owner/repo", 1, "default", ".", "project", command.Plan, owner)
	if err != nil {
		t.Fatal(err)
	}
	_, err = locker.TryLockPull("owner/repo", 1, command.Apply, WorkingDirLockMetadata{})
	lockErr, ok := err.(*workingDirLockError)
	if !ok {
		t.Fatalf("expected typed lock error, got %T", err)
	}
	if lockErr.metadata != owner || lockErr.multipleJobs {
		t.Fatalf("unexpected lock error: %#v", lockErr)
	}
}

func TestTryLockPullSelectsActiveProjectJobs(t *testing.T) {
	const repo = "owner/repo"
	const jobURL = "https://atlantis.example.com/jobs/job-id"
	tests := []struct {
		name         string
		jobURLs      []string
		wantJobURL   string
		multipleJobs bool
	}{
		{name: "zero"},
		{name: "one", jobURLs: []string{jobURL}, wantJobURL: jobURL},
		{name: "duplicate URL is one job", jobURLs: []string{jobURL, jobURL}, wantJobURL: jobURL},
		{name: "multiple", jobURLs: []string{jobURL + "-1", jobURL + "-2"}, multipleJobs: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			locker := NewDefaultWorkingDirLocker()
			_, err := locker.TryLockPull(repo, 1, command.Plan, WorkingDirLockMetadata{HeadCommit: "owner-sha", JobURL: "pull-job-is-ignored"})
			if err != nil {
				t.Fatal(err)
			}
			for i, url := range tt.jobURLs {
				_, err = locker.TryLock(repo, 1, "workspace", ".", []string{"a", "b"}[i], command.Plan, WorkingDirLockMetadata{JobURL: url})
				if err != nil {
					t.Fatal(err)
				}
			}
			_, err = locker.TryLockPull(repo, 1, command.Apply, WorkingDirLockMetadata{JobURL: "loser-job"})
			lockErr, ok := err.(*workingDirLockError)
			if !ok {
				t.Fatalf("expected typed lock error, got %T", err)
			}
			if lockErr.metadata.HeadCommit != "owner-sha" || lockErr.metadata.JobURL != tt.wantJobURL || lockErr.multipleJobs != tt.multipleJobs {
				t.Fatalf("unexpected lock error: %#v", lockErr)
			}
			if lockErr.Error() != "cannot run \"apply\": pull request 1 is currently locked by \"plan\" for commit owner-sha.\nWait until the previous command is complete and try again" {
				t.Fatalf("unexpected lock error message: %s", lockErr)
			}
		})
	}
}
