// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package runtime_test

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashicorp/go-version"
	. "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/server/core/runtime"
	runtimemocks "github.com/runatlantis/atlantis/server/core/runtime/mocks"
	tf "github.com/runatlantis/atlantis/server/core/terraform"
	"github.com/runatlantis/atlantis/server/core/terraform/mocks"
	tfclientmocks "github.com/runatlantis/atlantis/server/core/terraform/tfclient/mocks"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"

	. "github.com/runatlantis/atlantis/testing"
)

// symlinkedPlanFixture lays out a project directory whose plan file path is a
// committed symlink pointing at a file outside the project directory.
type symlinkedPlanFixture struct {
	projectDir  string
	planFile    string
	outsideFile string
}

func newSymlinkedPlanFixture(t *testing.T) symlinkedPlanFixture {
	t.Helper()
	projectDir := t.TempDir()
	outsideFile := filepath.Join(t.TempDir(), "escaped")
	planFile := filepath.Join(projectDir, "default.tfplan")
	Ok(t, os.Symlink(outsideFile, planFile))
	return symlinkedPlanFixture{projectDir: projectDir, planFile: planFile, outsideFile: outsideFile}
}

func (f symlinkedPlanFixture) assertOutsideUntouched(t *testing.T) {
	t.Helper()
	_, err := os.Lstat(f.outsideFile)
	Assert(t, errors.Is(err, fs.ErrNotExist), "expected symlink target to stay absent, got %v", err)
}

func isSymlink(t *testing.T, path string) bool {
	t.Helper()
	info, err := os.Lstat(path)
	if errors.Is(err, fs.ErrNotExist) {
		return false
	}
	Ok(t, err)
	return info.Mode()&os.ModeSymlink != 0
}

func symlinkTestContext(t *testing.T) command.ProjectContext {
	t.Helper()
	return command.ProjectContext{
		Log:        logging.NewNoopLogger(t),
		Workspace:  "default",
		RepoRelDir: ".",
		User:       models.User{Username: "username"},
		Pull: models.PullRequest{
			Num: 2,
		},
		BaseRepo: models.Repo{
			FullName: "owner/repo",
			Owner:    "owner",
			Name:     "repo",
		},
	}
}

// Test that a symlink committed at the plan file path is removed before
// terraform plan -out runs, so the plan is never written through the link.
func TestRun_RemovesSymlinkedPlanFileBeforePlan(t *testing.T) {
	RegisterMockTestingT(t)
	terraform := tfclientmocks.NewMockClient()
	commitStatusUpdater := runtimemocks.NewMockStatusUpdater()
	asyncTfExec := runtimemocks.NewMockAsyncTFExec()
	mockDownloader := mocks.NewMockDownloader()
	tfDistribution := tf.NewDistributionTerraformWithDownloader(mockDownloader)
	tfVersion, _ := version.NewVersion("1.5.0")
	s := runtime.NewPlanStepRunner(terraform, tfDistribution, tfVersion, commitStatusUpdater, asyncTfExec, &runtime.LocalPlanStore{})
	fixture := newSymlinkedPlanFixture(t)
	ctx := symlinkTestContext(t)

	planCmdSawSymlink := true
	When(terraform.RunCommandWithVersion(Any[command.ProjectContext](), Any[string](), Any[[]string](),
		Any[map[string]string](), Any[tf.Distribution](), Any[*version.Version](), Any[string]())).
		Then(func(params []Param) ReturnValues {
			args := params[2].([]string)
			if args[0] == "workspace" {
				return ReturnValues{"default\n", nil}
			}
			Equals(t, []string{"plan", "-input=false", "-refresh", "-out", fixture.planFile}, args)
			planCmdSawSymlink = isSymlink(t, fixture.planFile)
			return ReturnValues{"output", nil}
		})

	output, err := s.Run(ctx, nil, fixture.projectDir, nil)
	Ok(t, err)
	Equals(t, "output", output)

	terraform.VerifyWasCalledOnce().RunCommandWithVersion(Any[command.ProjectContext](), Eq(fixture.projectDir),
		Eq([]string{"plan", "-input=false", "-refresh", "-out", fixture.planFile}), Any[map[string]string](),
		Any[tf.Distribution](), Any[*version.Version](), Eq("default"))
	Assert(t, !planCmdSawSymlink, "expected the symlink to be removed before terraform plan ran")
	Assert(t, !isSymlink(t, fixture.planFile), "expected the symlink to stay removed")
	fixture.assertOutsideUntouched(t)
}

