// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.

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
	tc := t.testCase
	cloneDir := fmt.Sprintf("%s/%s-test", t.cloneDirRoot, tc.Name)
	branchName := fmt.Sprintf("e2e-%s-%s", tc.Name, time.Now().Format("20060102150405"))

	result := &E2EResult{
		testCase:   tc.Name,
		branchName: branchName,
	}

	log.Printf("creating dir %q", cloneDir)
	if err := os.MkdirAll(cloneDir, 0700); err != nil {
		return result, fmt.Errorf("failed to create dir %q: %v", cloneDir, err)
	}

	if err := t.vcsClient.Clone(cloneDir); err != nil {
		return result, err
	}

	log.Printf("checking out branch %q", branchName)
	checkoutCmd := exec.Command("git", "checkout", "-b", branchName)
	checkoutCmd.Dir = cloneDir
	if output, err := checkoutCmd.CombinedOutput(); err != nil {
		return result, fmt.Errorf("failed to git checkout branch %q: %v: %s", branchName, err, string(output))
	}

	// Determine file to mutate.
	mutateFile := tc.MutateFile
	if mutateFile == "" {
		mutateFile = fmt.Sprintf("%s.tf", tc.Name)
	}
	filePath := filepath.Join(cloneDir, tc.Dir, mutateFile)

	// Ensure parent directory exists (for cases writing to subdirs).
	if err := os.MkdirAll(filepath.Dir(filePath), 0700); err != nil {
		return result, fmt.Errorf("failed to create parent dir for %q: %v", filePath, err)
	}

	content := tc.MutateContent
	if content == "" {
		content = defaultMutateContent
	}

	log.Printf("writing mutation file %q", filePath)
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return result, fmt.Errorf("couldn't write file %s: %v", filePath, err)
	}

	log.Printf("git add %q", filePath)
	addCmd := exec.Command("git", "add", filePath)
	addCmd.Dir = cloneDir
	if output, err := addCmd.CombinedOutput(); err != nil {
		return result, fmt.Errorf("failed to git add file %q: %v: %s", filePath, err, string(output))
	}

	log.Printf("git commit")
	commitCmd := exec.Command("git", "commit", "-am", fmt.Sprintf("e2e: %s", tc.Name))
	commitCmd.Dir = cloneDir
	if output, err := commitCmd.CombinedOutput(); err != nil {
		return result, fmt.Errorf("failed to git commit in %q: %v: %s", cloneDir, err, string(output))
	}

	log.Printf("git push branch %q", branchName)
	pushCmd := exec.Command("git", "push", "origin", branchName)
	pushCmd.Dir = cloneDir
	if output, err := pushCmd.CombinedOutput(); err != nil {
		return result, fmt.Errorf("failed to git push branch %q: %v: %s", branchName, err, string(output))
	}

	title := fmt.Sprintf("[E2E] %s", tc.Name)
	url, pullID, err := t.vcsClient.CreatePullRequest(ctx, title, branchName)
	if err != nil {
		return result, err
	}
	result.pullRequestURL = url
	log.Printf("created pull request %s", url)

	defer func() {
		if cleanErr := cleanUp(ctx, t, pullID, branchName); cleanErr != nil {
			log.Printf("cleanup failed: %v", cleanErr)
		}
	}()

	// Wait for Atlantis to receive webhook and start processing.
	time.Sleep(initialWait)

	statusPrefix := tc.StatusPrefix
	if statusPrefix == "" {
		statusPrefix = "atlantis/plan"
	}

	state := "not started"
	var statusErr error
	i := 0
	for ; i < maxPolls && t.vcsClient.IsAtlantisInProgress(state); i++ {
		time.Sleep(pollInterval)
		state, statusErr = t.vcsClient.GetAtlantisStatus(ctx, branchName, statusPrefix, tc.ExpectedStatusCount)
		if statusErr != nil {
			log.Printf("[%s] error polling status: %v", tc.Name, statusErr)
			continue
		}
		if state == "" {
			log.Printf("[%s] atlantis run hasn't started yet", tc.Name)
			continue
		}
		log.Printf("[%s] atlantis status: %s", tc.Name, state)
	}
	if i == maxPolls {
		if statusErr != nil {
			state = fmt.Sprintf("timed out (last error: %v)", statusErr)
		} else {
			state = "timed out"
		}
	}

	log.Printf("[%s] final status: %q", tc.Name, state)
	result.testResult = state

	// Evaluate result against expectations.
	if tc.ExpectFailure {
		if !t.vcsClient.DidAtlantisFail(state) {
			return result, fmt.Errorf("[%s] expected failure but got %q", tc.Name, state)
		}
	} else {
		if !t.vcsClient.DidAtlantisSucceed(state) {
			return result, fmt.Errorf("[%s] expected success but got %q", tc.Name, state)
		}
	}

	// Assert expected comment substring if configured.
	if tc.ExpectedCommentSubstring != "" {
		if err := assertCommentContains(ctx, t.vcsClient, pullID, tc.Name, tc.ExpectedCommentSubstring); err != nil {
			return result, err
		}
	}

	return result, nil
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

func cleanUp(ctx context.Context, t *E2ETester, pullRequestNumber int, branchName string) error {
	err := t.vcsClient.ClosePullRequest(ctx, pullRequestNumber)
	if err != nil {
		return err
	}
	log.Printf("closed pull request %d", pullRequestNumber)

	err = t.vcsClient.DeleteBranch(ctx, branchName)
	if err != nil {
		return fmt.Errorf("error while deleting branch %s: %v", branchName, err)
	}
	log.Printf("deleted branch %s", branchName)

	return nil
}
