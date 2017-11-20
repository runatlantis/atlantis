package matchers

import (
	"reflect"

	slack "github.com/nlopes/slack"
	"github.com/petergtz/pegomock"
)

func AnyPtrToSlackAuthTestResponse() *slack.AuthTestResponse {
	pegomock.RegisterMatcher(pegomock.NewAnyMatcher(reflect.TypeOf((*(*slack.AuthTestResponse))(nil)).Elem()))
	var nullValue *slack.AuthTestResponse
	return nullValue
}

func EqPtrToSlackAuthTestResponse(value *slack.AuthTestResponse) *slack.AuthTestResponse {
	pegomock.RegisterMatcher(&pegomock.EqMatcher{Value: value})
	var nullValue *slack.AuthTestResponse
	return nullValue
}
