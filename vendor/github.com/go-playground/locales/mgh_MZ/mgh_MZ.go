package mgh_MZ

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type mgh_MZ struct {
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

// New returns a new instance of translator for the 'mgh_MZ' locale
func New() locales.Translator {
	return &mgh_MZ{
		locale:                 "mgh_MZ",
		pluralsCardinal:        nil,
		pluralsOrdinal:         nil,
		pluralsRange:           nil,
		decimal:                ",",
		group:                  ".",
		timeSeparator:          ":",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyPositivePrefix: " ",
		currencyPositiveSuffix: "K",
		currencyNegativePrefix: " ",
		currencyNegativeSuffix: "K",
		monthsAbbreviated:      []string{"", "Kwa", "Una", "Rar", "Che", "Tha", "Moc", "Sab", "Nan", "Tis", "Kum", "Moj", "Yel"},
		monthsNarrow:           []string{"", "K", "U", "R", "C", "T", "M", "S", "N", "T", "K", "M", "Y"},
		monthsWide:             []string{"", "Mweri wo kwanza", "Mweri wo unayeli", "Mweri wo uneraru", "Mweri wo unecheshe", "Mweri wo unethanu", "Mweri wo thanu na mocha", "Mweri wo saba", "Mweri wo nane", "Mweri wo tisa", "Mweri wo kumi", "Mweri wo kumi na moja", "Mweri wo kumi na yel’li"},
		daysAbbreviated:        []string{"Sab", "Jtt", "Jnn", "Jtn", "Ara", "Iju", "Jmo"},
		daysNarrow:             []string{"S", "J", "J", "J", "A", "I", "J"},
		daysWide:               []string{"Sabato", "Jumatatu", "Jumanne", "Jumatano", "Arahamisi", "Ijumaa", "Jumamosi"},
		periodsAbbreviated:     []string{"wichishu", "mchochil’l"},
		periodsWide:            []string{"wichishu", "mchochil’l"},
		erasAbbreviated:        []string{"HY", "YY"},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"Hinapiya yesu", "Yopia yesu"},
		timezones:              map[string]string{"OESZ": "OESZ", "GYT": "GYT", "AEDT": "AEDT", "WIB": "WIB", "EST": "EST", "HNPM": "HNPM", "HENOMX": "HENOMX", "ARST": "ARST", "CHAST": "CHAST", "AEST": "AEST", "JST": "JST", "JDT": "JDT", "HEOG": "HEOG", "HKST": "HKST", "VET": "VET", "CLT": "CLT", "CDT": "CDT", "AWDT": "AWDT", "WARST": "WARST", "BOT": "BOT", "ACST": "ACST", "CLST": "CLST", "OEZ": "OEZ", "EAT": "EAT", "HADT": "HADT", "UYT": "UYT", "HECU": "HECU", "AWST": "AWST", "BT": "BT", "HEPM": "HEPM", "TMST": "TMST", "ACWDT": "ACWDT", "IST": "IST", "HNT": "HNT", "HNPMX": "HNPMX", "HEPMX": "HEPMX", "MDT": "MDT", "WAT": "WAT", "AKST": "AKST", "UYST": "UYST", "HAST": "HAST", "CHADT": "CHADT", "AST": "AST", "GFT": "GFT", "NZST": "NZST", "MESZ": "MESZ", "HNEG": "HNEG", "HAT": "HAT", "WEZ": "WEZ", "WESZ": "WESZ", "NZDT": "NZDT", "ECT": "ECT", "ACWST": "ACWST", "HNOG": "HNOG", "HEEG": "HEEG", "SRT": "SRT", "WIT": "WIT", "LHST": "LHST", "COT": "COT", "CST": "CST", "ADT": "ADT", "AKDT": "AKDT", "EDT": "EDT", "HKT": "HKT", "ART": "ART", "ChST": "ChST", "MST": "MST", "WAST": "WAST", "WART": "WART", "WITA": "WITA", "TMT": "TMT", "COST": "COST", "HNCU": "HNCU", "MYT": "MYT", "SGT": "SGT", "ACDT": "ACDT", "MEZ": "MEZ", "PST": "PST", "LHDT": "LHDT", "HNNOMX": "HNNOMX", "CAT": "CAT", "GMT": "GMT", "∅∅∅": "∅∅∅", "PDT": "PDT", "SAST": "SAST"},
	}
}

