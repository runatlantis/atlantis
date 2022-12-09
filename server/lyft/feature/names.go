package feature

type Name string

// list of feature names used in the code base. These must be kept in sync with any external config.
const (
	PlatformMode Name = "platform-mode"
	PolicyV2     Name = "policy-v2"
)
