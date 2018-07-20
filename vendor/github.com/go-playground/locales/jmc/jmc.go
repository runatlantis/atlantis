package jmc

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type jmc struct {
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

// New returns a new instance of translator for the 'jmc' locale
func New() locales.Translator {
	return &jmc{
		locale:             "jmc",
		pluralsCardinal:    []locales.PluralRule{2, 6},
		pluralsOrdinal:     nil,
		pluralsRange:       nil,
		timeSeparator:      ":",
		currencies:         []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TSh", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		monthsAbbreviated:  []string{"", "Jan", "Feb", "Mac", "Apr", "Mei", "Jun", "Jul", "Ago", "Sep", "Okt", "Nov", "Des"},
		monthsNarrow:       []string{"", "J", "F", "M", "A", "M", "J", "J", "A", "S", "O", "N", "D"},
		monthsWide:         []string{"", "Januari", "Februari", "Machi", "Aprilyi", "Mei", "Junyi", "Julyai", "Agusti", "Septemba", "Oktoba", "Novemba", "Desemba"},
		daysAbbreviated:    []string{"Jpi", "Jtt", "Jnn", "Jtn", "Alh", "Iju", "Jmo"},
		daysNarrow:         []string{"J", "J", "J", "J", "A", "I", "J"},
		daysWide:           []string{"Jumapilyi", "Jumatatuu", "Jumanne", "Jumatanu", "Alhamisi", "Ijumaa", "Jumamosi"},
		periodsAbbreviated: []string{"utuko", "kyiukonyi"},
		periodsWide:        []string{"utuko", "kyiukonyi"},
		erasAbbreviated:    []string{"KK", "BK"},
		erasNarrow:         []string{"", ""},
		erasWide:           []string{"Kabla ya Kristu", "Baada ya Kristu"},
		timezones:          map[string]string{"SAST": "SAST", "IST": "IST", "HNT": "HNT", "HENOMX": "HENOMX", "LHDT": "LHDT", "ART": "ART", "COT": "COT", "CST": "CST", "AEDT": "AEDT", "WAST": "WAST", "GFT": "GFT", "ACST": "ACST", "WARST": "WARST", "ARST": "ARST", "AEST": "AEST", "ACWDT": "ACWDT", "HEOG": "HEOG", "MESZ": "MESZ", "WART": "WART", "HADT": "HADT", "HECU": "HECU", "AST": "AST", "BT": "BT", "JDT": "JDT", "MEZ": "MEZ", "HEPM": "HEPM", "OESZ": "OESZ", "HAST": "HAST", "PST": "PST", "WAT": "WAT", "HAT": "HAT", "HNPM": "HNPM", "WESZ": "WESZ", "JST": "JST", "ACWST": "ACWST", "AKDT": "AKDT", "GMT": "GMT", "UYST": "UYST", "HNCU": "HNCU", "PDT": "PDT", "HNPMX": "HNPMX", "WEZ": "WEZ", "NZST": "NZST", "EDT": "EDT", "HEEG": "HEEG", "HNNOMX": "HNNOMX", "CLST": "CLST", "∅∅∅": "∅∅∅", "UYT": "UYT", "MST": "MST", "SGT": "SGT", "ACDT": "ACDT", "HEPMX": "HEPMX", "HKST": "HKST", "SRT": "SRT", "CAT": "CAT", "CDT": "CDT", "MDT": "MDT", "AKST": "AKST", "ECT": "ECT", "HNOG": "HNOG", "EAT": "EAT", "TMST": "TMST", "OEZ": "OEZ", "CHADT": "CHADT", "ADT": "ADT", "BOT": "BOT", "COST": "COST", "NZDT": "NZDT", "GYT": "GYT", "EST": "EST", "WIT": "WIT", "TMT": "TMT", "ChST": "ChST", "CHAST": "CHAST", "AWST": "AWST", "CLT": "CLT", "HNEG": "HNEG", "WITA": "WITA", "AWDT": "AWDT", "WIB": "WIB", "MYT": "MYT", "HKT": "HKT", "LHST": "LHST", "VET": "VET"},
	}
}

// Locale returns the current translators string locale
func (jmc *jmc) Locale() string {
	return jmc.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'jmc'
func (jmc *jmc) PluralsCardinal() []locales.PluralRule {
	return jmc.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'jmc'
func (jmc *jmc) PluralsOrdinal() []locales.PluralRule {
	return jmc.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'jmc'
func (jmc *jmc) PluralsRange() []locales.PluralRule {
	return jmc.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'jmc'
func (jmc *jmc) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'jmc'
func (jmc *jmc) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'jmc'
func (jmc *jmc) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (jmc *jmc) MonthAbbreviated(month time.Month) string {
	return jmc.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (jmc *jmc) MonthsAbbreviated() []string {
	return jmc.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (jmc *jmc) MonthNarrow(month time.Month) string {
	return jmc.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (jmc *jmc) MonthsNarrow() []string {
	return jmc.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (jmc *jmc) MonthWide(month time.Month) string {
	return jmc.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (jmc *jmc) MonthsWide() []string {
	return jmc.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (jmc *jmc) WeekdayAbbreviated(weekday time.Weekday) string {
	return jmc.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (jmc *jmc) WeekdaysAbbreviated() []string {
	return jmc.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (jmc *jmc) WeekdayNarrow(weekday time.Weekday) string {
	return jmc.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (jmc *jmc) WeekdaysNarrow() []string {
	return jmc.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (jmc *jmc) WeekdayShort(weekday time.Weekday) string {
	return jmc.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (jmc *jmc) WeekdaysShort() []string {
	return jmc.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (jmc *jmc) WeekdayWide(weekday time.Weekday) string {
	return jmc.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (jmc *jmc) WeekdaysWide() []string {
	return jmc.daysWide
}

// Decimal returns the decimal point of number
func (jmc *jmc) Decimal() string {
	return jmc.decimal
}

// Group returns the group of number
func (jmc *jmc) Group() string {
	return jmc.group
}

// Group returns the minus sign of number
func (jmc *jmc) Minus() string {
	return jmc.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'jmc' and handles both Whole and Real numbers based on 'v'
func (jmc *jmc) FmtNumber(num float64, v uint64) string {

	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'jmc' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (jmc *jmc) FmtPercent(num float64, v uint64) string {
	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'jmc'
func (jmc *jmc) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := jmc.currencies[currency]
	l := len(s) + len(symbol) + 0
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, jmc.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, jmc.group[0])
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

	if num < 0 {
		b = append(b, jmc.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, jmc.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'jmc'
// in accounting notation.
func (jmc *jmc) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := jmc.currencies[currency]
	l := len(s) + len(symbol) + 0
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, jmc.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, jmc.group[0])
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

		b = append(b, jmc.minus[0])

	} else {

		for j := len(symbol) - 1; j >= 0; j-- {
			b = append(b, symbol[j])
		}

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, jmc.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'jmc'
func (jmc *jmc) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'jmc'
func (jmc *jmc) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, jmc.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'jmc'
func (jmc *jmc) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, jmc.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'jmc'
func (jmc *jmc) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, jmc.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, jmc.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'jmc'
func (jmc *jmc) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, jmc.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'jmc'
func (jmc *jmc) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, jmc.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, jmc.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'jmc'
func (jmc *jmc) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, jmc.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, jmc.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'jmc'
func (jmc *jmc) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, jmc.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, jmc.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := jmc.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
