package runtime_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	version "github.com/hashicorp/go-version"
	"go.uber.org/mock/gomock"
	"github.com/pkg/errors"

	"github.com/runatlantis/atlantis/server/core/runtime"
	tf "github.com/runatlantis/atlantis/server/core/terraform"
	"github.com/runatlantis/atlantis/server/core/terraform/mocks"
	tfclientmocks "github.com/runatlantis/atlantis/server/core/terraform/tfclient/mocks"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

func TestRun_UsesGetOrInitForRightVersion(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDownloader := mocks.NewMockDownloader(ctrl)
	tfDistribution := tf.NewDistributionTerraformWithDownloader(mockDownloader)
	cases := []struct {
		version string
		expCmd  string
	}{
		{
			"0.8.9",
			"get",
		},
		{
			"0.9.0",
			"init",
		},
		{
			"0.9.1",
			"init",
		},
		{
			"0.10.0",
			"init",
		},
	}

	for _, c := range cases {
		t.Run(c.version, func(t *testing.T) {
			terraform := tfclientmocks.NewMockClient(ctrl)

			logger := logging.NewNoopLogger(t)
			ctx := command.ProjectContext{
				Workspace:  "workspace",
				RepoRelDir: ".",
				Log:        logger,
			}

			tfVersion, _ := version.NewVersion(c.version)
			iso := runtime.InitStepRunner{
				TerraformExecutor:     terraform,
				DefaultTFDistribution: tfDistribution,
				DefaultTFVersion:      tfVersion,
			}
			
			// If using init then we specify -input=false but not for get.
			expArgs := []string{c.expCmd, "-input=false", "-upgrade", "extra", "args"}
			if c.expCmd == "get" {
				expArgs = []string{c.expCmd, "-upgrade", "extra", "args"}
			}
			
			terraform.EXPECT().RunCommandWithVersion(
				ctx, "/path", expArgs, map[string]string(nil), tfDistribution, tfVersion, "workspace").
				Return("output", nil)

			output, err := iso.Run(ctx, []string{"extra", "args"}, "/path", map[string]string(nil))
			Ok(t, err)
			// When there is no error, should not return init output to PR.
			Equals(t, "", output)
		})
	}
}

func TestInitStepRunner_TestRun_UsesConfiguredDistribution(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDownloader := mocks.NewMockDownloader(ctrl)
	tfDistribution := tf.NewDistributionTerraformWithDownloader(mockDownloader)
	cases := []struct {
		version      string
		distribution string
		expCmd       string
	}{
		{
			"0.8.9",
			"opentofu",
			"get",
		},
		{
			"0.9.0",
			"opentofu",
			"init",
		},
		{
			"0.9.1",
			"opentofu",
			"init",
		},
		{
			"0.10.0",
			"opentofu",
			"init",
		},
		{
			"1.5.0",
			"opentofu",
			"init",
		},
		{
			"1.6.0",
			"opentofu",
			"init",
		},
	}

	for _, c := range cases {
		t.Run(c.version, func(t *testing.T) {
			terraform := tfclientmocks.NewMockClient(ctrl)
			projDistribution := c.distribution
			logger := logging.NewNoopLogger(t)
			ctx := command.ProjectContext{
				Workspace:             "workspace",
				RepoRelDir:            ".",
				Log:                   logger,
				TerraformDistribution: &projDistribution,
			}

			tfVersion, _ := version.NewVersion(c.version)
			iso := runtime.InitStepRunner{
				TerraformExecutor:     terraform,
				DefaultTFDistribution: tfDistribution,
				DefaultTFVersion:      tfVersion,
			}
			
			// If using init then we specify -input=false but not for get.
			expArgs := []string{c.expCmd, "-input=false", "-upgrade", "extra", "args"}
			if c.expCmd == "get" {
				expArgs = []string{c.expCmd, "-upgrade", "extra", "args"}
			}
			
			// Expect non-default distribution
			notDefaultDistribution := gomock.Not(gomock.Eq(tfDistribution))
			terraform.EXPECT().RunCommandWithVersion(
				ctx, "/path", expArgs, map[string]string(nil), notDefaultDistribution, tfVersion, "workspace").
				Return("output", nil)

			output, err := iso.Run(ctx, []string{"extra", "args"}, "/path", map[string]string(nil))
			Ok(t, err)
			// When there is no error, should not return init output to PR.
			Equals(t, "", output)
		})
	}
}

