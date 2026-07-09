// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	. "github.com/runatlantis/atlantis/testing"
)

func TestJobErrorMessageUsesJobMessageForWrappedStepOutput(t *testing.T) {
	inner := errors.New("running terraform in path: exit status 1")
	output := strings.Repeat("      + latest_version = (known after apply)\n", 5)
	stepErr := errorWithStepOutput(inner, []string{output}, true)
	wrapped := fmt.Errorf("plan failed: %w", stepErr)

	Assert(t, !strings.Contains(jobErrorMessage(wrapped), "+ latest_version"),
		"expected wrapped step output error not to replay terraform output in job message, got %q", jobErrorMessage(wrapped))
	Equals(t, "running terraform in path: exit status 1", jobErrorMessage(wrapped))
}

func TestErrWithStepOutputJobMessageIncludesOutputWhenNotStreamed(t *testing.T) {
	inner := errors.New("step failed")
	stepErr := errorWithStepOutput(inner, []string{"not streamed output"}, false).(errWithStepOutput)

	Equals(t, stepErr.Error(), stepErr.JobMessage())
	Assert(t, strings.Contains(stepErr.JobMessage(), "not streamed output"),
		"expected non-streamed output in job message, got %q", stepErr.JobMessage())
}

func TestErrWithStepOutputJobMessageOmitsOutputWhenStreamed(t *testing.T) {
	inner := errors.New("step failed")
	stepErr := errorWithStepOutput(inner, []string{"already streamed output"}, true).(errWithStepOutput)

	Equals(t, "step failed", stepErr.JobMessage())
	Assert(t, !strings.Contains(stepErr.JobMessage(), "already streamed output"),
		"expected streamed output to be omitted from job message, got %q", stepErr.JobMessage())
}
