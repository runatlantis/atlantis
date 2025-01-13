package runtime

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/go-version"
	. "github.com/petergtz/pegomock/v4"
	tf "github.com/runatlantis/atlantis/server/core/terraform"
	"github.com/runatlantis/atlantis/server/core/terraform/mocks"
	tfclientmocks "github.com/runatlantis/atlantis/server/core/terraform/tfclient/mocks"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

func TestImportStepRunner_Run_Success(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	workspace := "default"
	tmpDir := t.TempDir()
	planPath := filepath.Join(tmpDir, fmt.Sprintf("%s.tfplan", workspace))
	err := os.WriteFile(planPath, nil, 0600)
	Ok(t, err)

	context := command.ProjectContext{
		Log:                logger,
		EscapedCommentArgs: []string{"-var", "foo=bar", "addr", "id"},
		Workspace:          workspace,
	}

	RegisterMockTestingT(t)
	terraform := tfclientmocks.NewMockClient()
	mockDownloader := mocks.NewMockDownloader()
	tfDistribution := tf.NewDistributionTerraformWithDownloader(mockDownloader)
	tfVersion, _ := version.NewVersion("0.15.0")
	s := NewImportStepRunner(terraform, tfDistribution, tfVersion)

	When(terraform.RunCommandWithVersion(Any[command.ProjectContext](), Any[string](), Any[[]string](), Any[map[string]string](), Any[tf.Distribution](), Any[*version.Version](), Any[string]())).
		ThenReturn("output", nil)
	output, err := s.Run(context, []string{}, tmpDir, map[string]string(nil))
	Ok(t, err)
	Equals(t, "output", output)
	commands := []string{"import", "-var", "foo=bar", "addr", "id"}
	terraform.VerifyWasCalledOnce().RunCommandWithVersion(context, tmpDir, commands, map[string]string(nil), tfDistribution, tfVersion, workspace)
	_, err = os.Stat(planPath)
	Assert(t, os.IsNotExist(err), "planfile should be deleted")
}

func TestImportStepRunner_Run_Workspace(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	workspace := "something"
	tmpDir := t.TempDir()
	planPath := filepath.Join(tmpDir, fmt.Sprintf("%s.tfplan", workspace))
	err := os.WriteFile(planPath, nil, 0600)
	Ok(t, err)

	context := command.ProjectContext{
		Log:                logger,
		EscapedCommentArgs: []string{"-var", "foo=bar", "addr", "id"},
		Workspace:          workspace,
	}

	RegisterMockTestingT(t)
	terraform := tfclientmocks.NewMockClient()
	tfVersion, _ := version.NewVersion("0.15.0")
	mockDownloader := mocks.NewMockDownloader()
	tfDistribution := tf.NewDistributionTerraformWithDownloader(mockDownloader)
	s := NewImportStepRunner(terraform, tfDistribution, tfVersion)

	When(terraform.RunCommandWithVersion(Any[command.ProjectContext](), Any[string](), Any[[]string](), Any[map[string]string](), Any[tf.Distribution](), Any[*version.Version](), Any[string]())).
		ThenReturn("output", nil)
	output, err := s.Run(context, []string{}, tmpDir, map[string]string(nil))
	Ok(t, err)
	Equals(t, "output", output)

	// switch workspace
	terraform.VerifyWasCalledOnce().RunCommandWithVersion(context, tmpDir, []string{"workspace", "show"}, map[string]string(nil), tfDistribution, tfVersion, workspace)
	terraform.VerifyWasCalledOnce().RunCommandWithVersion(context, tmpDir, []string{"workspace", "select", workspace}, map[string]string(nil), tfDistribution, tfVersion, workspace)

	// exec import
	commands := []string{"import", "-var", "foo=bar", "addr", "id"}
	terraform.VerifyWasCalledOnce().RunCommandWithVersion(context, tmpDir, commands, map[string]string(nil), tfDistribution, tfVersion, workspace)

	_, err = os.Stat(planPath)
	Assert(t, os.IsNotExist(err), "planfile should be deleted")
}

func TestImportStepRunner_Run_UsesConfiguredDistribution(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	workspace := "something"
	tmpDir := t.TempDir()
	planPath := filepath.Join(tmpDir, fmt.Sprintf("%s.tfplan", workspace))
	err := os.WriteFile(planPath, nil, 0600)
	Ok(t, err)

	projTFDistribution := "opentofu"
	context := command.ProjectContext{
		Log:                   logger,
		EscapedCommentArgs:    []string{"-var", "foo=bar", "addr", "id"},
		Workspace:             workspace,
		TerraformDistribution: &projTFDistribution,
	}

	RegisterMockTestingT(t)
	terraform := tfclientmocks.NewMockClient()
	tfVersion, _ := version.NewVersion("0.15.0")
	mockDownloader := mocks.NewMockDownloader()
	tfDistribution := tf.NewDistributionTerraformWithDownloader(mockDownloader)
	s := NewImportStepRunner(terraform, tfDistribution, tfVersion)

	When(terraform.RunCommandWithVersion(Any[command.ProjectContext](), Any[string](), Any[[]string](), Any[map[string]string](), Any[tf.Distribution](), Any[*version.Version](), Any[string]())).
		ThenReturn("output", nil)
	output, err := s.Run(context, []string{}, tmpDir, map[string]string(nil))
	Ok(t, err)
	Equals(t, "output", output)

	// switch workspace
	terraform.VerifyWasCalledOnce().RunCommandWithVersion(Eq(context), Eq(tmpDir), Eq([]string{"workspace", "show"}), Eq(map[string]string(nil)), NotEq(tfDistribution), Eq(tfVersion), Eq(workspace))
	terraform.VerifyWasCalledOnce().RunCommandWithVersion(Eq(context), Eq(tmpDir), Eq([]string{"workspace", "select", workspace}), Eq(map[string]string(nil)), NotEq(tfDistribution), Eq(tfVersion), Eq(workspace))

	// exec import
	commands := []string{"import", "-var", "foo=bar", "addr", "id"}
	terraform.VerifyWasCalledOnce().RunCommandWithVersion(Eq(context), Eq(tmpDir), Eq(commands), Eq(map[string]string(nil)), NotEq(tfDistribution), Eq(tfVersion), Eq(workspace))

	_, err = os.Stat(planPath)
	Assert(t, os.IsNotExist(err), "planfile should be deleted")
}
