// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package is

import (
	"testing"

	"github.com/go-ozzo/ozzo-validation"
	"github.com/stretchr/testify/assert"
)

func TestAll(t *testing.T) {
	tests := []struct {
		tag            string
		rule           validation.Rule
		valid, invalid string
		err            string
	}{
		{"Email", Email, "test@example.com", "example.com", "must be a valid email address"},
		{"URL", URL, "http://example.com", "examplecom", "must be a valid URL"},
		{"RequestURL", RequestURL, "http://example.com", "examplecom", "must be a valid request URL"},
		{"RequestURI", RequestURI, "http://example.com", "examplecom", "must be a valid request URI"},
		{"Alpha", Alpha, "abcd", "ab12", "must contain English letters only"},
		{"Digit", Digit, "123", "12ab", "must contain digits only"},
		{"Alphanumeric", Alphanumeric, "abc123", "abc.123", "must contain English letters and digits only"},
		{"UTFLetter", UTFLetter, "ａｂｃ", "１２３", "must contain unicode letter characters only"},
		{"UTFDigit", UTFDigit, "１２３", "ａｂｃ", "must contain unicode decimal digits only"},
		{"UTFNumeric", UTFNumeric, "１２３", "ａｂｃ.１２３", "must contain unicode number characters only"},
		{"UTFLetterNumeric", UTFLetterNumeric, "ａｂｃ１２３", "ａｂｃ.１２３", "must contain unicode letters and numbers only"},
		{"LowerCase", LowerCase, "ａｂc", "Aｂｃ", "must be in lower case"},
		{"UpperCase", UpperCase, "ABC", "ABｃ", "must be in upper case"},
		{"IP", IP, "74.125.19.99", "74.125.19.999", "must be a valid IP address"},
		{"IPv4", IPv4, "74.125.19.99", "2001:4860:0:2001::68", "must be a valid IPv4 address"},
		{"IPv6", IPv6, "2001:4860:0:2001::68", "74.125.19.99", "must be a valid IPv6 address"},
		{"MAC", MAC, "0123.4567.89ab", "74.125.19.99", "must be a valid MAC address"},
		{"DNSName", DNSName, "example.com", "abc%", "must be a valid DNS name"},
		{"Host", Host, "example.com", "abc%", "must be a valid IP address or DNS name"},
		{"Port", Port, "123", "99999", "must be a valid port number"},
		{"Latitude", Latitude, "23.123", "100", "must be a valid latitude"},
		{"Longitude", Longitude, "123.123", "abc", "must be a valid longitude"},
		{"SSN", SSN, "100-00-1000", "100-0001000", "must be a valid social security number"},
		{"Semver", Semver, "1.0.0", "1.0.0.0", "must be a valid semantic version"},
		{"ISBN", ISBN, "1-61729-085-8", "1-61729-085-81", "must be a valid ISBN"},
		{"ISBN10", ISBN10, "1-61729-085-8", "1-61729-085-81", "must be a valid ISBN-10"},
		{"ISBN13", ISBN13, "978-4-87311-368-5", "978-4-87311-368-a", "must be a valid ISBN-13"},
		{"UUID", UUID, "a987fbc9-4bed-3078-cf07-9141ba07c9f1", "a987fbc9-4bed-3078-cf07-9141ba07c9f3a", "must be a valid UUID"},
		{"UUIDv3", UUIDv3, "b987fbc9-4bed-3078-cf07-9141ba07c9f3", "b987fbc9-4bed-4078-cf07-9141ba07c9f3", "must be a valid UUID v3"},
		{"UUIDv4", UUIDv4, "57b73598-8764-4ad0-a76a-679bb6640eb1", "b987fbc9-4bed-3078-cf07-9141ba07c9f3", "must be a valid UUID v4"},
		{"UUIDv5", UUIDv5, "987fbc97-4bed-5078-af07-9141ba07c9f3", "b987fbc9-4bed-3078-cf07-9141ba07c9f3", "must be a valid UUID v5"},
		{"MongoID", MongoID, "507f1f77bcf86cd799439011", "507f1f77bcf86cd79943901", "must be a valid hex-encoded MongoDB ObjectId"},
		{"CreditCard", CreditCard, "375556917985515", "375556917985516", "must be a valid credit card number"},
		{"JSON", JSON, "[1, 2]", "[1, 2,]", "must be in valid JSON format"},
		{"ASCII", ASCII, "abc", "ａabc", "must contain ASCII characters only"},
		{"PrintableASCII", PrintableASCII, "abc", "ａabc", "must contain printable ASCII characters only"},
		{"CountryCode2", CountryCode2, "US", "XY", "must be a valid two-letter country code"},
		{"CountryCode3", CountryCode3, "USA", "XYZ", "must be a valid three-letter country code"},
		{"DialString", DialString, "localhost.local:1", "localhost.loc:100000", "must be a valid dial string"},
		{"DataURI", DataURI, "data:image/png;base64,TG9yZW0gaXBzdW0gZG9sb3Igc2l0IGFtZXQsIGNvbnNlY3RldHVyIGFkaXBpc2NpbmcgZWxpdC4=", "image/gif;base64,U3VzcGVuZGlzc2UgbGVjdHVzIGxlbw==", "must be a Base64-encoded data URI"},
		{"Base64", Base64, "TG9yZW0gaXBzdW0gZG9sb3Igc2l0IGFtZXQsIGNvbnNlY3RldHVyIGFkaXBpc2NpbmcgZWxpdC4=", "image", "must be encoded in Base64"},
		{"Multibyte", Multibyte, "ａｂｃ", "abc", "must contain multibyte characters"},
		{"FullWidth", FullWidth, "３ー０", "abc", "must contain full-width characters"},
		{"HalfWidth", HalfWidth, "abc123い", "００１１", "must contain half-width characters"},
		{"VariableWidth", VariableWidth, "３ー０123", "abc", "must contain both full-width and half-width characters"},
		{"Hexadecimal", Hexadecimal, "FEF", "FTF", "must be a valid hexadecimal number"},
		{"HexColor", HexColor, "F00", "FTF", "must be a valid hexadecimal color code"},
		{"RGBColor", RGBColor, "rgb(100, 200, 1)", "abc", "must be a valid RGB color code"},
		{"Int", Int, "100", "1.1", "must be an integer number"},
		{"Float", Float, "1.1", "a.1", "must be a floating point number"},
		{"VariableWidth", VariableWidth, "", "", ""},
	}

	for _, test := range tests {
		err := test.rule.Validate("")
		assert.Nil(t, err, test.tag)
		err = test.rule.Validate(test.valid)
		assert.Nil(t, err, test.tag)
		err = test.rule.Validate(&test.valid)
		assert.Nil(t, err, test.tag)
		err = test.rule.Validate(test.invalid)
		assertError(t, test.err, err, test.tag)
		err = test.rule.Validate(&test.invalid)
		assertError(t, test.err, err, test.tag)
	}
}

func assertError(t *testing.T, expected string, err error, tag string) {
	if expected == "" {
		assert.Nil(t, err, tag)
	} else if assert.NotNil(t, err, tag) {
		assert.Equal(t, expected, err.Error(), tag)
	}
}
