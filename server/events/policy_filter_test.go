package events

// not using a separate test package to be able to test some private fields in struct ApprovedPolicyFilter

import (
	"context"
	"github.com/runatlantis/atlantis/server/core/config/valid"
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
	teamFetcher := &mockTeamMemberFetcher{
		members: []string{ownerA, ownerB, ownerC},
	}
	policyFilter := NewApprovedPolicyFilter(reviewFetcher, teamFetcher)
	failedPolicies := []valid.PolicySet{
		{Name: policyName, Owner: policyOwner},
	}

	filteredPolicies, err := policyFilter.Filter(context.Background(), 0, models.Repo{}, 0, failedPolicies)
	assert.NoError(t, err)
	assert.True(t, reviewFetcher.isCalled)
	assert.Empty(t, filteredPolicies)
}

func TestFilter_ApprovedCacheMiss(t *testing.T) {
	team := []string{ownerA, ownerB, ownerC}
	reviewFetcher := &mockReviewFetcher{
		approvers: []string{ownerB},
	}
	teamFetcher := &mockTeamMemberFetcher{
		members: team,
	}
	policyFilter := NewApprovedPolicyFilter(reviewFetcher, teamFetcher)
	policyFilter.owners = map[string][]string{
		policyOwner: team,
	}
	failedPolicies := []valid.PolicySet{
		{Name: policyName, Owner: policyOwner},
	}

	filteredPolicies, err := policyFilter.Filter(context.Background(), 0, models.Repo{}, 0, failedPolicies)
	assert.NoError(t, err)
	assert.True(t, reviewFetcher.isCalled)
	assert.False(t, teamFetcher.isCalled)
	assert.Empty(t, filteredPolicies)
}

func TestFilter_NotApproved(t *testing.T) {
	reviewFetcher := &mockReviewFetcher{
		approvers: []string{ownerB},
	}
	teamFetcher := &mockTeamMemberFetcher{
		members: []string{ownerA, ownerC},
	}
	policyFilter := NewApprovedPolicyFilter(reviewFetcher, teamFetcher)
	failedPolicies := []valid.PolicySet{
		{Name: policyName, Owner: policyOwner},
	}

	filteredPolicies, err := policyFilter.Filter(context.Background(), 0, models.Repo{}, 0, failedPolicies)
	assert.NoError(t, err)
	assert.True(t, reviewFetcher.isCalled)
	assert.True(t, teamFetcher.isCalled)
	assert.Equal(t, failedPolicies, filteredPolicies)
}

func TestFilter_NoFailedPolicies(t *testing.T) {
	reviewFetcher := &mockReviewFetcher{
		approvers: []string{ownerB},
	}
	teamFetcher := &mockTeamMemberFetcher{
		members: []string{ownerA, ownerB, ownerC},
	}
	policyFilter := NewApprovedPolicyFilter(reviewFetcher, teamFetcher)
	var failedPolicies []valid.PolicySet

	filteredPolicies, err := policyFilter.Filter(context.Background(), 0, models.Repo{}, 0, failedPolicies)
	assert.NoError(t, err)
	assert.False(t, reviewFetcher.isCalled)
	assert.False(t, teamFetcher.isCalled)
	assert.Empty(t, filteredPolicies)
}

func TestFilter_FailedPRReviewFetch(t *testing.T) {
	reviewFetcher := &mockReviewFetcher{
		error: assert.AnError,
	}
	teamFetcher := &mockTeamMemberFetcher{
		members: []string{ownerA, ownerB, ownerC},
	}
	policyFilter := NewApprovedPolicyFilter(reviewFetcher, teamFetcher)
	failedPolicies := []valid.PolicySet{
		{Name: policyName, Owner: policyOwner},
	}

	filteredPolicies, err := policyFilter.Filter(context.Background(), 0, models.Repo{}, 0, failedPolicies)
	assert.Error(t, err)
	assert.True(t, reviewFetcher.isCalled)
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
	policyFilter := NewApprovedPolicyFilter(reviewFetcher, teamFetcher)
	failedPolicies := []valid.PolicySet{
		{Name: policyName, Owner: policyOwner},
	}

	filteredPolicies, err := policyFilter.Filter(context.Background(), 0, models.Repo{}, 0, failedPolicies)
	assert.Error(t, err)
	assert.True(t, reviewFetcher.isCalled)
	assert.True(t, teamFetcher.isCalled)
	assert.Equal(t, failedPolicies, filteredPolicies)
}

type mockReviewFetcher struct {
	approvers []string
	error     error
	isCalled  bool
}

func (f *mockReviewFetcher) ListApprovalReviewers(_ context.Context, _ int64, _ models.Repo, _ int) ([]string, error) {
	f.isCalled = true
	return f.approvers, f.error
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
