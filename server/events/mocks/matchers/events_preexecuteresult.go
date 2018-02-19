package matchers

import (
	"reflect"

	"github.com/petergtz/pegomock"
	events "github.com/runatlantis/atlantis/server/events"
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
