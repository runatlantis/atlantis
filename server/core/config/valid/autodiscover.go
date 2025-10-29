package valid

import "github.com/bmatcuk/doublestar/v4"

// AutoDiscoverMode enum
type AutoDiscoverMode string

const (
	AutoDiscoverEnabledMode  AutoDiscoverMode = "enabled"
	AutoDiscoverDisabledMode AutoDiscoverMode = "disabled"
	AutoDiscoverAutoMode     AutoDiscoverMode = "auto"
)

type AutoDiscover struct {
	Mode        AutoDiscoverMode
	IgnorePaths []string
}

func (a AutoDiscover) IsPathIgnored(path string) bool {
	if a.IgnorePaths == nil {
		return false
	}
	for i := 0; i < len(a.IgnorePaths); i++ {
		// Per documentation https://pkg.go.dev/github.com/bmatcuk/doublestar, if you run ValidatePattern()
		// against a pattern, which we do, you can run MatchUnvalidated for a slight performance gain,
		// and also no need to explicitly check for an error
		if doublestar.MatchUnvalidated(a.IgnorePaths[i], path) {
			return true
		}
	}
	return false
}
