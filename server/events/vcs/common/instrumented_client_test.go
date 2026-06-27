// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package common_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/events/vcs/common"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
	tally "github.com/uber-go/tally/v4"
)

type updateStatusErrorClient struct {
	*vcs.NotConfiguredVCSClient
	err error
}

func (c *updateStatusErrorClient) UpdateStatus(_ logging.SimpleLogging, _ models.Repo, _ models.PullRequest, _ models.CommitStatus, _ string, _ string, _ string) error {
	return c.err
}

func TestInstrumentedClient_UpdateStatusLogsErrorContext(t *testing.T) {
	tests := []struct {
		name        string
		src         string
		url         string
		wantLog     string
		notWantLogs []string
	}{
		{
			name:    "source without URL",
			src:     "atlantis/plan",
			wantLog: "Unable to update status (from 'atlantis/plan'), error: status denied",
			notWantLogs: []string{
				"at url:",
			},
		},
		{
			name:    "URL present",
			src:     "atlantis/plan",
			url:     "https://example.com/status",
			wantLog: "Unable to update status at url: https://example.com/status, error: status denied",
		},
		{
			name:    "no source or URL",
			wantLog: "Unable to update status, error: status denied",
			notWantLogs: []string{
				"at url:",
				"(from",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := logging.NewNoopLogger(t).WithHistory()
			client := &common.InstrumentedClient{
				Client: &updateStatusErrorClient{
					NotConfiguredVCSClient: &vcs.NotConfiguredVCSClient{},
					err:                    errors.New("status denied"),
				},
				StatsScope: tally.NoopScope,
			}

			err := client.UpdateStatus(
				logger,
				models.Repo{FullName: "owner/repo"},
				models.PullRequest{Num: 1},
				models.FailedCommitStatus,
				tt.src,
				"description",
				tt.url,
			)

			ErrContains(t, "status denied", err)
			history := logger.GetHistory()
			Assert(t, strings.Contains(history, tt.wantLog), "expected log history %q to contain %q", history, tt.wantLog)
			for _, notWantLog := range tt.notWantLogs {
				Assert(t, !strings.Contains(history, notWantLog), "expected log history %q not to contain %q", history, notWantLog)
			}
		})
	}
}
