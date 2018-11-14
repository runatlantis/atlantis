// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequired(t *testing.T) {
	s1 := "123"
	s2 := ""
	tests := []struct {
		tag   string
		value interface{}
		err   string
	}{
		{"t1", 123, ""},
		{"t2", "", "cannot be blank"},
		{"t3", &s1, ""},
		{"t4", &s2, "cannot be blank"},
		{"t5", nil, "cannot be blank"},
	}

	for _, test := range tests {
		r := Required
		err := r.Validate(test.value)
		assertError(t, test.err, err, test.tag)
	}
}

func TestNilOrNotEmpty(t *testing.T) {
	s1 := "123"
	s2 := ""
	tests := []struct {
		tag   string
		value interface{}
		err   string
	}{
		{"t1", 123, ""},
		{"t2", "", "cannot be blank"},
		{"t3", &s1, ""},
		{"t4", &s2, "cannot be blank"},
		{"t5", nil, ""},
	}

	for _, test := range tests {
		r := NilOrNotEmpty
		err := r.Validate(test.value)
		assertError(t, test.err, err, test.tag)
	}
}

func Test_requiredRule_Error(t *testing.T) {
	r := Required
	assert.Equal(t, "cannot be blank", r.message)
	assert.False(t, r.skipNil)
	r2 := r.Error("123")
	assert.Equal(t, "cannot be blank", r.message)
	assert.False(t, r.skipNil)
	assert.Equal(t, "123", r2.message)
	assert.False(t, r2.skipNil)

	r = NilOrNotEmpty
	assert.Equal(t, "cannot be blank", r.message)
	assert.True(t, r.skipNil)
	r2 = r.Error("123")
	assert.Equal(t, "cannot be blank", r.message)
	assert.True(t, r.skipNil)
	assert.Equal(t, "123", r2.message)
	assert.True(t, r2.skipNil)
}
