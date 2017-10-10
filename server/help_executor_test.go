package server_test

import (
	"testing"

	"github.com/hootsuite/atlantis/server"
	. "github.com/hootsuite/atlantis/testing_util"
)

func TestExecute(t *testing.T) {
	h := server.HelpExecutor{}
	ctx := server.CommandContext{}
	r := h.Execute(&ctx)
	Equals(t, server.CommandResponse{}, r)
}
