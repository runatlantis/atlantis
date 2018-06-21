package matchers

import (
	"reflect"

	"github.com/petergtz/pegomock"
	valid "github.com/runatlantis/atlantis/server/events/yaml/valid"
)

func AnyValidSpec() valid.Spec {
	pegomock.RegisterMatcher(pegomock.NewAnyMatcher(reflect.TypeOf((*(valid.Spec))(nil)).Elem()))
	var nullValue valid.Spec
	return nullValue
}

func EqValidSpec(value valid.Spec) valid.Spec {
	pegomock.RegisterMatcher(&pegomock.EqMatcher{Value: value})
	var nullValue valid.Spec
	return nullValue
}
