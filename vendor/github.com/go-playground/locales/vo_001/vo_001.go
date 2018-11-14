package vo_001

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type vo_001 struct {
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

// New returns a new instance of translator for the 'vo_001' locale
func New() locales.Translator {
	return &vo_001{
		locale:            "vo_001",
		pluralsCardinal:   []locales.PluralRule{2, 6},
		pluralsOrdinal:    nil,
		pluralsRange:      nil,
		timeSeparator:     ":",
		currencies:        []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		monthsAbbreviated: []string{"", "yan", "feb", "mäz", "prl", "may", "yun", "yul", "gst", "set", "ton", "nov", "dek"},
		monthsNarrow:      []string{"", "Y", "F", "M", "P", "M", "Y", "Y", "G", "S", "T", "N", "D"},
		monthsWide:        []string{"", "yanul", "febul", "mäzul", "prilul", "mayul", "yunul", "yulul", "gustul", "setul", "tobul", "novul", "dekul"},
		daysAbbreviated:   []string{"su.", "mu.", "tu.", "ve.", "dö.", "fr.", "zä."},
		daysNarrow:        []string{"S", "M", "T", "V", "D", "F", "Z"},
		daysWide:          []string{"sudel", "mudel", "tudel", "vedel", "dödel", "fridel", "zädel"},
		erasAbbreviated:   []string{"b. t. kr.", "p. t. kr."},
		erasNarrow:        []string{"", ""},
		erasWide:          []string{"b. t. kr.", "p. t. kr."},
		timezones:         map[string]string{"EAT": "EAT", "ART": "ART", "WAST": "WAST", "EDT": "EDT", "AST": "AST", "WAT": "WAT", "AKST": "AKST", "WARST": "WARST", "HNT": "HNT", "TMT": "TMT", "HAST": "HAST", "CDT": "CDT", "AWDT": "AWDT", "HNEG": "HNEG", "MDT": "MDT", "CAT": "CAT", "COST": "COST", "HNCU": "HNCU", "HECU": "HECU", "HEPMX": "HEPMX", "SAST": "SAST", "BT": "BT", "JST": "JST", "ACST": "ACST", "MST": "MST", "CLT": "CLT", "WIT": "WIT", "CHAST": "CHAST", "GMT": "GMT", "WIB": "WIB", "HAT": "HAT", "HADT": "HADT", "MYT": "MYT", "HNOG": "HNOG", "EST": "EST", "SRT": "SRT", "CLST": "CLST", "CST": "CST", "JDT": "JDT", "PDT": "PDT", "AEDT": "AEDT", "MESZ": "MESZ", "IST": "IST", "HNPM": "HNPM", "OEZ": "OEZ", "UYST": "UYST", "CHADT": "CHADT", "WITA": "WITA", "HNPMX": "HNPMX", "SGT": "SGT", "ECT": "ECT", "WART": "WART", "HENOMX": "HENOMX", "OESZ": "OESZ", "AKDT": "AKDT", "MEZ": "MEZ", "ADT": "ADT", "WESZ": "WESZ", "NZDT": "NZDT", "VET": "VET", "HEPM": "HEPM", "COT": "COT", "UYT": "UYT", "AEST": "AEST", "WEZ": "WEZ", "NZST": "NZST", "HEOG": "HEOG", "ARST": "ARST", "∅∅∅": "∅∅∅", "ChST": "ChST", "PST": "PST", "LHDT": "LHDT", "AWST": "AWST", "GFT": "GFT", "ACDT": "ACDT", "ACWDT": "ACWDT", "HEEG": "HEEG", "HNNOMX": "HNNOMX", "TMST": "TMST", "BOT": "BOT", "ACWST": "ACWST", "GYT": "GYT", "HKT": "HKT", "HKST": "HKST", "LHST": "LHST"},
	}
}

// Locale returns the current translators string locale
func (vo *vo_001) Locale() string {
	return vo.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'vo_001'
func (vo *vo_001) PluralsCardinal() []locales.PluralRule {
	return vo.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'vo_001'
func (vo *vo_001) PluralsOrdinal() []locales.PluralRule {
	return vo.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'vo_001'
func (vo *vo_001) PluralsRange() []locales.PluralRule {
	return vo.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'vo_001'
func (vo *vo_001) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'vo_001'
func (vo *vo_001) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'vo_001'
func (vo *vo_001) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (vo *vo_001) MonthAbbreviated(month time.Month) string {
	return vo.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (vo *vo_001) MonthsAbbreviated() []string {
	return vo.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (vo *vo_001) MonthNarrow(month time.Month) string {
	return vo.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (vo *vo_001) MonthsNarrow() []string {
	return vo.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (vo *vo_001) MonthWide(month time.Month) string {
	return vo.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (vo *vo_001) MonthsWide() []string {
	return vo.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (vo *vo_001) WeekdayAbbreviated(weekday time.Weekday) string {
	return vo.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (vo *vo_001) WeekdaysAbbreviated() []string {
	return vo.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (vo *vo_001) WeekdayNarrow(weekday time.Weekday) string {
	return vo.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (vo *vo_001) WeekdaysNarrow() []string {
	return vo.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (vo *vo_001) WeekdayShort(weekday time.Weekday) string {
	return vo.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (vo *vo_001) WeekdaysShort() []string {
	return vo.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (vo *vo_001) WeekdayWide(weekday time.Weekday) string {
	return vo.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (vo *vo_001) WeekdaysWide() []string {
	return vo.daysWide
}

// Decimal returns the decimal point of number
func (vo *vo_001) Decimal() string {
	return vo.decimal
}

// Group returns the group of number
func (vo *vo_001) Group() string {
	return vo.group
}

// Group returns the minus sign of number
func (vo *vo_001) Minus() string {
	return vo.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'vo_001' and handles both Whole and Real numbers based on 'v'
func (vo *vo_001) FmtNumber(num float64, v uint64) string {

	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'vo_001' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (vo *vo_001) FmtPercent(num float64, v uint64) string {
	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'vo_001'
func (vo *vo_001) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := vo.currencies[currency]
	return string(append(append([]byte{}, symbol...), s...))
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'vo_001'
// in accounting notation.
func (vo *vo_001) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := vo.currencies[currency]
	return string(append(append([]byte{}, symbol...), s...))
}

// FmtDateShort returns the short date representation of 't' for 'vo_001'
func (vo *vo_001) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'vo_001'
func (vo *vo_001) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x20}...)
	b = append(b, vo.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x2e, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'vo_001'
func (vo *vo_001) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x20}...)
	b = append(b, vo.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'vo_001'
func (vo *vo_001) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x20}...)
	b = append(b, vo.monthsWide[t.Month()]...)
	b = append(b, []byte{0x61}...)
	b = append(b, []byte{0x20, 0x64}...)
	b = append(b, []byte{0x2e, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x69, 0x64}...)

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'vo_001'
func (vo *vo_001) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, vo.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'vo_001'
func (vo *vo_001) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, vo.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, vo.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'vo_001'
func (vo *vo_001) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, vo.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, vo.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'vo_001'
func (vo *vo_001) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, vo.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, vo.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := vo.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
