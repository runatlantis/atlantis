// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.

package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"
)

type E2ETester struct {
	vcsClient    VCSClient
	hookID       int64
	cloneDirRoot string
	testCase     TestCase
}

type E2EResult struct {
	testCase       string
	pullRequestURL string
	branchName     string
	testResult     string
	err            error
}

const (
	pollInterval = 2 * time.Second
	maxPolls     = 30
	initialWait  = 2 * time.Second
)

func (t *E2ETester) Start(ctx context.Context) (*E2EResult, error) {
	switch t.testCase.Scenario {
	case ScenarioPlanOnly:
		return t.runPlanOnly(ctx)
	case ScenarioPlanThenApply:
		return t.runPlanThenApply(ctx)
	case ScenarioPlanThenReplanThenApply:
		return t.runPlanThenReplanThenApply(ctx)
	case ScenarioPlanThenApplyExpectFailure:
		return t.runPlanThenApplyExpectFailure(ctx)
	case ScenarioOnApplyLockPreservation:
		return t.runOnApplyLockPreservation(ctx)
	default:
		return nil, fmt.Errorf("unknown E2E scenario %d", t.testCase.Scenario)
	}
}

func (t *E2ETester) runPlanOnly(ctx context.Context) (result *E2EResult, err error) {
	tc := t.testCase
	branchName := fmt.Sprintf("e2e-%s-%s", tc.Name, e2eRunNonce())
	result = &E2EResult{
		testCase:   tc.Name,
		branchName: branchName,
	}
	pull, err := t.createFixturePull(ctx, "", branchName, tc.MutateFile, nil)
	if err != nil {
		return result, err
	}
	result.pullRequestURL = pull.url
	log.Printf("created pull request %s", pull.url)

	defer func() {
		cleanupCtx, cancel := newLifecycleCleanupContext(ctx)
		defer cancel()
		if cleanupErr := cleanUpFixturePull(cleanupCtx, t, pull); cleanupErr != nil {
			err = errors.Join(err, cleanupErr)
		}
	}()

	// Wait for Atlantis to receive webhook and start processing.
	time.Sleep(initialWait)

	state := pollInitialPlanStatus(ctx, t.vcsClient, branchName, tc.Name)
	result.testResult = state

	if err := t.assertPlanResult(ctx, pull, state); err != nil {
		return result, err
	}
	return result, nil
}

func pollInitialPlanStatus(ctx context.Context, client VCSClient, branchName, caseName string) string {
	state := "not started"
	var statusErr error
	for i := 0; i < maxPolls && client.IsAtlantisInProgress(state); i++ {
		select {
		case <-ctx.Done():
			return fmt.Sprintf("canceled (%v)", ctx.Err())
		case <-time.After(pollInterval):
		}
		state, statusErr = client.GetAtlantisStatus(ctx, branchName)
		if statusErr != nil {
			log.Printf("[%s] error polling plan status: %v", caseName, statusErr)
			continue
		}
		if state == "" {
			log.Printf("[%s] atlantis plan has not started yet", caseName)
			continue
		}
		log.Printf("[%s] atlantis plan status: %s", caseName, state)
	}
	if client.IsAtlantisInProgress(state) {
		if statusErr != nil {
			return fmt.Sprintf("timed out (last error: %v)", statusErr)
		}
		return "timed out"
	}
	return state
}

func (t *E2ETester) runPlanThenApply(ctx context.Context) (result *E2EResult, err error) {
	return t.runPlanApplyLifecycle(ctx, false, false)
}

func (t *E2ETester) runPlanThenReplanThenApply(ctx context.Context) (result *E2EResult, err error) {
	return t.runPlanApplyLifecycle(ctx, true, false)
}

func (t *E2ETester) runPlanThenApplyExpectFailure(ctx context.Context) (result *E2EResult, err error) {
	return t.runPlanApplyLifecycle(ctx, false, true)
}

