package request

// Types defined here should not be used internally, as our goal should be to eventually swap these out for something less brittle than json translation
type Repo struct {
	// FullName is the owner and repo name separated
	// by a "/"
	FullName string
	// Owner is just the repo owner
	Owner string
	// Name is just the repo name, this will never have
	// /'s in it.
	Name string
	// URL is the ssh clone URL (ie. git@github.com:owner/repo.git)
	URL string
	// Flag to determine if open PRs for a root need to be rebased
	RebaseEnabled bool
	// Repo's default branch
	DefaultBranch string

	Credentials AppCredentials
}

type AppCredentials struct {
	InstallationToken int64
}

type User struct {
	Name string
}
