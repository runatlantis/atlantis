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

package testing

import (
	"strings"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/go-test/deep"
)

// Assert fails the test if the condition is false.
// Taken from https://github.com/benbjohnson/testing.
func Assert(tb testing.TB, condition bool, msg string, v ...interface{}) {
	tb.Helper()
	if !condition {
		errLog(tb, msg, v...)
		tb.FailNow()
	}
}

// Ok fails the test if an err is not nil.
// Taken from https://github.com/benbjohnson/testing.
func Ok(tb testing.TB, err error) {
	tb.Helper()
	if err != nil {
		errLog(tb, "unexpected error: %s", err.Error())
		tb.FailNow()
	}
}

// Equals fails the test if exp is not equal to act.
// Taken from https://github.com/benbjohnson/testing.
func Equals(tb testing.TB, exp, act interface{}) {
	tb.Helper()
	if diff := deep.Equal(exp, act); diff != nil {
		errLog(tb, "%s\n\nexp: %s******\ngot: %s", diff, spew.Sdump(exp), spew.Sdump(act))
		tb.FailNow()
	}
}

// ErrEquals fails the test if act is nil or act.Error() != exp
func ErrEquals(tb testing.TB, exp string, act error) {
	tb.Helper()
	if act == nil {
		errLog(tb, "exp err %q but err was nil\n", exp)
		tb.FailNow()
	}
	if act.Error() != exp {
		errLog(tb, "exp err: %q but got: %q\n", exp, act.Error())
		tb.FailNow()
	}
}

// ErrContains fails the test if act is nil or act.Error() does not contain
// substr.
func ErrContains(tb testing.TB, substr string, act error) {
	tb.Helper()
	if act == nil {
		errLog(tb, "exp err to contain %q but err was nil", substr)
		tb.FailNow()
	}
	if !strings.Contains(act.Error(), substr) {
		errLog(tb, "exp err %q to contain %q", act.Error(), substr)
		tb.FailNow()
	}
}

// Contains fails the test if the slice doesn't contain the expected element
func Contains(tb testing.TB, exp interface{}, slice []string) {
	tb.Helper()
	for _, v := range slice {
		if v == exp {
			return
		}
	}
	errLog(tb, "exp: %#v\n\n\twas not in: %#v", exp, slice)
	tb.FailNow()
}

func errLog(tb testing.TB, fmt string, args ...interface{}) {
	tb.Helper()
	tb.Logf("\033[31m"+fmt+"\033[39m", args...)
}
