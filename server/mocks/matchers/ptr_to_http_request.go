package matchers

import (
	http "net/http"
	"reflect"

	"github.com/petergtz/pegomock"
)

func AnyPtrToHttpRequest() *http.Request {
	pegomock.RegisterMatcher(pegomock.NewAnyMatcher(reflect.TypeOf((*(*http.Request))(nil)).Elem()))
	var nullValue *http.Request
	return nullValue
}

func EqPtrToHttpRequest(value *http.Request) *http.Request {
	pegomock.RegisterMatcher(&pegomock.EqMatcher{Value: value})
	var nullValue *http.Request
	return nullValue
}
