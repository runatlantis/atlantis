package status

import (
	"time"

	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
)

// StatusState represents the current state of all status checks for a PR
type StatusState struct {
	PullRequest models.PullRequest
	Repository  models.Repo
	Statuses    map[command.Name]CommandStatus
	UpdatedAt   time.Time
}

// CommandStatus represents the status of a specific command
type CommandStatus struct {
	Command    command.Name
	Status     models.CommitStatus
	NumSuccess int
	NumTotal   int
	UpdatedAt  time.Time
}

// StatusOperation represents an operation to be performed on status
type StatusOperation int

const (
	// OperationSet sets a status normally
	OperationSet StatusOperation = iota
	// OperationClear explicitly clears/removes existing status
	OperationClear
	// OperationSilence skips setting status due to silence flags (no VCS interaction)
	OperationSilence
)

// StatusDecision represents a decision about what status operation to perform
type StatusDecision struct {
	Operation   StatusOperation
	Status      models.CommitStatus
	NumSuccess  int
	NumTotal    int
	Reason      string // Human-readable reason for the decision
	SilenceType string // Which silence flag caused this (for debugging)
}

// SilenceReason contains common silence reasons
type SilenceReason struct {
	NoProjects      string
	NoPlans         string
	ForkPRError     string
	ExplicitSilence string
}

var DefaultSilenceReasons = SilenceReason{
	NoProjects:      "silence enabled and no projects found",
	NoPlans:         "silence enabled and no plans generated", 
	ForkPRError:     "fork PR with silence enabled",
	ExplicitSilence: "explicit silence configuration",
}

// CommitStatusUpdater defines what the status package needs from a commit status updater
// Any type that implements these methods will automatically satisfy this interface
type CommitStatusUpdater interface {
	UpdateCombined(logger logging.SimpleLogging, repo models.Repo, pull models.PullRequest, status models.CommitStatus, cmdName command.Name) error
	UpdateCombinedCount(logger logging.SimpleLogging, repo models.Repo, pull models.PullRequest, status models.CommitStatus, cmdName command.Name, numSuccess int, numTotal int) error
}