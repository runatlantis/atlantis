package github

type CheckRunState string
type CheckRunConclusion string

const (
	CheckRunSuccess  CheckRunConclusion = "success"
	CheckRunFailure  CheckRunConclusion = "failure"
	CheckRunComplete CheckRunState      = "completed"
	CheckRunPending  CheckRunState      = "in_progress"
	CheckRunQueued   CheckRunState      = "queued"
)
