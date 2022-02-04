package runtime

import (
	"errors"
	"testing"

	"github.com/hashicorp/go-version"
	. "github.com/petergtz/pegomock"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/core/runtime/mocks"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

func TestRun(t *testing.T) {
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger(t)
	workspace := "default"
	v, _ := version.NewVersion("1.0")
	workdir := "/path"
	executablePath := "some/path/conftest"

	context := models.ProjectCommandContext{
		Log:                logger,
		EscapedCommentArgs: []string{"comment", "args"},
		Workspace:          workspace,
		RepoRelDir:         ".",
		User:               models.User{Username: "username"},
		Pull: models.PullRequest{
			Num: 2,
		},
		BaseRepo: models.Repo{
			FullName: "owner/repo",
			Owner:    "owner",
			Name:     "repo",
		},
		PolicySets: valid.PolicySets{
			Version:    v,
			PolicySets: []valid.PolicySet{},
		},
	}

	executorWorkflow := mocks.NewMockVersionedExecutorWorkflow()
	s := &PolicyCheckStepRunner{
		versionEnsurer: executorWorkflow,
		executor:       executorWorkflow,
	}

	t.Run("success", func(t *testing.T) {
		extraArgs := []string{"extra", "args"}
		When(executorWorkflow.EnsureExecutorVersion(logger, v)).ThenReturn(executablePath, nil)
		When(executorWorkflow.Run(context, executablePath, map[string]string(nil), workdir, extraArgs)).ThenReturn("Success!", nil)

		output, err := s.Run(context, extraArgs, workdir, map[string]string(nil))

		Ok(t, err)
		Equals(t, "Success!", output)
	})

	t.Run("ensure version failure", func(t *testing.T) {
		extraArgs := []string{"extra", "args"}
		expectedErr := errors.New("error ensuring version")
		When(executorWorkflow.EnsureExecutorVersion(logger, v)).ThenReturn("", expectedErr)

		_, err := s.Run(context, extraArgs, workdir, map[string]string(nil))

		Assert(t, err != nil, "error is not nil")
	})
	t.Run("executor failure", func(t *testing.T) {
		extraArgs := []string{"extra", "args"}
		When(executorWorkflow.EnsureExecutorVersion(logger, v)).ThenReturn(executablePath, nil)
		When(executorWorkflow.Run(context, executablePath, map[string]string(nil), workdir, extraArgs)).ThenReturn("", errors.New("error running executor"))

		_, err := s.Run(context, extraArgs, workdir, map[string]string(nil))

		Assert(t, err != nil, "error is not nil")
	})
}
