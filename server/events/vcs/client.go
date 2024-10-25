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

package vcs

import (
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
)

//go:generate pegomock generate --package mocks -o mocks/mock_client.go github.com/runatlantis/atlantis/server/events/vcs Client

// Client is used to make API calls to a VCS host like GitHub or GitLab.
type Client interface {
	// GetModifiedFiles returns the names of files that were modified in the merge request
	// relative to the repo root, e.g. parent/child/file.txt.
	GetModifiedFiles(logger logging.SimpleLogging, repo models.Repo, pull models.PullRequest) ([]string, error)
	CreateComment(logger logging.SimpleLogging, repo models.Repo, pullNum int, comment string, command string) error

	ReactToComment(logger logging.SimpleLogging, repo models.Repo, pullNum int, commentID int64, reaction string) error
	HidePrevCommandComments(logger logging.SimpleLogging, repo models.Repo, pullNum int, command string, dir string) error
	PullIsApproved(logger logging.SimpleLogging, repo models.Repo, pull models.PullRequest) (models.ApprovalStatus, error)
	PullIsMergeable(logger logging.SimpleLogging, repo models.Repo, pull models.PullRequest, vcsstatusname string) (bool, error)
	// UpdateStatus updates the commit status to state for pull. src is the
	// source of this status. This should be relatively static across runs,
	// ex. atlantis/plan or atlantis/apply.
	// description is a description of this particular status update and can
	// change across runs.
	// url is an optional link that users should click on for more information
	// about this status.
	UpdateStatus(logger logging.SimpleLogging, repo models.Repo, pull models.PullRequest, state models.CommitStatus, src string, description string, url string) error
	DiscardReviews(repo models.Repo, pull models.PullRequest) error
	MergePull(logger logging.SimpleLogging, pull models.PullRequest, pullOptions models.PullRequestOptions) error
	MarkdownPullLink(pull models.PullRequest) (string, error)
	GetTeamNamesForUser(repo models.Repo, user models.User) ([]string, error)

	// GetFileContent a repository file content from VCS (which support fetch a single file from repository)
	// The first return value indicates whether the repo contains a file or not
	// if BaseRepo had a file, its content will placed on the second return value
	GetFileContent(logger logging.SimpleLogging, repo models.Repo, branch string, fileName string) (bool, []byte, error)
	SupportsSingleFileDownload(repo models.Repo) bool
	GetCloneURL(logger logging.SimpleLogging, VCSHostType models.VCSHostType, repo string) (string, error)

	// GetPullLabels returns the labels of a pull request
	GetPullLabels(logger logging.SimpleLogging, repo models.Repo, pull models.PullRequest) ([]string, error)
}
