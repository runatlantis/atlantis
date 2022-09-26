package checks_test

import (
	"context"
	"testing"

	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/events/vcs/types"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/lyft/checks"
	"github.com/runatlantis/atlantis/server/lyft/feature"
	"github.com/stretchr/testify/assert"
	"github.com/uber-go/tally/v4"
)

type testFeatureAllocator struct {
	isChecksEnabled bool
}

func (t testFeatureAllocator) ShouldAllocate(featureID feature.Name, featureCtx feature.FeatureContext) (bool, error) {
	return t.isChecksEnabled, nil
}

func TestChecksClientWrapper(t *testing.T) {
	mergeabilityChecker := vcs.NewPullMergeabilityChecker("atlantis")
	client, err := vcs.NewGithubClient("github.com", &vcs.GithubUserCredentials{"user", "pass"}, logging.NewNoopCtxLogger(t), mergeabilityChecker)
	assert.Nil(t, err)

	scope := tally.NewTestScope("test", nil)
	t.Run("should increment commit status counter when checks is not enabled", func(t *testing.T) {
		clientWrapper := checks.ChecksClientWrapper{
			GithubClient: client,
			FeatureAllocator: testFeatureAllocator{
				isChecksEnabled: false,
			},
			Logger: logging.NewNoopCtxLogger(t),
			Scope:  scope,
		}

		_, err = clientWrapper.UpdateStatus(context.TODO(), types.UpdateStatusRequest{})
		for _, counter := range scope.Snapshot().Counters() {
			if counter.Name() == "test.commit_status" {
				assert.Equal(t, counter.Value(), int64(1))
			}
		}
	})

	t.Run("should increment checks counter when checks is enabled", func(t *testing.T) {
		clientWrapper := checks.ChecksClientWrapper{
			GithubClient: client,
			FeatureAllocator: testFeatureAllocator{
				isChecksEnabled: true,
			},
			Logger: logging.NewNoopCtxLogger(t),
			Scope:  scope,
		}

		_, err = clientWrapper.UpdateStatus(context.TODO(), types.UpdateStatusRequest{})
		for _, counter := range scope.Snapshot().Counters() {
			if counter.Name() == "test.checks" {
				assert.Equal(t, counter.Value(), int64(1))
			}
		}
	})

}
