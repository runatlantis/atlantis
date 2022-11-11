package decorators

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/lyft/aws/sns"
)

// AtlantisJobState represent current state of the job
// Job can be in 3 states:
//   - RUNNING - when the job is initiated
//   - FAILURE - when the job fails the execution
//   - SUCCESS - when the job runs successfully
type AtlantisJobState string

// AtlantisJobType represent the type of the job
// Currently only apply is supported
type AtlantisJobType string

const (
	AtlantisJobStateRunning AtlantisJobState = "RUNNING"
	AtlantisJobStateSuccess AtlantisJobState = "SUCCESS"
	AtlantisJobStateFailure AtlantisJobState = "FAILURE"

	AtlantisApplyJob AtlantisJobType = "APPLY"
)

// AuditProjectCommandWrapper is a decorator that notifies sns topic
// about the state of the command. It is used for auditing purposes
type AuditProjectCommandWrapper struct {
	SnsWriter sns.Writer
	events.ProjectCommandRunner
}

func (p *AuditProjectCommandWrapper) Apply(ctx command.ProjectContext) command.ProjectResult {
	id := ctx.JobID
	startTime := strconv.FormatInt(time.Now().Unix(), 10)

	atlantisJobEvent := &AtlantisJobEvent{
		Version:        1,
		ID:             id,
		RootName:       ctx.ProjectName,
		JobType:        AtlantisApplyJob,
		Repository:     ctx.BaseRepo.FullName,
		Environment:    ctx.Tags["environment"],
		PullNumber:     ctx.Pull.Num,
		InitiatingUser: ctx.User.Username,
		Project:        ctx.Tags["service_name"],
		ForceApply:     ctx.ForceApply,
		StartTime:      startTime,
		Revision:       ctx.Pull.HeadCommit,
	}

	if err := p.emit(ctx, AtlantisJobStateRunning, atlantisJobEvent); err != nil {
		// return an error if we are not able to write to sns
		return command.ProjectResult{
			Error: errors.Wrap(err, "emitting atlantis job event"),
		}
	}

	result := p.ProjectCommandRunner.Apply(ctx)

	if result.Error != nil || result.Failure != "" {
		if err := p.emit(ctx, AtlantisJobStateFailure, atlantisJobEvent); err != nil {
			ctx.Log.ErrorContext(ctx.RequestCtx, fmt.Sprintf("failed to emit atlantis job event %v", err))
		}

		return result
	}

	if err := p.emit(ctx, AtlantisJobStateSuccess, atlantisJobEvent); err != nil {
		ctx.Log.ErrorContext(ctx.RequestCtx, fmt.Sprintf("failed to emit atlantis job event %v", err))
	}

	return result
}

func (p *AuditProjectCommandWrapper) emit(
	ctx command.ProjectContext,
	state AtlantisJobState,
	atlantisJobEvent *AtlantisJobEvent,
) error {
	atlantisJobEvent.State = state

	if state == AtlantisJobStateFailure || state == AtlantisJobStateSuccess {
		atlantisJobEvent.EndTime = strconv.FormatInt(time.Now().Unix(), 10)
	}

	payload, err := atlantisJobEvent.Marshal()
	if err != nil {
		return errors.Wrap(err, "marshaling atlantis job event")
	}

	if err := p.SnsWriter.Write(payload); err != nil {
		return errors.Wrap(err, "writing to sns topic")
	}

	return nil
}

// AtlantisJobEvent contains metadata of the state of the AtlantisJobType command
type AtlantisJobEvent struct {
	Version        int              `json:"version"`
	ID             string           `json:"id"`
	State          AtlantisJobState `json:"state"`
	JobType        AtlantisJobType  `json:"job_type"`
	Revision       string           `json:"revision"`
	Repository     string           `json:"repository"`
	PullNumber     int              `json:"pull_number"`
	Environment    string           `json:"environment"`
	InitiatingUser string           `json:"initiating_user"`
	StartTime      string           `json:"start_time"`
	EndTime        string           `json:"end_time"`
	ForceApply     bool             `json:"force_apply"`

	// Service name in the manifest.yaml
	Project string `json:"project"`
	// ProjectName in the atlantis.yaml
	RootName string `json:"root_name"`

	// Currently we do not track approvers metadata.
	// ORCA-954 will implement this feature
	ApprovedBy   string `json:"approved_by"`
	ApprovedTime string `json:"approved_time"`
}

func (a *AtlantisJobEvent) Marshal() ([]byte, error) {
	eventPayload, err := json.Marshal(a)
	if err != nil {
		return nil, errors.Wrap(err, "marshaling atlantis job event")
	}

	return eventPayload, nil
}
