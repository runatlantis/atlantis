package handlers_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/handlers"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

func TestProjectCommandOutputHandler(t *testing.T) {
	t.Run("Should Group by Project Info", func(t *testing.T) {
		tempchan := make(chan *models.ProjectCmdOutputLine)
		logger := logging.NewNoopLogger(t)
		projectOutputHandler := handlers.NewProjectCommandOutputHandler(tempchan, logger)
		ctx := models.ProjectCommandContext{
			BaseRepo: models.Repo{
				Name:  "test-repo",
				Owner: "test-org",
			},
			HeadRepo: models.Repo{
				Name:  "test-repo",
				Owner: "test-org",
			},
			Pull: models.PullRequest{
				Num:        1,
				HeadBranch: "add-feat",
				BaseBranch: "master",
				Author:     "acme",
			},
			User: models.User{
				Username: "acme-user",
			},
			Log:         logger,
			Workspace:   "myworkspace",
			RepoRelDir:  "mydir",
			ProjectName: "test-project",
		}
		fmt.Printf("Testing Handle func in unit test")

		var wg sync.WaitGroup

		go func() {
			projectOutputHandler.Handle()
		}()

		go func() {
			projectOutputHandler.Send(ctx, "Test Terraform Output")
		}()

		expectMsg := ""
		wg.Add(1)
		go func() {
			err := projectOutputHandler.Receive(ctx.PullInfo(), func(msg string) error {
				expectMsg = msg
				wg.Done()
				return nil
			})
			Ok(t, err)
		}()
		wg.Wait()
		Equals(t, expectMsg, "Test Terraform Output")

		time.Sleep(1 * time.Second)

	})
}

func TestProjectCommandOutputHandler_ClearBuff(t *testing.T) {
	t.Run("Should Group by Project Info", func(t *testing.T) {
		tempchan := make(chan *models.ProjectCmdOutputLine)
		logger := logging.NewNoopLogger(t)
		projectOutputHandler := handlers.NewProjectCommandOutputHandler(tempchan, logger)
		ctx := models.ProjectCommandContext{
			BaseRepo: models.Repo{
				Name:  "test-repo",
				Owner: "test-org",
			},
			HeadRepo: models.Repo{
				Name:  "test-repo",
				Owner: "test-org",
			},
			Pull: models.PullRequest{
				Num:        1,
				HeadBranch: "add-feat",
				BaseBranch: "master",
				Author:     "acme",
			},
			User: models.User{
				Username: "acme-user",
			},
			Log:         logger,
			Workspace:   "myworkspace",
			RepoRelDir:  "mydir",
			ProjectName: "test-project",
		}
		fmt.Printf("Testing Handle func in unit test")

		var wg sync.WaitGroup

		go func() {
			projectOutputHandler.Handle()
		}()

		wg.Add(1)
		go func() {
			projectOutputHandler.Send(ctx, "Test Terraform Output")
		}()

		wg.Add(1)
		go func() {
			projectOutputHandler.Clear(ctx)
		}()

		wg.Add(1)
		go func() {
			projectOutputHandler.Send(ctx, "Test Terraform Output")
		}()

		expectMsg := ""
		go func() {
			err := projectOutputHandler.Receive(ctx.PullInfo(), func(msg string) error {
				expectMsg = msg
				wg.Done()
				return nil
			})
			Ok(t, err)
		}()
		wg.Wait()

		Equals(t, expectMsg, "Test Terraform Output")

		time.Sleep(1 * time.Second)

	})
}
