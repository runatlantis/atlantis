package cmd

import (
	"context"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/activities"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/job"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/root"
	"go.temporal.io/sdk/workflow"
)

type executeCommandActivities interface {
	ExecuteCommand(context.Context, activities.ExecuteCommandRequest) (activities.ExecuteCommandResponse, error)
}

type Runner struct {
	Activity executeCommandActivities
}

func (r *Runner) Run(executionContext *job.ExecutionContext, localRoot *root.LocalRoot, step job.Step) (string, error) {
	relPath := localRoot.RelativePathFromRepo()

	envVars := map[string]string{
		"REPO_NAME":    localRoot.Repo.Name,
		"REPO_OWNER":   localRoot.Repo.Owner,
		"DIR":          executionContext.Path,
		"HEAD_COMMIT":  localRoot.Repo.HeadCommit.Ref,
		"PROJECT_NAME": localRoot.Root.Name,
		"REPO_REL_DIR": relPath,
		"USER_NAME":    localRoot.Repo.HeadCommit.Author.Username,
	}

	var resp activities.ExecuteCommandResponse
	err := workflow.ExecuteActivity(executionContext.Context, r.Activity.ExecuteCommand, activities.ExecuteCommandRequest{
		Step:    step,
		Path:    executionContext.Path,
		EnvVars: envVars,
	}).Get(executionContext, &resp)
	if err != nil {
		return "", errors.Wrap(err, "executing activity")
	}

	return resp.Output, nil
}
