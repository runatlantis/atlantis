// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/runatlantis/atlantis/server/core/db"
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

var apiErrorURLCredentialRE = regexp.MustCompile(`(?i)(https?://)([^\s/@]+:)?[^\s/@]+@`)

type APIController struct {
	APISecret                       []byte
	Locker                          locking.Locker `validate:"required"`
	DriftStorage                    drift.Storage
	RemediationService              drift.RemediationService
	ApplyLockChecker                locking.ApplyLockChecker
	EnableDriftRemediation          bool
	Logger                          logging.SimpleLogging           `validate:"required"`
	Parser                          events.EventParsing             `validate:"required"`
	ProjectCommandBuilder           events.ProjectCommandBuilder    `validate:"required"`
	ProjectPlanCommandRunner        events.ProjectPlanCommandRunner `validate:"required"`
	ProjectPolicyCheckCommandRunner events.ProjectPolicyCheckCommandRunner
	ProjectApplyCommandRunner       events.ProjectApplyCommandRunner `validate:"required"`
	FailOnPreWorkflowHookError      bool
	PreWorkflowHooksCommandRunner   events.PreWorkflowHooksCommandRunner  `validate:"required"`
	PostWorkflowHooksCommandRunner  events.PostWorkflowHooksCommandRunner `validate:"required"`
	RepoAllowlistChecker            *events.RepoAllowlistChecker          `validate:"required"`
	Scope                           tally.Scope                           `validate:"required"`
	VCSClient                       vcs.Client                            `validate:"required"`
	WorkingDir                      events.WorkingDir                     `validate:"required"`
	WorkingDirLocker                events.WorkingDirLocker               `validate:"required"`
	CommitStatusUpdater             events.CommitStatusUpdater            `validate:"required"`
	// Database is required for PR-backed API plan/apply requests so their
	// managed artifacts and follow-on policy results use the same durable
	// generation boundary as comment-driven workflows. Synthetic non-PR API
	// requests intentionally retain their in-memory lifecycle.
	Database db.Database
	// PullReqStatusFetcher is optional. When set and the API request supplies a
	// PR number, it is used to populate command.Context.PullRequestStatus so
	// apply requirements like 'mergeable' and 'approved' evaluate against real
	// VCS state instead of always failing.
	PullReqStatusFetcher vcs.PullReqStatusFetcher
	// PullStatusFetcher is optional. When set and the API request supplies a PR
	// number, it is used to populate command.Context.PullStatus so generated
	// policy_check contexts can preserve existing policy approvals.
	PullStatusFetcher events.PullStatusFetcher
	// LivePullHeadFetcher is optional for tests. In production it is used for
	// PR-backed API requests to seed live PR identity data such as base branch.
	LivePullHeadFetcher events.LivePullHeadFetcher
	// DriftWebhookSender sends webhook notifications when drift is detected.
	// Nil when no drift webhooks are configured.
	DriftWebhookSender *webhooks.DriftWebhookSender
	// SilenceVCSStatusNoProjects is whether API should set commit status if no projects are found
	SilenceVCSStatusNoProjects bool

	// apiMiddleware provides common authentication and response utilities.
	// Initialized lazily via getAPIMiddleware() with sync.Once for thread safety.
	apiMiddleware           *APIMiddleware
	apiMiddlewareOnce       sync.Once
	driftFullDetectionLocks sync.Map
}

