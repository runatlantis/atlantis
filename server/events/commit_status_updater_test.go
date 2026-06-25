// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.

package events_test

import (
	"fmt"
	"strings"
	"testing"

	. "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs/mocks"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

const (
	githubStatusContextLimit         = 255
	truncatedContextHashSuffixLength = 13
)

func assertTruncatedStatusContext(t *testing.T, src string, original string) {
	t.Helper()
	if len(original) <= githubStatusContextLimit {
		t.Fatalf("expected original context %q to exceed %d characters", original, githubStatusContextLimit)
	}
	Equals(t, githubStatusContextLimit, len(src))

	prefixLength := githubStatusContextLimit - truncatedContextHashSuffixLength
	Equals(t, original[:prefixLength], src[:prefixLength])
	if src[prefixLength] != '-' {
		t.Fatalf("expected truncated context %q to include a hash suffix", src)
	}
	for _, char := range src[prefixLength+1:] {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f')) {
			t.Fatalf("expected truncated context suffix %q to be lowercase hex", src[prefixLength:])
		}
	}
}

func TestUpdateCombined(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	cases := []struct {
		status     models.CommitStatus
		command    command.Name
		expDescrip string
	}{
		{
			status:     models.PendingCommitStatus,
			command:    command.Plan,
			expDescrip: "Plan in progress...",
		},
		{
			status:     models.FailedCommitStatus,
			command:    command.Plan,
			expDescrip: "Plan failed.",
		},
		{
			status:     models.SuccessCommitStatus,
			command:    command.Plan,
			expDescrip: "Plan succeeded.",
		},
		{
			status:     models.PendingCommitStatus,
			command:    command.Apply,
			expDescrip: "Apply in progress...",
		},
		{
			status:     models.FailedCommitStatus,
			command:    command.Apply,
			expDescrip: "Apply failed.",
		},
		{
			status:     models.SuccessCommitStatus,
			command:    command.Apply,
			expDescrip: "Apply succeeded.",
		},
	}

	for _, c := range cases {
		t.Run(c.expDescrip, func(t *testing.T) {
			RegisterMockTestingT(t)
			client := mocks.NewMockClient()
			s := events.DefaultCommitStatusUpdater{Client: client, StatusName: "atlantis"}
			err := s.UpdateCombined(logger, models.Repo{}, models.PullRequest{}, c.status, c.command)
			Ok(t, err)

			expSrc := fmt.Sprintf("atlantis/%s", c.command)
			client.VerifyWasCalledOnce().UpdateStatus(logger, models.Repo{}, models.PullRequest{}, c.status, expSrc, c.expDescrip, "")
		})
	}
}

func TestUpdateCombinedCount(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	cases := []struct {
		status     models.CommitStatus
		command    command.Name
		numSuccess int
		numTotal   int
		expDescrip string
	}{
		{
			status:     models.PendingCommitStatus,
			command:    command.Plan,
			numSuccess: 0,
			numTotal:   2,
			expDescrip: "0/2 projects planned successfully.",
		},
		{
			status:     models.FailedCommitStatus,
			command:    command.Plan,
			numSuccess: 1,
			numTotal:   2,
			expDescrip: "1/2 projects planned successfully.",
		},
		{
			status:     models.SuccessCommitStatus,
			command:    command.Plan,
			numSuccess: 2,
			numTotal:   2,
			expDescrip: "2/2 projects planned successfully.",
		},
		{
			status:     models.FailedCommitStatus,
			command:    command.Apply,
			numSuccess: 0,
			numTotal:   2,
			expDescrip: "0/2 projects applied successfully.",
		},
		{
			status:     models.PendingCommitStatus,
			command:    command.Apply,
			numSuccess: 1,
			numTotal:   2,
			expDescrip: "1/2 projects applied successfully.",
		},
		{
			status:     models.SuccessCommitStatus,
			command:    command.Apply,
			numSuccess: 2,
			numTotal:   2,
			expDescrip: "2/2 projects applied successfully.",
		},
	}

	for _, c := range cases {
		t.Run(c.expDescrip, func(t *testing.T) {
			RegisterMockTestingT(t)
			client := mocks.NewMockClient()
			s := events.DefaultCommitStatusUpdater{Client: client, StatusName: "atlantis-test"}
			err := s.UpdateCombinedCount(logger, models.Repo{}, models.PullRequest{}, c.status, c.command, c.numSuccess, c.numTotal)
			Ok(t, err)

			expSrc := fmt.Sprintf("%s/%s", s.StatusName, c.command)
			client.VerifyWasCalledOnce().UpdateStatus(logger, models.Repo{}, models.PullRequest{}, c.status, expSrc, c.expDescrip, "")
		})
	}
}

