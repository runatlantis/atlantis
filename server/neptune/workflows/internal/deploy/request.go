package deploy

type Request struct {
	Repo Repo
	Root Root
}

// Repo Names and Root Names are assumed to be static throughout the lifetime
// of a workflow since that's what our ID is based on
//
// We need these values at minimum during startup atm.
type Repo struct {
	FullName string
}

type Root struct {
	Name string
}
