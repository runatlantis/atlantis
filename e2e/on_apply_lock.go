// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"
)

const (
	onApplyLockProjectName   = "locking-on-apply-preservation"
	onApplyLockApplyCommand  = "atlantis apply -p " + onApplyLockProjectName
	onApplyLockPlanCommand   = "atlantis plan"
	lifecycleCommandMaxPolls = 60
	lifecycleCleanupTimeout  = 30 * time.Second
)

var e2eNonceCounter atomic.Uint64

func (t *E2ETester) runOnApplyLockPreservation(ctx context.Context) (result *E2EResult, err error) {
	tc := t.testCase
	result = &E2EResult{testCase: tc.Name}
	var prs []*fixturePull
	defer func() {
		cleanupCtx, cancel := newLifecycleCleanupContext(ctx)
		defer cancel()

		cleanupErr := cleanUpFixturePulls(cleanupCtx, t, prs...)
		switch {
		case err != nil && cleanupErr != nil:
			err = fmt.Errorf("%w; cleanup failed: %v", err, cleanupErr)
		case cleanupErr != nil:
			err = cleanupErr
		}
	}()

	err = t.runOnApplyLockPreservationBody(ctx, result, &prs)
	return result, err
}

func (t *E2ETester) runOnApplyLockPreservationBody(ctx context.Context, result *E2EResult, prs *[]*fixturePull) error {
	tc := t.testCase
	nonce := e2eRunNonce()

	pr1, err := t.createOnApplyLockPR(ctx, "pr1", fmt.Sprintf("e2e-lock-pr1-%s", nonce), tc.MutateFile)
	if err != nil {
		return err
	}
	*prs = append(*prs, pr1)
	result.pullRequestURL = pr1.url
	result.branchName = pr1.branchName

	log.Printf("[%s] PR1 branch %q pull request %s", tc.Name, pr1.branchName, pr1.url)
	time.Sleep(initialWait)

	state, err := pollAtlantisCommandStatusAfter(ctx, t.vcsClient, pr1.branchName, "plan", tc.Name, CommitStatus{})
	if err != nil {
		return err
	}
	result.testResult = state
	if !t.vcsClient.DidAtlantisSucceed(state) {
		return fmt.Errorf("[%s] PR1 plan expected success but got %q", tc.Name, state)
	}
	if err := assertProjectStatuses(ctx, t.vcsClient, pr1.branchName, "plan", tc.Name, tc.ExpectedStatusContexts, tc.ForbidExtraProjectStatuses); err != nil {
		return err
	}

	state, err = t.postAtlantisCommandAndWait(ctx, pr1.pullID, pr1.branchName, tc.Name, "apply", onApplyLockApplyCommand)
	if err != nil {
		return err
	}
	result.testResult = state
	if !t.vcsClient.DidAtlantisSucceed(state) {
		return fmt.Errorf("[%s] PR1 apply expected success but got %q", tc.Name, state)
	}

	planBaseline, err := t.vcsClient.GetCommitStatus(ctx, pr1.branchName, atlantisCommandStatusContext("plan"))
	if err != nil {
		return fmt.Errorf("[%s] getting PR1 plan status before cleanup trigger: %w", tc.Name, err)
	}
	projectPlanBaseline, err := t.vcsClient.GetCommitStatus(ctx, pr1.branchName, onApplyLockProjectPlanStatusContext())
	if err != nil {
		return fmt.Errorf("[%s] getting PR1 project plan status before cleanup trigger: %w", tc.Name, err)
	}
	log.Printf("[%s] posting PR1 cleanup-trigger plan command", tc.Name)
	if err := t.vcsClient.PostPRComment(ctx, pr1.pullID, onApplyLockPlanCommand); err != nil {
		return fmt.Errorf("[%s] posting PR1 plan command: %w", tc.Name, err)
	}
	state, err = pollAtlantisCommandStatusAfter(ctx, t.vcsClient, pr1.branchName, "plan", tc.Name, planBaseline)
	if err != nil {
		return err
	}
	result.testResult = state
	if !t.vcsClient.DidAtlantisSucceed(state) {
		return fmt.Errorf("[%s] PR1 cleanup-trigger plan expected success but got %q", tc.Name, state)
	}
	projectPlanState, err := pollCommitStatusAfter(ctx, t.vcsClient, pr1.branchName, onApplyLockProjectPlanStatusContext(), tc.Name, projectPlanBaseline)
	if err != nil {
		return err
	}
	if !t.vcsClient.DidAtlantisSucceed(projectPlanState) {
		return fmt.Errorf("[%s] PR1 cleanup-trigger project plan for %s expected success but got %q", tc.Name, onApplyLockProjectName, projectPlanState)
	}

	pr2, err := t.createOnApplyLockPR(ctx, "pr2", fmt.Sprintf("e2e-lock-pr2-%s", nonce), "e2e_pr2_trigger.tf")
	if err != nil {
		return err
	}
	*prs = append(*prs, pr2)
	result.pullRequestURL = fmt.Sprintf("PR1: %s PR2: %s", pr1.url, pr2.url)
	log.Printf("[%s] PR2 branch %q pull request %s", tc.Name, pr2.branchName, pr2.url)
	time.Sleep(initialWait)

	state, err = pollAtlantisCommandStatusAfter(ctx, t.vcsClient, pr2.branchName, "plan", tc.Name, CommitStatus{})
	if err != nil {
		return err
	}
	result.testResult = state
	if !t.vcsClient.DidAtlantisSucceed(state) {
		return fmt.Errorf("[%s] PR2 plan expected success but got %q", tc.Name, state)
	}
	if err := assertProjectStatuses(ctx, t.vcsClient, pr2.branchName, "plan", tc.Name, tc.ExpectedStatusContexts, tc.ForbidExtraProjectStatuses); err != nil {
		return err
	}

	state, err = t.postAtlantisCommandAndWait(ctx, pr2.pullID, pr2.branchName, tc.Name, "apply", onApplyLockApplyCommand)
	if err != nil {
		return err
	}
	result.testResult = state
	if !t.vcsClient.DidAtlantisFail(state) {
		return fmt.Errorf("[%s] PR2 apply expected lock-blocked failure but got %q; PR1=%s PR2=%s", tc.Name, state, pr1.url, pr2.url)
	}
	if err := assertLockConflictComment(ctx, t.vcsClient, pr2.pullID, tc.Name, pr1.pullID); err != nil {
		return err
	}

	result.testResult = "success"
	return nil
}

