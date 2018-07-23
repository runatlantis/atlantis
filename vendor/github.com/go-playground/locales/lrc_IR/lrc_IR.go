package lrc_IR

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type lrc_IR struct {
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

// New returns a new instance of translator for the 'lrc_IR' locale
func New() locales.Translator {
	return &lrc_IR{
		locale:                 "lrc_IR",
		pluralsCardinal:        nil,
		pluralsOrdinal:         nil,
		pluralsRange:           nil,
		decimal:                ".",
		group:                  ",",
		minus:                  "-",
		percent:                "%",
		timeSeparator:          ":",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyPositivePrefix: " ",
		currencyNegativePrefix: " ",
		monthsAbbreviated:      []string{"", "جانڤیە", "فئڤریە", "مارس", "آڤریل", "مئی", "جوٙأن", "جوٙلا", "آگوست", "سئپتامر", "ئوکتوڤر", "نوڤامر", "دئسامر"},
		monthsNarrow:           []string{"", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12"},
		monthsWide:             []string{"", "جانڤیە", "فئڤریە", "مارس", "آڤریل", "مئی", "جوٙأن", "جوٙلا", "آگوست", "سئپتامر", "ئوکتوڤر", "نوڤامر", "دئسامر"},
		periodsAbbreviated:     []string{"AM", "PM"},
		periodsWide:            []string{"AM", "PM"},
		erasAbbreviated:        []string{"BCE", "CE"},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"", ""},
		timezones:              map[string]string{"CAT": "CAT", "ECT": "ECT", "HEEG": "HEEG", "MEZ": "MEZ", "HENOMX": "HENOMX", "ART": "ART", "COST": "COST", "GMT": "GMT", "EST": "EST", "OEZ": "OEZ", "ARST": "ARST", "CHADT": "CHADT", "HECU": "HECU", "ADT": "ADT", "GFT": "GFT", "AKST": "AKST", "ACST": "ACST", "SRT": "SRT", "MESZ": "MESZ", "HKT": "HKT", "ACDT": "ACDT", "MST": "MST", "HADT": "HADT", "BT": "BT", "LHST": "LHST", "HAST": "HAST", "HAT": "HAT", "CLT": "CLT", "CHAST": "CHAST", "WESZ": "WESZ", "ACWDT": "ACWDT", "HNOG": "HNOG", "WITA": "WITA", "JST": "JST", "HKST": "HKST", "ChST": "ChST", "CDT": "روٙشنایی نئهادار روٙز", "AWST": "AWST", "WEZ": "WEZ", "HNNOMX": "HNNOMX", "MDT": "MDT", "TMT": "TMT", "HNPMX": "HNPMX", "HNPM": "HNPM", "TMST": "TMST", "SGT": "SGT", "WARST": "WARST", "WIT": "WIT", "UYST": "UYST", "PST": "PST", "PDT": "PDT", "NZDT": "NZDT", "AKDT": "AKDT", "HNEG": "HNEG", "EAT": "EAT", "OESZ": "OESZ", "HEPMX": "HEPMX", "AST": "AST", "AEST": "AEST", "WAT": "WAT", "WIB": "WIB", "JDT": "JDT", "IST": "IST", "HEOG": "HEOG", "EDT": "EDT", "HEPM": "HEPM", "COT": "COT", "UYT": "UYT", "WAST": "WAST", "LHDT": "LHDT", "CLST": "CLST", "CST": "گاٛت مینجاٛیی ئستاٛنداٛرد", "VET": "VET", "HNT": "HNT", "HNCU": "HNCU", "AWDT": "AWDT", "NZST": "NZST", "BOT": "BOT", "ACWST": "ACWST", "WART": "WART", "GYT": "GYT", "AEDT": "AEDT", "SAST": "SAST", "MYT": "MYT", "∅∅∅": "∅∅∅"},
	}
}

// Locale returns the current translators string locale
func (lrc *lrc_IR) Locale() string {
	return lrc.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'lrc_IR'
func (lrc *lrc_IR) PluralsCardinal() []locales.PluralRule {
	return lrc.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'lrc_IR'
func (lrc *lrc_IR) PluralsOrdinal() []locales.PluralRule {
	return lrc.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'lrc_IR'
func (lrc *lrc_IR) PluralsRange() []locales.PluralRule {
	return lrc.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'lrc_IR'
func (lrc *lrc_IR) CardinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'lrc_IR'
func (lrc *lrc_IR) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'lrc_IR'
func (lrc *lrc_IR) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (lrc *lrc_IR) MonthAbbreviated(month time.Month) string {
	return lrc.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (lrc *lrc_IR) MonthsAbbreviated() []string {
	return lrc.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (lrc *lrc_IR) MonthNarrow(month time.Month) string {
	return lrc.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (lrc *lrc_IR) MonthsNarrow() []string {
	return lrc.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (lrc *lrc_IR) MonthWide(month time.Month) string {
	return lrc.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (lrc *lrc_IR) MonthsWide() []string {
	return lrc.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (lrc *lrc_IR) WeekdayAbbreviated(weekday time.Weekday) string {
	return lrc.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (lrc *lrc_IR) WeekdaysAbbreviated() []string {
	return lrc.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (lrc *lrc_IR) WeekdayNarrow(weekday time.Weekday) string {
	return lrc.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (lrc *lrc_IR) WeekdaysNarrow() []string {
	return lrc.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (lrc *lrc_IR) WeekdayShort(weekday time.Weekday) string {
	return lrc.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (lrc *lrc_IR) WeekdaysShort() []string {
	return lrc.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (lrc *lrc_IR) WeekdayWide(weekday time.Weekday) string {
	return lrc.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (lrc *lrc_IR) WeekdaysWide() []string {
	return lrc.daysWide
}

// Decimal returns the decimal point of number
func (lrc *lrc_IR) Decimal() string {
	return lrc.decimal
}

// Group returns the group of number
func (lrc *lrc_IR) Group() string {
	return lrc.group
}

// Group returns the minus sign of number
func (lrc *lrc_IR) Minus() string {
	return lrc.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'lrc_IR' and handles both Whole and Real numbers based on 'v'
func (lrc *lrc_IR) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, lrc.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, lrc.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, lrc.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'lrc_IR' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (lrc *lrc_IR) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 3
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, lrc.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, lrc.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, lrc.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'lrc_IR'
func (lrc *lrc_IR) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := lrc.currencies[currency]
	l := len(s) + len(symbol) + 4 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, lrc.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, lrc.group[0])
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

	for j := len(lrc.currencyPositivePrefix) - 1; j >= 0; j-- {
		b = append(b, lrc.currencyPositivePrefix[j])
	}

	if num < 0 {
		b = append(b, lrc.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, lrc.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'lrc_IR'
// in accounting notation.
func (lrc *lrc_IR) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := lrc.currencies[currency]
	l := len(s) + len(symbol) + 4 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, lrc.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, lrc.group[0])
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

		for j := len(lrc.currencyNegativePrefix) - 1; j >= 0; j-- {
			b = append(b, lrc.currencyNegativePrefix[j])
		}

		b = append(b, lrc.minus[0])

	} else {

		for j := len(symbol) - 1; j >= 0; j-- {
			b = append(b, symbol[j])
		}

		for j := len(lrc.currencyPositivePrefix) - 1; j >= 0; j-- {
			b = append(b, lrc.currencyPositivePrefix[j])
		}

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, lrc.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'lrc_IR'
func (lrc *lrc_IR) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'lrc_IR'
func (lrc *lrc_IR) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x20}...)
	b = append(b, lrc.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'lrc_IR'
func (lrc *lrc_IR) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x20}...)
	b = append(b, lrc.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'lrc_IR'
func (lrc *lrc_IR) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x20}...)
	b = append(b, lrc.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2c, 0x20}...)
	b = append(b, lrc.daysWide[t.Weekday()]...)

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'lrc_IR'
func (lrc *lrc_IR) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, lrc.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'lrc_IR'
func (lrc *lrc_IR) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, lrc.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, lrc.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'lrc_IR'
func (lrc *lrc_IR) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, lrc.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, lrc.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'lrc_IR'
func (lrc *lrc_IR) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, lrc.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, lrc.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := lrc.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
