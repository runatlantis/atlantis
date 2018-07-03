package vcs_test

import (
	"testing"

	"github.com/runatlantis/atlantis/server/events/vcs"
	. "github.com/runatlantis/atlantis/testing"
)

func TestStatus_String(t *testing.T) {
	cases := map[vcs.CommitStatus]string{
		vcs.Pending: "pending",
		vcs.Success: "success",
		vcs.Failed:  "failed",
	}
	for k, v := range cases {
		Equals(t, v, k.String())
	}
}
