package mer

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type mer struct {
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

// New returns a new instance of translator for the 'mer' locale
func New() locales.Translator {
	return &mer{
		locale:                 "mer",
		pluralsCardinal:        nil,
		pluralsOrdinal:         nil,
		pluralsRange:           nil,
		timeSeparator:          ":",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "Ksh", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyNegativePrefix: "(",
		currencyNegativeSuffix: ")",
		monthsAbbreviated:      []string{"", "JAN", "FEB", "MAC", "ĨPU", "MĨĨ", "NJU", "NJR", "AGA", "SPT", "OKT", "NOV", "DEC"},
		monthsNarrow:           []string{"", "J", "F", "M", "Ĩ", "M", "N", "N", "A", "S", "O", "N", "D"},
		monthsWide:             []string{"", "Januarĩ", "Feburuarĩ", "Machi", "Ĩpurũ", "Mĩĩ", "Njuni", "Njuraĩ", "Agasti", "Septemba", "Oktũba", "Novemba", "Dicemba"},
		daysAbbreviated:        []string{"KIU", "MRA", "WAI", "WET", "WEN", "WTN", "JUM"},
		daysNarrow:             []string{"K", "M", "W", "W", "W", "W", "J"},
		daysWide:               []string{"Kiumia", "Muramuko", "Wairi", "Wethatu", "Wena", "Wetano", "Jumamosi"},
		periodsAbbreviated:     []string{"RŨ", "ŨG"},
		periodsWide:            []string{"RŨ", "ŨG"},
		erasAbbreviated:        []string{"MK", "NK"},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"Mbere ya Kristũ", "Nyuma ya Kristũ"},
		timezones:              map[string]string{"GMT": "GMT", "GFT": "GFT", "EDT": "EDT", "HEEG": "HEEG", "JST": "JST", "IST": "IST", "WITA": "WITA", "WIT": "WIT", "HAST": "HAST", "WAST": "WAST", "SGT": "SGT", "MESZ": "MESZ", "HADT": "HADT", "EAT": "EAT", "AKDT": "AKDT", "HNPM": "HNPM", "ART": "ART", "HEPMX": "HEPMX", "AEDT": "AEDT", "NZST": "NZST", "VET": "VET", "BOT": "BOT", "OESZ": "OESZ", "ARST": "ARST", "CDT": "CDT", "AST": "AST", "MST": "MST", "WESZ": "WESZ", "BT": "BT", "CLT": "CLT", "HEPM": "HEPM", "CHAST": "CHAST", "PDT": "PDT", "WIB": "WIB", "HEOG": "HEOG", "HNEG": "HNEG", "LHST": "LHST", "LHDT": "LHDT", "HNNOMX": "HNNOMX", "COT": "COT", "ChST": "ChST", "PST": "PST", "ADT": "ADT", "ACWDT": "ACWDT", "HKT": "HKT", "WART": "WART", "TMT": "TMT", "UYST": "UYST", "HNPMX": "HNPMX", "ACWST": "ACWST", "GYT": "GYT", "ECT": "ECT", "HNOG": "HNOG", "WARST": "WARST", "CHADT": "CHADT", "SAST": "SAST", "MYT": "MYT", "ACDT": "ACDT", "MEZ": "MEZ", "CAT": "CAT", "CST": "CST", "AEST": "AEST", "JDT": "JDT", "AKST": "AKST", "ACST": "ACST", "SRT": "SRT", "TMST": "TMST", "∅∅∅": "∅∅∅", "HECU": "HECU", "AWST": "AWST", "WAT": "WAT", "NZDT": "NZDT", "HAT": "HAT", "COST": "COST", "AWDT": "AWDT", "WEZ": "WEZ", "EST": "EST", "HENOMX": "HENOMX", "OEZ": "OEZ", "UYT": "UYT", "HNCU": "HNCU", "MDT": "MDT", "HKST": "HKST", "HNT": "HNT", "CLST": "CLST"},
	}
}

// Locale returns the current translators string locale
func (mer *mer) Locale() string {
	return mer.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'mer'
func (mer *mer) PluralsCardinal() []locales.PluralRule {
	return mer.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'mer'
func (mer *mer) PluralsOrdinal() []locales.PluralRule {
	return mer.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'mer'
func (mer *mer) PluralsRange() []locales.PluralRule {
	return mer.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'mer'
func (mer *mer) CardinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'mer'
func (mer *mer) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'mer'
func (mer *mer) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (mer *mer) MonthAbbreviated(month time.Month) string {
	return mer.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (mer *mer) MonthsAbbreviated() []string {
	return mer.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (mer *mer) MonthNarrow(month time.Month) string {
	return mer.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (mer *mer) MonthsNarrow() []string {
	return mer.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (mer *mer) MonthWide(month time.Month) string {
	return mer.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (mer *mer) MonthsWide() []string {
	return mer.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (mer *mer) WeekdayAbbreviated(weekday time.Weekday) string {
	return mer.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (mer *mer) WeekdaysAbbreviated() []string {
	return mer.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (mer *mer) WeekdayNarrow(weekday time.Weekday) string {
	return mer.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (mer *mer) WeekdaysNarrow() []string {
	return mer.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (mer *mer) WeekdayShort(weekday time.Weekday) string {
	return mer.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (mer *mer) WeekdaysShort() []string {
	return mer.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (mer *mer) WeekdayWide(weekday time.Weekday) string {
	return mer.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (mer *mer) WeekdaysWide() []string {
	return mer.daysWide
}

// Decimal returns the decimal point of number
func (mer *mer) Decimal() string {
	return mer.decimal
}

// Group returns the group of number
func (mer *mer) Group() string {
	return mer.group
}

// Group returns the minus sign of number
func (mer *mer) Minus() string {
	return mer.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'mer' and handles both Whole and Real numbers based on 'v'
func (mer *mer) FmtNumber(num float64, v uint64) string {

	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'mer' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (mer *mer) FmtPercent(num float64, v uint64) string {
	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'mer'
func (mer *mer) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := mer.currencies[currency]
	l := len(s) + len(symbol) + 0
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, mer.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, mer.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	for j := len(symbol) - 1; j >= 0; j-- {
		b = append(b, symbol[j])
	}

	if num < 0 {
		b = append(b, mer.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, mer.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'mer'
// in accounting notation.
func (mer *mer) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := mer.currencies[currency]
	l := len(s) + len(symbol) + 2
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, mer.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, mer.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {

		for j := len(symbol) - 1; j >= 0; j-- {
			b = append(b, symbol[j])
		}

		b = append(b, mer.currencyNegativePrefix[0])

	} else {

		for j := len(symbol) - 1; j >= 0; j-- {
			b = append(b, symbol[j])
		}

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, mer.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, mer.currencyNegativeSuffix...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'mer'
func (mer *mer) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'mer'
func (mer *mer) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, mer.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'mer'
func (mer *mer) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, mer.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'mer'
func (mer *mer) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, mer.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, mer.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'mer'
func (mer *mer) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, mer.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'mer'
func (mer *mer) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, mer.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, mer.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'mer'
func (mer *mer) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, mer.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, mer.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'mer'
func (mer *mer) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, mer.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, mer.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := mer.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
