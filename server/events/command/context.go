package command

import (
	"context"
	"time"

	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/uber-go/tally/v4"
)

// CommandTrigger represents the how the command was triggered
type CommandTrigger int //nolint:revive // avoiding refactor while adding linter action

const (
	// Commands that are automatically triggered (ie. automatic plans)
	AutoTrigger CommandTrigger = iota

	// Commands that are triggered by comments (ie. atlantis plan)
	CommentTrigger
)

// Context represents the context of a command that should be executed
// for a pull request.
type Context struct {
	// HeadRepo is the repository that is getting merged into the BaseRepo.
	// If the pull request branch is from the same repository then HeadRepo will
	// be the same as BaseRepo.
	// See https://help.github.com/articles/about-pull-request-merges/.
	HeadRepo models.Repo
	Pull     models.PullRequest
	Scope    tally.Scope
	// User is the user that triggered this command.
	User models.User
	Log  logging.Logger

	// Current PR state
	PullRequestStatus models.PullReqStatus

	PullStatus *models.PullStatus

	Trigger CommandTrigger

	// Time Atlantis received VCS event, triggering command to be executed
	TriggerTimestamp time.Time
	RequestCtx       context.Context

	InstallationToken int64
}
