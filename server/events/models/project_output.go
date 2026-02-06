// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package models

import (
	"encoding/json"
	"fmt"
	"time"
)

// ProjectOutputStatus represents the status of a command execution
type ProjectOutputStatus int

const (
	PendingOutputStatus ProjectOutputStatus = iota
	SuccessOutputStatus
	FailedOutputStatus
)

func (s ProjectOutputStatus) String() string {
	switch s {
	case PendingOutputStatus:
		return "pending"
	case SuccessOutputStatus:
		return "success"
	case FailedOutputStatus:
		return "failed"
	default:
		return "unknown"
	}
}

// MarshalJSON implements json.Marshaler for ProjectOutputStatus
func (s ProjectOutputStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

// UnmarshalJSON implements json.Unmarshaler for ProjectOutputStatus
func (s *ProjectOutputStatus) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	switch str {
	case "pending":
		*s = PendingOutputStatus
	case "success":
		*s = SuccessOutputStatus
	case "failed":
		*s = FailedOutputStatus
	default:
		return fmt.Errorf("unknown status: %s", str)
	}
	return nil
}

// ResourceStats holds the count of resources to be added, changed, or destroyed
type ResourceStats struct {
	Import  int `json:"import"`
	Add     int `json:"add"`
	Change  int `json:"change"`
	Destroy int `json:"destroy"`
}

// ProjectOutput represents persisted output from a plan/apply command
type ProjectOutput struct {
	// Identifiers
	RepoFullName string `json:"repo_full_name"`
	PullNum      int    `json:"pull_num"`
	ProjectName  string `json:"project_name"`
	Workspace    string `json:"workspace"`
	Path         string `json:"path"`

	// Command info
	CommandName  string `json:"command_name"` // plan, apply, policy_check
	JobID        string `json:"job_id"`
	RunTimestamp int64  `json:"run_timestamp"` // Unix millis UTC

	// Output
	Output string              `json:"output"`
	Status ProjectOutputStatus `json:"status"`
	Error  string              `json:"error,omitempty"`

	// Resource changes (for plan)
	ResourceStats ResourceStats `json:"resource_stats"`

	// Policy results
	PolicyPassed bool   `json:"policy_passed"`
	PolicyOutput string `json:"policy_output,omitempty"`

	// Metadata
	TriggeredBy string    `json:"triggered_by"`
	StartedAt   time.Time `json:"started_at"`
	CompletedAt time.Time `json:"completed_at"`

	// PR metadata (captured at persist time from ctx.Pull)
	PullURL   string `json:"pull_url,omitempty"`   // Full VCS URL (e.g., https://github.com/org/repo/pull/123)
	PullTitle string `json:"pull_title,omitempty"` // PR title from VCS
}

// Key returns a unique key for this project output
// Format: {repo}::{pull}::{path}::{workspace}::{project}::{command}::{timestamp}
func (p ProjectOutput) Key() string {
	return fmt.Sprintf("%s::%d::%s::%s::%s::%s::%d",
		p.RepoFullName,
		p.PullNum,
		p.Path,
		p.Workspace,
		p.ProjectName,
		p.CommandName,
		p.RunTimestamp,
	)
}

// ProjectKey returns a key prefix for all outputs of this project
// Format: {repo}::{pull}::{path}::{workspace}::{project}::
func (p ProjectOutput) ProjectKey() string {
	return fmt.Sprintf("%s::%d::%s::%s::%s::",
		p.RepoFullName,
		p.PullNum,
		p.Path,
		p.Workspace,
		p.ProjectName,
	)
}

// PullKey returns a key identifying the pull request
// Format: {repo}::{pull}
func (p ProjectOutput) PullKey() string {
	return fmt.Sprintf("%s::%d", p.RepoFullName, p.PullNum)
}
