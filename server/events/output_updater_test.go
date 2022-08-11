package events_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/events/vcs/types"
	"github.com/stretchr/testify/assert"
)

type testRenderer struct {
	t                     *testing.T
	expectedResult        command.Result
	expectedCmdName       command.Name
	expectedRepo          models.Repo
	expectedProjectResult command.ProjectResult

	expectedOutput string
}

func (t *testRenderer) Render(res command.Result, cmdName command.Name, baseRepo models.Repo) string {
	assert.Equal(t.t, t.expectedResult, res)
	assert.Equal(t.t, t.expectedCmdName, cmdName)
	assert.Equal(t.t, t.expectedRepo, baseRepo)

	return t.expectedOutput
}
func (t *testRenderer) RenderProject(prjRes command.ProjectResult, cmdName command.Name, baseRepo models.Repo) string {
	assert.Equal(t.t, t.expectedProjectResult, prjRes)
	assert.Equal(t.t, t.expectedCmdName, cmdName)
	assert.Equal(t.t, t.expectedRepo, baseRepo)

	return t.expectedOutput
}

// this allows you to make multiple calls to the checks client with different results/requests
type strictTestChecksClient struct {
	clients []*testChecksClient

	count int
}

func (c *strictTestChecksClient) UpdateStatus(ctx context.Context, request types.UpdateStatusRequest) (string, error) {
	if c.count > (len(c.clients) - 1) {
		return "", errors.New("more calls than expected")
	}
	_, err := c.clients[c.count].UpdateStatus(ctx, request)
	c.count += 1
	return "", err
}

type testChecksClient struct {
	t               *testing.T
	expectedRequest types.UpdateStatusRequest
	expectedError   error

	called bool
}

func (c *testChecksClient) UpdateStatus(ctx context.Context, request types.UpdateStatusRequest) (string, error) {
	c.called = true
	assert.Equal(c.t, c.expectedRequest, request)

	return "", c.expectedError
}