// Test that it sets the "source" properly depending on if the project is
// named or not.
func TestDefaultCommitStatusUpdater_UpdateProjectSrc(t *testing.T) {
	RegisterMockTestingT(t)
	cases := []struct {
		projectName string
		repoRelDir  string
		workspace   string
		expSrc      string
	}{
		{
			projectName: "name",
			repoRelDir:  ".",
			workspace:   "default",
			expSrc:      "atlantis/plan: name",
		},
		{
			projectName: "",
			repoRelDir:  "dir1/dir2",
			workspace:   "workspace",
			expSrc:      "atlantis/plan: dir1/dir2/workspace",
		},
	}

	for _, c := range cases {
		t.Run(c.expSrc, func(t *testing.T) {
			client := mocks.NewMockClient()
			s := events.DefaultCommitStatusUpdater{Client: client, StatusName: "atlantis"}
			err := s.UpdateProject(command.ProjectContext{
				ProjectName: c.projectName,
				RepoRelDir:  c.repoRelDir,
				Workspace:   c.workspace,
			}, command.Plan, models.PendingCommitStatus, "url", nil)
			Ok(t, err)
			client.VerifyWasCalledOnce().UpdateStatus(
				Any[logging.SimpleLogging](), Eq(models.Repo{}), Eq(models.PullRequest{}), Eq(models.PendingCommitStatus), Eq(c.expSrc),
				Eq("Plan in progress..."), Eq("url"))
		})
	}
}

// Test that it uses the right words in the description.
func TestDefaultCommitStatusUpdater_UpdateProject(t *testing.T) {
	RegisterMockTestingT(t)
	cases := []struct {
		status     models.CommitStatus
		cmd        command.Name
		result     *command.ProjectCommandOutput
		expDescrip string
	}{
		{
			status:     models.PendingCommitStatus,
			cmd:        command.Plan,
			expDescrip: "Plan in progress...",
		},
		{
			status:     models.FailedCommitStatus,
			cmd:        command.Plan,
			expDescrip: "Plan failed.",
		},
		{
			status: models.SuccessCommitStatus,
			cmd:    command.Plan,
			result: &command.ProjectCommandOutput{
				PlanSuccess: &models.PlanSuccess{
					TerraformOutput: "aaa\nNote: Objects have changed outside of Terraform\nbbb\nPlan: 1 to add, 2 to change, 3 to destroy.\nbbb",
				},
			},
			expDescrip: "Plan: 1 to add, 2 to change, 3 to destroy.",
		},
		{
			status:     models.PendingCommitStatus,
			cmd:        command.Apply,
			expDescrip: "Apply in progress...",
		},
		{
			status:     models.FailedCommitStatus,
			cmd:        command.Apply,
			expDescrip: "Apply failed.",
		},
		{
			status: models.SuccessCommitStatus,
			cmd:    command.Apply,
			result: &command.ProjectCommandOutput{
				ApplySuccess: "success",
			},
			expDescrip: "Apply succeeded.",
		},
	}

	for _, c := range cases {
		t.Run(c.expDescrip, func(t *testing.T) {
			client := mocks.NewMockClient()
			s := events.DefaultCommitStatusUpdater{Client: client, StatusName: "atlantis"}
			err := s.UpdateProject(command.ProjectContext{
				RepoRelDir: ".",
				Workspace:  "default",
			}, c.cmd, c.status, "url", c.result)
			Ok(t, err)
			client.VerifyWasCalledOnce().UpdateStatus(Any[logging.SimpleLogging](), Eq(models.Repo{}), Eq(models.PullRequest{}), Eq(c.status),
				Eq(fmt.Sprintf("atlantis/%s: ./default", c.cmd.String())), Eq(c.expDescrip), Eq("url"))
		})
	}
}

// Test that the status context is truncated to 255 characters when the project
// path is very long, to avoid a 422 from the GitHub Statuses API.
func TestDefaultCommitStatusUpdater_UpdateProjectTruncatesLongContext(t *testing.T) {
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger(t)
	client := mocks.NewMockClient()
	s := events.DefaultCommitStatusUpdater{Client: client, StatusName: "atlantis"}
	// Build a directory name long enough to push the context over 255 chars.
	longDir := fmt.Sprintf("%s/deeply/nested/path", fmt.Sprintf("%0250d", 0))
	originalSrc := fmt.Sprintf("atlantis/plan: %s/default", longDir)
	err := s.UpdateProject(command.ProjectContext{
		Log:        logger,
		RepoRelDir: longDir,
		Workspace:  "default",
	}, command.Plan, models.PendingCommitStatus, "url", nil)
	Ok(t, err)
	_, _, _, _, src, _, _ := client.VerifyWasCalledOnce().UpdateStatus(
		Any[logging.SimpleLogging](), Eq(models.Repo{}), Eq(models.PullRequest{}),
		Eq(models.PendingCommitStatus), Any[string](), Eq("Plan in progress..."), Eq("url")).GetCapturedArguments()
	assertTruncatedStatusContext(t, src, originalSrc)
}

