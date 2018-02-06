package matchers

import (
	"reflect"

	events "github.com/atlantisnorth/atlantis/server/events"
	"github.com/petergtz/pegomock"
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
