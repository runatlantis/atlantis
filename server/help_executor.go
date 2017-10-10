package server

type HelpExecutor struct{}

func (h *HelpExecutor) Execute(ctx *CommandContext) CommandResponse {
	// We don't actually need to do anything since the comment renderer
	// will see that it is a help command and render the help comment.
	return CommandResponse{}
}
