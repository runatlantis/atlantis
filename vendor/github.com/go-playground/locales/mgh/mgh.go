package mgh

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type mgh struct {
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

// New returns a new instance of translator for the 'mgh' locale
func New() locales.Translator {
	return &mgh{
		locale:                 "mgh",
		pluralsCardinal:        nil,
		pluralsOrdinal:         nil,
		pluralsRange:           nil,
		decimal:                ",",
		group:                  ".",
		timeSeparator:          ":",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MTn", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
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
		timezones:              map[string]string{"HEPM": "HEPM", "CLST": "CLST", "SAST": "SAST", "WIB": "WIB", "BOT": "BOT", "AKDT": "AKDT", "ChST": "ChST", "PDT": "PDT", "WEZ": "WEZ", "MST": "MST", "CAT": "CAT", "WIT": "WIT", "OESZ": "OESZ", "GYT": "GYT", "HKT": "HKT", "HNT": "HNT", "TMT": "TMT", "MYT": "MYT", "VET": "VET", "TMST": "TMST", "COT": "COT", "GFT": "GFT", "MESZ": "MESZ", "WAST": "WAST", "HNOG": "HNOG", "HENOMX": "HENOMX", "HADT": "HADT", "CHAST": "CHAST", "AWDT": "AWDT", "AEDT": "AEDT", "HEPMX": "HEPMX", "NZDT": "NZDT", "ECT": "ECT", "MDT": "MDT", "OEZ": "OEZ", "ARST": "ARST", "GMT": "GMT", "HECU": "HECU", "ACDT": "ACDT", "HKST": "HKST", "IST": "IST", "LHDT": "LHDT", "COST": "COST", "AWST": "AWST", "EDT": "EDT", "NZST": "NZST", "HEOG": "HEOG", "ACWDT": "ACWDT", "HAST": "HAST", "AST": "AST", "WAT": "WAT", "JDT": "JDT", "BT": "BT", "LHST": "LHST", "HAT": "HAT", "UYST": "UYST", "HNCU": "HNCU", "SGT": "SGT", "WARST": "WARST", "HNNOMX": "HNNOMX", "EAT": "EAT", "HNPMX": "HNPMX", "AKST": "AKST", "WART": "WART", "CLT": "CLT", "CST": "CST", "WESZ": "WESZ", "ACST": "ACST", "MEZ": "MEZ", "∅∅∅": "∅∅∅", "ART": "ART", "UYT": "UYT", "AEST": "AEST", "JST": "JST", "ADT": "ADT", "ACWST": "ACWST", "HNPM": "HNPM", "CHADT": "CHADT", "WITA": "WITA", "HEEG": "HEEG", "SRT": "SRT", "CDT": "CDT", "PST": "PST", "EST": "EST", "HNEG": "HNEG"},
	}
}

// Locale returns the current translators string locale
func (mgh *mgh) Locale() string {
	return mgh.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'mgh'
func (mgh *mgh) PluralsCardinal() []locales.PluralRule {
	return mgh.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'mgh'
func (mgh *mgh) PluralsOrdinal() []locales.PluralRule {
	return mgh.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'mgh'
func (mgh *mgh) PluralsRange() []locales.PluralRule {
	return mgh.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'mgh'
func (mgh *mgh) CardinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'mgh'
func (mgh *mgh) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'mgh'
func (mgh *mgh) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (mgh *mgh) MonthAbbreviated(month time.Month) string {
	return mgh.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (mgh *mgh) MonthsAbbreviated() []string {
	return mgh.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (mgh *mgh) MonthNarrow(month time.Month) string {
	return mgh.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (mgh *mgh) MonthsNarrow() []string {
	return mgh.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (mgh *mgh) MonthWide(month time.Month) string {
	return mgh.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (mgh *mgh) MonthsWide() []string {
	return mgh.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (mgh *mgh) WeekdayAbbreviated(weekday time.Weekday) string {
	return mgh.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (mgh *mgh) WeekdaysAbbreviated() []string {
	return mgh.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (mgh *mgh) WeekdayNarrow(weekday time.Weekday) string {
	return mgh.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (mgh *mgh) WeekdaysNarrow() []string {
	return mgh.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (mgh *mgh) WeekdayShort(weekday time.Weekday) string {
	return mgh.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (mgh *mgh) WeekdaysShort() []string {
	return mgh.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (mgh *mgh) WeekdayWide(weekday time.Weekday) string {
	return mgh.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (mgh *mgh) WeekdaysWide() []string {
	return mgh.daysWide
}

// Decimal returns the decimal point of number
func (mgh *mgh) Decimal() string {
	return mgh.decimal
}

// Group returns the group of number
func (mgh *mgh) Group() string {
	return mgh.group
}

// Group returns the minus sign of number
func (mgh *mgh) Minus() string {
	return mgh.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'mgh' and handles both Whole and Real numbers based on 'v'
func (mgh *mgh) FmtNumber(num float64, v uint64) string {

	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'mgh' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (mgh *mgh) FmtPercent(num float64, v uint64) string {
	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'mgh'
func (mgh *mgh) FmtCurrency(num float64, v uint64, currency currency.Type) string {

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

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'mgh'
// in accounting notation.
func (mgh *mgh) FmtAccounting(num float64, v uint64, currency currency.Type) string {

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

// FmtDateShort returns the short date representation of 't' for 'mgh'
func (mgh *mgh) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'mgh'
func (mgh *mgh) FmtDateMedium(t time.Time) string {

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

// FmtDateLong returns the long date representation of 't' for 'mgh'
func (mgh *mgh) FmtDateLong(t time.Time) string {

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

// FmtDateFull returns the full date representation of 't' for 'mgh'
func (mgh *mgh) FmtDateFull(t time.Time) string {

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

// FmtTimeShort returns the short time representation of 't' for 'mgh'
func (mgh *mgh) FmtTimeShort(t time.Time) string {

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

// FmtTimeMedium returns the medium time representation of 't' for 'mgh'
func (mgh *mgh) FmtTimeMedium(t time.Time) string {

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

// FmtTimeLong returns the long time representation of 't' for 'mgh'
func (mgh *mgh) FmtTimeLong(t time.Time) string {

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

// FmtTimeFull returns the full time representation of 't' for 'mgh'
func (mgh *mgh) FmtTimeFull(t time.Time) string {

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
