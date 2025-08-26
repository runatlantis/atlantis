package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/runatlantis/atlantis/server/core/locking"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/logging"
	tally "github.com/uber-go/tally/v4"
)

const atlantisTokenHeader = "X-Atlantis-Token"

type APIController struct {
	APISecret                      []byte
	Locker                         locking.Locker                   `validate:"required"`
	Logger                         logging.SimpleLogging            `validate:"required"`
	Parser                         events.EventParsing              `validate:"required"`
	ProjectCommandBuilder          events.ProjectCommandBuilder     `validate:"required"`
	ProjectPlanCommandRunner       events.ProjectPlanCommandRunner  `validate:"required"`
	ProjectApplyCommandRunner      events.ProjectApplyCommandRunner `validate:"required"`
	FailOnPreWorkflowHookError     bool
	PreWorkflowHooksCommandRunner  events.PreWorkflowHooksCommandRunner  `validate:"required"`
	PostWorkflowHooksCommandRunner events.PostWorkflowHooksCommandRunner `validate:"required"`
	RepoAllowlistChecker           *events.RepoAllowlistChecker          `validate:"required"`
	Scope                          tally.Scope                           `validate:"required"`
	VCSClient                      vcs.Client                            `validate:"required"`
	WorkingDir                     events.WorkingDir                     `validate:"required"`
	WorkingDirLocker               events.WorkingDirLocker               `validate:"required"`
	CommitStatusUpdater            events.CommitStatusUpdater            `validate:"required"`
	// SilenceVCSStatusNoProjects is whether API should set commit status if no projects are found
	SilenceVCSStatusNoProjects     bool
}

type APIRequest struct {
	Repository string `validate:"required"`
	Ref        string `validate:"required"`
	Type       string `validate:"required"`
	PR         int
	Projects   []string
	Paths      []struct {
		Directory string
		Workspace string
	}
}

func (a *APIRequest) getCommands(ctx *command.Context, cmdBuilder func(*command.Context, *events.CommentCommand) ([]command.ProjectContext, error)) ([]command.ProjectContext, []*events.CommentCommand, error) {
	cc := make([]*events.CommentCommand, 0)

	for _, project := range a.Projects {
		cc = append(cc, &events.CommentCommand{
			ProjectName: project,
		})
	}
	for _, path := range a.Paths {
		cc = append(cc, &events.CommentCommand{
			RepoRelDir: strings.TrimRight(path.Directory, "/"),
			Workspace:  path.Workspace,
		})
	}

	cmds := make([]command.ProjectContext, 0)
	for _, commentCommand := range cc {
		projectCmds, err := cmdBuilder(ctx, commentCommand)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to build command: %v", err)
		}
		cmds = append(cmds, projectCmds...)
	}

	return cmds, cc, nil
}

func (a *APIController) apiReportError(w http.ResponseWriter, code int, err error) {
	response, _ := json.Marshal(map[string]string{
		"error": err.Error(),
	})
	a.respond(w, logging.Warn, code, "%s", string(response))
}

func (a *APIController) Plan(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	request, ctx, code, err := a.apiParseAndValidate(r)
	if err != nil {
		a.apiReportError(w, code, err)
		return
	}

	err = a.apiSetup(ctx)
	if err != nil {
		a.apiReportError(w, http.StatusInternalServerError, err)
		return
	}

	result, err := a.apiPlan(request, ctx)
	if err != nil {
		a.apiReportError(w, http.StatusInternalServerError, err)
		return
	}
	defer a.Locker.UnlockByPull(ctx.HeadRepo.FullName, ctx.Pull.Num) // nolint: errcheck
	if result.HasErrors() {
		code = http.StatusInternalServerError
	}

	// TODO: make a better response
	response, err := json.Marshal(result)
	if err != nil {
		a.apiReportError(w, http.StatusInternalServerError, err)
		return
	}
	a.respond(w, logging.Warn, code, "%s", string(response))
}

func (a *APIController) Apply(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	request, ctx, code, err := a.apiParseAndValidate(r)
	if err != nil {
		a.apiReportError(w, code, err)
		return
	}

	err = a.apiSetup(ctx)
	if err != nil {
		a.apiReportError(w, http.StatusInternalServerError, err)
		return
	}

	// We must first make the plan for all projects
	_, err = a.apiPlan(request, ctx)
	if err != nil {
		a.apiReportError(w, http.StatusInternalServerError, err)
		return
	}
	defer a.Locker.UnlockByPull(ctx.HeadRepo.FullName, ctx.Pull.Num) // nolint: errcheck

	// We can now prepare and run the apply step
	result, err := a.apiApply(request, ctx)
	if err != nil {
		a.apiReportError(w, http.StatusInternalServerError, err)
		return
	}
	if result.HasErrors() {
		code = http.StatusInternalServerError
	}

	response, err := json.Marshal(result)
	if err != nil {
		a.apiReportError(w, http.StatusInternalServerError, err)
		return
	}
	a.respond(w, logging.Warn, code, "%s", string(response))
}

