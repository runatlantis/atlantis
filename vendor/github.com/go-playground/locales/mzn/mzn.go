package mzn

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type mzn struct {
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

// New returns a new instance of translator for the 'mzn' locale
func New() locales.Translator {
	return &mzn{
		locale:            "mzn",
		pluralsCardinal:   nil,
		pluralsOrdinal:    nil,
		pluralsRange:      nil,
		timeSeparator:     ":",
		currencies:        []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "R$", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CA$", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CN¥", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "€", "FIM", "FJD", "FKP", "FRF", "£", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HK$", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "₪", "₹", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JP¥", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "₩", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MX$", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "NT$", "TZS", "UAH", "UAK", "UGS", "UGX", "$", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "₫", "VNN", "VUV", "WST", "FCFA", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "EC$", "XDR", "XEU", "XFO", "XFU", "CFA", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		monthsAbbreviated: []string{"", "ژانویه", "فوریه", "مارس", "آوریل", "مه", "ژوئن", "ژوئیه", "اوت", "سپتامبر", "اکتبر", "نوامبر", "دسامبر"},
		monthsWide:        []string{"", "ژانویه", "فوریه", "مارس", "آوریل", "مه", "ژوئن", "ژوئیه", "اوت", "سپتامبر", "اکتبر", "نوامبر", "دسامبر"},
		erasAbbreviated:   []string{"پ.م", "م."},
		erasNarrow:        []string{"", ""},
		erasWide:          []string{"قبل میلاد", "بعد میلاد"},
		timezones:         map[string]string{"JDT": "JDT", "NZDT": "NZDT", "HEPM": "HEPM", "HNCU": "HNCU", "AST": "AST", "SAST": "SAST", "HEEG": "HEEG", "ACST": "ACST", "HNPM": "HNPM", "BOT": "BOT", "JST": "JST", "AKDT": "AKDT", "ACWST": "ACWST", "CHADT": "CHADT", "MST": "MST", "WESZ": "WESZ", "EDT": "EDT", "WART": "WART", "COST": "COST", "HEPMX": "HEPMX", "EST": "EST", "CAT": "CAT", "HAST": "HAST", "HADT": "HADT", "AWST": "AWST", "HKT": "HKT", "WARST": "WARST", "TMST": "TMST", "CHAST": "CHAST", "SGT": "SGT", "MESZ": "MESZ", "CST": "CST", "IST": "IST", "UYST": "UYST", "AEDT": "AEDT", "WAT": "WAT", "LHST": "LHST", "HNT": "HNT", "COT": "COT", "∅∅∅": "∅∅∅", "ART": "ART", "WEZ": "WEZ", "MYT": "MYT", "HNOG": "HNOG", "HEOG": "HEOG", "TMT": "TMT", "UYT": "UYT", "ChST": "ChST", "PDT": "PDT", "MEZ": "MEZ", "HAT": "HAT", "VET": "VET", "WIT": "WIT", "AEST": "AEST", "HNEG": "HNEG", "ACDT": "ACDT", "WITA": "WITA", "HENOMX": "HENOMX", "ARST": "ARST", "GMT": "GMT", "GFT": "GFT", "WIB": "WIB", "WAST": "WAST", "NZST": "NZST", "HNNOMX": "HNNOMX", "GYT": "GYT", "PST": "PST", "MDT": "MDT", "AKST": "AKST", "OESZ": "OESZ", "AWDT": "AWDT", "BT": "BT", "HKST": "HKST", "SRT": "SRT", "EAT": "EAT", "CLT": "CLT", "CLST": "CLST", "HECU": "HECU", "CDT": "CDT", "ACWDT": "ACWDT", "ECT": "ECT", "LHDT": "LHDT", "OEZ": "OEZ", "HNPMX": "HNPMX", "ADT": "ADT"},
	}
}

// Locale returns the current translators string locale
func (mzn *mzn) Locale() string {
	return mzn.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'mzn'
func (mzn *mzn) PluralsCardinal() []locales.PluralRule {
	return mzn.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'mzn'
func (mzn *mzn) PluralsOrdinal() []locales.PluralRule {
	return mzn.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'mzn'
func (mzn *mzn) PluralsRange() []locales.PluralRule {
	return mzn.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'mzn'
func (mzn *mzn) CardinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'mzn'
func (mzn *mzn) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'mzn'
func (mzn *mzn) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (mzn *mzn) MonthAbbreviated(month time.Month) string {
	return mzn.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (mzn *mzn) MonthsAbbreviated() []string {
	return mzn.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (mzn *mzn) MonthNarrow(month time.Month) string {
	return mzn.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (mzn *mzn) MonthsNarrow() []string {
	return nil
}

// MonthWide returns the locales wide month given the 'month' provided
func (mzn *mzn) MonthWide(month time.Month) string {
	return mzn.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (mzn *mzn) MonthsWide() []string {
	return mzn.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (mzn *mzn) WeekdayAbbreviated(weekday time.Weekday) string {
	return mzn.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (mzn *mzn) WeekdaysAbbreviated() []string {
	return mzn.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (mzn *mzn) WeekdayNarrow(weekday time.Weekday) string {
	return mzn.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (mzn *mzn) WeekdaysNarrow() []string {
	return mzn.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (mzn *mzn) WeekdayShort(weekday time.Weekday) string {
	return mzn.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (mzn *mzn) WeekdaysShort() []string {
	return mzn.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (mzn *mzn) WeekdayWide(weekday time.Weekday) string {
	return mzn.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (mzn *mzn) WeekdaysWide() []string {
	return mzn.daysWide
}

// Decimal returns the decimal point of number
func (mzn *mzn) Decimal() string {
	return mzn.decimal
}

// Group returns the group of number
func (mzn *mzn) Group() string {
	return mzn.group
}

// Group returns the minus sign of number
func (mzn *mzn) Minus() string {
	return mzn.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'mzn' and handles both Whole and Real numbers based on 'v'
func (mzn *mzn) FmtNumber(num float64, v uint64) string {

	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'mzn' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (mzn *mzn) FmtPercent(num float64, v uint64) string {
	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'mzn'
func (mzn *mzn) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := mzn.currencies[currency]
	return string(append(append([]byte{}, symbol...), s...))
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'mzn'
// in accounting notation.
func (mzn *mzn) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := mzn.currencies[currency]
	return string(append(append([]byte{}, symbol...), s...))
}

// FmtDateShort returns the short date representation of 't' for 'mzn'
func (mzn *mzn) FmtDateShort(t time.Time) string {

	b := make([]byte, 0, 32)

	return string(b)
}

// FmtDateMedium returns the medium date representation of 't' for 'mzn'
func (mzn *mzn) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'mzn'
func (mzn *mzn) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'mzn'
func (mzn *mzn) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'mzn'
func (mzn *mzn) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'mzn'
func (mzn *mzn) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'mzn'
func (mzn *mzn) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'mzn'
func (mzn *mzn) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	return string(b)
}