func (t *E2ETester) runPlanApplyLifecycle(ctx context.Context, replan, expectApplyFailure bool) (result *E2EResult, err error) {
	tc := t.testCase
	if tc.ApplyCommand == "" {
		return nil, fmt.Errorf("[%s] plan/apply lifecycle requires ApplyCommand", tc.Name)
	}
	branchName := fmt.Sprintf("e2e-%s-%s", tc.Name, e2eRunNonce())
	result = &E2EResult{testCase: tc.Name, branchName: branchName}
	pull, err := t.createFixturePull(ctx, "", branchName, tc.MutateFile, nil)
	if err != nil {
		return result, err
	}
	result.pullRequestURL = pull.url
	defer func() {
		cleanupCtx, cancel := newLifecycleCleanupContext(ctx)
		defer cancel()
		if cleanupErr := cleanUpFixturePull(cleanupCtx, t, pull); cleanupErr != nil {
			err = errors.Join(err, cleanupErr)
		}
	}()

	log.Printf("[%s] pull request %s", tc.Name, pull.url)
	time.Sleep(initialWait)
	state, pollErr := pollAtlantisCommandStatusAfter(ctx, t.vcsClient, branchName, "plan", tc.Name, CommitStatus{})
	result.testResult = state
	if pollErr != nil {
		return result, pollErr
	}
	if err := t.assertPlanResult(ctx, pull, state); err != nil {
		return result, err
	}
	if replan {
		planBaseline, err := t.vcsClient.GetCommitStatus(ctx, branchName, atlantisCommandStatusContext("plan"))
		if err != nil {
			return result, fmt.Errorf("[%s] fetching plan baseline before replan: %w", tc.Name, err)
		}
		commentBaseline, err := t.vcsClient.GetPRComments(ctx, pull.pullID)
		if err != nil {
			return result, fmt.Errorf("[%s] fetching comments before replan: %w", tc.Name, err)
		}
		if err := t.pushFixtureMutation(pull, tc.ReplanMutateFile, tc.ReplanMutateContent, "generation-2"); err != nil {
			return result, fmt.Errorf("[%s] pushing replan mutation: %w", tc.Name, err)
		}
		state, err = pollAtlantisCommandStatusAfter(ctx, t.vcsClient, branchName, "plan", tc.Name, planBaseline)
		result.testResult = state
		if err != nil {
			return result, t.withLifecycleDiagnostics(ctx, pull, "plan", err)
		}
		if !t.vcsClient.DidAtlantisSucceed(state) {
			return result, t.withLifecycleDiagnostics(ctx, pull, "plan", fmt.Errorf("[%s] replan expected success but got %q", tc.Name, state))
		}
		if len(tc.ExpectedStatusContexts) > 0 {
			if err := assertProjectStatuses(ctx, t.vcsClient, branchName, "plan", tc.Name, tc.ExpectedStatusContexts, tc.ForbidExtraProjectStatuses); err != nil {
				return result, t.withLifecycleDiagnostics(ctx, pull, "plan", err)
			}
		}
		if tc.ExpectedReplanCommentSubstring != "" {
			if err := waitForNewCommentContaining(ctx, t.vcsClient, pull.pullID, tc.Name, commentBaseline, tc.ExpectedReplanCommentSubstring); err != nil {
				return result, t.withLifecycleDiagnostics(ctx, pull, "plan", err)
			}
		}
	}

	commentBaseline, err := t.vcsClient.GetPRComments(ctx, pull.pullID)
	if err != nil {
		return result, fmt.Errorf("[%s] fetching comments before apply: %w", tc.Name, err)
	}
	expectedApplyState := t.vcsClient.DidAtlantisSucceed
	if expectApplyFailure {
		expectedApplyState = t.vcsClient.DidAtlantisFail
	}
	state, applyErr := t.postAtlantisCommandAndWaitForExpectedState(
		ctx,
		pull.pullID,
		branchName,
		tc.Name,
		"apply",
		tc.ApplyCommand,
		expectedApplyState,
	)
	result.testResult = state
	if applyErr != nil {
		return result, t.withLifecycleDiagnostics(ctx, pull, "apply", applyErr)
	}
	if expectApplyFailure {
		if !t.vcsClient.DidAtlantisFail(state) {
			return result, t.withLifecycleDiagnostics(ctx, pull, "apply", fmt.Errorf("[%s] apply expected failure but got %q", tc.Name, state))
		}
		if err := assertFailedProjectStatuses(ctx, t.vcsClient, branchName, "apply", tc.Name, tc.ExpectedFailedApplyStatusContexts); err != nil {
			return result, t.withLifecycleDiagnostics(ctx, pull, "apply", err)
		}
	} else {
		if !t.vcsClient.DidAtlantisSucceed(state) {
			return result, t.withLifecycleDiagnostics(ctx, pull, "apply", fmt.Errorf("[%s] apply expected success but got %q", tc.Name, state))
		}
		if len(tc.ExpectedApplyStatusContexts) > 0 {
			if err := assertProjectStatuses(ctx, t.vcsClient, branchName, "apply", tc.Name, tc.ExpectedApplyStatusContexts, false); err != nil {
				return result, t.withLifecycleDiagnostics(ctx, pull, "apply", err)
			}
		}
	}
	if tc.ExpectedApplyCommentSubstring != "" {
		if err := waitForNewCommentContaining(ctx, t.vcsClient, pull.pullID, tc.Name, commentBaseline, tc.ExpectedApplyCommentSubstring); err != nil {
			return result, t.withLifecycleDiagnostics(ctx, pull, "apply", err)
		}
	}
	if tc.ForbiddenApplyCommentSubstring != "" {
		comments, err := t.vcsClient.GetPRComments(ctx, pull.pullID)
		if err != nil {
			return result, t.withLifecycleDiagnostics(ctx, pull, "apply", fmt.Errorf("[%s] fetching comments for forbidden marker: %w", tc.Name, err))
		}
		if newCommentContains(comments, commentBaseline, tc.ForbiddenApplyCommentSubstring) {
			return result, t.withLifecycleDiagnostics(ctx, pull, "apply", fmt.Errorf("[%s] forbidden apply marker %q appeared after command", tc.Name, tc.ForbiddenApplyCommentSubstring))
		}
	}
	result.testResult = "success"
	return result, nil
}

