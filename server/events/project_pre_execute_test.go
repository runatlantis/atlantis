package events_test

import (
	"errors"
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/hootsuite/atlantis/server/events"
	"github.com/hootsuite/atlantis/server/events/locking"
	lmocks "github.com/hootsuite/atlantis/server/events/locking/mocks"
	"github.com/hootsuite/atlantis/server/events/mocks"
	"github.com/hootsuite/atlantis/server/events/models"
	rmocks "github.com/hootsuite/atlantis/server/events/run/mocks"
	tmocks "github.com/hootsuite/atlantis/server/events/terraform/mocks"
	"github.com/hootsuite/atlantis/server/logging"
	. "github.com/hootsuite/atlantis/testing_util"
	"github.com/mohae/deepcopy"
	. "github.com/petergtz/pegomock"
)

var ctx = events.CommandContext{
	Command: &events.Command{
		Name: events.Plan,
	},
	Log: logging.NewNoopLogger(),
}
var project = models.Project{}

func TestExecute_LockErr(t *testing.T) {
	t.Log("when there is an error returned from TryLock we return it")
	p, l, _, _, _ := setupPreExecuteTest(t)
	When(l.TryLock(project, "", ctx.Pull, ctx.User)).ThenReturn(locking.TryLockResponse{}, errors.New("err"))

	res := p.Execute(&ctx, "", project)
	Equals(t, "acquiring lock: err", res.ProjectResult.Error.Error())
}

func TestExecute_LockFailed(t *testing.T) {
	t.Log("when we can't acquire a lock for this project and the lock is owned by a different pull, we get an error")
	p, l, _, _, _ := setupPreExecuteTest(t)
	// The response has LockAcquired: false and the pull request is a number
	// different than the current pull.
	When(l.TryLock(project, "", ctx.Pull, ctx.User)).ThenReturn(locking.TryLockResponse{
		LockAcquired: false,
		CurrLock:     models.ProjectLock{Pull: models.PullRequest{Num: ctx.Pull.Num + 1}},
	}, nil)

	res := p.Execute(&ctx, "", project)
	Equals(t, "This project is currently locked by #1. The locking plan must be applied or discarded before future plans can execute.", res.ProjectResult.Failure)
}

func TestExecute_ConfigErr(t *testing.T) {
	t.Log("when there is an error loading config, we return it")
	p, l, _, _, _ := setupPreExecuteTest(t)
	When(l.TryLock(project, "", ctx.Pull, ctx.User)).ThenReturn(locking.TryLockResponse{
		LockAcquired: true,
	}, nil)
	When(p.ConfigReader.Exists("")).ThenReturn(true)
	When(p.ConfigReader.Read("")).ThenReturn(events.ProjectConfig{}, errors.New("err"))

	res := p.Execute(&ctx, "", project)
	Equals(t, "err", res.ProjectResult.Error.Error())
}

func TestExecute_InitErr(t *testing.T) {
	t.Log("when the project is on tf >= 0.9 and we run `init` that returns an error we return it")
	p, l, _, tm, _ := setupPreExecuteTest(t)
	When(l.TryLock(project, "", ctx.Pull, ctx.User)).ThenReturn(locking.TryLockResponse{
		LockAcquired: true,
	}, nil)
	When(p.ConfigReader.Exists("")).ThenReturn(true)
	When(p.ConfigReader.Read("")).ThenReturn(events.ProjectConfig{}, nil)
	tfVersion, _ := version.NewVersion("0.9.0")
	When(tm.Version()).ThenReturn(tfVersion)
	When(tm.RunInitAndEnv(ctx.Log, "", "", nil, tfVersion)).ThenReturn(nil, errors.New("err"))

	res := p.Execute(&ctx, "", project)
	Equals(t, "err", res.ProjectResult.Error.Error())
}

func TestExecute_GetErr(t *testing.T) {
	t.Log("when the project is on tf < 0.9 and we run `get` that returns an error we return it")
	p, l, _, tm, _ := setupPreExecuteTest(t)
	When(l.TryLock(project, "", ctx.Pull, ctx.User)).ThenReturn(locking.TryLockResponse{
		LockAcquired: true,
	}, nil)
	When(p.ConfigReader.Exists("")).ThenReturn(true)
	When(p.ConfigReader.Read("")).ThenReturn(events.ProjectConfig{}, nil)
	tfVersion, _ := version.NewVersion("0.8")
	When(tm.Version()).ThenReturn(tfVersion)
	When(tm.RunCommandWithVersion(ctx.Log, "", []string{"get", "-no-color"}, tfVersion, "")).ThenReturn("", errors.New("err"))

	res := p.Execute(&ctx, "", project)
	Equals(t, "err", res.ProjectResult.Error.Error())
}

func TestExecute_PreCommandErr(t *testing.T) {
	t.Log("when we get an error running pre commands we return it")
	p, l, _, tm, r := setupPreExecuteTest(t)
	When(l.TryLock(project, "", ctx.Pull, ctx.User)).ThenReturn(locking.TryLockResponse{
		LockAcquired: true,
	}, nil)
	When(p.ConfigReader.Exists("")).ThenReturn(true)
	When(p.ConfigReader.Read("")).ThenReturn(events.ProjectConfig{
		PrePlan: events.PrePlan{Commands: []string{"command"}},
	}, nil)
	tfVersion, _ := version.NewVersion("0.9")
	When(tm.Version()).ThenReturn(tfVersion)
	When(tm.RunInitAndEnv(ctx.Log, "", "", nil, tfVersion)).ThenReturn(nil, nil)
	When(r.Execute(ctx.Log, []string{"command"}, "", "", tfVersion, "pre_plan")).ThenReturn("", errors.New("err"))

	res := p.Execute(&ctx, "", project)
	Equals(t, "running pre_plan commands: err", res.ProjectResult.Error.Error())
}

