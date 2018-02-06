package matchers

import (
	"reflect"

	logging "github.com/atlantisnorth/atlantis/server/logging"
	"github.com/petergtz/pegomock"
)

func AnyLoggingLogLevel() logging.LogLevel {
	pegomock.RegisterMatcher(pegomock.NewAnyMatcher(reflect.TypeOf((*(logging.LogLevel))(nil)).Elem()))
	var nullValue logging.LogLevel
	return nullValue
}

func EqLoggingLogLevel(value logging.LogLevel) logging.LogLevel {
	pegomock.RegisterMatcher(&pegomock.EqMatcher{Value: value})
	var nullValue logging.LogLevel
	return nullValue
}
