package events

type CommandResponse struct {
	Error          error
	Failure        string
	ProjectResults []ProjectResult
}
