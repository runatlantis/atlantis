// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

type fixturePull struct {
	label      string
	cloneDir   string
	branchName string
	url        string
	pullID     int
}

func (t *E2ETester) createFixturePull(ctx context.Context, label, branchName, mutateFile string, prepare func(string) error) (*fixturePull, error) {
	tc := t.testCase
	cloneName := fmt.Sprintf("%s-test", tc.Name)
	if label != "" {
		cloneName = fmt.Sprintf("%s-%s-test", tc.Name, label)
	}
	cloneDir := filepath.Join(t.cloneDirRoot, cloneName)
	log.Printf("[%s] creating %s clone dir %q", tc.Name, fixtureLabel(label), cloneDir)
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
	if prepare != nil {
		if err := prepare(cloneDir); err != nil {
			return nil, err
		}
	}
	if err := runGit(cloneDir, "add", "-A"); err != nil {
		return nil, err
	}
	commitMessage := fmt.Sprintf("e2e: %s", tc.Name)
	if label != "" {
		commitMessage += " " + label
	}
	if err := runGit(cloneDir, "commit", "-m", commitMessage); err != nil {
		return nil, err
	}
	if err := runGit(cloneDir, "push", "origin", branchName); err != nil {
		return nil, err
	}

	title := fmt.Sprintf("[E2E] %s", tc.Name)
	if label != "" {
		title += " " + label
	}
	url, pullID, err := t.vcsClient.CreatePullRequest(ctx, title, branchName)
	if err != nil {
		cleanupCtx, cancel := newLifecycleCleanupContext(ctx)
		defer cancel()

		if deleteErr := deleteRemoteBranch(cleanupCtx, cloneDir, branchName); deleteErr != nil {
			return nil, fmt.Errorf("creating pull request after pushing branch %q: %w; additionally failed to delete pushed branch: %v", branchName, err, deleteErr)
		}
		return nil, err
	}
	return &fixturePull{
		label:      fixtureLabel(label),
		cloneDir:   cloneDir,
		branchName: branchName,
		url:        url,
		pullID:     pullID,
	}, nil
}

func fixtureLabel(label string) string {
	if label == "" {
		return "lifecycle"
	}
	return label
}

func cleanUpFixturePulls(ctx context.Context, t *E2ETester, pulls ...*fixturePull) error {
	var cleanupErr error
	for _, pull := range slices.Backward(pulls) {
		if err := cleanUpFixturePull(ctx, t, pull); err != nil {
			cleanupErr = errors.Join(cleanupErr, err)
		}
	}
	return cleanupErr
}

func cleanUpFixturePull(ctx context.Context, t *E2ETester, pull *fixturePull) error {
	if pull == nil {
		return nil
	}
	var cleanupErr error
	if err := t.vcsClient.ClosePullRequest(ctx, pull.pullID); err != nil {
		cleanupErr = errors.Join(cleanupErr, fmt.Errorf("closing pull request %d: %w", pull.pullID, err))
	} else {
		log.Printf("closed pull request %d", pull.pullID)
	}
	if err := t.vcsClient.DeleteBranch(ctx, pull.branchName); err != nil {
		cleanupErr = errors.Join(cleanupErr, fmt.Errorf("deleting branch %s: %w", pull.branchName, err))
	} else {
		log.Printf("deleted branch %s", pull.branchName)
	}
	if cleanupErr != nil {
		return fmt.Errorf("[%s] cleanup failed for %s pull %d branch %q: %w", t.testCase.Name, pull.label, pull.pullID, pull.branchName, cleanupErr)
	}
	return nil
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

func runGitContext(ctx context.Context, dir string, args ...string) error {
	cmd := exec.CommandContext(ctx, "git", args...) //nolint:gosec // arguments are test-controlled branch/file names.
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, string(output))
	}
	return nil
}

func deleteRemoteBranch(ctx context.Context, cloneDir, branchName string) error {
	return runGitContext(ctx, cloneDir, "push", "origin", "--delete", branchName)
}
