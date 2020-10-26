package runtime_test

import (
	"testing"

	. "github.com/petergtz/pegomock"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/runtime"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

func TestRun_PolicyRunSuccess(t *testing.T) {
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger()
	workspace := "default"
	s := runtime.PolicyCheckStepRunner{}

	output, err := s.Run(models.ProjectCommandContext{
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
	}, []string{"extra", "args"}, "/path", map[string]string(nil))
	Ok(t, err)

	Equals(t, "Success!", output)
}
