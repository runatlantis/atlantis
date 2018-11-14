package lrc_IQ

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type lrc_IQ struct {
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

// New returns a new instance of translator for the 'lrc_IQ' locale
func New() locales.Translator {
	return &lrc_IQ{
		locale:                 "lrc_IQ",
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
		timezones:              map[string]string{"HEOG": "HEOG", "HADT": "HADT", "PDT": "PDT", "HEEG": "HEEG", "EDT": "EDT", "HEPM": "HEPM", "VET": "VET", "CDT": "روٙشنایی نئهادار روٙز", "WEZ": "WEZ", "ACWDT": "ACWDT", "CST": "گاٛت مینجاٛیی ئستاٛنداٛرد", "MST": "MST", "HNNOMX": "HNNOMX", "TMT": "TMT", "TMST": "TMST", "HAST": "HAST", "GFT": "GFT", "WIB": "WIB", "CLT": "CLT", "JST": "JST", "ACWST": "ACWST", "HNEG": "HNEG", "ACDT": "ACDT", "MEZ": "MEZ", "UYT": "UYT", "AEST": "AEST", "AWDT": "AWDT", "HNOG": "HNOG", "EST": "EST", "WART": "WART", "GYT": "GYT", "ChST": "ChST", "CHAST": "CHAST", "HNPMX": "HNPMX", "AEDT": "AEDT", "BT": "BT", "AKDT": "AKDT", "WITA": "WITA", "COT": "COT", "UYST": "UYST", "JDT": "JDT", "ECT": "ECT", "HAT": "HAT", "HEPMX": "HEPMX", "WAST": "WAST", "AKST": "AKST", "MESZ": "MESZ", "WARST": "WARST", "HNT": "HNT", "HENOMX": "HENOMX", "ART": "ART", "MDT": "MDT", "IST": "IST", "HNPM": "HNPM", "WIT": "WIT", "EAT": "EAT", "CHADT": "CHADT", "WESZ": "WESZ", "AST": "AST", "ACST": "ACST", "HKT": "HKT", "HKST": "HKST", "ARST": "ARST", "HNCU": "HNCU", "ADT": "ADT", "MYT": "MYT", "SRT": "SRT", "CLST": "CLST", "CAT": "CAT", "COST": "COST", "OESZ": "OESZ", "NZST": "NZST", "NZDT": "NZDT", "HECU": "HECU", "AWST": "AWST", "WAT": "WAT", "BOT": "BOT", "SGT": "SGT", "GMT": "GMT", "SAST": "SAST", "PST": "PST", "LHST": "LHST", "LHDT": "LHDT", "∅∅∅": "∅∅∅", "OEZ": "OEZ"},
	}
}

// Locale returns the current translators string locale
func (lrc *lrc_IQ) Locale() string {
	return lrc.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'lrc_IQ'
func (lrc *lrc_IQ) PluralsCardinal() []locales.PluralRule {
	return lrc.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'lrc_IQ'
func (lrc *lrc_IQ) PluralsOrdinal() []locales.PluralRule {
	return lrc.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'lrc_IQ'
func (lrc *lrc_IQ) PluralsRange() []locales.PluralRule {
	return lrc.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'lrc_IQ'
func (lrc *lrc_IQ) CardinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'lrc_IQ'
func (lrc *lrc_IQ) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'lrc_IQ'
func (lrc *lrc_IQ) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (lrc *lrc_IQ) MonthAbbreviated(month time.Month) string {
	return lrc.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (lrc *lrc_IQ) MonthsAbbreviated() []string {
	return lrc.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (lrc *lrc_IQ) MonthNarrow(month time.Month) string {
	return lrc.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (lrc *lrc_IQ) MonthsNarrow() []string {
	return lrc.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (lrc *lrc_IQ) MonthWide(month time.Month) string {
	return lrc.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (lrc *lrc_IQ) MonthsWide() []string {
	return lrc.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (lrc *lrc_IQ) WeekdayAbbreviated(weekday time.Weekday) string {
	return lrc.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (lrc *lrc_IQ) WeekdaysAbbreviated() []string {
	return lrc.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (lrc *lrc_IQ) WeekdayNarrow(weekday time.Weekday) string {
	return lrc.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (lrc *lrc_IQ) WeekdaysNarrow() []string {
	return lrc.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (lrc *lrc_IQ) WeekdayShort(weekday time.Weekday) string {
	return lrc.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (lrc *lrc_IQ) WeekdaysShort() []string {
	return lrc.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (lrc *lrc_IQ) WeekdayWide(weekday time.Weekday) string {
	return lrc.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (lrc *lrc_IQ) WeekdaysWide() []string {
	return lrc.daysWide
}

// Decimal returns the decimal point of number
func (lrc *lrc_IQ) Decimal() string {
	return lrc.decimal
}

// Group returns the group of number
func (lrc *lrc_IQ) Group() string {
	return lrc.group
}

// Group returns the minus sign of number
func (lrc *lrc_IQ) Minus() string {
	return lrc.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'lrc_IQ' and handles both Whole and Real numbers based on 'v'
func (lrc *lrc_IQ) FmtNumber(num float64, v uint64) string {

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

// FmtPercent returns 'num' with digits/precision of 'v' for 'lrc_IQ' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (lrc *lrc_IQ) FmtPercent(num float64, v uint64) string {
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

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'lrc_IQ'
func (lrc *lrc_IQ) FmtCurrency(num float64, v uint64, currency currency.Type) string {

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

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'lrc_IQ'
// in accounting notation.
func (lrc *lrc_IQ) FmtAccounting(num float64, v uint64, currency currency.Type) string {

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

// FmtDateShort returns the short date representation of 't' for 'lrc_IQ'
func (lrc *lrc_IQ) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'lrc_IQ'
func (lrc *lrc_IQ) FmtDateMedium(t time.Time) string {

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

// FmtDateLong returns the long date representation of 't' for 'lrc_IQ'
func (lrc *lrc_IQ) FmtDateLong(t time.Time) string {

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

// FmtDateFull returns the full date representation of 't' for 'lrc_IQ'
func (lrc *lrc_IQ) FmtDateFull(t time.Time) string {

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

// FmtTimeShort returns the short time representation of 't' for 'lrc_IQ'
func (lrc *lrc_IQ) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, lrc.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, lrc.periodsAbbreviated[0]...)
	} else {
		b = append(b, lrc.periodsAbbreviated[1]...)
	}

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'lrc_IQ'
func (lrc *lrc_IQ) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
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

	if t.Hour() < 12 {
		b = append(b, lrc.periodsAbbreviated[0]...)
	} else {
		b = append(b, lrc.periodsAbbreviated[1]...)
	}

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'lrc_IQ'
func (lrc *lrc_IQ) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
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

	if t.Hour() < 12 {
		b = append(b, lrc.periodsAbbreviated[0]...)
	} else {
		b = append(b, lrc.periodsAbbreviated[1]...)
	}

	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'lrc_IQ'
func (lrc *lrc_IQ) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
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

	if t.Hour() < 12 {
		b = append(b, lrc.periodsAbbreviated[0]...)
	} else {
		b = append(b, lrc.periodsAbbreviated[1]...)
	}

	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := lrc.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
