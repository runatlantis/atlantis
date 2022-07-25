package checks

import (
	"context"
	"fmt"

	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/events/vcs/types"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/lyft/feature"
)

// [WENGINES-4643] TODO: Remove this wrapper and add checks implementation to UpdateStatus() directly after github checks is stable
type ChecksClientWrapper struct {
	*vcs.GithubClient
	FeatureAllocator feature.Allocator
	Logger           logging.Logger
}

func (c *ChecksClientWrapper) UpdateStatus(ctx context.Context, request types.UpdateStatusRequest) error {
	shouldAllocate, err := c.FeatureAllocator.ShouldAllocate(feature.GithubChecks, feature.FeatureContext{
		RepoName:         request.Repo.FullName,
		PullCreationTime: request.PullCreationTime,
	})
	if err != nil {
		c.Logger.ErrorContext(ctx, fmt.Sprintf("unable to allocate for feature: %s", feature.GithubChecks), map[string]interface{}{
			"error": err.Error(),
		})
	}

	if !shouldAllocate {
		return c.GithubClient.UpdateStatus(ctx, request)
	}

	return c.GithubClient.UpdateChecksStatus(ctx, request)
}
