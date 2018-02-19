package fixtures

import "github.com/runatlantis/atlantis/server/events/models"

var Pull = models.PullRequest{
	Num:        1,
	HeadCommit: "16ca62f65c18ff456c6ef4cacc8d4826e264bb17",
	Branch:     "branch",
	Author:     "lkysow",
	URL:        "url",
}

var Repo = models.Repo{
	CloneURL:          "https://user:password@github.com/runatlantis/atlantis.git",
	FullName:          "runatlantis/atlantis",
	Owner:             "runatlantis",
	SanitizedCloneURL: "https://github.com/runatlantis/atlantis.git",
	Name:              "atlantis",
}

var User = models.User{
	Username: "lkysow",
}
