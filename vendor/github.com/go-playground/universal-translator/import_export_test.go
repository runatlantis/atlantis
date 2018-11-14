package ut

import (
	"fmt"
	"path/filepath"
	"testing"

	"os"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/nl"
)

// NOTES:
// - Run "go test" to run tests
// - Run "gocov test | gocov report" to report on test converage by file
// - Run "gocov test | gocov annotate -" to report on all code and functions, those ,marked with "MISS" were never called
//
// or
//
// -- may be a good idea to change to output path to somewherelike /tmp
// go test -coverprofile cover.out && go tool cover -html=cover.out -o cover.html
//

func TestExportImportBasic(t *testing.T) {

	e := en.New()
	uni := New(e, e)
	en, found := uni.GetTranslator("en") // or fallback if fails to find 'en'
	if !found {
		t.Fatalf("Expected '%t' Got '%t'", true, found)
	}

	translations := []struct {
		key           interface{}
		trans         string
		expected      error
		expectedError bool
		override      bool
	}{
		{
			key:      "test_trans",
			trans:    "Welcome {0}",
			expected: nil,
		},
		{
			key:      -1,
			trans:    "Welcome {0}",
			expected: nil,
		},
		{
			key:      "test_trans2",
			trans:    "{0} to the {1}.",
			expected: nil,
		},
		{
			key:      "test_trans3",
			trans:    "Welcome {0} to the {1}",
			expected: nil,
		},
		{
			key:      "test_trans4",
			trans:    "{0}{1}",
			expected: nil,
		},
		{
			key:           "test_trans",
			trans:         "{0}{1}",
			expected:      &ErrConflictingTranslation{locale: en.Locale(), key: "test_trans", text: "{0}{1}"},
			expectedError: true,
		},
		{
			key:           -1,
			trans:         "{0}{1}",
			expected:      &ErrConflictingTranslation{locale: en.Locale(), key: -1, text: "{0}{1}"},
			expectedError: true,
		},
		{
			key:      "test_trans",
			trans:    "Welcome {0} to the {1}.",
			expected: nil,
			override: true,
		},
	}

	for _, tt := range translations {

		err := en.Add(tt.key, tt.trans, tt.override)
		if err != tt.expected {
			if !tt.expectedError {
				t.Errorf("Expected '%s' Got '%s'", tt.expected, err)
			} else {
				if err.Error() != tt.expected.Error() {
					t.Errorf("Expected '%s' Got '%s'", tt.expected.Error(), err.Error())
				}
			}
		}
	}

	dirname := "testdata/translations"
	defer os.RemoveAll(dirname)

	err := uni.Export(FormatJSON, dirname)
	if err != nil {
		t.Fatalf("Expected '%v' Got '%s'", nil, err)
	}

	uni = New(e, e)

	err = uni.Import(FormatJSON, dirname)
	if err != nil {
		t.Fatalf("Expected '%v' Got '%s'", nil, err)
	}

	en, found = uni.GetTranslator("en") // or fallback if fails to find 'en'
	if !found {
		t.Fatalf("Expected '%t' Got '%t'", true, found)
	}

	tests := []struct {
		key           interface{}
		params        []string
		expected      string
		expectedError bool
	}{
		{
			key:      "test_trans",
			params:   []string{"Joeybloggs", "The Test"},
			expected: "Welcome Joeybloggs to the The Test.",
		},
		{
			key:      "test_trans2",
			params:   []string{"Joeybloggs", "The Test"},
			expected: "Joeybloggs to the The Test.",
		},
		{
			key:      "test_trans3",
			params:   []string{"Joeybloggs", "The Test"},
			expected: "Welcome Joeybloggs to the The Test",
		},
		{
			key:      "test_trans4",
			params:   []string{"Joeybloggs", "The Test"},
			expected: "JoeybloggsThe Test",
		},
		// bad translation
		{
			key:           "non-existant-key",
			params:        []string{"Joeybloggs", "The Test"},
			expected:      "",
			expectedError: true,
		},
	}

	for _, tt := range tests {

		s, err := en.T(tt.key, tt.params...)
		if s != tt.expected {
			if !tt.expectedError || (tt.expectedError && err != ErrUnknowTranslation) {
				t.Errorf("Expected '%s' Got '%s'", tt.expected, s)
			}
		}
	}
}

