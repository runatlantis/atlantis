package wae_CH

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type wae_CH struct {
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

// New returns a new instance of translator for the 'wae_CH' locale
func New() locales.Translator {
	return &wae_CH{
		locale:                 "wae_CH",
		pluralsCardinal:        []locales.PluralRule{2, 6},
		pluralsOrdinal:         nil,
		pluralsRange:           nil,
		decimal:                ",",
		group:                  "’",
		timeSeparator:          ":",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyPositivePrefix: " ",
		currencyPositiveSuffix: "K",
		currencyNegativePrefix: " ",
		currencyNegativeSuffix: "K",
		monthsAbbreviated:      []string{"", "Jen", "Hor", "Mär", "Abr", "Mei", "Brá", "Hei", "Öig", "Her", "Wím", "Win", "Chr"},
		monthsNarrow:           []string{"", "J", "H", "M", "A", "M", "B", "H", "Ö", "H", "W", "W", "C"},
		monthsWide:             []string{"", "Jenner", "Hornig", "Märze", "Abrille", "Meije", "Bráčet", "Heiwet", "Öigšte", "Herbštmánet", "Wímánet", "Wintermánet", "Chrištmánet"},
		daysAbbreviated:        []string{"Sun", "Män", "Ziš", "Mit", "Fró", "Fri", "Sam"},
		daysNarrow:             []string{"S", "M", "Z", "M", "F", "F", "S"},
		daysWide:               []string{"Sunntag", "Mäntag", "Zištag", "Mittwuč", "Fróntag", "Fritag", "Samštag"},
		erasAbbreviated:        []string{"v. Chr.", "n. Chr"},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"", ""},
		timezones:              map[string]string{"HNOG": "HNOG", "WAT": "WAT", "WEZ": "Wešteuropäiši Standardzit", "WESZ": "Wešteuropäiši Summerzit", "AKST": "AKST", "OEZ": "Ošteuropäiši Standardzit", "CAT": "CAT", "CLT": "CLT", "CLST": "CLST", "HEPMX": "HEPMX", "MYT": "MYT", "JST": "JST", "EST": "EST", "GYT": "GYT", "UYST": "UYST", "PST": "PST", "PDT": "PDT", "HAT": "HAT", "TMT": "TMT", "MST": "MST", "HEOG": "HEOG", "WIT": "WIT", "TMST": "TMST", "HAST": "HAST", "ChST": "ChST", "CHAST": "CHAST", "CST": "CST", "SGT": "SGT", "ACST": "ACST", "VET": "VET", "HNCU": "HNCU", "JDT": "JDT", "WITA": "WITA", "AEST": "AEST", "OESZ": "Ošteuropäiši Summerzit", "COT": "COT", "UYT": "UYT", "HECU": "HECU", "∅∅∅": "∅∅∅", "GFT": "GFT", "LHDT": "LHDT", "EAT": "EAT", "GMT": "GMT", "HNPMX": "HNPMX", "EDT": "EDT", "WARST": "WARST", "ADT": "Atlantiši Summerzit", "HKST": "HKST", "HEPM": "HEPM", "AKDT": "AKDT", "ACWDT": "ACWDT", "HENOMX": "HENOMX", "HADT": "HADT", "COST": "COST", "MDT": "MDT", "WAST": "WAST", "SRT": "SRT", "AST": "Atlantiši Standardzit", "SAST": "SAST", "BT": "BT", "MESZ": "Mitteleuropäiši Summerzit", "HKT": "HKT", "LHST": "LHST", "WART": "WART", "HNNOMX": "HNNOMX", "ART": "ART", "CDT": "CDT", "AWDT": "AWDT", "NZST": "NZST", "ACWST": "ACWST", "HNT": "HNT", "ARST": "ARST", "AWST": "AWST", "WIB": "WIB", "NZDT": "NZDT", "IST": "IST", "HNPM": "HNPM", "CHADT": "CHADT", "AEDT": "AEDT", "ACDT": "ACDT", "HEEG": "HEEG", "BOT": "BOT", "ECT": "ECT", "HNEG": "HNEG", "MEZ": "Mitteleuropäiši Standardzit"},
	}
}