func TestDefaultCommitStatusUpdater_UpdateProjectTruncatedContextsRemainUnique(t *testing.T) {
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger(t)
	client := mocks.NewMockClient()
	s := events.DefaultCommitStatusUpdater{Client: client, StatusName: "atlantis"}
	commonDirPrefix := strings.Repeat("a", 260)
	repoRelDirs := []string{
		commonDirPrefix + "/project-a",
		commonDirPrefix + "/project-b",
	}
	originalSrcs := make([]string, 0, len(repoRelDirs))

	for _, repoRelDir := range repoRelDirs {
		originalSrcs = append(originalSrcs, fmt.Sprintf("atlantis/plan: %s/default", repoRelDir))
		err := s.UpdateProject(command.ProjectContext{
			Log:        logger,
			RepoRelDir: repoRelDir,
			Workspace:  "default",
		}, command.Plan, models.PendingCommitStatus, "url", nil)
		Ok(t, err)
	}

	_, _, _, _, srcs, _, _ := client.VerifyWasCalled(Times(2)).UpdateStatus(
		Any[logging.SimpleLogging](), Eq(models.Repo{}), Eq(models.PullRequest{}),
		Eq(models.PendingCommitStatus), Any[string](), Eq("Plan in progress..."), Eq("url")).GetAllCapturedArguments()
	Equals(t, originalSrcs[0][:githubStatusContextLimit], originalSrcs[1][:githubStatusContextLimit])
	Assert(t, srcs[0] != srcs[1], "expected truncated contexts to remain unique")
	for i, src := range srcs {
		assertTruncatedStatusContext(t, src, originalSrcs[i])
	}
}

func TestDefaultCommitStatusUpdater_UpdateWorkflowHookTruncatedContextsRemainUnique(t *testing.T) {
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger(t)
	client := mocks.NewMockClient()
	s := events.DefaultCommitStatusUpdater{Client: client, StatusName: "atlantis"}
	commonHookDescription := strings.Repeat("a", 260)
	hookDescriptions := []string{
		commonHookDescription + " first hook",
		commonHookDescription + " second hook",
	}
	originalSrcs := make([]string, 0, len(hookDescriptions))

	for _, hookDescription := range hookDescriptions {
		originalSrcs = append(originalSrcs, fmt.Sprintf("atlantis/pre_workflow_hook: %s", hookDescription))
		err := s.UpdatePreWorkflowHook(logger, models.PullRequest{}, models.PendingCommitStatus, hookDescription, "", "url")
		Ok(t, err)
	}

	_, _, _, _, srcs, _, _ := client.VerifyWasCalled(Times(2)).UpdateStatus(
		Any[logging.SimpleLogging](), Eq(models.Repo{}), Eq(models.PullRequest{}),
		Eq(models.PendingCommitStatus), Any[string](), Eq("in progress..."), Eq("url")).GetAllCapturedArguments()
	Equals(t, originalSrcs[0][:githubStatusContextLimit], originalSrcs[1][:githubStatusContextLimit])
	Assert(t, srcs[0] != srcs[1], "expected truncated workflow hook contexts to remain unique")
	for i, src := range srcs {
		assertTruncatedStatusContext(t, src, originalSrcs[i])
	}
}

// Test that we can set the status name.
func TestDefaultCommitStatusUpdater_UpdateProjectCustomStatusName(t *testing.T) {
	RegisterMockTestingT(t)
	client := mocks.NewMockClient()
	s := events.DefaultCommitStatusUpdater{Client: client, StatusName: "custom"}
	err := s.UpdateProject(command.ProjectContext{
		RepoRelDir: ".",
		Workspace:  "default",
	}, command.Apply, models.SuccessCommitStatus, "url", nil)
	Ok(t, err)
	client.VerifyWasCalledOnce().UpdateStatus(Any[logging.SimpleLogging](), Eq(models.Repo{}), Eq(models.PullRequest{}),
		Eq(models.SuccessCommitStatus), Eq("custom/apply: ./default"), Eq("Apply succeeded."), Eq("url"))
}