func TestExportImportCardinal(t *testing.T) {

	e := en.New()
	uni := New(e, e)
	en, found := uni.GetTranslator("en")
	if !found {
		t.Fatalf("Expected '%t' Got '%t'", true, found)
	}

	translations := []struct {
		key           interface{}
		trans         string
		rule          locales.PluralRule
		expected      error
		expectedError bool
		override      bool
	}{
		// bad translation
		{
			key:           "cardinal_test",
			trans:         "You have a day left.",
			rule:          locales.PluralRuleOne,
			expected:      &ErrCardinalTranslation{text: fmt.Sprintf("error: parameter '%s' not found, may want to use 'Add' instead of 'AddCardinal'. locale: '%s' key: '%v' text: '%s'", paramZero, en.Locale(), "cardinal_test", "You have a day left.")},
			expectedError: true,
		},
		{
			key:      "cardinal_test",
			trans:    "You have {0} day",
			rule:     locales.PluralRuleOne,
			expected: nil,
		},
		{
			key:      "cardinal_test",
			trans:    "You have {0} days left.",
			rule:     locales.PluralRuleOther,
			expected: nil,
		},
		{
			key:           "cardinal_test",
			trans:         "You have {0} days left.",
			rule:          locales.PluralRuleOther,
			expected:      &ErrConflictingTranslation{locale: en.Locale(), key: "cardinal_test", rule: locales.PluralRuleOther, text: "You have {0} days left."},
			expectedError: true,
		},
		{
			key:      "cardinal_test",
			trans:    "You have {0} day left.",
			rule:     locales.PluralRuleOne,
			expected: nil,
			override: true,
		},
	}

	for _, tt := range translations {

		err := en.AddCardinal(tt.key, tt.trans, tt.rule, tt.override)
		if err != tt.expected {
			if !tt.expectedError || err.Error() != tt.expected.Error() {
				t.Errorf("Expected '%s' Got '%s'", tt.expected, err)
			}
		}
	}

	dirname := "testdata/translations"
	defer os.RemoveAll(dirname)

	err := uni.Export(FormatJSON, dirname)
	if err != nil {
		t.Fatalf("Expected '%v' Got '%s'", nil, err)
	}

	uni = New(e, e)

	err = uni.Import(FormatJSON, dirname)
	if err != nil {
		t.Fatalf("Expected '%v' Got '%s'", nil, err)
	}

	en, found = uni.GetTranslator("en") // or fallback if fails to find 'en'
	if !found {
		t.Fatalf("Expected '%t' Got '%t'", true, found)
	}

	tests := []struct {
		key           interface{}
		num           float64
		digits        uint64
		param         string
		expected      string
		expectedError bool
	}{
		{
			key:      "cardinal_test",
			num:      1,
			digits:   0,
			param:    string(en.FmtNumber(1, 0)),
			expected: "You have 1 day left.",
		},
		// bad translation key
		{
			key:           "non-existant",
			num:           1,
			digits:        0,
			param:         string(en.FmtNumber(1, 0)),
			expected:      "",
			expectedError: true,
		},
	}

	for _, tt := range tests {

		s, err := en.C(tt.key, tt.num, tt.digits, tt.param)
		if err != nil {
			if !tt.expectedError && err != ErrUnknowTranslation {
				t.Errorf("Expected '<nil>' Got '%s'", err)
			}
		}

		if s != tt.expected {
			t.Errorf("Expected '%s' Got '%s'", tt.expected, s)
		}
	}
}

