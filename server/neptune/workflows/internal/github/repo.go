package github

type Repo struct {
	// Owner is just the repo owner
	Owner string
	// Name is just the repo name, this will never have
	// /'s in it.
	Name string
	// URL is the ssh clone URL (ie. git@github.com:owner/repo.git)
	URL string

	Credentials AppCredentials
}

func (r Repo) GetFullName() string {
	return r.Owner + "/" + r.Name
}

type AppCredentials struct {
	InstallationToken int64
}