type LockDetail struct {
	Name            string
	ProjectName     string
	ProjectRepo     string
	ProjectRepoPath string
	PullID          int `json:",string"`
	PullURL         string
	User            string
	Workspace       string
	Time            time.Time
}

type ListLocksResult struct {
	Locks []LockDetail
}

func (a *APIController) ListLocks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	locks, err := a.Locker.List()
	if err != nil {
		a.apiReportError(w, http.StatusInternalServerError, err)
		return
	}

	result := ListLocksResult{}
	for name, lock := range locks {
		lockDetail := LockDetail{
			name,
			lock.Project.ProjectName,
			lock.Project.RepoFullName,
			lock.Project.Path,
			lock.Pull.Num,
			lock.Pull.URL,
			lock.User.Username,
			lock.Workspace,
			lock.Time,
		}
		result.Locks = append(result.Locks, lockDetail)
	}

	response, err := json.Marshal(result)
	if err != nil {
		a.apiReportError(w, http.StatusInternalServerError, err)
		return
	}
	a.respond(w, logging.Warn, http.StatusOK, "%s", string(response))
}

func (a *APIController) apiSetup(ctx *command.Context) error {
	pull := ctx.Pull
	baseRepo := ctx.Pull.BaseRepo
	headRepo := ctx.HeadRepo

	unlockFn, err := a.WorkingDirLocker.TryLock(baseRepo.FullName, pull.Num, events.DefaultWorkspace, events.DefaultRepoRelDir)
	if err != nil {
		return err
	}
	ctx.Log.Debug("got workspace lock")
	defer unlockFn()

	// ensure workingDir is present
	_, err = a.WorkingDir.Clone(ctx.Log, headRepo, pull, events.DefaultWorkspace)
	if err != nil {
		return err
	}

	return nil
}

func (a *APIController) apiPlan(request *APIRequest, ctx *command.Context) (*command.Result, error) {
	cmds, cc, err := request.getCommands(ctx, a.ProjectCommandBuilder.BuildPlanCommands)
	if err != nil {
		return nil, err
	}

	// Update the combined plan commit status to pending
	if err := a.CommitStatusUpdater.UpdateCombined(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.PendingCommitStatus, command.Plan); err != nil {
		ctx.Log.Warn("unable to update plan commit status: %s", err)
	}

	if len(cmds) == 0 {
		ctx.Log.Info("determined there was no project to run plan in")
		if a.SilenceVCSStatusNoProjects {
			// If silence is enabled but a pending status was already set above,
			// we need to clear it to avoid leaving the PR check stuck in pending state
			ctx.Log.Debug("clearing pending status since no projects found and silence is enabled")
			if err := a.CommitStatusUpdater.UpdateCombinedCount(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.SuccessCommitStatus, command.Plan, 0, 0); err != nil {
				ctx.Log.Warn("unable to clear pending plan status: %s", err)
			}
			if err := a.CommitStatusUpdater.UpdateCombinedCount(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.SuccessCommitStatus, command.PolicyCheck, 0, 0); err != nil {
				ctx.Log.Warn("unable to clear pending policy check status: %s", err)
			}
			if err := a.CommitStatusUpdater.UpdateCombinedCount(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.SuccessCommitStatus, command.Apply, 0, 0); err != nil {
				ctx.Log.Warn("unable to clear pending apply status: %s", err)
			}
		}
		return &command.Result{ProjectResults: []command.ProjectResult{}}, nil
	}

	var projectResults []command.ProjectResult
	for i, cmd := range cmds {
		err = a.PreWorkflowHooksCommandRunner.RunPreHooks(ctx, cc[i])
		if err != nil {
			if a.FailOnPreWorkflowHookError {
				return nil, err
			}
		}

		res := a.ProjectPlanCommandRunner.Plan(cmd)
		projectResults = append(projectResults, res)

		a.PostWorkflowHooksCommandRunner.RunPostHooks(ctx, cc[i]) // nolint: errcheck
	}
	return &command.Result{ProjectResults: projectResults}, nil
}

