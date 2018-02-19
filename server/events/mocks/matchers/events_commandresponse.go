package matchers

import (
	"reflect"

	"github.com/petergtz/pegomock"
	events "github.com/runatlantis/atlantis/server/events"
)

func AnyEventsCommandResponse() events.CommandResponse {
	pegomock.RegisterMatcher(pegomock.NewAnyMatcher(reflect.TypeOf((*(events.CommandResponse))(nil)).Elem()))
	var nullValue events.CommandResponse
	return nullValue
}

func EqEventsCommandResponse(value events.CommandResponse) events.CommandResponse {
	pegomock.RegisterMatcher(&pegomock.EqMatcher{Value: value})
	var nullValue events.CommandResponse
	return nullValue
}
