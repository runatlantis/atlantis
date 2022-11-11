package activities

import (
	"context"
	"encoding/json"
	"io"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/github"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/terraform"
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

const (
	EnvironmentTagKey = "environment"
	ProjectTagKey     = "service_name"
)

type AuditJobRequest struct {
	Repo           github.Repo
	Root           terraform.Root
	Revision       string
	InitiatingUser github.User
	JobID          string
	State          AtlantisJobState
	StartTime      string
	EndTime        string
	IsForceApply   bool
	Tags           map[string]string
}

type auditActivities struct {
	SnsWriter io.Writer
}

func (a *auditActivities) AuditJob(ctx context.Context, req AuditJobRequest) error {
	atlantisJobEvent := &AtlantisJobEvent{
		Version:        1,
		ID:             req.JobID,
		State:          req.State,
		RootName:       req.Root.Name,
		JobType:        AtlantisApplyJob,
		Repository:     req.Repo.GetFullName(),
		InitiatingUser: req.InitiatingUser.Username,
		ForceApply:     req.IsForceApply,
		StartTime:      req.StartTime,
		Revision:       req.Revision,
		Project:        req.Tags[ProjectTagKey],
		Environment:    req.Tags[EnvironmentTagKey],
	}

	if req.State == AtlantisJobStateFailure || req.State == AtlantisJobStateSuccess {
		atlantisJobEvent.EndTime = req.EndTime
	}

	payload, err := atlantisJobEvent.Marshal()
	if err != nil {
		return errors.Wrap(err, "marshaling atlantis job event")
	}

	if _, err := a.SnsWriter.Write(payload); err != nil {
		return errors.Wrap(err, "writing to sns topic")
	}

	return nil
}
