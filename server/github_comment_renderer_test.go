package server_test

import (
	"errors"
	"testing"

	"github.com/hootsuite/atlantis/server"
	. "github.com/hootsuite/atlantis/testing_util"
)

func TestRenderErr(t *testing.T) {
	err := errors.New("err")
	cases := []struct {
		Description string
		Command     server.CommandName
		Error       error
		Expected    string
	}{
		{
			"apply error",
			server.Apply,
			err,
			"**Apply Error**\n```\nerr\n```\n\n",
		},
		{
			"plan error",
			server.Plan,
			err,
			"**Plan Error**\n```\nerr\n```\n\n",
		},
	}

	r := server.GithubCommentRenderer{}
	for _, c := range cases {
		res := server.CommandResponse{
			Command: c.Command,
			Error:   c.Error,
		}
		for _, verbose := range []bool{true, false} {
			t.Log("testing " + c.Description)
			s := r.Render(res, "log", verbose)
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
		Command     server.CommandName
		Failure     string
		Expected    string
	}{
		{
			"apply failure",
			server.Apply,
			"failure",
			"**Apply Failed**: failure\n\n",
		},
		{
			"plan failure",
			server.Plan,
			"failure",
			"**Plan Failed**: failure\n\n",
		},
	}

	r := server.GithubCommentRenderer{}
	for _, c := range cases {
		res := server.CommandResponse{
			Command: c.Command,
			Failure: c.Failure,
		}
		for _, verbose := range []bool{true, false} {
			t.Log("testing " + c.Description)
			s := r.Render(res, "log", verbose)
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
	r := server.GithubCommentRenderer{}
	res := server.CommandResponse{
		Command: server.Plan,
		Error:   errors.New("error"),
		Failure: "failure",
	}
	s := r.Render(res, "", false)
	Equals(t, "**Plan Error**\n```\nerror\n```\n\n", s)
}

func TestRenderProjectResults(t *testing.T) {
	cases := []struct {
		Description    string
		Command        server.CommandName
		ProjectResults []server.ProjectResult
		Expected       string
	}{
		{
			"single successful plan",
			server.Plan,
			[]server.ProjectResult{
				{
					PlanSuccess: &server.PlanSuccess{
						"terraform-output",
						"lock-url",
					},
				},
			},
			"```diff\nterraform-output\n```\n\n* To **discard** this plan click [here](lock-url).\n\n",
		},
		{
			"single successful apply",
			server.Apply,
			[]server.ProjectResult{
				{
					ApplySuccess: "success",
				},
			},
			"```diff\nsuccess\n```\n\n",
		},
		{
			"multiple successful plans",
			server.Plan,
			[]server.ProjectResult{
				{
					Path: "path",
					PlanSuccess: &server.PlanSuccess{
						"terraform-output",
						"lock-url",
					},
				},
				{
					Path: "path2",
					PlanSuccess: &server.PlanSuccess{
						"terraform-output2",
						"lock-url2",
					},
				},
			},
			"Ran Plan in 2 directories:\n * `path`\n * `path2`\n\n## path/\n```diff\nterraform-output\n```\n\n* To **discard** this plan click [here](lock-url).\n---\n## path2/\n```diff\nterraform-output2\n```\n\n* To **discard** this plan click [here](lock-url2).\n---\n\n",
		},
		{
			"multiple successful applies",
			server.Apply,
			[]server.ProjectResult{
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
			server.Plan,
			[]server.ProjectResult{
				{
					Error: errors.New("error"),
				},
			},
			"**Plan Error**\n```\nerror\n```\n\n\n",
		},
		{
			"single failed plan",
			server.Plan,
			[]server.ProjectResult{
				{
					Failure: "failure",
				},
			},
			"**Plan Failed**: failure\n\n\n",
		},
		{
			"successful, failed, and errored plan",
			server.Plan,
			[]server.ProjectResult{
				{
					Path: "path",
					PlanSuccess: &server.PlanSuccess{
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
			server.Apply,
			[]server.ProjectResult{
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

	r := server.GithubCommentRenderer{}
	for _, c := range cases {
		res := server.CommandResponse{
			Command:        c.Command,
			ProjectResults: c.ProjectResults,
		}
		for _, verbose := range []bool{true, false} {
			t.Log("testing " + c.Description)
			s := r.Render(res, "log", verbose)
			if !verbose {
				Equals(t, c.Expected, s)
			} else {
				Equals(t, c.Expected+"<details><summary>Log</summary>\n  <p>\n\n```\nlog```\n</p></details>\n", s)
			}
		}
	}
}
