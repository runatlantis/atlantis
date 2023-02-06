package job

import (
	"context"
	"sort"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/execute"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/terraform"
	"go.temporal.io/sdk/workflow"
)

func GetDefaultEnvVars(ctx *ExecutionContext, localRoot *terraform.LocalRoot) map[string]string {
	relPath := localRoot.RelativePathFromRepo()
	return map[string]string{
		"BASE_REPO_NAME":  localRoot.Repo.Name,
		"BASE_REPO_OWNER": localRoot.Repo.Owner,
		"DIR":             ctx.Path,
		"PROJECT_NAME":    localRoot.Root.Name,
		"REPO_REL_DIR":    relPath,
	}
}

type executeCommandActivities interface {
	ExecuteCommand(context.Context, activities.ExecuteCommandRequest) (activities.ExecuteCommandResponse, error)
}

type CmdStepRunner struct {
	Activity executeCommandActivities
}

func getDefaultStringEnvs(ctx *ExecutionContext, localRoot *terraform.LocalRoot) []StringEnvVar {
	var envs []StringEnvVar
	for k, v := range GetDefaultEnvVars(ctx, localRoot) {
		envs = append(envs, NewEnvVarFromString(k, v))
	}

	// sort so our activity inputs are deterministic
	sort.Slice(envs, func(i, j int) bool {
		return envs[i].name < envs[j].name
	})

	return envs
}

func (r *CmdStepRunner) Run(executionContext *ExecutionContext, localRoot *terraform.LocalRoot, step execute.Step) (string, error) {
	var envs []EnvVar
	for _, e := range getDefaultStringEnvs(executionContext, localRoot) {
		envs = append(envs, e)
	}

	envs = append(envs, executionContext.Envs...)

	var resp activities.ExecuteCommandResponse
	err := workflow.ExecuteActivity(executionContext.Context, r.Activity.ExecuteCommand, activities.ExecuteCommandRequest{
		Step:           step,
		Path:           executionContext.Path,
		DynamicEnvVars: toActivityEnvs(envs),
		EnvVars:        map[string]string{},
	}).Get(executionContext, &resp)
	if err != nil {
		return "", errors.Wrap(err, "executing activity")
	}

	return resp.Output, nil
}

func toActivityEnvs(envs []EnvVar) []activities.EnvVar {
	var result []activities.EnvVar
	for _, e := range envs {
		result = append(result, e.ToActivityEnvVar())
	}
	return result
}
