package matchers

import (
	"reflect"

	events "github.com/hootsuite/atlantis/server/events"
	"github.com/petergtz/pegomock"
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
