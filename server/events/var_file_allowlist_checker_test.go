package events_test

import (
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
			"Empty Allowlist, single var file outside the repo directory",
			"",
			[]string{"-var-file=/path/to/file"},
			"var file path /path/to/file is not allowed by the current allowlist: []",
		},
		{
			"Empty Allowlist, single var file under the parent directory of the repo directory",
			"",
			[]string{"-var-file=../test.tfvars"},
			"var file path ../test.tfvars is not allowed by the current allowlist: []",
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
			"var file path ../test.tfvars is not allowed by the current allowlist: [/path]",
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
