package valid_test

import (
	"fmt"
	"github.com/mohae/deepcopy"
	"github.com/runatlantis/atlantis/server/events/yaml/valid"
	. "github.com/runatlantis/atlantis/testing"
	"regexp"
	"testing"
)

func TestNewGlobalCfg(t *testing.T) {
	expDefaultWorkflow := valid.Workflow{
		Name: "default",
		Apply: valid.Stage{
			Steps: []valid.Step{
				{
					StepName: "apply",
				},
			},
		},
		Plan: valid.Stage{
			Steps: []valid.Step{
				{
					StepName: "init",
				},
				{
					StepName: "plan",
				},
			},
		},
	}
	baseCfg := valid.GlobalCfg{
		Repos: []valid.Repo{
			{
				IDRegex:              regexp.MustCompile(".*"),
				ApplyRequirements:    []string{},
				Workflow:             &expDefaultWorkflow,
				AllowedOverrides:     []string{},
				AllowCustomWorkflows: Bool(false),
			},
		},
		Workflows: map[string]valid.Workflow{
			"default": expDefaultWorkflow,
		},
	}

	cases := []struct {
		allowRepoCfg bool
		approvedReq  bool
		mergeableReq bool
	}{
		{
			allowRepoCfg: false,
			approvedReq:  false,
			mergeableReq: false,
		},
		{
			allowRepoCfg: true,
			approvedReq:  false,
			mergeableReq: false,
		},
		{
			allowRepoCfg: false,
			approvedReq:  true,
			mergeableReq: false,
		},
		{
			allowRepoCfg: false,
			approvedReq:  false,
			mergeableReq: true,
		},
		{
			allowRepoCfg: false,
			approvedReq:  true,
			mergeableReq: true,
		},
		{
			allowRepoCfg: true,
			approvedReq:  true,
			mergeableReq: true,
		},
	}

	for _, c := range cases {
		caseName := fmt.Sprintf("allow_repo: %t, approved: %t, mergeable: %t",
			c.allowRepoCfg, c.approvedReq, c.mergeableReq)
		t.Run(caseName, func(t *testing.T) {
			act := valid.NewGlobalCfg(c.allowRepoCfg, c.mergeableReq, c.approvedReq)

			// For each test, we change our expected cfg based on the parameters.
			exp := deepcopy.Copy(baseCfg).(valid.GlobalCfg)
			exp.Repos[0].IDRegex = regexp.MustCompile(".*") // deepcopy doesn't copy the regex.

			if c.allowRepoCfg {
				exp.Repos[0].AllowCustomWorkflows = Bool(true)
				exp.Repos[0].AllowedOverrides = []string{"apply_requirements", "workflow"}
			}
			if c.mergeableReq {
				exp.Repos[0].ApplyRequirements = append(exp.Repos[0].ApplyRequirements, "mergeable")
			}
			if c.approvedReq {
				exp.Repos[0].ApplyRequirements = append(exp.Repos[0].ApplyRequirements, "approved")
			}
			Equals(t, exp, act)

			// Have to hand-compare regexes because Equals doesn't do it.
			for i, actRepo := range act.Repos {
				expRepo := exp.Repos[i]
				if expRepo.IDRegex != nil {
					Assert(t, expRepo.IDRegex.String() == actRepo.IDRegex.String(),
						"%q != %q for repos[%d]", expRepo.IDRegex.String(), actRepo.IDRegex.String(), i)
				}
			}
		})
	}
}

