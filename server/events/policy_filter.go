package events

import (
	"context"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/events/models"
	"sync"
)

type prReviewsFetcher interface {
	ListApprovalReviewers(ctx context.Context, installationToken int64, repo models.Repo, prNum int) ([]string, error)
}

type teamMemberFetcher interface {
	ListTeamMembers(ctx context.Context, installationToken int64, teamSlug string) ([]string, error)
}

type ApprovedPolicyFilter struct {
	owners sync.Map //cache

	prReviewsFetcher  prReviewsFetcher
	teamMemberFetcher teamMemberFetcher
}

func NewApprovedPolicyFilter(prReviewsFetcher prReviewsFetcher, teamMemberFetcher teamMemberFetcher) *ApprovedPolicyFilter {
	return &ApprovedPolicyFilter{
		prReviewsFetcher:  prReviewsFetcher,
		teamMemberFetcher: teamMemberFetcher,
		owners:            sync.Map{},
	}
}

// Filter will remove failed policies if the underlying PR has been approved by a policy owner
func (p *ApprovedPolicyFilter) Filter(ctx context.Context, installationToken int64, repo models.Repo, prNum int, failedPolicies []valid.PolicySet) ([]valid.PolicySet, error) {
	// Skip GH API calls if no policies failed
	if len(failedPolicies) == 0 {
		return failedPolicies, nil
	}

	// Fetch reviewers who approved the PR
	approvedReviewers, err := p.prReviewsFetcher.ListApprovalReviewers(ctx, installationToken, repo, prNum)
	if err != nil {
		return failedPolicies, errors.Wrap(err, "failed to fetch GH PR reviews")
	}

	// Filter out policies that already have been approved within GH
	var filteredFailedPolicies []valid.PolicySet
	for _, failedPolicy := range failedPolicies {
		approved, err := p.policyApproved(ctx, installationToken, approvedReviewers, failedPolicy)
		if err != nil {
			return failedPolicies, errors.Wrap(err, "validating policy approval")
		}
		if !approved {
			filteredFailedPolicies = append(filteredFailedPolicies, failedPolicy)
		}
	}
	return filteredFailedPolicies, nil
}

func (p *ApprovedPolicyFilter) policyApproved(ctx context.Context, installationToken int64, approvedReviewers []string, failedPolicy valid.PolicySet) (bool, error) {
	// Check if any reviewer is an owner of the failed policy set
	if p.ownersContainsPolicy(approvedReviewers, failedPolicy) {
		return true, nil
	}

	// Cache miss, fetch and try again
	err := p.getOwners(ctx, installationToken, failedPolicy)
	if err != nil {
		return false, errors.Wrap(err, "updating owners cache")
	}
	return p.ownersContainsPolicy(approvedReviewers, failedPolicy), nil
}

func (p *ApprovedPolicyFilter) ownersContainsPolicy(approvedReviewers []string, failedPolicy valid.PolicySet) bool {
	// Check if any reviewer is an owner of the failed policy set
	owners, ok := p.owners.Load(failedPolicy.Owner)
	if !ok {
		return false
	}
	for _, owner := range owners.([]string) {
		for _, reviewer := range approvedReviewers {
			if reviewer == owner {
				return true
			}
		}
	}
	return false
}

func (p *ApprovedPolicyFilter) getOwners(ctx context.Context, installationToken int64, policy valid.PolicySet) error {
	members, err := p.teamMemberFetcher.ListTeamMembers(ctx, installationToken, policy.Owner)
	if err != nil {
		return errors.Wrap(err, "failed to fetch GH team members")
	}
	p.owners.Store(policy.Owner, members)
	return nil
}
