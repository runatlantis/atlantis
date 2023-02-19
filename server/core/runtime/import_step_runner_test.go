package runtime

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/go-version"
	. "github.com/petergtz/pegomock"
	runtimematchers "github.com/runatlantis/atlantis/server/core/runtime/mocks/matchers"
	"github.com/runatlantis/atlantis/server/core/terraform/mocks"
	tfmatchers "github.com/runatlantis/atlantis/server/core/terraform/mocks/matchers"
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
	terraform := mocks.NewMockClient()
	tfVersion, _ := version.NewVersion("0.15.0")
	s := NewImportStepRunner(terraform, tfVersion)

	When(terraform.RunCommandWithVersion(runtimematchers.AnyCommandProjectContext(), AnyString(), AnyStringSlice(), tfmatchers.AnyMapOfStringToString(), tfmatchers.AnyPtrToGoVersionVersion(), AnyString())).
		ThenReturn("output", nil)
	output, err := s.Run(context, []string{}, tmpDir, map[string]string(nil))
	Ok(t, err)
	Equals(t, "output", output)
	commands := []string{"import", "-var", "foo=bar", "addr", "id"}
	terraform.VerifyWasCalledOnce().RunCommandWithVersion(context, tmpDir, commands, map[string]string(nil), tfVersion, workspace)
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
	terraform := mocks.NewMockClient()
	tfVersion, _ := version.NewVersion("0.15.0")
	s := NewImportStepRunner(terraform, tfVersion)

	When(terraform.RunCommandWithVersion(runtimematchers.AnyCommandProjectContext(), AnyString(), AnyStringSlice(), runtimematchers.AnyMapOfStringToString(), runtimematchers.AnyPtrToGoVersionVersion(), AnyString())).
		ThenReturn("output", nil)
	output, err := s.Run(context, []string{}, tmpDir, map[string]string(nil))
	Ok(t, err)
	Equals(t, "output", output)

	// switch workspace
	terraform.VerifyWasCalledOnce().RunCommandWithVersion(context, tmpDir, []string{"workspace", "show"}, map[string]string(nil), tfVersion, workspace)
	terraform.VerifyWasCalledOnce().RunCommandWithVersion(context, tmpDir, []string{"workspace", "select", workspace}, map[string]string(nil), tfVersion, workspace)

	// exec import
	commands := []string{"import", "-var", "foo=bar", "addr", "id"}
	terraform.VerifyWasCalledOnce().RunCommandWithVersion(context, tmpDir, commands, map[string]string(nil), tfVersion, workspace)

	_, err = os.Stat(planPath)
	Assert(t, os.IsNotExist(err), "planfile should be deleted")
}
