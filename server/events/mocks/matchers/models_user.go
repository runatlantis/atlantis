package matchers

import (
	"reflect"

	models "github.com/hootsuite/atlantis/server/events/models"
	"github.com/petergtz/pegomock"
)

func AnyModelsUser() models.User {
	pegomock.RegisterMatcher(pegomock.NewAnyMatcher(reflect.TypeOf((*(models.User))(nil)).Elem()))
	var nullValue models.User
	return nullValue
}

func EqModelsUser(value models.User) models.User {
	pegomock.RegisterMatcher(&pegomock.EqMatcher{Value: value})
	var nullValue models.User
	return nullValue
}