func TestExportImportOrdinal(t *testing.T) {

	e := en.New()
	uni := New(e, e)
	en, found := uni.GetTranslator("en")
	if !found {
		t.Fatalf("Expected '%t' Got '%t'", true, found)
	}

	translations := []struct {
		key           interface{}
		trans         string
		rule          locales.PluralRule
		expected      error
		expectedError bool
		override      bool
	}{
		// bad translation
		{
			key:           "day",
			trans:         "st",
			rule:          locales.PluralRuleOne,
			expected:      &ErrOrdinalTranslation{text: fmt.Sprintf("error: parameter '%s' not found, may want to use 'Add' instead of 'AddOrdinal'. locale: '%s' key: '%v' text: '%s'", paramZero, en.Locale(), "day", "st")},
			expectedError: true,
		},
		{
			key:      "day",
			trans:    "{0}sfefewt",
			rule:     locales.PluralRuleOne,
			expected: nil,
		},
		{
			key:      "day",
			trans:    "{0}nd",
			rule:     locales.PluralRuleTwo,
			expected: nil,
		},
		{
			key:      "day",
			trans:    "{0}rd",
			rule:     locales.PluralRuleFew,
			expected: nil,
		},
		{
			key:      "day",
			trans:    "{0}th",
			rule:     locales.PluralRuleOther,
			expected: nil,
		},
		// bad translation
		{
			key:           "day",
			trans:         "{0}th",
			rule:          locales.PluralRuleOther,
			expected:      &ErrConflictingTranslation{locale: en.Locale(), key: "day", rule: locales.PluralRuleOther, text: "{0}th"},
			expectedError: true,
		},
		{
			key:      "day",
			trans:    "{0}st",
			rule:     locales.PluralRuleOne,
			expected: nil,
			override: true,
		},
	}

	for _, tt := range translations {

		err := en.AddOrdinal(tt.key, tt.trans, tt.rule, tt.override)
		if err != tt.expected {
			if !tt.expectedError || err.Error() != tt.expected.Error() {
				t.Errorf("Expected '<nil>' Got '%s'", err)
			}
		}
	}

	dirname := "testdata/translations"
	defer os.RemoveAll(dirname)

	err := uni.Export(FormatJSON, dirname)
	if err != nil {
		t.Fatalf("Expected '%v' Got '%s'", nil, err)
	}

	uni = New(e, e)

	err = uni.Import(FormatJSON, dirname)
	if err != nil {
		t.Fatalf("Expected '%v' Got '%s'", nil, err)
	}

	en, found = uni.GetTranslator("en") // or fallback if fails to find 'en'
	if !found {
		t.Fatalf("Expected '%t' Got '%t'", true, found)
	}

	tests := []struct {
		key           interface{}
		num           float64
		digits        uint64
		param         string
		expected      string
		expectedError bool
	}{
		{
			key:      "day",
			num:      1,
			digits:   0,
			param:    string(en.FmtNumber(1, 0)),
			expected: "1st",
		},
		{
			key:      "day",
			num:      2,
			digits:   0,
			param:    string(en.FmtNumber(2, 0)),
			expected: "2nd",
		},
		{
			key:      "day",
			num:      3,
			digits:   0,
			param:    string(en.FmtNumber(3, 0)),
			expected: "3rd",
		},
		{
			key:      "day",
			num:      4,
			digits:   0,
			param:    string(en.FmtNumber(4, 0)),
			expected: "4th",
		},
		{
			key:      "day",
			num:      10258.43,
			digits:   0,
			param:    string(en.FmtNumber(10258.43, 0)),
			expected: "10,258th",
		},
		// bad translation
		{
			key:           "d-day",
			num:           10258.43,
			digits:        0,
			param:         string(en.FmtNumber(10258.43, 0)),
			expected:      "",
			expectedError: true,
		},
	}

	for _, tt := range tests {

		s, err := en.O(tt.key, tt.num, tt.digits, tt.param)
		if err != nil {
			if !tt.expectedError && err != ErrUnknowTranslation {
				t.Errorf("Expected '<nil>' Got '%s'", err)
			}
		}

		if s != tt.expected {
			t.Errorf("Expected '%s' Got '%s'", tt.expected, s)
		}
	}
}

