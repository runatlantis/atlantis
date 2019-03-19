package runtime_test

import (
	"fmt"
	"testing"

	"github.com/runatlantis/atlantis/server/events/runtime"
	. "github.com/runatlantis/atlantis/testing"
)

func TestGetPlanFilename(t *testing.T) {
	cases := []struct {
		workspace   string
		projectName string
		exp         string
	}{
		{
			"workspace",
			"",
			"workspace.tfplan",
		},
		{
			"workspace",
			"",
			"workspace.tfplan",
		},
		{
			"workspace",
			"project",
			"project-workspace.tfplan",
		},
		{
			"workspace",
			"project/with/slash",
			"project-with-slash-workspace.tfplan",
		},
		{
			"workspace",
			"project with space",
			"project with space-workspace.tfplan",
		},
		{
			"workspace😀",
			"project😀",
			"project😀-workspace😀.tfplan",
		},
		{
			"default",
			`all.invalid.chars \/"*?<>`,
			"all.invalid.chars --------default.tfplan",
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			Equals(t, c.exp, runtime.GetPlanFilename(c.workspace, c.projectName))
		})
	}
}
