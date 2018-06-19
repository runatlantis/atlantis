// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIn(t *testing.T) {
	var v int = 1
	var v2 *int
	tests := []struct {
		tag    string
		values []interface{}
		value  interface{}
		err    string
	}{
		{"t0", []interface{}{1, 2}, 0, ""},
		{"t1", []interface{}{1, 2}, 1, ""},
		{"t2", []interface{}{1, 2}, 2, ""},
		{"t3", []interface{}{1, 2}, 3, "must be a valid value"},
		{"t4", []interface{}{}, 3, "must be a valid value"},
		{"t5", []interface{}{1, 2}, "1", "must be a valid value"},
		{"t6", []interface{}{1, 2}, &v, ""},
		{"t7", []interface{}{1, 2}, v2, ""},
	}

	for _, test := range tests {
		r := In(test.values...)
		err := r.Validate(test.value)
		assertError(t, test.err, err, test.tag)
	}
}

func Test_inRule_Error(t *testing.T) {
	r := In(1, 2, 3)
	assert.Equal(t, "must be a valid value", r.message)
	r.Error("123")
	assert.Equal(t, "123", r.message)
}
