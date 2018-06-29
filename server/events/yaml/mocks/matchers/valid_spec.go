package matchers

import (
	"reflect"

	"github.com/petergtz/pegomock"
	valid "github.com/runatlantis/atlantis/server/events/yaml/valid"
)

func AnyValidConfig() valid.Config {
	pegomock.RegisterMatcher(pegomock.NewAnyMatcher(reflect.TypeOf((*(valid.Config))(nil)).Elem()))
	var nullValue valid.Config
	return nullValue
}

func EqValidConfig(value valid.Config) valid.Config {
	pegomock.RegisterMatcher(&pegomock.EqMatcher{Value: value})
	var nullValue valid.Config
	return nullValue
}