func (t *E2ETester) assertPlanResult(ctx context.Context, pull *fixturePull, state string) error {
	tc := t.testCase
	if tc.ExpectFailure {
		if !t.vcsClient.DidAtlantisFail(state) {
			return fmt.Errorf("[%s] expected failure but got %q", tc.Name, state)
		}
	} else if !t.vcsClient.DidAtlantisSucceed(state) {
		return fmt.Errorf("[%s] expected success but got %q", tc.Name, state)
	}
	if len(tc.ExpectedStatusContexts) > 0 {
		if err := assertProjectStatuses(ctx, t.vcsClient, pull.branchName, "plan", tc.Name, tc.ExpectedStatusContexts, tc.ForbidExtraProjectStatuses); err != nil {
			return err
		}
	}
	if tc.ExpectedCommentSubstring != "" {
		return assertCommentContains(ctx, t.vcsClient, pull.pullID, tc.Name, tc.ExpectedCommentSubstring)
	}
	return nil
}

// assertProjectStatuses verifies that the expected per-project status contexts
// are present and (optionally) that no unexpected project statuses exist.
func assertProjectStatuses(ctx context.Context, client VCSClient, branchName, command, caseName string, expectedContexts []string, forbidExtra bool) error {
	projectStatuses, err := client.GetProjectStatuses(ctx, branchName, command)
	if err != nil {
		return fmt.Errorf("[%s] failed to get %s project statuses: %v", caseName, command, err)
	}

	// GetProjectStatuses returns nil on GitLab (unsupported).
	if projectStatuses == nil {
		log.Printf("[%s] skipping %s project status assertion (not supported on this VCS)", caseName, command)
		return nil
	}

	// Check all expected contexts are present and successful.
	for _, expected := range expectedContexts {
		state, ok := projectStatuses[expected]
		if !ok {
			var found []string
			for k := range projectStatuses {
				found = append(found, k)
			}
			sort.Strings(found)
			return fmt.Errorf("[%s] expected status context %q not found. Found: %v",
				caseName, expected, found)
		}
		if state != "success" {
			return fmt.Errorf("[%s] status context %q has state %q, expected success",
				caseName, expected, state)
		}
	}

	// Check no unexpected project statuses appear.
	if forbidExtra {
		expectedSet := make(map[string]bool)
		for _, e := range expectedContexts {
			expectedSet[e] = true
		}
		for statusCtx := range projectStatuses {
			if !expectedSet[statusCtx] {
				return fmt.Errorf("[%s] unexpected project status %q (expected only %v)",
					caseName, statusCtx, expectedContexts)
			}
		}
	}

	log.Printf("[%s] %s project status contexts verified: %v", caseName, command, expectedContexts)
	return nil
}

