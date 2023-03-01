package models_test

import (
	"testing"

	"github.com/runatlantis/atlantis/server/events/models"
	. "github.com/runatlantis/atlantis/testing"
)

func TestStatus_String(t *testing.T) {
	cases := map[models.VCSStatus]string{
		models.PendingVCSStatus: "pending",
		models.QueuedVCSStatus:  "queued",
		models.SuccessVCSStatus: "success",
		models.FailedVCSStatus:  "failed",
	}
	for k, v := range cases {
		Equals(t, v, k.String())
	}
}
