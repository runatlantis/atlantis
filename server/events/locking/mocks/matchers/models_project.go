package matchers

import (
	"reflect"

	models "github.com/hootsuite/atlantis/server/events/models"
	"github.com/petergtz/pegomock"
)

func AnyModelsProject() models.Project {
	pegomock.RegisterMatcher(pegomock.NewAnyMatcher(reflect.TypeOf((*(models.Project))(nil)).Elem()))
	var nullValue models.Project
	return nullValue
}

func EqModelsProject(value models.Project) models.Project {
	pegomock.RegisterMatcher(&pegomock.EqMatcher{Value: value})
	var nullValue models.Project
	return nullValue
}
