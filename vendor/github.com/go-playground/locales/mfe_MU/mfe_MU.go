package mfe_MU

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type mfe_MU struct {
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

// New returns a new instance of translator for the 'mfe_MU' locale
func New() locales.Translator {
	return &mfe_MU{
		locale:                 "mfe_MU",
		pluralsCardinal:        nil,
		pluralsOrdinal:         nil,
		pluralsRange:           nil,
		group:                  " ",
		timeSeparator:          ":",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyPositivePrefix: " ",
		currencyPositiveSuffix: "K",
		currencyNegativePrefix: " ",
		currencyNegativeSuffix: "K",
		monthsAbbreviated:      []string{"", "zan", "fev", "mar", "avr", "me", "zin", "zil", "out", "sep", "okt", "nov", "des"},
		monthsNarrow:           []string{"", "z", "f", "m", "a", "m", "z", "z", "o", "s", "o", "n", "d"},
		monthsWide:             []string{"", "zanvie", "fevriye", "mars", "avril", "me", "zin", "zilye", "out", "septam", "oktob", "novam", "desam"},
		daysAbbreviated:        []string{"dim", "lin", "mar", "mer", "ze", "van", "sam"},
		daysNarrow:             []string{"d", "l", "m", "m", "z", "v", "s"},
		daysWide:               []string{"dimans", "lindi", "mardi", "merkredi", "zedi", "vandredi", "samdi"},
		erasAbbreviated:        []string{"av. Z-K", "ap. Z-K"},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"avan Zezi-Krist", "apre Zezi-Krist"},
		timezones:              map[string]string{"CLT": "CLT", "CLST": "CLST", "AWDT": "AWDT", "ADT": "ADT", "WESZ": "WESZ", "HNNOMX": "HNNOMX", "UYST": "UYST", "HECU": "HECU", "WEZ": "WEZ", "GFT": "GFT", "JDT": "JDT", "AKST": "AKST", "HKT": "HKT", "CHADT": "CHADT", "NZST": "NZST", "ACDT": "ACDT", "MST": "MST", "HAST": "HAST", "WAST": "WAST", "SGT": "SGT", "HEOG": "HEOG", "EST": "EST", "EAT": "EAT", "WART": "WART", "IST": "IST", "MDT": "MDT", "OESZ": "OESZ", "HNEG": "HNEG", "LHDT": "LHDT", "TMST": "TMST", "HADT": "HADT", "COT": "COT", "HNPMX": "HNPMX", "BOT": "BOT", "HAT": "HAT", "WARST": "WARST", "PDT": "PDT", "WIB": "WIB", "MYT": "MYT", "LHST": "LHST", "CHAST": "CHAST", "HNCU": "HNCU", "ACWST": "ACWST", "HNT": "HNT", "HENOMX": "HENOMX", "HEPMX": "HEPMX", "HNOG": "HNOG", "MESZ": "MESZ", "∅∅∅": "∅∅∅", "HNPM": "HNPM", "HEPM": "HEPM", "UYT": "UYT", "AWST": "AWST", "NZDT": "NZDT", "AKDT": "AKDT", "MEZ": "MEZ", "EDT": "EDT", "SRT": "SRT", "TMT": "TMT", "OEZ": "OEZ", "ART": "ART", "ARST": "ARST", "SAST": "SAST", "ACST": "ACST", "WITA": "WITA", "HEEG": "HEEG", "ACWDT": "ACWDT", "WIT": "WIT", "GMT": "GMT", "PST": "PST", "AEDT": "AEDT", "HKST": "HKST", "GYT": "GYT", "ChST": "ChST", "WAT": "WAT", "JST": "JST", "BT": "BT", "COST": "COST", "CAT": "CAT", "CST": "CST", "CDT": "CDT", "AST": "AST", "AEST": "AEST", "ECT": "ECT", "VET": "VET"},
	}
}

