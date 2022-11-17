package temporal

import (
	"context"
	"go.temporal.io/sdk/activity"
	"time"
)

const HeartbeatTimeout = 5 * time.Second

// Adapted from dynajoe/temporal-terraform-demo:
// https://github.com/dynajoe/temporal-terraform-demo/blob/b468ac13cd9400ec0ffeb1b96eb8135e4b36d8ee/heartbeat/heartbeat.go#L10
func StartHeartbeat(ctx context.Context, frequency time.Duration) func() {
	ctx, cancel := context.WithCancel(ctx)
	go startHeartbeatTicks(ctx, frequency)
	return cancel
}

func startHeartbeatTicks(ctx context.Context, duration time.Duration) {
	ticker := time.NewTicker(duration)
	defer ticker.Stop()

	activity.RecordHeartbeat(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			activity.RecordHeartbeat(ctx)
		}
	}
}
