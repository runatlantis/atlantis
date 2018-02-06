package matchers

import (
	"reflect"

	events "github.com/atlantisnorth/atlantis/server/events"
	"github.com/petergtz/pegomock"
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
