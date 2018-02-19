package matchers

import (
	"reflect"

	"github.com/petergtz/pegomock"
	events "github.com/runatlantis/atlantis/server/events"
)

func AnyPtrToEventsCommand() *events.Command {
	pegomock.RegisterMatcher(pegomock.NewAnyMatcher(reflect.TypeOf((*(*events.Command))(nil)).Elem()))
	var nullValue *events.Command
	return nullValue
}

func EqPtrToEventsCommand(value *events.Command) *events.Command {
	pegomock.RegisterMatcher(&pegomock.EqMatcher{Value: value})
	var nullValue *events.Command
	return nullValue
}
