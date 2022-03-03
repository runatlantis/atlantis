package feature

type Name string

// list of feature names used in the code base. These must be kept in sync with any external config.
const LogStreaming Name = "log-streaming"
const ForceApply Name = "force-apply"
const LogPersistence Name = "log-persistence"