func (t *E2ETester) createOnApplyLockPR(ctx context.Context, label, branchName, mutateFile string) (*fixturePull, error) {
	return t.createFixturePull(ctx, label, branchName, mutateFile, enableOnApplyRepoLocksForFixture)
}

func newLifecycleCleanupContext(parent context.Context) (context.Context, context.CancelFunc) {
	base := context.Background()
	if parent != nil {
		base = context.WithoutCancel(parent)
	}
	return context.WithTimeout(base, lifecycleCleanupTimeout)
}

func e2eRunNonce() string {
	now := time.Now().UTC().UnixNano()
	counter := e2eNonceCounter.Add(1)
	if runID := os.Getenv("GITHUB_RUN_ID"); runID != "" {
		if attempt := os.Getenv("GITHUB_RUN_ATTEMPT"); attempt != "" {
			return fmt.Sprintf("%s-%s-%d-%d", runID, attempt, now, counter)
		}
		return fmt.Sprintf("%s-%d-%d", runID, now, counter)
	}
	return fmt.Sprintf("%d-%d-%d", now, os.Getpid(), counter)
}

func onApplyLockProjectPlanStatusContext() string {
	return projectStatusPrefix + onApplyLockProjectName
}

func (t *E2ETester) postAtlantisCommandAndWait(ctx context.Context, pullID int, branchName, caseName, statusCommand, body string) (string, error) {
	return t.postAtlantisCommandAndWaitForExpectedState(ctx, pullID, branchName, caseName, statusCommand, body, nil)
}

func (t *E2ETester) postAtlantisCommandAndWaitForExpectedState(
	ctx context.Context,
	pullID int,
	branchName, caseName, statusCommand, body string,
	expectedState func(string) bool,
) (string, error) {
	statusContext := atlantisCommandStatusContext(statusCommand)
	baseline, err := t.vcsClient.GetCommitStatus(ctx, branchName, statusContext)
	if err != nil {
		return "", fmt.Errorf("[%s] getting baseline %s status: %w", caseName, statusContext, err)
	}
	log.Printf("[%s] posting command to PR %d: %s", caseName, pullID, body)
	if err := t.vcsClient.PostPRComment(ctx, pullID, body); err != nil {
		return "", fmt.Errorf("[%s] posting command %q: %w", caseName, body, err)
	}
	return pollAtlantisCommandStatusAfterExpectedState(ctx, t.vcsClient, branchName, statusCommand, caseName, baseline, expectedState)
}

func pollAtlantisCommandStatusAfter(ctx context.Context, client VCSClient, branchName, command, caseName string, baseline CommitStatus) (string, error) {
	return pollAtlantisCommandStatusAfterExpectedState(ctx, client, branchName, command, caseName, baseline, nil)
}

