// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	onApplyLockProjectName   = "locking-on-apply-preservation"
	onApplyLockApplyCommand  = "atlantis apply -p " + onApplyLockProjectName
	onApplyLockPlanCommand   = "atlantis plan"
	lifecycleCommandMaxPolls = 60
)

var lockConflictCommentSubstrings = []string{
	"currently locked",
	"locked by",
	"delete the lock",
	"apply that plan",
}

type lockPreservationPR struct {
	label      string
	cloneDir   string
	branchName string
	url        string
	pullID     int
}

func (t *E2ETester) runOnApplyLockPreservation(ctx context.Context) (*E2EResult, error) {
	tc := t.testCase
	result := &E2EResult{testCase: tc.Name}
	timestamp := time.Now().UTC().Format("20060102150405")

	pr1, err := t.createOnApplyLockPR(ctx, "pr1", fmt.Sprintf("e2e-lock-pr1-%s", timestamp), tc.MutateFile)
	if err != nil {
		return result, err
	}
	result.pullRequestURL = pr1.url
	result.branchName = pr1.branchName
	defer cleanUpLockPR(ctx, t, pr1)

	log.Printf("[%s] PR1 branch %q pull request %s", tc.Name, pr1.branchName, pr1.url)
	time.Sleep(initialWait)

	state, err := pollAtlantisCommandStatusAfter(ctx, t.vcsClient, pr1.branchName, "plan", tc.Name, CommitStatus{})
	if err != nil {
		return result, err
	}
	result.testResult = state
	if !t.vcsClient.DidAtlantisSucceed(state) {
		return result, fmt.Errorf("[%s] PR1 plan expected success but got %q", tc.Name, state)
	}
	if err := assertProjectStatuses(ctx, t.vcsClient, pr1.branchName, tc); err != nil {
		return result, err
	}

	state, err = t.postAtlantisCommandAndWait(ctx, pr1.pullID, pr1.branchName, tc.Name, "apply", onApplyLockApplyCommand)
	if err != nil {
		return result, err
	}
	result.testResult = state
	if !t.vcsClient.DidAtlantisSucceed(state) {
		return result, fmt.Errorf("[%s] PR1 apply expected success but got %q", tc.Name, state)
	}

	planBaseline, err := t.vcsClient.GetCommitStatus(ctx, pr1.branchName, atlantisCommandStatusContext("plan"))
	if err != nil {
		return result, fmt.Errorf("[%s] getting PR1 plan status before cleanup trigger: %w", tc.Name, err)
	}
	log.Printf("[%s] posting PR1 cleanup-trigger plan command", tc.Name)
	if err := t.vcsClient.PostPRComment(ctx, pr1.pullID, onApplyLockPlanCommand); err != nil {
		return result, fmt.Errorf("[%s] posting PR1 plan command: %w", tc.Name, err)
	}
	state, err = pollAtlantisCommandStatusAfter(ctx, t.vcsClient, pr1.branchName, "plan", tc.Name, planBaseline)
	if err != nil {
		return result, err
	}
	result.testResult = state
	if !t.vcsClient.DidAtlantisSucceed(state) {
		return result, fmt.Errorf("[%s] PR1 cleanup-trigger plan expected success but got %q", tc.Name, state)
	}

	pr2, err := t.createOnApplyLockPR(ctx, "pr2", fmt.Sprintf("e2e-lock-pr2-%s", timestamp), "e2e_pr2_trigger.tf")
	if err != nil {
		return result, err
	}
	defer cleanUpLockPR(ctx, t, pr2)
	log.Printf("[%s] PR2 branch %q pull request %s", tc.Name, pr2.branchName, pr2.url)
	time.Sleep(initialWait)

	state, err = pollAtlantisCommandStatusAfter(ctx, t.vcsClient, pr2.branchName, "plan", tc.Name, CommitStatus{})
	if err != nil {
		return result, err
	}
	result.testResult = state
	if !t.vcsClient.DidAtlantisSucceed(state) {
		return result, fmt.Errorf("[%s] PR2 plan expected success but got %q", tc.Name, state)
	}
	if err := assertProjectStatuses(ctx, t.vcsClient, pr2.branchName, tc); err != nil {
		return result, err
	}

	state, err = t.postAtlantisCommandAndWait(ctx, pr2.pullID, pr2.branchName, tc.Name, "apply", onApplyLockApplyCommand)
	if err != nil {
		return result, err
	}
	result.testResult = state
	if !t.vcsClient.DidAtlantisFail(state) {
		return result, fmt.Errorf("[%s] PR2 apply expected lock-blocked failure but got %q; PR1=%s PR2=%s", tc.Name, state, pr1.url, pr2.url)
	}
	if err := assertLockConflictComment(ctx, t.vcsClient, pr2.pullID, tc.Name); err != nil {
		return result, err
	}

	result.testResult = "success"
	return result, nil
}

