package wrappers

import (
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/lyft/aws/sns"
	"github.com/runatlantis/atlantis/server/lyft/decorators"
	"github.com/runatlantis/atlantis/server/sync"
)

type projectCommand struct {
	events.ProjectCommandRunner
}

func WrapProjectRunner(projectRunner events.ProjectCommandRunner) *projectCommand { //nolint:golint // avoiding refactor while adding linter action
	return &projectCommand{
		projectRunner,
	}
}

// WithSync add project level locking to projects
func (d *projectCommand) WithSync(
	projectLocker events.ProjectLocker,
	projectLockURL events.LockURLGenerator,
) *projectCommand {
	d.ProjectCommandRunner = &sync.ProjectSyncer{
		ProjectCommandRunner: d.ProjectCommandRunner,
		Locker:               projectLocker,
		LockURLGenerator:     projectLockURL,
	}

	return d
}

// WithJobs adds streaming capabilities to terraform output. With it end user
// can see their terraform command's execution in real time.
func (d *projectCommand) WithJobs(
	projectStatusUpdater command.StatusUpdater,
) *projectCommand {
	d.ProjectCommandRunner = &events.ProjectOutputWrapper{
		ProjectCommandRunner: d.ProjectCommandRunner,
		ProjectStatusUpdater: projectStatusUpdater,
	}
	return d
}

func (d *projectCommand) WithAuditing(
	snsWriter sns.Writer,
) *projectCommand {
	d.ProjectCommandRunner = &decorators.AuditProjectCommandWrapper{
		ProjectCommandRunner: d.ProjectCommandRunner,
		SnsWriter:            snsWriter,
	}
	return d
}

func (d *projectCommand) WithInstrumentation() *projectCommand {
	d.ProjectCommandRunner = &events.InstrumentedProjectCommandRunner{
		ProjectCommandRunner: d.ProjectCommandRunner,
	}
	return d
}