func TestExportImportRange(t *testing.T) {

	n := nl.New()
	uni := New(n, n)

	// dutch
	nl, found := uni.GetTranslator("nl")
	if !found {
		t.Fatalf("Expected '%t' Got '%t'", true, found)
	}

	translations := []struct {
		key           interface{}
		trans         string
		rule          locales.PluralRule
		expected      error
		expectedError bool
		override      bool
	}{
		// bad translation
		{
			key:           "day",
			trans:         "er -{1} dag vertrokken",
			rule:          locales.PluralRuleOne,
			expected:      &ErrRangeTranslation{text: fmt.Sprintf("error: parameter '%s' not found, are you sure you're adding a Range Translation? locale: '%s' key: '%s' text: '%s'", paramZero, nl.Locale(), "day", "er -{1} dag vertrokken")},
			expectedError: true,
		},
		// bad translation
		{
			key:           "day",
			trans:         "er {0}- dag vertrokken",
			rule:          locales.PluralRuleOne,
			expected:      &ErrRangeTranslation{text: fmt.Sprintf("error: parameter '%s' not found, a Range Translation requires two parameters. locale: '%s' key: '%s' text: '%s'", paramOne, nl.Locale(), "day", "er {0}- dag vertrokken")},
			expectedError: true,
		},
		{
			key:      "day",
			trans:    "er {0}-{1} dag",
			rule:     locales.PluralRuleOne,
			expected: nil,
		},
		{
			key:      "day",
			trans:    "er zijn {0}-{1} dagen over",
			rule:     locales.PluralRuleOther,
			expected: nil,
		},
		// bad translation
		{
			key:           "day",
			trans:         "er zijn {0}-{1} dagen over",
			rule:          locales.PluralRuleOther,
			expected:      &ErrConflictingTranslation{locale: nl.Locale(), key: "day", rule: locales.PluralRuleOther, text: "er zijn {0}-{1} dagen over"},
			expectedError: true,
		},
		{
			key:      "day",
			trans:    "er {0}-{1} dag vertrokken",
			rule:     locales.PluralRuleOne,
			expected: nil,
			override: true,
		},
	}

	for _, tt := range translations {

		err := nl.AddRange(tt.key, tt.trans, tt.rule, tt.override)
		if err != tt.expected {
			if !tt.expectedError || err.Error() != tt.expected.Error() {
				t.Errorf("Expected '%#v' Got '%s'", tt.expected, err)
			}
		}
	}

	dirname := "testdata/translations"
	defer os.RemoveAll(dirname)

	err := uni.Export(FormatJSON, dirname)
	if err != nil {
		t.Fatalf("Expected '%v' Got '%s'", nil, err)
	}

	uni = New(n, n)

	err = uni.Import(FormatJSON, dirname)
	if err != nil {
		t.Fatalf("Expected '%v' Got '%s'", nil, err)
	}

	nl, found = uni.GetTranslator("nl") // or fallback if fails to find 'en'
	if !found {
		t.Fatalf("Expected '%t' Got '%t'", true, found)
	}

	tests := []struct {
		key           interface{}
		num1          float64
		digits1       uint64
		num2          float64
		digits2       uint64
		param1        string
		param2        string
		expected      string
		expectedError bool
	}{
		{
			key:      "day",
			num1:     1,
			digits1:  0,
			num2:     2,
			digits2:  0,
			param1:   string(nl.FmtNumber(1, 0)),
			param2:   string(nl.FmtNumber(2, 0)),
			expected: "er zijn 1-2 dagen over",
		},
		{
			key:      "day",
			num1:     0,
			digits1:  0,
			num2:     1,
			digits2:  0,
			param1:   string(nl.FmtNumber(0, 0)),
			param2:   string(nl.FmtNumber(1, 0)),
			expected: "er 0-1 dag vertrokken",
		},
		{
			key:      "day",
			num1:     0,
			digits1:  0,
			num2:     2,
			digits2:  0,
			param1:   string(nl.FmtNumber(0, 0)),
			param2:   string(nl.FmtNumber(2, 0)),
			expected: "er zijn 0-2 dagen over",
		},
		// bad translations from here
		{
			key:           "d-day",
			num1:          0,
			digits1:       0,
			num2:          2,
			digits2:       0,
			param1:        string(nl.FmtNumber(0, 0)),
			param2:        string(nl.FmtNumber(2, 0)),
			expected:      "",
			expectedError: true,
		},
	}

	for _, tt := range tests {

		s, err := nl.R(tt.key, tt.num1, tt.digits1, tt.num2, tt.digits2, tt.param1, tt.param2)
		if err != nil {
			if !tt.expectedError && err != ErrUnknowTranslation {
				t.Errorf("Expected '<nil>' Got '%s'", err)
			}
		}

		if s != tt.expected {
			t.Errorf("Expected '%s' Got '%s'", tt.expected, s)
		}
	}
}

func TestImportRecursive(t *testing.T) {

	e := en.New()
	uni := New(e, e)

	dirname := "testdata/nested1"
	err := uni.Import(FormatJSON, dirname)
	if err != nil {
		t.Fatalf("Expected '%v' Got '%s'", nil, err)
	}

	en, found := uni.GetTranslator("en") // or fallback if fails to find 'en'
	if !found {
		t.Fatalf("Expected '%t' Got '%t'", true, found)
	}

	tests := []struct {
		key           interface{}
		params        []string
		expected      string
		expectedError bool
	}{
		{
			key:      "test_trans",
			params:   []string{"Joeybloggs", "The Test"},
			expected: "Welcome Joeybloggs to the The Test.",
		},
		{
			key:      "test_trans2",
			params:   []string{"Joeybloggs", "The Test"},
			expected: "Joeybloggs to the The Test.",
		},
		{
			key:      "test_trans3",
			params:   []string{"Joeybloggs", "The Test"},
			expected: "Welcome Joeybloggs to the The Test",
		},
		{
			key:      "test_trans4",
			params:   []string{"Joeybloggs", "The Test"},
			expected: "JoeybloggsThe Test",
		},
		// bad translation
		{
			key:           "non-existant-key",
			params:        []string{"Joeybloggs", "The Test"},
			expected:      "",
			expectedError: true,
		},
	}

	for _, tt := range tests {

		s, err := en.T(tt.key, tt.params...)
		if s != tt.expected {
			if !tt.expectedError || (tt.expectedError && err != ErrUnknowTranslation) {
				t.Errorf("Expected '%s' Got '%s'", tt.expected, s)
			}
		}
	}
}

