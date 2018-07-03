// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package validation

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {
	slice := []String123{String123("abc"), String123("123"), String123("xyz")}
	mp := map[string]String123{"c": String123("abc"), "b": String123("123"), "a": String123("xyz")}
	tests := []struct {
		tag   string
		value interface{}
		err   string
	}{
		{"t1", 123, ""},
		{"t2", String123("123"), ""},
		{"t3", String123("abc"), "error 123"},
		{"t4", []String123{}, ""},
		{"t5", slice, "0: error 123; 2: error 123."},
		{"t6", &slice, "0: error 123; 2: error 123."},
		{"t7", mp, "a: error 123; c: error 123."},
		{"t8", &mp, "a: error 123; c: error 123."},
		{"t9", map[string]String123{}, ""},
	}
	for _, test := range tests {
		err := Validate(test.value)
		assertError(t, test.err, err, test.tag)
	}

	// with rules
	err := Validate("123", &validateAbc{}, &validateXyz{})
	if assert.NotNil(t, err) {
		assert.Equal(t, "error abc", err.Error())
	}
	err = Validate("abc", &validateAbc{}, &validateXyz{})
	if assert.NotNil(t, err) {
		assert.Equal(t, "error xyz", err.Error())
	}
	err = Validate("abcxyz", &validateAbc{}, &validateXyz{})
	assert.Nil(t, err)

	err = Validate("123", &validateAbc{}, Skip, &validateXyz{})
	if assert.NotNil(t, err) {
		assert.Equal(t, "error abc", err.Error())
	}
	err = Validate("abc", &validateAbc{}, Skip, &validateXyz{})
	assert.Nil(t, err)
}

func TestBy(t *testing.T) {
	abcRule := By(func(value interface{}) error {
		s, _ := value.(string)
		if s != "abc" {
			return errors.New("must be abc")
		}
		return nil
	})
	assert.Nil(t, Validate("abc", abcRule))
	err := Validate("xyz", abcRule)
	if assert.NotNil(t, err) {
		assert.Equal(t, "must be abc", err.Error())
	}
}

func Test_skipRule_Validate(t *testing.T) {
	assert.Nil(t, Skip.Validate(100))
}

func assertError(t *testing.T, expected string, err error, tag string) {
	if expected == "" {
		assert.Nil(t, err, tag)
	} else if assert.NotNil(t, err, tag) {
		assert.Equal(t, expected, err.Error(), tag)
	}
}

type validateAbc struct{}

func (v *validateAbc) Validate(obj interface{}) error {
	if !strings.Contains(obj.(string), "abc") {
		return errors.New("error abc")
	}
	return nil
}

type validateXyz struct{}

func (v *validateXyz) Validate(obj interface{}) error {
	if !strings.Contains(obj.(string), "xyz") {
		return errors.New("error xyz")
	}
	return nil
}

type validateInternalError struct{}

func (v *validateInternalError) Validate(obj interface{}) error {
	if strings.Contains(obj.(string), "internal") {
		return NewInternalError(errors.New("error internal"))
	}
	return nil
}

type Model1 struct {
	A string
	B string
	c string
	D *string
	E String123
	F *String123
	G string `json:"g"`
}

type String123 string

func (s String123) Validate() error {
	if !strings.Contains(string(s), "123") {
		return errors.New("error 123")
	}
	return nil
}

type Model2 struct {
	Model3
	M3 Model3
	B  string
}

type Model3 struct {
	A string
}

func (m Model3) Validate() error {
	return ValidateStruct(&m,
		Field(&m.A, &validateAbc{}),
	)
}
