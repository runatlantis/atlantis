// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package validation

import (
	"testing"

	"database/sql"

	"github.com/stretchr/testify/assert"
)

func TestEnsureString(t *testing.T) {
	str := "abc"
	bytes := []byte("abc")

	tests := []struct {
		tag      string
		value    interface{}
		expected string
		hasError bool
	}{
		{"t1", "abc", "abc", false},
		{"t2", &str, "", true},
		{"t3", bytes, "abc", false},
		{"t4", &bytes, "", true},
		{"t5", 100, "", true},
	}
	for _, test := range tests {
		s, err := EnsureString(test.value)
		if test.hasError {
			assert.NotNil(t, err, test.tag)
		} else {
			assert.Nil(t, err, test.tag)
			assert.Equal(t, test.expected, s, test.tag)
		}
	}
}

type MyString string

func TestStringOrBytes(t *testing.T) {
	str := "abc"
	bytes := []byte("abc")
	var str2 string
	var bytes2 []byte
	var str3 MyString = "abc"
	var str4 *string

	tests := []struct {
		tag      string
		value    interface{}
		str      string
		bs       []byte
		isString bool
		isBytes  bool
	}{
		{"t1", str, "abc", nil, true, false},
		{"t2", &str, "", nil, false, false},
		{"t3", bytes, "", []byte("abc"), false, true},
		{"t4", &bytes, "", nil, false, false},
		{"t5", 100, "", nil, false, false},
		{"t6", str2, "", nil, true, false},
		{"t7", &str2, "", nil, false, false},
		{"t8", bytes2, "", nil, false, true},
		{"t9", &bytes2, "", nil, false, false},
		{"t10", str3, "abc", nil, true, false},
		{"t11", &str3, "", nil, false, false},
		{"t12", str4, "", nil, false, false},
	}
	for _, test := range tests {
		isString, str, isBytes, bs := StringOrBytes(test.value)
		assert.Equal(t, test.str, str, test.tag)
		assert.Equal(t, test.bs, bs, test.tag)
		assert.Equal(t, test.isString, isString, test.tag)
		assert.Equal(t, test.isBytes, isBytes, test.tag)
	}
}

func TestLengthOfValue(t *testing.T) {
	var a [3]int

	tests := []struct {
		tag    string
		value  interface{}
		length int
		err    string
	}{
		{"t1", "abc", 3, ""},
		{"t2", []int{1, 2}, 2, ""},
		{"t3", map[string]int{"A": 1, "B": 2}, 2, ""},
		{"t4", a, 3, ""},
		{"t5", &a, 0, "cannot get the length of ptr"},
		{"t6", 123, 0, "cannot get the length of int"},
	}

	for _, test := range tests {
		l, err := LengthOfValue(test.value)
		assert.Equal(t, test.length, l, test.tag)
		assertError(t, test.err, err, test.tag)
	}
}

func TestToInt(t *testing.T) {
	var a int

	tests := []struct {
		tag    string
		value  interface{}
		result int64
		err    string
	}{
		{"t1", 1, 1, ""},
		{"t2", int8(1), 1, ""},
		{"t3", int16(1), 1, ""},
		{"t4", int32(1), 1, ""},
		{"t5", int64(1), 1, ""},
		{"t6", &a, 0, "cannot convert ptr to int64"},
		{"t7", uint(1), 0, "cannot convert uint to int64"},
		{"t8", float64(1), 0, "cannot convert float64 to int64"},
		{"t9", "abc", 0, "cannot convert string to int64"},
		{"t10", []int{1, 2}, 0, "cannot convert slice to int64"},
		{"t11", map[string]int{"A": 1}, 0, "cannot convert map to int64"},
	}

	for _, test := range tests {
		l, err := ToInt(test.value)
		assert.Equal(t, test.result, l, test.tag)
		assertError(t, test.err, err, test.tag)
	}
}

func TestToUint(t *testing.T) {
	var a int
	var b uint

	tests := []struct {
		tag    string
		value  interface{}
		result uint64
		err    string
	}{
		{"t1", uint(1), 1, ""},
		{"t2", uint8(1), 1, ""},
		{"t3", uint16(1), 1, ""},
		{"t4", uint32(1), 1, ""},
		{"t5", uint64(1), 1, ""},
		{"t6", 1, 0, "cannot convert int to uint64"},
		{"t7", &a, 0, "cannot convert ptr to uint64"},
		{"t8", &b, 0, "cannot convert ptr to uint64"},
		{"t9", float64(1), 0, "cannot convert float64 to uint64"},
		{"t10", "abc", 0, "cannot convert string to uint64"},
		{"t11", []int{1, 2}, 0, "cannot convert slice to uint64"},
		{"t12", map[string]int{"A": 1}, 0, "cannot convert map to uint64"},
	}

	for _, test := range tests {
		l, err := ToUint(test.value)
		assert.Equal(t, test.result, l, test.tag)
		assertError(t, test.err, err, test.tag)
	}
}

