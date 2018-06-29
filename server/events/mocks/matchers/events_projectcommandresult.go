package matchers

import (
	"reflect"

	"github.com/petergtz/pegomock"
	events "github.com/runatlantis/atlantis/server/events"
)

func AnyEventsProjectCommandResult() events.ProjectCommandResult {
	pegomock.RegisterMatcher(pegomock.NewAnyMatcher(reflect.TypeOf((*(events.ProjectCommandResult))(nil)).Elem()))
	var nullValue events.ProjectCommandResult
	return nullValue
}

func EqEventsProjectCommandResult(value events.ProjectCommandResult) events.ProjectCommandResult {
	pegomock.RegisterMatcher(&pegomock.EqMatcher{Value: value})
	var nullValue events.ProjectCommandResult
	return nullValue
}
