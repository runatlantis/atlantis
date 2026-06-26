// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.

package testdata

import (
	"fmt"

	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/events/models"
)

var Pull = models.PullRequest{
	Num:        1,
	HeadCommit: "16ca62f65c18ff456c6ef4cacc8d4826e264bb17",
	HeadBranch: "branch",
	Author:     "lkysow",
	URL:        "url",
}

var GithubRepo = models.Repo{
	CloneURL:          "https://user:password@github.com/runatlantis/atlantis.git",
	FullName:          "runatlantis/atlantis",
	Owner:             "runatlantis",
	SanitizedCloneURL: "https://github.com/runatlantis/atlantis.git",
	Name:              "atlantis",
	VCSHost: models.VCSHost{
		Hostname: "github.com",
		Type:     models.Github,
	},
}

var GitlabRepo = models.Repo{
	CloneURL:          "https://user:password@github.com/runatlantis/atlantis.git",
	FullName:          "runatlantis/atlantis",
	Owner:             "runatlantis",
	SanitizedCloneURL: "https://gitlab.com/runatlantis/atlantis.git",
	Name:              "atlantis",
	VCSHost: models.VCSHost{
		Hostname: "gitlab.com",
		Type:     models.Gitlab,
	},
}

var User = models.User{
	Username: "lkysow",
}

var projectName = "test-project"

var Project = valid.Project{
	Name: &projectName,
}

var PullInfo = fmt.Sprintf("%s/%d/%s", GithubRepo.FullName, Pull.Num, *Project.Name)
