package preworkflow_test

import (
	"errors"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/models/fixtures"
	"github.com/runatlantis/atlantis/server/neptune/gateway/event/preworkflow"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRunPreHooks_Clone(t *testing.T) {
	testHook := &valid.PreWorkflowHook{
		StepName:   "test",
		RunCommand: "some command",
	}
	repoDir := "path/to/repo"

	t.Run("success hooks in cfg", func(t *testing.T) {
		globalCfg := valid.GlobalCfg{
			Repos: []valid.Repo{
				{
					ID: fixtures.GithubRepo.ID(),
					PreWorkflowHooks: []*valid.PreWorkflowHook{
						testHook,
					},
				},
			},
		}
		wh := preworkflow.HooksRunner{
			HookExecutor: &MockSuccessPreWorkflowHookExecutor{},
			GlobalCfg:    globalCfg,
		}
		err := wh.Run(fixtures.GithubRepo, repoDir)
		assert.NoError(t, err)
	})
	t.Run("success hooks not in cfg", func(t *testing.T) {
		globalCfg := valid.GlobalCfg{
			Repos: []valid.Repo{
				// one with hooks but mismatched id
				{
					ID: "id1",
					PreWorkflowHooks: []*valid.PreWorkflowHook{
						testHook,
					},
				},
				// one with the correct id but no hooks
				{
					ID:               fixtures.GithubRepo.ID(),
					PreWorkflowHooks: []*valid.PreWorkflowHook{},
				},
			},
		}
		wh := preworkflow.HooksRunner{
			HookExecutor: &MockSuccessPreWorkflowHookExecutor{},
			GlobalCfg:    globalCfg,
		}
		err := wh.Run(fixtures.GithubRepo, repoDir)
		assert.NoError(t, err)
	})
	t.Run("error running pre hook", func(t *testing.T) {
		globalCfg := valid.GlobalCfg{
			Repos: []valid.Repo{
				{
					ID: fixtures.GithubRepo.ID(),
					PreWorkflowHooks: []*valid.PreWorkflowHook{
						testHook,
					},
				},
			},
		}
		wh := preworkflow.HooksRunner{
			HookExecutor: &MockFailurePreWorkflowHookExecutor{},
			GlobalCfg:    globalCfg,
		}
		err := wh.Run(fixtures.GithubRepo, repoDir)
		assert.Error(t, err, "error not nil")
	})
}

type MockSuccessPreWorkflowHookExecutor struct {
}

func (m *MockSuccessPreWorkflowHookExecutor) Execute(_ *valid.PreWorkflowHook, _ models.Repo, _ string) error {
	return nil
}

type MockFailurePreWorkflowHookExecutor struct {
}

func (m *MockFailurePreWorkflowHookExecutor) Execute(_ *valid.PreWorkflowHook, _ models.Repo, _ string) error {
	return errors.New("some error")
}
