package runtime

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/go-version"
	. "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/server/core/terraform/mocks"
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

	RegisterMockTestingT(t)
	terraform := mocks.NewMockClient()
	tfVersion, _ := version.NewVersion("0.15.0")
	s := NewStateRmStepRunner(terraform, tfVersion)

	When(terraform.RunCommandWithVersion(Any[command.ProjectContext](), Any[string](), Any[[]string](), Any[map[string]string](), Any[*version.Version](), Any[string]())).
		ThenReturn("output", nil)
	output, err := s.Run(context, []string{}, tmpDir, map[string]string(nil))
	Ok(t, err)
	Equals(t, "output", output)
	commands := []string{"state", "rm", "-lock=false", "addr1", "addr2", "addr3"}
	terraform.VerifyWasCalledOnce().RunCommandWithVersion(context, tmpDir, commands, map[string]string(nil), tfVersion, "default")
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

	RegisterMockTestingT(t)
	terraform := mocks.NewMockClient()
	tfVersion, _ := version.NewVersion("0.15.0")
	s := NewStateRmStepRunner(terraform, tfVersion)

	When(terraform.RunCommandWithVersion(Any[command.ProjectContext](), Any[string](), Any[[]string](), Any[map[string]string](), Any[*version.Version](), Any[string]())).
		ThenReturn("output", nil)
	output, err := s.Run(context, []string{}, tmpDir, map[string]string(nil))
	Ok(t, err)
	Equals(t, "output", output)

	// switch workspace
	terraform.VerifyWasCalledOnce().RunCommandWithVersion(context, tmpDir, []string{"workspace", "show"}, map[string]string(nil), tfVersion, workspace)
	terraform.VerifyWasCalledOnce().RunCommandWithVersion(context, tmpDir, []string{"workspace", "select", workspace}, map[string]string(nil), tfVersion, workspace)

	// exec state rm
	commands := []string{"state", "rm", "-lock=false", "addr1", "addr2", "addr3"}
	terraform.VerifyWasCalledOnce().RunCommandWithVersion(context, tmpDir, commands, map[string]string(nil), tfVersion, workspace)

	_, err = os.Stat(planPath)
	Assert(t, os.IsNotExist(err), "planfile should be deleted")
}