func pollAtlantisCommandStatusAfterExpectedState(
	ctx context.Context,
	client VCSClient,
	branchName, command, caseName string,
	baseline CommitStatus,
	expectedState func(string) bool,
) (string, error) {
	statusContext := atlantisCommandStatusContext(command)
	return pollCommitStatusAfterExpectedState(ctx, client, branchName, statusContext, caseName, baseline, expectedState)
}

func pollCommitStatusAfter(
	ctx context.Context,
	client VCSClient,
	branchName, statusContext, caseName string,
	baseline CommitStatus,
) (string, error) {
	return pollCommitStatusAfterExpectedState(ctx, client, branchName, statusContext, caseName, baseline, nil)
}

func pollCommitStatusAfterExpectedState(
	ctx context.Context,
	client VCSClient,
	branchName, statusContext, caseName string,
	baseline CommitStatus,
	expectedState func(string) bool,
) (string, error) {
	state := "not started"
	var statusErr error
	for range lifecycleCommandMaxPolls {
		time.Sleep(pollInterval)
		status, err := client.GetCommitStatus(ctx, branchName, statusContext)
		if err != nil {
			statusErr = err
			log.Printf("[%s] error polling %s status for branch %q: %v", caseName, statusContext, branchName, err)
			continue
		}
		if status.State == "" {
			log.Printf("[%s] %s has not started yet on branch %q", caseName, statusContext, branchName)
			continue
		}
		state = status.State
		if !isNewCommitStatus(status, baseline) {
			log.Printf("[%s] %s status is still from previous run: state=%s id=%d", caseName, statusContext, status.State, status.ID)
			continue
		}
		log.Printf("[%s] %s status on branch %q: %s", caseName, statusContext, branchName, state)
		inProgress := client.IsAtlantisInProgress(state)
		if shouldReturnCommitStatus(state, inProgress, expectedState) {
			return state, nil
		}
		if !inProgress {
			log.Printf("[%s] %s status %q does not match the expected terminal state; continuing to poll", caseName, statusContext, state)
		}
	}
	if statusErr != nil {
		return state, fmt.Errorf("[%s] %s timed out after %d polls for branch %q; last state %q; last error: %w", caseName, statusContext, lifecycleCommandMaxPolls, branchName, state, statusErr)
	}
	return state, fmt.Errorf("[%s] %s timed out after %d polls for branch %q; last state %q", caseName, statusContext, lifecycleCommandMaxPolls, branchName, state)
}

func shouldReturnCommitStatus(state string, inProgress bool, expectedState func(string) bool) bool {
	if inProgress {
		return false
	}
	return expectedState == nil || expectedState(state)
}

func isNewCommitStatus(status, baseline CommitStatus) bool {
	if status.State == "" {
		return false
	}
	if baseline.ID == 0 && baseline.UpdatedAt.IsZero() {
		return true
	}
	if baseline.ID != 0 && status.ID != baseline.ID {
		return true
	}
	if !baseline.UpdatedAt.IsZero() && !status.UpdatedAt.IsZero() && status.UpdatedAt.After(baseline.UpdatedAt) {
		return true
	}
	return false
}

func enableOnApplyRepoLocksForFixture(cloneDir string) error {
	path := filepath.Join(cloneDir, "atlantis.yaml")
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("stat atlantis.yaml: %w", err)
	}
	contents, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading atlantis.yaml: %w", err)
	}
	patched, err := enableOnApplyRepoLocksForFixtureContent(string(contents))
	if err != nil {
		return err
	}
	if patched == string(contents) {
		return nil
	}
	if err := os.WriteFile(path, []byte(patched), info.Mode().Perm()); err != nil { //nolint:gosec // path is the fixed atlantis.yaml file in an E2E temp clone.
		return fmt.Errorf("writing atlantis.yaml: %w", err)
	}
	return nil
}

func enableOnApplyRepoLocksForFixtureContent(contents string) (string, error) {
	lines := strings.Split(contents, "\n")
	start, end, err := findProjectBlock(lines, onApplyLockProjectName)
	if err != nil {
		return "", err
	}
	for i := start; i < end; i++ {
		if strings.TrimSpace(lines[i]) != "repo_locks:" {
			continue
		}
		if i+1 < end && strings.TrimSpace(lines[i+1]) == "mode: on_apply" {
			return contents, nil
		}
		return "", fmt.Errorf("project %q already contains repo_locks but not mode: on_apply", onApplyLockProjectName)
	}

	insertAt := -1
	for i := start; i < end; i++ {
		if strings.TrimSpace(lines[i]) == "workspace: default" {
			insertAt = i + 1
			break
		}
	}
	if insertAt == -1 {
		return "", fmt.Errorf("project %q is missing workspace: default", onApplyLockProjectName)
	}

	insert := []string{
		"    repo_locks:",
		"      mode: on_apply",
	}
	patched := make([]string, 0, len(lines)+len(insert))
	patched = append(patched, lines[:insertAt]...)
	patched = append(patched, insert...)
	patched = append(patched, lines[insertAt:]...)
	return strings.Join(patched, "\n"), nil
}

