// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/runatlantis/atlantis/server/core/drift"
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
	DriftStorage                   drift.Storage
	RemediationService             drift.RemediationService
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
	SilenceVCSStatusNoProjects bool
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

	err = a.apiSetup(ctx, command.Plan)
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

	err = a.apiSetup(ctx, command.Apply)
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

// DriftStatus returns the drift status for a repository.
// This is a non-authenticated endpoint that returns cached drift detection results.
// Query parameters:
//   - repository: required, the full repository name (owner/repo)
//   - project: optional, filter by project name
//   - workspace: optional, filter by workspace
func (a *APIController) DriftStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Check if drift storage is configured
	if a.DriftStorage == nil {
		a.apiReportError(w, http.StatusServiceUnavailable, fmt.Errorf("drift detection is not enabled"))
		return
	}

	// Get query parameters
	repository := r.URL.Query().Get("repository")
	if repository == "" {
		a.apiReportError(w, http.StatusBadRequest, fmt.Errorf("repository parameter is required"))
		return
	}

	opts := drift.GetOptions{
		ProjectName: r.URL.Query().Get("project"),
		Workspace:   r.URL.Query().Get("workspace"),
	}

	// Retrieve drift results from storage
	drifts, err := a.DriftStorage.Get(repository, opts)
	if err != nil {
		a.apiReportError(w, http.StatusInternalServerError, err)
		return
	}

	// Build response
	result := models.NewDriftStatusResponse(repository, drifts)

	response, err := json.Marshal(result)
	if err != nil {
		a.apiReportError(w, http.StatusInternalServerError, err)
		return
	}
	a.respond(w, logging.Info, http.StatusOK, "%s", string(response))
}

