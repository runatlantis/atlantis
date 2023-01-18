package state

import (
	"fmt"
	"net/url"
	"time"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/terraform"
)

const (
	ConfirmAction = "Confirm"
	RejectAction  = "Reject"
)

type urlGenerator interface {
	Generate(jobID fmt.Stringer, BaseURL fmt.Stringer) (*url.URL, error)
}

type UpdateNotifier func(state *Workflow) error

type WorkflowStore struct {
	state              *Workflow
	notifier           UpdateNotifier
	outputURLGenerator urlGenerator
}

type UpdateOptions struct {
	PlanSummary  terraform.PlanSummary
	PlanApproval terraform.PlanApproval
	StartTime    time.Time
	EndTime      time.Time
}

func NewWorkflowStoreWithGenerator(notifier UpdateNotifier, g urlGenerator) *WorkflowStore {
	return &WorkflowStore{
		state:              &Workflow{},
		notifier:           notifier,
		outputURLGenerator: g,
	}
}

func NewWorkflowStore(notifier UpdateNotifier) *WorkflowStore {
	// Create a dummy route with the correct jobs path
	route := &mux.Route{}
	route.Path("/jobs/{job-id}")
	urlGenerator := &OutputURLGenerator{
		URLBuilder: route,
	}
	return NewWorkflowStoreWithGenerator(notifier, urlGenerator)
}

func (s *WorkflowStore) InitPlanJob(jobID fmt.Stringer, serverURL fmt.Stringer) error {
	outputURL, err := s.outputURLGenerator.Generate(jobID, serverURL)

	if err != nil {
		return errors.Wrap(err, "generating url for plan job")
	}
	s.state.Plan = &Job{
		ID: jobID.String(),
		Output: &JobOutput{
			URL: outputURL,
		},
		Status: WaitingJobStatus,
	}

	return s.notifier(s.state)
}

func convertApprovalToActions(approval terraform.PlanApproval) JobActions {
	if approval.Type == terraform.AutoApproval {
		return JobActions{}
	}

	return JobActions{
		Actions: []JobAction{
			{
				ID:   ConfirmAction,
				Info: "Confirm this plan to proceed to apply",
			},
			{
				ID:   RejectAction,
				Info: "Reject this plan to prevent the apply",
			},
		},
		Summary: approval.Reason,
	}
}

func (s *WorkflowStore) UpdateApprovalActions(approval terraform.PlanApproval) error {
	s.state.Apply.OnWaitingActions = convertApprovalToActions(approval)

	return s.notifier(s.state)
}

func (s *WorkflowStore) InitApplyJob(jobID fmt.Stringer, serverURL fmt.Stringer) error {
	outputURL, err := s.outputURLGenerator.Generate(jobID, serverURL)

	if err != nil {
		return errors.Wrap(err, "generating url for apply job")
	}
	s.state.Apply = &Job{
		ID: jobID.String(),
		Output: &JobOutput{
			URL: outputURL,
		},
		Status: WaitingJobStatus,
	}

	return s.notifier(s.state)
}

func (s *WorkflowStore) UpdatePlanJobWithStatus(status JobStatus, options ...UpdateOptions) error {
	s.state.Plan.Status = status

	for _, o := range options {
		s.state.Plan.Output.Summary = o.PlanSummary
	}
	return s.notifier(s.state)
}

func (s *WorkflowStore) UpdateApplyJobWithStatus(status JobStatus, options ...UpdateOptions) error {
	switch status {
	case InProgressJobStatus:
		s.state.Apply.StartTime = getStartTimeFromOpts(options...)

	case FailedJobStatus, SuccessJobStatus:
		s.state.Apply.EndTime = getEndTimeFromOpts(options...)
	}

	s.state.Apply.Status = status
	return s.notifier(s.state)
}

func (s *WorkflowStore) UpdateCompletion(result WorkflowResult) error {
	s.state.Result = result
	return s.notifier(s.state)
}

func getStartTimeFromOpts(options ...UpdateOptions) time.Time {
	for _, o := range options {
		if !o.StartTime.IsZero() {
			return o.StartTime
		}
	}
	return time.Time{}
}

func getEndTimeFromOpts(options ...UpdateOptions) time.Time {
	for _, o := range options {
		if !o.EndTime.IsZero() {
			return o.EndTime
		}
	}
	return time.Time{}
}