func TestChecksOutputUpdater_ProjectResults(t *testing.T) {

	repo := models.Repo{
		FullName: "nish/repo",
	}

	createdAt := time.Now()
	sha := "12345"

	pull := models.PullRequest{
		HeadCommit: sha,
		Num:        1,
		CreatedAt:  createdAt,
		BaseRepo:   repo,
	}

	cmdCtx := &command.Context{
		Pull:       pull,
		RequestCtx: context.Background(),
		HeadRepo:   repo,
	}

	output := "some output"

	t.Run("project result success", func(t *testing.T) {
		projectResult := command.ProjectResult{
			ProjectName: "project1",
			RepoRelDir:  "somedir",
			Workspace:   "default",
		}
		commandResult := command.Result{
			ProjectResults: []command.ProjectResult{
				projectResult,
			},
		}

		client := &testChecksClient{
			t: t,
			expectedRequest: types.UpdateStatusRequest{
				Repo:             repo,
				Ref:              sha,
				StatusName:       "nish/plan: project1",
				Description:      "**Project**: `project1`\n**Dir**: `somedir`\n**Workspace**: `default`",
				State:            models.SuccessCommitStatus,
				PullCreationTime: createdAt,
				Output:           output,
				PullNum:          1,
			},
		}
		subject := events.ChecksOutputUpdater{
			VCSClient: client,
			MarkdownRenderer: &testRenderer{
				t:                     t,
				expectedCmdName:       command.Plan,
				expectedResult:        commandResult,
				expectedRepo:          repo,
				expectedOutput:        output,
				expectedProjectResult: projectResult,
			},
			TitleBuilder: vcs.StatusTitleBuilder{"nish"},
		}

		subject.UpdateOutput(cmdCtx, events.AutoplanCommand{}, commandResult)

		assert.True(t, client.called)
	})

	t.Run("project result error", func(t *testing.T) {
		projectResult := command.ProjectResult{
			ProjectName: "project1",
			RepoRelDir:  "somedir",
			Workspace:   "default",
			Error:       assert.AnError,
		}
		commandResult := command.Result{
			ProjectResults: []command.ProjectResult{
				projectResult,
			},
		}

		client := &testChecksClient{
			t: t,
			expectedRequest: types.UpdateStatusRequest{
				Repo:             repo,
				Ref:              sha,
				StatusName:       "nish/plan: project1",
				Description:      "**Project**: `project1`\n**Dir**: `somedir`\n**Workspace**: `default`",
				State:            models.FailedCommitStatus,
				PullCreationTime: createdAt,
				Output:           output,
				PullNum:          1,
			},
		}
		subject := events.ChecksOutputUpdater{
			VCSClient: client,
			MarkdownRenderer: &testRenderer{
				t:                     t,
				expectedCmdName:       command.Plan,
				expectedResult:        commandResult,
				expectedRepo:          repo,
				expectedOutput:        output,
				expectedProjectResult: projectResult,
			},
			TitleBuilder: vcs.StatusTitleBuilder{"nish"},
		}

		subject.UpdateOutput(cmdCtx, events.AutoplanCommand{}, commandResult)

		assert.True(t, client.called)
	})

	t.Run("project result failure", func(t *testing.T) {
		projectResult := command.ProjectResult{
			ProjectName: "project1",
			RepoRelDir:  "somedir",
			Workspace:   "default",
			Failure:     "failure",
		}
		commandResult := command.Result{
			ProjectResults: []command.ProjectResult{
				projectResult,
			},
		}

		client := &testChecksClient{
			t: t,
			expectedRequest: types.UpdateStatusRequest{
				Repo:             repo,
				Ref:              sha,
				StatusName:       "nish/plan: project1",
				Description:      "**Project**: `project1`\n**Dir**: `somedir`\n**Workspace**: `default`",
				State:            models.FailedCommitStatus,
				PullCreationTime: createdAt,
				Output:           output,
				PullNum:          1,
			},
		}
		subject := events.ChecksOutputUpdater{
			VCSClient: client,
			MarkdownRenderer: &testRenderer{
				t:                     t,
				expectedCmdName:       command.Plan,
				expectedResult:        commandResult,
				expectedRepo:          repo,
				expectedOutput:        output,
				expectedProjectResult: projectResult,
			},
			TitleBuilder: vcs.StatusTitleBuilder{"nish"},
		}

		subject.UpdateOutput(cmdCtx, events.AutoplanCommand{}, commandResult)

		assert.True(t, client.called)
	})

}

func TestChecksOutputUpdater_ProjectResults_ApprovePolicies(t *testing.T) {
	repo := models.Repo{
		FullName: "nish/repo",
	}

	createdAt := time.Now()
	sha := "12345"

	pull := models.PullRequest{
		HeadCommit: sha,
		Num:        1,
		CreatedAt:  createdAt,
		BaseRepo:   repo,
	}

	cmdCtx := &command.Context{
		Pull:       pull,
		RequestCtx: context.Background(),
		HeadRepo:   repo,
	}

	output := "some output"

	t.Run("project result success", func(t *testing.T) {
		projectResult := command.ProjectResult{
			ProjectName:        "project1",
			RepoRelDir:         "somedir",
			Workspace:          "default",
			PolicyCheckSuccess: &models.PolicyCheckSuccess{},
		}
		commandResult := command.Result{
			ProjectResults: []command.ProjectResult{
				projectResult,
			},
		}

		client := &strictTestChecksClient{
			clients: []*testChecksClient{
				{
					t: t,
					expectedRequest: types.UpdateStatusRequest{
						Repo:             repo,
						Ref:              sha,
						StatusName:       "nish/policy_check: project1",
						Output:           "some output",
						State:            models.SuccessCommitStatus,
						PullCreationTime: createdAt,
						PullNum:          1,
					},
				},
			},
		}
		subject := events.ChecksOutputUpdater{
			VCSClient: client,
			MarkdownRenderer: &testRenderer{
				t:                     t,
				expectedCmdName:       command.PolicyCheck,
				expectedResult:        commandResult,
				expectedRepo:          repo,
				expectedOutput:        output,
				expectedProjectResult: projectResult,
			},
			TitleBuilder: vcs.StatusTitleBuilder{"nish"},
		}

		subject.UpdateOutput(cmdCtx, &command.Comment{
			Name: command.ApprovePolicies,
		}, commandResult)
	})
}
