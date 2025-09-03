package runtime

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/go-version"
	"go.uber.org/mock/gomock"
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

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	terraform := tfclientmocks.NewMockClient(ctrl)
	mockDownloader := mocks.NewMockDownloader(ctrl)
	tfDistribution := tf.NewDistributionTerraformWithDownloader(mockDownloader)
	tfVersion, _ := version.NewVersion("0.15.0")
	s := NewImportStepRunner(terraform, tfDistribution, tfVersion)

	commands := []string{"import", "-var", "foo=bar", "addr", "id"}
	terraform.EXPECT().RunCommandWithVersion(context, tmpDir, commands, map[string]string(nil), tfDistribution, tfVersion, workspace).Return("output", nil)
	output, err := s.Run(context, []string{}, tmpDir, map[string]string(nil))
	Ok(t, err)
	Equals(t, "output", output)
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

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	terraform := tfclientmocks.NewMockClient(ctrl)
	tfVersion, _ := version.NewVersion("0.15.0")
	mockDownloader := mocks.NewMockDownloader(ctrl)
	tfDistribution := tf.NewDistributionTerraformWithDownloader(mockDownloader)
	s := NewImportStepRunner(terraform, tfDistribution, tfVersion)

	// Set up expectations for workspace switching
	terraform.EXPECT().RunCommandWithVersion(context, tmpDir, []string{"workspace", "show"}, map[string]string(nil), tfDistribution, tfVersion, workspace).Return("", nil)
	terraform.EXPECT().RunCommandWithVersion(context, tmpDir, []string{"workspace", "select", workspace}, map[string]string(nil), tfDistribution, tfVersion, workspace).Return("", nil)

	// Set up expectation for import command
	commands := []string{"import", "-var", "foo=bar", "addr", "id"}
	terraform.EXPECT().RunCommandWithVersion(context, tmpDir, commands, map[string]string(nil), tfDistribution, tfVersion, workspace).Return("output", nil)

	output, err := s.Run(context, []string{}, tmpDir, map[string]string(nil))
	Ok(t, err)
	Equals(t, "output", output)

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

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	terraform := tfclientmocks.NewMockClient(ctrl)
	tfVersion, _ := version.NewVersion("0.15.0")
	mockDownloader := mocks.NewMockDownloader(ctrl)
	tfDistribution := tf.NewDistributionTerraformWithDownloader(mockDownloader)
	s := NewImportStepRunner(terraform, tfDistribution, tfVersion)

	// Set up expectations for workspace switching with configured distribution
	terraform.EXPECT().RunCommandWithVersion(gomock.Eq(context), gomock.Eq(tmpDir), gomock.Eq([]string{"workspace", "show"}), gomock.Eq(map[string]string(nil)), gomock.Not(gomock.Eq(tfDistribution)), gomock.Eq(tfVersion), gomock.Eq(workspace)).Return("", nil)
	terraform.EXPECT().RunCommandWithVersion(gomock.Eq(context), gomock.Eq(tmpDir), gomock.Eq([]string{"workspace", "select", workspace}), gomock.Eq(map[string]string(nil)), gomock.Not(gomock.Eq(tfDistribution)), gomock.Eq(tfVersion), gomock.Eq(workspace)).Return("", nil)

	// Set up expectation for import command with configured distribution
	commands := []string{"import", "-var", "foo=bar", "addr", "id"}
	terraform.EXPECT().RunCommandWithVersion(gomock.Eq(context), gomock.Eq(tmpDir), gomock.Eq(commands), gomock.Eq(map[string]string(nil)), gomock.Not(gomock.Eq(tfDistribution)), gomock.Eq(tfVersion), gomock.Eq(workspace)).Return("output", nil)

	output, err := s.Run(context, []string{}, tmpDir, map[string]string(nil))
	Ok(t, err)
	Equals(t, "output", output)

	_, err = os.Stat(planPath)
	Assert(t, os.IsNotExist(err), "planfile should be deleted")
}
