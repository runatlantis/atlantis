package events_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/events/vcs/types"
	"github.com/stretchr/testify/assert"
)

type testJobURLGenerator struct {
	t        *testing.T
	expJobID string
	url      string
	err      error
}

func (t *testJobURLGenerator) GenerateProjectJobURL(jobID string) (string, error) {
	assert.Equal(t.t, t.expJobID, jobID)
	return t.url, t.err
}

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
func (t *testRenderer) RenderProject(prjRes command.ProjectResult, cmdName fmt.Stringer, baseRepo models.Repo) string {
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
	c.count++
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
			Command:     command.Plan,
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
				State:            models.SuccessVCSStatus,
				PullCreationTime: createdAt,
				Output:           output,
				PullNum:          1,

				CommandName: "Plan",
				Project:     "project1",
				Workspace:   "default",
				Directory:   "somedir",
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
			TitleBuilder: vcs.StatusTitleBuilder{TitlePrefix: "nish"},
			JobURLGenerator: &testJobURLGenerator{
				t:        t,
				expJobID: "",
				url:      "",
				err:      nil,
			},
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
			Command:     command.Plan,
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
				State:            models.FailedVCSStatus,
				PullCreationTime: createdAt,
				Output:           output,
				PullNum:          1,

				CommandName: "Plan",
				Project:     "project1",
				Workspace:   "default",
				Directory:   "somedir",
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
			TitleBuilder: vcs.StatusTitleBuilder{TitlePrefix: "nish"},
			JobURLGenerator: &testJobURLGenerator{
				t:        t,
				expJobID: "",
				url:      "",
				err:      nil,
			},
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
			Command:     command.Plan,
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
				State:            models.FailedVCSStatus,
				PullCreationTime: createdAt,
				Output:           output,
				PullNum:          1,

				CommandName: "Plan",
				Project:     "project1",
				Workspace:   "default",
				Directory:   "somedir",
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
			TitleBuilder: vcs.StatusTitleBuilder{TitlePrefix: "nish"},
			JobURLGenerator: &testJobURLGenerator{
				t:        t,
				expJobID: "",
				url:      "",
				err:      nil,
			},
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

	result := command.ProjectResult{
		ProjectName: "project1",
		RepoRelDir:  "somedir",
		Workspace:   "default",
		Command:     command.ApprovePolicies,
	}

	t.Run("handle approve policies", func(t *testing.T) {
		commandResult := command.Result{
			ProjectResults: []command.ProjectResult{
				result,
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
						State:            models.SuccessVCSStatus,
						PullCreationTime: createdAt,
						PullNum:          1,

						CommandName: "Approve Policies",
						Project:     "project1",
						Workspace:   "default",
						Directory:   "somedir",
					},
				},
			},
		}
		subject := events.ChecksOutputUpdater{
			VCSClient: client,
			MarkdownRenderer: &testRenderer{
				t:                     t,
				expectedCmdName:       command.ApprovePolicies,
				expectedResult:        commandResult,
				expectedRepo:          repo,
				expectedOutput:        output,
				expectedProjectResult: result,
			},
			TitleBuilder: vcs.StatusTitleBuilder{TitlePrefix: "nish"},
			JobURLGenerator: &testJobURLGenerator{
				t:        t,
				expJobID: "",
				url:      "",
				err:      nil,
			},
		}

		subject.UpdateOutput(cmdCtx, &command.Comment{
			Name: command.ApprovePolicies,
		}, commandResult)
	})
}

func TestChecksOutputUpdater_CommandFailure(t *testing.T) {
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
	t.Run("approve policies", func(t *testing.T) {
		errorString := "error"
		commandResult := command.Result{
			Error: errors.New(errorString),
		}

		client := &strictTestChecksClient{
			clients: []*testChecksClient{
				{
					t: t,
					expectedRequest: types.UpdateStatusRequest{
						Repo:             repo,
						Ref:              sha,
						StatusName:       "nish/policy_check",
						Output:           errorString,
						State:            models.FailedVCSStatus,
						PullCreationTime: createdAt,
						PullNum:          1,
						CommandName:      "Approve Policies",
					},
				},
			},
		}
		subject := events.ChecksOutputUpdater{
			VCSClient:    client,
			TitleBuilder: vcs.StatusTitleBuilder{TitlePrefix: "nish"},
			JobURLGenerator: &testJobURLGenerator{
				t:        t,
				expJobID: "",
				url:      "",
				err:      nil,
			},
		}

		subject.UpdateOutput(cmdCtx, command.Comment{
			Name: command.ApprovePolicies,
		}, commandResult)
	})

	t.Run("error", func(t *testing.T) {
		errorString := "error"
		commandResult := command.Result{
			Error: errors.New(errorString),
		}

		client := &strictTestChecksClient{
			clients: []*testChecksClient{
				{
					t: t,
					expectedRequest: types.UpdateStatusRequest{
						Repo:             repo,
						Ref:              sha,
						StatusName:       "nish/plan",
						Output:           errorString,
						State:            models.FailedVCSStatus,
						PullCreationTime: createdAt,
						PullNum:          1,
						CommandName:      "Plan",
					},
				},
			},
		}
		subject := events.ChecksOutputUpdater{
			VCSClient:    client,
			TitleBuilder: vcs.StatusTitleBuilder{TitlePrefix: "nish"},
			JobURLGenerator: &testJobURLGenerator{
				t:        t,
				expJobID: "",
				url:      "",
				err:      nil,
			},
		}

		subject.UpdateOutput(cmdCtx, events.AutoplanCommand{}, commandResult)
	})

	t.Run("failure", func(t *testing.T) {
		failureString := "error"
		commandResult := command.Result{
			Failure: failureString,
		}

		client := &strictTestChecksClient{
			clients: []*testChecksClient{
				{
					t: t,
					expectedRequest: types.UpdateStatusRequest{
						Repo:             repo,
						Ref:              sha,
						StatusName:       "nish/plan",
						Output:           failureString,
						State:            models.FailedVCSStatus,
						PullCreationTime: createdAt,
						PullNum:          1,
						CommandName:      "Plan",
					},
				},
			},
		}
		subject := events.ChecksOutputUpdater{
			VCSClient:    client,
			TitleBuilder: vcs.StatusTitleBuilder{TitlePrefix: "nish"},
			JobURLGenerator: &testJobURLGenerator{
				t:        t,
				expJobID: "",
				url:      "",
				err:      nil,
			},
		}

		subject.UpdateOutput(cmdCtx, events.AutoplanCommand{}, commandResult)
	})
}