type driftFullDetectionLockKey struct {
	Repository string
	Type       string
	Ref        string
	BaseBranch string
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

func (a *APIController) lockFullDriftDetection(repository, vcsType, ref, baseBranch string) func() {
	key := driftFullDetectionLockKey{
		Repository: repository,
		Type:       vcsType,
		Ref:        ref,
		BaseBranch: baseBranch,
	}
	lock, _ := a.driftFullDetectionLocks.LoadOrStore(key, &sync.Mutex{})
	mu := lock.(*sync.Mutex)
	mu.Lock()
	return mu.Unlock
}

func (a *APIController) cleanupNonPRWorkingDir(ctx *command.Context) {
	if ctx.Pull.Num >= 0 {
		return
	}
	if err := a.WorkingDir.Delete(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull); err != nil {
		ctx.Log.Warn("cleaning up API working directory: %s", err)
	}
}

type APIRequest struct {
	Repository string `validate:"required"`
	Ref        string `validate:"required"`
	BaseBranch string `json:"base_branch,omitempty"`
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

	if ctx.SortByExecutionOrder {
		sortCommandPairsByExecutionOrder(cmds, keptCommentCommands)
	}
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

func apiRequestBaseBranch(ref, baseBranch string) string {
	if strings.TrimSpace(baseBranch) != "" {
		return normalizeAPIBranchRef(baseBranch)
	}
	return normalizeAPIBranchRef(ref)
}

func apiRequestStorageRef(ref string) string {
	return normalizeAPIBranchRef(ref)
}

func normalizeAPIBranchRef(ref string) string {
	return models.NormalizeAPIRef(ref)
}

func apiErrorStatusCode(err error) int {
	if errors.Is(err, events.ErrTeamAllowlistDenied) {
		return http.StatusForbidden
	}
	if db.IsPlanGenerationObsolete(err) {
		return http.StatusConflict
	}
	return http.StatusInternalServerError
}

func (a *APIController) apiReportLegacyError(w http.ResponseWriter, code int, err error) {
	a.getAPIMiddleware().Responder.writeJSON(w, code, map[string]string{
		"error": err.Error(),
	})
}

func sanitizeAPIError(ctx *command.Context, err error) error {
	if err == nil {
		return nil
	}
	return errors.New(sanitizeAPIErrorString(ctx, err.Error()))
}

func sanitizeAPIErrorString(ctx *command.Context, message string) string {
	if ctx != nil {
		message = sanitizeRepoCloneURL(message, ctx.HeadRepo)
		message = sanitizeRepoCloneURL(message, ctx.Pull.BaseRepo)
	}
	return apiErrorURLCredentialRE.ReplaceAllString(message, "$1<redacted>@")
}

func sanitizeRepoCloneURL(message string, repo models.Repo) string {
	if repo.CloneURL == "" || repo.SanitizedCloneURL == "" {
		return message
	}
	return strings.ReplaceAll(message, repo.CloneURL, repo.SanitizedCloneURL)
}

func (a *APIController) Plan(w http.ResponseWriter, r *http.Request) {
	middleware := a.getAPIMiddleware()
	responder := middleware.Responder

	request, ctx, code, err := a.apiParseAndValidate(r)
	if err != nil {
		a.apiReportLegacyError(w, code, err)
		return
	}

	err = a.apiSetup(ctx, command.Plan)
	if err != nil {
		a.apiReportLegacyError(w, http.StatusInternalServerError, err)
		return
	}
	defer a.cleanupNonPRWorkingDir(ctx)

	result, err := a.apiPlan(request, ctx, true)
	if err != nil {
		a.apiReportLegacyError(w, apiErrorStatusCode(err), err)
		return
	}
	if !ctx.CommandSkipped && ctx.Pull.Num <= 0 {
		defer a.Locker.UnlockByPull(ctx.HeadRepo.FullName, ctx.Pull.Num) // nolint: errcheck
	}

	statusCode := http.StatusOK
	if result.HasErrors() {
		statusCode = http.StatusInternalServerError
	}
	responder.writeJSON(w, statusCode, result)
}

func (a *APIController) Apply(w http.ResponseWriter, r *http.Request) {
	middleware := a.getAPIMiddleware()
	responder := middleware.Responder

	request, ctx, code, err := a.apiParseAndValidate(r)
	if err != nil {
		a.apiReportLegacyError(w, code, err)
		return
	}

	err = a.apiSetup(ctx, command.Apply)
	if err != nil {
		a.apiReportLegacyError(w, http.StatusInternalServerError, err)
		return
	}
	defer a.cleanupNonPRWorkingDir(ctx)

	// We must first make the plan for all projects
	result, err := a.apiPlan(request, ctx, false)
	if err != nil {
		a.apiReportLegacyError(w, apiErrorStatusCode(err), err)
		return
	}
	if ctx.CommandSkipped {
		responder.writeJSON(w, http.StatusOK, result)
		return
	}
	if ctx.Pull.Num > 0 && len(result.ProjectResults) == 0 {
		// A PR-backed apply with no selected plan projects is a no-op. It must not
		// discard an existing accepted plan or replace newer shared statuses with
		// synthetic 0/0 success.
		responder.writeJSON(w, http.StatusOK, result)
		return
	}
	if ctx.Pull.Num <= 0 {
		defer a.Locker.UnlockByPull(ctx.HeadRepo.FullName, ctx.Pull.Num) // nolint: errcheck
	}

	// The API apply endpoint runs plan first. Refresh PR status afterward so
	// apply requirements evaluate the VCS state the plan phase just produced.
	a.populatePullRequestStatus(ctx)
	if ctx.Pull.Num <= 0 {
		seedPullStatusFromPlanResult(ctx, result)
	}

	// We can now prepare and run the apply step
	result, err = a.apiApply(request, ctx)
	if err != nil {
		a.apiReportLegacyError(w, apiErrorStatusCode(err), err)
		return
	}

	statusCode := http.StatusOK
	if result.HasErrors() {
		statusCode = http.StatusInternalServerError
	}
	responder.writeJSON(w, statusCode, result)
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
		a.apiReportLegacyError(w, http.StatusInternalServerError, err)
		return
	}

	result := ListLocksResult{}
	for name, lock := range locks {
		result.Locks = append(result.Locks, LockDetail{
			Name:            name,
			ProjectName:     lock.Project.ProjectName,
			ProjectRepo:     lock.Project.RepoFullName,
			ProjectRepoPath: lock.Project.Path,
			PullID:          lock.Pull.Num,
			PullURL:         lock.Pull.URL,
			User:            lock.User.Username,
			Workspace:       lock.Workspace,
			Time:            lock.Time,
		})
	}

	responder.writeJSON(w, http.StatusOK, result)
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
//   - base_branch: optional, filter by branch context
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
	if !models.IsSupportedDriftVCSHostType(VCSHostType.String()) {
		responder.ValidationFailed(w, r, "unsupported VCS type",
			ValidationError{Field: "type", Message: "type must be one of: Github, Gitlab, Gitea"})
		return
	}
	if !models.IsValidAPIRepositoryForType(repository, VCSHostType.String()) {
		responder.ValidationFailed(w, r, "invalid repository",
			ValidationError{Field: "repository", Message: "repository must be a valid repository path for the VCS type"})
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

	pathFilter := r.URL.Query().Get("path")
	if pathFilter != "" {
		var ok bool
		pathFilter, ok = models.NormalizeAPIPath(pathFilter)
		if !ok {
			responder.ValidationFailed(w, r, "invalid path filter",
				ValidationError{Field: "path", Message: "path must be a clean repo-relative path"})
			return
		}
	}

	opts := drift.GetOptions{
		ProjectName: r.URL.Query().Get("project"),
		Path:        pathFilter,
		Workspace:   r.URL.Query().Get("workspace"),
		Ref:         apiRequestStorageRef(r.URL.Query().Get("ref")),
		BaseBranch:  normalizeAPIBranchRef(r.URL.Query().Get("base_branch")),
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

func (a *APIController) apiSetup(ctx *command.Context, cmdName command.Name) (err error) {
	if ctx.Pull.Num < 0 {
		defer func() {
			if err != nil {
				a.cleanupNonPRWorkingDir(ctx)
			}
		}()
	}

	pull := ctx.Pull
	baseRepo := ctx.Pull.BaseRepo
	headRepo := ctx.HeadRepo

	unlockFn, err := a.WorkingDirLocker.TryLock(baseRepo.FullName, pull.Num, events.DefaultWorkspace, events.DefaultRepoRelDir, "", cmdName, events.WorkingDirLockMetadataForPull(pull))
	if err != nil {
		return err
	}
	ctx.Log.Debug("got workspace lock")
	defer unlockFn()

	headBeforeResolve := ctx.Pull.HeadCommit
	baseBeforeResolve := ctx.Pull.BaseBranch
	if err := a.seedAPIPrBaseBranch(ctx); err != nil {
		return sanitizeAPIError(ctx, err)
	}

	// ensure workingDir is present
	repoDir, err := a.WorkingDir.Clone(ctx.Log, headRepo, ctx.Pull, events.DefaultWorkspace)
	if err != nil {
		return sanitizeAPIError(ctx, err)
	}

	if err := resolveAPIHeadCommit(ctx, repoDir, checkoutMergeEnabled(a.WorkingDir)); err != nil {
		return sanitizeAPIError(ctx, err)
	}
	if ctx.Pull.Num > 0 && (ctx.Pull.HeadCommit != headBeforeResolve || ctx.Pull.BaseBranch != baseBeforeResolve) {
		a.populatePullRequestStatus(ctx)
	}
	return sanitizeAPIError(ctx, verifyNonPRBaseBranchReachability(ctx, repoDir))
}

func (a *APIController) seedAPIPrBaseBranch(ctx *command.Context) error {
	if ctx.Pull.Num <= 0 || strings.TrimSpace(ctx.Pull.BaseBranch) != "" || a.LivePullHeadFetcher == nil {
		return nil
	}
	livePull, err := a.LivePullHeadFetcher.GetLivePullIdentity(command.ProjectContext{
		Log:        ctx.Log,
		Pull:       ctx.Pull,
		PullStatus: ctx.PullStatus,
		API:        ctx.API,
	})
	if err != nil {
		return fmt.Errorf("fetching live pull request: %w", err)
	}
	if strings.TrimSpace(livePull.BaseBranch) != "" {
		ctx.Pull.BaseBranch = livePull.BaseBranch
	}
	return nil
}

func resolveNonPRHeadCommit(ctx *command.Context, repoDir string) error {
	if ctx.Pull.Num >= 0 || repoDir == "" {
		return nil
	}
	return resolveAPIHeadCommit(ctx, repoDir, false)
}

func resolveAPIHeadCommit(ctx *command.Context, repoDir string, checkoutMerge bool) error {
	if repoDir == "" {
		return nil
	}
	if _, err := os.Stat(filepath.Join(repoDir, ".git")); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("checking API checkout git metadata: %w", err)
	}

	if ctx.Pull.Num > 0 && checkoutMerge {
		if headCommit, err := checkedOutAPICommit(repoDir, "HEAD^2"); err == nil && headCommit != "" {
			ctx.Pull.HeadCommit = headCommit
			return nil
		}
	}
	headCommit, err := checkedOutAPICommit(repoDir, "HEAD")
	if err != nil {
		return err
	}
	ctx.Pull.HeadCommit = headCommit
	return nil
}

func checkoutMergeEnabled(workingDir events.WorkingDir) bool {
	checkoutMerge, ok := workingDir.(interface {
		CheckoutMergeEnabled() bool
	})
	return ok && checkoutMerge.CheckoutMergeEnabled()
}

func checkedOutAPICommit(repoDir string, ref string) (string, error) {
	cmd := exec.Command("git", "rev-parse", ref) // nolint: gosec
	cmd.Dir = repoDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("resolving checked out API ref %q: %s: %w", ref, strings.TrimSpace(string(output)), err)
	}
	headCommit := strings.TrimSpace(string(output))
	if headCommit == "" {
		return "", fmt.Errorf("resolving checked out API ref %q: empty commit", ref)
	}
	return headCommit, nil
}

func verifyNonPRBaseBranchReachability(ctx *command.Context, repoDir string) error {
	if ctx.Pull.Num > 0 || repoDir == "" {
		return nil
	}
	if _, err := os.Stat(filepath.Join(repoDir, ".git")); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("checking API checkout git metadata: %w", err)
	}
	baseBranch := strings.TrimSpace(ctx.Pull.BaseBranch)
	headRef := strings.TrimSpace(ctx.Pull.HeadBranch)
	if baseBranch == "" || headRef == "" {
		return nil
	}
	if baseBranch == headRef && !models.RequiresBaseBranchForRef(headRef) {
		return nil
	}
	baseRef, err := checkedAPIBaseBranchRef(baseBranch)
	if err != nil {
		return err
	}
	remoteBaseRef := "refs/remotes/origin/" + baseRef
	fetchRef := fmt.Sprintf("+refs/heads/%s:%s", baseRef, remoteBaseRef)
	if output, err := fetchAPIBaseBranch(repoDir, fetchRef); err != nil {
		return fmt.Errorf("verifying API base branch %q: %s: %w", baseBranch, sanitizeAPIErrorString(ctx, strings.TrimSpace(output)), err)
	}
	if output, err := checkedOutCommitReachableFromAPIBase(repoDir, ctx.Pull.HeadCommit, remoteBaseRef); err == nil {
		return nil
	} else if !isShallowGitRepo(repoDir) {
		return fmt.Errorf("checked out API ref %q at commit %s is not reachable from base_branch %q: %s: %w", headRef, ctx.Pull.HeadCommit, baseBranch, sanitizeAPIErrorString(ctx, strings.TrimSpace(output)), err)
	}

	if _, err := unshallowAPIBaseBranch(repoDir, fetchRef); err != nil {
		ctx.Log.Debug("unable to unshallow API checkout while verifying base branch: %v", sanitizeAPIError(ctx, err))
	}
	if output, err := checkedOutCommitReachableFromAPIBase(repoDir, ctx.Pull.HeadCommit, remoteBaseRef); err != nil {
		return fmt.Errorf("checked out API ref %q at commit %s is not reachable from base_branch %q: %s: %w", headRef, ctx.Pull.HeadCommit, baseBranch, sanitizeAPIErrorString(ctx, strings.TrimSpace(output)), err)
	}
	return nil
}

func checkedAPIBaseBranchRef(baseBranch string) (string, error) {
	baseRef, ok := models.CheckedBaseBranchRef(baseBranch)
	if !ok {
		return "", fmt.Errorf("invalid API base_branch %q", baseBranch)
	}
	return baseRef, nil
}

func isShallowGitRepo(repoDir string) bool {
	_, err := os.Stat(filepath.Join(repoDir, ".git", "shallow"))
	return err == nil
}

func fetchAPIBaseBranch(repoDir string, fetchRef string) (string, error) {
	cmd := exec.Command("git", "fetch", "origin", "--", fetchRef) //nolint:gosec // fetchRef is constructed from a validated base branch.
	cmd.Dir = repoDir
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func unshallowAPIBaseBranch(repoDir string, fetchRef string) (string, error) {
	cmd := exec.Command("git", "fetch", "--unshallow", "origin", "--", fetchRef) //nolint:gosec // fetchRef is constructed from a validated base branch.
	cmd.Dir = repoDir
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func checkedOutCommitReachableFromAPIBase(repoDir string, headCommit string, remoteBaseRef string) (string, error) {
	cmd := exec.Command("git", "merge-base", "--is-ancestor", headCommit, remoteBaseRef) //nolint:gosec // refs are resolved from the checked-out commit and validated base branch.
	cmd.Dir = repoDir
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func (a *APIController) apiPlan(request *APIRequest, ctx *command.Context, replacePRStatusWhenEmpty bool) (*command.Result, error) {
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
		if ctx.Pull.Num > 0 {
			if !replacePRStatusWhenEmpty {
				return &command.Result{ProjectResults: []command.ProjectResult{}}, nil
			}
			if a.Database == nil {
				return nil, errors.New("database is required for PR-backed API plan")
			}
			publicationToken, claimErr := a.acquirePlanPublicationClaim(ctx, "replacing PR-backed API plan with no projects")
			if claimErr != nil {
				return nil, claimErr
			}
			beginResult, beginErr := a.Database.BeginPlanGenerationReplacing(ctx.Pull, nil, uuid.NewString(), publicationToken)
			if beginErr != nil {
				var statusErr error
				if !ctx.SuppressVCSStatus {
					statusErr = a.CommitStatusUpdater.UpdateCombined(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.FailedCommitStatus, command.Plan)
				}
				claimErr = a.finishPlanPublicationClaim(ctx, publicationToken, statusErr)
				return nil, errors.Join(fmt.Errorf("replacing PR-backed API plan with no projects: %w", beginErr), claimErr)
			}
			ctx.PullStatus = &beginResult.PullStatus
			var publicationErr error
			if !a.SilenceVCSStatusNoProjects && !ctx.SuppressVCSStatus {
				status := models.SuccessCommitStatus
				counts := models.ProjectCounts{}
				if len(beginResult.Canceled) > 0 {
					status = models.FailedCommitStatus
					counts = models.ProjectCounts{Total: len(beginResult.Canceled), Errored: len(beginResult.Canceled)}
				}
				publicationErr = errors.Join(
					a.CommitStatusUpdater.UpdateCombinedCount(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, status, command.Plan, counts),
					a.CommitStatusUpdater.UpdateCombinedCount(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, status, command.PolicyCheck, counts),
					a.CommitStatusUpdater.UpdateCombinedCount(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, status, command.Apply, counts),
				)
			}
			if claimErr := a.finishPlanPublicationClaim(ctx, publicationToken, errors.Join(
				a.publishCancelledPlanStatuses(apiCancelledPlanProjectContexts(ctx, beginResult.Canceled)),
				publicationErr,
			)); claimErr != nil {
				return nil, claimErr
			}
			return &command.Result{ProjectResults: []command.ProjectResult{}}, nil
		}
		if err := a.validateNonPRAPIRefUnchanged(ctx); err != nil {
			return nil, err
		}
		// When silence is enabled and no projects are found, don't set any VCS status
		if !a.SilenceVCSStatusNoProjects && !ctx.SuppressVCSStatus {
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
	var planCmds []command.ProjectContext
	var planCC []*events.CommentCommand
	var policyCmds []command.ProjectContext
	for i, cmd := range cmds {
		switch cmd.CommandName {
		case command.Plan:
			planCmds = append(planCmds, cmd)
			planCC = append(planCC, cc[i])
		case command.PolicyCheck:
			policyCmds = append(policyCmds, cmd)
		default:
			return nil, fmt.Errorf("%s is not supported", cmd.CommandName)
		}
	}

	planGeneration := ""
	publicationToken := ""
	if ctx.Pull.Num > 0 {
		if a.Database == nil {
			return nil, errors.New("database is required for PR-backed API plan")
		}
		publicationToken, err = a.acquirePlanPublicationClaim(ctx, "beginning PR-backed API plan")
		if err != nil {
			return nil, err
		}
		planGeneration = uuid.NewString()
		startedStatus, beginErr := a.Database.BeginPlanGeneration(ctx.Pull, apiPlanGenerationProjects(ctx, planCmds), planGeneration, publicationToken)
		if beginErr != nil {
			var statusErr error
			if !ctx.SuppressVCSStatus {
				statusErr = a.CommitStatusUpdater.UpdateCombined(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.FailedCommitStatus, command.Plan)
			}
			claimErr := a.finishPlanPublicationClaim(ctx, publicationToken, errors.Join(a.publishPlanGenerationStartFailureStatuses(planCmds, beginErr), statusErr))
			return nil, errors.Join(fmt.Errorf("beginning PR-backed API plan generation: %w", beginErr), claimErr)
		}
		var aggregatePendingErr error
		if !ctx.SuppressVCSStatus {
			aggregatePendingErr = a.CommitStatusUpdater.UpdateCombined(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.PendingCommitStatus, command.Plan)
		}
		if claimErr := a.finishPlanPublicationClaim(ctx, publicationToken, errors.Join(
			a.publishPendingPlanStatuses(planCmds),
			a.publishCancelledPlanStatuses(apiCancelledPlanProjectContexts(ctx, startedStatus.Canceled)),
			aggregatePendingErr,
		)); claimErr != nil {
			return nil, claimErr
		}
		prepareAPIProjectCommandsForPlanGeneration(planCmds, planGeneration)
		prepareAPIProjectCommandsForPlanGeneration(policyCmds, planGeneration)
		ctx.PullStatus = &startedStatus.PullStatus
	} else if !ctx.SuppressVCSStatus {
		if err := a.CommitStatusUpdater.UpdateCombined(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.PendingCommitStatus, command.Plan); err != nil {
			ctx.Log.Warn("unable to update plan commit status: %s", err)
		}
	}

	var projectResults []command.ProjectResult
	for i, cmd := range planCmds {
		if !ctx.PreWorkflowHooksAlreadyRun {
			err = a.runAPIPlanWorkflowHook(ctx, planGeneration, cmd, "pre-workflow hook", func() error {
				return a.PreWorkflowHooksCommandRunner.RunPreHooks(ctx, planCC[i])
			})
			if err != nil {
				if db.IsPlanGenerationObsolete(err) {
					result := &command.Result{ProjectResults: projectResults}
					a.completeUnpublishedPlanJobs(planCmds, result, "This plan was superseded before its workflow hook completed.")
					return result, err
				}
				if a.FailOnPreWorkflowHookError {
					failureResults := append(projectResults, apiPlanFailureResults(planCmds[i:], err)...)
					failureResult := &command.Result{ProjectResults: failureResults, Error: err}
					if planGeneration != "" {
						publicationToken, claimErr := a.waitForPlanPublicationClaim(ctx, "completing failed PR-backed API plan")
						if claimErr != nil {
							return failureResult, claimErr
						}
						completedStatus, completeErr := a.Database.CompletePlanGeneration(ctx.Pull, planGeneration, uniqueAPIProjectResults(failureResults), publicationToken)
						if completeErr != nil {
							claimErr := a.finishPlanPublicationClaim(ctx, publicationToken, nil)
							return failureResult, errors.Join(fmt.Errorf("completing failed PR-backed API plan generation: %w", completeErr), claimErr)
						}
						ctx.PullStatus = &completedStatus
						var aggregateErr error
						if !ctx.SuppressVCSStatus {
							aggregateErr = a.CommitStatusUpdater.UpdateCombined(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.FailedCommitStatus, command.Plan)
						}
						if claimErr := a.finishPlanPublicationClaim(ctx, publicationToken, errors.Join(
							a.publishDeferredPlanStatuses(planCmds, failureResult, models.FailedCommitStatus),
							aggregateErr,
						)); claimErr != nil {
							return failureResult, claimErr
						}
					} else {
						publicationErr := a.publishDeferredPlanStatuses(planCmds, failureResult, models.FailedCommitStatus)
						if !ctx.SuppressVCSStatus {
							if statusErr := a.CommitStatusUpdater.UpdateCombined(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.FailedCommitStatus, command.Plan); statusErr != nil {
								ctx.Log.Warn("unable to update plan commit status: %s", statusErr)
								publicationErr = errors.Join(publicationErr, statusErr)
							}
						}
						if publicationErr != nil {
							return failureResult, errors.Join(err, publicationErr)
						}
					}
					return failureResult, err
				}
			}
		}

		res := events.RunOneProjectCmd(a.ProjectPlanCommandRunner.Plan, cmd)
		projectResults = append(projectResults, res)

		if hookErr := a.runAPIPlanWorkflowHook(ctx, planGeneration, cmd, "post-workflow hook", func() error {
			return a.PostWorkflowHooksCommandRunner.RunPostHooks(ctx, planCC[i])
		}); hookErr != nil {
			if db.IsPlanGenerationObsolete(hookErr) || events.IsWorkflowHookStatusPublicationError(hookErr) {
				result := &command.Result{ProjectResults: projectResults}
				a.completeUnpublishedPlanJobs(planCmds, result, "This plan was superseded before its workflow hook completed.")
				return result, hookErr
			}
			ctx.Log.Warn("post-workflow hook failed, %v", hookErr)
		}
	}

	result := &command.Result{ProjectResults: projectResults}
	if planGeneration != "" {
		publicationToken, claimErr := a.waitForPlanPublicationClaim(ctx, "completing PR-backed API plan")
		if claimErr != nil {
			a.completeUnpublishedPlanJobs(planCmds, result, "Atlantis could not claim the durable publication boundary for this completed plan. Run the API plan again.")
			return result, claimErr
		}
		completedStatus, completeErr := a.Database.CompletePlanGeneration(ctx.Pull, planGeneration, uniqueAPIProjectResults(projectResults), publicationToken)
		if completeErr != nil {
			if !db.IsPlanGenerationObsolete(completeErr) {
				claimErr := a.finishPlanPublicationClaim(ctx, publicationToken, a.publishDeferredPlanStatuses(planCmds, result, models.FailedCommitStatus))
				return result, errors.Join(fmt.Errorf("completing PR-backed API plan generation: %w", completeErr), claimErr)
			}
			ctx.Log.Info("PR-backed API plan generation was superseded; ignoring obsolete results")
			claimErr := a.finishPlanPublicationClaim(ctx, publicationToken, nil)
			return result, errors.Join(fmt.Errorf("PR-backed API plan was superseded by a newer plan generation: %w", completeErr), claimErr)
		}
		ctx.PullStatus = &completedStatus
		planStatus := models.SuccessCommitStatus
		if result.HasErrors() {
			planStatus = models.FailedCommitStatus
		}
		var aggregateErr error
		if !ctx.SuppressVCSStatus {
			aggregateErr = a.CommitStatusUpdater.UpdateCombined(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, planStatus, command.Plan)
		}
		if claimErr := a.finishPlanPublicationClaim(ctx, publicationToken, errors.Join(
			a.publishDeferredPlanStatuses(planCmds, result, models.SuccessCommitStatus),
			aggregateErr,
		)); claimErr != nil {
			return result, claimErr
		}
	} else {
		if publicationErr := a.publishDeferredPlanStatuses(planCmds, result, models.SuccessCommitStatus); publicationErr != nil {
			return result, publicationErr
		}
	}
	if !ctx.RunPolicyChecks || result.HasErrors() || len(planCmds) == 0 {
		return result, nil
	}

	for _, cmd := range policyCmds {
		if a.ProjectPolicyCheckCommandRunner == nil {
			return nil, fmt.Errorf("policy check runner is not configured")
		}
		res := events.RunOneProjectCmd(a.ProjectPolicyCheckCommandRunner.PolicyCheck, cmd)
		projectResults = append(projectResults, res)
	}
	if planGeneration != "" && len(projectResults) > len(result.ProjectResults) {
		policyResults := uniqueAPIProjectResults(projectResults[len(result.ProjectResults):])
		publicationToken, claimErr := a.waitForPlanPublicationClaim(ctx, "persisting PR-backed API policy results")
		if claimErr != nil {
			return &command.Result{ProjectResults: projectResults}, claimErr
		}
		policyStatus, policyErr := a.Database.UpdatePolicyResultsForPlanGeneration(ctx.Pull, policyResults, publicationToken)
		if policyErr != nil {
			if db.IsPlanGenerationObsolete(policyErr) {
				ctx.Log.Info("PR-backed API policy results were superseded; ignoring obsolete results")
				claimErr := a.finishPlanPublicationClaim(ctx, publicationToken, nil)
				return &command.Result{ProjectResults: projectResults}, errors.Join(fmt.Errorf("PR-backed API policy results were superseded by a newer plan generation: %w", policyErr), claimErr)
			}
			claimErr := a.finishPlanPublicationClaim(ctx, publicationToken, policyErr)
			return &command.Result{ProjectResults: projectResults}, errors.Join(fmt.Errorf("persisting PR-backed API policy results: %w", policyErr), claimErr)
		}
		ctx.PullStatus = &policyStatus
		if claimErr := a.finishPlanPublicationClaim(ctx, publicationToken, nil); claimErr != nil {
			return &command.Result{ProjectResults: projectResults}, claimErr
		}
	}
	return &command.Result{ProjectResults: projectResults}, nil
}

func (a *APIController) apiApply(request *APIRequest, ctx *command.Context) (*command.Result, error) {
	publicationToken := ""
	claimActive := false
	finishClaim := func(publicationErr error) error {
		if !claimActive {
			return publicationErr
		}
		claimActive = false
		return a.finishPlanPublicationClaim(ctx, publicationToken, publicationErr)
	}
	defer func() {
		if claimActive {
			if claimErr := finishClaim(nil); claimErr != nil {
				ctx.Log.Err("releasing PR-backed API apply publication claim: %s", claimErr)
			}
		}
	}()

	if ctx.Pull.Num > 0 {
		if a.Database == nil {
			return nil, errors.New("database is required for PR-backed API apply")
		}
		expectedPullStatus := ctx.PullStatus
		var err error
		publicationToken, err = a.acquirePlanPublicationClaim(ctx, "executing PR-backed API apply")
		if err != nil {
			return nil, err
		}
		claimActive = true
		pullStatus, statusErr := a.Database.GetPullStatus(ctx.Pull)
		if statusErr != nil {
			return nil, errors.Join(fmt.Errorf("fetching current PR-backed API plan status: %w", statusErr), finishClaim(nil))
		}
		if pullStatus == nil {
			return nil, errors.Join(errors.New("current PR-backed API plan status is missing"), finishClaim(nil))
		}
		if expectedPullStatus != nil && !reflect.DeepEqual(expectedPullStatus, pullStatus) {
			return nil, errors.Join(errors.New("PR-backed API plan generation changed before apply; rerun the API apply request"), finishClaim(nil))
		}
		ctx.PullStatus = pullStatus
	}

	cmds, cc, err := request.getCommands(ctx, command.Apply, a.ProjectCommandBuilder.BuildApplyCommands)
	if events.IsIgnoredTargetedDir(err) {
		ctx.CommandSkipped = true
		return &command.Result{ProjectResults: []command.ProjectResult{}}, finishClaim(nil)
	}
	if err != nil {
		return nil, errors.Join(err, finishClaim(nil))
	}

	if len(cmds) == 0 {
		ctx.Log.Info("determined there was no project to run apply in")
		if ctx.Pull.Num > 0 {
			// A PR-backed no-project apply has no durable result to publish. Leaving
			// current statuses untouched avoids an obsolete API request replacing a
			// newer generation's user-visible state with synthetic 0/0 success.
			return &command.Result{ProjectResults: []command.ProjectResult{}}, finishClaim(nil)
		}
		if err := a.validateNonPRAPIRefUnchanged(ctx); err != nil {
			return nil, err
		}
		// When silence is enabled and no projects are found, don't set any VCS status
		if !a.SilenceVCSStatusNoProjects && !ctx.SuppressVCSStatus {
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

	if err := prepareAPIApplyCommandsForAcceptedGeneration(ctx, cmds); err != nil {
		return nil, errors.Join(err, finishClaim(nil))
	}

	// Update the combined apply commit status to pending
	if !ctx.SuppressVCSStatus {
		if err := a.CommitStatusUpdater.UpdateCombined(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.PendingCommitStatus, command.Apply); err != nil {
			if ctx.Pull.Num > 0 {
				return nil, finishClaim(err)
			}
			ctx.Log.Warn("unable to update apply commit status: %s", err)
		}
	}

	var projectResults []command.ProjectResult
	var currentExecutionGroup int
	var currentGroupHasResult bool
	var currentGroupHasErrors bool
	var currentGroupAbortOnFailure bool
	for i, cmd := range cmds {
		if ctx.SuppressVCSStatus && currentGroupHasResult && cmd.ExecutionOrderGroup != currentExecutionGroup && currentGroupHasErrors && currentGroupAbortOnFailure {
			ctx.Log.Info("abort on execution order when failed")
			break
		}
		if !currentGroupHasResult || cmd.ExecutionOrderGroup != currentExecutionGroup {
			currentExecutionGroup = cmd.ExecutionOrderGroup
			currentGroupHasResult = true
			currentGroupHasErrors = false
			currentGroupAbortOnFailure = cmd.AbortOnExecutionOrderFail
		}
		if !ctx.PreWorkflowHooksAlreadyRun {
			err = a.PreWorkflowHooksCommandRunner.RunPreHooks(ctx, cc[i])
			if err != nil {
				if events.IsWorkflowHookStatusPublicationError(err) {
					return nil, finishClaim(err)
				}
				if a.FailOnPreWorkflowHookError {
					var publicationErr error
					if ctx.Pull.Num > 0 && !ctx.SuppressVCSStatus {
						publicationErr = a.CommitStatusUpdater.UpdateCombined(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.FailedCommitStatus, command.Apply)
					}
					return nil, errors.Join(err, finishClaim(publicationErr))
				}
			}
		}

		res := events.RunOneProjectCmd(a.ProjectApplyCommandRunner.Apply, cmd)
		projectResults = append(projectResults, res)
		if res.Error != nil || res.Failure != "" {
			currentGroupHasErrors = true
		}
		if ctx.Pull.Num <= 0 {
			updatePullStatusFromProjectResult(ctx, res)
		}

		if hookErr := a.PostWorkflowHooksCommandRunner.RunPostHooks(ctx, cc[i]); hookErr != nil {
			if events.IsWorkflowHookStatusPublicationError(hookErr) {
				return &command.Result{ProjectResults: projectResults}, finishClaim(hookErr)
			}
			ctx.Log.Warn("post-workflow hook failed, %v", hookErr)
		}
	}
	result := &command.Result{ProjectResults: projectResults}
	if ctx.Pull.Num > 0 {
		pullStatus, updateErr := a.Database.UpdateApplyResultsForPlanGeneration(ctx.Pull, projectResults, publicationToken)
		if updateErr != nil {
			if db.IsPlanGenerationObsolete(updateErr) {
				ctx.Log.Warn("PR-backed API apply result was superseded after execution; durable plan state was not changed")
				claimErr := finishClaim(nil)
				return result, errors.Join(fmt.Errorf("PR-backed API apply result was superseded by a newer plan generation after execution; verify infrastructure state: %w", updateErr), claimErr)
			}
			publicationErr := a.publishDeferredApplyStatuses(cmds, result, models.FailedCommitStatus)
			claimErr := finishClaim(errors.Join(publicationErr, updateErr))
			return result, errors.Join(fmt.Errorf("persisting PR-backed API apply results: %w", updateErr), claimErr)
		}
		ctx.PullStatus = &pullStatus
		status := models.SuccessCommitStatus
		if result.HasErrors() {
			status = models.FailedCommitStatus
		}
		if claimErr := finishClaim(a.publishDeferredApplyStatuses(cmds, result, status)); claimErr != nil {
			return result, claimErr
		}
		return result, nil
	}
	if err := a.validateNonPRAPIRefUnchanged(ctx); err != nil {
		return result, errors.Join(err, a.publishDeferredApplyStatuses(cmds, result, models.FailedCommitStatus))
	}
	return result, a.publishDeferredApplyStatuses(cmds, result, models.SuccessCommitStatus)
}

func (a *APIController) validateNonPRAPIRefUnchanged(ctx *command.Context) error {
	if ctx == nil || !ctx.API || ctx.Pull.Num > 0 {
		return nil
	}
	repoDir, err := a.WorkingDir.GetWorkingDir(ctx.Pull.BaseRepo, ctx.Pull, events.DefaultWorkspace)
	if err != nil {
		return sanitizeAPIError(ctx, err)
	}
	return sanitizeAPIError(ctx, events.ValidateNonPRAPIRefUnchanged(command.ProjectContext{
		Log:  ctx.Log,
		Pull: ctx.Pull,
		API:  ctx.API,
	}, repoDir))
}

func (a *APIController) publishDeferredApplyStatuses(projectCmds []command.ProjectContext, result *command.Result, status models.CommitStatus) error {
	if result == nil {
		return nil
	}
	publisher, ok := a.ProjectApplyCommandRunner.(events.DeferredApplyStatusPublisher)
	if !ok {
		return nil
	}
	return publisher.PublishDeferredApplyStatuses(projectCmds, *result, status)
}

func (a *APIController) publishDeferredPlanStatuses(projectCmds []command.ProjectContext, result *command.Result, status models.CommitStatus) error {
	if result == nil {
		return nil
	}
	publisher, ok := a.ProjectPlanCommandRunner.(events.DeferredPlanStatusPublisher)
	if !ok {
		return nil
	}
	return publisher.PublishDeferredPlanStatuses(projectCmds, *result, status)
}

func (a *APIController) completeUnpublishedPlanJobs(projectCmds []command.ProjectContext, result *command.Result, message string) {
	if result == nil {
		return
	}
	publisher, ok := a.ProjectPlanCommandRunner.(events.DeferredPlanStatusPublisher)
	if !ok {
		return
	}
	publisher.CompleteUnpublishedPlanJobs(projectCmds, *result, message)
}

func (a *APIController) publishPendingPlanStatuses(projectCmds []command.ProjectContext) error {
	publisher, ok := a.ProjectPlanCommandRunner.(events.DeferredPlanStatusPublisher)
	if !ok {
		return nil
	}
	return publisher.PublishPendingPlanStatuses(projectCmds)
}

func (a *APIController) publishCancelledPlanStatuses(projectCmds []command.ProjectContext) error {
	if len(projectCmds) == 0 {
		return nil
	}
	publisher, ok := a.ProjectPlanCommandRunner.(events.DeferredPlanStatusPublisher)
	if !ok {
		return nil
	}
	return publisher.PublishCancelledPlanStatuses(projectCmds)
}

func (a *APIController) publishPlanGenerationStartFailureStatuses(projectCmds []command.ProjectContext, err error) error {
	publisher, ok := a.ProjectPlanCommandRunner.(events.DeferredPlanStatusPublisher)
	if !ok {
		return nil
	}
	return publisher.PublishPlanGenerationStartFailureStatuses(projectCmds, err)
}

func (a *APIController) acquirePlanPublicationClaim(ctx *command.Context, operation string) (string, error) {
	token := uuid.NewString()
	if err := a.Database.AcquirePlanPublicationClaim(ctx.Pull, token); err != nil {
		return "", fmt.Errorf("%s: %w", operation, err)
	}
	return token, nil
}

func (a *APIController) waitForPlanPublicationClaim(ctx *command.Context, operation string) (string, error) {
	token := uuid.NewString()
	for {
		err := a.Database.AcquirePlanPublicationClaim(ctx.Pull, token)
		if err == nil {
			return token, nil
		}
		if !errors.Is(err, db.ErrPlanPublicationBusy) {
			return "", fmt.Errorf("%s: %w", operation, err)
		}
		time.Sleep(25 * time.Millisecond)
	}
}

func (a *APIController) finishPlanPublicationClaim(ctx *command.Context, token string, publicationErr error) error {
	if publicationErr != nil {
		return fmt.Errorf("publishing PR-backed API plan state; publication claim retained for offline recovery: %w", publicationErr)
	}
	if err := a.Database.ReleasePlanPublicationClaim(ctx.Pull, token); err != nil {
		return fmt.Errorf("releasing PR-backed API plan publication claim: %w", err)
	}
	return nil
}

func (a *APIController) runAPIPlanWorkflowHook(
	ctx *command.Context,
	planGeneration string,
	project command.ProjectContext,
	phase string,
	run func() error,
) error {
	if ctx.Pull.Num <= 0 || planGeneration == "" {
		return run()
	}
	publicationToken, err := a.waitForPlanPublicationClaim(ctx, "running PR-backed API "+phase)
	if err != nil {
		return err
	}
	pullStatus, err := a.Database.GetPullStatus(ctx.Pull)
	if err != nil {
		return errors.Join(
			fmt.Errorf("reading plan status before %s: %w", phase, err),
			a.finishPlanPublicationClaim(ctx, publicationToken, nil),
		)
	}
	current := false
	if pullStatus != nil {
		for _, status := range pullStatus.Projects {
			if apiProjectIdentityMatches(status, project) && status.PlanGeneration == planGeneration {
				current = true
				break
			}
		}
	}
	if !current {
		claimErr := a.finishPlanPublicationClaim(ctx, publicationToken, nil)
		return errors.Join(fmt.Errorf("%w: PR-backed API %s belongs to generation %q", db.ErrPlanGenerationSuperseded, phase, planGeneration), claimErr)
	}

	hookErr := run()
	if events.IsWorkflowHookStatusPublicationError(hookErr) {
		return a.finishPlanPublicationClaim(ctx, publicationToken, hookErr)
	}
	return errors.Join(hookErr, a.finishPlanPublicationClaim(ctx, publicationToken, nil))
}

func apiPlanGenerationProjects(ctx *command.Context, cmds []command.ProjectContext) []models.ProjectStatus {
	projects := make([]models.ProjectStatus, 0, len(cmds))
	seen := make(map[string]struct{}, len(cmds))
	for _, cmd := range cmds {
		key := fmt.Sprintf("%s\x00%s\x00%s", cmd.Workspace, filepath.Clean(cmd.RepoRelDir), cmd.ProjectName)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		project := models.ProjectStatus{
			Workspace:   cmd.Workspace,
			RepoRelDir:  cmd.RepoRelDir,
			ProjectName: cmd.ProjectName,
		}
		if ctx.PullStatus != nil {
			for _, existing := range ctx.PullStatus.Projects {
				if existing.Workspace == cmd.Workspace && filepath.Clean(existing.RepoRelDir) == filepath.Clean(cmd.RepoRelDir) && existing.ProjectName == cmd.ProjectName {
					project.PolicyStatus = existing.PolicyStatus
					break
				}
			}
		}
		projects = append(projects, project)
	}
	return projects
}

func apiCancelledPlanProjectContexts(ctx *command.Context, projects []db.PlanGenerationProject) []command.ProjectContext {
	contexts := make([]command.ProjectContext, 0, len(projects))
	for _, project := range projects {
		contexts = append(contexts, command.ProjectContext{
			Log:         ctx.Log,
			CommandName: command.Plan,
			BaseRepo:    ctx.Pull.BaseRepo,
			Pull:        ctx.Pull,
			Workspace:   project.Workspace,
			RepoRelDir:  project.RepoRelDir,
			ProjectName: project.ProjectName,
		})
	}
	return contexts
}

func apiPlanFailureResults(cmds []command.ProjectContext, err error) []command.ProjectResult {
	results := make([]command.ProjectResult, 0, len(cmds))
	for _, cmd := range cmds {
		results = append(results, command.ProjectResult{
			ProjectCommandOutput: command.ProjectCommandOutput{Error: err},
			Command:              command.Plan,
			SubCommand:           cmd.SubCommand,
			RepoRelDir:           cmd.RepoRelDir,
			Workspace:            cmd.Workspace,
			ProjectName:          cmd.ProjectName,
			SilencePRComments:    cmd.SilencePRComments,
		})
	}
	return results
}

func uniqueAPIProjectResults(results []command.ProjectResult) []command.ProjectResult {
	unique := make([]command.ProjectResult, 0, len(results))
	indexes := make(map[string]int, len(results))
	for _, result := range results {
		key := fmt.Sprintf("%s\x00%s\x00%s", result.Workspace, filepath.Clean(result.RepoRelDir), result.ProjectName)
		if idx, ok := indexes[key]; ok {
			unique[idx] = result
			continue
		}
		indexes[key] = len(unique)
		unique = append(unique, result)
	}
	return unique
}

func prepareAPIProjectCommandsForPlanGeneration(cmds []command.ProjectContext, generation string) {
	for i := range cmds {
		cmds[i].PlanGeneration = generation
		cmds[i].AcceptedPlanGeneration = generation
		if !cmds[i].RequiresAtlantisManagedPlanFile {
			continue
		}
		managedPlanHash := ""
		cmds[i].GeneratedPlanHash = &managedPlanHash
	}
}

func prepareAPIApplyCommandsForAcceptedGeneration(ctx *command.Context, cmds []command.ProjectContext) error {
	if ctx.PullStatus == nil {
		if ctx.Pull.Num > 0 {
			return errors.New("current PR-backed API plan status is missing")
		}
		return nil
	}
	for i := range cmds {
		found := false
		for _, project := range ctx.PullStatus.Projects {
			if apiProjectIdentityMatches(project, cmds[i]) {
				if cmds[i].AcceptedPlanGeneration == "" {
					// Legacy/custom builders may omit this correlation field. The
					// publication claim and command-start PullStatus equality check
					// above make copying from this exact snapshot safe.
					cmds[i].AcceptedPlanGeneration = project.AcceptedPlanGeneration
				} else if cmds[i].AcceptedPlanGeneration != project.AcceptedPlanGeneration {
					return fmt.Errorf(
						"plan generation changed before PR-backed API apply for dir %q workspace %q project %q; rerun the API apply request",
						cmds[i].RepoRelDir, cmds[i].Workspace, cmds[i].ProjectName,
					)
				}
				found = true
				break
			}
		}
		if !found && ctx.Pull.Num > 0 {
			return fmt.Errorf(
				"no current plan status exists for PR-backed API apply dir %q workspace %q project %q; run `atlantis plan`",
				cmds[i].RepoRelDir, cmds[i].Workspace, cmds[i].ProjectName,
			)
		}
	}
	return nil
}

func apiProjectIdentityMatches(project models.ProjectStatus, ctx command.ProjectContext) bool {
	projectWorkspace := project.Workspace
	if projectWorkspace == "" {
		projectWorkspace = events.DefaultWorkspace
	}
	ctxWorkspace := ctx.Workspace
	if ctxWorkspace == "" {
		ctxWorkspace = events.DefaultWorkspace
	}
	return projectWorkspace == ctxWorkspace &&
		filepath.Clean(project.RepoRelDir) == filepath.Clean(ctx.RepoRelDir) &&
		project.ProjectName == ctx.ProjectName
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
	ctx.PullStatus.Pull = ctx.Pull
	for _, projectResult := range result.ProjectResults {
		if projectResult.Command != command.Plan && projectResult.Command != command.PolicyCheck && projectResult.Command != command.ApprovePolicies {
			continue
		}
		switch projectResult.Command {
		case command.Plan:
			upsertProjectStatus(ctx.PullStatus, models.ProjectStatus{
				Workspace:    projectResult.Workspace,
				RepoRelDir:   projectResult.RepoRelDir,
				ProjectName:  projectResult.ProjectName,
				PolicyStatus: projectResult.PolicyStatus(),
				Status:       projectResult.PlanStatus(),
			})
		case command.PolicyCheck, command.ApprovePolicies:
			upsertProjectPolicyStatus(ctx.PullStatus, projectResult)
		}
	}
}

func upsertProjectPolicyStatus(pullStatus *models.PullStatus, result command.ProjectResult) {
	status := models.ProjectStatus{
		Workspace:    result.Workspace,
		RepoRelDir:   result.RepoRelDir,
		ProjectName:  result.ProjectName,
		PolicyStatus: result.PolicyStatus(),
	}
	for idx := range pullStatus.Projects {
		project := &pullStatus.Projects[idx]
		if status.Workspace == project.Workspace &&
			status.RepoRelDir == project.RepoRelDir &&
			status.ProjectName == project.ProjectName {
			project.Status = planStatusAfterPolicyCheck(project.Status, result)
			project.PolicyStatus = mergePolicyStatuses(project.PolicyStatus, status.PolicyStatus)
			return
		}
	}
	status.Status = result.PlanStatus()
	pullStatus.Projects = append(pullStatus.Projects, status)
}

func planStatusAfterPolicyCheck(existing models.ProjectPlanStatus, result command.ProjectResult) models.ProjectPlanStatus {
	policyStatus := result.PlanStatus()
	if policyStatus == models.ErroredPolicyCheckStatus {
		return policyStatus
	}
	if existing == models.PlannedPlanStatus || existing == models.ErroredPolicyCheckStatus {
		return policyStatus
	}
	return existing
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
	syntheticNonPR := request.PR <= 0
	if syntheticNonPR {
		pullNum = nextNonPRPullNum()
	}

	baseBranch := apiRequestBaseBranch(request.Ref, request.BaseBranch)
	if !syntheticNonPR && strings.TrimSpace(request.BaseBranch) == "" {
		baseBranch = ""
	}
	pull := models.PullRequest{
		Num:                      pullNum,
		BaseBranch:               baseBranch,
		HeadBranch:               request.Ref,
		HeadCommit:               request.Ref,
		BaseRepo:                 baseRepo,
		HardenedNonPRRefCheckout: syntheticNonPR,
	}
	ctx := &command.Context{
		HeadRepo:                  baseRepo,
		Pull:                      pull,
		Scope:                     a.Scope,
		Log:                       a.Logger,
		API:                       true,
		SkipPRModifiedFiles:       syntheticNonPR,
		SkipPRRequirements:        syntheticNonPR,
		FailOnTeamAllowlistDenied: true,
		RunPolicyChecks:           true,
		SortByExecutionOrder:      true,
		ExactProjectNameMatching:  true,
	}
	a.populatePullRequestStatus(ctx)
	a.populatePullStatus(ctx)
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

func (a *APIController) populatePullStatus(ctx *command.Context) {
	if ctx.Pull.Num <= 0 || a.PullStatusFetcher == nil {
		return
	}

	status, err := a.PullStatusFetcher.GetPullStatus(ctx.Pull)
	if err != nil {
		ctx.PullStatus = nil
		ctx.Log.Warn("unable to fetch pull status: %s", err)
		return
	}
	ctx.PullStatus = status
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
	if request.Action == models.RemediationAutoApply && !a.EnableDriftRemediation {
		responder.ServiceUnavailable(w, r, "drift remediation apply is not enabled")
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
	executionRef := request.Ref
	request.Ref = apiRequestStorageRef(request.Ref)
	request.ExecutionRef = executionRef
	request.BaseBranch = apiRequestBaseBranch(executionRef, request.BaseBranch)
	request.StorageRepository = baseRepo.ID()

	// Create executor that bridges to existing plan/apply infrastructure
	executor := &apiRemediationExecutor{
		controller: a,
		baseRepo:   baseRepo,
		baseBranch: request.BaseBranch,
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

	switch result.Status {
	case models.RemediationStatusFailed:
		if len(result.Projects) == 0 && result.Error != "" {
			responder.InternalError(w, r, errors.New(result.Error))
			return
		}
		responder.Error(w, r, http.StatusConflict, NewAPIError(ErrCodeConflict, "drift remediation failed").WithDetails(apiResult))
	case models.RemediationStatusPartial:
		responder.Success(w, r, http.StatusMultiStatus, apiResult)
	default:
		responder.Success(w, r, http.StatusOK, apiResult)
	}
}

// apiRemediationExecutor implements drift.RemediationExecutor using the API controller's
// existing plan/apply infrastructure.
type apiRemediationExecutor struct {
	controller *APIController
	baseRepo   models.Repo
	baseBranch string
	logger     logging.SimpleLogging
}

// ExecutePlan runs a plan for the given project using the API infrastructure.
func (e *apiRemediationExecutor) ExecutePlan(repository, ref, vcsType, projectName, path, workspace string) (string, *models.DriftSummary, error) {
	// Create a minimal API request for the plan
	request := &APIRequest{
		Repository:       repository,
		Ref:              ref,
		BaseBranch:       e.baseBranchForRef(ref),
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
			Num:                      nextNonPRPullNum(), // Synthetic non-PR workflow ID.
			BaseBranch:               e.baseBranchForRef(ref),
			HeadBranch:               ref,
			HeadCommit:               ref,
			BaseRepo:                 e.baseRepo,
			HardenedNonPRRefCheckout: true,
		},
		Scope:                     e.controller.Scope,
		Log:                       e.logger,
		API:                       true,
		SkipPRModifiedFiles:       true,
		SkipPRRequirements:        true,
		SuppressVCSStatus:         true,
		SuppressJobOutput:         true,
		SuppressApplyWebhooks:     true,
		RunPolicyChecks:           true,
		FailOnTeamAllowlistDenied: true,
		ExactProjectNameMatching:  true,
		SortByExecutionOrder:      true,
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
	ctx.PreWorkflowHooksAlreadyRun = true

	// Execute plan
	result, err := e.controller.apiPlan(request, ctx, true)
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
		BaseBranch:       e.baseBranchForRef(ref),
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
			Num:                      nextNonPRPullNum(), // Synthetic non-PR workflow ID.
			BaseBranch:               e.baseBranchForRef(ref),
			HeadBranch:               ref,
			HeadCommit:               ref,
			BaseRepo:                 e.baseRepo,
			HardenedNonPRRefCheckout: true,
		},
		Scope:                     e.controller.Scope,
		Log:                       e.logger,
		API:                       true,
		SkipPRModifiedFiles:       true,
		SkipPRRequirements:        true,
		SuppressVCSStatus:         true,
		SuppressJobOutput:         true,
		SuppressApplyWebhooks:     true,
		RunPolicyChecks:           true,
		FailOnTeamAllowlistDenied: true,
		FailOnMissingDependencies: true,
		ExactProjectNameMatching:  true,
		SortByExecutionOrder:      true,
	}

	if err := e.ensureApplyUnlocked(); err != nil {
		return nil, err
	}

	if err := e.controller.apiSetup(ctx, command.Apply); err != nil {
		return nil, fmt.Errorf("setup failed: %w", err)
	}
	defer e.controller.cleanupNonPRWorkingDir(ctx)
	if err := validateCachedDriftCommit(projects, ctx.Pull.HeadCommit); err != nil {
		return nil, err
	}

	preHookCmd := &events.CommentCommand{Name: command.Plan}
	if err := e.controller.PreWorkflowHooksCommandRunner.RunPreHooks(ctx, preHookCmd); err != nil {
		if e.controller.FailOnPreWorkflowHookError {
			return nil, fmt.Errorf("pre-workflow hook failed: %w", err)
		}
		e.logger.Warn("pre-workflow hook error (continuing): %v", err)
	}
	ctx.PreWorkflowHooksAlreadyRun = true

	planResult, err := e.controller.apiPlan(request, ctx, true)
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
	ctx.PreWorkflowHooksAlreadyRun = false

	applyResult, err := e.controller.apiApply(request, ctx)
	if err != nil {
		markRunningRemediationResultsFailed(remediationResults, err.Error())
		return remediationResults, err
	}
	remediationResults = mergeApplyRemediationResults(remediationResults, applyResult)
	if applyResult.HasErrors() {
		markRunningRemediationResultsFailed(remediationResults, "apply skipped because an earlier execution group failed")
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
		BaseBranch:       e.baseBranchForRef(ref),
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
			Num:                      nextNonPRPullNum(), // Synthetic non-PR workflow ID.
			BaseBranch:               e.baseBranchForRef(ref),
			HeadBranch:               ref,
			HeadCommit:               ref,
			BaseRepo:                 e.baseRepo,
			HardenedNonPRRefCheckout: true,
		},
		Scope:                     e.controller.Scope,
		Log:                       e.logger,
		API:                       true,
		SkipPRModifiedFiles:       true,
		SkipPRRequirements:        true,
		SuppressVCSStatus:         true,
		SuppressJobOutput:         true,
		SuppressApplyWebhooks:     true,
		RunPolicyChecks:           true,
		FailOnTeamAllowlistDenied: true,
		FailOnMissingDependencies: true,
		ExactProjectNameMatching:  true,
		SortByExecutionOrder:      true,
	}

	if err := e.ensureApplyUnlocked(); err != nil {
		return "", err
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
	ctx.PreWorkflowHooksAlreadyRun = true

	// First run plan (required before apply)
	planResult, err := e.controller.apiPlan(request, ctx, true)
	if err != nil {
		return "", fmt.Errorf("plan failed: %w", err)
	}
	defer e.controller.Locker.UnlockByPull(ctx.HeadRepo.FullName, ctx.Pull.Num) // nolint: errcheck
	if planResult.HasErrors() {
		output, _ := planRemediationOutput(planResult)
		return output.String(), fmt.Errorf("plan had errors")
	}
	seedPullStatusFromPlanResult(ctx, planResult)
	ctx.PreWorkflowHooksAlreadyRun = false

	// Execute apply
	result, err := e.controller.apiApply(request, ctx)
	if err != nil {
		if result != nil {
			output := applyRemediationOutput(result)
			return output.String(), err
		}
		return "", err
	}

	output := applyRemediationOutput(result)

	if result.HasErrors() {
		return output.String(), fmt.Errorf("apply had errors")
	}

	return output.String(), nil
}

func (e *apiRemediationExecutor) baseBranchForRef(ref string) string {
	return apiRequestBaseBranch(ref, e.baseBranch)
}

func validateCachedDriftCommit(projects []models.ProjectDrift, checkedOutCommit string) error {
	if checkedOutCommit == "" {
		return nil
	}
	for _, project := range projects {
		if project.ResolvedCommit == "" {
			if project.DetectionID != "" || !project.LastChecked.IsZero() || project.Drift.HasDrift {
				return fmt.Errorf("cached drift for project %q does not include a resolved commit; rerun drift detection before remediation apply", project.ProjectName)
			}
			continue
		}
		if project.ResolvedCommit != checkedOutCommit {
			return fmt.Errorf("cached drift for project %q was detected at commit %s, but %s now resolves to %s; rerun drift detection before remediation apply", project.ProjectName, project.ResolvedCommit, project.Ref, checkedOutCommit)
		}
	}
	return nil
}

func (e *apiRemediationExecutor) ensureApplyUnlocked() error {
	if e.controller.ApplyLockChecker == nil {
		return nil
	}
	lock, err := e.controller.ApplyLockChecker.CheckApplyLock()
	if err != nil {
		return fmt.Errorf("checking global apply lock: %w", err)
	}
	if lock.Locked {
		return fmt.Errorf("apply is disabled globally")
	}
	return nil
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
		if projectResult.Command == command.Plan {
			planOutput, _ := planProjectRemediationOutput(projectResult)
			remediationResults[idx].PlanOutput = planOutput
		}
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
	for i := range results {
		if results[i].ProjectName != "" || results[i].Path != path {
			continue
		}
		if results[i].Workspace != "" && results[i].Workspace != workspace {
			continue
		}
		results[i].ProjectName = projectName
		if results[i].Workspace == "" {
			results[i].Workspace = workspace
		}
		return i
	}
	for i := range results {
		if results[i].ProjectName == "" || results[i].ProjectName != projectName || results[i].Path != "" {
			continue
		}
		if results[i].Workspace != "" && results[i].Workspace != workspace {
			continue
		}
		results[i].Path = path
		if results[i].Workspace == "" {
			results[i].Workspace = workspace
		}
		return i
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
	baseRepo, ok := a.parseAllowlistedRepo(w, r, repository, vcsType)
	if !ok {
		return
	}

	// Get the result only after the requested repository has been authorized.
	result, err := a.RemediationService.GetResult(id)
	if err != nil {
		responder.NotFound(w, r, fmt.Sprintf("remediation result not found: %v", err))
		return
	}
	if !remediationResultMatchesRepo(result, baseRepo.ID(), repository) {
		responder.NotFound(w, r, "remediation result not found")
		return
	}

	// Convert to API DTO and return
	apiResult := NewRemediationResultAPI(result)
	responder.Success(w, r, http.StatusOK, apiResult)
}

func (a *APIController) parseAllowlistedRepo(w http.ResponseWriter, r *http.Request, repository, vcsType string) (models.Repo, bool) {
	responder := a.getAPIMiddleware().Responder
	VCSHostType, err := models.NewVCSHostType(vcsType)
	if err != nil {
		responder.ValidationFailed(w, r, "invalid VCS type",
			ValidationError{Field: "type", Message: err.Error()})
		return models.Repo{}, false
	}
	if !models.IsSupportedDriftVCSHostType(VCSHostType.String()) {
		responder.ValidationFailed(w, r, "unsupported VCS type",
			ValidationError{Field: "type", Message: "type must be one of: Github, Gitlab, Gitea"})
		return models.Repo{}, false
	}
	if !models.IsValidAPIRepositoryForType(repository, VCSHostType.String()) {
		responder.ValidationFailed(w, r, "invalid repository",
			ValidationError{Field: "repository", Message: "repository must be a valid repository path for the VCS type"})
		return models.Repo{}, false
	}
	cloneURL, err := a.VCSClient.GetCloneURL(a.Logger, VCSHostType, repository)
	if err != nil {
		responder.InternalError(w, r, fmt.Errorf("failed to get clone URL: %w", err))
		return models.Repo{}, false
	}
	baseRepo, err := a.Parser.ParseAPIPlanRequest(VCSHostType, repository, cloneURL)
	if err != nil {
		responder.ValidationFailed(w, r, fmt.Sprintf("failed to parse repository: %v", err))
		return models.Repo{}, false
	}
	if !a.RepoAllowlistChecker.IsAllowlisted(baseRepo.FullName, baseRepo.VCSHost.Hostname) {
		responder.Forbidden(w, r, "repository is not in the allowlist")
		return models.Repo{}, false
	}
	return baseRepo, true
}

func remediationResultMatchesRepo(result *models.RemediationResult, storageRepository, repository string) bool {
	if result.StorageRepository != "" {
		return result.StorageRepository == storageRepository
	}
	return result.Repository == repository
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
	if !models.IsSupportedDriftVCSHostType(VCSHostType.String()) {
		responder.ValidationFailed(w, r, "unsupported VCS type",
			ValidationError{Field: "type", Message: "type must be one of: Github, Gitlab, Gitea"})
		return
	}
	if !models.IsValidAPIRepositoryForType(repository, VCSHostType.String()) {
		responder.ValidationFailed(w, r, "invalid repository",
			ValidationError{Field: "repository", Message: "repository must be a valid repository path for the VCS type"})
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
	if values, ok := r.URL.Query()["limit"]; ok {
		limitStr := ""
		if len(values) > 0 {
			limitStr = values[0]
		}
		parsedLimit, err := strconv.Atoi(strings.TrimSpace(limitStr))
		if err != nil || parsedLimit <= 0 {
			responder.ValidationFailed(w, r, "invalid limit parameter",
				ValidationError{Field: "limit", Message: "must be a positive integer"})
			return
		}
		limit = min(parsedLimit, 100)
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

type driftProjectIdentity struct {
	projectName string
	path        string
	workspace   string
	ref         string
	baseBranch  string
}

func newDriftProjectIdentity(project models.ProjectDrift) driftProjectIdentity {
	return driftProjectIdentity{
		projectName: project.ProjectName,
		path:        project.Path,
		workspace:   project.Workspace,
		ref:         project.Ref,
		baseBranch:  project.BaseBranch,
	}
}

func driftProjectsFromCommandResult(result *command.Result, ref, baseBranch, resolvedCommit, detectionID string) []models.ProjectDrift {
	if result == nil {
		return []models.ProjectDrift{}
	}

	projects := make([]models.ProjectDrift, 0, len(result.ProjectResults))
	indexByIdentity := map[driftProjectIdentity]int{}

	for _, pr := range result.ProjectResults {
		if pr.Command != command.Plan {
			continue
		}
		projectDrift := newProjectDriftFromResult(pr, ref, baseBranch, resolvedCommit, detectionID)
		indexByIdentity[newDriftProjectIdentity(projectDrift)] = len(projects)
		projects = append(projects, projectDrift)
	}

	for _, pr := range result.ProjectResults {
		if pr.Command != command.PolicyCheck {
			continue
		}
		errMessage := projectResultErrorMessage(pr)
		if errMessage == "" {
			continue
		}
		projectDrift := newProjectDriftFromResult(pr, ref, baseBranch, resolvedCommit, detectionID)
		projectDrift.Error = fmt.Sprintf("policy_check failed: %s", errMessage)
		identity := newDriftProjectIdentity(projectDrift)
		if idx, ok := indexByIdentity[identity]; ok {
			if projects[idx].Error == "" {
				projects[idx].Error = projectDrift.Error
			} else {
				projects[idx].Error = projects[idx].Error + "; " + projectDrift.Error
			}
			continue
		}
		indexByIdentity[identity] = len(projects)
		projects = append(projects, projectDrift)
	}

	return projects
}

func newProjectDriftFromResult(pr command.ProjectResult, ref, baseBranch, resolvedCommit, detectionID string) models.ProjectDrift {
	projectDrift := models.ProjectDrift{
		ProjectName:    pr.ProjectName,
		Path:           pr.RepoRelDir,
		Workspace:      pr.Workspace,
		Ref:            ref,
		BaseBranch:     baseBranch,
		ResolvedCommit: resolvedCommit,
		DetectionID:    detectionID,
		LastChecked:    time.Now(),
	}

	if pr.Error != nil {
		projectDrift.Error = pr.Error.Error()
		projectDrift.Drift = models.DriftSummary{HasDrift: false}
	} else if pr.Failure != "" {
		projectDrift.Error = pr.Failure
		projectDrift.Drift = models.DriftSummary{HasDrift: false}
	} else if pr.PlanSuccess != nil {
		projectDrift.Drift = models.NewDriftSummaryFromPlanSuccess(pr.PlanSuccess)
		projectDrift.PlanOutput = pr.PlanSuccess.TerraformOutput
	}

	return projectDrift
}

func projectResultErrorMessage(pr command.ProjectResult) string {
	if pr.Error != nil {
		return pr.Error.Error()
	}
	return pr.Failure
}

func appendDriftProjectError(existing, additional string) string {
	if existing == "" {
		return additional
	}
	return existing + "; " + additional
}

func driftDetectionHasErrors(result *models.DriftDetectionResult) bool {
	if result == nil {
		return false
	}
	for _, project := range result.Projects {
		if project.Error != "" {
			return true
		}
	}
	return false
}

func (a *APIController) reconcileDriftStorage(repository, ref, baseBranch string, detected map[driftProjectIdentity]struct{}, startedAt time.Time) error {
	existing, err := a.DriftStorage.Get(repository, drift.GetOptions{Ref: ref, BaseBranch: baseBranch})
	if err != nil {
		return err
	}

	for _, project := range existing {
		if _, ok := detected[newDriftProjectIdentity(project)]; ok {
			continue
		}
		if project.LastChecked.After(startedAt) {
			continue
		}
		if err := a.DriftStorage.DeleteMatching(repository, drift.GetOptions{
			ProjectName: project.ProjectName,
			Path:        project.Path,
			Workspace:   project.Workspace,
			Ref:         project.Ref,
			BaseBranch:  project.BaseBranch,
			Exact:       true,
		}); err != nil {
			return err
		}
	}
	return nil
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
	normalizedRef := apiRequestStorageRef(request.Ref)
	normalizedBaseBranch := apiRequestBaseBranch(request.Ref, request.BaseBranch)
	fullDetection := len(request.Projects) == 0 && len(request.Paths) == 0
	if fullDetection {
		unlockFullDetection := a.lockFullDriftDetection(baseRepo.ID(), request.Type, normalizedRef, normalizedBaseBranch)
		defer unlockFullDetection()
	}
	detectionStartedAt := time.Now()

	// Build API request for plan
	apiRequest := &APIRequest{
		Repository:       request.Repository,
		Ref:              normalizedRef,
		BaseBranch:       normalizedBaseBranch,
		Type:             request.Type,
		DiscoverProjects: true, // Enable auto-discovery when no projects/paths specified
	}

	if len(request.Projects) > 0 {
		apiRequest.Projects = request.Projects
	} else if len(request.Paths) > 0 {
		for _, p := range request.Paths {
			dir, _ := models.NormalizeAPIPath(p.Directory)
			apiRequest.Paths = append(apiRequest.Paths, APIRequestPath{Directory: dir, Workspace: p.Workspace})
		}
	}

	// Build the command context
	ctx := &command.Context{
		HeadRepo: baseRepo,
		Pull: models.PullRequest{
			Num:                      nextNonPRPullNum(), // Synthetic non-PR workflow ID.
			BaseBranch:               normalizedBaseBranch,
			HeadBranch:               normalizedRef,
			HeadCommit:               request.Ref,
			BaseRepo:                 baseRepo,
			HardenedNonPRRefCheckout: true,
		},
		Scope:                     a.Scope,
		Log:                       a.Logger,
		API:                       true,
		SkipPRModifiedFiles:       true,
		SuppressVCSStatus:         true,
		SuppressJobOutput:         true,
		RunPolicyChecks:           true,
		FailOnTeamAllowlistDenied: true,
		ExactProjectNameMatching:  true,
		SortByExecutionOrder:      true,
	}

	// Setup working directory
	if err := a.apiSetup(ctx, command.Plan); err != nil {
		responder.InternalError(w, r, fmt.Errorf("setup failed: %w", err))
		return
	}
	defer a.cleanupNonPRWorkingDir(ctx)

	// Run pre-workflow hooks before project discovery so hooks can
	// dynamically generate atlantis.yaml or other config files.
	preHookFailed := false
	preHookCmd := &events.CommentCommand{Name: command.Plan}
	if err := a.PreWorkflowHooksCommandRunner.RunPreHooks(ctx, preHookCmd); err != nil {
		preHookFailed = true
		if a.FailOnPreWorkflowHookError {
			responder.InternalError(w, r, fmt.Errorf("pre-workflow hook failed: %w", err))
			return
		}
		a.Logger.Warn("pre-workflow hook error (continuing): %v", err)
	}
	ctx.PreWorkflowHooksAlreadyRun = true

	result, err := a.apiPlan(apiRequest, ctx, true)
	if err != nil {
		if errors.Is(err, events.ErrTeamAllowlistDenied) {
			responder.Forbidden(w, r, err.Error())
			return
		}
		responder.InternalError(w, r, err)
		return
	}
	defer a.Locker.UnlockByPull(ctx.HeadRepo.FullName, ctx.Pull.Num) // nolint: errcheck

	// Process results and store drift data
	detectionResult := models.NewDriftDetectionResult(request.Repository)
	detectedProjects := map[driftProjectIdentity]struct{}{}
	storeFailed := false
	projectDrifts := driftProjectsFromCommandResult(result, normalizedRef, normalizedBaseBranch, ctx.Pull.HeadCommit, detectionResult.ID)
	for _, projectDrift := range projectDrifts {
		detectedProjects[newDriftProjectIdentity(projectDrift)] = struct{}{}
		// PlanOutput is only ever returned in the immediate detect response;
		// Store strips it before persisting, so it is never persisted.
		if err := a.DriftStorage.Store(baseRepo.ID(), projectDrift); err != nil {
			storeFailed = true
			projectDrift.Error = appendDriftProjectError(projectDrift.Error, fmt.Sprintf("storing drift result: %v", err))
			a.Logger.Warn("failed to store drift data: %v", err)
		}
		detectionResult.AddProject(projectDrift)
	}

	if fullDetection && !storeFailed && !preHookFailed && !driftDetectionHasErrors(detectionResult) {
		if err := a.reconcileDriftStorage(baseRepo.ID(), normalizedRef, normalizedBaseBranch, detectedProjects, detectionStartedAt); err != nil {
			a.Logger.Warn("failed to reconcile drift data: %v", err)
		}
	}

	// Send drift webhook notifications for completed detections, including
	// no-drift heartbeat results.
	if a.DriftWebhookSender != nil && !driftDetectionHasErrors(detectionResult) {
		webhookResult := convertToDriftWebhookResult(detectionResult, normalizedRef)
		if err := a.DriftWebhookSender.Send(a.Logger, webhookResult); err != nil {
			a.Logger.Warn("failed to send drift webhook: %v", err)
		}
	}

	// Convert to API DTO and return
	apiResult := NewDriftDetectionResultAPI(detectionResult, request.IncludePlanOutput)

	code := http.StatusOK
	if driftDetectionHasErrors(detectionResult) {
		code = http.StatusMultiStatus // 207 - some projects may have failed
	}
	responder.Success(w, r, code, apiResult)
}

// convertToDriftWebhookResult converts a DriftDetectionResult to a webhook DriftResult.
func convertToDriftWebhookResult(dr *models.DriftDetectionResult, requestRef string) webhooks.DriftResult {
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
	ref := requestRef
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
