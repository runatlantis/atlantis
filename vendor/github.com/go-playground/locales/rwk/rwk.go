package rwk

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type rwk struct {
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

// New returns a new instance of translator for the 'rwk' locale
func New() locales.Translator {
	return &rwk{
		locale:             "rwk",
		pluralsCardinal:    []locales.PluralRule{2, 6},
		pluralsOrdinal:     nil,
		pluralsRange:       nil,
		timeSeparator:      ":",
		currencies:         []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TSh", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
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
		timezones:          map[string]string{"CST": "CST", "LHST": "LHST", "HAST": "HAST", "GMT": "GMT", "WIT": "WIT", "WESZ": "WESZ", "ECT": "ECT", "ACST": "ACST", "HECU": "HECU", "PST": "PST", "HNOG": "HNOG", "ACDT": "ACDT", "OESZ": "OESZ", "VET": "VET", "AEST": "AEST", "WAST": "WAST", "AKDT": "AKDT", "COST": "COST", "MST": "MST", "SAST": "SAST", "MESZ": "MESZ", "HEPM": "HEPM", "CLT": "CLT", "TMST": "TMST", "WIB": "WIB", "BT": "BT", "WARST": "WARST", "NZST": "NZST", "HADT": "HADT", "AWDT": "AWDT", "MEZ": "MEZ", "HKT": "HKT", "HNPM": "HNPM", "CHAST": "CHAST", "CHADT": "CHADT", "AEDT": "AEDT", "SGT": "SGT", "ACWST": "ACWST", "HEPMX": "HEPMX", "HAT": "HAT", "ChST": "ChST", "PDT": "PDT", "MDT": "MDT", "AST": "AST", "IST": "IST", "HENOMX": "HENOMX", "CAT": "CAT", "ARST": "ARST", "HNCU": "HNCU", "WEZ": "WEZ", "HNEG": "HNEG", "HKST": "HKST", "CLST": "CLST", "ART": "ART", "OEZ": "OEZ", "UYT": "UYT", "GYT": "GYT", "∅∅∅": "∅∅∅", "WART": "WART", "HNNOMX": "HNNOMX", "CDT": "CDT", "WITA": "WITA", "HNT": "HNT", "SRT": "SRT", "NZDT": "NZDT", "HEEG": "HEEG", "LHDT": "LHDT", "TMT": "TMT", "EAT": "EAT", "ADT": "ADT", "WAT": "WAT", "JST": "JST", "HNPMX": "HNPMX", "MYT": "MYT", "ACWDT": "ACWDT", "EST": "EST", "EDT": "EDT", "COT": "COT", "AWST": "AWST", "BOT": "BOT", "GFT": "GFT", "HEOG": "HEOG", "JDT": "JDT", "AKST": "AKST", "UYST": "UYST"},
	}
}

// Locale returns the current translators string locale
func (rwk *rwk) Locale() string {
	return rwk.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'rwk'
func (rwk *rwk) PluralsCardinal() []locales.PluralRule {
	return rwk.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'rwk'
func (rwk *rwk) PluralsOrdinal() []locales.PluralRule {
	return rwk.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'rwk'
func (rwk *rwk) PluralsRange() []locales.PluralRule {
	return rwk.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'rwk'
func (rwk *rwk) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'rwk'
func (rwk *rwk) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'rwk'
func (rwk *rwk) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (rwk *rwk) MonthAbbreviated(month time.Month) string {
	return rwk.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (rwk *rwk) MonthsAbbreviated() []string {
	return rwk.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (rwk *rwk) MonthNarrow(month time.Month) string {
	return rwk.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (rwk *rwk) MonthsNarrow() []string {
	return rwk.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (rwk *rwk) MonthWide(month time.Month) string {
	return rwk.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (rwk *rwk) MonthsWide() []string {
	return rwk.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (rwk *rwk) WeekdayAbbreviated(weekday time.Weekday) string {
	return rwk.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (rwk *rwk) WeekdaysAbbreviated() []string {
	return rwk.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (rwk *rwk) WeekdayNarrow(weekday time.Weekday) string {
	return rwk.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (rwk *rwk) WeekdaysNarrow() []string {
	return rwk.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (rwk *rwk) WeekdayShort(weekday time.Weekday) string {
	return rwk.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (rwk *rwk) WeekdaysShort() []string {
	return rwk.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (rwk *rwk) WeekdayWide(weekday time.Weekday) string {
	return rwk.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (rwk *rwk) WeekdaysWide() []string {
	return rwk.daysWide
}

// Decimal returns the decimal point of number
func (rwk *rwk) Decimal() string {
	return rwk.decimal
}

// Group returns the group of number
func (rwk *rwk) Group() string {
	return rwk.group
}

// Group returns the minus sign of number
func (rwk *rwk) Minus() string {
	return rwk.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'rwk' and handles both Whole and Real numbers based on 'v'
func (rwk *rwk) FmtNumber(num float64, v uint64) string {

	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'rwk' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (rwk *rwk) FmtPercent(num float64, v uint64) string {
	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'rwk'
func (rwk *rwk) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := rwk.currencies[currency]
	l := len(s) + len(symbol) + 0
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, rwk.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, rwk.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, rwk.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, rwk.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	b = append(b, symbol...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'rwk'
// in accounting notation.
func (rwk *rwk) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := rwk.currencies[currency]
	l := len(s) + len(symbol) + 0
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, rwk.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, rwk.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {

		b = append(b, rwk.minus[0])

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, rwk.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, symbol...)
	} else {

		b = append(b, symbol...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'rwk'
func (rwk *rwk) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'rwk'
func (rwk *rwk) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, rwk.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'rwk'
func (rwk *rwk) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, rwk.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'rwk'
func (rwk *rwk) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, rwk.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, rwk.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'rwk'
func (rwk *rwk) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, rwk.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'rwk'
func (rwk *rwk) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, rwk.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, rwk.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'rwk'
func (rwk *rwk) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, rwk.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, rwk.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'rwk'
func (rwk *rwk) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, rwk.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, rwk.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := rwk.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
