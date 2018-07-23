package rw_RW

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type rw_RW struct {
	locale                 string
	pluralsCardinal        []locales.PluralRule
	pluralsOrdinal         []locales.PluralRule
	pluralsRange           []locales.PluralRule
	decimal                string
	group                  string
	minus                  string
	percent                string
	perMille               string
	timeSeparator          string
	inifinity              string
	currencies             []string // idx = enum of currency code
	currencyPositivePrefix string
	currencyNegativePrefix string
	monthsAbbreviated      []string
	monthsNarrow           []string
	monthsWide             []string
	daysAbbreviated        []string
	daysNarrow             []string
	daysShort              []string
	daysWide               []string
	periodsAbbreviated     []string
	periodsNarrow          []string
	periodsShort           []string
	periodsWide            []string
	erasAbbreviated        []string
	erasNarrow             []string
	erasWide               []string
	timezones              map[string]string
}

// New returns a new instance of translator for the 'rw_RW' locale
func New() locales.Translator {
	return &rw_RW{
		locale:                 "rw_RW",
		pluralsCardinal:        nil,
		pluralsOrdinal:         nil,
		pluralsRange:           nil,
		decimal:                ",",
		group:                  ".",
		minus:                  "-",
		percent:                "%",
		perMille:               "‰",
		timeSeparator:          ":",
		inifinity:              "∞",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyPositivePrefix: " ",
		currencyNegativePrefix: " ",
		monthsAbbreviated:      []string{"", "mut.", "gas.", "wer.", "mat.", "gic.", "kam.", "nya.", "kan.", "nze.", "ukw.", "ugu.", "uku."},
		monthsNarrow:           []string{"", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12"},
		monthsWide:             []string{"", "Mutarama", "Gashyantare", "Werurwe", "Mata", "Gicuransi", "Kamena", "Nyakanga", "Kanama", "Nzeli", "Ukwakira", "Ugushyingo", "Ukuboza"},
		daysAbbreviated:        []string{"cyu.", "mbe.", "kab.", "gtu.", "kan.", "gnu.", "gnd."},
		daysNarrow:             []string{"S", "M", "T", "W", "T", "F", "S"},
		daysShort:              []string{"cyu.", "mbe.", "kab.", "gtu.", "kan.", "gnu.", "gnd."},
		daysWide:               []string{"Ku cyumweru", "Kuwa mbere", "Kuwa kabiri", "Kuwa gatatu", "Kuwa kane", "Kuwa gatanu", "Kuwa gatandatu"},
		periodsAbbreviated:     []string{"AM", "PM"},
		periodsNarrow:          []string{"AM", "PM"},
		periodsWide:            []string{"AM", "PM"},
		erasAbbreviated:        []string{"BCE", "CE"},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"BCE", "CE"},
		timezones:              map[string]string{"OESZ": "OESZ", "COT": "COT", "AKST": "AKST", "HEOG": "HEOG", "HADT": "HADT", "ART": "ART", "CDT": "CDT", "BOT": "BOT", "MESZ": "MESZ", "WART": "WART", "WARST": "WARST", "HNNOMX": "HNNOMX", "AWST": "AWST", "HNPM": "HNPM", "UYST": "UYST", "CHAST": "CHAST", "HNOG": "HNOG", "EDT": "EDT", "WIB": "WIB", "SRT": "SRT", "TMT": "TMT", "GMT": "GMT", "GYT": "GYT", "HECU": "HECU", "WAT": "WAT", "HNEG": "HNEG", "HEPM": "HEPM", "WITA": "WITA", "ARST": "ARST", "HNCU": "HNCU", "AEST": "AEST", "SGT": "SGT", "CST": "CST", "WESZ": "WESZ", "GFT": "GFT", "ACST": "ACST", "HKT": "HKT", "HNT": "HNT", "TMST": "TMST", "EAT": "EAT", "AWDT": "AWDT", "JST": "JST", "HAT": "HAT", "OEZ": "OEZ", "COST": "COST", "AST": "AST", "WEZ": "WEZ", "IST": "IST", "ChST": "ChST", "HEPMX": "HEPMX", "NZST": "NZST", "MYT": "MYT", "ACDT": "ACDT", "NZDT": "NZDT", "AKDT": "AKDT", "MEZ": "MEZ", "LHDT": "LHDT", "CLT": "CLT", "WIT": "WIT", "PST": "PST", "BT": "BT", "ACWST": "ACWST", "VET": "VET", "CHADT": "CHADT", "HNPMX": "HNPMX", "ADT": "ADT", "SAST": "SAST", "∅∅∅": "∅∅∅", "UYT": "UYT", "AEDT": "AEDT", "WAST": "WAST", "ACWDT": "ACWDT", "HEEG": "HEEG", "HKST": "HKST", "LHST": "LHST", "HENOMX": "HENOMX", "MDT": "MDT", "JDT": "JDT", "ECT": "ECT", "MST": "MST", "CLST": "CLST", "CAT": "CAT", "HAST": "HAST", "PDT": "PDT", "EST": "EST"},
	}
}

