package events

import (
	"strings"

	"github.com/runatlantis/atlantis/server/events/models"
)

// WorkflowHooksCommandResult is the result of executing pre workflow hooks for a
// repository.
type WorkflowHooksCommandResult struct {
	WorkflowHookResults []models.WorkflowHookResult
}

// HasErrors returns true if there were any errors during the execution,
// even if it was only in one project.
func (w WorkflowHooksCommandResult) HasErrors() bool {
	for _, r := range w.WorkflowHookResults {
		if !r.IsSuccessful() {
			return true
		}
	}
	return false
}

func (w WorkflowHooksCommandResult) Errors() string {
	errors := make([]string, 0)
	for _, r := range w.WorkflowHookResults {
		errors = append(errors, r.Error.Error())
	}

	return strings.Join(errors, "\n")
}
