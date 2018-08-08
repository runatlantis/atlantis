package matchers

import (
	"reflect"

	"github.com/petergtz/pegomock"
	events "github.com/runatlantis/atlantis/server/events"
)

func AnyEventsProjectResult() events.ProjectResult {
	pegomock.RegisterMatcher(pegomock.NewAnyMatcher(reflect.TypeOf((*(events.ProjectResult))(nil)).Elem()))
	var nullValue events.ProjectResult
	return nullValue
}

func EqEventsProjectResult(value events.ProjectResult) events.ProjectResult {
	pegomock.RegisterMatcher(&pegomock.EqMatcher{Value: value})
	var nullValue events.ProjectResult
	return nullValue
}
