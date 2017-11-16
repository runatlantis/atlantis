package vcs

type Host int

const (
	Github Host = iota
	Gitlab
)

func (h Host) String() string {
	switch h {
	case Github:
		return "Github"
	case Gitlab:
		return "Gitlab"
	}
	return "<missing String() implementation>"
}

// CommitStatus is the result of executing an Atlantis command for the commit.
// In Github the options are: error, failure, pending, success.
// In Gitlab the options are: failed, canceled, pending, running, success.
// We only support Failed, Pending, Success.
type CommitStatus int

const (
	Pending CommitStatus = iota
	Success
	Failed
)

func (s CommitStatus) String() string {
	switch s {
	case Pending:
		return "pending"
	case Success:
		return "success"
	case Failed:
		return "failed"
	}
	return "failed"
}
