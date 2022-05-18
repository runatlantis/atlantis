package common

import (
	"os/exec"
	"reflect"
	"strings"
	"testing"

	. "github.com/runatlantis/atlantis/testing"
)

func Test_DeDuplicateExtraArgs(t *testing.T) {
	cases := []struct {
		description  string
		inputArgs    []string
		extraArgs    []string
		expectedArgs []string
	}{
		{
			"No extra args",
			[]string{"init", "-input=false", "-no-color", "-upgrade"},
			[]string{},
			[]string{"init", "-input=false", "-no-color", "-upgrade"},
		},
		{
			"Override -upgrade",
			[]string{"init", "-input=false", "-no-color", "-upgrade"},
			[]string{"-upgrade=false"},
			[]string{"init", "-input=false", "-no-color", "-upgrade=false"},
		},
		{
			"Override -input",
			[]string{"init", "-input=false", "-no-color", "-upgrade"},
			[]string{"-input=true"},
			[]string{"init", "-input=true", "-no-color", "-upgrade"},
		},
		{
			"Override -input and -upgrade",
			[]string{"init", "-input=false", "-no-color", "-upgrade"},
			[]string{"-input=true", "-upgrade=false"},
			[]string{"init", "-input=true", "-no-color", "-upgrade=false"},
		},
		{
			"Non duplicate extra args",
			[]string{"init", "-input=false", "-no-color", "-upgrade"},
			[]string{"extra", "args"},
			[]string{"init", "-input=false", "-no-color", "-upgrade", "extra", "args"},
		},
		{
			"Override upgrade with extra args",
			[]string{"init", "-input=false", "-no-color", "-upgrade"},
			[]string{"extra", "args", "-upgrade=false"},
			[]string{"init", "-input=false", "-no-color", "-upgrade=false", "extra", "args"},
		},
		{
			"Override -input (using --input)",
			[]string{"init", "-input=false", "-no-color", "-upgrade"},
			[]string{"--input=true"},
			[]string{"init", "--input=true", "-no-color", "-upgrade"},
		},
		{
			"Override -input (using --input) and -upgrade (using --upgrade)",
			[]string{"init", "-input=false", "-no-color", "-upgrade"},
			[]string{"--input=true", "--upgrade=false"},
			[]string{"init", "--input=true", "-no-color", "--upgrade=false"},
		},
		{
			"Override long form flag ",
			[]string{"init", "--input=false", "-no-color", "-upgrade"},
			[]string{"--input=true"},
			[]string{"init", "--input=true", "-no-color", "-upgrade"},
		},
		{
			"Override --input using (-input) ",
			[]string{"init", "--input=false", "-no-color", "-upgrade"},
			[]string{"-input=true"},
			[]string{"init", "-input=true", "-no-color", "-upgrade"},
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			finalArgs := DeDuplicateExtraArgs(c.inputArgs, c.extraArgs)

			if !reflect.DeepEqual(c.expectedArgs, finalArgs) {
				t.Fatalf("finalArgs (%v) does not match expectedArgs (%v)", finalArgs, c.expectedArgs)
			}
		})
	}
}

func runCmd(t *testing.T, dir string, name string, args ...string) string {
	t.Helper()
	cpCmd := exec.Command(name, args...)
	cpCmd.Dir = dir
	cpOut, err := cpCmd.CombinedOutput()
	Assert(t, err == nil, "err running %q: %s", strings.Join(append([]string{name}, args...), " "), cpOut)
	return string(cpOut)
}

func initRepo(t *testing.T) (string, func()) {
	repoDir, cleanup := TempDir(t)
	runCmd(t, repoDir, "git", "init")
	runCmd(t, repoDir, "touch", ".gitkeep")
	runCmd(t, repoDir, "git", "add", ".gitkeep")
	runCmd(t, repoDir, "git", "config", "--local", "user.email", "atlantisbot@runatlantis.io")
	runCmd(t, repoDir, "git", "config", "--local", "user.name", "atlantisbot")
	runCmd(t, repoDir, "git", "commit", "-m", "initial commit")
	runCmd(t, repoDir, "git", "branch", "branch")
	return repoDir, cleanup
}

func TestIsFileTracked(t *testing.T) {
	// Initialize the git repo.
	repoDir, cleanup := initRepo(t)
	defer cleanup()

	// file1 should not be tracked
	tracked, err := IsFileTracked(repoDir, "file1")
	Ok(t, err)
	Equals(t, tracked, false)

	// stage file1
	runCmd(t, repoDir, "touch", "file1")
	runCmd(t, repoDir, "git", "add", "file1")
	runCmd(t, repoDir, "git", "commit", "-m", "add file1")

	// file1 should  be tracked
	tracked, err = IsFileTracked(repoDir, "file1")
	Ok(t, err)
	Equals(t, tracked, true)

	// .terraform.lock.hcl should not be tracked
	tracked, err = IsFileTracked(repoDir, ".terraform.lock.hcl")
	Ok(t, err)
	Equals(t, tracked, false)

	// stage .terraform.lock.hcl
	runCmd(t, repoDir, "touch", ".terraform.lock.hcl")
	runCmd(t, repoDir, "git", "add", ".terraform.lock.hcl")
	runCmd(t, repoDir, "git", "commit", "-m", "add .terraform.lock.hcl")

	// file1 should  be tracked
	tracked, err = IsFileTracked(repoDir, ".terraform.lock.hcl")
	Ok(t, err)
	Equals(t, tracked, true)
}