func TestRun_ShowInitOutputOnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	tfClient := tfclientmocks.NewMockClient(ctrl)
	expCmd := "output"
	expErr := "error"
	tfClient.EXPECT().RunCommandWithVersion(
		gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(expCmd, errors.New(expErr))
	
	mockDownloader := mocks.NewMockDownloader(ctrl)
	tfDistribution := tf.NewDistributionTerraformWithDownloader(mockDownloader)
	tfVersion, _ := version.NewVersion("0.11.0")
	logger := logging.NewNoopLogger(t)
	ctx := command.ProjectContext{
		Workspace:  "workspace",
		RepoRelDir: ".",
		Log:        logger,
	}
	iso := runtime.InitStepRunner{
		TerraformExecutor:     tfClient,
		DefaultTFDistribution: tfDistribution,
		DefaultTFVersion:      tfVersion,
	}

	output, err := iso.Run(ctx, []string{"extra", "args"}, "/path", map[string]string(nil))
	ErrContains(t, expErr, err)
	Equals(t, expCmd, output)
}

func TestRun_InitOmitsUpgradeFlagIfLockFilePresent(t *testing.T) {
	tmpDir := t.TempDir()
	lockFilePath := filepath.Join(tmpDir, ".terraform.lock.hcl")
	err := os.WriteFile(lockFilePath, []byte("# I'm a lock file"), 0600)
	Ok(t, err)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	logger := logging.NewNoopLogger(t)
	ctx := command.ProjectContext{
		Workspace:  "workspace",
		RepoRelDir: ".",
		Log:        logger,
	}
	repoDir := tmpDir

	terraform := tfclientmocks.NewMockClient(ctrl)
	mockDownloader := mocks.NewMockDownloader(ctrl)
	tfDistribution := tf.NewDistributionTerraformWithDownloader(mockDownloader)
	tfVersion, _ := version.NewVersion("0.14.0")

	iso := runtime.InitStepRunner{
		TerraformExecutor:     terraform,
		DefaultTFDistribution: tfDistribution,
		DefaultTFVersion:      tfVersion,
	}
	
	expectedArgs := []string{"init", "-input=false", "extra", "args"}
	terraform.EXPECT().RunCommandWithVersion(
		ctx, repoDir, expectedArgs, map[string]string(nil), tfDistribution, tfVersion, "workspace").
		Return("output", nil)

	output, err := iso.Run(ctx, []string{"extra", "args"}, repoDir, map[string]string(nil))
	Ok(t, err)
	// When there is no error, should not return init output to PR.
	Equals(t, "", output)
}

func TestRun_InitKeepsUpgradeFlagIfLockFilePresentAndUpgradeOptionSet(t *testing.T) {
	tmpDir := t.TempDir()
	lockFilePath := filepath.Join(tmpDir, ".terraform.lock.hcl")
	err := os.WriteFile(lockFilePath, []byte("# I'm a lock file"), 0600)
	Ok(t, err)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	logger := logging.NewNoopLogger(t)
	ctx := command.ProjectContext{
		Workspace:  "workspace",
		RepoRelDir: ".",
		Log:        logger,
	}

	terraform := tfclientmocks.NewMockClient(ctrl)
	mockDownloader := mocks.NewMockDownloader(ctrl)
	tfDistribution := tf.NewDistributionTerraformWithDownloader(mockDownloader)
	tfVersion, _ := version.NewVersion("0.14.0")

	iso := runtime.InitStepRunner{
		TerraformExecutor:     terraform,
		DefaultTFDistribution: tfDistribution,
		DefaultTFVersion:      tfVersion,
	}
	
	expectedArgs := []string{"init", "-input=false", "-upgrade", "extra", "args"}
	terraform.EXPECT().RunCommandWithVersion(
		ctx, tmpDir, expectedArgs, map[string]string(nil), tfDistribution, tfVersion, "workspace").
		Return("output", nil)

	output, err := iso.Run(ctx, []string{"extra", "args"}, tmpDir, map[string]string(nil))
	Ok(t, err)
	// When there is no error, should not return init output to PR.
	Equals(t, "", output)
}

