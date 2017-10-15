package events

import (
	"fmt"

	"strings"

	"github.com/hootsuite/atlantis/server/events/github"
	"github.com/hootsuite/atlantis/server/events/models"
)

type Status int

const (
	statusContext = "Atlantis"
)

const (
	Pending Status = iota
	Success
	Failure
	Error
)

//go:generate pegomock generate --use-experimental-model-gen --package mocks -o mocks/mock_github_status.go GHStatusUpdater
type GHStatusUpdater interface {
	Update(repo models.Repo, pull models.PullRequest, status Status, cmd *Command) error
	UpdateProjectResult(ctx *CommandContext, res CommandResponse) error
}

type GithubStatus struct {
	Client github.Client
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

func (g *GithubStatus) Update(repo models.Repo, pull models.PullRequest, status Status, cmd *Command) error {
	description := fmt.Sprintf("%s %s", strings.Title(cmd.Name.String()), strings.Title(status.String()))
	return g.Client.UpdateStatus(repo, pull, status.String(), description, statusContext)
}

func (g *GithubStatus) UpdateProjectResult(ctx *CommandContext, res CommandResponse) error {
	var status Status
	if res.Error != nil {
		status = Error
	} else if res.Failure != "" {
		status = Failure
	} else {
		var statuses []Status
		for _, p := range res.ProjectResults {
			statuses = append(statuses, p.Status())
		}
		status = g.worstStatus(statuses)
	}
	return g.Update(ctx.BaseRepo, ctx.Pull, status, ctx.Command)
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
