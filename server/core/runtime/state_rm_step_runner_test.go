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

func TestStateRmStepRunner_Run_Success(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	workspace := "default"
	tmpDir := t.TempDir()
	planPath := filepath.Join(tmpDir, fmt.Sprintf("%s.tfplan", workspace))
	err := os.WriteFile(planPath, nil, 0600)
	Ok(t, err)

	context := command.ProjectContext{
		Log:                logger,
		EscapedCommentArgs: []string{"-lock=false", "addr1", "addr2", "addr3"},
		Workspace:          workspace,
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	terraform := tfclientmocks.NewMockClient(ctrl)
	tfVersion, _ := version.NewVersion("0.15.0")
	mockDownloader := mocks.NewMockDownloader(ctrl)
	tfDistribution := tf.NewDistributionTerraformWithDownloader(mockDownloader)
	s := NewStateRmStepRunner(terraform, tfDistribution, tfVersion)

	commands := []string{"state", "rm", "-lock=false", "addr1", "addr2", "addr3"}
	terraform.EXPECT().RunCommandWithVersion(context, tmpDir, commands, map[string]string(nil), tfDistribution, tfVersion, "default").
		Return("output", nil)
	
	output, err := s.Run(context, []string{}, tmpDir, map[string]string(nil))
	Ok(t, err)
	Equals(t, "output", output)
	
	_, err = os.Stat(planPath)
	Assert(t, os.IsNotExist(err), "planfile should be deleted")
}

func TestStateRmStepRunner_Run_Workspace(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	workspace := "something"
	tmpDir := t.TempDir()
	planPath := filepath.Join(tmpDir, fmt.Sprintf("%s.tfplan", workspace))
	err := os.WriteFile(planPath, nil, 0600)
	Ok(t, err)

	context := command.ProjectContext{
		Log:                logger,
		EscapedCommentArgs: []string{"-lock=false", "addr1", "addr2", "addr3"},
		Workspace:          workspace,
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	terraform := tfclientmocks.NewMockClient(ctrl)
	tfVersion, _ := version.NewVersion("0.15.0")
	mockDownloader := mocks.NewMockDownloader(ctrl)
	tfDistribution := tf.NewDistributionTerraformWithDownloader(mockDownloader)
	s := NewStateRmStepRunner(terraform, tfDistribution, tfVersion)

	// Set up expectations in order
	// First: workspace show
	terraform.EXPECT().RunCommandWithVersion(context, tmpDir, []string{"workspace", "show"}, map[string]string(nil), tfDistribution, tfVersion, workspace).
		Return("default", nil)
	// Second: workspace select
	terraform.EXPECT().RunCommandWithVersion(context, tmpDir, []string{"workspace", "select", workspace}, map[string]string(nil), tfDistribution, tfVersion, workspace).
		Return("", nil)
	// Third: state rm command
	commands := []string{"state", "rm", "-lock=false", "addr1", "addr2", "addr3"}
	terraform.EXPECT().RunCommandWithVersion(context, tmpDir, commands, map[string]string(nil), tfDistribution, tfVersion, workspace).
		Return("output", nil)
	
	output, err := s.Run(context, []string{}, tmpDir, map[string]string(nil))
	Ok(t, err)
	Equals(t, "output", output)

	_, err = os.Stat(planPath)
	Assert(t, os.IsNotExist(err), "planfile should be deleted")
}

func TestStateRmStepRunner_Run_UsesConfiguredDistribution(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	workspace := "something"
	tmpDir := t.TempDir()
	planPath := filepath.Join(tmpDir, fmt.Sprintf("%s.tfplan", workspace))
	err := os.WriteFile(planPath, nil, 0600)
	Ok(t, err)

	projTFDistribution := "opentofu"

	context := command.ProjectContext{
		Log:                   logger,
		EscapedCommentArgs:    []string{"-lock=false", "addr1", "addr2", "addr3"},
		Workspace:             workspace,
		TerraformDistribution: &projTFDistribution,
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	terraform := tfclientmocks.NewMockClient(ctrl)
	tfVersion, _ := version.NewVersion("0.15.0")
	mockDownloader := mocks.NewMockDownloader(ctrl)
	tfDistribution := tf.NewDistributionTerraformWithDownloader(mockDownloader)
	s := NewStateRmStepRunner(terraform, tfDistribution, tfVersion)

	// Create a custom matcher for checking distribution is not the default one
	notDefaultDistribution := gomock.Not(gomock.Eq(tfDistribution))
	
	// Set up expectations in order
	// First: workspace show with custom distribution
	terraform.EXPECT().RunCommandWithVersion(context, tmpDir, []string{"workspace", "show"}, map[string]string(nil), notDefaultDistribution, tfVersion, workspace).
		Return("default", nil)
	// Second: workspace select with custom distribution
	terraform.EXPECT().RunCommandWithVersion(context, tmpDir, []string{"workspace", "select", workspace}, map[string]string(nil), notDefaultDistribution, tfVersion, workspace).
		Return("", nil)
	// Third: state rm command with custom distribution
	commands := []string{"state", "rm", "-lock=false", "addr1", "addr2", "addr3"}
	terraform.EXPECT().RunCommandWithVersion(context, tmpDir, commands, map[string]string(nil), notDefaultDistribution, tfVersion, workspace).
		Return("output", nil)
	
	output, err := s.Run(context, []string{}, tmpDir, map[string]string(nil))
	Ok(t, err)
	Equals(t, "output", output)

	_, err = os.Stat(planPath)
	Assert(t, os.IsNotExist(err), "planfile should be deleted")
}