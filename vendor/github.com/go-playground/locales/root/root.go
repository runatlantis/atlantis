package root

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type root struct {
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

// New returns a new instance of translator for the 'root' locale
func New() locales.Translator {
	return &root{
		locale:             "root",
		pluralsCardinal:    []locales.PluralRule{6},
		pluralsOrdinal:     []locales.PluralRule{6},
		pluralsRange:       nil,
		decimal:            ".",
		group:              ",",
		minus:              "-",
		percent:            "%",
		perMille:           "‰",
		timeSeparator:      ":",
		inifinity:          "∞",
		currencies:         []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "Kz", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "$", "ATS", "A$", "AWG", "AZM", "AZN", "BAD", "KM", "BAN", "$", "৳", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "$", "$", "Bs", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "R$", "BRN", "BRR", "BRZ", "$", "BTN", "BUK", "P", "BYB", "р.", "BYR", "$", "CA$", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "$", "CNH", "CNX", "CN¥", "$", "COU", "₡", "CSD", "CSK", "$", "$", "CVE", "CYP", "Kč", "DDM", "DEM", "DJF", "kr", "$", "DZD", "ECS", "ECV", "EEK", "E£", "ERN", "ESA", "ESB", "₧", "ETB", "€", "FIM", "$", "£", "FRF", "£", "GEK", "₾", "GHC", "GHS", "£", "GMD", "FG", "GNS", "GQE", "GRD", "Q", "GWE", "GWP", "$", "HK$", "L", "HRD", "kn", "HTG", "Ft", "Rp", "IEP", "ILP", "ILR", "₪", "₹", "IQD", "IRR", "ISJ", "kr", "ITL", "$", "JOD", "JP¥", "KES", "KGS", "៛", "CF", "₩", "KRH", "KRO", "₩", "KWD", "$", "₸", "₭", "L£", "Rs", "$", "LSL", "Lt", "LTT", "LUC", "LUF", "LUL", "Ls", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "Ar", "MGF", "MKD", "MKN", "MLF", "K", "₮", "MOP", "MRO", "MTL", "MTP", "Rs", "MVP", "MVR", "MWK", "MX$", "MXP", "MXV", "RM", "MZE", "MZM", "MZN", "$", "₦", "NIC", "C$", "NLG", "kr", "Rs", "NZ$", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "₱", "Rs", "zł", "PLZ", "PTE", "₲", "QAR", "RHD", "ROL", "lei", "RSD", "₽", "р.", "RF", "SAR", "$", "SCR", "SDD", "SDG", "SDP", "kr", "$", "£", "SIT", "SKK", "SLL", "SOS", "$", "SRG", "£", "Db", "STN", "SUR", "SVC", "£", "SZL", "฿", "TJR", "TJS", "TMM", "TMT", "TND", "T$", "TPE", "TRL", "₺", "$", "NT$", "TZS", "₴", "UAK", "UGS", "UGX", "US$", "USN", "USS", "UYI", "UYP", "$", "UZS", "VEB", "Bs", "₫", "VNN", "VUV", "WST", "FCFA", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "EC$", "XDR", "XEU", "XFO", "XFU", "CFA", "XPD", "CFPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "R", "ZMK", "ZK", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		monthsNarrow:       []string{"", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12"},
		monthsWide:         []string{"", "M01", "M02", "M03", "M04", "M05", "M06", "M07", "M08", "M09", "M10", "M11", "M12"},
		daysNarrow:         []string{"S", "M", "T", "W", "T", "F", "S"},
		daysWide:           []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"},
		periodsAbbreviated: []string{"AM", "PM"},
		periodsNarrow:      []string{"", ""},
		periodsWide:        []string{"", ""},
		erasAbbreviated:    []string{"BCE", "CE"},
		erasNarrow:         []string{"", ""},
		erasWide:           []string{"", ""},
		timezones:          map[string]string{"CLT": "CLT", "ChST": "ChST", "AWST": "AWST", "HEPMX": "HEPMX", "CST": "CST", "PDT": "PDT", "WAST": "WAST", "WITA": "WITA", "ACST": "ACST", "ART": "ART", "HADT": "HADT", "CHADT": "CHADT", "AST": "AST", "WESZ": "WESZ", "JDT": "JDT", "EDT": "EDT", "∅∅∅": "∅∅∅", "HNEG": "HNEG", "HENOMX": "HENOMX", "WIB": "WIB", "BT": "BT", "AKST": "AKST", "HAT": "HAT", "PST": "PST", "IST": "IST", "HNPMX": "HNPMX", "AEST": "AEST", "SAST": "SAST", "NZDT": "NZDT", "ACWDT": "ACWDT", "GYT": "GYT", "WIT": "WIT", "TMST": "TMST", "UYST": "UYST", "CHAST": "CHAST", "HNCU": "HNCU", "AEDT": "AEDT", "ACWST": "ACWST", "WARST": "WARST", "MEZ": "MEZ", "EST": "EST", "ECT": "ECT", "HNOG": "HNOG", "MESZ": "MESZ", "WEZ": "WEZ", "HAST": "HAST", "HEOG": "HEOG", "COT": "COT", "HNNOMX": "HNNOMX", "SRT": "SRT", "OESZ": "OESZ", "ADT": "ADT", "AKDT": "AKDT", "HEEG": "HEEG", "HNPM": "HNPM", "BOT": "BOT", "HECU": "HECU", "HNT": "HNT", "MDT": "MDT", "EAT": "EAT", "OEZ": "OEZ", "HEPM": "HEPM", "AWDT": "AWDT", "COST": "COST", "LHST": "LHST", "LHDT": "LHDT", "MST": "MST", "UYT": "UYT", "GMT": "GMT", "GFT": "GFT", "JST": "JST", "WART": "WART", "ACDT": "ACDT", "NZST": "NZST", "ARST": "ARST", "CDT": "CDT", "MYT": "MYT", "SGT": "SGT", "TMT": "TMT", "CLST": "CLST", "CAT": "CAT", "WAT": "WAT", "HKT": "HKT", "HKST": "HKST", "VET": "VET"},
	}
}

