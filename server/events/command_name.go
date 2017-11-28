package events

// CommandName is the type of command.
type CommandName int

const (
	Apply CommandName = iota
	Plan
	Help
	// Adding more? Don't forget to update String() below
)

// String returns the string representation of c.
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
