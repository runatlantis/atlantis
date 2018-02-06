package matchers

import (
	"reflect"

	events "github.com/atlantisnorth/atlantis/server/events"
	"github.com/petergtz/pegomock"
)

func AnyEventsPreExecuteResult() events.PreExecuteResult {
	pegomock.RegisterMatcher(pegomock.NewAnyMatcher(reflect.TypeOf((*(events.PreExecuteResult))(nil)).Elem()))
	var nullValue events.PreExecuteResult
	return nullValue
}

func EqEventsPreExecuteResult(value events.PreExecuteResult) events.PreExecuteResult {
	pegomock.RegisterMatcher(&pegomock.EqMatcher{Value: value})
	var nullValue events.PreExecuteResult
	return nullValue
}
