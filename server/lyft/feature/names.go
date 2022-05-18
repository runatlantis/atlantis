package feature

type Name string

// list of feature names used in the code base. These must be kept in sync with any external config.
const LogPersistence Name = "log-persistence"
const PlatformMode Name = "platform-mode"
const GithubChecks Name = "github-checks"
