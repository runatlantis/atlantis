// Copyright 2017 HootSuite Media Inc.
//
// Licensed under the Apache License, Version 2.0 (the License);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an AS IS BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// Modified hereafter by contributors to runatlantis/atlantis.
//
// Package models holds all models that are needed across packages.
// We place these models in their own package so as to avoid circular
// dependencies between packages (which is a compile error).
package models

import (
	"fmt"
	"net/url"
	paths "path"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// Repo is a VCS repository.
type Repo struct {
	// FullName is the owner and repo name separated
	// by a "/", ex. "runatlantis/atlantis".
	FullName string
	// Owner is just the repo owner, ex. "runatlantis".
	Owner string
	// Name is just the repo name, ex. "atlantis".
	Name string
	// CloneURL is the full HTTPS url for cloning with username and token string
	// ex. "https://username:token@github.com/atlantis/atlantis.git".
	CloneURL string
	// SanitizedCloneURL is the full HTTPS url for cloning without the username and password.
	// ex. "https://github.com/atlantis/atlantis.git".
	SanitizedCloneURL string
	// VCSHost is where this repo is hosted.
	VCSHost VCSHost
}

func NewRepo(vcsHostType VCSHostType, repoFullName string, cloneURL string, vcsUser string, vcsToken string) (Repo, error) {
	if repoFullName == "" {
		return Repo{}, errors.New("repoFullName can't be empty")
	}
	if cloneURL == "" {
		return Repo{}, errors.New("cloneURL can't be empty")
	}

	// Ensure the Clone URL is for the same repo to avoid something malicious.
	cloneURLParsed, err := url.Parse(cloneURL)
	if err != nil {
		return Repo{}, errors.Wrap(err, "invalid clone url")
	}
	expClonePath := fmt.Sprintf("/%s.git", repoFullName)
	if expClonePath != cloneURLParsed.Path {
		return Repo{}, fmt.Errorf("expected clone url to have path %q but had %q", expClonePath, cloneURLParsed.Path)
	}

	// Construct clone urls with http auth. Need to do both https and http
	// because in GitLab's docs they have some http urls.
	auth := fmt.Sprintf("%s:%s@", vcsUser, vcsToken)
	authedCloneURL := strings.Replace(cloneURL, "https://", "https://"+auth, -1)
	authedCloneURL = strings.Replace(authedCloneURL, "http://", "http://"+auth, -1)

	// Get the owner and repo names from the full name.
	var owner string
	var repo string
	pathSplit := strings.Split(repoFullName, "/")
	if len(pathSplit) != 2 || pathSplit[0] == "" || pathSplit[1] == "" {
		return Repo{}, fmt.Errorf("invalid repo format %q", repoFullName)
	}
	owner = pathSplit[0]
	repo = pathSplit[1]

	return Repo{
		FullName:          repoFullName,
		Owner:             owner,
		Name:              repo,
		CloneURL:          authedCloneURL,
		SanitizedCloneURL: cloneURL,
		VCSHost: VCSHost{
			Type:     vcsHostType,
			Hostname: cloneURLParsed.Hostname(),
		},
	}, nil
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
	// ex. "https://github.com/runatlantis/atlantis/pull/1"
	URL string
	// Branch is the name of the head branch (not the base).
	Branch string
	// Author is the username of the pull request author.
	Author string
	// State will be one of Open or Closed.
	// Gitlab supports an additional "merged" state but Github doesn't so we map
	// merged to Closed.
	State PullRequestState
	// BaseRepo is the repository that the pull request will be merged into.
	BaseRepo Repo
	// HeadRepo is the repository that is getting merged into the BaseRepo.
	// If the pull request branch is from the same repository then HeadRepo will
	// be the same as BaseRepo.
	// See https://help.github.com/articles/about-pull-request-merges/.
	HeadRepo Repo
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
	// RepoFullName is the owner and repo name, ex. "runatlantis/atlantis"
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

// VCSHost is a Git hosting provider, for example GitHub.
type VCSHost struct {
	// Hostname is the hostname of the VCS provider, ex. "github.com" or
	// "github-enterprise.example.com".
	Hostname string

	// Type is which type of VCS host this is, ex. GitHub or GitLab.
	Type VCSHostType
}

type VCSHostType int

const (
	Github VCSHostType = iota
	Gitlab
)

func (h VCSHostType) String() string {
	switch h {
	case Github:
		return "Github"
	case Gitlab:
		return "Gitlab"
	}
	return "<missing String() implementation>"
}
