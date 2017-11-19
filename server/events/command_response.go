package events

// CommandResponse is the result of running a Command.
type CommandResponse struct {
	Error          error
	Failure        string
	ProjectResults []ProjectResult
}
