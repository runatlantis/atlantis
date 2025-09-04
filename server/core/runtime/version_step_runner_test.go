package runtime

import (
	"testing"

	"github.com/hashicorp/go-version"
	"go.uber.org/mock/gomock"
	tf "github.com/runatlantis/atlantis/server/core/terraform"
	"github.com/runatlantis/atlantis/server/core/terraform/mocks"
	tfclientmocks "github.com/runatlantis/atlantis/server/core/terraform/tfclient/mocks"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

func TestRunVersionStep(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	logger := logging.NewNoopLogger(t)
	workspace := "default"

	context := command.ProjectContext{
		Log:                logger,
		EscapedCommentArgs: []string{"comment", "args"},
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

	terraform := tfclientmocks.NewMockClient(ctrl)
	mockDownloader := mocks.NewMockDownloader(ctrl)
	tfDistribution := tf.NewDistributionTerraformWithDownloader(mockDownloader)
	tfVersion, _ := version.NewVersion("0.15.0")
	tmpDir := t.TempDir()

	s := &VersionStepRunner{
		TerraformExecutor:     terraform,
		DefaultTFDistribution: tfDistribution,
		DefaultTFVersion:      tfVersion,
	}

	t.Run("ensure runs", func(t *testing.T) {
		terraform.EXPECT().RunCommandWithVersion(context, tmpDir, []string{"version"}, map[string]string(nil), tfDistribution, tfVersion, "default").
			Return("", nil)
		_, err := s.Run(context, []string{}, tmpDir, map[string]string(nil))
		Ok(t, err)
	})
}

func TestVersionStepRunner_Run_UsesConfiguredDistribution(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	logger := logging.NewNoopLogger(t)
	workspace := "default"
	projTFDistribution := "opentofu"
	context := command.ProjectContext{
		Log:                logger,
		EscapedCommentArgs: []string{"comment", "args"},
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
		TerraformDistribution: &projTFDistribution,
	}

	terraform := tfclientmocks.NewMockClient(ctrl)
	mockDownloader := mocks.NewMockDownloader(ctrl)
	tfDistribution := tf.NewDistributionTerraformWithDownloader(mockDownloader)
	tfVersion, _ := version.NewVersion("0.15.0")
	tmpDir := t.TempDir()

	s := &VersionStepRunner{
		TerraformExecutor:     terraform,
		DefaultTFDistribution: tfDistribution,
		DefaultTFVersion:      tfVersion,
	}

	t.Run("ensure runs", func(t *testing.T) {
		// Use gomock.Not to assert that distribution is not the default
		notDefaultDistribution := gomock.Not(gomock.Eq(tfDistribution))
		terraform.EXPECT().RunCommandWithVersion(context, tmpDir, []string{"version"}, map[string]string(nil), notDefaultDistribution, tfVersion, "default").
			Return("", nil)
		_, err := s.Run(context, []string{}, tmpDir, map[string]string(nil))
		Ok(t, err)
	})
}