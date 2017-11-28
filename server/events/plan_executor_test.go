package events_test

import (
	"errors"
	"testing"

	"github.com/hootsuite/atlantis/server/events"
	"github.com/hootsuite/atlantis/server/events/locking"
	lmocks "github.com/hootsuite/atlantis/server/events/locking/mocks"
	"github.com/hootsuite/atlantis/server/events/mocks"
	"github.com/hootsuite/atlantis/server/events/models"
	rmocks "github.com/hootsuite/atlantis/server/events/run/mocks"
	tmocks "github.com/hootsuite/atlantis/server/events/terraform/mocks"
	vcsmocks "github.com/hootsuite/atlantis/server/events/vcs/mocks"
	"github.com/hootsuite/atlantis/server/events/vcs/mocks/matchers"
	"github.com/hootsuite/atlantis/server/logging"
	. "github.com/hootsuite/atlantis/testing"
	. "github.com/petergtz/pegomock"
)

var planCtx = events.CommandContext{
	Command: &events.Command{
		Name:        events.Plan,
		Environment: "env",
	},
	Log:      logging.NewNoopLogger(),
	BaseRepo: models.Repo{},
	HeadRepo: models.Repo{},
	Pull:     models.PullRequest{},
	User: models.User{
		Username: "anubhavmishra",
	},
}

func TestExecute_ModifiedFilesErr(t *testing.T) {
	t.Log("If GetModifiedFiles returns an error we return an error")
	p, _, _ := setupPlanExecutorTest(t)
	When(p.VCSClient.GetModifiedFiles(matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest(), matchers.AnyVcsHost())).ThenReturn(nil, errors.New("err"))
	r := p.Execute(&planCtx)

	Assert(t, r.Error != nil, "exp .Error to be set")
	Equals(t, "getting modified files: err", r.Error.Error())
}

func TestExecute_NoModifiedProjects(t *testing.T) {
	t.Log("If there are no modified projects we return a failure")
	p, _, _ := setupPlanExecutorTest(t)
	// We don't need to actually mock VCSClient.GetModifiedFiles because by
	// default it will return an empty slice which is what we want for this test.
	r := p.Execute(&planCtx)

	Equals(t, "No Terraform files were modified.", r.Failure)
}

func TestExecute_CloneErr(t *testing.T) {
	t.Log("If Workspace.Clone returns an error we return an error")
	p, _, _ := setupPlanExecutorTest(t)
	When(p.VCSClient.GetModifiedFiles(matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest(), matchers.AnyVcsHost())).ThenReturn([]string{"file.tf"}, nil)
	When(p.Workspace.Clone(planCtx.Log, planCtx.BaseRepo, planCtx.HeadRepo, planCtx.Pull, "env")).ThenReturn("", errors.New("err"))
	r := p.Execute(&planCtx)

	Assert(t, r.Error != nil, "exp .Error to be set")
	Equals(t, "err", r.Error.Error())
}

func TestExecute_Success(t *testing.T) {
	t.Log("If there are no errors, the plan should be returned")
	p, runner, _ := setupPlanExecutorTest(t)
	When(p.VCSClient.GetModifiedFiles(matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest(), matchers.AnyVcsHost())).ThenReturn([]string{"file.tf"}, nil)
	When(p.Workspace.Clone(planCtx.Log, planCtx.BaseRepo, planCtx.HeadRepo, planCtx.Pull, "env")).
		ThenReturn("/tmp/clone-repo", nil)
	When(p.ProjectPreExecute.Execute(&planCtx, "/tmp/clone-repo", models.Project{RepoFullName: "", Path: "."})).
		ThenReturn(events.PreExecuteResult{
			LockResponse: locking.TryLockResponse{
				LockKey: "key",
			},
		})

	r := p.Execute(&planCtx)

	runner.VerifyWasCalledOnce().RunCommandWithVersion(
		planCtx.Log,
		"/tmp/clone-repo",
		[]string{"plan", "-refresh", "-no-color", "-out", "/tmp/clone-repo/env.tfplan", "-var", "atlantis_user=anubhavmishra"},
		nil,
		"env",
	)
	Assert(t, len(r.ProjectResults) == 1, "exp one project result")
	result := r.ProjectResults[0]
	Assert(t, result.PlanSuccess != nil, "exp plan success to not be nil")
	Equals(t, "", result.PlanSuccess.TerraformOutput)
	Equals(t, "lockurl-key", result.PlanSuccess.LockURL)
}

func TestExecute_PreExecuteResult(t *testing.T) {
	t.Log("If DefaultProjectPreExecutor.Execute returns a ProjectResult we should return it")
	p, _, _ := setupPlanExecutorTest(t)
	When(p.VCSClient.GetModifiedFiles(matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest(), matchers.AnyVcsHost())).ThenReturn([]string{"file.tf"}, nil)
	When(p.Workspace.Clone(planCtx.Log, planCtx.BaseRepo, planCtx.HeadRepo, planCtx.Pull, "env")).
		ThenReturn("/tmp/clone-repo", nil)
	projectResult := events.ProjectResult{
		Failure: "failure",
	}
	When(p.ProjectPreExecute.Execute(&planCtx, "/tmp/clone-repo", models.Project{RepoFullName: "", Path: "."})).
		ThenReturn(events.PreExecuteResult{ProjectResult: projectResult})
	r := p.Execute(&planCtx)

	Assert(t, len(r.ProjectResults) == 1, "exp one project result")
	result := r.ProjectResults[0]
	Equals(t, "failure", result.Failure)
}

