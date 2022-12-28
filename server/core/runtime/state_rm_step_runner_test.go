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
	"github.com/runatlantis/atlantis/server/events/models"
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
		RepoRelDir:         ".",
		User:               models.User{Username: "username"},
		Pull: models.PullRequest{
			Num: 2,
		},
		BaseRepo: models.Repo{
			FullName: "owner/repo",
			Owner:    "owner",
			Name:     "repo",
		},
	}

	RegisterMockTestingT(t)
	terraform := mocks.NewMockClient()
	tfVersion, _ := version.NewVersion("0.15.0")
	s := &StateRmStepRunner{
		TerraformExecutor: terraform,
		DefaultTFVersion:  tfVersion,
	}

	When(terraform.RunCommandWithVersion(runtimematchers.AnyCommandProjectContext(), AnyString(), AnyStringSlice(), tfmatchers.AnyMapOfStringToString(), tfmatchers.AnyPtrToGoVersionVersion(), AnyString())).
		ThenReturn("output", nil)
	output, err := s.Run(context, []string{}, tmpDir, map[string]string(nil))
	Ok(t, err)
	Equals(t, "output", output)
	commands := []string{"state", "rm", "-lock=false", "addr1", "addr2", "addr3"}
	terraform.VerifyWasCalledOnce().RunCommandWithVersion(context, tmpDir, commands, map[string]string(nil), tfVersion, "default")
	_, err = os.Stat(planPath)
	Assert(t, os.IsNotExist(err), "planfile should be deleted")
}
