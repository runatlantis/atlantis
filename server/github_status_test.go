package server_test

import (
	"testing"

	"errors"
	"strings"

	"github.com/hootsuite/atlantis/github/mocks"
	"github.com/hootsuite/atlantis/models"
	"github.com/hootsuite/atlantis/server"
	. "github.com/hootsuite/atlantis/testing_util"
	. "github.com/petergtz/pegomock"
)

var repoModel = models.Repo{}
var pullModel = models.PullRequest{}
var status = server.Success
var cmd = server.Command{
	Name: server.Plan,
}

func TestStatus_String(t *testing.T) {
	cases := map[server.Status]string{
		server.Pending: "pending",
		server.Success: "success",
		server.Failure: "failure",
		server.Error:   "error",
	}
	for k, v := range cases {
		Equals(t, v, k.String())
	}
}

func TestUpdate(t *testing.T) {
	RegisterMockTestingT(t)
	client := mocks.NewMockClient()
	s := server.GithubStatus{client}
	err := s.Update(repoModel, pullModel, status, &cmd)
	Ok(t, err)
	client.VerifyWasCalledOnce().UpdateStatus(repoModel, pullModel, "success", "Plan Success", "Atlantis")
}

func TestUpdateProjectResult_Error(t *testing.T) {
	RegisterMockTestingT(t)
	ctx := &server.CommandContext{
		BaseRepo: repoModel,
		Pull:     pullModel,
		Command:  &server.Command{Name: server.Plan},
	}
	client := mocks.NewMockClient()
	s := server.GithubStatus{client}
	s.UpdateProjectResult(ctx, server.CommandResponse{Error: errors.New("err")})
	client.VerifyWasCalledOnce().UpdateStatus(repoModel, pullModel, server.Error.String(), "Plan Error", "Atlantis")
}

func TestUpdateProjectResult_Failure(t *testing.T) {
	RegisterMockTestingT(t)
	ctx := &server.CommandContext{
		BaseRepo: repoModel,
		Pull:     pullModel,
		Command:  &server.Command{Name: server.Plan},
	}
	client := mocks.NewMockClient()
	s := server.GithubStatus{client}
	s.UpdateProjectResult(ctx, server.CommandResponse{Failure: "failure"})
	client.VerifyWasCalledOnce().UpdateStatus(repoModel, pullModel, server.Failure.String(), "Plan Failure", "Atlantis")
}

func TestUpdateProjectResult(t *testing.T) {
	t.Log("should use worst status")
	RegisterMockTestingT(t)

	ctx := &server.CommandContext{
		BaseRepo: repoModel,
		Pull:     pullModel,
		Command:  &server.Command{Name: server.Plan},
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
		var results []server.ProjectResult
		for _, statusStr := range c.Statuses {
			var result server.ProjectResult
			switch statusStr {
			case "failure":
				result = server.ProjectResult{Failure: "failure"}
			case "error":
				result = server.ProjectResult{Error: errors.New("err")}
			default:
				result = server.ProjectResult{}
			}
			results = append(results, result)
		}
		resp := server.CommandResponse{ProjectResults: results}

		client := mocks.NewMockClient()
		s := server.GithubStatus{client}
		s.UpdateProjectResult(ctx, resp)
		client.VerifyWasCalledOnce().UpdateStatus(repoModel, pullModel, c.Expected, "Plan "+strings.Title(c.Expected), "Atlantis")
	}
}
