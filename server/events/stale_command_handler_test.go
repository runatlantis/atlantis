package events_test

import (
	. "github.com/petergtz/pegomock"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/models"
	. "github.com/runatlantis/atlantis/testing"
	"github.com/uber-go/tally"
	"testing"
	"time"
)

func TestStaleCommandHandler_CommandIsStale(t *testing.T) {
	olderTimestamp := time.Unix(123, 456)
	newerTimestamp := time.Unix(124, 457)
	testScope := tally.NewTestScope("test", nil)
	cases := []struct {
		Description      string
		PullStatus       models.PullStatus
		CommandTimestamp time.Time
		Expected         bool
	}{
		{
			Description: "simple stale command",
			PullStatus: models.PullStatus{
				UpdatedAt: newerTimestamp.Unix(),
			},
			CommandTimestamp: olderTimestamp,
			Expected:         true,
		},
		{
			Description: "simple not stale command",
			PullStatus: models.PullStatus{
				UpdatedAt: olderTimestamp.Unix(),
			},
			CommandTimestamp: newerTimestamp,
			Expected:         false,
		},
	}
	for _, c := range cases {
		t.Run(c.Description, func(t *testing.T) {
			RegisterMockTestingT(t)
			ctx := &events.CommandContext{
				TriggerTimestamp: c.CommandTimestamp,
				PullStatus:       &c.PullStatus,
			}
			staleCommandHandler := &events.StaleCommandHandler{
				StaleStatsScope: testScope,
			}
			Assert(t, c.Expected == staleCommandHandler.CommandIsStale(ctx),
				"CommandIsStale returned value should be %v", c.Expected)
		})
	}
	Assert(t, testScope.Snapshot().Counters()["test.dropped_commands+"].Value() == 1, "counted commands doesn't equal 1")
}

func TestStaleCommandHandler_CommandIsStale_NilPullModel(t *testing.T) {
	RegisterMockTestingT(t)
	testScope := tally.NewTestScope("test", nil)
	staleCommandHandler := &events.StaleCommandHandler{
		StaleStatsScope: testScope,
	}
	Assert(t, staleCommandHandler.CommandIsStale(&events.CommandContext{}) == false,
		"CommandIsStale returned value should be false")
	Assert(t, len(testScope.Snapshot().Counters()) == 0, "no counters should have started")
}
