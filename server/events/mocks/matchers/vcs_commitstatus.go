package matchers

import (
	"reflect"

	"github.com/petergtz/pegomock"
	models "github.com/runatlantis/atlantis/server/events/models"
)

func AnyVcsCommitStatus() models.CommitStatus {
	pegomock.RegisterMatcher(pegomock.NewAnyMatcher(reflect.TypeOf((*(models.CommitStatus))(nil)).Elem()))
	var nullValue models.CommitStatus
	return nullValue
}

func EqVcsCommitStatus(value models.CommitStatus) models.CommitStatus {
	pegomock.RegisterMatcher(&pegomock.EqMatcher{Value: value})
	var nullValue models.CommitStatus
	return nullValue
}
