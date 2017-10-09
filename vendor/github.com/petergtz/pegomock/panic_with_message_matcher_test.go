package pegomock_test

import (
	"fmt"
	"reflect"

	"github.com/onsi/gomega/format"
	"github.com/onsi/gomega/types"
	"github.com/petergtz/pegomock/internal/verify"
)

type PanicWithMatcher struct {
	expectedWith interface{}
	actualWith   interface{}
}

func PanicWith(object interface{}) types.GomegaMatcher {
	verify.Argument(object != nil, "You must provide a non-nil object to PanicWith")
	return &PanicWithMatcher{expectedWith: object}
}

func (matcher *PanicWithMatcher) Match(actual interface{}) (success bool, err error) {
	if actual == nil {
		return false, fmt.Errorf("PanicWithMatcher expects a non-nil actual.")
	}

	actualType := reflect.TypeOf(actual)
	if actualType.Kind() != reflect.Func {
		return false, fmt.Errorf("PanicWithMatcher expects a function.  Got:\n%s", format.Object(actual, 1))
	}
	if !(actualType.NumIn() == 0 && actualType.NumOut() == 0) {
		return false, fmt.Errorf("PanicWithMatcher expects a function with no arguments and no return value.  Got:\n%s", format.Object(actual, 1))
	}

	success = false
	defer func() {
		if object := recover(); object == matcher.expectedWith {
			success = true
		} else {
			matcher.actualWith = object
		}
	}()

	reflect.ValueOf(actual).Call([]reflect.Value{})

	return
}

func (matcher *PanicWithMatcher) FailureMessage(actual interface{}) (message string) {
	if matcher.actualWith == "" {
		return format.Message(actual, "to panic")
	} else {
		// TODO: can we reuse format.Message somehow?
		return fmt.Sprintf("Expected\n\t<func ()>: %v\n\tpanicking with <%T>: %v\n\nto panic with\n\t<%T>: %v",
			actual, matcher.actualWith, matcher.actualWith, matcher.expectedWith, matcher.expectedWith)
	}
}

func (matcher *PanicWithMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	if matcher.actualWith == "" {
		return format.Message(actual, "not to panic")
	} else {
		return fmt.Sprintf("Expected\n\t<func ()>: %v\n\tpanicking with <%T>: %v\n\nnot to panic with\n\t<%T>: %v",
			actual, matcher.actualWith, matcher.actualWith, matcher.expectedWith, matcher.expectedWith)
	}

}
