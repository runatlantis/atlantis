package pegomock_test

import (
	"fmt"
	"reflect"

	"github.com/onsi/gomega/format"
	"github.com/onsi/gomega/types"
	"github.com/petergtz/pegomock/internal/verify"
)

type PanicWithMessageToMatcher struct {
	expectedWith types.GomegaMatcher
	actualWith   interface{}
}

func PanicWithMessageTo(object types.GomegaMatcher) types.GomegaMatcher {
	verify.Argument(object != nil, "You must provide a non-nil object to PanicWith")
	return &PanicWithMessageToMatcher{expectedWith: object}
}

func (matcher *PanicWithMessageToMatcher) Match(actual interface{}) (success bool, err error) {
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
		if object := recover(); object != nil {
			var e error
			success, e = matcher.expectedWith.Match(object)
			if e != nil {
				// TODO is there something needed here?
			}
			if !success {
				matcher.actualWith = object
			}
		} else {
			matcher.actualWith = object
		}
	}()

	reflect.ValueOf(actual).Call([]reflect.Value{})

	return
}

func (matcher *PanicWithMessageToMatcher) FailureMessage(actual interface{}) (message string) {
	if matcher.actualWith == nil {
		return format.Message(actual, "to panic")
	} else {
		return fmt.Sprintf("Panic message does not match.\n\n%v", matcher.expectedWith.FailureMessage(matcher.actualWith))
	}
}

func (matcher *PanicWithMessageToMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	panic("Not implemented")
}
