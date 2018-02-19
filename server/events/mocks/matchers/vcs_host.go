package matchers

import (
	"reflect"

	"github.com/petergtz/pegomock"
	vcs "github.com/runatlantis/atlantis/server/events/vcs"
)

func AnyVcsHost() vcs.Host {
	pegomock.RegisterMatcher(pegomock.NewAnyMatcher(reflect.TypeOf((*(vcs.Host))(nil)).Elem()))
	var nullValue vcs.Host
	return nullValue
}

func EqVcsHost(value vcs.Host) vcs.Host {
	pegomock.RegisterMatcher(&pegomock.EqMatcher{Value: value})
	var nullValue vcs.Host
	return nullValue
}
