package server_test

import (
	"log"
	"os"
	"testing"

	"github.com/hootsuite/atlantis/github/mocks"
	"github.com/hootsuite/atlantis/logging"
	"github.com/hootsuite/atlantis/models"
	"github.com/hootsuite/atlantis/server"
	. "github.com/petergtz/pegomock"
)

func TestExecute(t *testing.T) {
	RegisterMockTestingT(t)
	client := mocks.NewMockClient()

	h := server.HelpExecutor{client}
	ctx := server.CommandContext{
		BaseRepo: models.Repo{},
		Pull:     models.PullRequest{},
		Log:      logging.NewSimpleLogger("", log.New(os.Stderr, "", log.LstdFlags), false, logging.Debug),
	}
	h.Execute(&ctx)
	client.VerifyWasCalledOnce().CreateComment(EqRepo(ctx.BaseRepo), EqPull(ctx.Pull), AnyString())
}

func EqRepo(value models.Repo) models.Repo {
	RegisterMatcher(&EqMatcher{Value: value})
	return models.Repo{}
}
func EqPull(value models.PullRequest) models.PullRequest {
	RegisterMatcher(&EqMatcher{Value: value})
	return models.PullRequest{}
}
