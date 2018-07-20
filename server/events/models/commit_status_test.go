package models_test

import (
	"testing"

	"github.com/runatlantis/atlantis/server/events/models"
	. "github.com/runatlantis/atlantis/testing"
)

func TestStatus_String(t *testing.T) {
	cases := map[models.CommitStatus]string{
		models.PendingCommitStatus: "pending",
		models.SuccessCommitStatus: "success",
		models.FailedCommitStatus:  "failed",
	}
	for k, v := range cases {
		Equals(t, v, k.String())
	}
}
