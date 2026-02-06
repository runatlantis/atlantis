// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package models_test

import (
	"encoding/json"
	"testing"

	"github.com/runatlantis/atlantis/server/events/models"
	. "github.com/runatlantis/atlantis/testing"
)

func TestProjectOutput_Key(t *testing.T) {
	po := models.ProjectOutput{
		RepoFullName: "owner/repo",
		PullNum:      123,
		ProjectName:  "myproject",
		Workspace:    "default",
		Path:         "terraform/staging",
		CommandName:  "plan",
		RunTimestamp: 1707000000000,
	}

	key := po.Key()
	Equals(t, "owner/repo::123::terraform/staging::default::myproject::plan::1707000000000", key)
}

func TestProjectOutput_KeyWithEmptyProjectName(t *testing.T) {
	po := models.ProjectOutput{
		RepoFullName: "owner/repo",
		PullNum:      123,
		Workspace:    "default",
		Path:         "terraform/staging",
		CommandName:  "apply",
		RunTimestamp: 1707000000000,
	}

	key := po.Key()
	Equals(t, "owner/repo::123::terraform/staging::default::::apply::1707000000000", key)
}

func TestProjectOutput_ProjectKey(t *testing.T) {
	po := models.ProjectOutput{
		RepoFullName: "owner/repo",
		PullNum:      123,
		ProjectName:  "myproject",
		Workspace:    "default",
		Path:         "terraform/staging",
	}

	key := po.ProjectKey()
	Equals(t, "owner/repo::123::terraform/staging::default::myproject::", key)
}

func TestProjectOutput_PullKey(t *testing.T) {
	po := models.ProjectOutput{
		RepoFullName: "owner/repo",
		PullNum:      123,
	}

	key := po.PullKey()
	Equals(t, "owner/repo::123", key)
}

func TestProjectOutputStatus_String(t *testing.T) {
	tests := []struct {
		status   models.ProjectOutputStatus
		expected string
	}{
		{models.PendingOutputStatus, "pending"},
		{models.SuccessOutputStatus, "success"},
		{models.FailedOutputStatus, "failed"},
	}

	for _, tt := range tests {
		Equals(t, tt.expected, tt.status.String())
	}
}

func TestProjectOutputStatus_JSONMarshal(t *testing.T) {
	status := models.SuccessOutputStatus
	data, err := json.Marshal(status)
	Ok(t, err)
	Equals(t, `"success"`, string(data))
}

func TestProjectOutputStatus_JSONUnmarshal(t *testing.T) {
	var status models.ProjectOutputStatus
	err := json.Unmarshal([]byte(`"failed"`), &status)
	Ok(t, err)
	Equals(t, models.FailedOutputStatus, status)
}
