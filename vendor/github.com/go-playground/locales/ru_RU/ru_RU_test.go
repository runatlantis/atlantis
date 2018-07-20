package ru_RU

import (
	"testing"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

func TestLocale(t *testing.T) {

	trans := New()
	expected := "ru_RU"

	if trans.Locale() != expected {
		t.Errorf("Expected '%s' Got '%s'", expected, trans.Locale())
	}
}

func TestPluralsRange(t *testing.T) {

	trans := New()

	tests := []struct {
		expected locales.PluralRule
	}{
		{
			expected: locales.PluralRuleOther,
		},
	}

	rules := trans.PluralsRange()
	expected := 4
	if len(rules) != expected {
		t.Errorf("Expected '%d' Got '%d'", expected, len(rules))
	}

	for _, tt := range tests {

		r := locales.PluralRuleUnknown

		for i := 0; i < len(rules); i++ {
			if rules[i] == tt.expected {
				r = rules[i]
				break
			}
		}
		if r == locales.PluralRuleUnknown {
			t.Errorf("Expected '%s' Got '%s'", tt.expected, r)
		}
	}
}

func TestPluralsOrdinal(t *testing.T) {

	trans := New()

	tests := []struct {
		expected locales.PluralRule
	}{
		{
			expected: locales.PluralRuleOther,
		},
	}

	rules := trans.PluralsOrdinal()
	expected := 1
	if len(rules) != expected {
		t.Errorf("Expected '%d' Got '%d'", expected, len(rules))
	}

	for _, tt := range tests {

		r := locales.PluralRuleUnknown

		for i := 0; i < len(rules); i++ {
			if rules[i] == tt.expected {
				r = rules[i]
				break
			}
		}
		if r == locales.PluralRuleUnknown {
			t.Errorf("Expected '%s' Got '%s'", tt.expected, r)
		}
	}
}

func TestPluralsCardinal(t *testing.T) {

	trans := New()

	tests := []struct {
		expected locales.PluralRule
	}{
		{
			expected: locales.PluralRuleOne,
		},
		{
			expected: locales.PluralRuleFew,
		},
		{
			expected: locales.PluralRuleMany,
		},
		{
			expected: locales.PluralRuleOther,
		},
	}

	rules := trans.PluralsCardinal()
	expected := 4
	if len(rules) != expected {
		t.Errorf("Expected '%d' Got '%d'", expected, len(rules))
	}

	for _, tt := range tests {

		r := locales.PluralRuleUnknown

		for i := 0; i < len(rules); i++ {
			if rules[i] == tt.expected {
				r = rules[i]
				break
			}
		}
		if r == locales.PluralRuleUnknown {
			t.Errorf("Expected '%s' Got '%s'", tt.expected, r)
		}
	}
}

func TestRangePlurals(t *testing.T) {

	trans := New()

	tests := []struct {
		num1     float64
		v1       uint64
		num2     float64
		v2       uint64
		expected locales.PluralRule
	}{
		{
			num1:     1,
			v1:       1,
			num2:     2,
			v2:       2,
			expected: locales.PluralRuleOther,
		},
		{
			num1:     1,
			v1:       0,
			num2:     2,
			v2:       0,
			expected: locales.PluralRuleFew,
		},
		{
			num1:     1,
			v1:       0,
			num2:     21,
			v2:       0,
			expected: locales.PluralRuleOne,
		},
		{
			num1:     1,
			v1:       0,
			num2:     5,
			v2:       0,
			expected: locales.PluralRuleMany,
		},
		{
			num1:     1,
			v1:       0,
			num2:     10,
			v2:       0,
			expected: locales.PluralRuleMany,
		},
		{
			num1:     1,
			v1:       0,
			num2:     10.0,
			v2:       1,
			expected: locales.PluralRuleOther,
		},
		{
			num1:     2,
			v1:       0,
			num2:     21,
			v2:       0,
			expected: locales.PluralRuleOne,
		},
		{
			num1:     2,
			v1:       0,
			num2:     22,
			v2:       0,
			expected: locales.PluralRuleFew,
		},
		{
			num1:     2,
			v1:       0,
			num2:     5,
			v2:       0,
			expected: locales.PluralRuleMany,
		},
		{
			num1:     2,
			v1:       0,
			num2:     10,
			v2:       1,
			expected: locales.PluralRuleOther,
		},
		{
			num1:     0,
			v1:       0,
			num2:     1,
			v2:       0,
			expected: locales.PluralRuleOne,
		},
		{
			num1:     0,
			v1:       0,
			num2:     2,
			v2:       0,
			expected: locales.PluralRuleFew,
		},
		{
			num1:     0,
			v1:       0,
			num2:     5,
			v2:       0,
			expected: locales.PluralRuleMany,
		},
		{
			num1:     0,
			v1:       0,
			num2:     10,
			v2:       1,
			expected: locales.PluralRuleOther,
		},
		{
			num1:     0.0,
			v1:       1,
			num2:     1,
			v2:       0,
			expected: locales.PluralRuleOne,
		},
		{
			num1:     0.0,
			v1:       1,
			num2:     2,
			v2:       0,
			expected: locales.PluralRuleFew,
		},
		{
			num1:     0.0,
			v1:       1,
			num2:     5,
			v2:       0,
			expected: locales.PluralRuleMany,
		},
		{
			num1:     0.0,
			v1:       1,
			num2:     10.0,
			v2:       1,
			expected: locales.PluralRuleOther,
		},
	}

	for _, tt := range tests {
		rule := trans.RangePluralRule(tt.num1, tt.v1, tt.num2, tt.v2)
		if rule != tt.expected {
			t.Errorf("Expected '%s' Got '%s'", tt.expected, rule)
		}
	}
}

func TestOrdinalPlurals(t *testing.T) {

	trans := New()

	tests := []struct {
		num      float64
		v        uint64
		expected locales.PluralRule
	}{
		{
			num:      1,
			v:        0,
			expected: locales.PluralRuleOther,
		},
		{
			num:      2,
			v:        0,
			expected: locales.PluralRuleOther,
		},
		{
			num:      3,
			v:        0,
			expected: locales.PluralRuleOther,
		},
		{
			num:      4,
			v:        0,
			expected: locales.PluralRuleOther,
		},
	}

	for _, tt := range tests {
		rule := trans.OrdinalPluralRule(tt.num, tt.v)
		if rule != tt.expected {
			t.Errorf("Expected '%s' Got '%s'", tt.expected, rule)
		}
	}
}

func TestCardinalPlurals(t *testing.T) {

	trans := New()

	tests := []struct {
		num      float64
		v        uint64
		expected locales.PluralRule
	}{
		{
			num:      1,
			v:        0,
			expected: locales.PluralRuleOne,
		},
		{
			num:      21,
			v:        0,
			expected: locales.PluralRuleOne,
		},
		{
			num:      31,
			v:        0,
			expected: locales.PluralRuleOne,
		},
		{
			num:      2,
			v:        0,
			expected: locales.PluralRuleFew,
		},
		{
			num:      3,
			v:        0,
			expected: locales.PluralRuleFew,
		},
		{
			num:      22,
			v:        0,
			expected: locales.PluralRuleFew,
		},
		{
			num:      23,
			v:        0,
			expected: locales.PluralRuleFew,
		},
		{
			num:      0,
			v:        0,
			expected: locales.PluralRuleMany,
		},
		{
			num:      5,
			v:        0,
			expected: locales.PluralRuleMany,
		},
		{
			num:      11,
			v:        0,
			expected: locales.PluralRuleMany,
		},
		{
			num:      100,
			v:        0,
			expected: locales.PluralRuleMany,
		},
	}

	for _, tt := range tests {
		rule := trans.CardinalPluralRule(tt.num, tt.v)
		if rule != tt.expected {
			t.Errorf("Expected '%s' Got '%s'", tt.expected, rule)
		}
	}
}

func TestDaysAbbreviated(t *testing.T) {

	trans := New()
	days := trans.WeekdaysAbbreviated()

	for i, day := range days {
		s := trans.WeekdayAbbreviated(time.Weekday(i))
		if s != day {
			t.Errorf("Expected '%s' Got '%s'", day, s)
		}
	}

	tests := []struct {
		idx      int
		expected string
	}{
		{
			idx:      0,
			expected: "вс",
		},
		{
			idx:      1,
			expected: "пн",
		},
		{
			idx:      2,
			expected: "вт",
		},
		{
			idx:      3,
			expected: "ср",
		},
		{
			idx:      4,
			expected: "чт",
		},
		{
			idx:      5,
			expected: "пт",
		},
		{
			idx:      6,
			expected: "сб",
		},
	}

	for _, tt := range tests {
		s := trans.WeekdayAbbreviated(time.Weekday(tt.idx))
		if s != tt.expected {
			t.Errorf("Expected '%s' Got '%s'", tt.expected, s)
		}
	}
}

func TestDaysNarrow(t *testing.T) {

	trans := New()
	days := trans.WeekdaysNarrow()

	for i, day := range days {
		s := trans.WeekdayNarrow(time.Weekday(i))
		if s != day {
			t.Errorf("Expected '%s' Got '%s'", string(day), s)
		}
	}

	tests := []struct {
		idx      int
		expected string
	}{
		{
			idx:      0,
			expected: "вс",
		},
		{
			idx:      1,
			expected: "пн",
		},
		{
			idx:      2,
			expected: "вт",
		},
		{
			idx:      3,
			expected: "ср",
		},
		{
			idx:      4,
			expected: "чт",
		},
		{
			idx:      5,
			expected: "пт",
		},
		{
			idx:      6,
			expected: "сб",
		},
	}

	for _, tt := range tests {
		s := trans.WeekdayNarrow(time.Weekday(tt.idx))
		if s != tt.expected {
			t.Errorf("Expected '%s' Got '%s'", tt.expected, s)
		}
	}
}

func TestDaysShort(t *testing.T) {

	trans := New()
	days := trans.WeekdaysShort()

	for i, day := range days {
		s := trans.WeekdayShort(time.Weekday(i))
		if s != day {
			t.Errorf("Expected '%s' Got '%s'", day, s)
		}
	}

	tests := []struct {
		idx      int
		expected string
	}{
		{
			idx:      0,
			expected: "вс",
		},
		{
			idx:      1,
			expected: "пн",
		},
		{
			idx:      2,
			expected: "вт",
		},
		{
			idx:      3,
			expected: "ср",
		},
		{
			idx:      4,
			expected: "чт",
		},
		{
			idx:      5,
			expected: "пт",
		},
		{
			idx:      6,
			expected: "сб",
		},
	}

	for _, tt := range tests {
		s := trans.WeekdayShort(time.Weekday(tt.idx))
		if s != tt.expected {
			t.Errorf("Expected '%s' Got '%s'", tt.expected, s)
		}
	}
}

func TestDaysWide(t *testing.T) {

	trans := New()
	days := trans.WeekdaysWide()

	for i, day := range days {
		s := trans.WeekdayWide(time.Weekday(i))
		if s != day {
			t.Errorf("Expected '%s' Got '%s'", day, s)
		}
	}

	tests := []struct {
		idx      int
		expected string
	}{
		{
			idx:      0,
			expected: "воскресенье",
		},
		{
			idx:      1,
			expected: "понедельник",
		},
		{
			idx:      2,
			expected: "вторник",
		},
		{
			idx:      3,
			expected: "среда",
		},
		{
			idx:      4,
			expected: "четверг",
		},
		{
			idx:      5,
			expected: "пятница",
		},
		{
			idx:      6,
			expected: "суббота",
		},
	}

	for _, tt := range tests {
		s := trans.WeekdayWide(time.Weekday(tt.idx))
		if s != tt.expected {
			t.Errorf("Expected '%s' Got '%s'", tt.expected, s)
		}
	}
}

func TestMonthsAbbreviated(t *testing.T) {

	trans := New()
	months := trans.MonthsAbbreviated()

	for i, month := range months {
		s := trans.MonthAbbreviated(time.Month(i + 1))
		if s != month {
			t.Errorf("Expected '%s' Got '%s'", month, s)
		}
	}

	tests := []struct {
		idx      int
		expected string
	}{
		{
			idx:      1,
			expected: "янв.",
		},
		{
			idx:      2,
			expected: "февр.",
		},
		{
			idx:      3,
			expected: "мар.",
		},
		{
			idx:      4,
			expected: "апр.",
		},
		{
			idx:      5,
			expected: "мая",
		},
		{
			idx:      6,
			expected: "июн.",
		},
		{
			idx:      7,
			expected: "июл.",
		},
		{
			idx:      8,
			expected: "авг.",
		},
		{
			idx:      9,
			expected: "сент.",
		},
		{
			idx:      10,
			expected: "окт.",
		},
		{
			idx:      11,
			expected: "нояб.",
		},
		{
			idx:      12,
			expected: "дек.",
		},
	}

	for _, tt := range tests {
		s := trans.MonthAbbreviated(time.Month(tt.idx))
		if s != tt.expected {
			t.Errorf("Expected '%s' Got '%s'", tt.expected, s)
		}
	}
}

func TestMonthsNarrow(t *testing.T) {

	trans := New()
	months := trans.MonthsNarrow()

	for i, month := range months {
		s := trans.MonthNarrow(time.Month(i + 1))
		if s != month {
			t.Errorf("Expected '%s' Got '%s'", month, s)
		}
	}

	tests := []struct {
		idx      int
		expected string
	}{
		{
			idx:      1,
			expected: "Я",
		},
		{
			idx:      2,
			expected: "Ф",
		},
		{
			idx:      3,
			expected: "М",
		},
		{
			idx:      4,
			expected: "А",
		},
		{
			idx:      5,
			expected: "М",
		},
		{
			idx:      6,
			expected: "И",
		},
		{
			idx:      7,
			expected: "И",
		},
		{
			idx:      8,
			expected: "А",
		},
		{
			idx:      9,
			expected: "С",
		},
		{
			idx:      10,
			expected: "О",
		},
		{
			idx:      11,
			expected: "Н",
		},
		{
			idx:      12,
			expected: "Д",
		},
	}

	for _, tt := range tests {
		s := trans.MonthNarrow(time.Month(tt.idx))
		if s != tt.expected {
			t.Errorf("Expected '%s' Got '%s'", tt.expected, s)
		}
	}
}

func TestMonthsWide(t *testing.T) {

	trans := New()
	months := trans.MonthsWide()

	for i, month := range months {
		s := trans.MonthWide(time.Month(i + 1))
		if s != month {
			t.Errorf("Expected '%s' Got '%s'", month, s)
		}
	}

	tests := []struct {
		idx      int
		expected string
	}{
		{
			idx:      1,
			expected: "января",
		},
		{
			idx:      2,
			expected: "февраля",
		},
		{
			idx:      3,
			expected: "марта",
		},
		{
			idx:      4,
			expected: "апреля",
		},
		{
			idx:      5,
			expected: "мая",
		},
		{
			idx:      6,
			expected: "июня",
		},
		{
			idx:      7,
			expected: "июля",
		},
		{
			idx:      8,
			expected: "августа",
		},
		{
			idx:      9,
			expected: "сентября",
		},
		{
			idx:      10,
			expected: "октября",
		},
		{
			idx:      11,
			expected: "ноября",
		},
		{
			idx:      12,
			expected: "декабря",
		},
	}

	for _, tt := range tests {
		s := string(trans.MonthWide(time.Month(tt.idx)))
		if s != tt.expected {
			t.Errorf("Expected '%s' Got '%s'", tt.expected, s)
		}
	}
}

func TestFmtTimeFull(t *testing.T) {

	loc, err := time.LoadLocation("America/Toronto")
	if err != nil {
		t.Errorf("Expected '<nil>' Got '%s'", err)
	}

	fixed := time.FixedZone("OTHER", -4)

	tests := []struct {
		t        time.Time
		expected string
	}{
		{
			t:        time.Date(2016, 02, 03, 9, 5, 1, 0, loc),
			expected: "9:05:01 Восточная Америка, стандартное время",
		},
		{
			t:        time.Date(2016, 02, 03, 20, 5, 1, 0, fixed),
			expected: "20:05:01 OTHER",
		},
	}

	trans := New()

	for _, tt := range tests {
		s := trans.FmtTimeFull(tt.t)
		if s != tt.expected {
			t.Errorf("Expected '%s' Got '%s'", tt.expected, s)
		}
	}
}

func TestFmtTimeLong(t *testing.T) {

	loc, err := time.LoadLocation("America/Toronto")
	if err != nil {
		t.Errorf("Expected '<nil>' Got '%s'", err)
	}

	tests := []struct {
		t        time.Time
		expected string
	}{
		{
			t:        time.Date(2016, 02, 03, 9, 5, 1, 0, loc),
			expected: "9:05:01 EST",
		},
		{
			t:        time.Date(2016, 02, 03, 20, 5, 1, 0, loc),
			expected: "20:05:01 EST",
		},
	}

	trans := New()

	for _, tt := range tests {
		s := trans.FmtTimeLong(tt.t)
		if s != tt.expected {
			t.Errorf("Expected '%s' Got '%s'", tt.expected, s)
		}
	}
}

func TestFmtTimeMedium(t *testing.T) {

	tests := []struct {
		t        time.Time
		expected string
	}{
		{
			t:        time.Date(2016, 02, 03, 9, 5, 1, 0, time.UTC),
			expected: "9:05:01",
		},
		{
			t:        time.Date(2016, 02, 03, 20, 5, 1, 0, time.UTC),
			expected: "20:05:01",
		},
	}

	trans := New()

	for _, tt := range tests {
		s := trans.FmtTimeMedium(tt.t)
		if s != tt.expected {
			t.Errorf("Expected '%s' Got '%s'", tt.expected, s)
		}
	}
}

func TestFmtTimeShort(t *testing.T) {

	tests := []struct {
		t        time.Time
		expected string
	}{
		{
			t:        time.Date(2016, 02, 03, 9, 5, 1, 0, time.UTC),
			expected: "9:05",
		},
		{
			t:        time.Date(2016, 02, 03, 20, 5, 1, 0, time.UTC),
			expected: "20:05",
		},
	}

	trans := New()

	for _, tt := range tests {
		s := trans.FmtTimeShort(tt.t)
		if s != tt.expected {
			t.Errorf("Expected '%s' Got '%s'", tt.expected, s)
		}
	}
}

func TestFmtDateFull(t *testing.T) {

	tests := []struct {
		t        time.Time
		expected string
	}{
		{
			t:        time.Date(2016, 02, 03, 9, 0, 1, 0, time.UTC),
			expected: "среда, 3 февраля 2016 г.",
		},
	}

	trans := New()

	for _, tt := range tests {
		s := trans.FmtDateFull(tt.t)
		if s != tt.expected {
			t.Errorf("Expected '%s' Got '%s'", tt.expected, s)
		}
	}
}

func TestFmtDateLong(t *testing.T) {

	tests := []struct {
		t        time.Time
		expected string
	}{
		{
			t:        time.Date(2016, 02, 03, 9, 0, 1, 0, time.UTC),
			expected: "3 февраля 2016 г.",
		},
	}

	trans := New()

	for _, tt := range tests {
		s := trans.FmtDateLong(tt.t)
		if s != tt.expected {
			t.Errorf("Expected '%s' Got '%s'", tt.expected, s)
		}
	}
}

func TestFmtDateMedium(t *testing.T) {

	tests := []struct {
		t        time.Time
		expected string
	}{
		{
			t:        time.Date(2016, 02, 03, 9, 0, 1, 0, time.UTC),
			expected: "3 февр. 2016 г.",
		},
	}

	trans := New()

	for _, tt := range tests {
		s := trans.FmtDateMedium(tt.t)
		if s != tt.expected {
			t.Errorf("Expected '%s' Got '%s'", tt.expected, s)
		}
	}
}

func TestFmtDateShort(t *testing.T) {

	tests := []struct {
		t        time.Time
		expected string
	}{
		{
			t:        time.Date(2016, 02, 03, 9, 0, 1, 0, time.UTC),
			expected: "03.02.2016", // date format changed from v29 dd.MM.yy to v30 dd.MM.y so adjusted test for new CLDR data
		},
		{
			t:        time.Date(-500, 02, 03, 9, 0, 1, 0, time.UTC),
			expected: "03.02.500",
		},
	}

	trans := New()

	for _, tt := range tests {
		s := trans.FmtDateShort(tt.t)
		if s != tt.expected {
			t.Errorf("Expected '%s' Got '%s'", tt.expected, s)
		}
	}
}

func TestFmtNumber(t *testing.T) {

	tests := []struct {
		num      float64
		v        uint64
		expected string
	}{
		{
			num:      1123456.5643,
			v:        2,
			expected: "1 123 456,56",
		},
		{
			num:      1123456.5643,
			v:        1,
			expected: "1 123 456,6",
		},
		{
			num:      221123456.5643,
			v:        3,
			expected: "221 123 456,564",
		},
		{
			num:      -221123456.5643,
			v:        3,
			expected: "-221 123 456,564",
		},
		{
			num:      -221123456.5643,
			v:        3,
			expected: "-221 123 456,564",
		},
		{
			num:      0,
			v:        2,
			expected: "0,00",
		},
		{
			num:      -0,
			v:        2,
			expected: "0,00",
		},
		{
			num:      -0,
			v:        2,
			expected: "0,00",
		},
	}

	trans := New()

	for _, tt := range tests {
		s := trans.FmtNumber(tt.num, tt.v)
		if s != tt.expected {
			t.Errorf("Expected '%s' Got '%s'", tt.expected, s)
		}
	}
}

func TestFmtCurrency(t *testing.T) {

	tests := []struct {
		num      float64
		v        uint64
		currency currency.Type
		expected string
	}{
		{
			num:      1123456.5643,
			v:        2,
			currency: currency.USD,
			expected: "1 123 456,56 USD",
		},
		{
			num:      1123456.5643,
			v:        1,
			currency: currency.USD,
			expected: "1 123 456,60 USD",
		},
		{
			num:      221123456.5643,
			v:        3,
			currency: currency.USD,
			expected: "221 123 456,564 USD",
		},
		{
			num:      -221123456.5643,
			v:        3,
			currency: currency.USD,
			expected: "-221 123 456,564 USD",
		},
		{
			num:      -221123456.5643,
			v:        3,
			currency: currency.CAD,
			expected: "-221 123 456,564 CAD",
		},
		{
			num:      0,
			v:        2,
			currency: currency.USD,
			expected: "0,00 USD",
		},
		{
			num:      -0,
			v:        2,
			currency: currency.USD,
			expected: "0,00 USD",
		},
		{
			num:      -0,
			v:        2,
			currency: currency.CAD,
			expected: "0,00 CAD",
		},
		{
			num:      1.23,
			v:        0,
			currency: currency.USD,
			expected: "1,00 USD",
		},
	}

	trans := New()

	for _, tt := range tests {
		s := trans.FmtCurrency(tt.num, tt.v, tt.currency)
		if s != tt.expected {
			t.Errorf("Expected '%s' Got '%s'", tt.expected, s)
		}
	}
}

func TestFmtAccounting(t *testing.T) {

	tests := []struct {
		num      float64
		v        uint64
		currency currency.Type
		expected string
	}{
		{
			num:      1123456.5643,
			v:        2,
			currency: currency.USD,
			expected: "1 123 456,56 USD",
		},
		{
			num:      1123456.5643,
			v:        1,
			currency: currency.USD,
			expected: "1 123 456,60 USD",
		},
		{
			num:      221123456.5643,
			v:        3,
			currency: currency.USD,
			expected: "221 123 456,564 USD",
		},
		{
			num:      -221123456.5643,
			v:        3,
			currency: currency.USD,
			expected: "-221 123 456,564 USD",
		},
		{
			num:      -221123456.5643,
			v:        3,
			currency: currency.CAD,
			expected: "-221 123 456,564 CAD",
		},
		{
			num:      -0,
			v:        2,
			currency: currency.USD,
			expected: "0,00 USD",
		},
		{
			num:      -0,
			v:        2,
			currency: currency.CAD,
			expected: "0,00 CAD",
		},
		{
			num:      1.23,
			v:        0,
			currency: currency.USD,
			expected: "1,00 USD",
		},
	}

	trans := New()

	for _, tt := range tests {
		s := trans.FmtAccounting(tt.num, tt.v, tt.currency)
		if s != tt.expected {
			t.Errorf("Expected '%s' Got '%s'", tt.expected, s)
		}
	}
}

func TestFmtPercent(t *testing.T) {

	tests := []struct {
		num      float64
		v        uint64
		expected string
	}{
		{
			num:      15,
			v:        0,
			expected: "15%",
		},
		{
			num:      15,
			v:        2,
			expected: "15,00%",
		},
		{
			num:      434.45,
			v:        0,
			expected: "434%",
		},
		{
			num:      34.4,
			v:        2,
			expected: "34,40%",
		},
		{
			num:      -34,
			v:        0,
			expected: "-34%",
		},
	}

	trans := New()

	for _, tt := range tests {
		s := trans.FmtPercent(tt.num, tt.v)
		if s != tt.expected {
			t.Errorf("Expected '%s' Got '%s'", tt.expected, s)
		}
	}
}
