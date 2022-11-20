package command_test

import (
	"errors"
	"testing"

	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	. "github.com/runatlantis/atlantis/testing"
)

func TestCommandResult_HasErrors(t *testing.T) {
	cases := map[string]struct {
		cr  command.Result
		exp bool
	}{
		"error": {
			cr: command.Result{
				Error: errors.New("err"),
			},
			exp: true,
		},
		"failure": {
			cr: command.Result{
				Failure: "failure",
			},
			exp: true,
		},
		"empty results list": {
			cr: command.Result{
				ProjectResults: []command.ProjectResult{},
			},
			exp: false,
		},
		"successful plan": {
			cr: command.Result{
				ProjectResults: []command.ProjectResult{
					{
						PlanSuccess: &models.PlanSuccess{},
					},
				},
			},
			exp: false,
		},
		"successful apply": {
			cr: command.Result{
				ProjectResults: []command.ProjectResult{
					{
						ApplySuccess: "success",
					},
				},
			},
			exp: false,
		},
		"single errored project": {
			cr: command.Result{
				ProjectResults: []command.ProjectResult{
					{
						Error: errors.New("err"),
					},
				},
			},
			exp: true,
		},
		"single failed project": {
			cr: command.Result{
				ProjectResults: []command.ProjectResult{
					{
						Failure: "failure",
					},
				},
			},
			exp: true,
		},
		"two successful projects": {
			cr: command.Result{
				ProjectResults: []command.ProjectResult{
					{
						PlanSuccess: &models.PlanSuccess{},
					},
					{
						ApplySuccess: "success",
					},
				},
			},
			exp: false,
		},
		"one successful, one failed project": {
			cr: command.Result{
				ProjectResults: []command.ProjectResult{
					{
						PlanSuccess: &models.PlanSuccess{},
					},
					{
						Failure: "failed",
					},
				},
			},
			exp: true,
		},
	}

	for descrip, c := range cases {
		t.Run(descrip, func(t *testing.T) {
			Equals(t, c.exp, c.cr.HasErrors())
		})
	}
}
