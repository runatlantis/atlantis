// Copyright 2017 HootSuite Media Inc.
//
// Licensed under the Apache License, Version 2.0 (the License);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an AS IS BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// Modified hereafter by contributors to runatlantis/atlantis.

package events

import (
	"context"
	"fmt"

	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
)

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_commit_status_updater.go CommitStatusUpdater
type CommitStatusUpdater interface {
	// UpdateCombined updates the combined status of the head commit of pull.
	// A combined status represents all the projects modified in the pull.
	UpdateCombined(ctx context.Context, repo models.Repo, pull models.PullRequest, status models.CommitStatus, cmdName fmt.Stringer) error
	// UpdateCombinedCount updates the combined status to reflect the
	// numSuccess out of numTotal.
	UpdateCombinedCount(ctx context.Context, repo models.Repo, pull models.PullRequest, status models.CommitStatus, cmdName fmt.Stringer, numSuccess int, numTotal int) error
	// UpdateProject sets the commit status for the project represented by
	// ctx.
	UpdateProject(ctx context.Context, projectCtx command.ProjectContext, cmdName fmt.Stringer, status models.CommitStatus, url string) (string, error)
}
