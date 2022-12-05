package events_test

import (
	"context"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	ownerA     = "A"
	ownerB     = "B"
	ownerC     = "C"
	policyName = "some-policy"
)

func TestFilter_Approved(t *testing.T) {
	reviewFetcher := &mockReviewFetcher{
		approvers: []string{ownerB},
	}
	globalCfg := valid.GlobalCfg{
		PolicySets: valid.PolicySets{
			Owners: valid.PolicyOwners{
				Users: []string{ownerA, ownerB, ownerC},
			},
		},
	}
	policyFilter := events.ApprovedPolicyFilter{
		GlobalCfg:        globalCfg,
		PRReviewsFetcher: reviewFetcher,
	}
	failedPolicies := []valid.PolicySet{
		{Name: policyName},
	}

	filteredPolicies, err := policyFilter.Filter(context.Background(), 0, models.Repo{}, 0, failedPolicies)
	assert.NoError(t, err)
	assert.True(t, reviewFetcher.isCalled)
	assert.Empty(t, filteredPolicies)
}

func TestFilter_NotApproved(t *testing.T) {
	reviewFetcher := &mockReviewFetcher{
		approvers: []string{ownerB},
	}
	globalCfg := valid.GlobalCfg{
		PolicySets: valid.PolicySets{
			Owners: valid.PolicyOwners{
				Users: []string{ownerA, ownerC},
			},
		},
	}
	policyFilter := events.ApprovedPolicyFilter{
		GlobalCfg:        globalCfg,
		PRReviewsFetcher: reviewFetcher,
	}
	failedPolicies := []valid.PolicySet{
		{Name: policyName},
	}

	filteredPolicies, err := policyFilter.Filter(context.Background(), 0, models.Repo{}, 0, failedPolicies)
	assert.NoError(t, err)
	assert.True(t, reviewFetcher.isCalled)
	assert.Equal(t, failedPolicies, filteredPolicies)
}

func TestFilter_NoFailedPolicies(t *testing.T) {
	reviewFetcher := &mockReviewFetcher{
		approvers: []string{ownerB},
	}
	globalCfg := valid.GlobalCfg{
		PolicySets: valid.PolicySets{
			Owners: valid.PolicyOwners{
				Users: []string{ownerA, ownerB, ownerC},
			},
		},
	}
	policyFilter := events.ApprovedPolicyFilter{
		GlobalCfg:        globalCfg,
		PRReviewsFetcher: reviewFetcher,
	}
	var failedPolicies []valid.PolicySet

	filteredPolicies, err := policyFilter.Filter(context.Background(), 0, models.Repo{}, 0, failedPolicies)
	assert.NoError(t, err)
	assert.False(t, reviewFetcher.isCalled)
	assert.Empty(t, filteredPolicies)
}

func TestFilter_FailedPRReviewFetch(t *testing.T) {
	reviewFetcher := &mockReviewFetcher{
		error: assert.AnError,
	}
	globalCfg := valid.GlobalCfg{
		PolicySets: valid.PolicySets{
			Owners: valid.PolicyOwners{
				Users: []string{ownerA, ownerB, ownerC},
			},
		},
	}
	policyFilter := events.ApprovedPolicyFilter{
		GlobalCfg:        globalCfg,
		PRReviewsFetcher: reviewFetcher,
	}
	failedPolicies := []valid.PolicySet{
		{Name: policyName},
	}

	filteredPolicies, err := policyFilter.Filter(context.Background(), 0, models.Repo{}, 0, failedPolicies)
	assert.Error(t, err)
	assert.True(t, reviewFetcher.isCalled)
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