// Locale returns the current translators string locale
func (mgh *mgh_MZ) Locale() string {
	return mgh.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'mgh_MZ'
func (mgh *mgh_MZ) PluralsCardinal() []locales.PluralRule {
	return mgh.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'mgh_MZ'
func (mgh *mgh_MZ) PluralsOrdinal() []locales.PluralRule {
	return mgh.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'mgh_MZ'
func (mgh *mgh_MZ) PluralsRange() []locales.PluralRule {
	return mgh.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'mgh_MZ'
func (mgh *mgh_MZ) CardinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'mgh_MZ'
func (mgh *mgh_MZ) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'mgh_MZ'
func (mgh *mgh_MZ) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (mgh *mgh_MZ) MonthAbbreviated(month time.Month) string {
	return mgh.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (mgh *mgh_MZ) MonthsAbbreviated() []string {
	return mgh.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (mgh *mgh_MZ) MonthNarrow(month time.Month) string {
	return mgh.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (mgh *mgh_MZ) MonthsNarrow() []string {
	return mgh.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (mgh *mgh_MZ) MonthWide(month time.Month) string {
	return mgh.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (mgh *mgh_MZ) MonthsWide() []string {
	return mgh.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (mgh *mgh_MZ) WeekdayAbbreviated(weekday time.Weekday) string {
	return mgh.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (mgh *mgh_MZ) WeekdaysAbbreviated() []string {
	return mgh.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (mgh *mgh_MZ) WeekdayNarrow(weekday time.Weekday) string {
	return mgh.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (mgh *mgh_MZ) WeekdaysNarrow() []string {
	return mgh.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (mgh *mgh_MZ) WeekdayShort(weekday time.Weekday) string {
	return mgh.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (mgh *mgh_MZ) WeekdaysShort() []string {
	return mgh.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (mgh *mgh_MZ) WeekdayWide(weekday time.Weekday) string {
	return mgh.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (mgh *mgh_MZ) WeekdaysWide() []string {
	return mgh.daysWide
}

// Decimal returns the decimal point of number
func (mgh *mgh_MZ) Decimal() string {
	return mgh.decimal
}

// Group returns the group of number
func (mgh *mgh_MZ) Group() string {
	return mgh.group
}

// Group returns the minus sign of number
func (mgh *mgh_MZ) Minus() string {
	return mgh.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'mgh_MZ' and handles both Whole and Real numbers based on 'v'
func (mgh *mgh_MZ) FmtNumber(num float64, v uint64) string {

	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'mgh_MZ' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (mgh *mgh_MZ) FmtPercent(num float64, v uint64) string {
	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'mgh_MZ'
func (mgh *mgh_MZ) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := mgh.currencies[currency]
	l := len(s) + len(symbol) + 4

	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, mgh.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	for j := len(symbol) - 1; j >= 0; j-- {
		b = append(b, symbol[j])
	}

	for j := len(mgh.currencyPositivePrefix) - 1; j >= 0; j-- {
		b = append(b, mgh.currencyPositivePrefix[j])
	}

	if num < 0 {
		b = append(b, mgh.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, mgh.currencyPositiveSuffix...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'mgh_MZ'
// in accounting notation.
func (mgh *mgh_MZ) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := mgh.currencies[currency]
	l := len(s) + len(symbol) + 4

	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, mgh.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {

		for j := len(symbol) - 1; j >= 0; j-- {
			b = append(b, symbol[j])
		}

		for j := len(mgh.currencyNegativePrefix) - 1; j >= 0; j-- {
			b = append(b, mgh.currencyNegativePrefix[j])
		}

		b = append(b, mgh.minus[0])

	} else {

		for j := len(symbol) - 1; j >= 0; j-- {
			b = append(b, symbol[j])
		}

		for j := len(mgh.currencyPositivePrefix) - 1; j >= 0; j-- {
			b = append(b, mgh.currencyPositivePrefix[j])
		}

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if num < 0 {
		b = append(b, mgh.currencyNegativeSuffix...)
	} else {

		b = append(b, mgh.currencyPositiveSuffix...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'mgh_MZ'
func (mgh *mgh_MZ) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'mgh_MZ'
func (mgh *mgh_MZ) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, mgh.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'mgh_MZ'
func (mgh *mgh_MZ) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, mgh.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'mgh_MZ'
func (mgh *mgh_MZ) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, mgh.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, mgh.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'mgh_MZ'
func (mgh *mgh_MZ) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, mgh.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'mgh_MZ'
func (mgh *mgh_MZ) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, mgh.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, mgh.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'mgh_MZ'
func (mgh *mgh_MZ) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, mgh.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, mgh.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'mgh_MZ'
func (mgh *mgh_MZ) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, mgh.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, mgh.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := mgh.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
