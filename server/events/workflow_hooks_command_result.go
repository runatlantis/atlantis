package events

import "github.com/runatlantis/atlantis/server/events/models"

// WorkflowHookCommandResult is the result of executing pre workflow hooks for a
// repository.
type WorkflowHookCommandResult struct {
	WorkflowHookResults []models.WorkflowHookResult
}

// HasErrors returns true if there were any errors during the execution,
// even if it was only in one project.
func (w WorkflowHookCommandResult) HasErrors() bool {
	for _, r := range w.WorkflowHookResults {
		if !r.IsSuccessful() {
			return true
		}
	}
	return false
}
