package command

// Result is the result of running a Command.
type Result struct {
	Error          error
	Failure        string
	ProjectResults []ProjectResult
	// PlansDeleted is true if all plans created during this command were
	// deleted. This happens if automerging is enabled and one project has an
	// error since automerging requires all plans to succeed.
	PlansDeleted bool
}

// HasErrors returns true if there were any errors during the execution,
// even if it was only in one project.
func (c Result) HasErrors() bool {
	if c.Error != nil || c.Failure != "" {
		return true
	}
	for _, r := range c.ProjectResults {
		if !r.IsSuccessful() {
			return true
		}
	}
	return false
}
