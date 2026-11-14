// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events_test

import (
	"errors"
	"testing"

	. "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/mocks"
	"github.com/runatlantis/atlantis/server/events/models"
	vcsmocks "github.com/runatlantis/atlantis/server/events/vcs/mocks"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

func TestUnlockCommandRunner_Project(t *testing.T) {
	for _, tc := range []struct {
		name         string
		labels       []string
		labelErr     error
		deleteErr    error
		silence      bool
		shouldDelete bool
		message      string
	}{
		{
			name: "success", shouldDelete: true,
			message: `Atlantis locks for project "production" in this PR have been unlocked and its plans discarded. Run plan again before applying.`,
		},
		{
			name: "unknown project is not silenced", shouldDelete: true, silence: true,
			deleteErr: errors.New(`no Atlantis lock found for project "production" in this PR`),
			message:   `Could not unlock project "production": no Atlantis lock found for project "production" in this PR`,
		},
		{
			name: "cleanup failure", shouldDelete: true,
			deleteErr: errors.New("deleting plan"),
			message:   `Could not unlock project "production": deleting plan`,
		},
		{
			name: "disabled by label", labels: []string{"do-not-unlock"}, silence: true,
			message: "Not allowed to unlock PR with do-not-unlock label",
		},
		{
			name: "label lookup failure", labelErr: errors.New("labels unavailable"), silence: true,
			message: "Failed to retrieve PR labels... Not unlocking",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			RegisterMockTestingT(t)
			vcsClient := vcsmocks.NewMockClient()
			deleter := mocks.NewMockDeleteLockCommand()
			ctx := &command.Context{
				Log:  logging.NewNoopLogger(t),
				Pull: models.PullRequest{Num: 42, BaseRepo: models.Repo{FullName: "owner/repo"}},
			}
			When(vcsClient.GetPullLabels(Any[logging.SimpleLogging](), Eq(ctx.Pull.BaseRepo), Eq(ctx.Pull))).ThenReturn(tc.labels, tc.labelErr)
			count := 1
			if tc.deleteErr != nil {
				count = 0
			}
			When(deleter.DeleteLocksByProject(Any[logging.SimpleLogging](), Eq("owner/repo"), Eq(42), Eq("production"))).ThenReturn(count, tc.deleteErr)
			runner := events.NewUnlockCommandRunner(deleter, vcsClient, tc.silence, "do-not-unlock")
			parsed := commentParser.Parse("atlantis unlock -p production", models.Github)
			Assert(t, parsed.Command != nil, "expected parsed project unlock")
			runner.Run(ctx, parsed.Command)

			calls := 0
			if tc.shouldDelete {
				calls = 1
			}
			deleter.VerifyWasCalled(Times(calls)).DeleteLocksByProject(Any[logging.SimpleLogging](), Eq("owner/repo"), Eq(42), Eq("production"))
			deleter.VerifyWasCalled(Never()).DeleteLocksByPull(Any[logging.SimpleLogging](), Any[string](), Any[int]())
			vcsClient.VerifyWasCalledOnce().CreateComment(Any[logging.SimpleLogging](), Eq(ctx.Pull.BaseRepo), Eq(42), Eq(tc.message), Eq("unlock"))
		})
	}
}
