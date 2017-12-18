// Package models holds all models that are needed across packages.
// We place these models in their own package so as to avoid circular
// dependencies between packages (which is a compile error).
package models

import (
	paths "path"
	"time"
)

// Repo is a VCS repository.
type Repo struct {
	// FullName is the owner and repo name separated
	// by a "/", ex. "hootsuite/atlantis".
	FullName string
	// Owner is just the repo owner, ex. "hootsuite".
	Owner string
	// Name is just the repo name, ex. "atlantis".
	Name string
	// CloneURL is the full HTTPS url for cloning with username and token string
	// ex. "https://username:token@github.com/atlantis/atlantis.git".
	CloneURL string
	// SanitizedCloneURL is the full HTTPS url for cloning without the username and password.
	// ex. "https://github.com/atlantis/atlantis.git".
	SanitizedCloneURL string
}

// PullRequest is a VCS pull request.
// GitLab calls these Merge Requests.
type PullRequest struct {
	// Num is the pull request number or ID.
	Num int
	// HeadCommit points to the head of the branch that is being
	// pull requested into the base.
	HeadCommit string
	// URL is the url of the pull request.
	// ex. "https://github.com/hootsuite/atlantis/pull/1"
	URL string
	// Branch is the name of the head branch (not the base).
	Branch string
	// Author is the username of the pull request author.
	Author string
	// State will be one of Open or Closed.
	// Gitlab supports an additional "merged" state but Github doesn't so we map
	// merged to Closed.
	State PullRequestState
}

type PullRequestState int

const (
	Open PullRequestState = iota
	Closed
)

// User is a VCS user.
type User struct {
	Username string
}

// ProjectLock represents a lock on a project.
type ProjectLock struct {
	// Project is the project that is being locked.
	Project Project
	// Pull is the pull request from which the command was run that
	// created this lock.
	Pull PullRequest
	// User is the username of the user that ran the command
	// that created this lock.
	User User
	// Workspace is the Terraform workspace that this
	// lock is being held against.
	Workspace string
	// Time is the time at which the lock was first created.
	Time time.Time
}

// Project represents a Terraform project. Since there may be multiple
// Terraform projects in a single repo we also include Path to the project
// root relative to the repo root.
type Project struct {
	// RepoFullName is the owner and repo name, ex. "hootsuite/atlantis"
	RepoFullName string
	// Path to project root in the repo.
	// If "." then project is at root.
	// Never ends in "/".
	Path string
}

// Plan is the result of running an Atlantis plan command.
// This model is used to represent a plan on disk.
type Plan struct {
	// Project is the project this plan is for.
	Project Project
	// LocalPath is the absolute path to the plan on disk
	// (versus the relative path from the repo root).
	LocalPath string
}

// NewProject constructs a Project. Use this constructor because it
// sets Path correctly.
func NewProject(repoFullName string, path string) Project {
	path = paths.Clean(path)
	if path == "/" {
		path = "."
	}
	return Project{
		RepoFullName: repoFullName,
		Path:         path,
	}
}
