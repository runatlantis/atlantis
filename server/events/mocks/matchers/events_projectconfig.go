package matchers

import (
	"reflect"

	"github.com/petergtz/pegomock"
	events "github.com/runatlantis/atlantis/server/events"
)

func AnyEventsProjectConfig() events.ProjectConfig {
	pegomock.RegisterMatcher(pegomock.NewAnyMatcher(reflect.TypeOf((*(events.ProjectConfig))(nil)).Elem()))
	var nullValue events.ProjectConfig
	return nullValue
}

func EqEventsProjectConfig(value events.ProjectConfig) events.ProjectConfig {
	pegomock.RegisterMatcher(&pegomock.EqMatcher{Value: value})
	var nullValue events.ProjectConfig
	return nullValue
}