// Locale returns the current translators string locale
func (rw *rw_RW) Locale() string {
	return rw.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'rw_RW'
func (rw *rw_RW) PluralsCardinal() []locales.PluralRule {
	return rw.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'rw_RW'
func (rw *rw_RW) PluralsOrdinal() []locales.PluralRule {
	return rw.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'rw_RW'
func (rw *rw_RW) PluralsRange() []locales.PluralRule {
	return rw.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'rw_RW'
func (rw *rw_RW) CardinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'rw_RW'
func (rw *rw_RW) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'rw_RW'
func (rw *rw_RW) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (rw *rw_RW) MonthAbbreviated(month time.Month) string {
	return rw.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (rw *rw_RW) MonthsAbbreviated() []string {
	return rw.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (rw *rw_RW) MonthNarrow(month time.Month) string {
	return rw.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (rw *rw_RW) MonthsNarrow() []string {
	return rw.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (rw *rw_RW) MonthWide(month time.Month) string {
	return rw.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (rw *rw_RW) MonthsWide() []string {
	return rw.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (rw *rw_RW) WeekdayAbbreviated(weekday time.Weekday) string {
	return rw.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (rw *rw_RW) WeekdaysAbbreviated() []string {
	return rw.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (rw *rw_RW) WeekdayNarrow(weekday time.Weekday) string {
	return rw.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (rw *rw_RW) WeekdaysNarrow() []string {
	return rw.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (rw *rw_RW) WeekdayShort(weekday time.Weekday) string {
	return rw.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (rw *rw_RW) WeekdaysShort() []string {
	return rw.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (rw *rw_RW) WeekdayWide(weekday time.Weekday) string {
	return rw.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (rw *rw_RW) WeekdaysWide() []string {
	return rw.daysWide
}

// Decimal returns the decimal point of number
func (rw *rw_RW) Decimal() string {
	return rw.decimal
}

// Group returns the group of number
func (rw *rw_RW) Group() string {
	return rw.group
}

// Group returns the minus sign of number
func (rw *rw_RW) Minus() string {
	return rw.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'rw_RW' and handles both Whole and Real numbers based on 'v'
func (rw *rw_RW) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, rw.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, rw.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, rw.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'rw_RW' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (rw *rw_RW) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 3
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, rw.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, rw.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, rw.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'rw_RW'
func (rw *rw_RW) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := rw.currencies[currency]
	l := len(s) + len(symbol) + 4 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, rw.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, rw.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	for j := len(symbol) - 1; j >= 0; j-- {
		b = append(b, symbol[j])
	}

	for j := len(rw.currencyPositivePrefix) - 1; j >= 0; j-- {
		b = append(b, rw.currencyPositivePrefix[j])
	}

	if num < 0 {
		b = append(b, rw.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, rw.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'rw_RW'
// in accounting notation.
func (rw *rw_RW) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := rw.currencies[currency]
	l := len(s) + len(symbol) + 4 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, rw.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, rw.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {

		for j := len(symbol) - 1; j >= 0; j-- {
			b = append(b, symbol[j])
		}

		for j := len(rw.currencyNegativePrefix) - 1; j >= 0; j-- {
			b = append(b, rw.currencyNegativePrefix[j])
		}

		b = append(b, rw.minus[0])

	} else {

		for j := len(symbol) - 1; j >= 0; j-- {
			b = append(b, symbol[j])
		}

		for j := len(rw.currencyPositivePrefix) - 1; j >= 0; j-- {
			b = append(b, rw.currencyPositivePrefix[j])
		}

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, rw.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'rw_RW'
func (rw *rw_RW) FmtDateShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x2d}...)

	if t.Month() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Month()), 10)

	b = append(b, []byte{0x2d}...)

	if t.Day() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Day()), 10)

	return string(b)
}

// FmtDateMedium returns the medium date representation of 't' for 'rw_RW'
func (rw *rw_RW) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x20}...)
	b = append(b, rw.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'rw_RW'
func (rw *rw_RW) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x20}...)
	b = append(b, rw.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'rw_RW'
func (rw *rw_RW) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x20}...)
	b = append(b, rw.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2c, 0x20}...)
	b = append(b, rw.daysWide[t.Weekday()]...)

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'rw_RW'
func (rw *rw_RW) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, rw.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'rw_RW'
func (rw *rw_RW) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, rw.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, rw.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'rw_RW'
func (rw *rw_RW) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, rw.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, rw.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'rw_RW'
func (rw *rw_RW) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, rw.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, rw.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := rw.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
