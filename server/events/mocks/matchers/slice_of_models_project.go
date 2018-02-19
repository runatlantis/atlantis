package matchers

import (
	"reflect"

	"github.com/petergtz/pegomock"
	models "github.com/runatlantis/atlantis/server/events/models"
)

func AnySliceOfModelsProject() []models.Project {
	pegomock.RegisterMatcher(pegomock.NewAnyMatcher(reflect.TypeOf((*([]models.Project))(nil)).Elem()))
	var nullValue []models.Project
	return nullValue
}

func EqSliceOfModelsProject(value []models.Project) []models.Project {
	pegomock.RegisterMatcher(&pegomock.EqMatcher{Value: value})
	var nullValue []models.Project
	return nullValue
}
