package runtime_test

import (
	"fmt"
	"testing"

	"github.com/runatlantis/atlantis/server/core/runtime"
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
			"project",
			"project-workspace.tfplan",
		},
		{
			"workspace",
			"project/with/slash",
			"project::with::slash-workspace.tfplan",
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
		// Previously we replaced invalid chars with -'s, however we now
		// rely on validation of the atlantis.yaml file to ensure the name's
		// don't contain chars that need to be url encoded. So now these
		// chars shouldn't get replaced.
		{
			"default",
			`all.invalid.chars \/"*?<>`,
			"all.invalid.chars \\::\"*?<>-default.tfplan",
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			Equals(t, c.exp, runtime.GetPlanFilename(c.workspace, c.projectName))
		})
	}
}

func TestProjectNameFromPlanfile(t *testing.T) {
	cases := []struct {
		workspace string
		filename  string
		exp       string
	}{
		{
			"workspace",
			"workspace.tfplan",
			"",
		},
		{
			"workspace",
			"project-workspace.tfplan",
			"project",
		},
		{
			"workspace",
			"project-workspace-workspace.tfplan",
			"project-workspace",
		},
		{
			"workspace",
			"project::with::slashes::-workspace.tfplan",
			"project/with/slashes/",
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			act, err := runtime.ProjectNameFromPlanfile(c.workspace, c.filename)
			Ok(t, err)
			Equals(t, c.exp, act)
		})
	}
}
