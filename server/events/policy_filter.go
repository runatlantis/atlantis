package events

import (
	"context"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/events/models"
)

type prReviewsFetcher interface {
	ListApprovalReviewers(ctx context.Context, installationToken int64, repo models.Repo, prNum int) ([]string, error)
}

type ApprovedPolicyFilter struct {
	//TODO: remove in favor of an in-memory cache of owners
	GlobalCfg        valid.GlobalCfg
	PRReviewsFetcher prReviewsFetcher
}

// Filter performs an all-or-nothing filter against failed policies for now. If PR reviewer is a global policy owner,
// all policies will pass. TODO: This will change to make filter decisions for each failed policy individually.
func (p *ApprovedPolicyFilter) Filter(ctx context.Context, installationToken int64, repo models.Repo, prNum int, failedPolicies []valid.PolicySet) ([]valid.PolicySet, error) {
	// Skip GH API calls if no policies failed
	if len(failedPolicies) == 0 {
		return failedPolicies, nil
	}

	// Fetch reviewers who approved the PR
	approvedReviewers, err := p.PRReviewsFetcher.ListApprovalReviewers(ctx, installationToken, repo, prNum)
	if err != nil {
		return failedPolicies, errors.Wrap(err, "failed to fetch GH PR reviews")
	}

	// TODO: check each policy set separately when per-policy owners configuration is implemented
	for _, reviewer := range approvedReviewers {
		if p.GlobalCfg.PolicySets.IsOwner(reviewer) {
			return nil, nil
		}
	}
	return failedPolicies, nil
}
