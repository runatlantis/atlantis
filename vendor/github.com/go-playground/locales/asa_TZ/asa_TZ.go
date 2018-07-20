package asa_TZ

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type asa_TZ struct {
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
	currencyPositiveSuffix string
	currencyNegativeSuffix string
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

// New returns a new instance of translator for the 'asa_TZ' locale
func New() locales.Translator {
	return &asa_TZ{
		locale:                 "asa_TZ",
		pluralsCardinal:        []locales.PluralRule{2, 6},
		pluralsOrdinal:         nil,
		pluralsRange:           nil,
		timeSeparator:          ":",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyPositiveSuffix: " ",
		currencyNegativeSuffix: " ",
		monthsAbbreviated:      []string{"", "Jan", "Feb", "Mac", "Apr", "Mei", "Jun", "Jul", "Ago", "Sep", "Okt", "Nov", "Dec"},
		monthsNarrow:           []string{"", "J", "F", "M", "A", "M", "J", "J", "A", "S", "O", "N", "D"},
		monthsWide:             []string{"", "Januari", "Februari", "Machi", "Aprili", "Mei", "Juni", "Julai", "Agosti", "Septemba", "Oktoba", "Novemba", "Desemba"},
		daysAbbreviated:        []string{"Jpi", "Jtt", "Jnn", "Jtn", "Alh", "Ijm", "Jmo"},
		daysNarrow:             []string{"J", "J", "J", "J", "A", "I", "J"},
		daysWide:               []string{"Jumapili", "Jumatatu", "Jumanne", "Jumatano", "Alhamisi", "Ijumaa", "Jumamosi"},
		periodsAbbreviated:     []string{"icheheavo", "ichamthi"},
		periodsWide:            []string{"icheheavo", "ichamthi"},
		erasAbbreviated:        []string{"KM", "BM"},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"Kabla yakwe Yethu", "Baada yakwe Yethu"},
		timezones:              map[string]string{"JST": "JST", "ACDT": "ACDT", "ACWST": "ACWST", "ACWDT": "ACWDT", "OEZ": "OEZ", "CHAST": "CHAST", "WAST": "WAST", "GFT": "GFT", "HKST": "HKST", "WARST": "WARST", "AWST": "AWST", "ECT": "ECT", "MEZ": "MEZ", "HNT": "HNT", "HENOMX": "HENOMX", "MST": "MST", "WIT": "WIT", "PST": "PST", "HKT": "HKT", "GYT": "GYT", "CST": "CST", "SAST": "SAST", "EDT": "EDT", "WART": "WART", "WITA": "WITA", "MDT": "MDT", "COT": "COT", "WESZ": "WESZ", "AKST": "AKST", "CLST": "CLST", "ARST": "ARST", "WIB": "WIB", "NZST": "NZST", "SRT": "SRT", "EAT": "EAT", "WEZ": "WEZ", "GMT": "GMT", "ChST": "ChST", "AKDT": "AKDT", "CLT": "CLT", "PDT": "PDT", "HEPMX": "HEPMX", "AST": "AST", "AEDT": "AEDT", "OESZ": "OESZ", "UYST": "UYST", "CHADT": "CHADT", "CDT": "CDT", "VET": "VET", "HNPM": "HNPM", "CAT": "CAT", "UYT": "UYT", "ADT": "ADT", "HEOG": "HEOG", "SGT": "SGT", "IST": "IST", "TMT": "TMT", "COST": "COST", "HNCU": "HNCU", "WAT": "WAT", "BT": "BT", "ACST": "ACST", "MESZ": "MESZ", "LHST": "LHST", "BOT": "BOT", "HNEG": "HNEG", "HAT": "HAT", "∅∅∅": "∅∅∅", "AWDT": "AWDT", "HNPMX": "HNPMX", "NZDT": "NZDT", "EST": "EST", "HEEG": "HEEG", "TMST": "TMST", "HECU": "HECU", "MYT": "MYT", "JDT": "JDT", "HNNOMX": "HNNOMX", "HADT": "HADT", "AEST": "AEST", "HNOG": "HNOG", "HAST": "HAST", "ART": "ART", "LHDT": "LHDT", "HEPM": "HEPM"},
	}
}

