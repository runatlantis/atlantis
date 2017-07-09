package server

import (
	"fmt"

	"strings"

	"github.com/hootsuite/atlantis/github"
	"github.com/hootsuite/atlantis/models"
)

type Status int

const (
	statusContext        = "Atlantis"
	Pending       Status = iota
	Success
	Failure
	Error
	PlanStep  = "plan"
	ApplyStep = "apply"
)

type GithubStatus struct {
	client *github.Client
}

func (s Status) String() string {
	switch s {
	case Pending:
		return "pending"
	case Success:
		return "success"
	case Failure:
		return "failure"
	case Error:
		return "error"
	}
	return "error"
}

func (g *GithubStatus) Update(repo models.Repo, pull models.PullRequest, status Status, step string) error {
	description := fmt.Sprintf("%s %s", strings.Title(step), strings.Title(status.String()))
	return g.client.UpdateStatus(repo, pull, status.String(), description, statusContext)
}

func (g *GithubStatus) UpdateProjectResult(ctx *CommandContext, projectResults []ProjectResult) error {
	var statuses []Status
	for _, p := range projectResults {
		statuses = append(statuses, p.Status())
	}
	worst := g.worstStatus(statuses)
	return g.Update(ctx.BaseRepo, ctx.Pull, worst, ctx.Command.Name.String())
}

func (g *GithubStatus) worstStatus(ss []Status) Status {
	worst := Success
	for _, s := range ss {
		if s > worst {
			worst = s
		}
	}
	return worst
}
