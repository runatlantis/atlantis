package vun_TZ

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type vun_TZ struct {
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

// New returns a new instance of translator for the 'vun_TZ' locale
func New() locales.Translator {
	return &vun_TZ{
		locale:             "vun_TZ",
		pluralsCardinal:    []locales.PluralRule{2, 6},
		pluralsOrdinal:     nil,
		pluralsRange:       nil,
		timeSeparator:      ":",
		currencies:         []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
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
		timezones:          map[string]string{"HEPM": "HEPM", "EDT": "EDT", "HEOG": "HEOG", "LHDT": "LHDT", "OEZ": "OEZ", "COST": "COST", "UYT": "UYT", "UYST": "UYST", "GFT": "GFT", "ACDT": "ACDT", "WART": "WART", "LHST": "LHST", "CST": "CST", "SAST": "SAST", "AKST": "AKST", "IST": "IST", "HNNOMX": "HNNOMX", "EAT": "EAT", "AEST": "AEST", "ACWDT": "ACWDT", "MESZ": "MESZ", "WARST": "WARST", "CLT": "CLT", "WIT": "WIT", "SGT": "SGT", "HEEG": "HEEG", "HAT": "HAT", "HNPM": "HNPM", "HADT": "HADT", "COT": "COT", "GYT": "GYT", "AWST": "AWST", "WEZ": "WEZ", "JST": "JST", "HNOG": "HNOG", "NZDT": "NZDT", "ECT": "ECT", "EST": "EST", "MDT": "MDT", "CAT": "CAT", "WAT": "WAT", "BT": "BT", "NZST": "NZST", "WESZ": "WESZ", "HNT": "HNT", "HNEG": "HNEG", "TMT": "TMT", "AST": "AST", "ADT": "ADT", "WAST": "WAST", "WIB": "WIB", "AKDT": "AKDT", "AWDT": "AWDT", "MEZ": "MEZ", "HKST": "HKST", "SRT": "SRT", "CHAST": "CHAST", "HECU": "HECU", "CDT": "CDT", "PST": "PST", "BOT": "BOT", "ACST": "ACST", "GMT": "GMT", "HEPMX": "HEPMX", "HAST": "HAST", "ChST": "ChST", "HNPMX": "HNPMX", "∅∅∅": "∅∅∅", "PDT": "PDT", "CHADT": "CHADT", "HNCU": "HNCU", "JDT": "JDT", "TMST": "TMST", "ARST": "ARST", "MST": "MST", "OESZ": "OESZ", "AEDT": "AEDT", "MYT": "MYT", "WITA": "WITA", "HENOMX": "HENOMX", "CLST": "CLST", "ART": "ART", "ACWST": "ACWST", "HKT": "HKT", "VET": "VET"},
	}
}

// Locale returns the current translators string locale
func (vun *vun_TZ) Locale() string {
	return vun.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'vun_TZ'
func (vun *vun_TZ) PluralsCardinal() []locales.PluralRule {
	return vun.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'vun_TZ'
func (vun *vun_TZ) PluralsOrdinal() []locales.PluralRule {
	return vun.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'vun_TZ'
func (vun *vun_TZ) PluralsRange() []locales.PluralRule {
	return vun.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'vun_TZ'
func (vun *vun_TZ) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'vun_TZ'
func (vun *vun_TZ) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'vun_TZ'
func (vun *vun_TZ) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (vun *vun_TZ) MonthAbbreviated(month time.Month) string {
	return vun.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (vun *vun_TZ) MonthsAbbreviated() []string {
	return vun.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (vun *vun_TZ) MonthNarrow(month time.Month) string {
	return vun.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (vun *vun_TZ) MonthsNarrow() []string {
	return vun.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (vun *vun_TZ) MonthWide(month time.Month) string {
	return vun.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (vun *vun_TZ) MonthsWide() []string {
	return vun.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (vun *vun_TZ) WeekdayAbbreviated(weekday time.Weekday) string {
	return vun.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (vun *vun_TZ) WeekdaysAbbreviated() []string {
	return vun.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (vun *vun_TZ) WeekdayNarrow(weekday time.Weekday) string {
	return vun.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (vun *vun_TZ) WeekdaysNarrow() []string {
	return vun.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (vun *vun_TZ) WeekdayShort(weekday time.Weekday) string {
	return vun.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (vun *vun_TZ) WeekdaysShort() []string {
	return vun.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (vun *vun_TZ) WeekdayWide(weekday time.Weekday) string {
	return vun.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (vun *vun_TZ) WeekdaysWide() []string {
	return vun.daysWide
}

// Decimal returns the decimal point of number
func (vun *vun_TZ) Decimal() string {
	return vun.decimal
}

// Group returns the group of number
func (vun *vun_TZ) Group() string {
	return vun.group
}

// Group returns the minus sign of number
func (vun *vun_TZ) Minus() string {
	return vun.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'vun_TZ' and handles both Whole and Real numbers based on 'v'
func (vun *vun_TZ) FmtNumber(num float64, v uint64) string {

	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'vun_TZ' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (vun *vun_TZ) FmtPercent(num float64, v uint64) string {
	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'vun_TZ'
func (vun *vun_TZ) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := vun.currencies[currency]
	l := len(s) + len(symbol) + 0
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, vun.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, vun.group[0])
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
		b = append(b, vun.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, vun.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'vun_TZ'
// in accounting notation.
func (vun *vun_TZ) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := vun.currencies[currency]
	l := len(s) + len(symbol) + 0
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, vun.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, vun.group[0])
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

		b = append(b, vun.minus[0])

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
			b = append(b, vun.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'vun_TZ'
func (vun *vun_TZ) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'vun_TZ'
func (vun *vun_TZ) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, vun.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'vun_TZ'
func (vun *vun_TZ) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, vun.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'vun_TZ'
func (vun *vun_TZ) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, vun.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, vun.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'vun_TZ'
func (vun *vun_TZ) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, vun.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'vun_TZ'
func (vun *vun_TZ) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, vun.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, vun.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'vun_TZ'
func (vun *vun_TZ) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, vun.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, vun.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'vun_TZ'
func (vun *vun_TZ) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, vun.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, vun.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := vun.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
