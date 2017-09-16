package server_test

import (
	"testing"
	"github.com/hootsuite/atlantis/server"
	. "github.com/hootsuite/atlantis/testing_util"
	"github.com/golang/mock/gomock"
	"github.com/hootsuite/atlantis/github/mocks"
	"github.com/hootsuite/atlantis/models"
	"errors"
	"strings"
)

var repoModel = models.Repo{}
var pullModel = models.PullRequest{}
var status = server.Success
var step = "step"

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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mock := mocks.NewMockClient(ctrl)
	mock.EXPECT().UpdateStatus(repoModel, pullModel, "success", "Step Success", "Atlantis")

	s := server.GithubStatus{mock}
	err := s.Update(repoModel, pullModel, status, step)
	Ok(t, err)
}

func TestUpdateProjectResult(t *testing.T) {
	t.Log("should use worst status")
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mock := mocks.NewMockClient(ctrl)
	ctx := &server.CommandContext{
		BaseRepo: repoModel,
		Pull:     pullModel,
		Command:  &server.Command{Name: server.Plan},
	}
	s := server.GithubStatus{mock}

	cases := []struct{
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

		mock.EXPECT().UpdateStatus(repoModel, pullModel, c.Expected, "Plan " + strings.Title(c.Expected), "Atlantis")
		s.UpdateProjectResult(ctx, results)
	}
}