func TestBadImport(t *testing.T) {

	// test non existant file
	e := en.New()
	uni := New(e, e)

	filename := "testdata/non-existant-file.json"
	expected := "stat testdata/non-existant-file.json: no such file or directory"
	err := uni.Import(FormatJSON, filename)
	if err == nil || err.Error() != expected {
		t.Fatalf("Expected '%s' Got '%s'", expected, err)
	}

	// test bad parameter basic translation
	filename = "testdata/bad-translation1.json"
	expected = "error: bad parameter syntax, missing parameter '{0}' in translation. locale: 'en' key: 'test_trans3' text: 'Welcome {lettersnotpermitted} to the {1}'"
	err = uni.Import(FormatJSON, filename)
	if err == nil || err.Error() != expected {
		t.Fatalf("Expected '%s' Got '%s'", expected, err)
	}

	// test missing bracket basic translation
	filename = "testdata/bad-translation2.json"
	expected = "error: missing bracket '{}', in translation. locale: 'en' key: 'test_trans3' text: 'Welcome {0 to the {1}'"
	err = uni.Import(FormatJSON, filename)
	if err == nil || err.Error() != expected {
		t.Fatalf("Expected '%s' Got '%s'", expected, err)
	}

	// test missing locale basic translation
	filename = "testdata/bad-translation3.json"
	expected = "error: locale 'nl' not registered."
	err = uni.Import(FormatJSON, filename)
	if err == nil || err.Error() != expected {
		t.Fatalf("Expected '%s' Got '%s'", expected, err)
	}

	// test bad plural definition
	filename = "testdata/bad-translation4.json"
	expected = "error: bad plural definition 'ut.translation{Locale:\"en\", Key:\"cardinal_test\", Translation:\"You have {0} day left.\", PluralType:\"NotAPluralType\", PluralRule:\"One\", OverrideExisting:false}'"
	err = uni.Import(FormatJSON, filename)
	if err == nil || err.Error() != expected {
		t.Fatalf("Expected '%s' Got '%s'", expected, err)
	}

	// test bad plural rule for locale
	filename = "testdata/bad-translation5.json"
	expected = "error: cardinal plural rule 'Many' does not exist for locale 'en' key: 'cardinal_test' text: 'You have {0} day left.'"
	err = uni.Import(FormatJSON, filename)
	if err == nil || err.Error() != expected {
		t.Fatalf("Expected '%s' Got '%s'", expected, err)
	}

	// test invalid JSON
	filename = "testdata/bad-translation6.json"
	expected = "invalid character ']' after object key:value pair"
	err = uni.Import(FormatJSON, filename)
	if err == nil || err.Error() != expected {
		t.Fatalf("Expected '%s' Got '%s'", expected, err)
	}

	// test bad io.Reader
	f, err := os.Open(filename)
	if err != nil {
		t.Fatalf("Expected '%v' Got '%s'", nil, err)
	}
	f.Close()

	expected = "read testdata/bad-translation6.json: bad file descriptor"
	err = uni.ImportByReader(FormatJSON, f)
	if err == nil || err.Error() != expected {
		t.Fatalf("Expected '%s' Got '%s'", expected, err)
	}
}

func TestBadExport(t *testing.T) {

	// test readonly directory
	e := en.New()
	uni := New(e, e)

	en, found := uni.GetTranslator("en") // or fallback if fails to find 'en'
	if !found {
		t.Fatalf("Expected '%t' Got '%t'", true, found)
	}

	dirname := "testdata/readonly"
	err := os.Mkdir(dirname, 0444)
	if err != nil {
		t.Fatalf("Expected '%v' Got '%s'", nil, err)
	}
	defer os.RemoveAll(dirname)

	en.Add("day", "this is a day", false)

	expected := "open testdata/readonly/en.json: permission denied"
	err = uni.Export(FormatJSON, dirname)
	if err == nil || err.Error() != expected {
		t.Fatalf("Expected '%s' Got '%s'", expected, err)
	}

	// test exporting into directory inside readonly directory
	expected = "stat testdata/readonly/inner: permission denied"
	err = uni.Export(FormatJSON, filepath.Join(dirname, "inner"))
	if err == nil || err.Error() != expected {
		t.Fatalf("Expected '%s' Got '%s'", expected, err)
	}
}