func TestExecute_SuccessTF9(t *testing.T) {
	t.Log("when the project is on tf >= 0.9 it should be successful")
	p, l, _, tm, _ := setupPreExecuteTest(t)
	lockResponse := locking.TryLockResponse{
		LockAcquired: true,
	}
	When(l.TryLock(project, "", ctx.Pull, ctx.User)).ThenReturn(lockResponse, nil)
	When(p.ConfigReader.Exists("")).ThenReturn(true)
	When(p.ConfigReader.Read("")).ThenReturn(events.ProjectConfig{}, nil)
	tfVersion, _ := version.NewVersion("0.9")
	When(tm.Version()).ThenReturn(tfVersion)
	When(tm.RunInitAndEnv(ctx.Log, "", "", nil, tfVersion)).ThenReturn(nil, nil)

	res := p.Execute(&ctx, "", project)
	Equals(t, events.PreExecuteResult{
		ProjectConfig:    events.ProjectConfig{},
		TerraformVersion: tfVersion,
		LockResponse:     lockResponse,
	}, res)
	tm.VerifyWasCalledOnce().RunInitAndEnv(ctx.Log, "", "", nil, tfVersion)
}

func TestExecute_SuccessTF8(t *testing.T) {
	t.Log("when the project is on tf < 0.9 it should be successful")
	p, l, _, tm, _ := setupPreExecuteTest(t)
	lockResponse := locking.TryLockResponse{
		LockAcquired: true,
	}
	When(l.TryLock(project, "", ctx.Pull, ctx.User)).ThenReturn(lockResponse, nil)
	When(p.ConfigReader.Exists("")).ThenReturn(true)
	When(p.ConfigReader.Read("")).ThenReturn(events.ProjectConfig{}, nil)
	tfVersion, _ := version.NewVersion("0.8")
	When(tm.Version()).ThenReturn(tfVersion)

	res := p.Execute(&ctx, "", project)
	Equals(t, events.PreExecuteResult{
		ProjectConfig:    events.ProjectConfig{},
		TerraformVersion: tfVersion,
		LockResponse:     lockResponse,
	}, res)
	tm.VerifyWasCalledOnce().RunCommandWithVersion(ctx.Log, "", []string{"get", "-no-color"}, tfVersion, "")
}

func TestExecute_SuccessPrePlan(t *testing.T) {
	t.Log("when there are pre_plan commands they are run")
	p, l, _, tm, r := setupPreExecuteTest(t)
	lockResponse := locking.TryLockResponse{
		LockAcquired: true,
	}
	When(l.TryLock(project, "", ctx.Pull, ctx.User)).ThenReturn(lockResponse, nil)
	When(p.ConfigReader.Exists("")).ThenReturn(true)
	config := events.ProjectConfig{
		PrePlan: events.PrePlan{Commands: []string{"command"}},
	}
	When(p.ConfigReader.Read("")).ThenReturn(config, nil)
	tfVersion, _ := version.NewVersion("0.9")
	When(tm.Version()).ThenReturn(tfVersion)

	res := p.Execute(&ctx, "", project)
	Equals(t, events.PreExecuteResult{
		ProjectConfig:    config,
		TerraformVersion: tfVersion,
		LockResponse:     lockResponse,
	}, res)
	r.VerifyWasCalledOnce().Execute(ctx.Log, []string{"command"}, "", "", tfVersion, "pre_plan")
}

func TestExecute_SuccessPreApply(t *testing.T) {
	t.Log("when there are pre_apply commands they are run")
	p, l, _, tm, r := setupPreExecuteTest(t)
	lockResponse := locking.TryLockResponse{
		LockAcquired: true,
	}
	When(l.TryLock(project, "", ctx.Pull, ctx.User)).ThenReturn(lockResponse, nil)
	When(p.ConfigReader.Exists("")).ThenReturn(true)
	config := events.ProjectConfig{
		PreApply: events.PreApply{Commands: []string{"command"}},
	}
	When(p.ConfigReader.Read("")).ThenReturn(config, nil)
	tfVersion, _ := version.NewVersion("0.9")
	When(tm.Version()).ThenReturn(tfVersion)

	cpCtx := deepcopy.Copy(ctx).(events.CommandContext)
	cpCtx.Command = &events.Command{
		Name: events.Apply,
	}
	cpCtx.Log = logging.NewNoopLogger()

	res := p.Execute(&cpCtx, "", project)
	Equals(t, events.PreExecuteResult{
		ProjectConfig:    config,
		TerraformVersion: tfVersion,
		LockResponse:     lockResponse,
	}, res)
	r.VerifyWasCalledOnce().Execute(cpCtx.Log, []string{"command"}, "", "", tfVersion, "pre_apply")
}

func setupPreExecuteTest(t *testing.T) (*events.ProjectPreExecute, *lmocks.MockLocker, *mocks.MockProjectConfigReader, *tmocks.MockRunner, *rmocks.MockRunner) {
	RegisterMockTestingT(t)
	l := lmocks.NewMockLocker()
	cr := mocks.NewMockProjectConfigReader()
	tm := tmocks.NewMockRunner()
	r := rmocks.NewMockRunner()
	return &events.ProjectPreExecute{
		Locker:       l,
		ConfigReader: cr,
		Terraform:    tm,
		Run:          r,
	}, l, cr, tm, r
}
