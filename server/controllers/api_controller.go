// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/runatlantis/atlantis/server/core/drift"
	"github.com/runatlantis/atlantis/server/core/locking"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/events/webhooks"
	"github.com/runatlantis/atlantis/server/logging"
	tally "github.com/uber-go/tally/v4"
)

const atlantisTokenHeader = "X-Atlantis-Token"

var nonPRPullCounter atomic.Int64

type APIController struct {
	APISecret                      []byte
	Locker                         locking.Locker `validate:"required"`
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
	// PullReqStatusFetcher is optional. When set and the API request supplies a
	// PR number, it is used to populate command.Context.PullRequestStatus so
	// apply requirements like 'mergeable' and 'approved' evaluate against real
	// VCS state instead of always failing.
	PullReqStatusFetcher vcs.PullReqStatusFetcher
	// DriftWebhookSender sends webhook notifications when drift is detected.
	// Nil when no drift webhooks are configured.
	DriftWebhookSender *webhooks.DriftWebhookSender
	// SilenceVCSStatusNoProjects is whether API should set commit status if no projects are found
	SilenceVCSStatusNoProjects bool

	// apiMiddleware provides common authentication and response utilities.
	// Initialized lazily via getAPIMiddleware() with sync.Once for thread safety.
	apiMiddleware     *APIMiddleware
	apiMiddlewareOnce sync.Once
}

// getAPIMiddleware returns the APIMiddleware, initializing it lazily with sync.Once.
func (a *APIController) getAPIMiddleware() *APIMiddleware {
	a.apiMiddlewareOnce.Do(func() {
		a.apiMiddleware = NewAPIMiddleware(a.APISecret, a.Logger)
	})
	return a.apiMiddleware
}

func nextNonPRPullNum() int {
	return -int((time.Now().UnixNano() & 0x3fffffff) + nonPRPullCounter.Add(1))
}

func (a *APIController) cleanupNonPRWorkingDir(ctx *command.Context) {
	if ctx.Pull.Num > 0 {
		return
	}
	if err := a.WorkingDir.Delete(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull); err != nil {
		ctx.Log.Warn("cleaning up API working directory: %s", err)
	}
}

type APIRequest struct {
	Repository string `validate:"required"`
	Ref        string `validate:"required"`
	Type       string `validate:"required"`
	PR         int
	Projects   []string
	Paths      []APIRequestPath
	// DiscoverProjects enables all-project discovery when no projects or paths
	// are specified. Only drift detection and remediation set this.
	DiscoverProjects bool `json:"-"`
}

type APIRequestPath struct {
	ProjectName string `json:"project_name,omitempty"`
	Directory   string `json:"directory"`
	Workspace   string `json:"workspace,omitempty"`
}

func (a *APIRequest) getCommands(ctx *command.Context, cmdName command.Name, cmdBuilder func(*command.Context, *events.CommentCommand) ([]command.ProjectContext, error)) ([]command.ProjectContext, []*events.CommentCommand, error) {
	cc := make([]*events.CommentCommand, 0)

	for _, project := range a.Projects {
		cc = append(cc, &events.CommentCommand{
			Name:        cmdName,
			ProjectName: project,
		})
	}
	for _, path := range a.Paths {
		cc = append(cc, &events.CommentCommand{
			Name:        cmdName,
			ProjectName: path.ProjectName,
			RepoRelDir:  strings.TrimRight(path.Directory, "/"),
			Workspace:   path.Workspace,
		})
	}

	// When no specific projects or paths are provided and DiscoverProjects
	// is set, enumerate all projects without consulting PR modified-file APIs.
	if len(cc) == 0 && a.DiscoverProjects {
		cc = append(cc, &events.CommentCommand{
			Name:                cmdName,
			DiscoverAllProjects: true,
		})
	}

	cmds := make([]command.ProjectContext, 0)
	keptCommentCommands := make([]*events.CommentCommand, 0)
	ignoredCommands := 0
	nonIgnoredCommands := 0
	for _, commentCommand := range cc {
		projectCmds, err := cmdBuilder(ctx, commentCommand)
		if err != nil {
			if events.IsIgnoredTargetedDir(err) {
				ignoredCommands++
				continue
			}
			return nil, nil, fmt.Errorf("failed to build command: %w", err)
		}
		nonIgnoredCommands++
		for _, projectCmd := range projectCmds {
			cmds = append(cmds, projectCmd)
			keptCommentCommands = append(keptCommentCommands, commentCommand)
		}
	}
	if ignoredCommands > 0 && nonIgnoredCommands == 0 {
		return nil, nil, events.ErrIgnoredTargetedDir
	}

	sortCommandPairsByExecutionOrder(cmds, keptCommentCommands)
	return cmds, keptCommentCommands, nil
}

func sortCommandPairsByExecutionOrder(cmds []command.ProjectContext, commentCommands []*events.CommentCommand) {
	if len(cmds) != len(commentCommands) {
		return
	}
	type commandPair struct {
		cmd            command.ProjectContext
		commentCommand *events.CommentCommand
		index          int
	}
	pairs := make([]commandPair, len(cmds))
	for i := range cmds {
		pairs[i] = commandPair{cmd: cmds[i], commentCommand: commentCommands[i], index: i}
	}
	sort.SliceStable(pairs, func(i, j int) bool {
		if pairs[i].cmd.ExecutionOrderGroup == pairs[j].cmd.ExecutionOrderGroup {
			return pairs[i].index < pairs[j].index
		}
		return pairs[i].cmd.ExecutionOrderGroup < pairs[j].cmd.ExecutionOrderGroup
	})
	for i := range pairs {
		cmds[i] = pairs[i].cmd
		commentCommands[i] = pairs[i].commentCommand
	}
}

