package matchers

import (
	"reflect"

	slack "github.com/nlopes/slack"
	"github.com/petergtz/pegomock"
)

func AnySlackPostMessageParameters() slack.PostMessageParameters {
	pegomock.RegisterMatcher(pegomock.NewAnyMatcher(reflect.TypeOf((*(slack.PostMessageParameters))(nil)).Elem()))
	var nullValue slack.PostMessageParameters
	return nullValue
}

func EqSlackPostMessageParameters(value slack.PostMessageParameters) slack.PostMessageParameters {
	pegomock.RegisterMatcher(&pegomock.EqMatcher{Value: value})
	var nullValue slack.PostMessageParameters
	return nullValue
}
