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
package events

import (
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
)

// CommandTrigger represents the how the command was triggered
type CommandTrigger int

const (
	// Commands that are automatically triggered (ie. automatic plans)
	Auto CommandTrigger = iota

	// Commands that are triggered by comments (ie. atlantis plan)
	Comment
)

// CommandContext represents the context of a command that should be executed
// for a pull request.
type CommandContext struct {
	// HeadRepo is the repository that is getting merged into the BaseRepo.
	// If the pull request branch is from the same repository then HeadRepo will
	// be the same as BaseRepo.
	// See https://help.github.com/articles/about-pull-request-merges/.
	HeadRepo models.Repo
	Pull     models.PullRequest
	// User is the user that triggered this command.
	User models.User
	Log  logging.SimpleLogging

	// Current PR state
	PullRequestStatus models.PullReqStatus

	PullStatus *models.PullStatus

	Trigger CommandTrigger
}
