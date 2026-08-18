// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.

package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"

	multierror "github.com/hashicorp/go-multierror"
)

var defaultAtlantisURL = "http://localhost:4141"

func getVCSClient() (VCSClient, error) {
	if os.Getenv("ATLANTIS_GH_USER") != "" || os.Getenv("ATLANTIS_GH_APP_ID") != "" {
		log.Print("Running tests for github")
		return NewGithubClient(), nil
	}
	if os.Getenv("ATLANTIS_GITLAB_USER") != "" {
		log.Print("Running tests for gitlab")
		return NewGitlabClient(), nil
	}

	return nil, errors.New("could not determine which vcs client")
}

func isGitHub() bool {
	return os.Getenv("ATLANTIS_GH_USER") != "" || os.Getenv("ATLANTIS_GH_APP_ID") != ""
}

func main() {
	atlantisURL := os.Getenv("ATLANTIS_URL")
	if atlantisURL == "" {
		atlantisURL = defaultAtlantisURL
	}
	atlantisURL = fmt.Sprintf("%s/events", atlantisURL)

	cloneDirRoot := os.Getenv("CLONE_DIR")
	if cloneDirRoot == "" {
		cloneDirRoot = "/tmp/atlantis-tests"
	}

	log.Print("cleaning workspace")
	if err := cleanDir(cloneDirRoot); err != nil {
		log.Fatalf("failed to clean dir: %v", err)
	}

	vcsClient, err := getVCSClient()
	if err != nil {
		log.Fatalf("failed to get vcs client: %v", err)
	}

	ctx := context.Background()
	log.Print("creating atlantis webhook")
	hookID, err := vcsClient.CreateAtlantisWebhook(ctx, atlantisURL)
	if err != nil {
		log.Fatalf("error creating atlantis webhook: %v", err)
	}

	cases := activeCases()
	log.Printf("running %d test cases", len(cases))

	results, err := runCases(ctx, vcsClient, hookID, cloneDirRoot, cases)

	// Print results summary.
	log.Printf("\nTest Results\n---------------------------")
	for _, r := range results {
		status := r.testResult
		if r.err != nil {
			status = fmt.Sprintf("FAIL: %v", r.err)
		}
		fmt.Printf("  %-35s %s\n", r.testCase, status)
		if r.pullRequestURL != "" {
			fmt.Printf("  %-35s PR: %s\n", "", r.pullRequestURL)
		}
	}
	fmt.Println("---------------------------")

	if err != nil {
		log.Fatalf("%v", err)
	}
}

func activeCases() []TestCase {
	optIn := os.Getenv("E2E_OPT_IN") == "1"
	gh := isGitHub()

	var active []TestCase
	for _, tc := range testCases {
		switch tc.Status {
		case CaseDisabled:
			log.Printf("skipping disabled case %q: %s", tc.Name, tc.SkipReason)
			continue
		case CaseOptIn:
			if !optIn {
				log.Printf("skipping opt-in case %q (set E2E_OPT_IN=1): %s", tc.Name, tc.SkipReason)
				continue
			}
		}

		switch tc.VCS {
		case VCSGitHub:
			if !gh {
				continue
			}
		case VCSGitLab:
			if gh {
				continue
			}
		}

		active = append(active, tc)
	}
	return active
}

func runCases(ctx context.Context, vcsClient VCSClient, hookID int64, cloneDirRoot string, cases []TestCase) ([]*E2EResult, error) {
	var results []*E2EResult
	var testErrors *multierror.Error

	defer vcsClient.DeleteAtlantisHook(ctx, hookID) // nolint: errcheck

	for _, tc := range cases {
		log.Printf("━━━ starting: %s ━━━", tc.Name)
		e2e := &E2ETester{
			vcsClient:    vcsClient,
			hookID:       hookID,
			cloneDirRoot: cloneDirRoot,
			testCase:     tc,
		}
		result, err := e2e.Start(ctx)
		if err != nil {
			result.err = err
			log.Printf("━━━ FAILED: %s — %v ━━━", tc.Name, err)
		} else {
			log.Printf("━━━ passed: %s ━━━", tc.Name)
		}
		results = append(results, result)
		testErrors = multierror.Append(testErrors, err)
	}

	return results, testErrors.ErrorOrNil()
}

// cleanDir validates the path is confined to an approved temp workspace and removes it.
func cleanDir(path string) error {
	cleanPath, err := validateCleanPath(path)
	if err != nil {
		return err
	}
	return os.RemoveAll(cleanPath) //nolint:gosec // path confined to known temp root by validateCleanPath
}

