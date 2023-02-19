package templates

import (
	"io"
	"testing"
	"time"

	. "github.com/runatlantis/atlantis/testing"
)

func TestIndexTemplate(t *testing.T) {
	err := IndexTemplate.Execute(io.Discard, IndexData{
		Locks: []LockIndexData{
			{
				LockPath:      "lock path",
				RepoFullName:  "repo full name",
				PullNum:       1,
				Path:          "path",
				Workspace:     "workspace",
				Time:          time.Now(),
				TimeFormatted: "02-01-2006 15:04:05",
			},
		},
		ApplyLock: ApplyLockData{
			Locked:        true,
			Time:          time.Now(),
			TimeFormatted: "02-01-2006 15:04:05",
		},
		AtlantisVersion: "v0.0.0",
		CleanedBasePath: "/path",
	})
	Ok(t, err)
}

func TestLockTemplate(t *testing.T) {
	err := LockTemplate.Execute(io.Discard, LockDetailData{
		LockKeyEncoded:  "lock key encoded",
		LockKey:         "lock key",
		PullRequestLink: "https://example.com",
		LockedBy:        "locked by",
		Workspace:       "workspace",
		AtlantisVersion: "v0.0.0",
		CleanedBasePath: "/path",
		RepoOwner:       "repo owner",
		RepoName:        "repo name",
	})
	Ok(t, err)
}

func TestProjectJobsTemplate(t *testing.T) {
	err := ProjectJobsTemplate.Execute(io.Discard, ProjectJobData{
		AtlantisVersion: "v0.0.0",
		ProjectPath:     "project path",
		CleanedBasePath: "/path",
	})
	Ok(t, err)
}

func TestProjectJobsErrorTemplate(t *testing.T) {
	err := ProjectJobsTemplate.Execute(io.Discard, ProjectJobsError{
		AtlantisVersion: "v0.0.0",
		ProjectPath:     "project path",
		CleanedBasePath: "/path",
	})
	Ok(t, err)
}

func TestGithubAppSetupTemplate(t *testing.T) {
	err := GithubAppSetupTemplate.Execute(io.Discard, GithubSetupData{
		Target:          "target",
		Manifest:        "manifest",
		ID:              1,
		Key:             "key",
		WebhookSecret:   "webhook secret",
		URL:             "https://example.com",
		CleanedBasePath: "/path",
	})
	Ok(t, err)
}
