// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"strings"
	"time"

	"github.com/runatlantis/atlantis/server/core/db"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/jobs"
)

// OutputPersister saves command output to the database
type OutputPersister struct {
	db            db.Database
	outputHandler jobs.ProjectCommandOutputHandler
}

// NewOutputPersister creates a new OutputPersister
func NewOutputPersister(database db.Database, outputHandler jobs.ProjectCommandOutputHandler) *OutputPersister {
	return &OutputPersister{db: database, outputHandler: outputHandler}
}

// PersistResult saves the result of a project command to the database
func (p *OutputPersister) PersistResult(ctx command.ProjectContext, result command.ProjectResult) error {
	output := models.ProjectOutput{
		RepoFullName: ctx.BaseRepo.FullName,
		PullNum:      ctx.Pull.Num,
		ProjectName:  ctx.ProjectName,
		Workspace:    ctx.Workspace,
		Path:         ctx.RepoRelDir,
		CommandName:  ctx.CommandName.String(),
		JobID:        ctx.JobID,
		RunTimestamp: time.Now().UTC().UnixMilli(),
		TriggeredBy:  ctx.User.Username,
		StartedAt:    time.Now().UTC(),
		CompletedAt:  time.Now().UTC(),
		PullURL:      ctx.Pull.URL,
		PullTitle:    ctx.Pull.Title,
	}

	// Determine status and extract output
	if result.Error != nil {
		output.Status = models.FailedOutputStatus
		output.Error = result.Error.Error()
	} else if result.Failure != "" {
		output.Status = models.FailedOutputStatus
		output.Error = result.Failure
	} else {
		output.Status = models.SuccessOutputStatus
	}

	// Try to get streaming output buffer (preserves ANSI colors)
	if p.outputHandler != nil && ctx.JobID != "" {
		buffer := p.outputHandler.GetProjectOutputBuffer(ctx.JobID)
		if len(buffer.Buffer) > 0 {
			output.Output = strings.Join(buffer.Buffer, "\n")
		}
	}

	// Extract command-specific output and stats
	switch result.Command {
	case command.Plan:
		if result.PlanSuccess != nil {
			// Only use TerraformOutput if we didn't get streaming buffer
			if output.Output == "" {
				output.Output = result.PlanSuccess.TerraformOutput
			}
			stats := result.PlanSuccess.Stats()
			output.ResourceStats = models.ResourceStats{
				Import:  stats.Import,
				Add:     stats.Add,
				Change:  stats.Change,
				Destroy: stats.Destroy,
			}
		}
	case command.Apply:
		// Only use ApplySuccess if we didn't get streaming buffer
		if output.Output == "" {
			output.Output = result.ApplySuccess
		}
	case command.PolicyCheck:
		if result.PolicyCheckResults != nil {
			output.PolicyPassed = true
			for _, psr := range result.PolicyCheckResults.PolicySetResults {
				if !psr.Passed {
					output.PolicyPassed = false
				}
				output.PolicyOutput += psr.PolicyOutput + "\n"
			}
		}
	}

	return p.db.SaveProjectOutput(output)
}
