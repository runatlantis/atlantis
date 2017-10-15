package events

type CommandName int

const (
	Apply CommandName = iota
	Plan
	Help
	// Adding more? Don't forget to update String() below
)

func (c CommandName) String() string {
	switch c {
	case Apply:
		return "apply"
	case Plan:
		return "plan"
	case Help:
		return "help"
	}
	return ""
}
