// Copyright 2017 HootSuite Media Inc.
//
// Licensed under the Apache License, Version 2.0 (the License);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an AS IS BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// Modified hereafter by contributors to runatlantis/atlantis.
//
package testing

import (
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

// Assert fails the test if the condition is false.
// Taken from https://github.com/benbjohnson/testing.
func Assert(tb testing.TB, condition bool, msg string, v ...interface{}) {
	if !condition {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: "+msg+"\033[39m\n\n", append([]interface{}{filepath.Base(file), line}, v...)...)
		tb.FailNow()
	}
}

// Ok fails the test if an err is not nil.
// Taken from https://github.com/benbjohnson/testing.
func Ok(tb testing.TB, err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: unexpected error: %s\033[39m\n\n", filepath.Base(file), line, err.Error())
		tb.FailNow()
	}
}

// Equals fails the test if exp is not equal to act.
// Taken from https://github.com/benbjohnson/testing.
func Equals(tb testing.TB, exp, act interface{}) {
	if !reflect.DeepEqual(exp, act) {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d:\n\n\texp: %#v\n\n\tgot: %#v\033[39m\n\n", filepath.Base(file), line, exp, act)
		tb.FailNow()
	}
}

// ErrEquals fails the test if act is nil or act.Error() != exp
func ErrEquals(tb testing.TB, exp string, act error) {
	tb.Helper()
	if act == nil {
		tb.Errorf("exp err %q but err was nil", exp)
	}
	if act.Error() != exp {
		tb.Errorf("exp err: %q but got: %q", exp, act.Error())
	}
}

// Contains fails the test if the slice doesn't contain the expected element
func Contains(tb testing.TB, exp interface{}, slice []string) {
	for _, v := range slice {
		if v == exp {
			return
		}
	}
	_, file, line, _ := runtime.Caller(1)
	fmt.Printf("\033[31m%s:%d:\n\n\texp: %#v\n\n\twas not in: %#v\033[39m\n\n", filepath.Base(file), line, exp, slice)
	tb.FailNow()
}
