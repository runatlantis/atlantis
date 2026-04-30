// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.

package recovery_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/runatlantis/atlantis/server/recovery"
)

func TestStack(t *testing.T) {
	tests := []struct {
		skip           int
		expContains    []string
		expNotContains []string
	}{
		{
			skip: 0,
			expContains: []string{
				"runtime.Caller(i)",
				"TestStack.func1.1: return string(recovery.Stack(tt.skip))",
				"recoveryTestFunc2: return f()",
				"recoveryTestFunc1: return recoveryTestFunc2(f)",
			},
			expNotContains: []string{},
		},
		{
			skip: 1,
			expContains: []string{
				"TestStack.func1.1: return string(recovery.Stack(tt.skip))",
				"recoveryTestFunc2: return f()",
				"recoveryTestFunc1: return recoveryTestFunc2(f)",
			},
			expNotContains: []string{
				"runtime.Caller(i)",
			},
		},
		{
			skip: 2,
			expContains: []string{
				"recoveryTestFunc2: return f()",
				"recoveryTestFunc1: return recoveryTestFunc2(f)",
			},
			expNotContains: []string{
				"runtime.Caller(i)",
				"TestStack.func1.1: return string(recovery.Stack(tt.skip))",
			},
		},
		{
			skip: 3,
			expContains: []string{
				"recoveryTestFunc1: return recoveryTestFunc2(f)",
			},
			expNotContains: []string{
				"runtime.Caller(i)",
				"TestStack.func1.1: return string(recovery.Stack(tt.skip))",
				"recoveryTestFunc2: return f()",
			},
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("skip %d", tt.skip), func(t *testing.T) {
			got := recoveryTestFunc1(func() string {
				return string(recovery.Stack(tt.skip))
			})
			for _, contain := range tt.expContains {
				if !strings.Contains(got, contain) {
					t.Fatalf("expected stack to contain %q but got:\n%s", contain, got)
				}
			}
			for _, notContain := range tt.expNotContains {
				if strings.Contains(got, notContain) {
					t.Fatalf("expected stack to not contain %q but got:\n%s", notContain, got)
				}
			}
		})
	}
}

func recoveryTestFunc1(f func() string) string {
	return recoveryTestFunc2(f)
}

func recoveryTestFunc2(f func() string) string {
	return f()
}