// apiHandleParseError maps HTTP status codes from apiParseAndValidate to API responses.
func (a *APIController) apiHandleParseError(w http.ResponseWriter, r *http.Request, responder *APIResponder, code int, err error) {
	switch code {
	case http.StatusBadRequest:
		// Validation or parsing errors
		responder.ValidationFailed(w, r, err.Error())
	case http.StatusUnauthorized:
		responder.Unauthorized(w, r, err.Error())
	case http.StatusForbidden:
		responder.Forbidden(w, r, err.Error())
	case http.StatusServiceUnavailable:
		responder.ServiceUnavailable(w, r, err.Error())
	default:
		responder.InternalError(w, r, err)
	}
}

func (a *APIController) Plan(w http.ResponseWriter, r *http.Request) {
	middleware := a.getAPIMiddleware()
	responder := middleware.Responder

	request, ctx, code, err := a.apiParseAndValidate(r)
	if err != nil {
		a.apiHandleParseError(w, r, responder, code, err)
		return
	}

	err = a.apiSetup(ctx, command.Plan)
	if err != nil {
		responder.InternalError(w, r, err)
		return
	}
	defer a.cleanupNonPRWorkingDir(ctx)

	result, err := a.apiPlan(request, ctx)
	if err != nil {
		responder.InternalError(w, r, err)
		return
	}
	if !ctx.CommandSkipped {
		defer a.Locker.UnlockByPull(ctx.HeadRepo.FullName, ctx.Pull.Num) // nolint: errcheck
	}

	// Convert to API response format
	apiResult := NewCommandResultAPI(result, command.Plan.String())

	// Return per-project statuses inside the success envelope. Project-level
	// failures use 207 Multi-Status so clients can read per-project results
	// without needing to parse an error envelope.
	statusCode := http.StatusOK
	if result.HasErrors() {
		statusCode = http.StatusMultiStatus
	}
	responder.Success(w, r, statusCode, apiResult)
}

func (a *APIController) Apply(w http.ResponseWriter, r *http.Request) {
	middleware := a.getAPIMiddleware()
	responder := middleware.Responder

	request, ctx, code, err := a.apiParseAndValidate(r)
	if err != nil {
		a.apiHandleParseError(w, r, responder, code, err)
		return
	}

	err = a.apiSetup(ctx, command.Apply)
	if err != nil {
		responder.InternalError(w, r, err)
		return
	}
	defer a.cleanupNonPRWorkingDir(ctx)

	// We must first make the plan for all projects
	result, err := a.apiPlan(request, ctx)
	if err != nil {
		responder.InternalError(w, r, err)
		return
	}
	if ctx.CommandSkipped {
		responder.Success(w, r, http.StatusOK, NewCommandResultAPI(result, command.Apply.String()))
		return
	}
	defer a.Locker.UnlockByPull(ctx.HeadRepo.FullName, ctx.Pull.Num) // nolint: errcheck

	// The API apply endpoint runs plan first. Refresh PR status afterward so
	// apply requirements evaluate the VCS state the plan phase just produced.
	a.populatePullRequestStatus(ctx)
	seedPullStatusFromPlanResult(ctx, result)

	// We can now prepare and run the apply step
	result, err = a.apiApply(request, ctx)
	if err != nil {
		responder.InternalError(w, r, err)
		return
	}

	// Convert to API response format
	apiResult := NewCommandResultAPI(result, command.Apply.String())

	// Return per-project statuses inside the success envelope. Project-level
	// failures use 207 Multi-Status so clients can read per-project results
	// without needing to parse an error envelope.
	statusCode := http.StatusOK
	if result.HasErrors() {
		statusCode = http.StatusMultiStatus
	}
	responder.Success(w, r, statusCode, apiResult)
}

// LockDetail is deprecated - use LockDetailAPI instead.
// Kept for backwards compatibility during migration.
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

// ListLocksResult is deprecated - use ListLocksResultAPI instead.
// Kept for backwards compatibility during migration.
type ListLocksResult struct {
	Locks []LockDetail
}

func (a *APIController) ListLocks(w http.ResponseWriter, r *http.Request) {
	middleware := a.getAPIMiddleware()
	responder := middleware.Responder

	locks, err := a.Locker.List()
	if err != nil {
		responder.InternalError(w, r, err)
		return
	}

	// Convert to API response format
	apiResult := NewListLocksResultAPI(locks)

	responder.Success(w, r, http.StatusOK, apiResult)
}

// DriftStatus returns cached drift detection results for a repository.
// This is an authenticated endpoint that requires the API secret.
// Query parameters:
//   - repository: required, the full repository name (owner/repo)
//   - type: required, the VCS provider type
//   - project: optional, filter by project name
//   - path: optional, filter by repository-relative project path
//   - workspace: optional, filter by workspace
//   - ref: optional, filter by git ref
func (a *APIController) DriftStatus(w http.ResponseWriter, r *http.Request) {
	middleware := a.getAPIMiddleware()
	responder := middleware.Responder

	if !middleware.RequireAuth(w, r) {
		return
	}

	// Check if drift storage is configured
	if a.DriftStorage == nil {
		responder.ServiceUnavailable(w, r, "drift detection is not enabled")
		return
	}

	// Get query parameters
	repository := r.URL.Query().Get("repository")
	if repository == "" {
		responder.ValidationFailed(w, r, "missing required parameter",
			ValidationError{Field: "repository", Message: "repository parameter is required"})
		return
	}
	vcsType := r.URL.Query().Get("type")
	if vcsType == "" {
		responder.ValidationFailed(w, r, "missing required parameter",
			ValidationError{Field: "type", Message: "type parameter is required"})
		return
	}
	VCSHostType, err := models.NewVCSHostType(vcsType)
	if err != nil {
		responder.ValidationFailed(w, r, "invalid VCS type",
			ValidationError{Field: "type", Message: err.Error()})
		return
	}
	cloneURL, err := a.VCSClient.GetCloneURL(a.Logger, VCSHostType, repository)
	if err != nil {
		responder.InternalError(w, r, fmt.Errorf("failed to get clone URL: %w", err))
		return
	}
	baseRepo, err := a.Parser.ParseAPIPlanRequest(VCSHostType, repository, cloneURL)
	if err != nil {
		responder.ValidationFailed(w, r, fmt.Sprintf("failed to parse repository: %v", err))
		return
	}
	if !a.RepoAllowlistChecker.IsAllowlisted(baseRepo.FullName, baseRepo.VCSHost.Hostname) {
		responder.Forbidden(w, r, "repository is not in the allowlist")
		return
	}

	opts := drift.GetOptions{
		ProjectName: r.URL.Query().Get("project"),
		Path:        r.URL.Query().Get("path"),
		Workspace:   r.URL.Query().Get("workspace"),
		Ref:         r.URL.Query().Get("ref"),
	}

	// Retrieve drift results from storage
	drifts, err := a.DriftStorage.Get(baseRepo.ID(), opts)
	if err != nil {
		responder.InternalError(w, r, err)
		return
	}

	// Build response using API DTO
	internalResult := models.NewDriftStatusResponse(repository, drifts)
	apiResult := NewDriftStatusAPI(internalResult)

	responder.Success(w, r, http.StatusOK, apiResult)
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
	repoDir, err := a.WorkingDir.Clone(ctx.Log, headRepo, pull, events.DefaultWorkspace)
	if err != nil {
		return err
	}

	return resolveNonPRHeadCommit(ctx, repoDir)
}

