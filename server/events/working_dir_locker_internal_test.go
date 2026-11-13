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
		pullURL  string
		wantURL  string
	}{
		{
			name:     "GitHub uses the fork-safe pull commit route",
			hostType: models.Github,
			pullURL:  "https://github.com/owner/repo/pull/123",
			wantURL:  "https://github.com/owner/repo/pull/123/commits/" + sha,
		},
		{
			name:     "GitLab uses the merge request diff route",
			hostType: models.Gitlab,
			pullURL:  "https://gitlab.example.com/gitlab/group/repo/-/merge_requests/123",
			wantURL:  "https://gitlab.example.com/gitlab/group/repo/-/merge_requests/123/diffs?commit_id=" + sha,
		},
		{
			name:     "Gitea uses the pull commit route",
			hostType: models.Gitea,
			pullURL:  "https://gitea.example.com/owner/repo/pulls/123/",
			wantURL:  "https://gitea.example.com/owner/repo/pulls/123/commits/" + sha,
		},
		{
			name:     "unsupported provider falls back to SHA",
			hostType: models.BitbucketCloud,
			pullURL:  "https://bitbucket.org/owner/repo/pull-requests/123",
		},
		{
			name:     "missing pull URL falls back to SHA",
			hostType: models.Github,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata := WorkingDirLockMetadataForPull(models.PullRequest{
				HeadCommit: sha,
				URL:        tt.pullURL,
				BaseRepo:   models.Repo{VCSHost: models.VCSHost{Type: tt.hostType}},
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
		URL:      "https://github.com/owner/repo/pull/123",
		BaseRepo: models.Repo{VCSHost: models.VCSHost{Type: models.Github}},
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

func TestTryLockPullKeepsBlockingCommitMetadataTogether(t *testing.T) {
	const repo = "owner/repo"
	locker := NewDefaultWorkingDirLocker()
	_, err := locker.TryLockPull(repo, 1, command.Plan, WorkingDirLockMetadata{HeadCommit: "old", CommitURL: "old-commit"})
	if err != nil {
		t.Fatal(err)
	}
	for i, metadata := range []WorkingDirLockMetadata{
		{HeadCommit: "old", JobURL: "old-job"},
		{HeadCommit: "new", JobURL: "new-job"},
	} {
		_, err = locker.TryLock(repo, 1, "workspace", ".", []string{"old", "new"}[i], command.Plan, metadata)
		if err != nil {
			t.Fatal(err)
		}
	}

	_, err = locker.TryLockPull(repo, 1, command.Apply, WorkingDirLockMetadata{HeadCommit: "new"})
	lockErr := err.(*workingDirLockError)
	if lockErr.metadata.HeadCommit != "old" || lockErr.metadata.CommitURL != "old-commit" || lockErr.metadata.JobURL != "old-job" || lockErr.multipleJobs {
		t.Fatalf("mixed blocking commits: %#v", lockErr)
	}
}

func TestTryLockPullFindsPlanForCurrentCommit(t *testing.T) {
	const repo = "owner/repo"
	locker := NewDefaultWorkingDirLocker()
	for i, metadata := range []WorkingDirLockMetadata{
		{HeadCommit: "old", JobURL: "old-job"},
		{HeadCommit: "new", JobURL: "new-job"},
	} {
		_, err := locker.TryLock(repo, 1, "workspace", ".", []string{"old", "new"}[i], command.Plan, metadata)
		if err != nil {
			t.Fatal(err)
		}
	}

	_, err := locker.TryLockPull(repo, 1, command.Apply, WorkingDirLockMetadata{HeadCommit: "new"})
	lockErr := err.(*workingDirLockError)
	if lockErr.metadata.HeadCommit != "new" || lockErr.metadata.JobURL != "new-job" || lockErr.multipleJobs {
		t.Fatalf("selected wrong blocking commit: %#v", lockErr)
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
				_, err = locker.TryLock(repo, 1, "workspace", ".", []string{"a", "b"}[i], command.Plan, WorkingDirLockMetadata{HeadCommit: "owner-sha", JobURL: url})
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
			if lockErr.Error() != "cannot run \"apply\": pull request 1 is currently locked by \"plan\" for commit owner-s.\nWait until the previous command is complete and try again" {
				t.Fatalf("unexpected lock error message: %s", lockErr)
			}
		})
	}
}