// Locale returns the current translators string locale
func (wae *wae_CH) Locale() string {
	return wae.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'wae_CH'
func (wae *wae_CH) PluralsCardinal() []locales.PluralRule {
	return wae.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'wae_CH'
func (wae *wae_CH) PluralsOrdinal() []locales.PluralRule {
	return wae.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'wae_CH'
func (wae *wae_CH) PluralsRange() []locales.PluralRule {
	return wae.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'wae_CH'
func (wae *wae_CH) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'wae_CH'
func (wae *wae_CH) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'wae_CH'
func (wae *wae_CH) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (wae *wae_CH) MonthAbbreviated(month time.Month) string {
	return wae.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (wae *wae_CH) MonthsAbbreviated() []string {
	return wae.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (wae *wae_CH) MonthNarrow(month time.Month) string {
	return wae.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (wae *wae_CH) MonthsNarrow() []string {
	return wae.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (wae *wae_CH) MonthWide(month time.Month) string {
	return wae.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (wae *wae_CH) MonthsWide() []string {
	return wae.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (wae *wae_CH) WeekdayAbbreviated(weekday time.Weekday) string {
	return wae.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (wae *wae_CH) WeekdaysAbbreviated() []string {
	return wae.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (wae *wae_CH) WeekdayNarrow(weekday time.Weekday) string {
	return wae.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (wae *wae_CH) WeekdaysNarrow() []string {
	return wae.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (wae *wae_CH) WeekdayShort(weekday time.Weekday) string {
	return wae.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (wae *wae_CH) WeekdaysShort() []string {
	return wae.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (wae *wae_CH) WeekdayWide(weekday time.Weekday) string {
	return wae.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (wae *wae_CH) WeekdaysWide() []string {
	return wae.daysWide
}

// Decimal returns the decimal point of number
func (wae *wae_CH) Decimal() string {
	return wae.decimal
}

// Group returns the group of number
func (wae *wae_CH) Group() string {
	return wae.group
}

// Group returns the minus sign of number
func (wae *wae_CH) Minus() string {
	return wae.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'wae_CH' and handles both Whole and Real numbers based on 'v'
func (wae *wae_CH) FmtNumber(num float64, v uint64) string {

	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'wae_CH' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (wae *wae_CH) FmtPercent(num float64, v uint64) string {
	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'wae_CH'
func (wae *wae_CH) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := wae.currencies[currency]
	l := len(s) + len(symbol) + 4

	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, wae.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	for j := len(symbol) - 1; j >= 0; j-- {
		b = append(b, symbol[j])
	}

	for j := len(wae.currencyPositivePrefix) - 1; j >= 0; j-- {
		b = append(b, wae.currencyPositivePrefix[j])
	}

	if num < 0 {
		b = append(b, wae.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, wae.currencyPositiveSuffix...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'wae_CH'
// in accounting notation.
func (wae *wae_CH) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := wae.currencies[currency]
	l := len(s) + len(symbol) + 4

	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, wae.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {

		for j := len(symbol) - 1; j >= 0; j-- {
			b = append(b, symbol[j])
		}

		for j := len(wae.currencyNegativePrefix) - 1; j >= 0; j-- {
			b = append(b, wae.currencyNegativePrefix[j])
		}

		b = append(b, wae.minus[0])

	} else {

		for j := len(symbol) - 1; j >= 0; j-- {
			b = append(b, symbol[j])
		}

		for j := len(wae.currencyPositivePrefix) - 1; j >= 0; j-- {
			b = append(b, wae.currencyPositivePrefix[j])
		}

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if num < 0 {
		b = append(b, wae.currencyNegativeSuffix...)
	} else {

		b = append(b, wae.currencyPositiveSuffix...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'wae_CH'
func (wae *wae_CH) FmtDateShort(t time.Time) string {

	b := make([]byte, 0, 32)

	return string(b)
}

// FmtDateMedium returns the medium date representation of 't' for 'wae_CH'
func (wae *wae_CH) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2e, 0x20}...)
	b = append(b, wae.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'wae_CH'
func (wae *wae_CH) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2e, 0x20}...)
	b = append(b, wae.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'wae_CH'
func (wae *wae_CH) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, wae.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2e, 0x20}...)
	b = append(b, wae.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'wae_CH'
func (wae *wae_CH) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'wae_CH'
func (wae *wae_CH) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'wae_CH'
func (wae *wae_CH) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'wae_CH'
func (wae *wae_CH) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	return string(b)
}