func findProjectBlock(lines []string, projectName string) (int, int, error) {
	start := -1
	end := -1
	for i := range len(lines) {
		if !strings.HasPrefix(lines[i], "  - ") {
			continue
		}
		candidateEnd := len(lines)
		for j := i + 1; j < len(lines); j++ {
			if strings.HasPrefix(lines[j], "  - ") || isTopLevelYAMLSection(lines[j]) {
				candidateEnd = j
				break
			}
		}
		for j := i; j < candidateEnd; j++ {
			if strings.TrimSpace(lines[j]) != "name: "+projectName {
				continue
			}
			if start != -1 {
				return 0, 0, fmt.Errorf("found multiple project entries named %q", projectName)
			}
			start = i
			end = candidateEnd
			break
		}
	}
	if start == -1 {
		return 0, 0, fmt.Errorf("project %q not found in atlantis.yaml", projectName)
	}
	return start, end, nil
}

func isTopLevelYAMLSection(line string) bool {
	return line != "" && !strings.HasPrefix(line, " ") && strings.HasSuffix(strings.TrimSpace(line), ":")
}

func assertLockConflictComment(ctx context.Context, client VCSClient, pullNumber int, caseName string, lockOwnerPullNumber int) error {
	var comments []string
	var commentsErr error
	for range lifecycleCommandMaxPolls {
		comments, commentsErr = client.GetPRComments(ctx, pullNumber)
		if commentsErr != nil {
			log.Printf("[%s] error polling PR comments for lock conflict: %v", caseName, commentsErr)
		} else if comment, ok := findLockConflictComment(comments, lockOwnerPullNumber); ok {
			log.Printf("[%s] found expected lock conflict comment: %q", caseName, truncateForLog(comment, 160))
			return nil
		}
		time.Sleep(pollInterval)
	}
	if commentsErr != nil {
		return fmt.Errorf("[%s] expected lock conflict comment for lock owner pull #%d on pull #%d but comment polling failed: %w", caseName, lockOwnerPullNumber, pullNumber, commentsErr)
	}
	return fmt.Errorf("[%s] expected repo-lock conflict comment for lock owner pull #%d on pull #%d not found in %d comments:\n%s", caseName, lockOwnerPullNumber, pullNumber, len(comments), strings.Join(comments, "\n---\n"))
}

func findLockConflictComment(comments []string, lockOwnerPullNumber int) (string, bool) {
	lockPhrase := "this project is currently locked by an unapplied plan from pull "
	deletePhrase := "delete the lock from "
	applyPhrase := "apply that plan and merge the pull request"
	for _, comment := range comments {
		lower := strings.ToLower(comment)
		if !containsExactPullRefAfterPhrase(comment, lockPhrase, lockOwnerPullNumber) {
			continue
		}
		if containsExactPullRefAfterPhrase(comment, deletePhrase, lockOwnerPullNumber) || strings.Contains(lower, applyPhrase) {
			return comment, true
		}
	}
	return "", false
}

func containsExactPullRefAfterPhrase(comment, phrase string, pullNumber int) bool {
	lower := strings.ToLower(comment)
	target := strings.ToLower(phrase) + fmt.Sprintf("#%d", pullNumber)
	idx := 0
	for {
		pos := strings.Index(lower[idx:], target)
		if pos == -1 {
			return false
		}
		end := idx + pos + len(target)
		if hasPullRefNumericBoundary(comment, end) {
			return true
		}
		idx = end
	}
}

func containsExactPullRef(comment string, pullNumber int) bool {
	ref := fmt.Sprintf("#%d", pullNumber)
	idx := 0
	for {
		pos := strings.Index(comment[idx:], ref)
		if pos == -1 {
			return false
		}
		end := idx + pos + len(ref)
		if hasPullRefNumericBoundary(comment, end) {
			return true
		}
		idx = end
	}
}

func hasPullRefNumericBoundary(comment string, end int) bool {
	return end == len(comment) || comment[end] < '0' || comment[end] > '9'
}

func truncateForLog(value string, maxLen int) string {
	if len(value) <= maxLen {
		return value
	}
	return value[:maxLen] + "..."
}
