// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/runatlantis/atlantis/server/events"
	. "github.com/runatlantis/atlantis/testing"
)

func TestVarFileAllowlistChecker_IsAllowlisted(t *testing.T) {
	cases := []struct {
		Description string
		Allowlist   string
		Flags       []string
		ExpErr      string
	}{
		{
			// The kernel follows a symlink before applying "..", so a path
			// like d/../x reads relative to d's target, not to the project.
			// Cleaning it lexically to "x" would approve a file outside the
			// repository, so any ".." element is refused.
			"Parent element inside a relative path is rejected",
			"",
			[]string{"-var-file=d/../secret.tfvars"},
			"var file path d/../secret.tfvars must not contain '..' elements",
		},
		{
			"Parent element is rejected for the double-dash separate form",
			"",
			[]string{"--var-file", "sub/../x.tfvars"},
			"var file path sub/../x.tfvars must not contain '..' elements",
		},
		{
			"Parent element inside an allowlisted absolute path is rejected",
			"/allowed",
			[]string{"-var-file=/allowed/a/../b.tfvars"},
			"var file path /allowed/a/../b.tfvars must not contain '..' elements",
		},
		{
			"Dots in a file name are not a parent element",
			"",
			[]string{"-var-file=env..prod.tfvars"},
			"",
		},
		{
			"Empty Allowlist, no var file",
			"",
			[]string{""},
			"",
		},
		{
			"Empty Allowlist, single var file under the repo directory",
			"",
			[]string{"-var-file=test.tfvars"},
			"",
		},
		{
			"Empty Allowlist, single var file under the repo directory, specified in separate flags",
			"",
			[]string{"-var-file", "test.tfvars"},
			"",
		},
		{
			"Empty Allowlist, single var file under the subdirectory of the repo directory",
			"",
			[]string{"-var-file=sub/test.tfvars"},
			"",
		},
		{
			// Terraform accepts both a single and a double dash for every
			// flag, so the allowlist has to recognise both spellings.
			"Double dash spelling is checked, joined form",
			"/allowed",
			[]string{"--var-file=/etc/passwd"},
			"var file path /etc/passwd is not allowed by the current allowlist: [/allowed]",
		},
		{
			"Double dash spelling is checked, separate flags",
			"/allowed",
			[]string{"--var-file", "/etc/passwd"},
			"var file path /etc/passwd is not allowed by the current allowlist: [/allowed]",
		},
		{
			"Single dash spelling is still checked, joined form",
			"/allowed",
			[]string{"-var-file=/etc/passwd"},
			"var file path /etc/passwd is not allowed by the current allowlist: [/allowed]",
		},
		{
			"Double dash spelling is allowed when the path is allowlisted",
			"/allowed",
			[]string{"--var-file=/allowed/test.tfvars"},
			"",
		},
		{
			// Dash count is matched permissively on purpose: the flag name is
			// compared after trimming every leading dash, so any dash-prefixed
			// spelling is checked. Terraform accepts one or two dashes; a
			// spelling it does not accept can only cause it to error, never to
			// read the file, so over-matching here is the safe direction for a
			// bug whose whole shape is under-matching the parser.
			"Extra dashes are still treated as the var-file flag",
			"/allowed",
			[]string{"---var-file=/etc/passwd"},
			"var file path /etc/passwd is not allowed by the current allowlist: [/allowed]",
		},
		{
			// A path may itself contain '=', so only the first one separates
			// the flag from its value.
			"Value containing an equals sign is still checked",
			"/allowed",
			[]string{"-var-file=/etc/pass=wd"},
			"var file path /etc/pass=wd is not allowed by the current allowlist: [/allowed]",
		},
		{
			"Empty Allowlist, single var file outside the repo directory",
			"",
			[]string{"-var-file=/path/to/file"},
			"var file path /path/to/file is not allowed by the current allowlist: []",
		},
		{
			"Empty Allowlist, single var file under the parent directory of the repo directory",
			"",
			[]string{"-var-file=../test.tfvars"},
			"var file path ../test.tfvars must not contain '..' elements",
		},
		{
			"Empty Allowlist, single var file under the home directory",
			"",
			[]string{"-var-file=~/test.tfvars"},
			"var file path ~/test.tfvars is not allowed by the current allowlist: []",
		},
		{
			"Single path in allowlist, no var file",
			"/path",
			[]string{""},
			"",
		},
		{
			"Single path in allowlist, single var file under the repo directory",
			"/path",
			[]string{"-var-file=test.tfvars"},
			"",
		},
		{
			"Single path in allowlist, single var file under the allowlisted directory",
			"/path",
			[]string{"-var-file=/path/test.tfvars"},
			"",
		},
		{
			"Single path with ending slash in allowlist, single var file under the allowlisted directory",
			"/path/",
			[]string{"-var-file=/path/test.tfvars"},
			"",
		},
		{
			"Single path in allowlist, single var file in the parent directory of the repo directory",
			"/path",
			[]string{"-var-file=../test.tfvars"},
			"var file path ../test.tfvars must not contain '..' elements",
		},
		{
			"Single path in allowlist, single var file outside the allowlisted directory",
			"/path",
			[]string{"-var-file=/path_not_allowed/test.tfvars"},
			"var file path /path_not_allowed/test.tfvars is not allowed by the current allowlist: [/path]",
		},
		{
			"Single path in allowlist, single var file in the parent directory of the allowlisted directory",
			"/path",
			[]string{"-var-file=/test.tfvars"},
			"var file path /test.tfvars is not allowed by the current allowlist: [/path]",
		},
		{
			"Root path in allowlist, with multiple var files",
			"/",
			[]string{"-var-file=test.tfvars", "-var-file=/path/test.tfvars", "-var-file=/test.tfvars"},
			"",
		},
		{
			"Multiple paths in allowlist, with multiple var files under allowlisted directories",
			"/path,/another/path",
			[]string{"-var-file=test.tfvars", "-var-file", "/path/test.tfvars", "unused-flag", "-var-file=/another/path/sub/test.tfvars"},
			"",
		},
	}

	for _, c := range cases {
		t.Run(c.Description, func(t *testing.T) {
			v, err := events.NewVarFileAllowlistChecker(c.Allowlist)
			Ok(t, err)

			err = v.Check(c.Flags)
			if c.ExpErr != "" {
				ErrEquals(t, c.ExpErr, err)
			} else {
				Ok(t, err)
			}
		})
	}
}

