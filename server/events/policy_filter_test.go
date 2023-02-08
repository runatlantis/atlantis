package events

// not using a separate test package to be able to test some private fields in struct ApprovedPolicyFilter

import (
	"context"
	"github.com/google/go-github/v45/github"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	ownerA      = "A"
	ownerB      = "B"
	ownerC      = "C"
	policyName  = "some-policy"
	policyOwner = "team"
)

func TestFilter_Approved(t *testing.T) {
	reviewFetcher := &mockReviewFetcher{
		approvers: []string{ownerB},
	}
	reviewDismisser := &mockReviewDismisser{}
	teamFetcher := &mockTeamMemberFetcher{
		members: []string{ownerA, ownerB, ownerC},
	}
	failedPolicies := []valid.PolicySet{
		{Name: policyName, Owner: policyOwner},
	}

	policyFilter := NewApprovedPolicyFilter(reviewFetcher, reviewDismisser, teamFetcher, failedPolicies)
	filteredPolicies, err := policyFilter.Filter(context.Background(), 0, models.Repo{}, 0, command.PRReviewTrigger, failedPolicies)
	assert.NoError(t, err)
	assert.True(t, reviewFetcher.listUsernamesIsCalled)
	assert.False(t, reviewFetcher.listApprovalsIsCalled)
	assert.True(t, teamFetcher.isCalled)
	assert.False(t, reviewDismisser.isCalled)
	assert.Empty(t, filteredPolicies)
}

func TestFilter_NotApproved(t *testing.T) {
	reviewFetcher := &mockReviewFetcher{
		reviews: []*github.PullRequestReview{
			{
				User: &github.User{Login: github.String(ownerA)},
			},
			{
				User: &github.User{Login: github.String(ownerB)},
			},
		},
	}
	teamFetcher := &mockTeamMemberFetcher{
		members: []string{ownerC},
	}
	reviewDismisser := &mockReviewDismisser{}
	failedPolicies := []valid.PolicySet{
		{Name: policyName, Owner: policyOwner},
	}

	policyFilter := NewApprovedPolicyFilter(reviewFetcher, reviewDismisser, teamFetcher, failedPolicies)
	filteredPolicies, err := policyFilter.Filter(context.Background(), 0, models.Repo{}, 0, command.AutoTrigger, failedPolicies)
	assert.NoError(t, err)
	assert.False(t, reviewFetcher.listUsernamesIsCalled)
	assert.True(t, reviewFetcher.listApprovalsIsCalled)
	assert.True(t, teamFetcher.isCalled)
	assert.False(t, reviewDismisser.isCalled)
	assert.Equal(t, failedPolicies, filteredPolicies)
}

func TestFilter_NotApproved_Dismissal(t *testing.T) {
	reviewFetcher := &mockReviewFetcher{
		reviews: []*github.PullRequestReview{
			{
				User: &github.User{Login: github.String(ownerA)},
			},
		},
	}
	teamFetcher := &mockTeamMemberFetcher{
		members: []string{ownerA},
	}
	reviewDismisser := &mockReviewDismisser{}
	failedPolicies := []valid.PolicySet{
		{Name: policyName, Owner: policyOwner},
	}

	policyFilter := NewApprovedPolicyFilter(reviewFetcher, reviewDismisser, teamFetcher, failedPolicies)
	filteredPolicies, err := policyFilter.Filter(context.Background(), 0, models.Repo{}, 0, command.AutoTrigger, failedPolicies)
	assert.NoError(t, err)
	assert.False(t, reviewFetcher.listUsernamesIsCalled)
	assert.True(t, reviewFetcher.listApprovalsIsCalled)
	assert.True(t, teamFetcher.isCalled)
	assert.True(t, reviewDismisser.isCalled)
	assert.Equal(t, failedPolicies, filteredPolicies)
}

func TestFilter_NoFailedPolicies(t *testing.T) {
	reviewFetcher := &mockReviewFetcher{
		approvers: []string{ownerB},
	}
	teamFetcher := &mockTeamMemberFetcher{
		members: []string{ownerA, ownerB, ownerC},
	}
	reviewDismisser := &mockReviewDismisser{}

	var failedPolicies []valid.PolicySet
	policyFilter := NewApprovedPolicyFilter(reviewFetcher, reviewDismisser, teamFetcher, failedPolicies)
	filteredPolicies, err := policyFilter.Filter(context.Background(), 0, models.Repo{}, 0, command.PRReviewTrigger, failedPolicies)
	assert.NoError(t, err)
	assert.False(t, reviewFetcher.listUsernamesIsCalled)
	assert.False(t, reviewFetcher.listApprovalsIsCalled)
	assert.False(t, teamFetcher.isCalled)
	assert.False(t, reviewDismisser.isCalled)
	assert.Empty(t, filteredPolicies)
}

