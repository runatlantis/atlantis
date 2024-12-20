package runtime

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/runatlantis/atlantis/server/core/terraform"
	"github.com/runatlantis/atlantis/server/events/command"
)

// workspaceStepRunnerDelegate ensures that a given step runner run on switched workspace
type workspaceStepRunnerDelegate struct {
	terraformExecutor     TerraformExec
	defaultTfDistribution terraform.Distribution
	defaultTfVersion      *version.Version
	delegate              Runner
}

func NewWorkspaceStepRunnerDelegate(terraformExecutor TerraformExec, defaultTfDistribution terraform.Distribution, defaultTfVersion *version.Version, delegate Runner) Runner {
	return &workspaceStepRunnerDelegate{
		terraformExecutor:     terraformExecutor,
		defaultTfDistribution: defaultTfDistribution,
		defaultTfVersion:      defaultTfVersion,
		delegate:              delegate,
	}
}

func (r *workspaceStepRunnerDelegate) Run(ctx command.ProjectContext, extraArgs []string, path string, envs map[string]string) (string, error) {
	tfDistribution := r.defaultTfDistribution
	tfVersion := r.defaultTfVersion
	if ctx.TerraformDistribution != nil {
		tfDistribution = terraform.NewDistribution(*ctx.TerraformDistribution)
	}
	if ctx.TerraformVersion != nil {
		tfVersion = ctx.TerraformVersion
	}

	// We only need to switch workspaces in version 0.9.*. In older versions,
	// there is no such thing as a workspace so we don't need to do anything.
	if err := r.switchWorkspace(ctx, path, tfDistribution, tfVersion, envs); err != nil {
		return "", err
	}

	return r.delegate.Run(ctx, extraArgs, path, envs)
}

// switchWorkspace changes the terraform workspace if necessary and will create
// it if it doesn't exist. It handles differences between versions.
func (r *workspaceStepRunnerDelegate) switchWorkspace(ctx command.ProjectContext, path string, tfDistribution terraform.Distribution, tfVersion *version.Version, envs map[string]string) error {
	// In versions less than 0.9 there is no support for workspaces.
	noWorkspaceSupport := MustConstraint("<0.9").Check(tfVersion)
	// If the user tried to set a specific workspace in the comment but their
	// version of TF doesn't support workspaces then error out.
	if noWorkspaceSupport && ctx.Workspace != defaultWorkspace {
		return fmt.Errorf("terraform version %s does not support workspaces", tfVersion)
	}
	if noWorkspaceSupport {
		return nil
	}

	// In version 0.9.* the workspace command was called env.
	workspaceCmd := "workspace"
	runningZeroPointNine := MustConstraint(">=0.9,<0.10").Check(tfVersion)
	if runningZeroPointNine {
		workspaceCmd = "env"
	}

	// Use `workspace show` to find out what workspace we're in now. If we're
	// already in the right workspace then no need to switch. This will save us
	// about ten seconds. This command is only available in > 0.10.
	if !runningZeroPointNine {
		workspaceShowOutput, err := r.terraformExecutor.RunCommandWithVersion(ctx, path, []string{workspaceCmd, "show"}, envs, tfDistribution, tfVersion, ctx.Workspace)
		if err != nil {
			return err
		}
		// If `show` says we're already on this workspace then we're done.
		if strings.TrimSpace(workspaceShowOutput) == ctx.Workspace {
			return nil
		}
	}

	// Finally we'll have to select the workspace. We need to figure out if this
	// workspace exists so we can create it if it doesn't.
	// To do this we can either select and catch the error or use list and then
	// look for the workspace. Both commands take the same amount of time so
	// that's why we're running select here.
	_, err := r.terraformExecutor.RunCommandWithVersion(ctx, path, []string{workspaceCmd, "select", ctx.Workspace}, envs, tfDistribution, tfVersion, ctx.Workspace)
	if err != nil {
		// If terraform workspace select fails we run terraform workspace
		// new to create a new workspace automatically.
		out, err := r.terraformExecutor.RunCommandWithVersion(ctx, path, []string{workspaceCmd, "new", ctx.Workspace}, envs, tfDistribution, tfVersion, ctx.Workspace)
		if err != nil {
			return fmt.Errorf("%s: %s", err, out)
		}
	}
	return nil
}