func TestExecute_MultiProjectFailure(t *testing.T) {
	t.Log("If is an error planning in one project it should be returned. It shouldn't affect another project though.")
	p, runner, locker := setupPlanExecutorTest(t)
	// Two projects have been modified so we should run plan in two paths.
	When(p.VCSClient.GetModifiedFiles(matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest(), matchers.AnyVcsHost())).ThenReturn([]string{"path1/file.tf", "path2/file.tf"}, nil)
	When(p.Workspace.Clone(planCtx.Log, planCtx.BaseRepo, planCtx.HeadRepo, planCtx.Pull, "env")).
		ThenReturn("/tmp/clone-repo", nil)

	// Both projects will succeed in the PreExecute stage.
	When(p.ProjectPreExecute.Execute(&planCtx, "/tmp/clone-repo", models.Project{RepoFullName: "", Path: "path1"})).
		ThenReturn(events.PreExecuteResult{LockResponse: locking.TryLockResponse{LockKey: "key1"}})
	When(p.ProjectPreExecute.Execute(&planCtx, "/tmp/clone-repo", models.Project{RepoFullName: "", Path: "path2"})).
		ThenReturn(events.PreExecuteResult{LockResponse: locking.TryLockResponse{LockKey: "key2"}})

	// The first project will fail when running plan
	When(runner.RunCommandWithVersion(
		planCtx.Log,
		"/tmp/clone-repo/path1",
		[]string{"plan", "-refresh", "-no-color", "-out", "/tmp/clone-repo/path1/env.tfplan", "-var", "atlantis_user=anubhavmishra"},
		nil,
		"env",
	)).ThenReturn("", errors.New("path1 err"))
	// The second will succeed. We don't need to stub it because by default it
	// will return a nil error.
	r := p.Execute(&planCtx)

	// We expect Unlock to be called for the failed project.
	locker.VerifyWasCalledOnce().Unlock("key1")

	// So at the end we expect the first project to return an error and the second to be successful.
	Assert(t, len(r.ProjectResults) == 2, "exp two project results")
	result1 := r.ProjectResults[0]
	Assert(t, result1.Error != nil, "exp err to not be nil")
	Equals(t, "path1 err\n", result1.Error.Error())

	result2 := r.ProjectResults[1]
	Assert(t, result2.PlanSuccess != nil, "exp plan success to not be nil")
	Equals(t, "", result2.PlanSuccess.TerraformOutput)
	Equals(t, "lockurl-key2", result2.PlanSuccess.LockURL)
}

func TestExecute_PostPlanCommands(t *testing.T) {
	t.Log("Should execute post-plan commands and return if there is an error")
	p, _, _ := setupPlanExecutorTest(t)
	When(p.VCSClient.GetModifiedFiles(matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest(), matchers.AnyVcsHost())).ThenReturn([]string{"file.tf"}, nil)
	When(p.Workspace.Clone(planCtx.Log, planCtx.BaseRepo, planCtx.HeadRepo, planCtx.Pull, "env")).
		ThenReturn("/tmp/clone-repo", nil)
	When(p.ProjectPreExecute.Execute(&planCtx, "/tmp/clone-repo", models.Project{RepoFullName: "", Path: "."})).
		ThenReturn(events.PreExecuteResult{
			ProjectConfig: events.ProjectConfig{PostPlan: []string{"post-plan"}},
		})
	When(p.Run.Execute(planCtx.Log, []string{"post-plan"}, "/tmp/clone-repo", "env", nil, "post_plan")).
		ThenReturn("", errors.New("err"))

	r := p.Execute(&planCtx)

	Assert(t, len(r.ProjectResults) == 1, "exp one project result")
	result := r.ProjectResults[0]
	Assert(t, result.Error != nil, "exp plan error to not be nil")
	Equals(t, "running post plan commands: err", result.Error.Error())
}

func setupPlanExecutorTest(t *testing.T) (*events.PlanExecutor, *tmocks.MockRunner, *lmocks.MockLocker) {
	RegisterMockTestingT(t)
	vcsProxy := vcsmocks.NewMockClientProxy()
	w := mocks.NewMockWorkspace()
	ppe := mocks.NewMockProjectPreExecutor()
	runner := tmocks.NewMockRunner()
	locker := lmocks.NewMockLocker()
	run := rmocks.NewMockRunner()
	p := events.PlanExecutor{
		VCSClient:         vcsProxy,
		ProjectFinder:     &events.DefaultProjectFinder{},
		Workspace:         w,
		ProjectPreExecute: ppe,
		Terraform:         runner,
		Locker:            locker,
		Run:               run,
	}
	p.LockURL = func(id string) (url string) {
		return "lockurl-" + id
	}
	return &p, runner, locker
}