func TestFilter_FailedListLatestApprovalUsernames(t *testing.T) {
	reviewFetcher := &mockReviewFetcher{
		listUsernamesError: assert.AnError,
	}
	teamFetcher := &mockTeamMemberFetcher{}
	reviewDismisser := &mockReviewDismisser{}
	failedPolicies := []valid.PolicySet{
		{Name: policyName, Owner: policyOwner},
	}

	policyFilter := NewApprovedPolicyFilter(reviewFetcher, reviewDismisser, teamFetcher, failedPolicies)
	filteredPolicies, err := policyFilter.Filter(context.Background(), 0, models.Repo{}, 0, command.PRReviewTrigger, failedPolicies)
	assert.Error(t, err)
	assert.True(t, reviewFetcher.listUsernamesIsCalled)
	assert.False(t, reviewFetcher.listApprovalsIsCalled)
	assert.False(t, reviewDismisser.isCalled)
	assert.False(t, teamFetcher.isCalled)
	assert.Equal(t, failedPolicies, filteredPolicies)
}

func TestFilter_FailedListApprovalReviews(t *testing.T) {
	reviewFetcher := &mockReviewFetcher{
		listApprovalsError: assert.AnError,
	}
	teamFetcher := &mockTeamMemberFetcher{}
	reviewDismisser := &mockReviewDismisser{}
	failedPolicies := []valid.PolicySet{
		{Name: policyName, Owner: policyOwner},
	}

	policyFilter := NewApprovedPolicyFilter(reviewFetcher, reviewDismisser, teamFetcher, failedPolicies)
	filteredPolicies, err := policyFilter.Filter(context.Background(), 0, models.Repo{}, 0, command.CommentTrigger, failedPolicies)
	assert.Error(t, err)
	assert.False(t, reviewFetcher.listUsernamesIsCalled)
	assert.True(t, reviewFetcher.listApprovalsIsCalled)
	assert.False(t, reviewDismisser.isCalled)
	assert.False(t, teamFetcher.isCalled)
	assert.Equal(t, failedPolicies, filteredPolicies)
}

func TestFilter_FailedTeamMemberFetch(t *testing.T) {
	reviewFetcher := &mockReviewFetcher{
		approvers: []string{ownerB},
	}
	teamFetcher := &mockTeamMemberFetcher{
		error: assert.AnError,
	}
	reviewDismisser := &mockReviewDismisser{}
	failedPolicies := []valid.PolicySet{
		{Name: policyName, Owner: policyOwner},
	}

	policyFilter := NewApprovedPolicyFilter(reviewFetcher, reviewDismisser, teamFetcher, failedPolicies)
	filteredPolicies, err := policyFilter.Filter(context.Background(), 0, models.Repo{}, 0, command.PRReviewTrigger, failedPolicies)
	assert.Error(t, err)
	assert.True(t, reviewFetcher.listUsernamesIsCalled)
	assert.False(t, reviewFetcher.listApprovalsIsCalled)
	assert.True(t, teamFetcher.isCalled)
	assert.False(t, reviewDismisser.isCalled)
	assert.Equal(t, failedPolicies, filteredPolicies)
}

func TestFilter_FailedDismiss(t *testing.T) {
	reviewFetcher := &mockReviewFetcher{
		reviews: []*github.PullRequestReview{
			{
				User: &github.User{Login: github.String(ownerB)},
			},
		},
	}
	reviewDismisser := &mockReviewDismisser{
		error: assert.AnError,
	}
	teamFetcher := &mockTeamMemberFetcher{
		members: []string{ownerB},
	}
	failedPolicies := []valid.PolicySet{
		{Name: policyName, Owner: policyOwner},
	}

	policyFilter := NewApprovedPolicyFilter(reviewFetcher, reviewDismisser, teamFetcher, failedPolicies)
	filteredPolicies, err := policyFilter.Filter(context.Background(), 0, models.Repo{}, 0, command.AutoTrigger, failedPolicies)
	assert.Error(t, err)
	assert.False(t, reviewFetcher.listUsernamesIsCalled)
	assert.True(t, reviewFetcher.listApprovalsIsCalled)
	assert.True(t, teamFetcher.isCalled)
	assert.True(t, reviewDismisser.isCalled)
	assert.Equal(t, failedPolicies, filteredPolicies)
}

type mockReviewFetcher struct {
	approvers             []string
	listUsernamesIsCalled bool
	listUsernamesError    error
	reviews               []*github.PullRequestReview
	listApprovalsIsCalled bool
	listApprovalsError    error
}

func (f *mockReviewFetcher) ListLatestApprovalUsernames(_ context.Context, _ int64, _ models.Repo, _ int) ([]string, error) {
	f.listUsernamesIsCalled = true
	return f.approvers, f.listUsernamesError
}

func (f *mockReviewFetcher) ListApprovalReviews(_ context.Context, _ int64, _ models.Repo, _ int) ([]*github.PullRequestReview, error) {
	f.listApprovalsIsCalled = true
	return f.reviews, f.listApprovalsError
}

type mockReviewDismisser struct {
	error    error
	isCalled bool
}

func (d *mockReviewDismisser) Dismiss(_ context.Context, _ int64, _ models.Repo, _ int, _ int64) error {
	d.isCalled = true
	return d.error
}

type mockTeamMemberFetcher struct {
	members  []string
	error    error
	isCalled bool
}

func (m *mockTeamMemberFetcher) ListTeamMembers(_ context.Context, _ int64, _ string) ([]string, error) {
	m.isCalled = true
	return m.members, m.error
}
