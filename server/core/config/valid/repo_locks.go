package valid

// RepoLocksMode enum
type RepoLocksMode string

var DefaultRepoLocksMode = RepoLocksOnPlanMode
var DefaultRepoLocks = RepoLocks{
	Mode: DefaultRepoLocksMode,
}

const (
	RepoLocksDisabledMode RepoLocksMode = "disabled"
	RepoLocksOnPlanMode   RepoLocksMode = "on_plan"
	RepoLocksOnApplyMode  RepoLocksMode = "on_apply"
)

type RepoLocks struct {
	Mode RepoLocksMode
}
