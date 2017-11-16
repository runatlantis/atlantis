package events_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/hootsuite/atlantis/server/events"
	"github.com/hootsuite/atlantis/server/events/models"
	"github.com/hootsuite/atlantis/server/events/vcs"
	"github.com/hootsuite/atlantis/server/events/vcs/mocks"
	. "github.com/hootsuite/atlantis/testing"
	. "github.com/petergtz/pegomock"
)

var repoModel = models.Repo{}
var pullModel = models.PullRequest{}
var status = vcs.Success
var cmd = events.Command{
	Name: events.Plan,
}

func TestStatus_String(t *testing.T) {
	cases := map[vcs.CommitStatus]string{
		vcs.Pending: "pending",
		vcs.Success: "success",
		vcs.Failed:  "failed",
	}
	for k, v := range cases {
		Equals(t, v, k.String())
	}
}

func TestUpdate(t *testing.T) {
	RegisterMockTestingT(t)
	client := mocks.NewMockClientProxy()
	s := events.DefaultCommitStatusUpdater{Client: client}
	err := s.Update(repoModel, pullModel, status, &cmd, vcs.Github)
	Ok(t, err)
	client.VerifyWasCalledOnce().UpdateStatus(repoModel, pullModel, status, "Plan Success", vcs.Github)
}

func TestUpdateProjectResult_Error(t *testing.T) {
	RegisterMockTestingT(t)
	ctx := &events.CommandContext{
		BaseRepo: repoModel,
		Pull:     pullModel,
		Command:  &events.Command{Name: events.Plan},
		VCSHost:  vcs.Github,
	}
	client := mocks.NewMockClientProxy()
	s := events.DefaultCommitStatusUpdater{Client: client}
	err := s.UpdateProjectResult(ctx, events.CommandResponse{Error: errors.New("err")})
	Ok(t, err)
	client.VerifyWasCalledOnce().UpdateStatus(repoModel, pullModel, vcs.Failed, "Plan Failed", vcs.Github)
}

func TestUpdateProjectResult_Failure(t *testing.T) {
	RegisterMockTestingT(t)
	ctx := &events.CommandContext{
		BaseRepo: repoModel,
		Pull:     pullModel,
		Command:  &events.Command{Name: events.Plan},
		VCSHost:  vcs.Github,
	}
	client := mocks.NewMockClientProxy()
	s := events.DefaultCommitStatusUpdater{Client: client}
	err := s.UpdateProjectResult(ctx, events.CommandResponse{Failure: "failure"})
	Ok(t, err)
	client.VerifyWasCalledOnce().UpdateStatus(repoModel, pullModel, vcs.Failed, "Plan Failed", vcs.Github)
}

func TestUpdateProjectResult(t *testing.T) {
	t.Log("should use worst status")
	RegisterMockTestingT(t)

	ctx := &events.CommandContext{
		BaseRepo: repoModel,
		Pull:     pullModel,
		Command:  &events.Command{Name: events.Plan},
		VCSHost:  vcs.Github,
	}

	cases := []struct {
		Statuses []string
		Expected vcs.CommitStatus
	}{
		{
			[]string{"success", "failure", "error"},
			vcs.Failed,
		},
		{
			[]string{"failure", "error", "success"},
			vcs.Failed,
		},
		{
			[]string{"success", "failure"},
			vcs.Failed,
		},
		{
			[]string{"success", "error"},
			vcs.Failed,
		},
		{
			[]string{"failure", "error"},
			vcs.Failed,
		},
		{
			[]string{"success"},
			vcs.Success,
		},
		{
			[]string{"success", "success"},
			vcs.Success,
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

		client := mocks.NewMockClientProxy()
		s := events.DefaultCommitStatusUpdater{Client: client}
		err := s.UpdateProjectResult(ctx, resp)
		Ok(t, err)
		client.VerifyWasCalledOnce().UpdateStatus(repoModel, pullModel, c.Expected, "Plan "+strings.Title(c.Expected.String()), vcs.Github)
	}
}
