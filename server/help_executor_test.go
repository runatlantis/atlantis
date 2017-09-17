package server_test

import (
	"testing"
	"github.com/golang/mock/gomock"
	"github.com/hootsuite/atlantis/github/mocks"
	"github.com/hootsuite/atlantis/server"
	"github.com/hootsuite/atlantis/models"
	"github.com/hootsuite/atlantis/logging"
	"os"
	"log"
)

func TestExecute(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mock := mocks.NewMockClient(ctrl)

	h := server.HelpExecutor{mock}
	ctx := server.CommandContext{
		BaseRepo: models.Repo{},
		Pull: models.PullRequest{},
		Log: logging.NewSimpleLogger("", log.New(os.Stderr, "", log.LstdFlags), false, logging.Debug),
	}
	mock.EXPECT().CreateComment(ctx.BaseRepo, ctx.Pull, gomock.Any())
	h.Execute(&ctx)
}