func TestGlobalCfg_ValidateRepoCfg(t *testing.T) {
	cases := map[string]struct {
		gCfg   valid.GlobalCfg
		rCfg   valid.Config
		repoID string
		expErr string
	}{
		"workflow not allowed": {
			gCfg: valid.NewGlobalCfg(false, false, false),
			rCfg: valid.Config{
				Projects: []valid.Project{
					{
						Workflow: String("invalid"),
					},
				},
			},
			repoID: "github.com/owner/repo",
			expErr: "repo config not allowed to set 'workflow' key: server-side config needs 'allowed_overrides: [workflow]'",
		},
		"custom workflows not allowed": {
			gCfg: valid.NewGlobalCfg(false, false, false),
			rCfg: valid.Config{
				Workflows: map[string]valid.Workflow{
					"custom": {},
				},
			},
			repoID: "github.com/owner/repo",
			expErr: "repo config not allowed to define custom workflows: server-side config needs 'allow_custom_workflows: true'",
		},
		"custom workflows allowed": {
			gCfg: valid.NewGlobalCfg(true, false, false),
			rCfg: valid.Config{
				Workflows: map[string]valid.Workflow{
					"custom": {},
				},
			},
			repoID: "github.com/owner/repo",
			expErr: "",
		},
		"repo uses custom workflow defined on repo": {
			gCfg: valid.NewGlobalCfg(true, false, false),
			rCfg: valid.Config{
				Projects: []valid.Project{
					{
						Dir:       ".",
						Workspace: "default",
						Workflow:  String("repodefined"),
					},
				},
				Workflows: map[string]valid.Workflow{
					"repodefined": {},
				},
			},
			repoID: "github.com/owner/repo",
			expErr: "",
		},
		"custom workflows allowed for this repo only": {
			gCfg: valid.GlobalCfg{
				Repos: []valid.Repo{
					valid.NewGlobalCfg(false, false, false).Repos[0],
					{
						ID:                   "github.com/owner/repo",
						AllowCustomWorkflows: Bool(true),
					},
				},
			},
			rCfg: valid.Config{
				Workflows: map[string]valid.Workflow{
					"custom": {},
				},
			},
			repoID: "github.com/owner/repo",
			expErr: "",
		},
		"repo uses global workflow": {
			gCfg: valid.NewGlobalCfg(true, false, false),
			rCfg: valid.Config{
				Projects: []valid.Project{
					{
						Dir:       ".",
						Workspace: "default",
						Workflow:  String("default"),
					},
				},
			},
			repoID: "github.com/owner/repo",
			expErr: "",
		},
		"apply_reqs not allowed": {
			gCfg: valid.NewGlobalCfg(false, false, false),
			rCfg: valid.Config{
				Projects: []valid.Project{
					{
						Dir:               ".",
						Workspace:         "default",
						ApplyRequirements: []string{""},
					},
				},
			},
			repoID: "github.com/owner/repo",
			expErr: "repo config not allowed to set 'apply_requirements' key: server-side config needs 'allowed_overrides: [apply_requirements]'",
		},
		"repo workflow doesn't exist": {
			gCfg: valid.NewGlobalCfg(true, false, false),
			rCfg: valid.Config{
				Projects: []valid.Project{
					{
						Dir:       ".",
						Workspace: "default",
						Workflow:  String("doesntexist"),
					},
				},
			},
			repoID: "github.com/owner/repo",
			expErr: "workflow \"doesntexist\" is not defined anywhere",
		},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			actErr := c.gCfg.ValidateRepoCfg(c.rCfg, c.repoID)
			if c.expErr == "" {
				Ok(t, actErr)
			} else {
				ErrEquals(t, c.expErr, actErr)
			}
		})
	}
}

func TestRepo_IDMatches(t *testing.T) {
	// Test exact matches.
	Equals(t, false, (valid.Repo{ID: "github.com/owner/repo"}).IDMatches("github.com/runatlantis/atlantis"))
	Equals(t, true, (valid.Repo{ID: "github.com/owner/repo"}).IDMatches("github.com/owner/repo"))

	// Test regexes.
	Equals(t, true, (valid.Repo{IDRegex: regexp.MustCompile(".*")}).IDMatches("github.com/owner/repo"))
	Equals(t, true, (valid.Repo{IDRegex: regexp.MustCompile("github.com")}).IDMatches("github.com/owner/repo"))
	Equals(t, false, (valid.Repo{IDRegex: regexp.MustCompile("github.com/anotherowner")}).IDMatches("github.com/owner/repo"))
	Equals(t, true, (valid.Repo{IDRegex: regexp.MustCompile("github.com/(owner|runatlantis)")}).IDMatches("github.com/owner/repo"))
	Equals(t, true, (valid.Repo{IDRegex: regexp.MustCompile("github.com/owner.*")}).IDMatches("github.com/owner/repo"))
}

func TestRepo_IDString(t *testing.T) {
	Equals(t, "github.com/owner/repo", (valid.Repo{ID: "github.com/owner/repo"}).IDString())
	Equals(t, "/regex.*/", (valid.Repo{IDRegex: regexp.MustCompile("regex.*")}).IDString())
}

// String is a helper routine that allocates a new string value
// to store v and returns a pointer to it.
func String(v string) *string { return &v }

// Bool is a helper routine that allocates a new bool value
// to store v and returns a pointer to it.
func Bool(v bool) *bool { return &v }
