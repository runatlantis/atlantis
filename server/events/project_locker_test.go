// Copyright 2017 HootSuite Media Inc.
//
// Licensed under the Apache License, Version 2.0 (the License);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an AS IS BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// Modified hereafter by contributors to runatlantis/atlantis.
//
package events_test

//import (
//	. "github.com/petergtz/pegomock"
//	. "github.com/runatlantis/atlantis/testing"
//)

//var ctx = events.CommandContext{
//	Command: &events.Command{
//		Name: events.Plan,
//	},
//	Log: logging.NewNoopLogger(),
//}
//var project = models.Project{}
//
//func TestExecute_LockErr(t *testing.T) {
//	t.Log("when there is an error returned from TryLock we return it")
//	p, l, _, _ := setupPreExecuteTest(t)
//	When(l.TryLock(project, "", ctx.Pull, ctx.User)).ThenReturn(locking.TryLockResponse{}, errors.New("err"))
//
//	res := p.Execute(&ctx, "", project)
//	Equals(t, "acquiring lock: err", res.ProjectResult.Error.Error())
//}
//
//func TestExecute_LockFailed(t *testing.T) {
//	t.Log("when we can't acquire a lock for this project and the lock is owned by a different pull, we get an error")
//	p, l, _, _ := setupPreExecuteTest(t)
//	// The response has LockAcquired: false and the pull request is a number
//	// different than the current pull.
//	When(l.TryLock(project, "", ctx.Pull, ctx.User)).ThenReturn(locking.TryLockResponse{
//		LockAcquired: false,
//		CurrLock:     models.ProjectLock{Pull: models.PullRequest{Num: ctx.Pull.Num + 1}},
//	}, nil)
//
//	res := p.Execute(&ctx, "", project)
//	Equals(t, "This project is currently locked by #1. The locking plan must be applied or discarded before future plans can execute.", res.ProjectResult.Failure)
//}
//
//func TestExecute_ConfigErr(t *testing.T) {
//	t.Log("when there is an error loading config, we return it")
//	p, l, _, _ := setupPreExecuteTest(t)
//	When(l.TryLock(project, "", ctx.Pull, ctx.User)).ThenReturn(locking.TryLockResponse{
//		LockAcquired: true,
//	}, nil)
//	When(p.ParserValidator.Exists("")).ThenReturn(true)
//	When(p.ParserValidator.Read("")).ThenReturn(events.ProjectConfig{}, errors.New("err"))
//
//	res := p.Execute(&ctx, "", project)
//	Equals(t, "err", res.ProjectResult.Error.Error())
//}
//
//func TestExecute_PreInitErr(t *testing.T) {
//	t.Log("when the project is on tf >= 0.9 and we run a `pre_init` that returns an error we return it")
//	p, l, tm, r := setupPreExecuteTest(t)
//	When(l.TryLock(project, "", ctx.Pull, ctx.User)).ThenReturn(locking.TryLockResponse{
//		LockAcquired: true,
//	}, nil)
//	When(p.ParserValidator.Exists("")).ThenReturn(true)
//	When(p.ParserValidator.Read("")).ThenReturn(events.ProjectConfig{
//		PreInit: []string{"pre-init"},
//	}, nil)
//	tfVersion, _ := version.NewVersion("0.9.0")
//	When(tm.Version()).ThenReturn(tfVersion)
//	When(r.Execute(ctx.Log, []string{"pre-init"}, "", "", tfVersion, "pre_init")).ThenReturn("", errors.New("err"))
//
//	res := p.Execute(&ctx, "", project)
//	Equals(t, "running pre_init commands: err", res.ProjectResult.Error.Error())
//}
//
//func TestExecute_InitErr(t *testing.T) {
//	t.Log("when the project is on tf >= 0.9 and we run `init` that returns an error we return it")
//	p, l, tm, _ := setupPreExecuteTest(t)
//	When(l.TryLock(project, "", ctx.Pull, ctx.User)).ThenReturn(locking.TryLockResponse{
//		LockAcquired: true,
//	}, nil)
//	When(p.ParserValidator.Exists("")).ThenReturn(true)
//	When(p.ParserValidator.Read("")).ThenReturn(events.ProjectConfig{}, nil)
//	tfVersion, _ := version.NewVersion("0.9.0")
//	When(tm.Version()).ThenReturn(tfVersion)
//	When(tm.Init(ctx.Log, "", "", nil, tfVersion)).ThenReturn(nil, errors.New("err"))
//
//	res := p.Execute(&ctx, "", project)
//	Equals(t, "err", res.ProjectResult.Error.Error())
//}
//
//func TestExecute_PreGetErr(t *testing.T) {
//	t.Log("when the project is on tf < 0.9 and we run a `pre_get` that returns an error we return it")
//	p, l, tm, r := setupPreExecuteTest(t)
//	When(l.TryLock(project, "", ctx.Pull, ctx.User)).ThenReturn(locking.TryLockResponse{
//		LockAcquired: true,
//	}, nil)
//	When(p.ParserValidator.Exists("")).ThenReturn(true)
//	When(p.ParserValidator.Read("")).ThenReturn(events.ProjectConfig{
//		PreGet: []string{"pre-get"},
//	}, nil)
//	tfVersion, _ := version.NewVersion("0.8")
//	When(tm.Version()).ThenReturn(tfVersion)
//	When(r.Execute(ctx.Log, []string{"pre-get"}, "", "", tfVersion, "pre_get")).ThenReturn("", errors.New("err"))
//
//	res := p.Execute(&ctx, "", project)
//	Equals(t, "running pre_get commands: err", res.ProjectResult.Error.Error())
//}
//
//func TestExecute_GetErr(t *testing.T) {
//	t.Log("when the project is on tf < 0.9 and we run `get` that returns an error we return it")
//	p, l, tm, _ := setupPreExecuteTest(t)
//	When(l.TryLock(project, "", ctx.Pull, ctx.User)).ThenReturn(locking.TryLockResponse{
//		LockAcquired: true,
//	}, nil)
//	When(p.ParserValidator.Exists("")).ThenReturn(true)
//	When(p.ParserValidator.Read("")).ThenReturn(events.ProjectConfig{}, nil)
//	tfVersion, _ := version.NewVersion("0.8")
//	When(tm.Version()).ThenReturn(tfVersion)
//	When(tm.RunCommandWithVersion(ctx.Log, "", []string{"get", "-no-color"}, tfVersion, "")).ThenReturn("", errors.New("err"))
//
//	res := p.Execute(&ctx, "", project)
//	Equals(t, "err", res.ProjectResult.Error.Error())
//}
//
//func TestExecute_PreCommandErr(t *testing.T) {
//	t.Log("when we get an error running pre commands we return it")
//	p, l, tm, r := setupPreExecuteTest(t)
//	When(l.TryLock(project, "", ctx.Pull, ctx.User)).ThenReturn(locking.TryLockResponse{
//		LockAcquired: true,
//	}, nil)
//	When(p.ParserValidator.Exists("")).ThenReturn(true)
//	When(p.ParserValidator.Read("")).ThenReturn(events.ProjectConfig{
//		PrePlan: []string{"command"},
//	}, nil)
//	tfVersion, _ := version.NewVersion("0.9")
//	When(tm.Version()).ThenReturn(tfVersion)
//	When(tm.Init(ctx.Log, "", "", nil, tfVersion)).ThenReturn(nil, nil)
//	When(r.Execute(ctx.Log, []string{"command"}, "", "", tfVersion, "pre_plan")).ThenReturn("", errors.New("err"))
//
//	res := p.Execute(&ctx, "", project)
//	Equals(t, "running pre_plan commands: err", res.ProjectResult.Error.Error())
//}
//
//func TestExecute_SuccessTF9(t *testing.T) {
//	t.Log("when the project is on tf >= 0.9 it should be successful")
//	p, l, tm, r := setupPreExecuteTest(t)
//	lockResponse := locking.TryLockResponse{
//		LockAcquired: true,
//	}
//	When(l.TryLock(project, "", ctx.Pull, ctx.User)).ThenReturn(lockResponse, nil)
//	When(p.ParserValidator.Exists("")).ThenReturn(true)
//	config := events.ProjectConfig{
//		PreInit: []string{"pre-init"},
//	}
//	When(p.ParserValidator.Read("")).ThenReturn(config, nil)
//	tfVersion, _ := version.NewVersion("0.9")
//	When(tm.Version()).ThenReturn(tfVersion)
//	When(tm.Init(ctx.Log, "", "", nil, tfVersion)).ThenReturn(nil, nil)
//
//	res := p.Execute(&ctx, "", project)
//	Equals(t, events.TryLockResponse{
//		ProjectConfig:    config,
//		TerraformVersion: tfVersion,
//		LockResponse:     lockResponse,
//	}, res)
//	tm.VerifyWasCalledOnce().Init(ctx.Log, "", "", nil, tfVersion)
//	r.VerifyWasCalledOnce().Execute(ctx.Log, []string{"pre-init"}, "", "", tfVersion, "pre_init")
//}
//
//func TestExecute_SuccessTF8(t *testing.T) {
//	t.Log("when the project is on tf < 0.9 it should be successful")
//	p, l, tm, r := setupPreExecuteTest(t)
//	lockResponse := locking.TryLockResponse{
//		LockAcquired: true,
//	}
//	When(l.TryLock(project, "", ctx.Pull, ctx.User)).ThenReturn(lockResponse, nil)
//	When(p.ParserValidator.Exists("")).ThenReturn(true)
//	config := events.ProjectConfig{
//		PreGet: []string{"pre-get"},
//	}
//	When(p.ParserValidator.Read("")).ThenReturn(config, nil)
//	tfVersion, _ := version.NewVersion("0.8")
//	When(tm.Version()).ThenReturn(tfVersion)
//
//	res := p.Execute(&ctx, "", project)
//	Equals(t, events.TryLockResponse{
//		ProjectConfig:    config,
//		TerraformVersion: tfVersion,
//		LockResponse:     lockResponse,
//	}, res)
//	tm.VerifyWasCalledOnce().RunCommandWithVersion(ctx.Log, "", []string{"get", "-no-color"}, tfVersion, "")
//	r.VerifyWasCalledOnce().Execute(ctx.Log, []string{"pre-get"}, "", "", tfVersion, "pre_get")
//}
//
//func TestExecute_SuccessPrePlan(t *testing.T) {
//	t.Log("when there are pre_plan commands they are run")
//	p, l, tm, r := setupPreExecuteTest(t)
//	lockResponse := locking.TryLockResponse{
//		LockAcquired: true,
//	}
//	When(l.TryLock(project, "", ctx.Pull, ctx.User)).ThenReturn(lockResponse, nil)
//	When(p.ParserValidator.Exists("")).ThenReturn(true)
//	config := events.ProjectConfig{
//		PrePlan: []string{"command"},
//	}
//	When(p.ParserValidator.Read("")).ThenReturn(config, nil)
//	tfVersion, _ := version.NewVersion("0.9")
//	When(tm.Version()).ThenReturn(tfVersion)
//
//	res := p.Execute(&ctx, "", project)
//	Equals(t, events.TryLockResponse{
//		ProjectConfig:    config,
//		TerraformVersion: tfVersion,
//		LockResponse:     lockResponse,
//	}, res)
//	r.VerifyWasCalledOnce().Execute(ctx.Log, []string{"command"}, "", "", tfVersion, "pre_plan")
//}
//
//func TestExecute_SuccessPreApply(t *testing.T) {
//	t.Log("when there are pre_apply commands they are run")
//	p, l, tm, r := setupPreExecuteTest(t)
//	lockResponse := locking.TryLockResponse{
//		LockAcquired: true,
//	}
//	When(l.TryLock(project, "", ctx.Pull, ctx.User)).ThenReturn(lockResponse, nil)
//	When(p.ParserValidator.Exists("")).ThenReturn(true)
//	config := events.ProjectConfig{
//		PreApply: []string{"command"},
//	}
//	When(p.ParserValidator.Read("")).ThenReturn(config, nil)
//	tfVersion, _ := version.NewVersion("0.9")
//	When(tm.Version()).ThenReturn(tfVersion)
//
//	cpCtx := deepcopy.Copy(ctx).(events.CommandContext)
//	cpCtx.Command = &events.Command{
//		Name: events.Apply,
//	}
//	cpCtx.Log = logging.NewNoopLogger()
//
//	res := p.Execute(&cpCtx, "", project)
//	Equals(t, events.TryLockResponse{
//		ProjectConfig:    config,
//		TerraformVersion: tfVersion,
//		LockResponse:     lockResponse,
//	}, res)
//	r.VerifyWasCalledOnce().Execute(cpCtx.Log, []string{"command"}, "", "", tfVersion, "pre_apply")
//}
//
//func setupPreExecuteTest(t *testing.T) (*events.DefaultProjectLocker, *lmocks.MockLocker, *tmocks.MockClient, *rmocks.MockRunner) {
//	RegisterMockTestingT(t)
//	l := lmocks.NewMockLocker()
//	cr := mocks.NewMockProjectConfigReader()
//	tm := tmocks.NewMockClient()
//	r := rmocks.NewMockRunner()
//	return &events.DefaultProjectLocker{
//		Locker:       l,
//		ParserValidator: cr,
//		Terraform:    tm,
//		Run:          r,
//	}, l, tm, r
//}
