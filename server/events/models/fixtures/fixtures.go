package fixtures

import "github.com/atlantisnorth/atlantis/server/events/models"

var Pull = models.PullRequest{
	Num:        1,
	HeadCommit: "16ca62f65c18ff456c6ef4cacc8d4826e264bb17",
	Branch:     "branch",
	Author:     "lkysow",
	URL:        "url",
}

var Repo = models.Repo{
	CloneURL:          "https://user:password@github.com/atlantisnorth/atlantis.git",
	FullName:          "atlantisnorth/atlantis",
	Owner:             "atlantisnorth",
	SanitizedCloneURL: "https://github.com/atlantisnorth/atlantis.git",
	Name:              "atlantis",
}

var User = models.User{
	Username: "lkysow",
}
