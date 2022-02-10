package events

import (
	"github.com/uber-go/tally"
)

type StaleCommandHandler struct {
	StaleStatsScope tally.Scope
}

func (s *StaleCommandHandler) CommandIsStale(ctx *CommandContext) bool {
	status := ctx.PullStatus
	if status != nil && status.UpdatedAt > ctx.TriggerTimestamp.Unix() {
		s.StaleStatsScope.Counter("dropped_commands").Inc(1)
		return true
	}
	return false
}
