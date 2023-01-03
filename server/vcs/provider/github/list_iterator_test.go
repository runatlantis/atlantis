package github_test

import (
	"context"
	gh "github.com/google/go-github/v45/github"
	"github.com/runatlantis/atlantis/server/vcs/provider/github"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

const (
	user1  = "user1"
	user2  = "user2"
	denied = "denied"
)

func TestIterate(t *testing.T) {
	testResults := []*gh.PullRequestReview{
		{
			User:  &gh.User{Login: gh.String(user1)},
			State: gh.String(github.ApprovalState),
		},
		{
			User:  &gh.User{Login: gh.String(user2)},
			State: gh.String(denied),
		},
	}
	testResponse := &gh.Response{
		Response: &http.Response{StatusCode: http.StatusOK},
	}
	run := func(ctx context.Context, nextPage int) ([]*gh.PullRequestReview, *gh.Response, error) {
		return testResults, testResponse, nil
	}
	results, err := github.Iterate(context.Background(), run)
	assert.NoError(t, err)
	assert.Equal(t, results, testResults)
}

func TestIterate_NotOKStatus(t *testing.T) {
	testResults := []*gh.PullRequestReview{
		{
			User:  &gh.User{Login: gh.String(user1)},
			State: gh.String(github.ApprovalState),
		},
	}
	testResponse := &gh.Response{
		Response: &http.Response{},
	}
	run := func(ctx context.Context, nextPage int) ([]*gh.PullRequestReview, *gh.Response, error) {
		return testResults, testResponse, nil
	}
	results, err := github.Iterate(context.Background(), run)
	assert.Error(t, err)
	assert.Nil(t, results)
}

func TestIterate_ErrorRun(t *testing.T) {
	testResults := []*gh.PullRequestReview{
		{
			User:  &gh.User{Login: gh.String(user1)},
			State: gh.String(github.ApprovalState),
		},
	}
	testResponse := &gh.Response{
		Response: &http.Response{StatusCode: http.StatusOK},
	}
	run := func(ctx context.Context, nextPage int) ([]*gh.PullRequestReview, *gh.Response, error) {
		return testResults, testResponse, assert.AnError
	}
	results, err := github.Iterate(context.Background(), run)
	assert.Error(t, err)
	assert.Nil(t, results)
}
