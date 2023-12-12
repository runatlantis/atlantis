package valid

// AutoDiscoverMode enum
type AutoDiscoverMode string

const (
	AutoDiscoverEnabledMode  AutoDiscoverMode = "enabled"
	AutoDiscoverDisabledMode AutoDiscoverMode = "disabled"
	AutoDiscoverAutoMode     AutoDiscoverMode = "auto"
)

type AutoDiscover struct {
	Mode AutoDiscoverMode
}
