package github

import (
	"fmt"

	"github.com/google/go-github/v45/github"
)

const (
	UnlockLabel       = "Unlock"
	UnlockDescription = "Unlock this plan to proceed"
)

type CheckRunState string

type CheckRunAction struct {
	Description string
	Label       string
}

func (a CheckRunAction) ToGithubAction() *github.CheckRunAction {
	return &github.CheckRunAction{
		Label:       a.Label,
		Description: a.Description,

		// we encode the label as the id since there's a 20 char limit anyways
		// and we can use the check run external id to map to the correct workflow
		Identifier: a.Label,
	}
}

func CreateUnlockAction() CheckRunAction {
	return CheckRunAction{
		Description: UnlockDescription,
		Label:       UnlockLabel,
	}
}

func CreatePlanReviewAction(t PlanReviewActionType) CheckRunAction {
	return CheckRunAction{
		Description: fmt.Sprintf("%s this plan to proceed", string(t)),
		Label:       string(t),
	}

}

type PlanReviewActionType string

const (
	CheckRunSuccess        CheckRunState = "success"
	CheckRunFailure        CheckRunState = "failure"
	CheckRunTimeout        CheckRunState = "timed_out"
	CheckRunPending        CheckRunState = "in_progress"
	CheckRunQueued         CheckRunState = "queued"
	CheckRunActionRequired CheckRunState = "action_required"
	CheckRunUnknown        CheckRunState = ""

	Approve PlanReviewActionType = "Approve"
	Reject  PlanReviewActionType = "Reject"
)
