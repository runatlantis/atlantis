package matchers

import (
	"reflect"

	vcs "github.com/hootsuite/atlantis/server/events/vcs"
	"github.com/petergtz/pegomock"
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
