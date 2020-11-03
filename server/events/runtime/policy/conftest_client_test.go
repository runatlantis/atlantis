package policy

import (
	"errors"
	"testing"

	"github.com/hashicorp/go-version"
	. "github.com/petergtz/pegomock"
	"github.com/runatlantis/atlantis/server/events/runtime/cache/mocks"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

func TestEnsureExecutorVersion(t *testing.T) {

	defaultVersion, _ := version.NewVersion("1.0")
	expectedPath := "some/path"

	RegisterMockTestingT(t)

	mockCache := mocks.NewMockExecutionVersionCache()
	log := logging.NewNoopLogger()

	t.Run("no specified version or default version", func(t *testing.T) {
		subject := &ConfTestExecutorWorkflow{
			VersionCache: mockCache,
		}

		_, err := subject.EnsureExecutorVersion(log, nil)

		Assert(t, err != nil, "expected error finding version")
	})

	t.Run("use default version", func(t *testing.T) {
		subject := &ConfTestExecutorWorkflow{
			VersionCache:           mockCache,
			DefaultConftestVersion: defaultVersion,
		}

		When(mockCache.Get(defaultVersion)).ThenReturn(expectedPath, nil)

		path, err := subject.EnsureExecutorVersion(log, nil)

		Ok(t, err)

		Assert(t, path == expectedPath, "path is expected")
	})

	t.Run("use specified version", func(t *testing.T) {
		subject := &ConfTestExecutorWorkflow{
			VersionCache:           mockCache,
			DefaultConftestVersion: defaultVersion,
		}

		versionInput, _ := version.NewVersion("2.0")

		When(mockCache.Get(versionInput)).ThenReturn(expectedPath, nil)

		path, err := subject.EnsureExecutorVersion(log, versionInput)

		Ok(t, err)

		Assert(t, path == expectedPath, "path is expected")
	})

	t.Run("cache error", func(t *testing.T) {
		subject := &ConfTestExecutorWorkflow{
			VersionCache:           mockCache,
			DefaultConftestVersion: defaultVersion,
		}

		versionInput, _ := version.NewVersion("2.0")

		When(mockCache.Get(versionInput)).ThenReturn(expectedPath, errors.New("some err"))

		_, err := subject.EnsureExecutorVersion(log, versionInput)

		Assert(t, err != nil, "path is expected")
	})

}
