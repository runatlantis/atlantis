package ksb

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type ksb struct {
	locale             string
	pluralsCardinal    []locales.PluralRule
	pluralsOrdinal     []locales.PluralRule
	pluralsRange       []locales.PluralRule
	decimal            string
	group              string
	minus              string
	percent            string
	perMille           string
	timeSeparator      string
	inifinity          string
	currencies         []string // idx = enum of currency code
	monthsAbbreviated  []string
	monthsNarrow       []string
	monthsWide         []string
	daysAbbreviated    []string
	daysNarrow         []string
	daysShort          []string
	daysWide           []string
	periodsAbbreviated []string
	periodsNarrow      []string
	periodsShort       []string
	periodsWide        []string
	erasAbbreviated    []string
	erasNarrow         []string
	erasWide           []string
	timezones          map[string]string
}

// New returns a new instance of translator for the 'ksb' locale
func New() locales.Translator {
	return &ksb{
		locale:             "ksb",
		pluralsCardinal:    []locales.PluralRule{2, 6},
		pluralsOrdinal:     nil,
		pluralsRange:       nil,
		timeSeparator:      ":",
		currencies:         []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TSh", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		monthsAbbreviated:  []string{"", "Jan", "Feb", "Mac", "Apr", "Mei", "Jun", "Jul", "Ago", "Sep", "Okt", "Nov", "Des"},
		monthsNarrow:       []string{"", "J", "F", "M", "A", "M", "J", "J", "A", "S", "O", "N", "D"},
		monthsWide:         []string{"", "Januali", "Febluali", "Machi", "Aplili", "Mei", "Juni", "Julai", "Agosti", "Septemba", "Oktoba", "Novemba", "Desemba"},
		daysAbbreviated:    []string{"Jpi", "Jtt", "Jmn", "Jtn", "Alh", "Iju", "Jmo"},
		daysNarrow:         []string{"2", "3", "4", "5", "A", "I", "1"},
		daysWide:           []string{"Jumaapii", "Jumaatatu", "Jumaane", "Jumaatano", "Alhamisi", "Ijumaa", "Jumaamosi"},
		periodsAbbreviated: []string{"makeo", "nyiaghuo"},
		periodsWide:        []string{"makeo", "nyiaghuo"},
		erasAbbreviated:    []string{"KK", "BK"},
		erasNarrow:         []string{"", ""},
		erasWide:           []string{"Kabla ya Klisto", "Baada ya Klisto"},
		timezones:          map[string]string{"AST": "AST", "HEOG": "HEOG", "WITA": "WITA", "ART": "ART", "GMT": "GMT", "GYT": "GYT", "WAT": "WAT", "NZDT": "NZDT", "HKST": "HKST", "HNT": "HNT", "VET": "VET", "CLT": "CLT", "TMST": "TMST", "AKST": "AKST", "HNOG": "HNOG", "TMT": "TMT", "CHAST": "CHAST", "PST": "PST", "AWST": "AWST", "EDT": "EDT", "HEEG": "HEEG", "HECU": "HECU", "JDT": "JDT", "HNEG": "HNEG", "MEZ": "MEZ", "LHST": "LHST", "EAT": "EAT", "OEZ": "OEZ", "∅∅∅": "∅∅∅", "CDT": "CDT", "HNPMX": "HNPMX", "AEDT": "AEDT", "WEZ": "WEZ", "GFT": "GFT", "HENOMX": "HENOMX", "OESZ": "OESZ", "ARST": "ARST", "UYST": "UYST", "CHADT": "CHADT", "PDT": "PDT", "WIB": "WIB", "BT": "BT", "NZST": "NZST", "JST": "JST", "WART": "WART", "MST": "MST", "MDT": "MDT", "SRT": "SRT", "AWDT": "AWDT", "HEPMX": "HEPMX", "LHDT": "LHDT", "HEPM": "HEPM", "HNNOMX": "HNNOMX", "WAST": "WAST", "CAT": "CAT", "WIT": "WIT", "COT": "COT", "UYT": "UYT", "CST": "CST", "ADT": "ADT", "CLST": "CLST", "HADT": "HADT", "COST": "COST", "SGT": "SGT", "ECT": "ECT", "ACWST": "ACWST", "IST": "IST", "HNPM": "HNPM", "HAST": "HAST", "ChST": "ChST", "BOT": "BOT", "ACST": "ACST", "MESZ": "MESZ", "WARST": "WARST", "HNCU": "HNCU", "WESZ": "WESZ", "AKDT": "AKDT", "EST": "EST", "ACDT": "ACDT", "ACWDT": "ACWDT", "HKT": "HKT", "AEST": "AEST", "SAST": "SAST", "MYT": "MYT", "HAT": "HAT"},
	}
}

