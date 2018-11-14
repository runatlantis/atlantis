// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package validation

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatch(t *testing.T) {
	var v2 *string
	tests := []struct {
		tag   string
		re    string
		value interface{}
		err   string
	}{
		{"t1", "[a-z]+", "abc", ""},
		{"t2", "[a-z]+", "", ""},
		{"t3", "[a-z]+", v2, ""},
		{"t4", "[a-z]+", "123", "must be in a valid format"},
		{"t5", "[a-z]+", []byte("abc"), ""},
		{"t6", "[a-z]+", []byte("123"), "must be in a valid format"},
		{"t7", "[a-z]+", []byte(""), ""},
		{"t8", "[a-z]+", nil, ""},
	}

	for _, test := range tests {
		r := Match(regexp.MustCompile(test.re))
		err := r.Validate(test.value)
		assertError(t, test.err, err, test.tag)
	}
}

func Test_matchRule_Error(t *testing.T) {
	r := Match(regexp.MustCompile("[a-z]+"))
	assert.Equal(t, "must be in a valid format", r.message)
	r.Error("123")
	assert.Equal(t, "123", r.message)
}
