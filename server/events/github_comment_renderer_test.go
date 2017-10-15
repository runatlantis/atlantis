package events_test

	import (
		"errors"
	"testing"

	. "github.com/hootsuite/atlantis/testing_util"
	"github.com/hootsuite/atlantis/server/events"
)

func TestRenderErr(t *testing.T) {
	err := errors.New("err")
	cases := []struct {
		Description string
		Command     events.CommandName
		Error       error
		Expected    string
	}{
		{
			"apply error",
			events.Apply,
			err,
			"**Apply Error**\n```\nerr\n```\n\n",
		},
		{
			"plan error",
			events.Plan,
			err,
			"**Plan Error**\n```\nerr\n```\n\n",
		},
	}

	r := events.GithubCommentRenderer{}
	for _, c := range cases {
		res := events.CommandResponse{
			Error: c.Error,
		}
		for _, verbose := range []bool{true, false} {
			t.Log("testing " + c.Description)
			s := r.Render(res, c.Command, "log", verbose)
			if !verbose {
				Equals(t, c.Expected, s)
			} else {
				Equals(t, c.Expected+"<details><summary>Log</summary>\n  <p>\n\n```\nlog```\n</p></details>\n", s)
			}
		}
	}
}

func TestRenderFailure(t *testing.T) {
	cases := []struct {
		Description string
		Command     events.CommandName
		Failure     string
		Expected    string
	}{
		{
			"apply failure",
			events.Apply,
			"failure",
			"**Apply Failed**: failure\n\n",
		},
		{
			"plan failure",
			events.Plan,
			"failure",
			"**Plan Failed**: failure\n\n",
		},
	}

	r := events.GithubCommentRenderer{}
	for _, c := range cases {
		res := events.CommandResponse{
			Failure: c.Failure,
		}
		for _, verbose := range []bool{true, false} {
			t.Log("testing " + c.Description)
			s := r.Render(res, c.Command, "log", verbose)
			if !verbose {
				Equals(t, c.Expected, s)
			} else {
				Equals(t, c.Expected+"<details><summary>Log</summary>\n  <p>\n\n```\nlog```\n</p></details>\n", s)
			}
		}
	}
}

func TestRenderErrAndFailure(t *testing.T) {
	t.Log("if there is an error and a failure, the error should be printed")
	r := events.GithubCommentRenderer{}
	res := events.CommandResponse{
		Error:   errors.New("error"),
		Failure: "failure",
	}
	s := r.Render(res, events.Plan, "", false)
	Equals(t, "**Plan Error**\n```\nerror\n```\n\n", s)
}

func TestRenderProjectResults(t *testing.T) {
	cases := []struct {
		Description    string
		Command        events.CommandName
		ProjectResults []events.ProjectResult
		Expected       string
	}{
		{
			"single successful plan",
			events.Plan,
			[]events.ProjectResult{
				{
					PlanSuccess: &events.PlanSuccess{
						"terraform-output",
						"lock-url",
					},
				},
			},
			"```diff\nterraform-output\n```\n\n* To **discard** this plan click [here](lock-url).\n\n",
		},
		{
			"single successful apply",
			events.Apply,
			[]events.ProjectResult{
				{
					ApplySuccess: "success",
				},
			},
			"```diff\nsuccess\n```\n\n",
		},
		{
			"multiple successful plans",
			events.Plan,
			[]events.ProjectResult{
				{
					Path: "path",
					PlanSuccess: &events.PlanSuccess{
						"terraform-output",
						"lock-url",
					},
				},
				{
					Path: "path2",
					PlanSuccess: &events.PlanSuccess{
						"terraform-output2",
						"lock-url2",
					},
				},
			},
			"Ran Plan in 2 directories:\n * `path`\n * `path2`\n\n## path/\n```diff\nterraform-output\n```\n\n* To **discard** this plan click [here](lock-url).\n---\n## path2/\n```diff\nterraform-output2\n```\n\n* To **discard** this plan click [here](lock-url2).\n---\n\n",
		},
		{
			"multiple successful applies",
			events.Apply,
			[]events.ProjectResult{
				{
					Path:         "path",
					ApplySuccess: "success",
				},
				{
					Path:         "path2",
					ApplySuccess: "success2",
				},
			},
			"Ran Apply in 2 directories:\n * `path`\n * `path2`\n\n## path/\n```diff\nsuccess\n```\n---\n## path2/\n```diff\nsuccess2\n```\n---\n\n",
		},
		{
			"single errored plan",
			events.Plan,
			[]events.ProjectResult{
				{
					Error: errors.New("error"),
				},
			},
			"**Plan Error**\n```\nerror\n```\n\n\n",
		},
		{
			"single failed plan",
			events.Plan,
			[]events.ProjectResult{
				{
					Failure: "failure",
				},
			},
			"**Plan Failed**: failure\n\n\n",
		},
		{
			"successful, failed, and errored plan",
			events.Plan,
			[]events.ProjectResult{
				{
					Path: "path",
					PlanSuccess: &events.PlanSuccess{
						"terraform-output",
						"lock-url",
					},
				},
				{
					Path:    "path2",
					Failure: "failure",
				},
				{
					Path:  "path3",
					Error: errors.New("error"),
				},
			},
			"Ran Plan in 3 directories:\n * `path`\n * `path2`\n * `path3`\n\n## path/\n```diff\nterraform-output\n```\n\n* To **discard** this plan click [here](lock-url).\n---\n## path2/\n**Plan Failed**: failure\n\n---\n## path3/\n**Plan Error**\n```\nerror\n```\n\n---\n\n",
		},
		{
			"successful, failed, and errored apply",
			events.Apply,
			[]events.ProjectResult{
				{
					Path:         "path",
					ApplySuccess: "success",
				},
				{
					Path:    "path2",
					Failure: "failure",
				},
				{
					Path:  "path3",
					Error: errors.New("error"),
				},
			},
			"Ran Apply in 3 directories:\n * `path`\n * `path2`\n * `path3`\n\n## path/\n```diff\nsuccess\n```\n---\n## path2/\n**Apply Failed**: failure\n\n---\n## path3/\n**Apply Error**\n```\nerror\n```\n\n---\n\n",
		},
	}

	r := events.GithubCommentRenderer{}
	for _, c := range cases {
		res := events.CommandResponse{
			ProjectResults: c.ProjectResults,
		}
		for _, verbose := range []bool{true, false} {
			t.Log("testing " + c.Description)
			s := r.Render(res, c.Command, "log", verbose)
			if !verbose {
				Equals(t, c.Expected, s)
			} else {
				Equals(t, c.Expected+"<details><summary>Log</summary>\n  <p>\n\n```\nlog```\n</p></details>\n", s)
			}
		}
	}
}
