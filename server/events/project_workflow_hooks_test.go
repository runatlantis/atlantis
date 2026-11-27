// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events_test

import (
	"testing"

	. "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/models/testdata"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

func TestWorkflowHook_RunsPerProject(t *testing.T) {
	Equals(t, false, valid.WorkflowHook{}.RunsPerProject())
	Equals(t, false, valid.WorkflowHook{Scope: valid.WorkflowHookScopeRepo}.RunsPerProject())
	Equals(t, true, valid.WorkflowHook{Scope: valid.WorkflowHookScopeProject}.RunsPerProject())
}

func TestRunPreHooks_SkipsProjectScopedHooks(t *testing.T) {
	preWorkflowHooksSetup(t)
	log := logging.NewNoopLogger(t)
	pull := testdata.Pull
	pull.BaseRepo = testdata.GithubRepo
	ctx := &command.Context{Pull: pull, HeadRepo: testdata.GithubRepo, User: testdata.User, Log: log}

	repoHook := &valid.WorkflowHook{RunCommand: "repo-hook"}
	projectHook := &valid.WorkflowHook{RunCommand: "project-hook", Scope: valid.WorkflowHookScopeProject}
	preWh.GlobalCfg = valid.GlobalCfg{Repos: []valid.Repo{{
		ID:               testdata.GithubRepo.ID(),
		PreWorkflowHooks: []*valid.WorkflowHook{repoHook, projectHook},
	}}}

	repoDir := "path/to/repo"
	When(preWhWorkingDirLocker.TryLock(testdata.GithubRepo.FullName, pull.Num, events.DefaultWorkspace, events.DefaultRepoRelDir, "", command.Plan, events.WorkingDirLockMetadataForPull(pull))).ThenReturn(func() {}, nil)
	When(preWhWorkingDir.Clone(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(pull), Eq(events.DefaultWorkspace))).ThenReturn(repoDir, nil)
	When(whPreWorkflowHookRunner.Run(Any[models.WorkflowHookCommandContext](), Any[string](), Any[string](), Any[string](), Eq(repoDir))).ThenReturn("", "", nil)

	Ok(t, preWh.RunPreHooks(ctx, &events.CommentCommand{Name: command.Plan}))

	whPreWorkflowHookRunner.VerifyWasCalledOnce().Run(Any[models.WorkflowHookCommandContext](), Eq("repo-hook"), Any[string](), Any[string](), Eq(repoDir))
	whPreWorkflowHookRunner.VerifyWasCalled(Never()).Run(Any[models.WorkflowHookCommandContext](), Eq("project-hook"), Any[string](), Any[string](), Eq(repoDir))
}

func TestRunPreHooksForProject_RunsOnlyProjectScopedHooks(t *testing.T) {
	preWorkflowHooksSetup(t)
	log := logging.NewNoopLogger(t)
	pull := testdata.Pull
	pull.BaseRepo = testdata.GithubRepo

	repoHook := &valid.WorkflowHook{RunCommand: "repo-hook"}
	projectHook := &valid.WorkflowHook{RunCommand: "project-hook", Scope: valid.WorkflowHookScopeProject}
	preWh.GlobalCfg = valid.GlobalCfg{Repos: []valid.Repo{{
		ID:               testdata.GithubRepo.ID(),
		PreWorkflowHooks: []*valid.WorkflowHook{repoHook, projectHook},
	}}}

	pctx := command.ProjectContext{
		Log:         log,
		Pull:        pull,
		HeadRepo:    testdata.GithubRepo,
		BaseRepo:    testdata.GithubRepo,
		User:        testdata.User,
		RepoRelDir:  "dir1",
		Workspace:   "default",
		ProjectName: "proj1",
		CommandName: command.Apply,
	}

	repoDir := "path/to/repo"
	When(preWhWorkingDirLocker.TryLock(testdata.GithubRepo.FullName, pull.Num, events.DefaultWorkspace, events.DefaultRepoRelDir, "proj1", command.Apply, events.WorkingDirLockMetadataForPull(pull))).ThenReturn(func() {}, nil)
	When(preWhWorkingDir.Clone(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(pull), Eq(events.DefaultWorkspace))).ThenReturn(repoDir, nil)
	When(whPreWorkflowHookRunner.Run(Any[models.WorkflowHookCommandContext](), Any[string](), Any[string](), Any[string](), Eq(repoDir))).ThenReturn("", "", nil)

	Ok(t, preWh.RunPreHooksForProject(pctx))

	whPreWorkflowHookRunner.VerifyWasCalledOnce().Run(Any[models.WorkflowHookCommandContext](), Eq("project-hook"), Any[string](), Any[string](), Eq(repoDir))
	whPreWorkflowHookRunner.VerifyWasCalled(Never()).Run(Any[models.WorkflowHookCommandContext](), Eq("repo-hook"), Any[string](), Any[string](), Eq(repoDir))
}

func TestRunPostHooksForProject_RunsOnlyProjectScopedHooks(t *testing.T) {
	postWorkflowHooksSetup(t)
	log := logging.NewNoopLogger(t)
	pull := testdata.Pull
	pull.BaseRepo = testdata.GithubRepo

	repoHook := &valid.WorkflowHook{RunCommand: "repo-hook"}
	projectHook := &valid.WorkflowHook{RunCommand: "project-hook", Scope: valid.WorkflowHookScopeProject}
	postWh.GlobalCfg = valid.GlobalCfg{Repos: []valid.Repo{{
		ID:                testdata.GithubRepo.ID(),
		PostWorkflowHooks: []*valid.WorkflowHook{repoHook, projectHook},
	}}}

	pctx := command.ProjectContext{
		Log:         log,
		Pull:        pull,
		HeadRepo:    testdata.GithubRepo,
		BaseRepo:    testdata.GithubRepo,
		User:        testdata.User,
		RepoRelDir:  "dir1",
		Workspace:   "default",
		ProjectName: "proj1",
		CommandName: command.Apply,
	}

	repoDir := "path/to/repo"
	When(postWhWorkingDirLocker.TryLock(testdata.GithubRepo.FullName, pull.Num, events.DefaultWorkspace, events.DefaultRepoRelDir, "proj1", command.Apply, events.WorkingDirLockMetadataForPull(pull))).ThenReturn(func() {}, nil)
	When(postWhWorkingDir.Clone(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(pull), Eq(events.DefaultWorkspace))).ThenReturn(repoDir, nil)
	When(whPostWorkflowHookRunner.Run(Any[models.WorkflowHookCommandContext](), Any[string](), Any[string](), Any[string](), Eq(repoDir))).ThenReturn("", "", nil)

	Ok(t, postWh.RunPostHooksForProject(pctx, false))

	whPostWorkflowHookRunner.VerifyWasCalledOnce().Run(Any[models.WorkflowHookCommandContext](), Eq("project-hook"), Any[string](), Any[string](), Eq(repoDir))
	whPostWorkflowHookRunner.VerifyWasCalled(Never()).Run(Any[models.WorkflowHookCommandContext](), Eq("repo-hook"), Any[string](), Any[string](), Eq(repoDir))
}