func (a *APIController) apiSetup(ctx *command.Context, cmdName command.Name) error {
	pull := ctx.Pull
	baseRepo := ctx.Pull.BaseRepo
	headRepo := ctx.HeadRepo

	unlockFn, err := a.WorkingDirLocker.TryLock(baseRepo.FullName, pull.Num, events.DefaultWorkspace, events.DefaultRepoRelDir, "", cmdName)
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

	if len(cmds) == 0 {
		ctx.Log.Info("determined there was no project to run plan in")
		// When silence is enabled and no projects are found, don't set any VCS status
		if !a.SilenceVCSStatusNoProjects {
			ctx.Log.Debug("setting VCS status to success with no projects found")
			if err := a.CommitStatusUpdater.UpdateCombinedCount(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.SuccessCommitStatus, command.Plan, 0, 0); err != nil {
				ctx.Log.Warn("unable to update plan status: %s", err)
			}
			if err := a.CommitStatusUpdater.UpdateCombinedCount(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.SuccessCommitStatus, command.PolicyCheck, 0, 0); err != nil {
				ctx.Log.Warn("unable to update policy check status: %s", err)
			}
			if err := a.CommitStatusUpdater.UpdateCombinedCount(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.SuccessCommitStatus, command.Apply, 0, 0); err != nil {
				ctx.Log.Warn("unable to update apply status: %s", err)
			}
		} else {
			ctx.Log.Debug("silence enabled and no projects found - not setting any VCS status")
		}
		return &command.Result{ProjectResults: []command.ProjectResult{}}, nil
	}

	// Update the combined plan commit status to pending
	if err := a.CommitStatusUpdater.UpdateCombined(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.PendingCommitStatus, command.Plan); err != nil {
		ctx.Log.Warn("unable to update plan commit status: %s", err)
	}

	var projectResults []command.ProjectResult
	for i, cmd := range cmds {
		err = a.PreWorkflowHooksCommandRunner.RunPreHooks(ctx, cc[i])
		if err != nil {
			if a.FailOnPreWorkflowHookError {
				return nil, err
			}
		}

		res := events.RunOneProjectCmd(a.ProjectPlanCommandRunner.Plan, cmd)
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

	if len(cmds) == 0 {
		ctx.Log.Info("determined there was no project to run apply in")
		// When silence is enabled and no projects are found, don't set any VCS status
		if !a.SilenceVCSStatusNoProjects {
			ctx.Log.Debug("setting VCS status to success with no projects found")
			if err := a.CommitStatusUpdater.UpdateCombinedCount(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.SuccessCommitStatus, command.Plan, 0, 0); err != nil {
				ctx.Log.Warn("unable to update plan status: %s", err)
			}
			if err := a.CommitStatusUpdater.UpdateCombinedCount(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.SuccessCommitStatus, command.PolicyCheck, 0, 0); err != nil {
				ctx.Log.Warn("unable to update policy check status: %s", err)
			}
			if err := a.CommitStatusUpdater.UpdateCombinedCount(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.SuccessCommitStatus, command.Apply, 0, 0); err != nil {
				ctx.Log.Warn("unable to update apply status: %s", err)
			}
		} else {
			ctx.Log.Debug("silence enabled and no projects found - not setting any VCS status")
		}
		return &command.Result{ProjectResults: []command.ProjectResult{}}, nil
	}

	// Update the combined apply commit status to pending
	if err := a.CommitStatusUpdater.UpdateCombined(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.PendingCommitStatus, command.Apply); err != nil {
		ctx.Log.Warn("unable to update apply commit status: %s", err)
	}

	var projectResults []command.ProjectResult
	for i, cmd := range cmds {
		err = a.PreWorkflowHooksCommandRunner.RunPreHooks(ctx, cc[i])
		if err != nil {
			if a.FailOnPreWorkflowHookError {
				return nil, err
			}
		}

		res := events.RunOneProjectCmd(a.ProjectApplyCommandRunner.Apply, cmd)
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

func (a *APIController) respond(w http.ResponseWriter, lvl logging.LogLevel, responseCode int, format string, args ...any) {
	response := fmt.Sprintf(format, args...)
	a.Logger.Log(lvl, response)
	w.WriteHeader(responseCode)
	fmt.Fprintln(w, response)
}

// Remediate handles POST /api/drift/remediate requests.
// It executes drift remediation (plan or apply) for the specified projects.
// This is an authenticated endpoint that requires the API secret.
func (a *APIController) Remediate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Check API is enabled
	if len(a.APISecret) == 0 {
		a.apiReportError(w, http.StatusBadRequest, fmt.Errorf("ignoring request since API is disabled"))
		return
	}

	// Validate the secret token
	secret := r.Header.Get(atlantisTokenHeader)
	if secret != string(a.APISecret) {
		a.apiReportError(w, http.StatusUnauthorized, fmt.Errorf("header %s did not match expected secret", atlantisTokenHeader))
		return
	}

	// Check if remediation service is configured
	if a.RemediationService == nil {
		a.apiReportError(w, http.StatusServiceUnavailable, fmt.Errorf("drift remediation is not enabled"))
		return
	}

	// Parse the JSON payload
	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		a.apiReportError(w, http.StatusBadRequest, fmt.Errorf("failed to read request: %v", err))
		return
	}

	var request models.RemediationRequest
	if err = json.Unmarshal(bytes, &request); err != nil {
		a.apiReportError(w, http.StatusBadRequest, fmt.Errorf("failed to parse request: %v", err))
		return
	}

	// Validate required fields
	if request.Repository == "" {
		a.apiReportError(w, http.StatusBadRequest, fmt.Errorf("repository is required"))
		return
	}
	if request.Ref == "" {
		a.apiReportError(w, http.StatusBadRequest, fmt.Errorf("ref is required"))
		return
	}
	if request.Type == "" {
		a.apiReportError(w, http.StatusBadRequest, fmt.Errorf("type is required"))
		return
	}

	// Check if the repo is allowlisted
	VCSHostType, err := models.NewVCSHostType(request.Type)
	if err != nil {
		a.apiReportError(w, http.StatusBadRequest, err)
		return
	}
	cloneURL, err := a.VCSClient.GetCloneURL(a.Logger, VCSHostType, request.Repository)
	if err != nil {
		a.apiReportError(w, http.StatusInternalServerError, err)
		return
	}

	baseRepo, err := a.Parser.ParseAPIPlanRequest(VCSHostType, request.Repository, cloneURL)
	if err != nil {
		a.apiReportError(w, http.StatusBadRequest, fmt.Errorf("failed to parse request: %v", err))
		return
	}

	if !a.RepoAllowlistChecker.IsAllowlisted(baseRepo.FullName, baseRepo.VCSHost.Hostname) {
		a.apiReportError(w, http.StatusForbidden, fmt.Errorf("repo not allowlisted"))
		return
	}

	// Create executor that bridges to existing plan/apply infrastructure
	executor := &apiRemediationExecutor{
		controller: a,
		baseRepo:   baseRepo,
		logger:     a.Logger,
	}

	// Execute remediation
	result, err := a.RemediationService.Remediate(request, executor)
	if err != nil {
		a.apiReportError(w, http.StatusInternalServerError, err)
		return
	}

	// Return result
	response, err := json.Marshal(result)
	if err != nil {
		a.apiReportError(w, http.StatusInternalServerError, err)
		return
	}

	code := http.StatusOK
	if result.Status == models.RemediationStatusFailed {
		code = http.StatusInternalServerError
	}
	a.respond(w, logging.Info, code, "%s", string(response))
}

// apiRemediationExecutor implements drift.RemediationExecutor using the API controller's
// existing plan/apply infrastructure.
type apiRemediationExecutor struct {
	controller *APIController
	baseRepo   models.Repo
	logger     logging.SimpleLogging
}

// ExecutePlan runs a plan for the given project using the API infrastructure.
func (e *apiRemediationExecutor) ExecutePlan(repository, ref, vcsType, projectName, path, workspace string) (string, *models.DriftSummary, error) {
	// Create a minimal API request for the plan
	request := &APIRequest{
		Repository: repository,
		Ref:        ref,
		Type:       vcsType,
	}

	if projectName != "" {
		request.Projects = []string{projectName}
	} else if path != "" || workspace != "" {
		request.Paths = []struct {
			Directory string
			Workspace string
		}{{Directory: path, Workspace: workspace}}
	}

	// Build the command context
	ctx := &command.Context{
		HeadRepo: e.baseRepo,
		Pull: models.PullRequest{
			Num:        0, // Non-PR workflow
			BaseBranch: ref,
			HeadBranch: ref,
			HeadCommit: ref,
			BaseRepo:   e.baseRepo,
		},
		Scope: e.controller.Scope,
		Log:   e.logger,
		API:   true,
	}

	// Setup working directory
	if err := e.controller.apiSetup(ctx, command.Plan); err != nil {
		return "", nil, fmt.Errorf("setup failed: %w", err)
	}

	// Execute plan
	result, err := e.controller.apiPlan(request, ctx)
	if err != nil {
		return "", nil, err
	}

	// Extract output and drift summary
	var output strings.Builder
	var driftSummary *models.DriftSummary

	for _, pr := range result.ProjectResults {
		if pr.Error != nil {
			output.WriteString(fmt.Sprintf("Error: %v\n", pr.Error))
		} else if pr.Failure != "" {
			output.WriteString(fmt.Sprintf("Failure: %s\n", pr.Failure))
		} else if pr.PlanSuccess != nil {
			output.WriteString(pr.PlanSuccess.TerraformOutput)
			// Parse drift from plan output
			summary := models.NewDriftSummaryFromPlanSuccess(pr.PlanSuccess)
			driftSummary = &summary
		}
	}

	if result.HasErrors() {
		return output.String(), driftSummary, fmt.Errorf("plan had errors")
	}

	return output.String(), driftSummary, nil
}

// ExecuteApply runs an apply for the given project using the API infrastructure.
func (e *apiRemediationExecutor) ExecuteApply(repository, ref, vcsType, projectName, path, workspace string) (string, error) {
	// Create a minimal API request for the apply
	request := &APIRequest{
		Repository: repository,
		Ref:        ref,
		Type:       vcsType,
	}

	if projectName != "" {
		request.Projects = []string{projectName}
	} else if path != "" || workspace != "" {
		request.Paths = []struct {
			Directory string
			Workspace string
		}{{Directory: path, Workspace: workspace}}
	}

	// Build the command context
	ctx := &command.Context{
		HeadRepo: e.baseRepo,
		Pull: models.PullRequest{
			Num:        0, // Non-PR workflow
			BaseBranch: ref,
			HeadBranch: ref,
			HeadCommit: ref,
			BaseRepo:   e.baseRepo,
		},
		Scope: e.controller.Scope,
		Log:   e.logger,
		API:   true,
	}

	// Setup working directory
	if err := e.controller.apiSetup(ctx, command.Apply); err != nil {
		return "", fmt.Errorf("setup failed: %w", err)
	}

	// First run plan (required before apply)
	_, err := e.controller.apiPlan(request, ctx)
	if err != nil {
		return "", fmt.Errorf("plan failed: %w", err)
	}

	// Execute apply
	result, err := e.controller.apiApply(request, ctx)
	if err != nil {
		return "", err
	}

	// Extract output
	var output strings.Builder
	for _, pr := range result.ProjectResults {
		if pr.Error != nil {
			output.WriteString(fmt.Sprintf("Error: %v\n", pr.Error))
		} else if pr.Failure != "" {
			output.WriteString(fmt.Sprintf("Failure: %s\n", pr.Failure))
		} else if pr.ApplySuccess != "" {
			output.WriteString(pr.ApplySuccess)
		}
	}

	if result.HasErrors() {
		return output.String(), fmt.Errorf("apply had errors")
	}

	return output.String(), nil
}

// GetRemediationResult handles GET /api/drift/remediate/{id} requests.
// It retrieves a specific remediation result by ID.
// This is an authenticated endpoint that requires the API secret.
func (a *APIController) GetRemediationResult(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Check API is enabled
	if len(a.APISecret) == 0 {
		a.apiReportError(w, http.StatusBadRequest, fmt.Errorf("ignoring request since API is disabled"))
		return
	}

	// Validate the secret token
	secret := r.Header.Get(atlantisTokenHeader)
	if secret != string(a.APISecret) {
		a.apiReportError(w, http.StatusUnauthorized, fmt.Errorf("header %s did not match expected secret", atlantisTokenHeader))
		return
	}

	// Check if remediation service is configured
	if a.RemediationService == nil {
		a.apiReportError(w, http.StatusServiceUnavailable, fmt.Errorf("drift remediation is not enabled"))
		return
	}

	// Get the ID from query parameter (or path, depending on router)
	id := r.URL.Query().Get("id")
	if id == "" {
		// Try to extract from path for routers that support path parameters
		// Path format: /api/drift/remediate/{id}
		path := r.URL.Path
		if strings.HasPrefix(path, "/api/drift/remediate/") {
			id = strings.TrimPrefix(path, "/api/drift/remediate/")
		}
	}

	if id == "" {
		a.apiReportError(w, http.StatusBadRequest, fmt.Errorf("id parameter is required"))
		return
	}

	// Get the result
	result, err := a.RemediationService.GetResult(id)
	if err != nil {
		a.apiReportError(w, http.StatusNotFound, err)
		return
	}

	// Return result
	response, err := json.Marshal(result)
	if err != nil {
		a.apiReportError(w, http.StatusInternalServerError, err)
		return
	}

	a.respond(w, logging.Info, http.StatusOK, "%s", string(response))
}

// ListRemediationResults handles GET /api/drift/remediate requests.
// It lists remediation results for a repository.
// Query parameters:
//   - repository: required, the full repository name (owner/repo)
//   - limit: optional, maximum number of results to return (default: 10)
//
// This is an authenticated endpoint that requires the API secret.
func (a *APIController) ListRemediationResults(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Check API is enabled
	if len(a.APISecret) == 0 {
		a.apiReportError(w, http.StatusBadRequest, fmt.Errorf("ignoring request since API is disabled"))
		return
	}

	// Validate the secret token
	secret := r.Header.Get(atlantisTokenHeader)
	if secret != string(a.APISecret) {
		a.apiReportError(w, http.StatusUnauthorized, fmt.Errorf("header %s did not match expected secret", atlantisTokenHeader))
		return
	}

	// Check if remediation service is configured
	if a.RemediationService == nil {
		a.apiReportError(w, http.StatusServiceUnavailable, fmt.Errorf("drift remediation is not enabled"))
		return
	}

	// Get query parameters
	repository := r.URL.Query().Get("repository")
	if repository == "" {
		a.apiReportError(w, http.StatusBadRequest, fmt.Errorf("repository parameter is required"))
		return
	}

	// Parse limit (default: 10)
	limit := 10
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if _, err := fmt.Sscanf(limitStr, "%d", &limit); err != nil {
			a.apiReportError(w, http.StatusBadRequest, fmt.Errorf("invalid limit parameter: %v", err))
			return
		}
		if limit <= 0 {
			limit = 10
		}
		if limit > 100 {
			limit = 100
		}
	}

	// Get results
	results, err := a.RemediationService.ListResults(repository, limit)
	if err != nil {
		a.apiReportError(w, http.StatusInternalServerError, err)
		return
	}

	// Return results
	responseData := struct {
		Repository string                      `json:"repository"`
		Count      int                         `json:"count"`
		Results    []*models.RemediationResult `json:"results"`
	}{
		Repository: repository,
		Count:      len(results),
		Results:    results,
	}

	response, err := json.Marshal(responseData)
	if err != nil {
		a.apiReportError(w, http.StatusInternalServerError, err)
		return
	}

	a.respond(w, logging.Info, http.StatusOK, "%s", string(response))
}