func TestVarFileAllowlistChecker_CheckResolved(t *testing.T) {
	// Layout:
	//   root/repo/project/   the project directory Terraform runs in
	//   root/repo/           the repository root
	//   root/outside/        a directory the repository must not read from
	//   root/allowed/        a directory that can be put on the allowlist
	root := t.TempDir()
	repoDir := filepath.Join(root, "repo")
	projectDir := filepath.Join(repoDir, "project")
	outsideDir := filepath.Join(root, "outside")
	allowedDir := filepath.Join(root, "allowed")
	for _, dir := range []string{projectDir, outsideDir, allowedDir} {
		Ok(t, os.MkdirAll(dir, 0o755))
	}
	write := func(path string) {
		Ok(t, os.WriteFile(path, []byte("x = 1\n"), 0o600))
	}
	write(filepath.Join(projectDir, "real.tfvars"))
	write(filepath.Join(repoDir, "shared.tfvars"))
	write(filepath.Join(outsideDir, "secret.tfvars"))
	write(filepath.Join(allowedDir, "ok.tfvars"))
	// Links a pull request could commit to the repository.
	Ok(t, os.Symlink(filepath.Join(outsideDir, "secret.tfvars"), filepath.Join(projectDir, "evil.tfvars")))
	Ok(t, os.Symlink(outsideDir, filepath.Join(projectDir, "evil-dir")))
	// d points two levels into a directory outside the repository, so d/..
	// resolves to that directory's parent, not to the project directory.
	Ok(t, os.MkdirAll(filepath.Join(outsideDir, "a", "b"), 0o755))
	Ok(t, os.Symlink(filepath.Join(outsideDir, "a", "b"), filepath.Join(projectDir, "d")))
	Ok(t, os.Symlink(filepath.Join(repoDir, "shared.tfvars"), filepath.Join(projectDir, "fine.tfvars")))
	Ok(t, os.Symlink(filepath.Join(allowedDir, "ok.tfvars"), filepath.Join(projectDir, "allowed.tfvars")))
	// A link whose target does not exist yet. If the target is created after
	// the check (by a later step or a concurrent run), Terraform would follow
	// the link outside the repository, so it must be rejected now.
	Ok(t, os.Symlink(filepath.Join(outsideDir, "later.tfvars"), filepath.Join(projectDir, "dangling.tfvars")))
	Ok(t, os.Symlink(filepath.Join(projectDir, "later.tfvars"), filepath.Join(projectDir, "dangling-inside.tfvars")))

	// The temp dir may itself sit behind a symlink (macOS), so the resolved
	// target in the error message is computed the same way the checker does.
	resolvedSecret, err := filepath.EvalSymlinks(filepath.Join(outsideDir, "secret.tfvars"))
	Ok(t, err)
	outsideErr := func(path, allowlist string) string {
		return fmt.Sprintf("var file path %s resolves to %s, which is not allowed by the current allowlist: [%s]", path, resolvedSecret, allowlist)
	}

	cases := []struct {
		Description string
		Allowlist   string
		Flags       []string
		ExpErr      string
	}{
		{
			"regular file in the project directory",
			"",
			[]string{"-var-file=real.tfvars"},
			"",
		},
		{
			"symlink to a file elsewhere in the repository",
			"",
			[]string{"-var-file=fine.tfvars"},
			"",
		},
		{
			"symlink to a file outside the repository",
			"",
			[]string{"-var-file=evil.tfvars"},
			outsideErr("evil.tfvars", ""),
		},
		{
			"symlink to a file outside the repository, double dash and separate value",
			"",
			[]string{"--var-file", "evil.tfvars"},
			outsideErr("evil.tfvars", ""),
		},
		{
			"file reached through a symlinked directory outside the repository",
			"",
			[]string{"-var-file=evil-dir/secret.tfvars"},
			outsideErr("evil-dir/secret.tfvars", ""),
		},
		{
			"symlink into an allowlisted directory",
			allowedDir,
			[]string{"-var-file=allowed.tfvars"},
			"",
		},
		{
			"symlink outside the repository with an unrelated allowlist",
			allowedDir,
			[]string{"-var-file=evil.tfvars"},
			outsideErr("evil.tfvars", allowedDir),
		},
		{
			"absolute path to an allowlisted file",
			allowedDir,
			[]string{"-var-file=" + filepath.Join(allowedDir, "ok.tfvars")},
			"",
		},
		{
			// A plain file that does not exist yet may be produced by an
			// earlier workflow step, so it is allowed when its directory is
			// inside the repository.
			"missing file is left for Terraform to report",
			"",
			[]string{"-var-file=missing.tfvars"},
			"",
		},
		{
			"dangling symlink to a missing file outside the repository",
			"",
			[]string{"-var-file=dangling.tfvars"},
			"var file path dangling.tfvars is a symbolic link whose target does not exist",
		},
		{
			"dangling symlink is rejected even when it points inside the repository",
			"",
			[]string{"-var-file=dangling-inside.tfvars"},
			"var file path dangling-inside.tfvars is a symbolic link whose target does not exist",
		},
		{
			"parent element after a symlinked directory is rejected",
			"",
			[]string{"-var-file=d/../secret.tfvars"},
			"var file path d/../secret.tfvars must not contain '..' elements",
		},
		{
			"parent element after a symlinked directory is rejected for a missing leaf",
			"",
			[]string{"-var-file=d/../later.tfvars"},
			"var file path d/../later.tfvars must not contain '..' elements",
		},
		{
			"absolute path through the clone with a parent element is rejected",
			"",
			[]string{"-var-file=" + projectDir + "/d/../later.tfvars"},
			"var file path " + projectDir + "/d/../later.tfvars must not contain '..' elements",
		},
		{
			"missing file below a symlinked directory outside the repository",
			"",
			[]string{"-var-file=evil-dir/not-there.tfvars"},
			fmt.Sprintf("var file path evil-dir/not-there.tfvars resolves to %s, which is not allowed by the current allowlist: []", filepath.Join(filepath.Dir(resolvedSecret), "not-there.tfvars")),
		},
		{
			"missing file in a missing subdirectory inside the repository",
			"",
			[]string{"-var-file=generated/vars.tfvars"},
			"",
		},
		{
			"only the first offending file is reported",
			"",
			[]string{"-var-file=real.tfvars", "-var-file", "evil.tfvars", "-var-file=evil-dir/secret.tfvars"},
			outsideErr("evil.tfvars", ""),
		},
	}

	for _, c := range cases {
		t.Run(c.Description, func(t *testing.T) {
			v, err := events.NewVarFileAllowlistChecker(c.Allowlist)
			Ok(t, err)

			err = v.CheckResolved(repoDir, projectDir, c.Flags)
			if c.ExpErr != "" {
				ErrEquals(t, c.ExpErr, err)
			} else {
				Ok(t, err)
			}
		})
	}
}
