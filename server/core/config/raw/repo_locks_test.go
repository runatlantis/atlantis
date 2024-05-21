package raw_test

import (
	"testing"

	"github.com/runatlantis/atlantis/server/core/config/raw"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	. "github.com/runatlantis/atlantis/testing"
)

func TestRepoLocks_UnmarshalYAML(t *testing.T) {
	repoLocksOnPlan := valid.RepoLocksOnPlanMode
	cases := []struct {
		description string
		input       string
		exp         raw.RepoLocks
	}{
		{
			description: "omit unset fields",
			input:       "",
			exp: raw.RepoLocks{
				Mode: nil,
			},
		},
		{
			description: "all fields set",
			input: `
mode: on_plan
`,
			exp: raw.RepoLocks{
				Mode: &repoLocksOnPlan,
			},
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			var a raw.RepoLocks
			err := unmarshalString(c.input, &a)
			Ok(t, err)
			Equals(t, c.exp, a)
		})
	}
}

func TestRepoLocks_Validate(t *testing.T) {
	repoLocksDisabled := valid.RepoLocksDisabledMode
	repoLocksOnPlan := valid.RepoLocksOnPlanMode
	repoLocksOnApply := valid.RepoLocksOnApplyMode
	randomString := valid.RepoLocksMode("random_string")
	cases := []struct {
		description string
		input       raw.RepoLocks
		errContains *string
	}{
		{
			description: "nothing set",
			input:       raw.RepoLocks{},
			errContains: nil,
		},
		{
			description: "mode set to disabled",
			input: raw.RepoLocks{
				Mode: &repoLocksDisabled,
			},
			errContains: nil,
		},
		{
			description: "mode set to on_plan",
			input: raw.RepoLocks{
				Mode: &repoLocksOnPlan,
			},
			errContains: nil,
		},
		{
			description: "mode set to on_apply",
			input: raw.RepoLocks{
				Mode: &repoLocksOnApply,
			},
			errContains: nil,
		},
		{
			description: "mode set to random string",
			input: raw.RepoLocks{
				Mode: &randomString,
			},
			errContains: String("valid value"),
		},
	}
	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			if c.errContains == nil {
				Ok(t, c.input.Validate())
			} else {
				ErrContains(t, *c.errContains, c.input.Validate())
			}
		})
	}
}

func TestRepoLocks_ToValid(t *testing.T) {
	repoLocksOnApply := valid.RepoLocksOnApplyMode
	cases := []struct {
		description string
		input       raw.RepoLocks
		exp         *valid.RepoLocks
	}{
		{
			description: "nothing set",
			input:       raw.RepoLocks{},
			exp:         &valid.DefaultRepoLocks,
		},
		{
			description: "value set",
			input: raw.RepoLocks{
				Mode: &repoLocksOnApply,
			},
			exp: &valid.RepoLocks{
				Mode: valid.RepoLocksOnApplyMode,
			},
		},
	}
	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			Equals(t, c.exp, c.input.ToValid())
		})
	}
}
