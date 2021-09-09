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
	"github.com/stretchr/testify/assert"
)

func createProjectCommandOutputHandler(t *testing.T) handlers.ProjectCommandOutputHandler {
	logger := logging.NewNoopLogger(t)
	prjCmdOutputChan := make(chan *models.ProjectCmdOutputLine)
	return handlers.NewProjectCommandOutputHandler(prjCmdOutputChan, logger)
}

func TestProjectCommandOutputHandler(t *testing.T) {

	RepoName := "test-repo"
	RepoOwner := "test-org"
	RepoBaseBranch := "master"
	User := "test-user"
	Workspace := "myworkspace"
	RepoDir := "test-dir"
	ProjectName := "test-project"
	Msg := "Test Terraform Output"

	logger := logging.NewNoopLogger(t)
	ctx := models.ProjectCommandContext{
		BaseRepo: models.Repo{
			Name:  RepoName,
			Owner: RepoOwner,
		},
		HeadRepo: models.Repo{
			Name:  RepoName,
			Owner: RepoOwner,
		},
		Pull: models.PullRequest{
			Num:        1,
			HeadBranch: RepoBaseBranch,
			BaseBranch: RepoBaseBranch,
			Author:     User,
		},
		User: models.User{
			Username: User,
		},
		Log:         logger,
		Workspace:   Workspace,
		RepoRelDir:  RepoDir,
		ProjectName: ProjectName,
	}

	t.Run("Should Receive Message Sent in the ProjectCmdOutput channel", func(t *testing.T) {
		var wg sync.WaitGroup
		var expectedMsg string

		projectOutputHandler := createProjectCommandOutputHandler(t)
		go func() {
			projectOutputHandler.Handle()
		}()

		wg.Add(1)
		ch := make(chan string)
		go func() {
			err := projectOutputHandler.Receive(ctx.PullInfo(), ch, func(msg string) error {
				expectedMsg = msg
				wg.Done()
				return nil
			})
			Ok(t, err)
		}()

		projectOutputHandler.Send(ctx, Msg)

		// Wait for the msg to be read.
		wg.Wait()
		Equals(t, expectedMsg, Msg)
	})

	t.Run("Should Clear ProjectOutputBuffer when new Plan", func(t *testing.T) {
		var wg sync.WaitGroup

		projectOutputHandler := createProjectCommandOutputHandler(t)
		go func() {
			projectOutputHandler.Handle()
		}()

		wg.Add(1)
		ch := make(chan string)
		go func() {
			err := projectOutputHandler.Receive(ctx.PullInfo(), ch, func(msg string) error {
				wg.Done()
				return nil
			})
			Ok(t, err)
		}()

		projectOutputHandler.Send(ctx, Msg)

		// Wait for the msg to be read.
		wg.Wait()

		// Send a clear msg
		projectOutputHandler.Clear(ctx)

		dfProjectOutputHandler, ok := projectOutputHandler.(*handlers.DefaultProjectCommandOutputHandler)
		assert.True(t, ok)

		// Wait for the clear msg to be received by handle()
		time.Sleep(1 * time.Second)
		assert.Empty(t, dfProjectOutputHandler.GetProjectOutputBuffer(ctx.PullInfo()))
	})

	t.Run("Should Cleanup receiverBuffers receiving WS channel closed", func(t *testing.T) {
		var wg sync.WaitGroup

		projectOutputHandler := createProjectCommandOutputHandler(t)
		go func() {
			projectOutputHandler.Handle()
		}()

		wg.Add(1)
		ch := make(chan string)
		go func() {
			err := projectOutputHandler.Receive(ctx.PullInfo(), ch, func(msg string) error {
				wg.Done()
				return nil
			})
			Ok(t, err)
		}()

		projectOutputHandler.Send(ctx, Msg)

		// Wait for the msg to be read.
		wg.Wait()

		// Close chan to execute cleanup.
		close(ch)
		time.Sleep(1 * time.Second)

		dfProjectOutputHandler, ok := projectOutputHandler.(*handlers.DefaultProjectCommandOutputHandler)
		assert.True(t, ok)

		x := dfProjectOutputHandler.GetReceiverBufferForPull(ctx.PullInfo())
		assert.Empty(t, x)
	})

	t.Run("Should copy over existing log messages to new WS channels", func(t *testing.T) {
		var wg sync.WaitGroup

		projectOutputHandler := createProjectCommandOutputHandler(t)
		go func() {
			projectOutputHandler.Handle()
		}()

		wg.Add(1)
		ch := make(chan string)
		go func() {
			err := projectOutputHandler.Receive(ctx.PullInfo(), ch, func(msg string) error {
				fmt.Println(msg)
				wg.Done()
				return nil
			})
			Ok(t, err)
		}()

		projectOutputHandler.Send(ctx, Msg)

		// Wait for the msg to be read.
		wg.Wait()

		// Close channel to close prev connection.
		// This should close the first go routine with receive call.
		close(ch)

		ch = make(chan string)

		// Expecting two calls to callback.
		wg.Add(2)

		expectedMsg := []string{}
		go func() {
			err := projectOutputHandler.Receive(ctx.PullInfo(), ch, func(msg string) error {
				expectedMsg = append(expectedMsg, msg)
				wg.Done()
				return nil
			})
			Ok(t, err)
		}()

		// Make sure addChan gets the buffer lock and adds ch to the map.
		time.Sleep(1 * time.Second)

		projectOutputHandler.Send(ctx, Msg)

		// Wait for the message to be read.
		wg.Wait()
		close(ch)
		assert.Equal(t, []string{Msg, Msg}, expectedMsg)
	})
}
