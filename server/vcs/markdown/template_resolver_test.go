package markdown_test

import (
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/terraform/filter"
	"github.com/runatlantis/atlantis/server/vcs/markdown"
	"github.com/stretchr/testify/assert"
	"regexp"
	"strings"
	"testing"
)

func TestRenderer_RenderProject_Filter(t *testing.T) {
	cases := []struct {
		Name          string
		CommmandName  command.Name
		Regex         string
		Expected      string
		ProjectResult command.ProjectResult
	}{
		{
			Name:         "error render",
			CommmandName: command.Plan,
			Regex:        "foo",
			Expected:     "**Plan Error**\n```\nbar\nbaz\n```",
			ProjectResult: command.ProjectResult{
				Error: errors.New("foo\nbar\nbaz"),
			},
		},
		{
			Name:         "apply render",
			CommmandName: command.Apply,
			Regex:        "bar",
			Expected:     "```diff\nfoo\nbaz\n```",
			ProjectResult: command.ProjectResult{
				ApplySuccess: "foo\nbar\nbaz",
			},
		},
		{
			Name:         "plan render",
			CommmandName: command.Plan,
			Regex:        "baz",
			Expected: `$$$diff
foo
bar
$$$

* :arrow_forward: To **apply** this plan, comment:
    * $$
* :put_litter_in_its_place: To **delete** this plan click [here]()
* :repeat: To **plan** this project again, comment:
    * $$
`,
			ProjectResult: command.ProjectResult{
				PlanSuccess: &models.PlanSuccess{
					TerraformOutput: "foo\nbar\nbaz",
				},
			},
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			renderer := markdown.Renderer{
				TemplateResolver: markdown.TemplateResolver{
					LogFilter: filter.LogFilter{
						Regexes: []*regexp.Regexp{
							regexp.MustCompile(c.Regex),
						},
					},
				},
			}
			actual := renderer.RenderProject(c.ProjectResult, c.CommmandName, models.Repo{})
			expWithBackticks := strings.ReplaceAll(c.Expected, "$", "`")
			assert.Equal(t, expWithBackticks, actual)
		})
	}
}
