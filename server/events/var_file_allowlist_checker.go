// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/runatlantis/atlantis/server/utils"
)

// VarFileAllowlistChecker implements checking if paths are allowlisted to be used with
// this Atlantis.
type VarFileAllowlistChecker struct {
	rules []string
}

// NewVarFileAllowlistChecker constructs a new checker and validates that the
// allowlist isn't malformed.
func NewVarFileAllowlistChecker(allowlist string) (*VarFileAllowlistChecker, error) {
	var rules []string
	paths := strings.Split(allowlist, ",")
	if paths[0] != "" {
		for _, path := range paths {
			absPath, err := filepath.Abs(path)
			if err != nil {
				return nil, fmt.Errorf("converting allowlist %q to absolute path: %w", path, err)
			}
			rules = append(rules, absPath)
		}
	}
	return &VarFileAllowlistChecker{
		rules: rules,
	}, nil
}

// varFileFlagName is the Terraform flag whose value the allowlist guards, with
// the leading dashes removed. Terraform accepts both a single and a double dash
// for every flag, so the spelling is normalised before comparing.
const varFileFlagName = "var-file"

// Check rejects any var-file flag whose path, as written in the comment, is
// neither inside the repository nor under an allowlisted directory. It runs
// before the repository is cloned, so it can only judge the spelling of the
// path; CheckResolved finishes the job once the files exist on disk.
func (p *VarFileAllowlistChecker) Check(flags []string) error {
	for _, path := range varFilePaths(flags) {
		if err := rejectParentElements(path); err != nil {
			return err
		}
		if !p.isAllowedPath(path) {
			return fmt.Errorf("var file path %s is not allowed by the current allowlist: [%s]",
				path, strings.Join(p.rules, ", "))
		}
	}
	return nil
}

// CheckResolved re-checks the var-file flags against the files they actually
// name once the repository is cloned. Check accepts a relative path on its
// spelling alone, but a pull request can commit a symbolic link at that path
// pointing anywhere on the server, and Terraform follows it. So each path is
// resolved relative to projectDir, the directory Terraform runs in, with
// symlinks followed, and the result must lie inside repoDir or inside an
// allowlisted directory. A dangling link is rejected; a plain file that does
// not exist yet is checked through its nearest existing directory.
func (p *VarFileAllowlistChecker) CheckResolved(repoDir, projectDir string, flags []string) error {
	resolvedRepo, err := filepath.EvalSymlinks(repoDir)
	if err != nil {
		return fmt.Errorf("resolving repository directory %q: %w", repoDir, err)
	}
	for _, path := range varFilePaths(flags) {
		if err := rejectParentElements(path); err != nil {
			return err
		}
		abs := path
		if !filepath.IsAbs(abs) {
			abs = filepath.Join(projectDir, abs)
		}
		resolved, err := resolveVarFilePath(abs)
		if errors.Is(err, errDanglingVarFile) {
			return fmt.Errorf("var file path %s is a symbolic link whose target does not exist", path)
		}
		if err != nil {
			return fmt.Errorf("resolving var file path %q: %w", path, err)
		}
		if utils.EnsureSubPath(resolvedRepo, resolved) == nil || p.inAllowlistedDir(resolved, true) {
			continue
		}
		return fmt.Errorf("var file path %s resolves to %s, which is not allowed by the current allowlist: [%s]",
			path, resolved, strings.Join(p.rules, ", "))
	}
	return nil
}

// rejectParentElements refuses a var file path that contains a ".." element.
//
// Both checks reason about a cleaned path, but Terraform hands the raw value
// to the kernel, which follows a symlink before applying "..". With a
// committed link d -> /outside/a/b, the value d/../secret.tfvars cleans to
// secret.tfvars inside the project yet opens /outside/a/secret.tfvars.
// Resolving the raw path would also work, but a ".." element has no
// legitimate use here: a relative path that starts with ".." was already
// refused, and an allowlisted or in-repository file can always be named
// without one.
func rejectParentElements(path string) error {
	elems := strings.FieldsFunc(path, func(r rune) bool { return r == '/' || r == filepath.Separator })
	if slices.Contains(elems, "..") {
		return fmt.Errorf("var file path %s must not contain '..' elements", path)
	}
	return nil
}

