package runtime

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/runatlantis/atlantis/server/events/models"
)

// atlantisUserTFVar is the name of the variable we execute terraform
// with, containing the vcs username of who is running the command
const atlantisUserTFVar = "atlantis_user"
// atlantisRepoTFVar is the name of the variable we execute terraform
// in, containing the vcs repo name of who is running the command
const atlantisRepoTFVar = "atlantis_repo"
const defaultWorkspace = "default"

type PlanStepRunner struct {
	TerraformExecutor TerraformExec
	DefaultTFVersion  *version.Version
}

func (p *PlanStepRunner) Run(ctx models.ProjectCommandContext, extraArgs []string, path string) (string, error) {
	tfVersion := p.DefaultTFVersion
	if ctx.ProjectConfig != nil && ctx.ProjectConfig.TerraformVersion != nil {
		tfVersion = ctx.ProjectConfig.TerraformVersion
	}

	// We only need to switch workspaces in version 0.9.*. In older versions,
	// there is no such thing as a workspace so we don't need to do anything.
	if err := p.switchWorkspace(ctx, path, tfVersion); err != nil {
		return "", err
	}

	planFile := filepath.Join(path, GetPlanFilename(ctx.Workspace, ctx.ProjectConfig))
	userVar := fmt.Sprintf("%s=%s", atlantisUserTFVar, ctx.User.Username)
	repoVar := fmt.Sprintf("%s=%s", atlantisiRepoTFVar, ctx.BaseRepo.FullName)
	tfPlanCmd := append(append([]string{"plan", "-input=false", "-refresh", "-no-color", "-out", planFile, "-var", userVar, "-var", repoVar}, extraArgs...), ctx.CommentArgs...)

	// Check if env/{workspace}.tfvars exist and include it. This is a use-case
	// from Hootsuite where Atlantis was first created so we're keeping this as
	// an homage and a favor so they don't need to refactor all their repos.
	// It's also a nice way to structure your repos to reduce duplication.
	optionalEnvFile := filepath.Join(path, "env", ctx.Workspace+".tfvars")
	if _, err := os.Stat(optionalEnvFile); err == nil {
		tfPlanCmd = append(tfPlanCmd, "-var-file", optionalEnvFile)
	}

	return p.TerraformExecutor.RunCommandWithVersion(ctx.Log, filepath.Join(path), tfPlanCmd, tfVersion, ctx.Workspace)
}

// switchWorkspace changes the terraform workspace if necessary and will create
// it if it doesn't exist. It handles differences between versions.
func (p *PlanStepRunner) switchWorkspace(ctx models.ProjectCommandContext, path string, tfVersion *version.Version) error {
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
		workspaceShowOutput, err := p.TerraformExecutor.RunCommandWithVersion(ctx.Log, path, []string{workspaceCmd, "show"}, tfVersion, ctx.Workspace)
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
	_, err := p.TerraformExecutor.RunCommandWithVersion(ctx.Log, path, []string{workspaceCmd, "select", "-no-color", ctx.Workspace}, tfVersion, ctx.Workspace)
	if err != nil {
		// If terraform workspace select fails we run terraform workspace
		// new to create a new workspace automatically.
		_, err = p.TerraformExecutor.RunCommandWithVersion(ctx.Log, path, []string{workspaceCmd, "new", "-no-color", ctx.Workspace}, tfVersion, ctx.Workspace)
		return err
	}
	return nil
}
