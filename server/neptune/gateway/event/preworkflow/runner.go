package preworkflow

import (
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/events/models"
)

type executor interface {
	Execute(hook *valid.PreWorkflowHook, repo models.Repo, path string) error
}

type HooksRunner struct {
	GlobalCfg    valid.GlobalCfg
	HookExecutor executor
}

func (r *HooksRunner) Run(baseRepo models.Repo, repoDir string) error {
	preWorkflowHooks := make([]*valid.PreWorkflowHook, 0)
	for _, repo := range r.GlobalCfg.Repos {
		if repo.IDMatches(baseRepo.ID()) && len(repo.PreWorkflowHooks) > 0 {
			preWorkflowHooks = append(preWorkflowHooks, repo.PreWorkflowHooks...)
		}
	}

	// short circuit any other calls if there are no pre-hooks configured
	if len(preWorkflowHooks) == 0 {
		return nil
	}

	// uses default zero values for some field in PreWorkflowHookCommandContext struct since they aren't relevant to fxn
	for _, hook := range preWorkflowHooks {
		err := r.HookExecutor.Execute(hook, baseRepo, repoDir)
		if err != nil {
			return errors.Wrap(err, "running pre workflow hooks")
		}
	}
	return nil
}
