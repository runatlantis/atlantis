package matchers

import (
	"reflect"

	models "github.com/hootsuite/atlantis/server/events/models"
	"github.com/petergtz/pegomock"
)

func AnyModelsRepo() models.Repo {
	pegomock.RegisterMatcher(pegomock.NewAnyMatcher(reflect.TypeOf((*(models.Repo))(nil)).Elem()))
	var nullValue models.Repo
	return nullValue
}

func EqModelsRepo(value models.Repo) models.Repo {
	pegomock.RegisterMatcher(&pegomock.EqMatcher{Value: value})
	var nullValue models.Repo
	return nullValue
}
