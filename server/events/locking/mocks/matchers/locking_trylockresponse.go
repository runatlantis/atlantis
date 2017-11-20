package matchers

import (
	"reflect"

	locking "github.com/hootsuite/atlantis/server/events/locking"
	"github.com/petergtz/pegomock"
)

func AnyLockingTryLockResponse() locking.TryLockResponse {
	pegomock.RegisterMatcher(pegomock.NewAnyMatcher(reflect.TypeOf((*(locking.TryLockResponse))(nil)).Elem()))
	var nullValue locking.TryLockResponse
	return nullValue
}

func EqLockingTryLockResponse(value locking.TryLockResponse) locking.TryLockResponse {
	pegomock.RegisterMatcher(&pegomock.EqMatcher{Value: value})
	var nullValue locking.TryLockResponse
	return nullValue
}
