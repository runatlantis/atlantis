package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

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
	Locker                         locking.Locker
	Logger                         logging.SimpleLogging
	Parser                         events.EventParsing
	ProjectCommandBuilder          events.ProjectCommandBuilder
	ProjectPlanCommandRunner       events.ProjectPlanCommandRunner
	ProjectApplyCommandRunner      events.ProjectApplyCommandRunner
	FailOnPreWorkflowHookError     bool
	PreWorkflowHooksCommandRunner  events.PreWorkflowHooksCommandRunner
	PostWorkflowHooksCommandRunner events.PostWorkflowHooksCommandRunner
	RepoAllowlistChecker           *events.RepoAllowlistChecker
	Scope                          tally.Scope
	VCSClient                      vcs.Client
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

	result, err := a.apiPlan(request, ctx)
	if err != nil {
		a.apiReportError(w, http.StatusInternalServerError, err)
		return
	}
	defer a.Locker.UnlockByPull(ctx.HeadRepo.FullName, 0) // nolint: errcheck
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

	// We must first make the plan for all projects
	_, err = a.apiPlan(request, ctx)
	if err != nil {
		a.apiReportError(w, http.StatusInternalServerError, err)
		return
	}
	defer a.Locker.UnlockByPull(ctx.HeadRepo.FullName, 0) // nolint: errcheck

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

func (a *APIController) apiPlan(request *APIRequest, ctx *command.Context) (*command.Result, error) {
	cmds, cc, err := request.getCommands(ctx, a.ProjectCommandBuilder.BuildPlanCommands)
	if err != nil {
		return nil, err
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
