package matchers

import (
	"reflect"

	"github.com/petergtz/pegomock"
	events "github.com/runatlantis/atlantis/server/events"
)

func AnyEventsCommandParseResult() events.CommandParseResult {
	pegomock.RegisterMatcher(pegomock.NewAnyMatcher(reflect.TypeOf((*(events.CommandParseResult))(nil)).Elem()))
	var nullValue events.CommandParseResult
	return nullValue
}

func EqEventsCommandParseResult(value events.CommandParseResult) events.CommandParseResult {
	pegomock.RegisterMatcher(&pegomock.EqMatcher{Value: value})
	var nullValue events.CommandParseResult
	return nullValue
}
