package runtime_test

import (
	"fmt"
	"testing"

	"github.com/runatlantis/atlantis/server/events/runtime"
	"github.com/runatlantis/atlantis/server/events/yaml/valid"
	. "github.com/runatlantis/atlantis/testing"
)

func TestGetPlanFilename(t *testing.T) {
	cases := []struct {
		workspace string
		maybeCfg  *valid.Project
		exp       string
	}{
		{
			"workspace",
			nil,
			"workspace.tfplan",
		},
		{
			"workspace",
			&valid.Project{},
			"workspace.tfplan",
		},
		{
			"workspace",
			&valid.Project{
				Name: String("project"),
			},
			"project-workspace.tfplan",
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			Equals(t, c.exp, runtime.GetPlanFilename(c.workspace, c.maybeCfg))
		})
	}
}

func String(v string) *string { return &v }