func assertFailedProjectStatuses(ctx context.Context, client VCSClient, branchName, command, caseName string, expectedContexts []string) error {
	if len(expectedContexts) == 0 {
		return nil
	}
	projectStatuses, err := client.GetProjectStatuses(ctx, branchName, command)
	if err != nil {
		return fmt.Errorf("[%s] failed to get %s project statuses: %v", caseName, command, err)
	}
	if projectStatuses == nil {
		log.Printf("[%s] skipping failed %s project status assertion (not supported on this VCS)", caseName, command)
		return nil
	}
	for _, expected := range expectedContexts {
		state, ok := projectStatuses[expected]
		if !ok {
			return fmt.Errorf("[%s] expected failed status context %q not found", caseName, expected)
		}
		if !client.DidAtlantisFail(state) {
			return fmt.Errorf("[%s] status context %q has state %q, expected failure", caseName, expected, state)
		}
	}
	return nil
}

func assertCommentContains(ctx context.Context, client VCSClient, pullNumber int, caseName, expected string) error {
	comments, err := client.GetPRComments(ctx, pullNumber)
	if err != nil {
		return fmt.Errorf("[%s] failed to fetch comments for marker assertion: %v", caseName, err)
	}
	for _, body := range comments {
		if strings.Contains(body, expected) {
			log.Printf("[%s] found expected marker in comment: %q", caseName, expected)
			return nil
		}
	}
	return fmt.Errorf("[%s] expected comment containing %q not found in %d comments", caseName, expected, len(comments))
}

func waitForNewCommentContaining(ctx context.Context, client VCSClient, pullNumber int, caseName string, baseline []string, expected string) error {
	var lastErr error
	for range maxPolls {
		comments, err := client.GetPRComments(ctx, pullNumber)
		if err != nil {
			lastErr = err
		} else if newCommentContains(comments, baseline, expected) {
			log.Printf("[%s] found expected marker in a new comment: %q", caseName, expected)
			return nil
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("[%s] waiting for new comment containing %q: %w", caseName, expected, ctx.Err())
		case <-time.After(pollInterval):
		}
	}
	if lastErr != nil {
		return fmt.Errorf("[%s] expected new comment containing %q not found after %d polls; last error: %w", caseName, expected, maxPolls, lastErr)
	}
	return fmt.Errorf("[%s] expected new comment containing %q not found after %d polls", caseName, expected, maxPolls)
}

func newCommentContains(comments, baseline []string, expected string) bool {
	remainingBaseline := make(map[string]int, len(baseline))
	for _, comment := range baseline {
		remainingBaseline[comment]++
	}
	for _, comment := range comments {
		if remainingBaseline[comment] > 0 {
			remainingBaseline[comment]--
			continue
		}
		if strings.Contains(comment, expected) {
			return true
		}
	}
	return false
}

func (t *E2ETester) withLifecycleDiagnostics(ctx context.Context, pull *fixturePull, command string, cause error) error {
	diagnostics := []string{"PR=" + pull.url}
	status, err := t.vcsClient.GetCommitStatus(ctx, pull.branchName, atlantisCommandStatusContext(command))
	if err != nil {
		diagnostics = append(diagnostics, "aggregate_status_error="+err.Error())
	} else {
		diagnostics = append(diagnostics, fmt.Sprintf("aggregate_status=%q id=%d updated=%s", status.State, status.ID, status.UpdatedAt.UTC().Format(time.RFC3339Nano)))
	}
	projectStatuses, err := t.vcsClient.GetProjectStatuses(ctx, pull.branchName, command)
	if err != nil {
		diagnostics = append(diagnostics, "project_status_error="+err.Error())
	} else {
		keys := make([]string, 0, len(projectStatuses))
		for statusContext := range projectStatuses {
			keys = append(keys, statusContext)
		}
		sort.Strings(keys)
		var statuses []string
		for _, statusContext := range keys {
			statuses = append(statuses, fmt.Sprintf("%s=%s", statusContext, projectStatuses[statusContext]))
		}
		diagnostics = append(diagnostics, "project_statuses=["+strings.Join(statuses, ", ")+"]")
	}
	comments, err := t.vcsClient.GetPRComments(ctx, pull.pullID)
	if err != nil {
		diagnostics = append(diagnostics, "comments_error="+err.Error())
	} else {
		if len(comments) > 5 {
			comments = comments[len(comments)-5:]
		}
		for i := range comments {
			comments[i] = strings.ReplaceAll(comments[i], "\n", " ")
			if len(comments[i]) > 500 {
				comments[i] = comments[i][:500] + "..."
			}
		}
		diagnostics = append(diagnostics, fmt.Sprintf("recent_comments=%q", comments))
	}
	return fmt.Errorf("%w; %s", cause, strings.Join(diagnostics, "; "))
}
