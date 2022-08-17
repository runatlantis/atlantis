package github

type CheckRunState string
type CheckRunConclusion string

const (
	Success          CheckRunConclusion = "success"
	Failure          CheckRunConclusion = "failure"
	CheckRunComplete CheckRunState      = "completed"
	CheckRunPending  CheckRunState      = "in_progress"
	CheckRunQueued   CheckRunState      = "queued"
)
