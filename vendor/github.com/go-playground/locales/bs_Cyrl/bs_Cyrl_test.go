package bs_Cyrl

import (
	"testing"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

func TestLocale(t *testing.T) {

	trans := New()
	expected := "bs_Cyrl"

	if trans.Locale() != expected {
		t.Errorf("Expected '%s' Got '%s'", expected, trans.Locale())
	}
}

func TestPluralsRange(t *testing.T) {

	trans := New()

	tests := []struct {
		expected locales.PluralRule
	}{
	// {
	// 	expected: locales.PluralRuleOther,
	// },
	}

	rules := trans.PluralsRange()
	// expected := 1
	// if len(rules) != expected {
	// 	t.Errorf("Expected '%d' Got '%d'", expected, len(rules))
	// }

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
	// {
	// 	expected: locales.PluralRuleOne,
	// },
	// {
	// 	expected: locales.PluralRuleTwo,
	// },
	// {
	// 	expected: locales.PluralRuleFew,
	// },
	// {
	// 	expected: locales.PluralRuleOther,
	// },
	}

	rules := trans.PluralsOrdinal()
	// expected := 4
	// if len(rules) != expected {
	// 	t.Errorf("Expected '%d' Got '%d'", expected, len(rules))
	// }

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
	// {
	// 	expected: locales.PluralRuleOne,
	// },
	// {
	// 	expected: locales.PluralRuleOther,
	// },
	}

	rules := trans.PluralsCardinal()
	// expected := 2
	// if len(rules) != expected {
	// 	t.Errorf("Expected '%d' Got '%d'", expected, len(rules))
	// }

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
	// {
	// 	num1:     1,
	// 	v1:       1,
	// 	num2:     2,
	// 	v2:       2,
	// 	expected: locales.PluralRuleOther,
	// },
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
	// {
	// 	num:      1,
	// 	v:        0,
	// 	expected: locales.PluralRuleOne,
	// },
	// {
	// 	num:      2,
	// 	v:        0,
	// 	expected: locales.PluralRuleTwo,
	// },
	// {
	// 	num:      3,
	// 	v:        0,
	// 	expected: locales.PluralRuleFew,
	// },
	// {
	// 	num:      4,
	// 	v:        0,
	// 	expected: locales.PluralRuleOther,
	// },
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
	// {
	// 	num:      1,
	// 	v:        0,
	// 	expected: locales.PluralRuleOne,
	// },
	// {
	// 	num:      4,
	// 	v:        0,
	// 	expected: locales.PluralRuleOther,
	// },
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
	// {
	// 	idx:      0,
	// 	expected: "Sun",
	// },
	// {
	// 	idx:      1,
	// 	expected: "Mon",
	// },
	// {
	// 	idx:      2,
	// 	expected: "Tue",
	// },
	// {
	// 	idx:      3,
	// 	expected: "Wed",
	// },
	// {
	// 	idx:      4,
	// 	expected: "Thu",
	// },
	// {
	// 	idx:      5,
	// 	expected: "Fri",
	// },
	// {
	// 	idx:      6,
	// 	expected: "Sat",
	// },
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
	// {
	// 	idx:      0,
	// 	expected: "S",
	// },
	// {
	// 	idx:      1,
	// 	expected: "M",
	// },
	// {
	// 	idx:      2,
	// 	expected: "T",
	// },
	// {
	// 	idx:      3,
	// 	expected: "W",
	// },
	// {
	// 	idx:      4,
	// 	expected: "T",
	// },
	// {
	// 	idx:      5,
	// 	expected: "F",
	// },
	// {
	// 	idx:      6,
	// 	expected: "S",
	// },
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
	// {
	// 	idx:      0,
	// 	expected: "Su",
	// },
	// {
	// 	idx:      1,
	// 	expected: "Mo",
	// },
	// {
	// 	idx:      2,
	// 	expected: "Tu",
	// },
	// {
	// 	idx:      3,
	// 	expected: "We",
	// },
	// {
	// 	idx:      4,
	// 	expected: "Th",
	// },
	// {
	// 	idx:      5,
	// 	expected: "Fr",
	// },
	// {
	// 	idx:      6,
	// 	expected: "Sa",
	// },
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
	// {
	// 	idx:      0,
	// 	expected: "Sunday",
	// },
	// {
	// 	idx:      1,
	// 	expected: "Monday",
	// },
	// {
	// 	idx:      2,
	// 	expected: "Tuesday",
	// },
	// {
	// 	idx:      3,
	// 	expected: "Wednesday",
	// },
	// {
	// 	idx:      4,
	// 	expected: "Thursday",
	// },
	// {
	// 	idx:      5,
	// 	expected: "Friday",
	// },
	// {
	// 	idx:      6,
	// 	expected: "Saturday",
	// },
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
	// {
	// 	idx:      1,
	// 	expected: "Jan",
	// },
	// {
	// 	idx:      2,
	// 	expected: "Feb",
	// },
	// {
	// 	idx:      3,
	// 	expected: "Mar",
	// },
	// {
	// 	idx:      4,
	// 	expected: "Apr",
	// },
	// {
	// 	idx:      5,
	// 	expected: "May",
	// },
	// {
	// 	idx:      6,
	// 	expected: "Jun",
	// },
	// {
	// 	idx:      7,
	// 	expected: "Jul",
	// },
	// {
	// 	idx:      8,
	// 	expected: "Aug",
	// },
	// {
	// 	idx:      9,
	// 	expected: "Sep",
	// },
	// {
	// 	idx:      10,
	// 	expected: "Oct",
	// },
	// {
	// 	idx:      11,
	// 	expected: "Nov",
	// },
	// {
	// 	idx:      12,
	// 	expected: "Dec",
	// },
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
	// {
	// 	idx:      1,
	// 	expected: "J",
	// },
	// {
	// 	idx:      2,
	// 	expected: "F",
	// },
	// {
	// 	idx:      3,
	// 	expected: "M",
	// },
	// {
	// 	idx:      4,
	// 	expected: "A",
	// },
	// {
	// 	idx:      5,
	// 	expected: "M",
	// },
	// {
	// 	idx:      6,
	// 	expected: "J",
	// },
	// {
	// 	idx:      7,
	// 	expected: "J",
	// },
	// {
	// 	idx:      8,
	// 	expected: "A",
	// },
	// {
	// 	idx:      9,
	// 	expected: "S",
	// },
	// {
	// 	idx:      10,
	// 	expected: "O",
	// },
	// {
	// 	idx:      11,
	// 	expected: "N",
	// },
	// {
	// 	idx:      12,
	// 	expected: "D",
	// },
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
	// {
	// 	idx:      1,
	// 	expected: "January",
	// },
	// {
	// 	idx:      2,
	// 	expected: "February",
	// },
	// {
	// 	idx:      3,
	// 	expected: "March",
	// },
	// {
	// 	idx:      4,
	// 	expected: "April",
	// },
	// {
	// 	idx:      5,
	// 	expected: "May",
	// },
	// {
	// 	idx:      6,
	// 	expected: "June",
	// },
	// {
	// 	idx:      7,
	// 	expected: "July",
	// },
	// {
	// 	idx:      8,
	// 	expected: "August",
	// },
	// {
	// 	idx:      9,
	// 	expected: "September",
	// },
	// {
	// 	idx:      10,
	// 	expected: "October",
	// },
	// {
	// 	idx:      11,
	// 	expected: "November",
	// },
	// {
	// 	idx:      12,
	// 	expected: "December",
	// },
	}

	for _, tt := range tests {
		s := string(trans.MonthWide(time.Month(tt.idx)))
		if s != tt.expected {
			t.Errorf("Expected '%s' Got '%s'", tt.expected, s)
		}
	}
}

