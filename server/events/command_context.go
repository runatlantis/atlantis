package events

import (
	"github.com/hootsuite/atlantis/server/events/models"
	"github.com/hootsuite/atlantis/server/events/vcs"
	"github.com/hootsuite/atlantis/server/logging"
)

// CommandContext represents the context of a command that came from a comment
// on a pull request.
type CommandContext struct {
	// BaseRepo is the repository that the pull request will be merged into.
	BaseRepo models.Repo
	// HeadRepo is the repository that is getting merged into the BaseRepo.
	// If the pull request branch is from the same repository then HeadRepo will
	// be the same as BaseRepo.
	// See https://help.github.com/articles/about-pull-request-merges/.
	HeadRepo models.Repo
	Pull     models.PullRequest
	// User is the user that triggered this command.
	User    models.User
	Command *Command
	Log     *logging.SimpleLogger
	// VCSHost is the host that the command came from.
	VCSHost vcs.Host
}
