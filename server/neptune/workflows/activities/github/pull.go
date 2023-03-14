package github

// Add more fields as necessary
type PullRequest struct {
	Number int
}

type State int64

const (
	Open State = iota
	Closed
	All
)

func (s State) String() string {
	switch s {
	case All:
		return "all"
	case Open:
		return "open"
	case Closed:
		return "closed"
	}

	// default to open
	return "open"
}
