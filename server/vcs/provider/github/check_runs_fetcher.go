package github

import (
	"context"
	gh "github.com/google/go-github/v45/github"
	"github.com/palantir/go-githubapp/githubapp"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/models"
	"regexp"
)

const (
	CompletedStatus  = "completed"
	FailedConclusion = "failed"
)

var checkRunRegex = regexp.MustCompile("atlantis/policy_check: .*")

type CheckRunsFetcher struct {
	ClientCreator githubapp.ClientCreator
	AppID         int64
}

func (r *CheckRunsFetcher) ListFailedPolicyCheckRuns(ctx context.Context, installationToken int64, repo models.Repo, ref string) ([]string, error) {
	client, err := r.ClientCreator.NewInstallationClient(installationToken)
	if err != nil {
		return nil, errors.Wrap(err, "creating installation client")
	}
	run := func(ctx context.Context, nextPage int) ([]*gh.CheckRun, *gh.Response, error) {
		listOptions := gh.ListCheckRunsOptions{
			Status: gh.String(CompletedStatus),
			AppID:  gh.Int64(r.AppID),
			ListOptions: gh.ListOptions{
				PerPage: 100,
			},
		}
		listOptions.Page = nextPage
		checkRunResults, resp, err := client.Checks.ListCheckRunsForRef(ctx, repo.Owner, repo.Name, ref, &listOptions)
		if checkRunResults != nil {
			return checkRunResults.CheckRuns, resp, err
		}
		return nil, nil, errors.New("unable to retrieve check runs from GH check run results")
	}

	checkRuns, err := Iterate(ctx, run)
	if err != nil {
		return nil, errors.Wrap(err, "iterating through entries")
	}
	var failedPolicyCheckRuns []string
	for _, checkRun := range checkRuns {
		if checkRunRegex.MatchString(checkRun.GetName()) && checkRun.GetConclusion() == FailedConclusion {
			failedPolicyCheckRuns = append(failedPolicyCheckRuns, checkRun.GetName())
		}
	}
	return failedPolicyCheckRuns, nil
}