// errDanglingVarFile reports a var file path that is, or passes through, a
// symbolic link whose target does not exist.
var errDanglingVarFile = errors.New("dangling symbolic link")

// resolveVarFilePath returns abs with every symbolic link resolved, so the
// caller can check where Terraform would actually read from.
//
// A file that does not exist yet is allowed, because an earlier workflow step
// may generate it: the longest existing ancestor directory is resolved and the
// missing components are appended, so a missing file below a directory link
// that leaves the repository is still caught. A dangling link is rejected
// outright. Accepting it would let its target be created after this check,
// by a later step or a concurrent run, and Terraform would then follow the
// already-approved link wherever it points.
func resolveVarFilePath(abs string) (string, error) {
	resolved, err := filepath.EvalSymlinks(abs)
	if err == nil {
		return resolved, nil
	}
	if !errors.Is(err, fs.ErrNotExist) {
		return "", err
	}
	// abs itself exists but cannot be resolved: it is a dangling link.
	if _, lerr := os.Lstat(abs); lerr == nil {
		return "", errDanglingVarFile
	}

	dir := abs
	var missing []string
	for {
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", err
		}
		missing = append([]string{filepath.Base(dir)}, missing...)
		dir = parent
		if _, lerr := os.Lstat(dir); lerr != nil {
			continue
		}
		resolvedDir, rerr := filepath.EvalSymlinks(dir)
		if errors.Is(rerr, fs.ErrNotExist) {
			return "", errDanglingVarFile
		}
		if rerr != nil {
			return "", rerr
		}
		return filepath.Join(append([]string{resolvedDir}, missing...)...), nil
	}
}

// varFilePaths returns the value of every var-file flag in flags, in order.
func varFilePaths(flags []string) []string {
	var paths []string
	for i, flag := range flags {
		name, value, joined := strings.Cut(flag, "=")
		// Cut on the first '=' only: a var file path may itself contain one.
		//
		// Compare after trimming every leading dash, so both spellings
		// Terraform accepts (-var-file and --var-file) are checked. This
		// over-matches any dash count, which is deliberate: an unrecognised
		// spelling only makes Terraform error, never read the file, so the
		// safe direction for this check is to match too much rather than miss
		// a spelling.
		if strings.TrimLeft(name, "-") != varFileFlagName {
			continue
		}

		switch {
		case joined:
			// Flags are in the format of []string{"-var-file=my-file.tfvars"}
			paths = append(paths, value)
		case i < len(flags)-1:
			// Flags are in the format of []string{"-var-file", "my-file.tfvars"}
			paths = append(paths, flags[i+1])
		}
	}
	return paths
}

func (p *VarFileAllowlistChecker) isAllowedPath(path string) bool {
	path = filepath.Clean(path)

	// If the path is within the repo directory, return true without checking the rules.
	if !filepath.IsAbs(path) {
		if !strings.HasPrefix(path, "..") && !strings.HasPrefix(path, "~") {
			return true
		}
	}

	return p.inAllowlistedDir(path, false)
}

// inAllowlistedDir reports whether path lies under one of the allowlisted
// directories. With resolveRules set, each rule is first resolved with
// symlinks followed so that it can be compared with a path that was resolved
// the same way; a rule that cannot be resolved is compared as written.
func (p *VarFileAllowlistChecker) inAllowlistedDir(path string, resolveRules bool) bool {
	for _, rule := range p.rules {
		if resolveRules {
			if resolved, err := filepath.EvalSymlinks(rule); err == nil {
				rule = resolved
			}
		}
		rel, err := filepath.Rel(rule, path)
		if err == nil && !strings.HasPrefix(rel, "..") {
			return true
		}
	}
	return false
}
