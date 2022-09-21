package state

import (
	"fmt"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/activities/terraform"
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
	PlanSummary terraform.PlanSummary
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
		Output: &JobOutput{
			URL: outputURL,
		},
		Status: WaitingJobStatus,
	}

	return s.notifier(s.state)
}

func (s *WorkflowStore) InitApplyJob(jobID fmt.Stringer, serverURL fmt.Stringer) error {
	outputURL, err := s.outputURLGenerator.Generate(jobID, serverURL)

	if err != nil {
		return errors.Wrap(err, "generating url for apply job")
	}
	s.state.Apply = &Job{
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
	s.state.Apply.Status = status
	return s.notifier(s.state)
}
