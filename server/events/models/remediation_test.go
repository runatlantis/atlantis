// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package models_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/runatlantis/atlantis/server/events/models"
	. "github.com/runatlantis/atlantis/testing"
)

func TestRemediationResultCompletedAtJSON(t *testing.T) {
	result := models.NewRemediationResult("id", "owner/repo", "main", models.RemediationPlanOnly)

	body, err := json.Marshal(result)
	Ok(t, err)
	Assert(t, !strings.Contains(string(body), "completed_at"), "expected incomplete remediation to omit completed_at: %s", body)

	result.Complete()
	body, err = json.Marshal(result)
	Ok(t, err)
	Assert(t, strings.Contains(string(body), "completed_at"), "expected completed remediation to include completed_at: %s", body)
}
