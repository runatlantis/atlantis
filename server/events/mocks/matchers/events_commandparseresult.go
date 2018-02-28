package matchers

import (
	"reflect"

	"github.com/petergtz/pegomock"
	events "github.com/runatlantis/atlantis/server/events"
)

func AnyEventsCommandParseResult() events.CommentParseResult {
	pegomock.RegisterMatcher(pegomock.NewAnyMatcher(reflect.TypeOf((*(events.CommentParseResult))(nil)).Elem()))
	var nullValue events.CommentParseResult
	return nullValue
}

func EqEventsCommandParseResult(value events.CommentParseResult) events.CommentParseResult {
	pegomock.RegisterMatcher(&pegomock.EqMatcher{Value: value})
	var nullValue events.CommentParseResult
	return nullValue
}
