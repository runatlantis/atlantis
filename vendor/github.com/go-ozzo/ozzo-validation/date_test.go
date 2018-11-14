// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package validation

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDate(t *testing.T) {
	tests := []struct {
		tag    string
		layout string
		value  interface{}
		err    string
	}{
		{"t1", time.ANSIC, "", ""},
		{"t2", time.ANSIC, "Wed Feb  4 21:00:57 2009", ""},
		{"t3", time.ANSIC, "Wed Feb  29 21:00:57 2009", "must be a valid date"},
		{"t4", "2006-01-02", "2009-11-12", ""},
		{"t5", "2006-01-02", "2009-11-12 21:00:57", "must be a valid date"},
		{"t6", "2006-01-02", "2009-1-12", "must be a valid date"},
		{"t7", "2006-01-02", "2009-01-12", ""},
		{"t8", "2006-01-02", "2009-01-32", "must be a valid date"},
		{"t9", "2006-01-02", 1, "must be either a string or byte slice"},
	}

	for _, test := range tests {
		r := Date(test.layout)
		err := r.Validate(test.value)
		assertError(t, test.err, err, test.tag)
	}
}

func TestDateRule_Error(t *testing.T) {
	r := Date(time.ANSIC)
	assert.Equal(t, "must be a valid date", r.message)
	assert.Equal(t, "the data is out of range", r.rangeMessage)
	r.Error("123")
	r.RangeError("456")
	assert.Equal(t, "123", r.message)
	assert.Equal(t, "456", r.rangeMessage)
}

func TestDateRule_MinMax(t *testing.T) {
	r := Date(time.ANSIC)
	assert.True(t, r.min.IsZero())
	assert.True(t, r.max.IsZero())
	r.Min(time.Now())
	assert.False(t, r.min.IsZero())
	assert.True(t, r.max.IsZero())
	r.Max(time.Now())
	assert.False(t, r.max.IsZero())

	r2 := Date("2006-01-02").Min(time.Date(2000, 12, 1, 0, 0, 0, 0, time.UTC)).Max(time.Date(2020, 2, 1, 0, 0, 0, 0, time.UTC))
	assert.Nil(t, r2.Validate("2010-01-02"))
	err := r2.Validate("1999-01-02")
	if assert.NotNil(t, err) {
		assert.Equal(t, "the data is out of range", err.Error())
	}
	err2 := r2.Validate("2021-01-02")
	if assert.NotNil(t, err) {
		assert.Equal(t, "the data is out of range", err2.Error())
	}
}
