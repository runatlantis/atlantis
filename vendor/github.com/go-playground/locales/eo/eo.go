package eo

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type eo struct {
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
	currencyPositiveSuffix string
	currencyNegativePrefix string
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

// New returns a new instance of translator for the 'eo' locale
func New() locales.Translator {
	return &eo{
		locale:                 "eo",
		pluralsCardinal:        []locales.PluralRule{2, 6},
		pluralsOrdinal:         nil,
		pluralsRange:           nil,
		decimal:                ",",
		group:                  " ",
		minus:                  "−",
		percent:                "%",
		perMille:               "‰",
		timeSeparator:          ":",
		inifinity:              "∞",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AU$", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "R$", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CA$", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CN¥", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "€", "FIM", "FJD", "FKP", "FRF", "£", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HK$", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "₹", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JP¥", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "₩", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MX$", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "฿", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "₺", "TTD", "NT$", "TZS", "UAH", "UAK", "UGS", "UGX", "US$", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyPositivePrefix: " ",
		currencyPositiveSuffix: "K",
		currencyNegativePrefix: " ",
		currencyNegativeSuffix: "K",
		monthsAbbreviated:      []string{"", "jan", "feb", "mar", "apr", "maj", "jun", "jul", "aŭg", "sep", "okt", "nov", "dec"},
		monthsNarrow:           []string{"", "J", "F", "M", "A", "M", "J", "J", "A", "S", "O", "N", "D"},
		monthsWide:             []string{"", "januaro", "februaro", "marto", "aprilo", "majo", "junio", "julio", "aŭgusto", "septembro", "oktobro", "novembro", "decembro"},
		daysAbbreviated:        []string{"di", "lu", "ma", "me", "ĵa", "ve", "sa"},
		daysNarrow:             []string{"D", "L", "M", "M", "Ĵ", "V", "S"},
		daysWide:               []string{"dimanĉo", "lundo", "mardo", "merkredo", "ĵaŭdo", "vendredo", "sabato"},
		periodsAbbreviated:     []string{"atm", "ptm"},
		periodsNarrow:          []string{"a", "p"},
		periodsWide:            []string{"atm", "ptm"},
		erasAbbreviated:        []string{"aK", "pK"},
		erasNarrow:             []string{"aK", "pK"},
		erasWide:               []string{"aK", "pK"},
		timezones:              map[string]string{"CHAST": "CHAST", "AST": "atlantika nord-amerika norma tempo", "AEDT": "orienta aŭstralia somera tempo", "SAST": "suda afrika tempo", "ACDT": "centra aŭstralia somera tempo", "HEEG": "HEEG", "PST": "pacifika nord-amerika norma tempo", "AEST": "orienta aŭstralia norma tempo", "WAT": "okcidenta afrika norma tempo", "JDT": "japana somera tempo", "TMST": "TMST", "ChST": "ChST", "WAST": "okcidenta afrika somera tempo", "WIB": "okcidenta indonezia tempo", "WARST": "WARST", "WITA": "centra indonezia tempo", "WIT": "orienta indonezia tempo", "CST": "centra nord-amerika norma tempo", "CDT": "centra nord-amerika somera tempo", "∅∅∅": "∅∅∅", "EDT": "orienta nord-amerika somera tempo", "HNT": "HNT", "HENOMX": "HENOMX", "TMT": "TMT", "ARST": "ARST", "ADT": "atlantika nord-amerika somera tempo", "WEZ": "okcidenta eŭropa norma tempo", "NZDT": "NZDT", "LHDT": "LHDT", "CLST": "CLST", "HNCU": "HNCU", "JST": "japana norma tempo", "HAT": "HAT", "HEPM": "HEPM", "EAT": "orienta afrika tempo", "CLT": "CLT", "EST": "orienta nord-amerika norma tempo", "ACWST": "centrokcidenta aŭstralia norma tempo", "LHST": "LHST", "HAST": "HAST", "HNPMX": "HNPMX", "PDT": "pacifika nord-amerika somera tempo", "BOT": "BOT", "HKST": "HKST", "IST": "barata tempo", "WART": "WART", "OEZ": "orienta eŭropa norma tempo", "GYT": "GYT", "BT": "BT", "HEOG": "HEOG", "HNPM": "HNPM", "OESZ": "orienta eŭropa somera tempo", "HADT": "HADT", "SGT": "SGT", "AWDT": "okcidenta aŭstralia somera tempo", "HEPMX": "HEPMX", "MST": "monta nord-amerika norma tempo", "AKST": "AKST", "ACWDT": "centrokcidenta aŭstralia somera tempo", "MEZ": "centra eŭropa norma tempo", "ART": "ART", "GMT": "universala tempo kunordigita", "HECU": "HECU", "MDT": "monta nord-amerika somera tempo", "WESZ": "okcidenta eŭropa somera tempo", "MYT": "MYT", "AKDT": "AKDT", "ECT": "ECT", "HNNOMX": "HNNOMX", "CHADT": "CHADT", "ACST": "centra aŭstralia norma tempo", "HNEG": "HNEG", "HNOG": "HNOG", "MESZ": "centra eŭropa somera tempo", "VET": "VET", "GFT": "GFT", "COST": "COST", "UYT": "UYT", "UYST": "UYST", "AWST": "okcidenta aŭstralia norma tempo", "NZST": "NZST", "HKT": "HKT", "SRT": "SRT", "CAT": "centra afrika tempo", "COT": "COT"},
	}
}

// Locale returns the current translators string locale
func (eo *eo) Locale() string {
	return eo.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'eo'
func (eo *eo) PluralsCardinal() []locales.PluralRule {
	return eo.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'eo'
func (eo *eo) PluralsOrdinal() []locales.PluralRule {
	return eo.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'eo'
func (eo *eo) PluralsRange() []locales.PluralRule {
	return eo.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'eo'
func (eo *eo) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'eo'
func (eo *eo) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'eo'
func (eo *eo) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (eo *eo) MonthAbbreviated(month time.Month) string {
	return eo.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (eo *eo) MonthsAbbreviated() []string {
	return eo.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (eo *eo) MonthNarrow(month time.Month) string {
	return eo.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (eo *eo) MonthsNarrow() []string {
	return eo.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (eo *eo) MonthWide(month time.Month) string {
	return eo.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (eo *eo) MonthsWide() []string {
	return eo.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (eo *eo) WeekdayAbbreviated(weekday time.Weekday) string {
	return eo.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (eo *eo) WeekdaysAbbreviated() []string {
	return eo.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (eo *eo) WeekdayNarrow(weekday time.Weekday) string {
	return eo.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (eo *eo) WeekdaysNarrow() []string {
	return eo.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (eo *eo) WeekdayShort(weekday time.Weekday) string {
	return eo.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (eo *eo) WeekdaysShort() []string {
	return eo.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (eo *eo) WeekdayWide(weekday time.Weekday) string {
	return eo.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (eo *eo) WeekdaysWide() []string {
	return eo.daysWide
}

// Decimal returns the decimal point of number
func (eo *eo) Decimal() string {
	return eo.decimal
}

// Group returns the group of number
func (eo *eo) Group() string {
	return eo.group
}

// Group returns the minus sign of number
func (eo *eo) Minus() string {
	return eo.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'eo' and handles both Whole and Real numbers based on 'v'
func (eo *eo) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 4 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, eo.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(eo.group) - 1; j >= 0; j-- {
					b = append(b, eo.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		for j := len(eo.minus) - 1; j >= 0; j-- {
			b = append(b, eo.minus[j])
		}
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'eo' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (eo *eo) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 5
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, eo.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		for j := len(eo.minus) - 1; j >= 0; j-- {
			b = append(b, eo.minus[j])
		}
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, eo.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'eo'
func (eo *eo) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := eo.currencies[currency]
	l := len(s) + len(symbol) + 7

	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, eo.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	for j := len(symbol) - 1; j >= 0; j-- {
		b = append(b, symbol[j])
	}

	for j := len(eo.currencyPositivePrefix) - 1; j >= 0; j-- {
		b = append(b, eo.currencyPositivePrefix[j])
	}

	if num < 0 {
		for j := len(eo.minus) - 1; j >= 0; j-- {
			b = append(b, eo.minus[j])
		}
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, eo.currencyPositiveSuffix...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'eo'
// in accounting notation.
func (eo *eo) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := eo.currencies[currency]
	l := len(s) + len(symbol) + 7

	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, eo.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {

		for j := len(symbol) - 1; j >= 0; j-- {
			b = append(b, symbol[j])
		}

		for j := len(eo.currencyNegativePrefix) - 1; j >= 0; j-- {
			b = append(b, eo.currencyNegativePrefix[j])
		}

		for j := len(eo.minus) - 1; j >= 0; j-- {
			b = append(b, eo.minus[j])
		}

	} else {

		for j := len(symbol) - 1; j >= 0; j-- {
			b = append(b, symbol[j])
		}

		for j := len(eo.currencyPositivePrefix) - 1; j >= 0; j-- {
			b = append(b, eo.currencyPositivePrefix[j])
		}

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if num < 0 {
		b = append(b, eo.currencyNegativeSuffix...)
	} else {

		b = append(b, eo.currencyPositiveSuffix...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'eo'
func (eo *eo) FmtDateShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 9 {
		b = append(b, strconv.Itoa(t.Year())[2:]...)
	} else {
		b = append(b, strconv.Itoa(t.Year())[1:]...)
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

// FmtDateMedium returns the medium date representation of 't' for 'eo'
func (eo *eo) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x2d}...)
	b = append(b, eo.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x2d}...)

	if t.Day() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Day()), 10)

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'eo'
func (eo *eo) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x2d}...)
	b = append(b, eo.monthsWide[t.Month()]...)
	b = append(b, []byte{0x2d}...)

	if t.Day() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Day()), 10)

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'eo'
func (eo *eo) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, eo.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2d, 0x61}...)
	b = append(b, []byte{0x20, 0x64, 0x65}...)
	b = append(b, []byte{0x20}...)
	b = append(b, eo.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'eo'
func (eo *eo) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, eo.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'eo'
func (eo *eo) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, eo.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, eo.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'eo'
func (eo *eo) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, eo.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, eo.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'eo'
func (eo *eo) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, []byte{0x2d, 0x61}...)
	b = append(b, []byte{0x20, 0x68, 0x6f, 0x72, 0x6f}...)
	b = append(b, []byte{0x20, 0x6b, 0x61, 0x6a}...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, eo.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := eo.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
