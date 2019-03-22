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
)

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_client.go Client

// Client is used to make API calls to a VCS host like GitHub or GitLab.
type Client interface {
	GetModifiedFiles(repo models.Repo, pull models.PullRequest) ([]string, error)
	CreateComment(repo models.Repo, pullNum int, comment string) error
	PullIsApproved(repo models.Repo, pull models.PullRequest) (bool, error)
	PullIsMergeable(repo models.Repo, pull models.PullRequest) (bool, error)
	// UpdateStatus updates the commit status to state for pull. src is the
	// source of this status. This should be relatively static across runs,
	// ex. atlantis/plan or atlantis/apply.
	// description is a description of this particular status update and can
	// change across runs.
	// url is an optional link that users should click on for more information
	// about this status.
	UpdateStatus(repo models.Repo, pull models.PullRequest, state models.CommitStatus, src string, description string, url string) error
	MergePull(pull models.PullRequest) error
}