// Locale returns the current translators string locale
func (asa *asa_TZ) Locale() string {
	return asa.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'asa_TZ'
func (asa *asa_TZ) PluralsCardinal() []locales.PluralRule {
	return asa.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'asa_TZ'
func (asa *asa_TZ) PluralsOrdinal() []locales.PluralRule {
	return asa.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'asa_TZ'
func (asa *asa_TZ) PluralsRange() []locales.PluralRule {
	return asa.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'asa_TZ'
func (asa *asa_TZ) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'asa_TZ'
func (asa *asa_TZ) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'asa_TZ'
func (asa *asa_TZ) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (asa *asa_TZ) MonthAbbreviated(month time.Month) string {
	return asa.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (asa *asa_TZ) MonthsAbbreviated() []string {
	return asa.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (asa *asa_TZ) MonthNarrow(month time.Month) string {
	return asa.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (asa *asa_TZ) MonthsNarrow() []string {
	return asa.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (asa *asa_TZ) MonthWide(month time.Month) string {
	return asa.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (asa *asa_TZ) MonthsWide() []string {
	return asa.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (asa *asa_TZ) WeekdayAbbreviated(weekday time.Weekday) string {
	return asa.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (asa *asa_TZ) WeekdaysAbbreviated() []string {
	return asa.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (asa *asa_TZ) WeekdayNarrow(weekday time.Weekday) string {
	return asa.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (asa *asa_TZ) WeekdaysNarrow() []string {
	return asa.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (asa *asa_TZ) WeekdayShort(weekday time.Weekday) string {
	return asa.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (asa *asa_TZ) WeekdaysShort() []string {
	return asa.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (asa *asa_TZ) WeekdayWide(weekday time.Weekday) string {
	return asa.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (asa *asa_TZ) WeekdaysWide() []string {
	return asa.daysWide
}

// Decimal returns the decimal point of number
func (asa *asa_TZ) Decimal() string {
	return asa.decimal
}

// Group returns the group of number
func (asa *asa_TZ) Group() string {
	return asa.group
}

// Group returns the minus sign of number
func (asa *asa_TZ) Minus() string {
	return asa.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'asa_TZ' and handles both Whole and Real numbers based on 'v'
func (asa *asa_TZ) FmtNumber(num float64, v uint64) string {

	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'asa_TZ' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (asa *asa_TZ) FmtPercent(num float64, v uint64) string {
	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'asa_TZ'
func (asa *asa_TZ) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := asa.currencies[currency]
	l := len(s) + len(symbol) + 2
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, asa.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, asa.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, asa.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, asa.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	b = append(b, asa.currencyPositiveSuffix...)

	b = append(b, symbol...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'asa_TZ'
// in accounting notation.
func (asa *asa_TZ) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := asa.currencies[currency]
	l := len(s) + len(symbol) + 2
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, asa.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, asa.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {

		b = append(b, asa.minus[0])

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, asa.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, asa.currencyNegativeSuffix...)
		b = append(b, symbol...)
	} else {

		b = append(b, asa.currencyPositiveSuffix...)
		b = append(b, symbol...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'asa_TZ'
func (asa *asa_TZ) FmtDateShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Day() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2f}...)

	if t.Month() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Month()), 10)

	b = append(b, []byte{0x2f}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateMedium returns the medium date representation of 't' for 'asa_TZ'
func (asa *asa_TZ) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, asa.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'asa_TZ'
func (asa *asa_TZ) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, asa.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'asa_TZ'
func (asa *asa_TZ) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, asa.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, asa.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'asa_TZ'
func (asa *asa_TZ) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, asa.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'asa_TZ'
func (asa *asa_TZ) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, asa.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, asa.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'asa_TZ'
func (asa *asa_TZ) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, asa.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, asa.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'asa_TZ'
func (asa *asa_TZ) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, asa.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, asa.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := asa.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
