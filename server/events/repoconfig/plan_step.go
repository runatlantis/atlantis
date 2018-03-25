package repoconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// atlantisUserTFVar is the name of the variable we execute terraform
// with, containing the vcs username of who is running the command
const atlantisUserTFVar = "atlantis_user"
const defaultWorkspace = "default"

// PlanStep runs `terraform plan`.
type PlanStep struct {
	ExtraArgs []string
	Meta      StepMeta
}

func (p *PlanStep) Run() (string, error) {
	// We only need to switch workspaces in version 0.9.*. In older versions,
	// there is no such thing as a workspace so we don't need to do anything.
	// In newer versions, the TF_WORKSPACE env var is respected and will handle
	// using the right workspace and even creating it if it doesn't exist.
	// This variable is set inside the Terraform executor.
	if err := p.switchWorkspace(); err != nil {
		return "", err
	}

	planFile := filepath.Join(p.Meta.AbsolutePath, fmt.Sprintf("%s.tfplan", p.Meta.Workspace))
	userVar := fmt.Sprintf("%s=%s", atlantisUserTFVar, p.Meta.Username)
	tfPlanCmd := append(append([]string{"plan", "-refresh", "-no-color", "-out", planFile, "-var", userVar}, p.ExtraArgs...), p.Meta.ExtraCommentArgs...)

	// Check if env/{workspace}.tfvars exist and include it. This is a use-case
	// from Hootsuite where Atlantis was first created so we're keeping this as
	// an homage and a favor so they don't need to refactor all their repos.
	// It's also a nice way to structure your repos to reduce duplication.
	optionalEnvFile := filepath.Join(p.Meta.AbsolutePath, "env", p.Meta.Workspace+".tfvars")
	if _, err := os.Stat(optionalEnvFile); err == nil {
		tfPlanCmd = append(tfPlanCmd, "-var-file", optionalEnvFile)
	}

	return p.Meta.TerraformExecutor.RunCommandWithVersion(p.Meta.Log, filepath.Join(p.Meta.AbsolutePath), tfPlanCmd, p.Meta.TerraformVersion, p.Meta.Workspace)
}

// switchWorkspace changes the terraform workspace if necessary and will create
// it if it doesn't exist. It handles differences between versions.
func (p *PlanStep) switchWorkspace() error {
	// In versions less than 0.9 there is no support for workspaces.
	noWorkspaceSupport := MustConstraint("<0.9").Check(p.Meta.TerraformVersion)
	if noWorkspaceSupport && p.Meta.Workspace != defaultWorkspace {
		return fmt.Errorf("terraform version %s does not support workspaces", p.Meta.TerraformVersion)
	}
	if noWorkspaceSupport {
		return nil
	}

	// In version 0.9.* the workspace command was called env.
	workspaceCmd := "workspace"
	runningZeroPointNine := MustConstraint(">=0.9,<0.10").Check(p.Meta.TerraformVersion)
	if runningZeroPointNine {
		workspaceCmd = "env"
	}

	// Use `workspace show` to find out what workspace we're in now. If we're
	// already in the right workspace then no need to switch. This will save us
	// about ten seconds. This command is only available in > 0.10.
	if !runningZeroPointNine {
		workspaceShowOutput, err := p.Meta.TerraformExecutor.RunCommandWithVersion(p.Meta.Log, p.Meta.AbsolutePath, []string{workspaceCmd, "show"}, p.Meta.TerraformVersion, p.Meta.Workspace)
		if err != nil {
			return err
		}
		// If `show` says we're already on this workspace then we're done.
		if strings.TrimSpace(workspaceShowOutput) == p.Meta.Workspace {
			return nil
		}
	}

	// Finally we'll have to select the workspace. Although we end up running
	// with TF_WORKSPACE set, we need to figure out if this workspace exists so
	// we can create it if it doesn't. To do this we can either select and catch
	// the error or use list and then look for the workspace. Both commands take
	// the same amount of time so that's why we're running select here.
	_, err := p.Meta.TerraformExecutor.RunCommandWithVersion(p.Meta.Log, p.Meta.AbsolutePath, []string{workspaceCmd, "select", "-no-color", p.Meta.Workspace}, p.Meta.TerraformVersion, p.Meta.Workspace)
	if err != nil {
		// If terraform workspace select fails we run terraform workspace
		// new to create a new workspace automatically.
		_, err = p.Meta.TerraformExecutor.RunCommandWithVersion(p.Meta.Log, p.Meta.AbsolutePath, []string{workspaceCmd, "new", "-no-color", p.Meta.Workspace}, p.Meta.TerraformVersion, p.Meta.Workspace)
		return err
	}
	return nil
}