// Locale returns the current translators string locale
func (mfe *mfe_MU) Locale() string {
	return mfe.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'mfe_MU'
func (mfe *mfe_MU) PluralsCardinal() []locales.PluralRule {
	return mfe.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'mfe_MU'
func (mfe *mfe_MU) PluralsOrdinal() []locales.PluralRule {
	return mfe.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'mfe_MU'
func (mfe *mfe_MU) PluralsRange() []locales.PluralRule {
	return mfe.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'mfe_MU'
func (mfe *mfe_MU) CardinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'mfe_MU'
func (mfe *mfe_MU) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'mfe_MU'
func (mfe *mfe_MU) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (mfe *mfe_MU) MonthAbbreviated(month time.Month) string {
	return mfe.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (mfe *mfe_MU) MonthsAbbreviated() []string {
	return mfe.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (mfe *mfe_MU) MonthNarrow(month time.Month) string {
	return mfe.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (mfe *mfe_MU) MonthsNarrow() []string {
	return mfe.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (mfe *mfe_MU) MonthWide(month time.Month) string {
	return mfe.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (mfe *mfe_MU) MonthsWide() []string {
	return mfe.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (mfe *mfe_MU) WeekdayAbbreviated(weekday time.Weekday) string {
	return mfe.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (mfe *mfe_MU) WeekdaysAbbreviated() []string {
	return mfe.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (mfe *mfe_MU) WeekdayNarrow(weekday time.Weekday) string {
	return mfe.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (mfe *mfe_MU) WeekdaysNarrow() []string {
	return mfe.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (mfe *mfe_MU) WeekdayShort(weekday time.Weekday) string {
	return mfe.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (mfe *mfe_MU) WeekdaysShort() []string {
	return mfe.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (mfe *mfe_MU) WeekdayWide(weekday time.Weekday) string {
	return mfe.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (mfe *mfe_MU) WeekdaysWide() []string {
	return mfe.daysWide
}

// Decimal returns the decimal point of number
func (mfe *mfe_MU) Decimal() string {
	return mfe.decimal
}

// Group returns the group of number
func (mfe *mfe_MU) Group() string {
	return mfe.group
}

// Group returns the minus sign of number
func (mfe *mfe_MU) Minus() string {
	return mfe.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'mfe_MU' and handles both Whole and Real numbers based on 'v'
func (mfe *mfe_MU) FmtNumber(num float64, v uint64) string {

	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'mfe_MU' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (mfe *mfe_MU) FmtPercent(num float64, v uint64) string {
	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'mfe_MU'
func (mfe *mfe_MU) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := mfe.currencies[currency]
	l := len(s) + len(symbol) + 3

	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, mfe.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	for j := len(symbol) - 1; j >= 0; j-- {
		b = append(b, symbol[j])
	}

	for j := len(mfe.currencyPositivePrefix) - 1; j >= 0; j-- {
		b = append(b, mfe.currencyPositivePrefix[j])
	}

	if num < 0 {
		b = append(b, mfe.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, mfe.currencyPositiveSuffix...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'mfe_MU'
// in accounting notation.
func (mfe *mfe_MU) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := mfe.currencies[currency]
	l := len(s) + len(symbol) + 3

	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, mfe.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {

		for j := len(symbol) - 1; j >= 0; j-- {
			b = append(b, symbol[j])
		}

		for j := len(mfe.currencyNegativePrefix) - 1; j >= 0; j-- {
			b = append(b, mfe.currencyNegativePrefix[j])
		}

		b = append(b, mfe.minus[0])

	} else {

		for j := len(symbol) - 1; j >= 0; j-- {
			b = append(b, symbol[j])
		}

		for j := len(mfe.currencyPositivePrefix) - 1; j >= 0; j-- {
			b = append(b, mfe.currencyPositivePrefix[j])
		}

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if num < 0 {
		b = append(b, mfe.currencyNegativeSuffix...)
	} else {

		b = append(b, mfe.currencyPositiveSuffix...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'mfe_MU'
func (mfe *mfe_MU) FmtDateShort(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2f}...)
	b = strconv.AppendInt(b, int64(t.Month()), 10)
	b = append(b, []byte{0x2f}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateMedium returns the medium date representation of 't' for 'mfe_MU'
func (mfe *mfe_MU) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, mfe.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x2c, 0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'mfe_MU'
func (mfe *mfe_MU) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, mfe.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'mfe_MU'
func (mfe *mfe_MU) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, mfe.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, mfe.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'mfe_MU'
func (mfe *mfe_MU) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, mfe.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'mfe_MU'
func (mfe *mfe_MU) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, mfe.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, mfe.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'mfe_MU'
func (mfe *mfe_MU) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, mfe.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, mfe.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'mfe_MU'
func (mfe *mfe_MU) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, mfe.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, mfe.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := mfe.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
