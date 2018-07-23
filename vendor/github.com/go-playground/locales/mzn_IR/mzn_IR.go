package mzn_IR

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type mzn_IR struct {
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

// New returns a new instance of translator for the 'mzn_IR' locale
func New() locales.Translator {
	return &mzn_IR{
		locale:            "mzn_IR",
		pluralsCardinal:   nil,
		pluralsOrdinal:    nil,
		pluralsRange:      nil,
		timeSeparator:     ":",
		currencies:        []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		monthsAbbreviated: []string{"", "ژانویه", "فوریه", "مارس", "آوریل", "مه", "ژوئن", "ژوئیه", "اوت", "سپتامبر", "اکتبر", "نوامبر", "دسامبر"},
		monthsWide:        []string{"", "ژانویه", "فوریه", "مارس", "آوریل", "مه", "ژوئن", "ژوئیه", "اوت", "سپتامبر", "اکتبر", "نوامبر", "دسامبر"},
		erasAbbreviated:   []string{"پ.م", "م."},
		erasNarrow:        []string{"", ""},
		erasWide:          []string{"قبل میلاد", "بعد میلاد"},
		timezones:         map[string]string{"NZDT": "NZDT", "MESZ": "MESZ", "AWST": "AWST", "EAT": "EAT", "COT": "COT", "HNPMX": "HNPMX", "SAST": "SAST", "WEZ": "WEZ", "AKST": "AKST", "HEPM": "HEPM", "TMT": "TMT", "HAST": "HAST", "GMT": "GMT", "HECU": "HECU", "AEST": "AEST", "IST": "IST", "LHST": "LHST", "HENOMX": "HENOMX", "SRT": "SRT", "ChST": "ChST", "ECT": "ECT", "MEZ": "MEZ", "WART": "WART", "WARST": "WARST", "∅∅∅": "∅∅∅", "GYT": "GYT", "CHADT": "CHADT", "CDT": "CDT", "ACWDT": "ACWDT", "HNPM": "HNPM", "TMST": "TMST", "OEZ": "OEZ", "ART": "ART", "BOT": "BOT", "MDT": "MDT", "HNCU": "HNCU", "HEPMX": "HEPMX", "AEDT": "AEDT", "HEOG": "HEOG", "EST": "EST", "HEEG": "HEEG", "HAT": "HAT", "CAT": "CAT", "HKT": "HKT", "PDT": "PDT", "HNNOMX": "HNNOMX", "CLST": "CLST", "GFT": "GFT", "SGT": "SGT", "EDT": "EDT", "WESZ": "WESZ", "WIT": "WIT", "COST": "COST", "ADT": "ADT", "WAT": "WAT", "WAST": "WAST", "AKDT": "AKDT", "ACDT": "ACDT", "CLT": "CLT", "LHDT": "LHDT", "WITA": "WITA", "ACWST": "ACWST", "WIB": "WIB", "NZST": "NZST", "MYT": "MYT", "HNT": "HNT", "CST": "CST", "PST": "PST", "ARST": "ARST", "UYT": "UYT", "UYST": "UYST", "CHAST": "CHAST", "JST": "JST", "JDT": "JDT", "BT": "BT", "HKST": "HKST", "HADT": "HADT", "VET": "VET", "AWDT": "AWDT", "HNOG": "HNOG", "OESZ": "OESZ", "AST": "AST", "ACST": "ACST", "HNEG": "HNEG", "MST": "MST"},
	}
}

// Locale returns the current translators string locale
func (mzn *mzn_IR) Locale() string {
	return mzn.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'mzn_IR'
func (mzn *mzn_IR) PluralsCardinal() []locales.PluralRule {
	return mzn.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'mzn_IR'
func (mzn *mzn_IR) PluralsOrdinal() []locales.PluralRule {
	return mzn.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'mzn_IR'
func (mzn *mzn_IR) PluralsRange() []locales.PluralRule {
	return mzn.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'mzn_IR'
func (mzn *mzn_IR) CardinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'mzn_IR'
func (mzn *mzn_IR) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'mzn_IR'
func (mzn *mzn_IR) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (mzn *mzn_IR) MonthAbbreviated(month time.Month) string {
	return mzn.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (mzn *mzn_IR) MonthsAbbreviated() []string {
	return mzn.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (mzn *mzn_IR) MonthNarrow(month time.Month) string {
	return mzn.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (mzn *mzn_IR) MonthsNarrow() []string {
	return nil
}

// MonthWide returns the locales wide month given the 'month' provided
func (mzn *mzn_IR) MonthWide(month time.Month) string {
	return mzn.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (mzn *mzn_IR) MonthsWide() []string {
	return mzn.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (mzn *mzn_IR) WeekdayAbbreviated(weekday time.Weekday) string {
	return mzn.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (mzn *mzn_IR) WeekdaysAbbreviated() []string {
	return mzn.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (mzn *mzn_IR) WeekdayNarrow(weekday time.Weekday) string {
	return mzn.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (mzn *mzn_IR) WeekdaysNarrow() []string {
	return mzn.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (mzn *mzn_IR) WeekdayShort(weekday time.Weekday) string {
	return mzn.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (mzn *mzn_IR) WeekdaysShort() []string {
	return mzn.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (mzn *mzn_IR) WeekdayWide(weekday time.Weekday) string {
	return mzn.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (mzn *mzn_IR) WeekdaysWide() []string {
	return mzn.daysWide
}

// Decimal returns the decimal point of number
func (mzn *mzn_IR) Decimal() string {
	return mzn.decimal
}

// Group returns the group of number
func (mzn *mzn_IR) Group() string {
	return mzn.group
}

// Group returns the minus sign of number
func (mzn *mzn_IR) Minus() string {
	return mzn.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'mzn_IR' and handles both Whole and Real numbers based on 'v'
func (mzn *mzn_IR) FmtNumber(num float64, v uint64) string {

	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'mzn_IR' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (mzn *mzn_IR) FmtPercent(num float64, v uint64) string {
	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'mzn_IR'
func (mzn *mzn_IR) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := mzn.currencies[currency]
	return string(append(append([]byte{}, symbol...), s...))
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'mzn_IR'
// in accounting notation.
func (mzn *mzn_IR) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := mzn.currencies[currency]
	return string(append(append([]byte{}, symbol...), s...))
}

// FmtDateShort returns the short date representation of 't' for 'mzn_IR'
func (mzn *mzn_IR) FmtDateShort(t time.Time) string {

	b := make([]byte, 0, 32)

	return string(b)
}

// FmtDateMedium returns the medium date representation of 't' for 'mzn_IR'
func (mzn *mzn_IR) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'mzn_IR'
func (mzn *mzn_IR) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'mzn_IR'
func (mzn *mzn_IR) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'mzn_IR'
func (mzn *mzn_IR) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'mzn_IR'
func (mzn *mzn_IR) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'mzn_IR'
func (mzn *mzn_IR) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'mzn_IR'
func (mzn *mzn_IR) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	return string(b)
}
