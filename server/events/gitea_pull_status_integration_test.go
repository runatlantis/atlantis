// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events_test

import (
	"os"
	"path/filepath"
	"testing"

	giteasdk "code.gitea.io/sdk/gitea"
	. "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/server/core/boltdb"
	"github.com/runatlantis/atlantis/server/core/config"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/core/runtime"
	tfclientmocks "github.com/runatlantis/atlantis/server/core/terraform/tfclient/mocks"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/mocks"
	"github.com/runatlantis/atlantis/server/events/models"
	giteamodels "github.com/runatlantis/atlantis/server/events/vcs/gitea"
	vcsmocks "github.com/runatlantis/atlantis/server/events/vcs/mocks"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/metrics/metricstest"
	. "github.com/runatlantis/atlantis/testing"
)

func TestGiteaAutoplanPullStatusRoundTripBuildsImmediateApply(t *testing.T) {
	RegisterMockTestingT(t)
	eventParser := events.EventParser{
		GiteaUser:  "atlantis",
		GiteaToken: "token",
	}
	repo := giteasdk.Repository{
		FullName: "RE-Germany/test",
		CloneURL: "https://gitea.example.com/RE-Germany/test.git",
	}
	giteaPull := giteasdk.PullRequest{
		Index:   6,
		HTMLURL: "https://gitea.example.com/RE-Germany/test/pulls/6",
		Body:    "update infrastructure",
		State:   giteasdk.StateOpen,
		Poster:  &giteasdk.User{UserName: "infra-author"},
		Head: &giteasdk.PRBranchInfo{
			Ref:        "update-infrastructure",
			Sha:        "d2eae324ca26242abca45d7b49d582cddb2a4f15",
			Repository: &repo,
		},
		Base: &giteasdk.PRBranchInfo{
			Ref:        "main",
			Repository: &repo,
		},
	}

	webhookPull, eventType, webhookBaseRepo, _, _, err := eventParser.ParseGiteaPullRequestEvent(giteaPull)
	Ok(t, err)
	Equals(t, models.OpenedPullEvent, eventType)

	database, err := boltdb.New(t.TempDir())
	Ok(t, err)
	t.Cleanup(func() {
		Ok(t, database.Close())
	})
	_, err = database.UpdatePullWithResults(webhookPull, []command.ProjectResult{{
		Command:     command.Plan,
		Workspace:   events.DefaultWorkspace,
		RepoRelDir:  "infrastructure",
		ProjectName: "gitea-autoplan",
		ProjectCommandOutput: command.ProjectCommandOutput{
			PlanSuccess: &models.PlanSuccess{},
		},
	}})
	Ok(t, err)

	commentPayload := giteamodels.GiteaIssueCommentPayload{
		Repository: repo,
		Issue:      giteasdk.Issue{Index: giteaPull.Index},
		Comment: giteasdk.Comment{
			Body:   "atlantis apply -p gitea-autoplan",
			Poster: &giteasdk.User{UserName: "infra-reviewer"},
		},
	}
	commentBaseRepo, _, commentPullNum, err := eventParser.ParseGiteaIssueCommentEvent(commentPayload)
	Ok(t, err)
	apiPull, apiBaseRepo, _, err := eventParser.ParseGiteaPull(&giteaPull)
	Ok(t, err)

	Equals(t, webhookBaseRepo.ID(), commentBaseRepo.ID())
	Equals(t, webhookBaseRepo.ID(), apiBaseRepo.ID())
	Equals(t, webhookPull.Num, commentPullNum)
	Equals(t, webhookPull.Num, apiPull.Num)
	Equals(t, webhookPull.HeadCommit, apiPull.HeadCommit)
	Equals(t, webhookPull.HeadBranch, apiPull.HeadBranch)
	Equals(t, webhookPull.BaseBranch, apiPull.BaseBranch)

	pullStatus, err := database.GetPullStatus(apiPull)
	Ok(t, err)
	Assert(t, pullStatus != nil, "expected apply-side Gitea identity to retrieve autoplan PullStatus")
	Equals(t, apiPull.HeadCommit, pullStatus.Pull.HeadCommit)
	Equals(t, apiPull.BaseBranch, pullStatus.Pull.BaseBranch)
	Equals(t, 1, len(pullStatus.Projects))
	Equals(t, "infrastructure", pullStatus.Projects[0].RepoRelDir)
	Equals(t, events.DefaultWorkspace, pullStatus.Projects[0].Workspace)
	Equals(t, "gitea-autoplan", pullStatus.Projects[0].ProjectName)
	Equals(t, models.PlannedPlanStatus, pullStatus.Projects[0].Status)

	repoDir := DirStructure(t, map[string]any{
		"atlantis.yaml": `version: 3
projects:
  - name: gitea-autoplan
    dir: infrastructure
    workspace: default
`,
		"infrastructure": map[string]any{
			"main.tf": `terraform {
  required_version = ">= 1.0"
}
`,
		},
	})
	planPath := filepath.Join(repoDir, "infrastructure", runtime.GetPlanFilename(events.DefaultWorkspace, "gitea-autoplan"))
	Ok(t, os.WriteFile(planPath, []byte("gitea autoplan"), 0600))

	workingDir := mocks.NewMockWorkingDir()
	When(workingDir.GetWorkingDir(Any[models.Repo](), Any[models.PullRequest](), Any[string]())).ThenReturn(repoDir, nil)
	logger := logging.NewNoopLogger(t)
	scope := metricstest.NewLoggingScope(t, logger, "atlantis")
	builder := events.NewProjectCommandBuilder(
		false,
		&config.ParserValidator{},
		&events.DefaultProjectFinder{},
		vcsmocks.NewMockClient(),
		workingDir,
		events.NewDefaultWorkingDirLocker(),
		valid.NewGlobalCfgFromArgs(valid.GlobalCfgArgs{AllowAllRepoSettings: true}),
		&events.DefaultPendingPlanFinder{},
		&events.CommentParser{ExecutableName: "atlantis"},
		false,
		false,
		false,
		false,
		false,
		"",
		"**/*.tf,**/*.tfvars,**/*.tfvars.json,**/terragrunt.hcl,**/.terraform.lock.hcl",
		false,
		"",
		false,
		false,
		"auto",
		scope,
		tfclientmocks.NewMockClient(),
		nil,
	)
	ctx := &command.Context{
		Log:        logger,
		Scope:      scope,
		Pull:       apiPull,
		PullStatus: pullStatus,
	}
	applyCommands, err := builder.BuildApplyCommands(ctx, &events.CommentCommand{
		Name:        command.Apply,
		ProjectName: "gitea-autoplan",
	})
	Ok(t, err)
	Equals(t, 1, len(applyCommands))
	Equals(t, apiPull.HeadCommit, applyCommands[0].Pull.HeadCommit)
	Assert(t, applyCommands[0].ExpectedPlanHash != "", "expected built-in apply command to capture the autoplan hash")

	validator := &events.DefaultApplyPlanValidator{PullStatusFetcher: database}
	err = validator.ValidateProjectPlan(applyCommands[0], filepath.Join(repoDir, "infrastructure"))
	Ok(t, err)
}
