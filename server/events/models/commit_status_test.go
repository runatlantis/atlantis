package models_test

import (
	"testing"

	"github.com/runatlantis/atlantis/server/events/models"
	. "github.com/runatlantis/atlantis/testing"
)

func TestStatus_String(t *testing.T) {
	cases := map[models.CommitStatus]string{
		models.Pending: "pending",
		models.Success: "success",
		models.Failed:  "failed",
	}
	for k, v := range cases {
		Equals(t, v, k.String())
	}
}
