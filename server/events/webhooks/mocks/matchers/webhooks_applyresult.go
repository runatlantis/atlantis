package matchers

import (
	"reflect"

	webhooks "github.com/atlantisnorth/atlantis/server/events/webhooks"
	"github.com/petergtz/pegomock"
)

func AnyWebhooksApplyResult() webhooks.ApplyResult {
	pegomock.RegisterMatcher(pegomock.NewAnyMatcher(reflect.TypeOf((*(webhooks.ApplyResult))(nil)).Elem()))
	var nullValue webhooks.ApplyResult
	return nullValue
}

func EqWebhooksApplyResult(value webhooks.ApplyResult) webhooks.ApplyResult {
	pegomock.RegisterMatcher(&pegomock.EqMatcher{Value: value})
	var nullValue webhooks.ApplyResult
	return nullValue
}
