package server

import (
	"fmt"

	"strings"

	"github.com/google/go-github/github"
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
	client *GithubClient
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
	repoStatus := github.RepoStatus{
		State:       github.String(status.String()),
		Description: github.String(fmt.Sprintf("%s %s", strings.Title(step), strings.Title(status.String()))),
		Context:     github.String(statusContext)}
	return g.client.UpdateStatus(repo, pull, &repoStatus)
}

func (g *GithubStatus) UpdatePathResult(ctx *CommandContext, pathResults []PathResult) error {
	var statuses []Status
	for _, p := range pathResults {
		statuses = append(statuses, p.Status)
	}
	worst := g.worstStatus(statuses)
	return g.Update(ctx.Repo, ctx.Pull, worst, ctx.Command.commandType.String())
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