// knownSafeTempRoots are the fixed filesystem temp directories.
// os.TempDir() is NOT included unconditionally because TMPDIR can be set
// to arbitrary paths (e.g. $HOME), which would expand the allowlist unsafely.
var knownSafeTempRoots = []string{"/tmp", "/var/tmp"}

// approvedTempRoots returns canonical paths of directories under which
// E2E workspace cleanup is permitted. Only includes os.TempDir() if it
// itself resolves under a known safe root.
func approvedTempRoots() []string {
	seen := make(map[string]bool)
	var roots []string

	for _, r := range knownSafeTempRoots {
		canon := canonicalize(r)
		if canon != "" && !seen[canon] {
			seen[canon] = true
			roots = append(roots, canon)
		}
	}

	// Include os.TempDir() only if it resolves under a known safe root.
	if td := canonicalize(os.TempDir()); td != "" && !seen[td] {
		for _, root := range roots {
			if isPathBelow(root, td) {
				seen[td] = true
				roots = append(roots, td)
				break
			}
		}
	}

	return roots
}

// validateCleanPath ensures path is a proper child of an approved temp root
// and does not overlap with protected runtime paths.
//
// Order of checks:
// 1. Reject empty/whitespace
// 2. Resolve to clean absolute path
// 3. Canonicalize for validation
// 4. Reject if overlaps protected runtime paths (cwd, repo root, home)
// 5. Reject if equals an approved temp root
// 6. Accept only if strictly below an approved temp root
func validateCleanPath(path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", errors.New("clone dir must not be empty")
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolving clone dir: %w", err)
	}
	cleanPath := filepath.Clean(absPath)
	candidateCanon := canonicalizeForValidation(cleanPath)

	// Reject overlap with protected runtime paths.
	if err := rejectProtectedPaths(candidateCanon); err != nil {
		return "", err
	}

	// Check allowlist in two phases: reject root equality first (regardless
	// of ordering), then accept children. This prevents a nested os.TempDir()
	// root from being accepted as a child of a parent root before its own
	// equality check runs.
	roots := approvedTempRoots()
	if slices.Contains(roots, candidateCanon) {
		return "", fmt.Errorf("refusing to clean temp root itself %q", cleanPath)
	}
	for _, root := range roots {
		if isPathBelow(root, candidateCanon) {
			return cleanPath, nil
		}
	}

	return "", fmt.Errorf("path %q is not under an approved temp root %v", cleanPath, roots)
}

// rejectProtectedPaths returns an error if candidateCanon overlaps any
// protected runtime path (cwd, repo root, home directory).
// "Overlaps" means the candidate equals, contains, or is contained by a protected path.
func rejectProtectedPaths(candidateCanon string) error {
	var protected []string

	if cwd, err := os.Getwd(); err == nil {
		cwdCanon := canonicalizeForValidation(filepath.Clean(cwd))
		protected = append(protected, cwdCanon)
		// Repo root (parent of e2e/ working dir).
		repoRoot := canonicalizeForValidation(filepath.Clean(filepath.Dir(cwd)))
		protected = append(protected, repoRoot)
	}

	if home, err := os.UserHomeDir(); err == nil {
		protected = append(protected, canonicalizeForValidation(filepath.Clean(home)))
	}

	for _, p := range protected {
		if pathsOverlap(candidateCanon, p) {
			return fmt.Errorf("refusing to clean %q: overlaps protected path %q", candidateCanon, p)
		}
	}
	return nil
}

// pathsOverlap returns true if a equals b, a is below b, or b is below a.
func pathsOverlap(a, b string) bool {
	if a == b {
		return true
	}
	return isPathBelow(a, b) || isPathBelow(b, a)
}

// canonicalize resolves symlinks for an existing path.
func canonicalize(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		return ""
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return filepath.Clean(abs)
	}
	return filepath.Clean(resolved)
}

// canonicalizeForValidation resolves the candidate path for comparison.
// If the path does not exist, resolves the nearest existing ancestor
// and appends the unresolved suffix.
func canonicalizeForValidation(path string) string {
	resolved, err := filepath.EvalSymlinks(path)
	if err == nil {
		return filepath.Clean(resolved)
	}

	parent := filepath.Dir(path)
	base := filepath.Base(path)
	if parent == path {
		return filepath.Clean(path)
	}
	resolvedParent := canonicalizeForValidation(parent)
	return filepath.Join(resolvedParent, base)
}

// isPathBelow returns true if candidate is strictly below base (not equal).
func isPathBelow(base, candidate string) bool {
	rel, err := filepath.Rel(base, candidate)
	if err != nil {
		return false
	}
	if rel == "." {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator))
}
