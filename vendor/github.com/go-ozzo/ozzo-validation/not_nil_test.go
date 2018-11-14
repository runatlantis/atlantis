// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type MyInterface interface {
	Hello()
}

func TestNotNil(t *testing.T) {
	var v1 []int
	var v2 map[string]int
	var v3 *int
	var v4 interface{}
	var v5 MyInterface
	tests := []struct {
		tag   string
		value interface{}
		err   string
	}{
		{"t1", v1, "is required"},
		{"t2", v2, "is required"},
		{"t3", v3, "is required"},
		{"t4", v4, "is required"},
		{"t5", v5, "is required"},
		{"t6", "", ""},
		{"t7", 0, ""},
	}

	for _, test := range tests {
		r := NotNil
		err := r.Validate(test.value)
		assertError(t, test.err, err, test.tag)
	}
}

func Test_notNilRule_Error(t *testing.T) {
	r := NotNil
	assert.Equal(t, "is required", r.message)
	r2 := r.Error("123")
	assert.Equal(t, "is required", r.message)
	assert.Equal(t, "123", r2.message)
}
