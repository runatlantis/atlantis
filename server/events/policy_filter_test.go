package events

// not using a separate test package to be able to test some private fields in struct ApprovedPolicyFilter

import (
	"context"
	"github.com/google/go-github/v45/github"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

const (
	ownerA      = "A"
	ownerB      = "B"
	ownerC      = "C"
	policyName  = "some-policy"
	policyOwner = "team"
)

func createCommit(time time.Time) *github.Commit {
	return &github.Commit{
		Committer: &github.CommitAuthor{
			Date: &time,
		},
	}
}

func TestFilter_Approved(t *testing.T) {
	time1 := time.UnixMicro(1)
	time2 := time.UnixMicro(2)

	reviewFetcher := &mockReviewFetcher{
		approvers: []string{ownerA, ownerB},
		reviews: []*github.PullRequestReview{
			{
				User: &github.User{Login: github.String(ownerA)},
			},
			{
				User:        &github.User{Login: github.String(ownerB)},
				SubmittedAt: &time2,
			},
		},
	}
	teamFetcher := &mockTeamMemberFetcher{
		members: []string{ownerA, ownerB, ownerC},
	}
	commitFetcher := &mockCommitFetcher{
		commit: createCommit(time1),
	}
	reviewDismisser := &mockReviewDismisser{}
	failedPolicies := []valid.PolicySet{
		{Name: policyName, Owner: policyOwner},
	}

	policyFilter := NewApprovedPolicyFilter(reviewFetcher, reviewDismisser, commitFetcher, teamFetcher, failedPolicies)
	filteredPolicies, err := policyFilter.Filter(context.Background(), 0, models.Repo{}, 0, failedPolicies)
	assert.NoError(t, err)
	assert.True(t, reviewFetcher.listUsernamesIsCalled)
	assert.True(t, reviewFetcher.listApprovalsIsCalled)
	assert.True(t, commitFetcher.isCalled)
	assert.True(t, teamFetcher.isCalled)
	assert.True(t, reviewDismisser.isCalled)
	assert.Empty(t, filteredPolicies)
}

func TestFilter_Approved_NoDismissal(t *testing.T) {
	time1 := time.UnixMicro(1)
	time2 := time.UnixMicro(2)
	reviewFetcher := &mockReviewFetcher{
		approvers: []string{ownerB},
		reviews: []*github.PullRequestReview{
			{
				User:        &github.User{Login: github.String(ownerB)},
				SubmittedAt: &time2,
			},
		},
	}
	commitFetcher := &mockCommitFetcher{
		commit: createCommit(time1),
	}
	reviewDismisser := &mockReviewDismisser{}
	teamFetcher := &mockTeamMemberFetcher{
		members: []string{ownerA, ownerB, ownerC},
	}
	failedPolicies := []valid.PolicySet{
		{Name: policyName, Owner: policyOwner},
	}

	policyFilter := NewApprovedPolicyFilter(reviewFetcher, reviewDismisser, commitFetcher, teamFetcher, failedPolicies)
	filteredPolicies, err := policyFilter.Filter(context.Background(), 0, models.Repo{}, 0, failedPolicies)
	assert.NoError(t, err)
	assert.True(t, reviewFetcher.listUsernamesIsCalled)
	assert.True(t, reviewFetcher.listApprovalsIsCalled)
	assert.True(t, commitFetcher.isCalled)
	assert.True(t, teamFetcher.isCalled)
	assert.False(t, reviewDismisser.isCalled)
	assert.Empty(t, filteredPolicies)
}

func TestFilter_NotApproved(t *testing.T) {
	time1 := time.UnixMicro(1)
	time2 := time.UnixMicro(2)

	reviewFetcher := &mockReviewFetcher{
		reviews: []*github.PullRequestReview{
			{
				User:        &github.User{Login: github.String(ownerA)},
				SubmittedAt: &time1,
			},
		},
	}
	teamFetcher := &mockTeamMemberFetcher{
		members: []string{ownerA, ownerB, ownerC},
	}
	commitFetcher := &mockCommitFetcher{
		commit: createCommit(time2),
	}
	reviewDismisser := &mockReviewDismisser{}
	failedPolicies := []valid.PolicySet{
		{Name: policyName, Owner: policyOwner},
	}

	policyFilter := NewApprovedPolicyFilter(reviewFetcher, reviewDismisser, commitFetcher, teamFetcher, failedPolicies)
	filteredPolicies, err := policyFilter.Filter(context.Background(), 0, models.Repo{}, 0, failedPolicies)
	assert.NoError(t, err)
	assert.True(t, reviewFetcher.listUsernamesIsCalled)
	assert.True(t, reviewFetcher.listApprovalsIsCalled)
	assert.True(t, commitFetcher.isCalled)
	assert.True(t, reviewDismisser.isCalled)
	assert.True(t, teamFetcher.isCalled)
	assert.Equal(t, failedPolicies, filteredPolicies)
}

func TestFilter_NotApproved_NoDismissal(t *testing.T) {
	reviewFetcher := &mockReviewFetcher{
		approvers: []string{ownerB},
	}
	teamFetcher := &mockTeamMemberFetcher{
		members: []string{ownerA, ownerC},
	}
	commitFetcher := &mockCommitFetcher{}
	reviewDismisser := &mockReviewDismisser{}
	failedPolicies := []valid.PolicySet{
		{Name: policyName, Owner: policyOwner},
	}

	policyFilter := NewApprovedPolicyFilter(reviewFetcher, reviewDismisser, commitFetcher, teamFetcher, failedPolicies)
	filteredPolicies, err := policyFilter.Filter(context.Background(), 0, models.Repo{}, 0, failedPolicies)
	assert.NoError(t, err)
	assert.True(t, reviewFetcher.listUsernamesIsCalled)
	assert.True(t, reviewFetcher.listApprovalsIsCalled)
	assert.True(t, commitFetcher.isCalled)
	assert.False(t, reviewDismisser.isCalled)
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
	commitFetcher := &mockCommitFetcher{}
	reviewDismisser := &mockReviewDismisser{}

	var failedPolicies []valid.PolicySet
	policyFilter := NewApprovedPolicyFilter(reviewFetcher, reviewDismisser, commitFetcher, teamFetcher, failedPolicies)
	filteredPolicies, err := policyFilter.Filter(context.Background(), 0, models.Repo{}, 0, failedPolicies)
	assert.NoError(t, err)
	assert.False(t, reviewFetcher.listUsernamesIsCalled)
	assert.False(t, reviewFetcher.listApprovalsIsCalled)
	assert.False(t, teamFetcher.isCalled)
	assert.False(t, commitFetcher.isCalled)
	assert.False(t, reviewDismisser.isCalled)
	assert.Empty(t, filteredPolicies)
}

func TestFilter_FailedListLatestApprovalUsernames(t *testing.T) {
	reviewFetcher := &mockReviewFetcher{
		listUsernamesError: assert.AnError,
	}
	teamFetcher := &mockTeamMemberFetcher{
		members: []string{ownerA, ownerB, ownerC},
	}
	commitFetcher := &mockCommitFetcher{}
	reviewDismisser := &mockReviewDismisser{}
	failedPolicies := []valid.PolicySet{
		{Name: policyName, Owner: policyOwner},
	}

	policyFilter := NewApprovedPolicyFilter(reviewFetcher, reviewDismisser, commitFetcher, teamFetcher, failedPolicies)
	filteredPolicies, err := policyFilter.Filter(context.Background(), 0, models.Repo{}, 0, failedPolicies)
	assert.Error(t, err)
	assert.True(t, reviewFetcher.listUsernamesIsCalled)
	assert.True(t, reviewFetcher.listApprovalsIsCalled)
	assert.True(t, commitFetcher.isCalled)
	assert.False(t, reviewDismisser.isCalled)
	assert.False(t, teamFetcher.isCalled)
	assert.Equal(t, failedPolicies, filteredPolicies)
}

func TestFilter_FailedListApprovalReviews(t *testing.T) {
	reviewFetcher := &mockReviewFetcher{
		listApprovalsError: assert.AnError,
	}
	teamFetcher := &mockTeamMemberFetcher{
		members: []string{ownerA, ownerB, ownerC},
	}
	commitFetcher := &mockCommitFetcher{}
	reviewDismisser := &mockReviewDismisser{}
	failedPolicies := []valid.PolicySet{
		{Name: policyName, Owner: policyOwner},
	}

	policyFilter := NewApprovedPolicyFilter(reviewFetcher, reviewDismisser, commitFetcher, teamFetcher, failedPolicies)
	filteredPolicies, err := policyFilter.Filter(context.Background(), 0, models.Repo{}, 0, failedPolicies)
	assert.Error(t, err)
	assert.False(t, reviewFetcher.listUsernamesIsCalled)
	assert.True(t, reviewFetcher.listApprovalsIsCalled)
	assert.False(t, commitFetcher.isCalled)
	assert.False(t, reviewDismisser.isCalled)
	assert.False(t, teamFetcher.isCalled)
	assert.Equal(t, failedPolicies, filteredPolicies)
}

func TestFilter_FailedFetchLatestCommitTime(t *testing.T) {
	reviewFetcher := &mockReviewFetcher{
		approvers: []string{ownerB},
	}
	teamFetcher := &mockTeamMemberFetcher{}
	commitFetcher := &mockCommitFetcher{
		error: assert.AnError,
	}
	reviewDismisser := &mockReviewDismisser{}

	failedPolicies := []valid.PolicySet{
		{Name: policyName, Owner: policyOwner},
	}
	policyFilter := NewApprovedPolicyFilter(reviewFetcher, reviewDismisser, commitFetcher, teamFetcher, failedPolicies)
	filteredPolicies, err := policyFilter.Filter(context.Background(), 0, models.Repo{}, 0, failedPolicies)
	assert.Error(t, err)
	assert.False(t, reviewFetcher.listUsernamesIsCalled)
	assert.True(t, reviewFetcher.listApprovalsIsCalled)
	assert.True(t, commitFetcher.isCalled)
	assert.False(t, teamFetcher.isCalled)
	assert.False(t, reviewDismisser.isCalled)
	assert.Equal(t, failedPolicies, filteredPolicies)
}

func TestFilter_FailedTeamMemberFetch(t *testing.T) {
	reviewFetcher := &mockReviewFetcher{
		approvers: []string{ownerB},
	}
	teamFetcher := &mockTeamMemberFetcher{
		error: assert.AnError,
	}
	commitFetcher := &mockCommitFetcher{}
	reviewDismisser := &mockReviewDismisser{}
	failedPolicies := []valid.PolicySet{
		{Name: policyName, Owner: policyOwner},
	}

	policyFilter := NewApprovedPolicyFilter(reviewFetcher, reviewDismisser, commitFetcher, teamFetcher, failedPolicies)
	filteredPolicies, err := policyFilter.Filter(context.Background(), 0, models.Repo{}, 0, failedPolicies)
	assert.Error(t, err)
	assert.True(t, reviewFetcher.listUsernamesIsCalled)
	assert.True(t, reviewFetcher.listApprovalsIsCalled)
	assert.True(t, commitFetcher.isCalled)
	assert.True(t, teamFetcher.isCalled)
	assert.False(t, reviewDismisser.isCalled)
	assert.Equal(t, failedPolicies, filteredPolicies)
}

func TestFilter_FailedDismiss(t *testing.T) {
	time1 := time.UnixMicro(1)
	time2 := time.UnixMicro(2)
	reviewFetcher := &mockReviewFetcher{
		approvers: []string{ownerB},
		reviews: []*github.PullRequestReview{
			{
				User:        &github.User{Login: github.String(ownerB)},
				SubmittedAt: &time1,
			},
		},
	}
	commitFetcher := &mockCommitFetcher{
		commit: createCommit(time2),
	}
	reviewDismisser := &mockReviewDismisser{
		error: assert.AnError,
	}
	teamFetcher := &mockTeamMemberFetcher{
		members: []string{ownerA, ownerB, ownerC},
	}
	failedPolicies := []valid.PolicySet{
		{Name: policyName, Owner: policyOwner},
	}

	policyFilter := NewApprovedPolicyFilter(reviewFetcher, reviewDismisser, commitFetcher, teamFetcher, failedPolicies)
	filteredPolicies, err := policyFilter.Filter(context.Background(), 0, models.Repo{}, 0, failedPolicies)
	assert.Error(t, err)
	assert.False(t, reviewFetcher.listUsernamesIsCalled)
	assert.True(t, reviewFetcher.listApprovalsIsCalled)
	assert.True(t, commitFetcher.isCalled)
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

type mockCommitFetcher struct {
	commit   *github.Commit
	error    error
	isCalled bool
}

func (c *mockCommitFetcher) FetchLatestPRCommit(_ context.Context, _ int64, _ models.Repo, _ int) (*github.Commit, error) {
	c.isCalled = true
	return c.commit, c.error
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
