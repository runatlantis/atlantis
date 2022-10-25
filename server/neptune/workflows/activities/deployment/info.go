package deployment

const InfoSchemaVersion = 1.0

// These objects are persisted and should be kept as lightweight as possible
// Additionally, please ensure changes to this object going forward are backwards compatible
// Version is there to help in case of incompatible changes.
type Info struct {
	Version  int
	ID       string
	Revision string
	Repo     Repo
	Root     Root
}

type Repo struct {
	Owner string
	Name  string
}

func (r Repo) GetFullName() string {
	return r.Owner + "/" + r.Name
}

type Root struct {
	Name    string
	Trigger string
}
