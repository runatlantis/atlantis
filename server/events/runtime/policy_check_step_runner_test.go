package runtime_test

import (
	"errors"
	"testing"

	"github.com/hashicorp/go-version"
	. "github.com/petergtz/pegomock"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/runtime"
	"github.com/runatlantis/atlantis/server/events/runtime/mocks"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

func TestRun(t *testing.T) {
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger()
	workspace := "default"
	v, _ := version.NewVersion("1.0")
	executablePath := "some/path/conftest"
	executableArgs := []string{"arg1", "arg2"}

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
		PolicySets: models.PolicySets{
			Version:    v,
			PolicySets: []models.PolicySet{},
		},
	}

	executorWorkflow := mocks.NewMockVersionedExecutorWorkflow()
	s := runtime.NewPolicyCheckStepRunner(executorWorkflow)

	t.Run("success", func(t *testing.T) {
		When(executorWorkflow.EnsureExecutorVersion(logger, v)).ThenReturn(executablePath, nil)
		When(executorWorkflow.ResolveArgs(context)).ThenReturn(executableArgs, nil)
		When(executorWorkflow.Run(logger, executablePath, map[string]string(nil), executableArgs)).ThenReturn("Success!", nil)

		output, err := s.Run(context, []string{"extra", "args"}, "/path", map[string]string(nil))

		Ok(t, err)
		Equals(t, "Success!", output)
	})

	t.Run("ensure version failure", func(t *testing.T) {
		expectedErr := errors.New("error ensuring version")
		When(executorWorkflow.EnsureExecutorVersion(logger, v)).ThenReturn("", expectedErr)

		_, err := s.Run(context, []string{"extra", "args"}, "/path", map[string]string(nil))

		Assert(t, err != nil, "error is not nil")
	})
	t.Run("resolve args failure", func(t *testing.T) {
		When(executorWorkflow.EnsureExecutorVersion(logger, v)).ThenReturn(executablePath, nil)
		When(executorWorkflow.ResolveArgs(context)).ThenReturn(executableArgs, errors.New("error resolving args"))

		_, err := s.Run(context, []string{"extra", "args"}, "/path", map[string]string(nil))

		Assert(t, err != nil, "error is not nil")
	})
	t.Run("executor failure", func(t *testing.T) {
		When(executorWorkflow.EnsureExecutorVersion(logger, v)).ThenReturn(executablePath, nil)
		When(executorWorkflow.ResolveArgs(context)).ThenReturn(executableArgs, nil)
		When(executorWorkflow.Run(logger, executablePath, map[string]string(nil), executableArgs)).ThenReturn("", errors.New("error running executor"))

		_, err := s.Run(context, []string{"extra", "args"}, "/path", map[string]string(nil))

		Assert(t, err != nil, "error is not nil")
	})
}