// Locale returns the current translators string locale
func (ksb *ksb) Locale() string {
	return ksb.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'ksb'
func (ksb *ksb) PluralsCardinal() []locales.PluralRule {
	return ksb.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'ksb'
func (ksb *ksb) PluralsOrdinal() []locales.PluralRule {
	return ksb.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'ksb'
func (ksb *ksb) PluralsRange() []locales.PluralRule {
	return ksb.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'ksb'
func (ksb *ksb) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'ksb'
func (ksb *ksb) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'ksb'
func (ksb *ksb) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (ksb *ksb) MonthAbbreviated(month time.Month) string {
	return ksb.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (ksb *ksb) MonthsAbbreviated() []string {
	return ksb.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (ksb *ksb) MonthNarrow(month time.Month) string {
	return ksb.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (ksb *ksb) MonthsNarrow() []string {
	return ksb.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (ksb *ksb) MonthWide(month time.Month) string {
	return ksb.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (ksb *ksb) MonthsWide() []string {
	return ksb.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (ksb *ksb) WeekdayAbbreviated(weekday time.Weekday) string {
	return ksb.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (ksb *ksb) WeekdaysAbbreviated() []string {
	return ksb.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (ksb *ksb) WeekdayNarrow(weekday time.Weekday) string {
	return ksb.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (ksb *ksb) WeekdaysNarrow() []string {
	return ksb.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (ksb *ksb) WeekdayShort(weekday time.Weekday) string {
	return ksb.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (ksb *ksb) WeekdaysShort() []string {
	return ksb.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (ksb *ksb) WeekdayWide(weekday time.Weekday) string {
	return ksb.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (ksb *ksb) WeekdaysWide() []string {
	return ksb.daysWide
}

// Decimal returns the decimal point of number
func (ksb *ksb) Decimal() string {
	return ksb.decimal
}

// Group returns the group of number
func (ksb *ksb) Group() string {
	return ksb.group
}

// Group returns the minus sign of number
func (ksb *ksb) Minus() string {
	return ksb.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'ksb' and handles both Whole and Real numbers based on 'v'
func (ksb *ksb) FmtNumber(num float64, v uint64) string {

	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'ksb' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (ksb *ksb) FmtPercent(num float64, v uint64) string {
	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'ksb'
func (ksb *ksb) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := ksb.currencies[currency]
	l := len(s) + len(symbol) + 0
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ksb.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, ksb.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, ksb.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, ksb.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	b = append(b, symbol...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'ksb'
// in accounting notation.
func (ksb *ksb) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := ksb.currencies[currency]
	l := len(s) + len(symbol) + 0
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ksb.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, ksb.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {

		b = append(b, ksb.minus[0])

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, ksb.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, symbol...)
	} else {

		b = append(b, symbol...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'ksb'
func (ksb *ksb) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'ksb'
func (ksb *ksb) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, ksb.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'ksb'
func (ksb *ksb) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, ksb.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'ksb'
func (ksb *ksb) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, ksb.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, ksb.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'ksb'
func (ksb *ksb) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, ksb.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'ksb'
func (ksb *ksb) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, ksb.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, ksb.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'ksb'
func (ksb *ksb) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, ksb.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, ksb.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'ksb'
func (ksb *ksb) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, ksb.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, ksb.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := ksb.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