// Test that the remote ops fallback does not write its fake plan file through
// a symlink at the plan file path, even one that appeared after the initial
// terraform plan attempt.
func TestRun_RemoteOpsDoesNotWritePlanThroughSymlink(t *testing.T) {
	RegisterMockTestingT(t)
	terraform := tfclientmocks.NewMockClient()
	commitStatusUpdater := runtimemocks.NewMockStatusUpdater()
	mockDownloader := mocks.NewMockDownloader()
	tfDistribution := tf.NewDistributionTerraformWithDownloader(mockDownloader)
	tfVersion, _ := version.NewVersion("1.1.0")
	asyncTf := &remotePlanMock{LinesToSend: remotePlanOutput}
	s := runtime.NewPlanStepRunner(terraform, tfDistribution, tfVersion, commitStatusUpdater, asyncTf, &runtime.LocalPlanStore{})
	fixture := newSymlinkedPlanFixture(t)
	ctx := symlinkTestContext(t)

	remoteOpsErr := `╷
│ Error: Saving a generated plan is currently not supported
│ 
│ Terraform Cloud does not support saving the generated execution plan
│ locally at this time.
╵
`
	When(terraform.RunCommandWithVersion(Any[command.ProjectContext](), Any[string](), Any[[]string](),
		Any[map[string]string](), Any[tf.Distribution](), Any[*version.Version](), Any[string]())).
		Then(func(params []Param) ReturnValues {
			args := params[2].([]string)
			if args[0] == "workspace" {
				return ReturnValues{"default\n", nil}
			}
			Assert(t, !isSymlink(t, fixture.planFile), "expected the symlink to be removed before terraform plan ran")
			// Simulate the link reappearing between the failed local plan and
			// the remote ops fallback so the fallback's own guard is exercised.
			Ok(t, os.Symlink(fixture.outsideFile, fixture.planFile))
			return ReturnValues{"\n" + remoteOpsErr, errors.New("exit status 1: err")}
		})

	output, err := s.Run(ctx, nil, fixture.projectDir, nil)
	Ok(t, err)
	Assert(t, strings.Contains(output, "Plan: 0 to add, 0 to change, 1 to destroy."), "expected remote plan output, got %q", output)

	Assert(t, !isSymlink(t, fixture.planFile), "expected the fake plan file to replace the symlink")
	bytes, err := os.ReadFile(fixture.planFile)
	Ok(t, err)
	Assert(t, strings.HasPrefix(string(bytes), "Atlantis: this plan was created by remote ops"), "expected remote plan header")
	fixture.assertOutsideUntouched(t)
}

// Test that an existing regular plan file is left for terraform to overwrite.
func TestRun_KeepsRegularPlanFile(t *testing.T) {
	RegisterMockTestingT(t)
	terraform := tfclientmocks.NewMockClient()
	commitStatusUpdater := runtimemocks.NewMockStatusUpdater()
	asyncTfExec := runtimemocks.NewMockAsyncTFExec()
	mockDownloader := mocks.NewMockDownloader()
	tfDistribution := tf.NewDistributionTerraformWithDownloader(mockDownloader)
	tfVersion, _ := version.NewVersion("1.5.0")
	s := runtime.NewPlanStepRunner(terraform, tfDistribution, tfVersion, commitStatusUpdater, asyncTfExec, &runtime.LocalPlanStore{})
	projectDir := t.TempDir()
	planFile := filepath.Join(projectDir, "default.tfplan")
	Ok(t, os.WriteFile(planFile, []byte("stale"), 0o600))
	ctx := symlinkTestContext(t)

	When(terraform.RunCommandWithVersion(Any[command.ProjectContext](), Any[string](), Any[[]string](),
		Any[map[string]string](), Any[tf.Distribution](), Any[*version.Version](), Any[string]())).
		Then(func(params []Param) ReturnValues {
			args := params[2].([]string)
			if args[0] == "workspace" {
				return ReturnValues{"default\n", nil}
			}
			bytes, err := os.ReadFile(planFile)
			Ok(t, err)
			Equals(t, "stale", string(bytes))
			return ReturnValues{"output", nil}
		})

	_, err := s.Run(ctx, nil, projectDir, nil)
	Ok(t, err)
	bytes, err := os.ReadFile(planFile)
	Ok(t, err)
	Equals(t, "stale", string(bytes))
}
