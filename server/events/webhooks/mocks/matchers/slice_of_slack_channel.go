package matchers

import (
	"reflect"

	slack "github.com/nlopes/slack"
	"github.com/petergtz/pegomock"
)

func AnySliceOfSlackChannel() []slack.Channel {
	pegomock.RegisterMatcher(pegomock.NewAnyMatcher(reflect.TypeOf((*([]slack.Channel))(nil)).Elem()))
	var nullValue []slack.Channel
	return nullValue
}

func EqSliceOfSlackChannel(value []slack.Channel) []slack.Channel {
	pegomock.RegisterMatcher(&pegomock.EqMatcher{Value: value})
	var nullValue []slack.Channel
	return nullValue
}