// DetectDrift handles POST /api/drift/detect requests.
// It triggers drift detection by running plans for the specified projects.
// This is an authenticated endpoint that requires the API secret.
func (a *APIController) DetectDrift(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Check API is enabled
	if len(a.APISecret) == 0 {
		a.apiReportError(w, http.StatusBadRequest, fmt.Errorf("ignoring request since API is disabled"))
		return
	}

	// Validate the secret token
	secret := r.Header.Get(atlantisTokenHeader)
	if secret != string(a.APISecret) {
		a.apiReportError(w, http.StatusUnauthorized, fmt.Errorf("header %s did not match expected secret", atlantisTokenHeader))
		return
	}

	// Check if drift storage is configured
	if a.DriftStorage == nil {
		a.apiReportError(w, http.StatusServiceUnavailable, fmt.Errorf("drift detection is not enabled"))
		return
	}

	// Parse the JSON payload
	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		a.apiReportError(w, http.StatusBadRequest, fmt.Errorf("failed to read request: %v", err))
		return
	}

	var request models.DriftDetectionRequest
	if err = json.Unmarshal(bytes, &request); err != nil {
		a.apiReportError(w, http.StatusBadRequest, fmt.Errorf("failed to parse request: %v", err))
		return
	}

	// Validate required fields
	if request.Repository == "" {
		a.apiReportError(w, http.StatusBadRequest, fmt.Errorf("repository is required"))
		return
	}
	if request.Ref == "" {
		a.apiReportError(w, http.StatusBadRequest, fmt.Errorf("ref is required"))
		return
	}
	if request.Type == "" {
		a.apiReportError(w, http.StatusBadRequest, fmt.Errorf("type is required"))
		return
	}

	// Check if the repo is allowlisted
	VCSHostType, err := models.NewVCSHostType(request.Type)
	if err != nil {
		a.apiReportError(w, http.StatusBadRequest, err)
		return
	}
	cloneURL, err := a.VCSClient.GetCloneURL(a.Logger, VCSHostType, request.Repository)
	if err != nil {
		a.apiReportError(w, http.StatusInternalServerError, err)
		return
	}

	baseRepo, err := a.Parser.ParseAPIPlanRequest(VCSHostType, request.Repository, cloneURL)
	if err != nil {
		a.apiReportError(w, http.StatusBadRequest, fmt.Errorf("failed to parse request: %v", err))
		return
	}

	if !a.RepoAllowlistChecker.IsAllowlisted(baseRepo.FullName, baseRepo.VCSHost.Hostname) {
		a.apiReportError(w, http.StatusForbidden, fmt.Errorf("repo not allowlisted"))
		return
	}

	// Build API request for plan
	apiRequest := &APIRequest{
		Repository: request.Repository,
		Ref:        request.Ref,
		Type:       request.Type,
	}

	if len(request.Projects) > 0 {
		apiRequest.Projects = request.Projects
	}
	if len(request.Paths) > 0 {
		for _, p := range request.Paths {
			apiRequest.Paths = append(apiRequest.Paths, struct {
				Directory string
				Workspace string
			}{Directory: p.Directory, Workspace: p.Workspace})
		}
	}

	// Build the command context
	ctx := &command.Context{
		HeadRepo: baseRepo,
		Pull: models.PullRequest{
			Num:        0, // Non-PR workflow
			BaseBranch: request.Ref,
			HeadBranch: request.Ref,
			HeadCommit: request.Ref,
			BaseRepo:   baseRepo,
		},
		Scope: a.Scope,
		Log:   a.Logger,
		API:   true,
	}

	// Setup and run plan
	if err := a.apiSetup(ctx, command.Plan); err != nil {
		a.apiReportError(w, http.StatusInternalServerError, fmt.Errorf("setup failed: %v", err))
		return
	}

	result, err := a.apiPlan(apiRequest, ctx)
	if err != nil {
		a.apiReportError(w, http.StatusInternalServerError, err)
		return
	}
	defer a.Locker.UnlockByPull(ctx.HeadRepo.FullName, ctx.Pull.Num) // nolint: errcheck

	// Process results and store drift data
	detectionResult := models.NewDriftDetectionResult(request.Repository)

	for _, pr := range result.ProjectResults {
		projectDrift := models.ProjectDrift{
			ProjectName: pr.ProjectName,
			Path:        pr.RepoRelDir,
			Workspace:   pr.Workspace,
			Ref:         request.Ref,
			LastChecked: time.Now(),
		}

		if pr.Error != nil {
			projectDrift.Drift = models.DriftSummary{
				HasDrift: false,
				Summary:  fmt.Sprintf("Error: %v", pr.Error),
			}
		} else if pr.Failure != "" {
			projectDrift.Drift = models.DriftSummary{
				HasDrift: false,
				Summary:  fmt.Sprintf("Failure: %s", pr.Failure),
			}
		} else if pr.PlanSuccess != nil {
			projectDrift.Drift = models.NewDriftSummaryFromPlanSuccess(pr.PlanSuccess)
		}

		// Store drift data
		if err := a.DriftStorage.Store(request.Repository, projectDrift); err != nil {
			a.Logger.Warn("failed to store drift data: %v", err)
		}

		detectionResult.AddProject(projectDrift)
	}

	// Return detection result
	response, err := json.Marshal(detectionResult)
	if err != nil {
		a.apiReportError(w, http.StatusInternalServerError, err)
		return
	}

	code := http.StatusOK
	if result.HasErrors() {
		code = http.StatusInternalServerError
	}
	a.respond(w, logging.Info, code, "%s", string(response))
}
