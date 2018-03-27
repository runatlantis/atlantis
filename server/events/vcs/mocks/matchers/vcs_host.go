package matchers

import (
	"reflect"

	"github.com/petergtz/pegomock"
	models "github.com/runatlantis/atlantis/server/events/models"
)

func AnyVcsHost() models.Host {
	pegomock.RegisterMatcher(pegomock.NewAnyMatcher(reflect.TypeOf((*(models.Host))(nil)).Elem()))
	var nullValue models.Host
	return nullValue
}

func EqVcsHost(value models.Host) models.Host {
	pegomock.RegisterMatcher(&pegomock.EqMatcher{Value: value})
	var nullValue models.Host
	return nullValue
}