func (a *APIController) apiApply(request *APIRequest, ctx *command.Context) (*command.Result, error) {
	cmds, cc, err := request.getCommands(ctx, a.ProjectCommandBuilder.BuildApplyCommands)
	if err != nil {
		return nil, err
	}

	// Update the combined apply commit status to pending
	if err := a.CommitStatusUpdater.UpdateCombined(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.PendingCommitStatus, command.Apply); err != nil {
		ctx.Log.Warn("unable to update apply commit status: %s", err)
	}

	if len(cmds) == 0 {
		ctx.Log.Info("determined there was no project to run apply in")
		if a.SilenceVCSStatusNoProjects {
			// If silence is enabled but a pending status was already set above,
			// we need to clear it to avoid leaving the PR check stuck in pending state
			ctx.Log.Debug("clearing pending status since no projects found and silence is enabled")
			if err := a.CommitStatusUpdater.UpdateCombinedCount(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.SuccessCommitStatus, command.Plan, 0, 0); err != nil {
				ctx.Log.Warn("unable to clear pending plan status: %s", err)
			}
			if err := a.CommitStatusUpdater.UpdateCombinedCount(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.SuccessCommitStatus, command.PolicyCheck, 0, 0); err != nil {
				ctx.Log.Warn("unable to clear pending policy check status: %s", err)
			}
			if err := a.CommitStatusUpdater.UpdateCombinedCount(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.SuccessCommitStatus, command.Apply, 0, 0); err != nil {
				ctx.Log.Warn("unable to clear pending apply status: %s", err)
			}
		}
		return &command.Result{ProjectResults: []command.ProjectResult{}}, nil
	}

	var projectResults []command.ProjectResult
	for i, cmd := range cmds {
		err = a.PreWorkflowHooksCommandRunner.RunPreHooks(ctx, cc[i])
		if err != nil {
			if a.FailOnPreWorkflowHookError {
				return nil, err
			}
		}

		res := a.ProjectApplyCommandRunner.Apply(cmd)
		projectResults = append(projectResults, res)

		a.PostWorkflowHooksCommandRunner.RunPostHooks(ctx, cc[i]) // nolint: errcheck
	}
	return &command.Result{ProjectResults: projectResults}, nil
}

func (a *APIController) apiParseAndValidate(r *http.Request) (*APIRequest, *command.Context, int, error) {
	if len(a.APISecret) == 0 {
		return nil, nil, http.StatusBadRequest, fmt.Errorf("ignoring request since API is disabled")
	}

	// Validate the secret token
	secret := r.Header.Get(atlantisTokenHeader)
	if secret != string(a.APISecret) {
		return nil, nil, http.StatusUnauthorized, fmt.Errorf("header %s did not match expected secret", atlantisTokenHeader)
	}

	// Parse the JSON payload
	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, nil, http.StatusBadRequest, fmt.Errorf("failed to read request")
	}
	var request APIRequest
	if err = json.Unmarshal(bytes, &request); err != nil {
		return nil, nil, http.StatusBadRequest, fmt.Errorf("failed to parse request: %v", err.Error())
	}
	if err = validator.New().Struct(request); err != nil {
		return nil, nil, http.StatusBadRequest, fmt.Errorf("request %q is missing fields", string(bytes))
	}

	VCSHostType, err := models.NewVCSHostType(request.Type)
	if err != nil {
		return nil, nil, http.StatusBadRequest, err
	}
	cloneURL, err := a.VCSClient.GetCloneURL(a.Logger, VCSHostType, request.Repository)
	if err != nil {
		return nil, nil, http.StatusInternalServerError, err
	}

	baseRepo, err := a.Parser.ParseAPIPlanRequest(VCSHostType, request.Repository, cloneURL)
	if err != nil {
		return nil, nil, http.StatusBadRequest, fmt.Errorf("failed to parse request: %v", err)
	}

	// Check if the repo is allowlisted
	if !a.RepoAllowlistChecker.IsAllowlisted(baseRepo.FullName, baseRepo.VCSHost.Hostname) {
		return nil, nil, http.StatusForbidden, fmt.Errorf("repo not allowlisted")
	}

	return &request, &command.Context{
		HeadRepo: baseRepo,
		Pull: models.PullRequest{
			Num:        request.PR,
			BaseBranch: request.Ref,
			HeadBranch: request.Ref,
			HeadCommit: request.Ref,
			BaseRepo:   baseRepo,
		},
		Scope: a.Scope,
		Log:   a.Logger,
		API:   true,
	}, http.StatusOK, nil
}

func (a *APIController) respond(w http.ResponseWriter, lvl logging.LogLevel, responseCode int, format string, args ...interface{}) {
	response := fmt.Sprintf(format, args...)
	a.Logger.Log(lvl, response)
	w.WriteHeader(responseCode)
	fmt.Fprintln(w, response)
}