func TestRun_InitKeepsUpgradeFlagIfLockFileNotPresent(t *testing.T) {
	tmpDir := t.TempDir()
	// Note: not creating a .terraform.lock.hcl file

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	logger := logging.NewNoopLogger(t)
	ctx := command.ProjectContext{
		Workspace:  "workspace",
		RepoRelDir: ".",
		Log:        logger,
	}

	terraform := tfclientmocks.NewMockClient(ctrl)
	mockDownloader := mocks.NewMockDownloader(ctrl)
	tfDistribution := tf.NewDistributionTerraformWithDownloader(mockDownloader)
	tfVersion, _ := version.NewVersion("0.14.0")

	iso := runtime.InitStepRunner{
		TerraformExecutor:     terraform,
		DefaultTFDistribution: tfDistribution,
		DefaultTFVersion:      tfVersion,
	}
	
	expectedArgs := []string{"init", "-input=false", "-upgrade", "extra", "args"}
	terraform.EXPECT().RunCommandWithVersion(
		ctx, tmpDir, expectedArgs, map[string]string(nil), tfDistribution, tfVersion, "workspace").
		Return("output", nil)

	output, err := iso.Run(ctx, []string{"extra", "args"}, tmpDir, map[string]string(nil))
	Ok(t, err)
	// When there is no error, should not return init output to PR.
	Equals(t, "", output)
}

// If the .terraform directory already exists, the -backend=false flag
// will be omitted since terraform will balk at it.
func TestRun_NoBackendInitFlagIfDotTerraformExists(t *testing.T) {
	cases := []struct {
		name         string
		expectedArgs []string
	}{
		{
			"standard",
			[]string{"init", "-input=false", "-upgrade"},
		},
		{
			"reconfigure",
			[]string{"init", "-input=false", "-reconfigure", "-upgrade"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			dirName := filepath.Join(tmpDir, ".terraform")
			err := os.MkdirAll(dirName, os.ModePerm)
			Ok(t, err)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			logger := logging.NewNoopLogger(t)
			ctx := command.ProjectContext{
				Workspace:  "workspace",
				RepoRelDir: ".",
				Log:        logger,
			}

			terraform := tfclientmocks.NewMockClient(ctrl)
			mockDownloader := mocks.NewMockDownloader(ctrl)
			tfDistribution := tf.NewDistributionTerraformWithDownloader(mockDownloader)
			tfVersion, _ := version.NewVersion("0.14.0")

			iso := runtime.InitStepRunner{
				TerraformExecutor:     terraform,
				DefaultTFDistribution: tfDistribution,
				DefaultTFVersion:      tfVersion,
			}
			
			terraform.EXPECT().RunCommandWithVersion(
				ctx, "/path", c.expectedArgs, map[string]string(nil), tfDistribution, tfVersion, "workspace").
				Return("output", nil)

			output, err := iso.Run(ctx, []string{}, "/path", map[string]string(nil))
			Ok(t, err)
			// When there is no error, should not return init output to PR.
			Equals(t, "", output)
		})
	}
}

