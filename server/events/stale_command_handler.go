package events

import (
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/uber-go/tally/v4"
)

type StaleCommandHandler struct {
	StaleStatsScope tally.Scope
}

func (s *StaleCommandHandler) CommandIsStale(ctx *command.Context) bool {
	status := ctx.PullStatus
	if status != nil && status.UpdatedAt > ctx.TriggerTimestamp.Unix() {
		s.StaleStatsScope.Counter("dropped_commands").Inc(1)
		return true
	}
	return false
}
