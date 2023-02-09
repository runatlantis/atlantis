package command

import (
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/uber-go/tally"
)

// Trigger represents the how the command was triggered
type Trigger int

const (
	// Commands that are automatically triggered (ie. automatic plans)
	AutoTrigger Trigger = iota

	// Commands that are triggered by comments (ie. atlantis plan)
	CommentTrigger
)

// Context represents the context of a command that should be executed
// for a pull request.
type Context struct {
	// HeadRepo is the repository that is getting merged into the BaseRepo.
	// If the pull request branch is from the same repository then HeadRepo will
	// be the same as BaseRepo.
	// See https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/incorporating-changes-from-a-pull-request/about-pull-request-merges
	HeadRepo models.Repo
	Pull     models.PullRequest
	Scope    tally.Scope
	// User is the user that triggered this command.
	User models.User
	Log  logging.SimpleLogging

	// Current PR state
	PullRequestStatus models.PullReqStatus

	PullStatus *models.PullStatus

	// PolicySet is the policy set to target (if specified) for the approve_policies command.
	PolicySet string

	Trigger Trigger
}
