package matchers

import (
	"reflect"

	models "github.com/hootsuite/atlantis/server/events/models"
	"github.com/petergtz/pegomock"
)

func AnyModelsPullRequest() models.PullRequest {
	pegomock.RegisterMatcher(pegomock.NewAnyMatcher(reflect.TypeOf((*(models.PullRequest))(nil)).Elem()))
	var nullValue models.PullRequest
	return nullValue
}

func EqModelsPullRequest(value models.PullRequest) models.PullRequest {
	pegomock.RegisterMatcher(&pegomock.EqMatcher{Value: value})
	var nullValue models.PullRequest
	return nullValue
}