func TestFmtTimeFull(t *testing.T) {

	// loc, err := time.LoadLocation("America/Toronto")
	// if err != nil {
	// 	t.Errorf("Expected '<nil>' Got '%s'", err)
	// }

	// fixed := time.FixedZone("OTHER", -4)

	tests := []struct {
		t        time.Time
		expected string
	}{
	// {
	// 	t:        time.Date(2016, 02, 03, 9, 5, 1, 0, loc),
	// 	expected: "9:05:01 am Eastern Standard Time",
	// },
	// {
	// 	t:        time.Date(2016, 02, 03, 20, 5, 1, 0, fixed),
	// 	expected: "8:05:01 pm OTHER",
	// },
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

	// loc, err := time.LoadLocation("America/Toronto")
	// if err != nil {
	// 	t.Errorf("Expected '<nil>' Got '%s'", err)
	// }

	tests := []struct {
		t        time.Time
		expected string
	}{
	// {
	// 	t:        time.Date(2016, 02, 03, 9, 5, 1, 0, loc),
	// 	expected: "9:05:01 am EST",
	// },
	// {
	// 	t:        time.Date(2016, 02, 03, 20, 5, 1, 0, loc),
	// 	expected: "8:05:01 pm EST",
	// },
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
	// {
	// 	t:        time.Date(2016, 02, 03, 9, 5, 1, 0, time.UTC),
	// 	expected: "9:05:01 am",
	// },
	// {
	// 	t:        time.Date(2016, 02, 03, 20, 5, 1, 0, time.UTC),
	// 	expected: "8:05:01 pm",
	// },
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
	// {
	// 	t:        time.Date(2016, 02, 03, 9, 5, 1, 0, time.UTC),
	// 	expected: "9:05 am",
	// },
	// {
	// 	t:        time.Date(2016, 02, 03, 20, 5, 1, 0, time.UTC),
	// 	expected: "8:05 pm",
	// },
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
	// {
	// 	t:        time.Date(2016, 02, 03, 9, 0, 1, 0, time.UTC),
	// 	expected: "Wednesday, February 3, 2016",
	// },
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
	// {
	// 	t:        time.Date(2016, 02, 03, 9, 0, 1, 0, time.UTC),
	// 	expected: "February 3, 2016",
	// },
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
	// {
	// 	t:        time.Date(2016, 02, 03, 9, 0, 1, 0, time.UTC),
	// 	expected: "Feb 3, 2016",
	// },
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
	// {
	// 	t:        time.Date(2016, 02, 03, 9, 0, 1, 0, time.UTC),
	// 	expected: "2/3/16",
	// },
	// {
	// 	t:        time.Date(-500, 02, 03, 9, 0, 1, 0, time.UTC),
	// 	expected: "2/3/500",
	// },
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
	// {
	// 	num:      1123456.5643,
	// 	v:        2,
	// 	expected: "1,123,456.56",
	// },
	// {
	// 	num:      1123456.5643,
	// 	v:        1,
	// 	expected: "1,123,456.6",
	// },
	// {
	// 	num:      221123456.5643,
	// 	v:        3,
	// 	expected: "221,123,456.564",
	// },
	// {
	// 	num:      -221123456.5643,
	// 	v:        3,
	// 	expected: "-221,123,456.564",
	// },
	// {
	// 	num:      -221123456.5643,
	// 	v:        3,
	// 	expected: "-221,123,456.564",
	// },
	// {
	// 	num:      0,
	// 	v:        2,
	// 	expected: "0.00",
	// },
	// {
	// 	num:      -0,
	// 	v:        2,
	// 	expected: "0.00",
	// },
	// {
	// 	num:      -0,
	// 	v:        2,
	// 	expected: "0.00",
	// },
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
	// {
	// 	num:      1123456.5643,
	// 	v:        2,
	// 	currency: currency.USD,
	// 	expected: "$1,123,456.56",
	// },
	// {
	// 	num:      1123456.5643,
	// 	v:        1,
	// 	currency: currency.USD,
	// 	expected: "$1,123,456.60",
	// },
	// {
	// 	num:      221123456.5643,
	// 	v:        3,
	// 	currency: currency.USD,
	// 	expected: "$221,123,456.564",
	// },
	// {
	// 	num:      -221123456.5643,
	// 	v:        3,
	// 	currency: currency.USD,
	// 	expected: "-$221,123,456.564",
	// },
	// {
	// 	num:      -221123456.5643,
	// 	v:        3,
	// 	currency: currency.CAD,
	// 	expected: "-CAD 221,123,456.564",
	// },
	// {
	// 	num:      0,
	// 	v:        2,
	// 	currency: currency.USD,
	// 	expected: "$0.00",
	// },
	// {
	// 	num:      -0,
	// 	v:        2,
	// 	currency: currency.USD,
	// 	expected: "$0.00",
	// },
	// {
	// 	num:      -0,
	// 	v:        2,
	// 	currency: currency.CAD,
	// 	expected: "CAD 0.00",
	// },
	// {
	// 	num:      1.23,
	// 	v:        0,
	// 	currency: currency.USD,
	// 	expected: "$1.00",
	// },
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
	// {
	// 	num:      1123456.5643,
	// 	v:        2,
	// 	currency: currency.USD,
	// 	expected: "$1,123,456.56",
	// },
	// {
	// 	num:      1123456.5643,
	// 	v:        1,
	// 	currency: currency.USD,
	// 	expected: "$1,123,456.60",
	// },
	// {
	// 	num:      221123456.5643,
	// 	v:        3,
	// 	currency: currency.USD,
	// 	expected: "$221,123,456.564",
	// },
	// {
	// 	num:      -221123456.5643,
	// 	v:        3,
	// 	currency: currency.USD,
	// 	expected: "($221,123,456.564)",
	// },
	// {
	// 	num:      -221123456.5643,
	// 	v:        3,
	// 	currency: currency.CAD,
	// 	expected: "(CAD 221,123,456.564)",
	// },
	// {
	// 	num:      -0,
	// 	v:        2,
	// 	currency: currency.USD,
	// 	expected: "$0.00",
	// },
	// {
	// 	num:      -0,
	// 	v:        2,
	// 	currency: currency.CAD,
	// 	expected: "CAD 0.00",
	// },
	// {
	// 	num:      1.23,
	// 	v:        0,
	// 	currency: currency.USD,
	// 	expected: "$1.00",
	// },
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
	// {
	// 	num:      15,
	// 	v:        0,
	// 	expected: "15%",
	// },
	// {
	// 	num:      15,
	// 	v:        2,
	// 	expected: "15.00%",
	// },
	// {
	// 	num:      434.45,
	// 	v:        0,
	// 	expected: "434%",
	// },
	// {
	// 	num:      34.4,
	// 	v:        2,
	// 	expected: "34.40%",
	// },
	// {
	// 	num:      -34,
	// 	v:        0,
	// 	expected: "-34%",
	// },
	}

	trans := New()

	for _, tt := range tests {
		s := trans.FmtPercent(tt.num, tt.v)
		if s != tt.expected {
			t.Errorf("Expected '%s' Got '%s'", tt.expected, s)
		}
	}
}
