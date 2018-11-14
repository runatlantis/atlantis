package kkj

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type kkj struct {
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

// New returns a new instance of translator for the 'kkj' locale
func New() locales.Translator {
	return &kkj{
		locale:                 "kkj",
		pluralsCardinal:        []locales.PluralRule{2, 6},
		pluralsOrdinal:         nil,
		pluralsRange:           nil,
		decimal:                ",",
		group:                  ".",
		timeSeparator:          ":",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "FCFA", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyPositivePrefix: " ",
		currencyPositiveSuffix: "K",
		currencyNegativePrefix: " ",
		currencyNegativeSuffix: "K",
		monthsWide:             []string{"", "pamba", "wanja", "mbiyɔ mɛndoŋgɔ", "Nyɔlɔmbɔŋgɔ", "Mɔnɔ ŋgbanja", "Nyaŋgwɛ ŋgbanja", "kuŋgwɛ", "fɛ", "njapi", "nyukul", "11", "ɓulɓusɛ"},
		daysAbbreviated:        []string{"sɔndi", "lundi", "mardi", "mɛrkɛrɛdi", "yedi", "vaŋdɛrɛdi", "mɔnɔ sɔndi"},
		daysNarrow:             []string{"so", "lu", "ma", "mɛ", "ye", "va", "ms"},
		daysShort:              []string{"sɔndi", "lundi", "mardi", "mɛrkɛrɛdi", "yedi", "vaŋdɛrɛdi", "mɔnɔ sɔndi"},
		daysWide:               []string{"sɔndi", "lundi", "mardi", "mɛrkɛrɛdi", "yedi", "vaŋdɛrɛdi", "mɔnɔ sɔndi"},
		timezones:              map[string]string{"CST": "CST", "WIB": "WIB", "MYT": "MYT", "LHST": "LHST", "WART": "WART", "HEPM": "HEPM", "HEPMX": "HEPMX", "CHADT": "CHADT", "GMT": "GMT", "PST": "PST", "ACST": "ACST", "HNEG": "HNEG", "MESZ": "MESZ", "OESZ": "OESZ", "SRT": "SRT", "TMST": "TMST", "OEZ": "OEZ", "HNPMX": "HNPMX", "WAST": "WAST", "SAST": "SAST", "AKDT": "AKDT", "HNNOMX": "HNNOMX", "EAT": "EAT", "CLST": "CLST", "WAT": "WAT", "COT": "COT", "AWST": "AWST", "AEST": "AEST", "HKT": "HKT", "MEZ": "MEZ", "HAT": "HAT", "ARST": "ARST", "HNCU": "HNCU", "JST": "JST", "SGT": "SGT", "HNOG": "HNOG", "WARST": "WARST", "HADT": "HADT", "UYT": "UYT", "NZDT": "NZDT", "ACWST": "ACWST", "HNPM": "HNPM", "CLT": "CLT", "WESZ": "WESZ", "ACWDT": "ACWDT", "HNT": "HNT", "TMT": "TMT", "AKST": "AKST", "AEDT": "AEDT", "WEZ": "WEZ", "EDT": "EDT", "ART": "ART", "ChST": "ChST", "AWDT": "AWDT", "VET": "VET", "MDT": "MDT", "BT": "BT", "GFT": "GFT", "HEOG": "HEOG", "MST": "MST", "CAT": "CAT", "HAST": "HAST", "ECT": "ECT", "HENOMX": "HENOMX", "UYST": "UYST", "CDT": "CDT", "ADT": "ADT", "JDT": "JDT", "EST": "EST", "LHDT": "LHDT", "WIT": "WIT", "PDT": "PDT", "HKST": "HKST", "IST": "IST", "COST": "COST", "AST": "AST", "GYT": "GYT", "HECU": "HECU", "CHAST": "CHAST", "NZST": "NZST", "ACDT": "ACDT", "HEEG": "HEEG", "∅∅∅": "∅∅∅", "WITA": "WITA", "BOT": "BOT"},
	}
}

// Locale returns the current translators string locale
func (kkj *kkj) Locale() string {
	return kkj.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'kkj'
func (kkj *kkj) PluralsCardinal() []locales.PluralRule {
	return kkj.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'kkj'
func (kkj *kkj) PluralsOrdinal() []locales.PluralRule {
	return kkj.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'kkj'
func (kkj *kkj) PluralsRange() []locales.PluralRule {
	return kkj.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'kkj'
func (kkj *kkj) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'kkj'
func (kkj *kkj) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'kkj'
func (kkj *kkj) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (kkj *kkj) MonthAbbreviated(month time.Month) string {
	return kkj.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (kkj *kkj) MonthsAbbreviated() []string {
	return nil
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (kkj *kkj) MonthNarrow(month time.Month) string {
	return kkj.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (kkj *kkj) MonthsNarrow() []string {
	return nil
}

// MonthWide returns the locales wide month given the 'month' provided
func (kkj *kkj) MonthWide(month time.Month) string {
	return kkj.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (kkj *kkj) MonthsWide() []string {
	return kkj.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (kkj *kkj) WeekdayAbbreviated(weekday time.Weekday) string {
	return kkj.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (kkj *kkj) WeekdaysAbbreviated() []string {
	return kkj.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (kkj *kkj) WeekdayNarrow(weekday time.Weekday) string {
	return kkj.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (kkj *kkj) WeekdaysNarrow() []string {
	return kkj.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (kkj *kkj) WeekdayShort(weekday time.Weekday) string {
	return kkj.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (kkj *kkj) WeekdaysShort() []string {
	return kkj.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (kkj *kkj) WeekdayWide(weekday time.Weekday) string {
	return kkj.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (kkj *kkj) WeekdaysWide() []string {
	return kkj.daysWide
}

// Decimal returns the decimal point of number
func (kkj *kkj) Decimal() string {
	return kkj.decimal
}

// Group returns the group of number
func (kkj *kkj) Group() string {
	return kkj.group
}

// Group returns the minus sign of number
func (kkj *kkj) Minus() string {
	return kkj.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'kkj' and handles both Whole and Real numbers based on 'v'
func (kkj *kkj) FmtNumber(num float64, v uint64) string {

	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'kkj' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (kkj *kkj) FmtPercent(num float64, v uint64) string {
	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'kkj'
func (kkj *kkj) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := kkj.currencies[currency]
	l := len(s) + len(symbol) + 4

	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, kkj.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	for j := len(symbol) - 1; j >= 0; j-- {
		b = append(b, symbol[j])
	}

	for j := len(kkj.currencyPositivePrefix) - 1; j >= 0; j-- {
		b = append(b, kkj.currencyPositivePrefix[j])
	}

	if num < 0 {
		b = append(b, kkj.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, kkj.currencyPositiveSuffix...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'kkj'
// in accounting notation.
func (kkj *kkj) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := kkj.currencies[currency]
	l := len(s) + len(symbol) + 4

	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, kkj.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {

		for j := len(symbol) - 1; j >= 0; j-- {
			b = append(b, symbol[j])
		}

		for j := len(kkj.currencyNegativePrefix) - 1; j >= 0; j-- {
			b = append(b, kkj.currencyNegativePrefix[j])
		}

		b = append(b, kkj.minus[0])

	} else {

		for j := len(symbol) - 1; j >= 0; j-- {
			b = append(b, symbol[j])
		}

		for j := len(kkj.currencyPositivePrefix) - 1; j >= 0; j-- {
			b = append(b, kkj.currencyPositivePrefix[j])
		}

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if num < 0 {
		b = append(b, kkj.currencyNegativeSuffix...)
	} else {

		b = append(b, kkj.currencyPositiveSuffix...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'kkj'
func (kkj *kkj) FmtDateShort(t time.Time) string {

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

	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateMedium returns the medium date representation of 't' for 'kkj'
func (kkj *kkj) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, kkj.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'kkj'
func (kkj *kkj) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, kkj.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'kkj'
func (kkj *kkj) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, kkj.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x20}...)

	if t.Day() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, kkj.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'kkj'
func (kkj *kkj) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, kkj.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'kkj'
func (kkj *kkj) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, kkj.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, kkj.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'kkj'
func (kkj *kkj) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'kkj'
func (kkj *kkj) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	return string(b)
}
