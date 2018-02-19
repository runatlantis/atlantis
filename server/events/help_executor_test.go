package events_test

import (
	"testing"

	"github.com/runatlantis/atlantis/server/events"
	. "github.com/runatlantis/atlantis/testing"
)

func TestExecute(t *testing.T) {
	h := events.HelpExecutor{}
	ctx := events.CommandContext{}
	r := h.Execute(&ctx)
	Equals(t, events.CommandResponse{}, r)
}