func resolveNonPRHeadCommit(ctx *command.Context, repoDir string) error {
	if ctx.Pull.Num > 0 || repoDir == "" {
		return nil
	}
	if _, err := os.Stat(filepath.Join(repoDir, ".git")); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("checking API checkout git metadata: %w", err)
	}

	cmd := exec.Command("git", "rev-parse", "HEAD") // nolint: gosec
	cmd.Dir = repoDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("resolving checked out API ref: %s: %w", strings.TrimSpace(string(output)), err)
	}
	headCommit := strings.TrimSpace(string(output))
	if headCommit == "" {
		return fmt.Errorf("resolving checked out API ref: empty commit")
	}
	ctx.Pull.HeadCommit = headCommit
	return nil
}

func (a *APIController) apiPlan(request *APIRequest, ctx *command.Context) (*command.Result, error) {
	cmds, cc, err := request.getCommands(ctx, command.Plan, a.ProjectCommandBuilder.BuildPlanCommands)
	if events.IsIgnoredTargetedDir(err) {
		ctx.CommandSkipped = true
		return &command.Result{ProjectResults: []command.ProjectResult{}}, nil
	}
	if err != nil {
		return nil, err
	}

	if len(cmds) == 0 {
		ctx.Log.Info("determined there was no project to run plan in")
		// When silence is enabled and no projects are found, don't set any VCS status
		if !a.SilenceVCSStatusNoProjects {
			ctx.Log.Debug("setting VCS status to success with no projects found")
			if err := a.CommitStatusUpdater.UpdateCombinedCount(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.SuccessCommitStatus, command.Plan, models.ProjectCounts{}); err != nil {
				ctx.Log.Warn("unable to update plan status: %s", err)
			}
			if err := a.CommitStatusUpdater.UpdateCombinedCount(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.SuccessCommitStatus, command.PolicyCheck, models.ProjectCounts{}); err != nil {
				ctx.Log.Warn("unable to update policy check status: %s", err)
			}
			if err := a.CommitStatusUpdater.UpdateCombinedCount(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.SuccessCommitStatus, command.Apply, models.ProjectCounts{}); err != nil {
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
	cmds, cc, err := request.getCommands(ctx, command.Apply, a.ProjectCommandBuilder.BuildApplyCommands)
	if events.IsIgnoredTargetedDir(err) {
		ctx.CommandSkipped = true
		return &command.Result{ProjectResults: []command.ProjectResult{}}, nil
	}
	if err != nil {
		return nil, err
	}

	if len(cmds) == 0 {
		ctx.Log.Info("determined there was no project to run apply in")
		// When silence is enabled and no projects are found, don't set any VCS status
		if !a.SilenceVCSStatusNoProjects {
			ctx.Log.Debug("setting VCS status to success with no projects found")
			if err := a.CommitStatusUpdater.UpdateCombinedCount(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.SuccessCommitStatus, command.Plan, models.ProjectCounts{}); err != nil {
				ctx.Log.Warn("unable to update plan status: %s", err)
			}
			if err := a.CommitStatusUpdater.UpdateCombinedCount(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.SuccessCommitStatus, command.PolicyCheck, models.ProjectCounts{}); err != nil {
				ctx.Log.Warn("unable to update policy check status: %s", err)
			}
			if err := a.CommitStatusUpdater.UpdateCombinedCount(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.SuccessCommitStatus, command.Apply, models.ProjectCounts{}); err != nil {
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
		updatePullStatusFromProjectResult(ctx, res)

		a.PostWorkflowHooksCommandRunner.RunPostHooks(ctx, cc[i]) // nolint: errcheck
	}
	return &command.Result{ProjectResults: projectResults}, nil
}

func updatePullStatusFromProjectResult(ctx *command.Context, result command.ProjectResult) {
	if ctx.PullStatus == nil {
		return
	}

	for projectIdx := range ctx.PullStatus.Projects {
		project := &ctx.PullStatus.Projects[projectIdx]
		if result.Workspace == project.Workspace &&
			result.RepoRelDir == project.RepoRelDir &&
			result.ProjectName == project.ProjectName {
			project.Status = result.PlanStatus()
			return
		}
	}
}

func seedPullStatusFromPlanResult(ctx *command.Context, result *command.Result) {
	if result == nil {
		return
	}
	if ctx.PullStatus == nil {
		ctx.PullStatus = &models.PullStatus{Pull: ctx.Pull}
	}
	for _, projectResult := range result.ProjectResults {
		if projectResult.Command != command.Plan && projectResult.Command != command.PolicyCheck && projectResult.Command != command.ApprovePolicies {
			continue
		}
		upsertProjectStatus(ctx.PullStatus, models.ProjectStatus{
			Workspace:    projectResult.Workspace,
			RepoRelDir:   projectResult.RepoRelDir,
			ProjectName:  projectResult.ProjectName,
			PolicyStatus: projectResult.PolicyStatus(),
			Status:       projectResult.PlanStatus(),
		})
	}
}

func upsertProjectStatus(pullStatus *models.PullStatus, status models.ProjectStatus) {
	for idx := range pullStatus.Projects {
		project := &pullStatus.Projects[idx]
		if status.Workspace == project.Workspace &&
			status.RepoRelDir == project.RepoRelDir &&
			status.ProjectName == project.ProjectName {
			project.Status = status.Status
			project.PolicyStatus = mergePolicyStatuses(project.PolicyStatus, status.PolicyStatus)
			return
		}
	}
	pullStatus.Projects = append(pullStatus.Projects, status)
}

func mergePolicyStatuses(existing []models.PolicySetStatus, incoming []models.PolicySetStatus) []models.PolicySetStatus {
	if len(incoming) == 0 {
		return existing
	}
	if len(existing) == 0 {
		return incoming
	}
	for _, newPolicySet := range incoming {
		updated := false
		for idx, oldPolicySet := range existing {
			if oldPolicySet.PolicySetName == newPolicySet.PolicySetName {
				existing[idx] = newPolicySet
				updated = true
				break
			}
		}
		if !updated {
			existing = append(existing, newPolicySet)
		}
	}
	return existing
}

func (a *APIController) apiParseAndValidate(r *http.Request) (*APIRequest, *command.Context, int, error) {
	if len(a.APISecret) == 0 {
		return nil, nil, http.StatusServiceUnavailable, fmt.Errorf("ignoring request since API is disabled")
	}

	// Validate the secret token using constant-time comparison to prevent timing attacks
	secret := r.Header.Get(atlantisTokenHeader)
	if subtle.ConstantTimeCompare([]byte(secret), a.APISecret) != 1 {
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

	pullNum := request.PR
	if pullNum <= 0 {
		pullNum = nextNonPRPullNum()
	}
	pull := models.PullRequest{
		Num:        pullNum,
		BaseBranch: request.Ref,
		HeadBranch: request.Ref,
		HeadCommit: request.Ref,
		BaseRepo:   baseRepo,
	}
	ctx := &command.Context{
		HeadRepo: baseRepo,
		Pull:     pull,
		Scope:    a.Scope,
		Log:      a.Logger,
		API:      true,
	}
	a.populatePullRequestStatus(ctx)
	return &request, ctx, http.StatusOK, nil
}

func (a *APIController) populatePullRequestStatus(ctx *command.Context) {
	if ctx.Pull.Num <= 0 || a.PullReqStatusFetcher == nil {
		return
	}

	status, err := a.PullReqStatusFetcher.FetchPullStatus(ctx.Log, ctx.Pull)
	if err != nil {
		ctx.PullRequestStatus = models.PullReqStatus{}
		ctx.Log.Warn("unable to get pull request status: %s. Continuing with mergeable and approved assumed false", err)
		return
	}

	ctx.PullRequestStatus = status
}

// Remediate handles POST /api/drift/remediate requests.
// It executes drift remediation (plan or apply) for the specified projects.
// This is an authenticated endpoint that requires the API secret.
func (a *APIController) Remediate(w http.ResponseWriter, r *http.Request) {
	middleware := a.getAPIMiddleware()
	responder := middleware.Responder

	// Authenticate
	if !middleware.RequireAuth(w, r) {
		return
	}

	// Check if remediation service is configured
	if a.RemediationService == nil {
		responder.ServiceUnavailable(w, r, "drift remediation is not enabled")
		return
	}

	// Parse the JSON payload
	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		responder.ValidationFailed(w, r, "failed to read request body")
		return
	}

	var request models.RemediationRequest
	if err = json.Unmarshal(bytes, &request); err != nil {
		responder.ValidationFailed(w, r, fmt.Sprintf("failed to parse JSON: %v", err))
		return
	}

	// Validate required fields using the model's Validate method
	if validationErrors := request.Validate(); len(validationErrors) > 0 {
		fields := make([]ValidationError, 0, len(validationErrors))
		for _, fe := range validationErrors {
			fields = append(fields, ValidationError{Field: fe.Field, Message: fe.Message})
		}
		responder.ValidationFailed(w, r, "validation failed", fields...)
		return
	}

	// Apply default values
	request.ApplyDefaults()

	// Check if the repo is allowlisted
	VCSHostType, err := models.NewVCSHostType(request.Type)
	if err != nil {
		responder.ValidationFailed(w, r, "invalid VCS type",
			ValidationError{Field: "type", Message: err.Error()})
		return
	}
	cloneURL, err := a.VCSClient.GetCloneURL(a.Logger, VCSHostType, request.Repository)
	if err != nil {
		responder.InternalError(w, r, fmt.Errorf("failed to get clone URL: %w", err))
		return
	}

	baseRepo, err := a.Parser.ParseAPIPlanRequest(VCSHostType, request.Repository, cloneURL)
	if err != nil {
		responder.ValidationFailed(w, r, fmt.Sprintf("failed to parse repository: %v", err))
		return
	}

	if !a.RepoAllowlistChecker.IsAllowlisted(baseRepo.FullName, baseRepo.VCSHost.Hostname) {
		responder.Forbidden(w, r, "repository is not in the allowlist")
		return
	}
	request.StorageRepository = baseRepo.ID()

	// Create executor that bridges to existing plan/apply infrastructure
	executor := &apiRemediationExecutor{
		controller: a,
		baseRepo:   baseRepo,
		logger:     a.Logger,
	}

	// Execute remediation
	result, err := a.RemediationService.Remediate(request, executor)
	if err != nil {
		responder.InternalError(w, r, err)
		return
	}

	// Convert to API DTO and return
	apiResult := NewRemediationResultAPI(result)

	code := http.StatusOK
	switch result.Status {
	case models.RemediationStatusFailed, models.RemediationStatusPartial:
		code = http.StatusMultiStatus // 207 - some projects succeeded, some failed
	}
	responder.Success(w, r, code, apiResult)
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
		Repository:       repository,
		Ref:              ref,
		Type:             vcsType,
		DiscoverProjects: true,
	}

	if projectName != "" || path != "" || workspace != "" {
		request.Paths = []APIRequestPath{{
			ProjectName: projectName,
			Directory:   path,
			Workspace:   workspace,
		}}
	}

	// Build the command context
	ctx := &command.Context{
		HeadRepo: e.baseRepo,
		Pull: models.PullRequest{
			Num:        nextNonPRPullNum(), // Synthetic non-PR workflow ID.
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
	defer e.controller.cleanupNonPRWorkingDir(ctx)

	// Run pre-workflow hooks before project discovery
	preHookCmd := &events.CommentCommand{Name: command.Plan}
	if err := e.controller.PreWorkflowHooksCommandRunner.RunPreHooks(ctx, preHookCmd); err != nil {
		if e.controller.FailOnPreWorkflowHookError {
			return "", nil, fmt.Errorf("pre-workflow hook failed: %w", err)
		}
		e.logger.Warn("pre-workflow hook error (continuing): %v", err)
	}

	// Execute plan
	result, err := e.controller.apiPlan(request, ctx)
	if err != nil {
		return "", nil, err
	}
	defer e.controller.Locker.UnlockByPull(ctx.HeadRepo.FullName, ctx.Pull.Num) // nolint: errcheck

	output, driftSummary := planRemediationOutput(result)

	if result.HasErrors() {
		return output.String(), driftSummary, fmt.Errorf("plan had errors")
	}

	return output.String(), driftSummary, nil
}

// ExecuteApplyProjects runs a pre-apply plan and apply for all projects in one
// API context so dependency checks and execution order can see sibling project
// statuses.
func (e *apiRemediationExecutor) ExecuteApplyProjects(repository, ref, vcsType string, projects []models.ProjectDrift) ([]models.ProjectRemediationResult, error) {
	request := &APIRequest{
		Repository:       repository,
		Ref:              ref,
		Type:             vcsType,
		DiscoverProjects: len(projects) == 0,
	}
	for _, project := range projects {
		request.Paths = append(request.Paths, APIRequestPath{
			ProjectName: project.ProjectName,
			Directory:   project.Path,
			Workspace:   project.Workspace,
		})
	}

	ctx := &command.Context{
		HeadRepo: e.baseRepo,
		Pull: models.PullRequest{
			Num:        nextNonPRPullNum(), // Synthetic non-PR workflow ID.
			BaseBranch: ref,
			HeadBranch: ref,
			HeadCommit: ref,
			BaseRepo:   e.baseRepo,
		},
		Scope: e.controller.Scope,
		Log:   e.logger,
		API:   true,
	}

	if err := e.controller.apiSetup(ctx, command.Apply); err != nil {
		return nil, fmt.Errorf("setup failed: %w", err)
	}
	defer e.controller.cleanupNonPRWorkingDir(ctx)

	preHookCmd := &events.CommentCommand{Name: command.Plan}
	if err := e.controller.PreWorkflowHooksCommandRunner.RunPreHooks(ctx, preHookCmd); err != nil {
		if e.controller.FailOnPreWorkflowHookError {
			return nil, fmt.Errorf("pre-workflow hook failed: %w", err)
		}
		e.logger.Warn("pre-workflow hook error (continuing): %v", err)
	}

	planResult, err := e.controller.apiPlan(request, ctx)
	if err != nil {
		return nil, fmt.Errorf("plan failed: %w", err)
	}
	defer e.controller.Locker.UnlockByPull(ctx.HeadRepo.FullName, ctx.Pull.Num) // nolint: errcheck

	remediationResults := projectRemediationResultsFromPlan(projects, planResult)
	if planResult.HasErrors() {
		markRunningRemediationResultsFailed(remediationResults, "apply skipped because pre-apply plan failed")
		return remediationResults, fmt.Errorf("plan had errors")
	}
	seedPullStatusFromPlanResult(ctx, planResult)

	applyResult, err := e.controller.apiApply(request, ctx)
	if err != nil {
		return remediationResults, err
	}
	remediationResults = mergeApplyRemediationResults(remediationResults, applyResult)
	if applyResult.HasErrors() {
		return remediationResults, fmt.Errorf("apply had errors")
	}

	return remediationResults, nil
}

// ExecuteApply runs an apply for the given project using the API infrastructure.
func (e *apiRemediationExecutor) ExecuteApply(repository, ref, vcsType, projectName, path, workspace string) (string, error) {
	// Create a minimal API request for the apply
	request := &APIRequest{
		Repository:       repository,
		Ref:              ref,
		Type:             vcsType,
		DiscoverProjects: true,
	}

	if projectName != "" || path != "" || workspace != "" {
		request.Paths = []APIRequestPath{{
			ProjectName: projectName,
			Directory:   path,
			Workspace:   workspace,
		}}
	}

	// Build the command context
	ctx := &command.Context{
		HeadRepo: e.baseRepo,
		Pull: models.PullRequest{
			Num:        nextNonPRPullNum(), // Synthetic non-PR workflow ID.
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
	defer e.controller.cleanupNonPRWorkingDir(ctx)

	// Run plan-scoped pre-workflow hooks before project discovery for the
	// pre-apply plan.
	preHookCmd := &events.CommentCommand{Name: command.Plan}
	if err := e.controller.PreWorkflowHooksCommandRunner.RunPreHooks(ctx, preHookCmd); err != nil {
		if e.controller.FailOnPreWorkflowHookError {
			return "", fmt.Errorf("pre-workflow hook failed: %w", err)
		}
		e.logger.Warn("pre-workflow hook error (continuing): %v", err)
	}

	// First run plan (required before apply)
	planResult, err := e.controller.apiPlan(request, ctx)
	if err != nil {
		return "", fmt.Errorf("plan failed: %w", err)
	}
	defer e.controller.Locker.UnlockByPull(ctx.HeadRepo.FullName, ctx.Pull.Num) // nolint: errcheck
	if planResult.HasErrors() {
		output, _ := planRemediationOutput(planResult)
		return output.String(), fmt.Errorf("plan had errors")
	}
	seedPullStatusFromPlanResult(ctx, planResult)

	// Execute apply
	result, err := e.controller.apiApply(request, ctx)
	if err != nil {
		return "", err
	}

	output := applyRemediationOutput(result)

	if result.HasErrors() {
		return output.String(), fmt.Errorf("apply had errors")
	}

	return output.String(), nil
}

func planRemediationOutput(result *command.Result) (strings.Builder, *models.DriftSummary) {
	var output strings.Builder
	var driftSummary *models.DriftSummary
	for _, pr := range result.ProjectResults {
		if pr.Error != nil {
			fmt.Fprintf(&output, "Error: %v\n", pr.Error)
		} else if pr.Failure != "" {
			fmt.Fprintf(&output, "Failure: %s\n", pr.Failure)
		} else if pr.PlanSuccess != nil {
			output.WriteString(pr.PlanSuccess.TerraformOutput)
			summary := models.NewDriftSummaryFromPlanSuccess(pr.PlanSuccess)
			driftSummary = &summary
		}
	}
	return output, driftSummary
}

func applyRemediationOutput(result *command.Result) strings.Builder {
	var output strings.Builder
	for _, pr := range result.ProjectResults {
		if pr.Error != nil {
			fmt.Fprintf(&output, "Error: %v\n", pr.Error)
		} else if pr.Failure != "" {
			fmt.Fprintf(&output, "Failure: %s\n", pr.Failure)
		} else if pr.ApplySuccess != "" {
			output.WriteString(pr.ApplySuccess)
		}
	}
	return output
}

func projectRemediationResultsFromPlan(projects []models.ProjectDrift, result *command.Result) []models.ProjectRemediationResult {
	remediationResults := make([]models.ProjectRemediationResult, 0, len(projects))
	for _, project := range projects {
		remediationResult := models.ProjectRemediationResult{
			ProjectName: project.ProjectName,
			Path:        project.Path,
			Workspace:   project.Workspace,
			Status:      models.RemediationStatusRunning,
		}
		if project.Drift.HasDrift {
			driftBefore := project.Drift
			remediationResult.DriftBefore = &driftBefore
		}
		remediationResults = append(remediationResults, remediationResult)
	}
	if result == nil {
		return remediationResults
	}
	for _, projectResult := range result.ProjectResults {
		idx := findProjectRemediationResult(remediationResults, projectResult.ProjectName, projectResult.RepoRelDir, projectResult.Workspace)
		if idx == -1 {
			remediationResults = append(remediationResults, models.ProjectRemediationResult{
				ProjectName: projectResult.ProjectName,
				Path:        projectResult.RepoRelDir,
				Workspace:   projectResult.Workspace,
				Status:      models.RemediationStatusRunning,
			})
			idx = len(remediationResults) - 1
		}
		planOutput, _ := planProjectRemediationOutput(projectResult)
		remediationResults[idx].PlanOutput = planOutput
		if projectResult.Error != nil {
			remediationResults[idx].Status = models.RemediationStatusFailed
			remediationResults[idx].Error = projectResult.Error.Error()
		} else if projectResult.Failure != "" {
			remediationResults[idx].Status = models.RemediationStatusFailed
			remediationResults[idx].Error = projectResult.Failure
		}
	}
	return remediationResults
}

func mergeApplyRemediationResults(remediationResults []models.ProjectRemediationResult, result *command.Result) []models.ProjectRemediationResult {
	if result == nil {
		return remediationResults
	}
	for _, projectResult := range result.ProjectResults {
		idx := findProjectRemediationResult(remediationResults, projectResult.ProjectName, projectResult.RepoRelDir, projectResult.Workspace)
		if idx == -1 {
			remediationResults = append(remediationResults, models.ProjectRemediationResult{
				ProjectName: projectResult.ProjectName,
				Path:        projectResult.RepoRelDir,
				Workspace:   projectResult.Workspace,
				Status:      models.RemediationStatusRunning,
			})
			idx = len(remediationResults) - 1
		}
		remediationResults[idx].ApplyOutput = applyProjectRemediationOutput(projectResult)
		if projectResult.Error != nil {
			remediationResults[idx].Status = models.RemediationStatusFailed
			remediationResults[idx].Error = projectResult.Error.Error()
		} else if projectResult.Failure != "" {
			remediationResults[idx].Status = models.RemediationStatusFailed
			remediationResults[idx].Error = projectResult.Failure
		} else {
			remediationResults[idx].Status = models.RemediationStatusSuccess
			remediationResults[idx].DriftAfter = &models.DriftSummary{
				HasDrift: false,
				Summary:  "Apply completed successfully",
			}
		}
	}
	return remediationResults
}

func findProjectRemediationResult(results []models.ProjectRemediationResult, projectName, path, workspace string) int {
	for i, result := range results {
		if result.ProjectName == projectName && result.Path == path && result.Workspace == workspace {
			return i
		}
	}
	return -1
}

func markRunningRemediationResultsFailed(results []models.ProjectRemediationResult, errorMessage string) {
	for i := range results {
		if results[i].Status == models.RemediationStatusRunning {
			results[i].Status = models.RemediationStatusFailed
			results[i].Error = errorMessage
		}
	}
}

func planProjectRemediationOutput(result command.ProjectResult) (string, *models.DriftSummary) {
	var output strings.Builder
	var driftSummary *models.DriftSummary
	if result.Error != nil {
		fmt.Fprintf(&output, "Error: %v\n", result.Error)
	} else if result.Failure != "" {
		fmt.Fprintf(&output, "Failure: %s\n", result.Failure)
	} else if result.PlanSuccess != nil {
		output.WriteString(result.PlanSuccess.TerraformOutput)
		summary := models.NewDriftSummaryFromPlanSuccess(result.PlanSuccess)
		driftSummary = &summary
	}
	return output.String(), driftSummary
}

func applyProjectRemediationOutput(result command.ProjectResult) string {
	var output strings.Builder
	if result.Error != nil {
		fmt.Fprintf(&output, "Error: %v\n", result.Error)
	} else if result.Failure != "" {
		fmt.Fprintf(&output, "Failure: %s\n", result.Failure)
	} else if result.ApplySuccess != "" {
		output.WriteString(result.ApplySuccess)
	}
	return output.String()
}

// GetRemediationResult handles GET /api/drift/remediate/{id} requests.
// It retrieves a specific remediation result by ID.
// This is an authenticated endpoint that requires the API secret.
func (a *APIController) GetRemediationResult(w http.ResponseWriter, r *http.Request) {
	middleware := a.getAPIMiddleware()
	responder := middleware.Responder

	// Authenticate
	if !middleware.RequireAuth(w, r) {
		return
	}

	// Check if remediation service is configured
	if a.RemediationService == nil {
		responder.ServiceUnavailable(w, r, "drift remediation is not enabled")
		return
	}

	// Get the ID from the gorilla/mux path variable.
	// Route registered as /api/drift/remediate/{id}.
	id := mux.Vars(r)["id"]
	if id == "" {
		// Fallback to query parameter for routers that do not populate path vars.
		id = r.URL.Query().Get("id")
	}

	if id == "" {
		responder.ValidationFailed(w, r, "missing required parameter",
			ValidationError{Field: "id", Message: "id parameter is required"})
		return
	}

	// Get the result
	result, err := a.RemediationService.GetResult(id)
	if err != nil {
		responder.NotFound(w, r, fmt.Sprintf("remediation result not found: %v", err))
		return
	}

	// Convert to API DTO and return
	apiResult := NewRemediationResultAPI(result)
	responder.Success(w, r, http.StatusOK, apiResult)
}

// ListRemediationResults handles GET /api/drift/remediate requests.
// It lists remediation results for a repository.
// Query parameters:
//   - repository: required, the full repository name (owner/repo)
//   - type: required, the VCS provider type
//   - limit: optional, maximum number of results to return (default: 10)
//
// This is an authenticated endpoint that requires the API secret.
func (a *APIController) ListRemediationResults(w http.ResponseWriter, r *http.Request) {
	middleware := a.getAPIMiddleware()
	responder := middleware.Responder

	// Authenticate
	if !middleware.RequireAuth(w, r) {
		return
	}

	// Check if remediation service is configured
	if a.RemediationService == nil {
		responder.ServiceUnavailable(w, r, "drift remediation is not enabled")
		return
	}

	// Get query parameters
	repository := r.URL.Query().Get("repository")
	if repository == "" {
		responder.ValidationFailed(w, r, "missing required parameter",
			ValidationError{Field: "repository", Message: "repository parameter is required"})
		return
	}
	vcsType := r.URL.Query().Get("type")
	if vcsType == "" {
		responder.ValidationFailed(w, r, "missing required parameter",
			ValidationError{Field: "type", Message: "type parameter is required"})
		return
	}
	VCSHostType, err := models.NewVCSHostType(vcsType)
	if err != nil {
		responder.ValidationFailed(w, r, "invalid VCS type",
			ValidationError{Field: "type", Message: err.Error()})
		return
	}
	cloneURL, err := a.VCSClient.GetCloneURL(a.Logger, VCSHostType, repository)
	if err != nil {
		responder.InternalError(w, r, fmt.Errorf("failed to get clone URL: %w", err))
		return
	}
	baseRepo, err := a.Parser.ParseAPIPlanRequest(VCSHostType, repository, cloneURL)
	if err != nil {
		responder.ValidationFailed(w, r, fmt.Sprintf("failed to parse repository: %v", err))
		return
	}
	if !a.RepoAllowlistChecker.IsAllowlisted(baseRepo.FullName, baseRepo.VCSHost.Hostname) {
		responder.Forbidden(w, r, "repository is not in the allowlist")
		return
	}

	// Parse limit (default: 10)
	limit := 10
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if _, err := fmt.Sscanf(limitStr, "%d", &limit); err != nil {
			responder.ValidationFailed(w, r, "invalid limit parameter",
				ValidationError{Field: "limit", Message: "must be a positive integer"})
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
	results, err := a.RemediationService.ListResults(baseRepo.ID(), limit)
	if err != nil {
		responder.InternalError(w, r, err)
		return
	}

	// Convert to API DTOs
	apiResults := make([]RemediationResultAPI, 0, len(results))
	for _, r := range results {
		apiResults = append(apiResults, NewRemediationResultAPI(r))
	}

	// Build response using DTO
	response := RemediationListAPI{
		Repository: repository,
		Count:      len(apiResults),
		Results:    apiResults,
	}

	responder.Success(w, r, http.StatusOK, response)
}

// DetectDrift handles POST /api/drift/detect requests.
// It triggers drift detection by running plans for the specified projects.
// This is an authenticated endpoint that requires the API secret.
func (a *APIController) DetectDrift(w http.ResponseWriter, r *http.Request) {
	middleware := a.getAPIMiddleware()
	responder := middleware.Responder

	// Authenticate
	if !middleware.RequireAuth(w, r) {
		return
	}

	// Check if drift storage is configured
	if a.DriftStorage == nil {
		responder.ServiceUnavailable(w, r, "drift detection is not enabled")
		return
	}

	// Parse the JSON payload
	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		responder.ValidationFailed(w, r, "failed to read request body")
		return
	}

	var request models.DriftDetectionRequest
	if err = json.Unmarshal(bytes, &request); err != nil {
		responder.ValidationFailed(w, r, fmt.Sprintf("failed to parse JSON: %v", err))
		return
	}

	// Validate required fields using the model's Validate method
	if validationErrors := request.Validate(); len(validationErrors) > 0 {
		fields := make([]ValidationError, 0, len(validationErrors))
		for _, fe := range validationErrors {
			fields = append(fields, ValidationError{Field: fe.Field, Message: fe.Message})
		}
		responder.ValidationFailed(w, r, "validation failed", fields...)
		return
	}

	// Check if the repo is allowlisted
	VCSHostType, err := models.NewVCSHostType(request.Type)
	if err != nil {
		responder.ValidationFailed(w, r, "invalid VCS type",
			ValidationError{Field: "type", Message: err.Error()})
		return
	}
	cloneURL, err := a.VCSClient.GetCloneURL(a.Logger, VCSHostType, request.Repository)
	if err != nil {
		responder.InternalError(w, r, fmt.Errorf("failed to get clone URL: %w", err))
		return
	}

	baseRepo, err := a.Parser.ParseAPIPlanRequest(VCSHostType, request.Repository, cloneURL)
	if err != nil {
		responder.ValidationFailed(w, r, fmt.Sprintf("failed to parse repository: %v", err))
		return
	}

	if !a.RepoAllowlistChecker.IsAllowlisted(baseRepo.FullName, baseRepo.VCSHost.Hostname) {
		responder.Forbidden(w, r, "repository is not in the allowlist")
		return
	}

	// Build API request for plan
	apiRequest := &APIRequest{
		Repository:       request.Repository,
		Ref:              request.Ref,
		Type:             request.Type,
		DiscoverProjects: true, // Enable auto-discovery when no projects/paths specified
	}

	if len(request.Projects) > 0 {
		apiRequest.Projects = request.Projects
	}
	if len(request.Paths) > 0 {
		for _, p := range request.Paths {
			apiRequest.Paths = append(apiRequest.Paths, APIRequestPath{Directory: p.Directory, Workspace: p.Workspace})
		}
	}

	// Build the command context
	ctx := &command.Context{
		HeadRepo: baseRepo,
		Pull: models.PullRequest{
			Num:        nextNonPRPullNum(), // Synthetic non-PR workflow ID.
			BaseBranch: request.Ref,
			HeadBranch: request.Ref,
			HeadCommit: request.Ref,
			BaseRepo:   baseRepo,
		},
		Scope: a.Scope,
		Log:   a.Logger,
		API:   true,
	}

	// Setup working directory
	if err := a.apiSetup(ctx, command.Plan); err != nil {
		responder.InternalError(w, r, fmt.Errorf("setup failed: %w", err))
		return
	}
	defer a.cleanupNonPRWorkingDir(ctx)

	// Run pre-workflow hooks before project discovery so hooks can
	// dynamically generate atlantis.yaml or other config files.
	preHookCmd := &events.CommentCommand{Name: command.Plan}
	if err := a.PreWorkflowHooksCommandRunner.RunPreHooks(ctx, preHookCmd); err != nil {
		if a.FailOnPreWorkflowHookError {
			responder.InternalError(w, r, fmt.Errorf("pre-workflow hook failed: %w", err))
			return
		}
		a.Logger.Warn("pre-workflow hook error (continuing): %v", err)
	}

	result, err := a.apiPlan(apiRequest, ctx)
	if err != nil {
		responder.InternalError(w, r, err)
		return
	}
	defer a.Locker.UnlockByPull(ctx.HeadRepo.FullName, ctx.Pull.Num) // nolint: errcheck

	// Process results and store drift data
	detectionResult := models.NewDriftDetectionResult(request.Repository)

	for _, pr := range result.ProjectResults {
		if pr.Command != command.Plan {
			continue
		}
		projectDrift := models.ProjectDrift{
			ProjectName: pr.ProjectName,
			Path:        pr.RepoRelDir,
			Workspace:   pr.Workspace,
			Ref:         request.Ref,
			DetectionID: detectionResult.ID,
			LastChecked: time.Now(),
		}

		if pr.Error != nil {
			projectDrift.Error = pr.Error.Error()
			projectDrift.Drift = models.DriftSummary{
				HasDrift: false,
			}
		} else if pr.Failure != "" {
			projectDrift.Error = pr.Failure
			projectDrift.Drift = models.DriftSummary{
				HasDrift: false,
			}
		} else if pr.PlanSuccess != nil {
			projectDrift.Drift = models.NewDriftSummaryFromPlanSuccess(pr.PlanSuccess)
		}

		// Store drift data
		if err := a.DriftStorage.Store(baseRepo.ID(), projectDrift); err != nil {
			a.Logger.Warn("failed to store drift data: %v", err)
		}

		detectionResult.AddProject(projectDrift)
	}

	// Send drift webhook notifications if drift was detected
	if a.DriftWebhookSender != nil && detectionResult.ProjectsWithDrift > 0 {
		webhookResult := convertToDriftWebhookResult(detectionResult)
		if err := a.DriftWebhookSender.Send(a.Logger, webhookResult); err != nil {
			a.Logger.Warn("failed to send drift webhook: %v", err)
		}
	}

	// Convert to API DTO and return
	apiResult := NewDriftDetectionResultAPI(detectionResult)

	code := http.StatusOK
	if result.HasErrors() {
		code = http.StatusMultiStatus // 207 - some projects may have failed
	}
	responder.Success(w, r, code, apiResult)
}

// convertToDriftWebhookResult converts a DriftDetectionResult to a webhook DriftResult.
func convertToDriftWebhookResult(dr *models.DriftDetectionResult) webhooks.DriftResult {
	projects := make([]webhooks.DriftProjectResult, 0, len(dr.Projects))
	for _, p := range dr.Projects {
		projects = append(projects, webhooks.DriftProjectResult{
			ProjectName: p.ProjectName,
			Path:        p.Path,
			Workspace:   p.Workspace,
			HasDrift:    p.Drift.HasDrift,
			ToAdd:       p.Drift.ToAdd,
			ToChange:    p.Drift.ToChange,
			ToDestroy:   p.Drift.ToDestroy,
			ToImport:    p.Drift.ToImport,
			ToForget:    p.Drift.ToForget,
			Summary:     p.Drift.Summary,
			Error:       p.Error,
		})
	}
	var ref string
	if len(dr.Projects) > 0 {
		ref = dr.Projects[0].Ref
	}
	return webhooks.DriftResult{
		Repository:        dr.Repository,
		Ref:               ref,
		DetectionID:       dr.ID,
		ProjectsWithDrift: dr.ProjectsWithDrift,
		TotalProjects:     dr.TotalProjects,
		Projects:          projects,
	}
}
