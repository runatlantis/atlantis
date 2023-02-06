package version

// This version removes an activity that computes env vars from commands and instead opts
// for lazy loading within each of the following steps.
const LazyLoadEnvVars = "lazy-load-env-vars"