// Locale returns the current translators string locale
func (root *root) Locale() string {
	return root.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'root'
func (root *root) PluralsCardinal() []locales.PluralRule {
	return root.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'root'
func (root *root) PluralsOrdinal() []locales.PluralRule {
	return root.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'root'
func (root *root) PluralsRange() []locales.PluralRule {
	return root.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'root'
func (root *root) CardinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'root'
func (root *root) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'root'
func (root *root) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (root *root) MonthAbbreviated(month time.Month) string {
	return root.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (root *root) MonthsAbbreviated() []string {
	return nil
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (root *root) MonthNarrow(month time.Month) string {
	return root.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (root *root) MonthsNarrow() []string {
	return root.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (root *root) MonthWide(month time.Month) string {
	return root.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (root *root) MonthsWide() []string {
	return root.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (root *root) WeekdayAbbreviated(weekday time.Weekday) string {
	return root.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (root *root) WeekdaysAbbreviated() []string {
	return root.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (root *root) WeekdayNarrow(weekday time.Weekday) string {
	return root.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (root *root) WeekdaysNarrow() []string {
	return root.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (root *root) WeekdayShort(weekday time.Weekday) string {
	return root.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (root *root) WeekdaysShort() []string {
	return root.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (root *root) WeekdayWide(weekday time.Weekday) string {
	return root.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (root *root) WeekdaysWide() []string {
	return root.daysWide
}

// Decimal returns the decimal point of number
func (root *root) Decimal() string {
	return root.decimal
}

// Group returns the group of number
func (root *root) Group() string {
	return root.group
}

// Group returns the minus sign of number
func (root *root) Minus() string {
	return root.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'root' and handles both Whole and Real numbers based on 'v'
func (root *root) FmtNumber(num float64, v uint64) string {

	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'root' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (root *root) FmtPercent(num float64, v uint64) string {
	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'root'
func (root *root) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := root.currencies[currency]
	return string(append(append([]byte{}, symbol...), s...))
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'root'
// in accounting notation.
func (root *root) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := root.currencies[currency]
	return string(append(append([]byte{}, symbol...), s...))
}

// FmtDateShort returns the short date representation of 't' for 'root'
func (root *root) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'root'
func (root *root) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x20}...)
	b = append(b, root.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'root'
func (root *root) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x20}...)
	b = append(b, root.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'root'
func (root *root) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x20}...)
	b = append(b, root.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2c, 0x20}...)
	b = append(b, root.daysWide[t.Weekday()]...)

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'root'
func (root *root) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, root.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'root'
func (root *root) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, root.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, root.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'root'
func (root *root) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, root.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, root.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'root'
func (root *root) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, root.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, root.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := root.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