// If -backend-config flag is provided and .terraform directory does not exist, -backend=false is omitted when used with any of these
// init subcommands: atlantis plan -- -init -backend-config=staging.backend.tfvars
// init subcommands: atlantis plan -- -init -backend-config staging.backend.tfvars
func TestRun_BackendConfigFlagParsing(t *testing.T) {
	tmpDir := t.TempDir()
	repoDir := tmpDir
	// No .terraform directory

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	logger := logging.NewNoopLogger(t)
	ctx := command.ProjectContext{
		Workspace:  "workspace",
		RepoRelDir: ".",
		Log:        logger,
	}

	terraform := tfclientmocks.NewMockClient(ctrl)
	mockDownloader := mocks.NewMockDownloader(ctrl)
	tfDistribution := tf.NewDistributionTerraformWithDownloader(mockDownloader)
	tfVersion, _ := version.NewVersion("0.14.0")

	iso := runtime.InitStepRunner{
		TerraformExecutor:     terraform,
		DefaultTFDistribution: tfDistribution,
		DefaultTFVersion:      tfVersion,
	}
	
	// Both parsing formats should work
	extraArgs := []string{"-backend-config=staging.backend.tfvars"}
	expectedArgs := []string{"init", "-input=false", "-upgrade", "-backend-config=staging.backend.tfvars"}
	
	terraform.EXPECT().RunCommandWithVersion(
		ctx, repoDir, expectedArgs, map[string]string(nil), tfDistribution, tfVersion, "workspace").
		Return("output", nil)

	output, err := iso.Run(ctx, extraArgs, repoDir, map[string]string(nil))
	Ok(t, err)
	// When there is no error, should not return init output to PR.
	Equals(t, "", output)
}

// GitIsNotInstalledErr is a special error that tells the init runner to run
// terraform init with -upgrade=false. We want to make sure that that
// flag is actually applied in this case.
type GitIsNotInstalledErr struct{}

func (e GitIsNotInstalledErr) Error() string {
	return exec.ErrNotFound.Error()
}

func (e GitIsNotInstalledErr) Unwrap() error {
	return exec.ErrNotFound
}

func TestRun_GetsUpgradeFalseIfGitNotInstalled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	cases := []struct {
		err     error
		version string
		expCmd  string
		expArgs []string
	}{
		{
			errors.New("not a git error"),
			"0.14.0",
			"init",
			[]string{"init", "-input=false", "-backend=false"},
		},
		{
			errors.New("not a git error"),
			"0.8.0",
			"get",
			[]string{"get", "-backend=false"},
		},
		{
			GitIsNotInstalledErr{},
			"0.14.0",
			"init",
			[]string{"init", "-input=false", "-upgrade=false", "-backend=false"},
		},
		{
			errors.Wrap(GitIsNotInstalledErr{}, "something else"),
			"0.14.0",
			"init",
			[]string{"init", "-input=false", "-upgrade=false", "-backend=false"},
		},
		{
			GitIsNotInstalledErr{},
			"0.8.0",
			"get",
			[]string{"get", "-upgrade=false", "-backend=false"},
		},
	}

	for _, c := range cases {
		descrip := strings.Split(c.err.Error(), "\n")[0]
		t.Run(descrip, func(t *testing.T) {
			// Setup the mocks.
			terraform := tfclientmocks.NewMockClient(ctrl)
			mockDownloader := mocks.NewMockDownloader(ctrl)
			tfDistribution := tf.NewDistributionTerraformWithDownloader(mockDownloader)
			tfVersion, _ := version.NewVersion(c.version)
			logger := logging.NewNoopLogger(t)
			ctx := command.ProjectContext{
				Workspace:  "workspace",
				RepoRelDir: ".",
				Log:        logger,
			}

			iso := runtime.InitStepRunner{
				TerraformExecutor:     terraform,
				DefaultTFDistribution: tfDistribution,
				DefaultTFVersion:      tfVersion,
			}

			// First, terraform workspace gets called and returns an error.
			terraform.EXPECT().RunCommandWithVersion(
				ctx, "/path", []string{"workspace", "list"}, map[string]string(nil), tfDistribution, tfVersion, "workspace").
				Return("", c.err)
			// Then init gets called with -upgrade=false if it's the Git error.
			terraform.EXPECT().RunCommandWithVersion(
				ctx, "/path", c.expArgs, map[string]string(nil), tfDistribution, tfVersion, "workspace").
				Return("output", nil)
			// Then since it succeeds, we call workspace list again. This is called
			// in to determine if this workspace needs to be created.
			terraform.EXPECT().RunCommandWithVersion(
				ctx, "/path", []string{"workspace", "list"}, map[string]string(nil), tfDistribution, tfVersion, "workspace").
				Return("output", nil)

			output, err := iso.Run(ctx, []string{}, "/path", map[string]string(nil))
			Ok(t, err)
			// When there is no error, should not return init output to PR.
			Equals(t, "", output)
		})
	}
}