package handlers

import (
	"context"
	"github.com/runatlantis/atlantis/server/http"
	"github.com/runatlantis/atlantis/server/neptune/gateway/event"
)

type PullRequestReviewEvent struct {
}

func (p PullRequestReviewEvent) Handle(ctx context.Context, e event.PullRequestReview, _ *http.BufferedRequest) error {
	//TODO implement me to run a modified approve command runner
	return nil
}

func NewPullRequestReviewEvent() *PullRequestReviewEvent {
	return &PullRequestReviewEvent{}
}