func (t *E2ETester) createOnApplyLockPR(ctx context.Context, label, branchName, mutateFile string) (*lockPreservationPR, error) {
	tc := t.testCase
	cloneDir := filepath.Join(t.cloneDirRoot, fmt.Sprintf("%s-%s-test", tc.Name, label))
	log.Printf("[%s] creating %s clone dir %q", tc.Name, label, cloneDir)
	if err := os.MkdirAll(cloneDir, 0700); err != nil {
		return nil, fmt.Errorf("creating clone dir %q: %w", cloneDir, err)
	}
	if err := t.vcsClient.Clone(cloneDir); err != nil {
		return nil, err
	}
	if err := runGit(cloneDir, "checkout", "-b", branchName); err != nil {
		return nil, err
	}

	if mutateFile == "" {
		mutateFile = fmt.Sprintf("%s.tf", tc.Name)
	}
	if err := writeFixtureMutation(cloneDir, tc.Dir, mutateFile, tc.MutateContent); err != nil {
		return nil, err
	}
	if err := enableOnApplyRepoLocksForFixture(cloneDir); err != nil {
		return nil, err
	}
	if err := runGit(cloneDir, "add", filepath.Join(tc.Dir, mutateFile), "atlantis.yaml"); err != nil {
		return nil, err
	}
	if err := runGit(cloneDir, "commit", "-m", fmt.Sprintf("e2e: %s %s", tc.Name, label)); err != nil {
		return nil, err
	}
	if err := runGit(cloneDir, "push", "origin", branchName); err != nil {
		return nil, err
	}

	title := fmt.Sprintf("[E2E] %s %s", tc.Name, label)
	url, pullID, err := t.vcsClient.CreatePullRequest(ctx, title, branchName)
	if err != nil {
		return nil, err
	}
	return &lockPreservationPR{
		label:      label,
		cloneDir:   cloneDir,
		branchName: branchName,
		url:        url,
		pullID:     pullID,
	}, nil
}

func cleanUpLockPR(ctx context.Context, t *E2ETester, pr *lockPreservationPR) {
	if pr == nil {
		return
	}
	if cleanErr := cleanUp(ctx, t, pr.pullID, pr.branchName); cleanErr != nil {
		log.Printf("[%s] cleanup failed for %s branch %q: %v", t.testCase.Name, pr.label, pr.branchName, cleanErr)
	}
}

func writeFixtureMutation(cloneDir, dir, mutateFile, content string) error {
	if content == "" {
		content = defaultMutateContent
	}
	filePath := filepath.Join(cloneDir, dir, mutateFile)
	if err := os.MkdirAll(filepath.Dir(filePath), 0700); err != nil {
		return fmt.Errorf("creating parent dir for %q: %w", filePath, err)
	}
	log.Printf("writing mutation file %q", filePath)
	if err := os.WriteFile(filePath, []byte(content), 0600); err != nil {
		return fmt.Errorf("writing mutation file %q: %w", filePath, err)
	}
	return nil
}

func runGit(dir string, args ...string) error {
	cmd := exec.Command("git", args...) //nolint:gosec // arguments are test-controlled branch/file names.
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, string(output))
	}
	return nil
}

func (t *E2ETester) postAtlantisCommandAndWait(ctx context.Context, pullID int, branchName, caseName, statusCommand, body string) (string, error) {
	statusContext := atlantisCommandStatusContext(statusCommand)
	baseline, err := t.vcsClient.GetCommitStatus(ctx, branchName, statusContext)
	if err != nil {
		return "", fmt.Errorf("[%s] getting baseline %s status: %w", caseName, statusContext, err)
	}
	log.Printf("[%s] posting command to PR %d: %s", caseName, pullID, body)
	if err := t.vcsClient.PostPRComment(ctx, pullID, body); err != nil {
		return "", fmt.Errorf("[%s] posting command %q: %w", caseName, body, err)
	}
	return pollAtlantisCommandStatusAfter(ctx, t.vcsClient, branchName, statusCommand, caseName, baseline)
}

func pollAtlantisCommandStatusAfter(ctx context.Context, client VCSClient, branchName, command, caseName string, baseline CommitStatus) (string, error) {
	statusContext := atlantisCommandStatusContext(command)
	state := "not started"
	var statusErr error
	for i := 0; i < lifecycleCommandMaxPolls; i++ {
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
		if !client.IsAtlantisInProgress(state) {
			return state, nil
		}
	}
	if statusErr != nil {
		return state, fmt.Errorf("[%s] %s timed out after %d polls for branch %q; last state %q; last error: %w", caseName, statusContext, lifecycleCommandMaxPolls, branchName, state, statusErr)
	}
	return state, fmt.Errorf("[%s] %s timed out after %d polls for branch %q; last state %q", caseName, statusContext, lifecycleCommandMaxPolls, branchName, state)
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
	if err := os.WriteFile(path, []byte(patched), info.Mode().Perm()); err != nil {
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
	for i := 0; i < len(lines); i++ {
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

func assertLockConflictComment(ctx context.Context, client VCSClient, pullNumber int, caseName string) error {
	var comments []string
	var commentsErr error
	for i := 0; i < lifecycleCommandMaxPolls; i++ {
		comments, commentsErr = client.GetPRComments(ctx, pullNumber)
		if commentsErr != nil {
			log.Printf("[%s] error polling PR comments for lock conflict: %v", caseName, commentsErr)
		} else if comment, ok := findLockConflictComment(comments); ok {
			log.Printf("[%s] found expected lock conflict comment: %q", caseName, truncateForLog(comment, 160))
			return nil
		}
		time.Sleep(pollInterval)
	}
	if commentsErr != nil {
		return fmt.Errorf("[%s] expected lock conflict comment but comment polling failed: %w", caseName, commentsErr)
	}
	return fmt.Errorf("[%s] expected lock conflict comment containing one of %v not found in %d comments:\n%s", caseName, lockConflictCommentSubstrings, len(comments), strings.Join(comments, "\n---\n"))
}

func findLockConflictComment(comments []string) (string, bool) {
	for _, comment := range comments {
		lower := strings.ToLower(comment)
		for _, candidate := range lockConflictCommentSubstrings {
			if strings.Contains(lower, candidate) {
				return comment, true
			}
		}
	}
	return "", false
}

func truncateForLog(value string, maxLen int) string {
	if len(value) <= maxLen {
		return value
	}
	return value[:maxLen] + "..."
}
