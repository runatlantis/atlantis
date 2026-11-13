// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.

package events_test

import (
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"

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
	if utf8.RuneCountInString(original) <= githubStatusContextLimit {
		t.Fatalf("expected original context %q to exceed %d characters", original, githubStatusContextLimit)
	}
	Assert(t, utf8.ValidString(src), "expected truncated context to be valid UTF-8")
	Equals(t, githubStatusContextLimit, utf8.RuneCountInString(src))

	prefixLength := githubStatusContextLimit - truncatedContextHashSuffixLength
	originalRunes := []rune(original)
	srcRunes := []rune(src)
	Equals(t, string(originalRunes[:prefixLength]), string(srcRunes[:prefixLength]))
	if srcRunes[prefixLength] != '-' {
		t.Fatalf("expected truncated context %q to include a hash suffix", src)
	}
	for _, char := range srcRunes[prefixLength+1:] {
		if (char < '0' || char > '9') && (char < 'a' || char > 'f') {
			t.Fatalf("expected truncated context suffix %q to be lowercase hex", string(srcRunes[prefixLength:]))
		}
	}
}

func leadingRunes(s string, count int) string {
	return string([]rune(s)[:count])
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

func TestUpdateCombinedTruncatesLongContext(t *testing.T) {
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger(t)
	client := mocks.NewMockClient()
	s := events.DefaultCommitStatusUpdater{Client: client, StatusName: strings.Repeat("a", 260)}
	originalSrc := fmt.Sprintf("%s/%s", s.StatusName, command.Plan)

	err := s.UpdateCombined(logger, models.Repo{}, models.PullRequest{}, models.PendingCommitStatus, command.Plan)
	Ok(t, err)

	_, _, _, _, src, _, _ := client.VerifyWasCalledOnce().UpdateStatus(
		Any[logging.SimpleLogging](), Eq(models.Repo{}), Eq(models.PullRequest{}),
		Eq(models.PendingCommitStatus), Any[string](), Eq("Plan in progress..."), Eq("")).GetCapturedArguments()
	assertTruncatedStatusContext(t, src, originalSrc)
}

func TestUpdateCombinedCount(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	cases := []struct {
		status     models.CommitStatus
		command    command.Name
		numSuccess int
		numTotal   int
		numErrored int
		expDescrip string
	}{
		{
			status:     models.PendingCommitStatus,
			command:    command.Plan,
			numSuccess: 0,
			numTotal:   2,
			expDescrip: "0/2 projects planned.",
		},
		{
			status:     models.FailedCommitStatus,
			command:    command.Plan,
			numSuccess: 1,
			numTotal:   2,
			expDescrip: "1/2 projects failed to plan.",
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
			command:    command.PolicyCheck,
			numSuccess: 0,
			numTotal:   2,
			numErrored: 1,
			expDescrip: "1/2 projects failed policy checks.",
		},
		{
			status:     models.FailedCommitStatus,
			command:    command.Apply,
			numSuccess: 0,
			numTotal:   2,
			numErrored: 2,
			expDescrip: "2/2 projects failed to apply.",
		},
		{
			status:     models.FailedCommitStatus,
			command:    command.Apply,
			numSuccess: 1,
			numTotal:   3,
			numErrored: 1,
			expDescrip: "1/3 projects failed to apply.",
		},
		{
			status:     models.PendingCommitStatus,
			command:    command.Apply,
			numSuccess: 1,
			numTotal:   2,
			expDescrip: "1/2 projects applied.",
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
			err := s.UpdateCombinedCount(logger, models.Repo{}, models.PullRequest{}, c.status, c.command, models.ProjectCounts{Success: c.numSuccess, Total: c.numTotal, Errored: c.numErrored})
			Ok(t, err)

			expSrc := fmt.Sprintf("%s/%s", s.StatusName, c.command)
			client.VerifyWasCalledOnce().UpdateStatus(logger, models.Repo{}, models.PullRequest{}, c.status, expSrc, c.expDescrip, "")
		})
	}
}

func TestUpdateCombinedCountNoChanges(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	cases := []struct {
		name         string
		status       models.CommitStatus
		numSuccess   int
		numTotal     int
		numNoChanges int
		expDescrip   string
	}{
		{
			name:         "all up to date",
			status:       models.SuccessCommitStatus,
			numSuccess:   2,
			numTotal:     2,
			numNoChanges: 2,
			expDescrip:   "2/2 projects up to date.",
		},
		{
			name:         "mixed: applied and no-change",
			status:       models.SuccessCommitStatus,
			numSuccess:   3,
			numTotal:     3,
			numNoChanges: 1,
			expDescrip:   "2/3 projects applied successfully (1 up to date).",
		},
		{
			name:         "pending: no-change and unapplied",
			status:       models.PendingCommitStatus,
			numSuccess:   1,
			numTotal:     2,
			numNoChanges: 1,
			expDescrip:   "0/2 projects applied (1 up to date).",
		},
		{
			name:         "pending: applied, no-change, and unapplied",
			status:       models.PendingCommitStatus,
			numSuccess:   2,
			numTotal:     3,
			numNoChanges: 1,
			expDescrip:   "1/3 projects applied (1 up to date).",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			RegisterMockTestingT(t)
			client := mocks.NewMockClient()
			s := events.DefaultCommitStatusUpdater{Client: client, StatusName: "atlantis-test"}
			err := s.UpdateCombinedCount(logger, models.Repo{}, models.PullRequest{}, c.status, command.Apply, models.ProjectCounts{Success: c.numSuccess, Total: c.numTotal, NoChanges: c.numNoChanges})
			Ok(t, err)

			expSrc := fmt.Sprintf("%s/apply", s.StatusName)
			client.VerifyWasCalledOnce().UpdateStatus(logger, models.Repo{}, models.PullRequest{}, c.status, expSrc, c.expDescrip, "")
		})
	}
}

func TestUpdateCombinedCountTruncatesLongContext(t *testing.T) {
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger(t)
	client := mocks.NewMockClient()
	s := events.DefaultCommitStatusUpdater{Client: client, StatusName: strings.Repeat("a", 260)}
	originalSrc := fmt.Sprintf("%s/%s", s.StatusName, command.Plan)

	err := s.UpdateCombinedCount(logger, models.Repo{}, models.PullRequest{}, models.PendingCommitStatus, command.Plan, models.ProjectCounts{Total: 2})
	Ok(t, err)

	_, _, _, _, src, _, _ := client.VerifyWasCalledOnce().UpdateStatus(
		Any[logging.SimpleLogging](), Eq(models.Repo{}), Eq(models.PullRequest{}),
		Eq(models.PendingCommitStatus), Any[string](), Eq("0/2 projects planned."), Eq("")).GetCapturedArguments()
	assertTruncatedStatusContext(t, src, originalSrc)
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

func TestDefaultCommitStatusUpdater_UpdateProjectTruncatesLongUTF8Context(t *testing.T) {
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger(t)
	client := mocks.NewMockClient()
	s := events.DefaultCommitStatusUpdater{Client: client, StatusName: "atlantis"}
	longDir := strings.Repeat("工程", 130)
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
	Equals(t, leadingRunes(originalSrcs[0], githubStatusContextLimit), leadingRunes(originalSrcs[1], githubStatusContextLimit))
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
	Equals(t, leadingRunes(originalSrcs[0], githubStatusContextLimit), leadingRunes(originalSrcs[1], githubStatusContextLimit))
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
