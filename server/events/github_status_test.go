package events_test

import (
	"testing"

	"errors"
	"strings"

	"github.com/hootsuite/atlantis/server/github/mocks"
	"github.com/hootsuite/atlantis/server/models"
	. "github.com/hootsuite/atlantis/testing_util"
	. "github.com/petergtz/pegomock"
	"github.com/hootsuite/atlantis/server/events"
)

var repoModel = models.Repo{}
var pullModel = models.PullRequest{}
var status = events.Success
var cmd = events.Command{
	Name: events.Plan,
}

func TestStatus_String(t *testing.T) {
	cases := map[events.Status]string{
		events.Pending: "pending",
		events.Success: "success",
		events.Failure: "failure",
		events.Error:   "error",
	}
	for k, v := range cases {
		Equals(t, v, k.String())
	}
}

func TestUpdate(t *testing.T) {
	RegisterMockTestingT(t)
	client := mocks.NewMockClient()
	s := events.GithubStatus{client}
	err := s.Update(repoModel, pullModel, status, &cmd)
	Ok(t, err)
	client.VerifyWasCalledOnce().UpdateStatus(repoModel, pullModel, "success", "Plan Success", "Atlantis")
}

func TestUpdateProjectResult_Error(t *testing.T) {
	RegisterMockTestingT(t)
	ctx := &events.CommandContext{
		BaseRepo: repoModel,
		Pull:     pullModel,
		Command:  &events.Command{Name: events.Plan},
	}
	client := mocks.NewMockClient()
	s := events.GithubStatus{client}
	s.UpdateProjectResult(ctx, events.CommandResponse{Error: errors.New("err")})
	client.VerifyWasCalledOnce().UpdateStatus(repoModel, pullModel, events.Error.String(), "Plan Error", "Atlantis")
}

func TestUpdateProjectResult_Failure(t *testing.T) {
	RegisterMockTestingT(t)
	ctx := &events.CommandContext{
		BaseRepo: repoModel,
		Pull:     pullModel,
		Command:  &events.Command{Name: events.Plan},
	}
	client := mocks.NewMockClient()
	s := events.GithubStatus{client}
	s.UpdateProjectResult(ctx, events.CommandResponse{Failure: "failure"})
	client.VerifyWasCalledOnce().UpdateStatus(repoModel, pullModel, events.Failure.String(), "Plan Failure", "Atlantis")
}

func TestUpdateProjectResult(t *testing.T) {
	t.Log("should use worst status")
	RegisterMockTestingT(t)

	ctx := &events.CommandContext{
		BaseRepo: repoModel,
		Pull:     pullModel,
		Command:  &events.Command{Name: events.Plan},
	}

	cases := []struct {
		Statuses []string
		Expected string
	}{
		{
			[]string{"success", "failure", "error"},
			"error",
		},
		{
			[]string{"failure", "error", "success"},
			"error",
		},
		{
			[]string{"success", "failure"},
			"failure",
		},
		{
			[]string{"success", "error"},
			"error",
		},
		{
			[]string{"failure", "error"},
			"error",
		},
	}

	for _, c := range cases {
		var results []events.ProjectResult
		for _, statusStr := range c.Statuses {
			var result events.ProjectResult
			switch statusStr {
			case "failure":
				result = events.ProjectResult{Failure: "failure"}
			case "error":
				result = events.ProjectResult{Error: errors.New("err")}
			default:
				result = events.ProjectResult{}
			}
			results = append(results, result)
		}
		resp := events.CommandResponse{ProjectResults: results}

		client := mocks.NewMockClient()
		s := events.GithubStatus{client}
		s.UpdateProjectResult(ctx, resp)
		client.VerifyWasCalledOnce().UpdateStatus(repoModel, pullModel, c.Expected, "Plan "+strings.Title(c.Expected), "Atlantis")
	}
}