func TestToFloat(t *testing.T) {
	var a int
	var b uint

	tests := []struct {
		tag    string
		value  interface{}
		result float64
		err    string
	}{
		{"t1", float32(1), 1, ""},
		{"t2", float64(1), 1, ""},
		{"t3", 1, 0, "cannot convert int to float64"},
		{"t4", uint(1), 0, "cannot convert uint to float64"},
		{"t5", &a, 0, "cannot convert ptr to float64"},
		{"t6", &b, 0, "cannot convert ptr to float64"},
		{"t7", "abc", 0, "cannot convert string to float64"},
		{"t8", []int{1, 2}, 0, "cannot convert slice to float64"},
		{"t9", map[string]int{"A": 1}, 0, "cannot convert map to float64"},
	}

	for _, test := range tests {
		l, err := ToFloat(test.value)
		assert.Equal(t, test.result, l, test.tag)
		assertError(t, test.err, err, test.tag)
	}
}

func TestIsEmpty(t *testing.T) {
	var s1 string
	var s2 string = "a"
	var s3 *string
	s4 := struct{}{}
	tests := []struct {
		tag   string
		value interface{}
		empty bool
	}{
		// nil
		{"t0", nil, true},
		// string
		{"t1.1", "", true},
		{"t1.2", "1", false},
		{"t1.3", MyString(""), true},
		{"t1.4", MyString("1"), false},
		// slice
		{"t2.1", []byte(""), true},
		{"t2.2", []byte("1"), false},
		// map
		{"t3.1", map[string]int{}, true},
		{"t3.2", map[string]int{"a": 1}, false},
		// bool
		{"t4.1", false, true},
		{"t4.2", true, false},
		// int
		{"t5.1", int(0), true},
		{"t5.2", int8(0), true},
		{"t5.3", int16(0), true},
		{"t5.4", int32(0), true},
		{"t5.5", int64(0), true},
		{"t5.6", int(1), false},
		{"t5.7", int8(1), false},
		{"t5.8", int16(1), false},
		{"t5.9", int32(1), false},
		{"t5.10", int64(1), false},
		// uint
		{"t6.1", uint(0), true},
		{"t6.2", uint8(0), true},
		{"t6.3", uint16(0), true},
		{"t6.4", uint32(0), true},
		{"t6.5", uint64(0), true},
		{"t6.6", uint(1), false},
		{"t6.7", uint8(1), false},
		{"t6.8", uint16(1), false},
		{"t6.9", uint32(1), false},
		{"t6.10", uint64(1), false},
		// float
		{"t7.1", float32(0), true},
		{"t7.2", float64(0), true},
		{"t7.3", float32(1), false},
		{"t7.4", float64(1), false},
		// interface, ptr
		{"t8.1", &s1, true},
		{"t8.2", &s2, false},
		{"t8.3", s3, true},
		// struct
		{"t9.1", s4, false},
		{"t9.2", &s4, false},
	}

	for _, test := range tests {
		empty := IsEmpty(test.value)
		assert.Equal(t, test.empty, empty, test.tag)
	}
}

func TestIndirect(t *testing.T) {
	var a int = 100
	var b *int
	var c *sql.NullInt64

	tests := []struct {
		tag    string
		value  interface{}
		result interface{}
		isNil  bool
	}{
		{"t1", 100, 100, false},
		{"t2", &a, 100, false},
		{"t3", b, nil, true},
		{"t4", nil, nil, true},
		{"t5", sql.NullInt64{Int64: 0, Valid: false}, nil, true},
		{"t6", sql.NullInt64{Int64: 1, Valid: false}, nil, true},
		{"t7", &sql.NullInt64{Int64: 0, Valid: false}, nil, true},
		{"t8", &sql.NullInt64{Int64: 1, Valid: false}, nil, true},
		{"t9", sql.NullInt64{Int64: 0, Valid: true}, int64(0), false},
		{"t10", sql.NullInt64{Int64: 1, Valid: true}, int64(1), false},
		{"t11", &sql.NullInt64{Int64: 0, Valid: true}, int64(0), false},
		{"t12", &sql.NullInt64{Int64: 1, Valid: true}, int64(1), false},
		{"t13", c, nil, true},
	}

	for _, test := range tests {
		result, isNil := Indirect(test.value)
		assert.Equal(t, test.result, result, test.tag)
		assert.Equal(t, test.isNil, isNil, test.tag)
	}
}
