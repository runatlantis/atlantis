package main

import (
	"context"

	"github.com/google/go-github/github"
)

type GithubClient struct {
	client   *github.Client
	ctx      context.Context
	username string
}
