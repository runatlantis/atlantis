package mer_KE

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type mer_KE struct {
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

// New returns a new instance of translator for the 'mer_KE' locale
func New() locales.Translator {
	return &mer_KE{
		locale:                 "mer_KE",
		pluralsCardinal:        nil,
		pluralsOrdinal:         nil,
		pluralsRange:           nil,
		timeSeparator:          ":",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
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
		timezones:              map[string]string{"HNPMX": "HNPMX", "MESZ": "MESZ", "OEZ": "OEZ", "WIT": "WIT", "HAST": "HAST", "HADT": "HADT", "OESZ": "OESZ", "WESZ": "WESZ", "IST": "IST", "HAT": "HAT", "MDT": "MDT", "TMT": "TMT", "UYST": "UYST", "GMT": "GMT", "CST": "CST", "BOT": "BOT", "VET": "VET", "HEPM": "HEPM", "CLST": "CLST", "GYT": "GYT", "PST": "PST", "AKST": "AKST", "ACWST": "ACWST", "HKT": "HKT", "ART": "ART", "HEPMX": "HEPMX", "GFT": "GFT", "JST": "JST", "ACST": "ACST", "HEEG": "HEEG", "EDT": "EDT", "WITA": "WITA", "EAT": "EAT", "AEST": "AEST", "SAST": "SAST", "JDT": "JDT", "HNOG": "HNOG", "EST": "EST", "HNPM": "HNPM", "∅∅∅": "∅∅∅", "AST": "AST", "COST": "COST", "WIB": "WIB", "ECT": "ECT", "ACDT": "ACDT", "LHST": "LHST", "WEZ": "WEZ", "HECU": "HECU", "AWDT": "AWDT", "NZDT": "NZDT", "ACWDT": "ACWDT", "LHDT": "LHDT", "TMST": "TMST", "CHADT": "CHADT", "WAST": "WAST", "MYT": "MYT", "WART": "WART", "HNT": "HNT", "SRT": "SRT", "CLT": "CLT", "COT": "COT", "AWST": "AWST", "ADT": "ADT", "HENOMX": "HENOMX", "HNEG": "HNEG", "NZST": "NZST", "MST": "MST", "UYT": "UYT", "CDT": "CDT", "AKDT": "AKDT", "HNNOMX": "HNNOMX", "HEOG": "HEOG", "WARST": "WARST", "AEDT": "AEDT", "PDT": "PDT", "BT": "BT", "SGT": "SGT", "HKST": "HKST", "MEZ": "MEZ", "CHAST": "CHAST", "ARST": "ARST", "ChST": "ChST", "HNCU": "HNCU", "WAT": "WAT", "CAT": "CAT"},
	}
}

// Locale returns the current translators string locale
func (mer *mer_KE) Locale() string {
	return mer.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'mer_KE'
func (mer *mer_KE) PluralsCardinal() []locales.PluralRule {
	return mer.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'mer_KE'
func (mer *mer_KE) PluralsOrdinal() []locales.PluralRule {
	return mer.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'mer_KE'
func (mer *mer_KE) PluralsRange() []locales.PluralRule {
	return mer.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'mer_KE'
func (mer *mer_KE) CardinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'mer_KE'
func (mer *mer_KE) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'mer_KE'
func (mer *mer_KE) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (mer *mer_KE) MonthAbbreviated(month time.Month) string {
	return mer.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (mer *mer_KE) MonthsAbbreviated() []string {
	return mer.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (mer *mer_KE) MonthNarrow(month time.Month) string {
	return mer.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (mer *mer_KE) MonthsNarrow() []string {
	return mer.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (mer *mer_KE) MonthWide(month time.Month) string {
	return mer.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (mer *mer_KE) MonthsWide() []string {
	return mer.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (mer *mer_KE) WeekdayAbbreviated(weekday time.Weekday) string {
	return mer.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (mer *mer_KE) WeekdaysAbbreviated() []string {
	return mer.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (mer *mer_KE) WeekdayNarrow(weekday time.Weekday) string {
	return mer.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (mer *mer_KE) WeekdaysNarrow() []string {
	return mer.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (mer *mer_KE) WeekdayShort(weekday time.Weekday) string {
	return mer.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (mer *mer_KE) WeekdaysShort() []string {
	return mer.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (mer *mer_KE) WeekdayWide(weekday time.Weekday) string {
	return mer.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (mer *mer_KE) WeekdaysWide() []string {
	return mer.daysWide
}

// Decimal returns the decimal point of number
func (mer *mer_KE) Decimal() string {
	return mer.decimal
}

// Group returns the group of number
func (mer *mer_KE) Group() string {
	return mer.group
}

// Group returns the minus sign of number
func (mer *mer_KE) Minus() string {
	return mer.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'mer_KE' and handles both Whole and Real numbers based on 'v'
func (mer *mer_KE) FmtNumber(num float64, v uint64) string {

	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'mer_KE' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (mer *mer_KE) FmtPercent(num float64, v uint64) string {
	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'mer_KE'
func (mer *mer_KE) FmtCurrency(num float64, v uint64, currency currency.Type) string {

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

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'mer_KE'
// in accounting notation.
func (mer *mer_KE) FmtAccounting(num float64, v uint64, currency currency.Type) string {

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

// FmtDateShort returns the short date representation of 't' for 'mer_KE'
func (mer *mer_KE) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'mer_KE'
func (mer *mer_KE) FmtDateMedium(t time.Time) string {

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

// FmtDateLong returns the long date representation of 't' for 'mer_KE'
func (mer *mer_KE) FmtDateLong(t time.Time) string {

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

// FmtDateFull returns the full date representation of 't' for 'mer_KE'
func (mer *mer_KE) FmtDateFull(t time.Time) string {

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

// FmtTimeShort returns the short time representation of 't' for 'mer_KE'
func (mer *mer_KE) FmtTimeShort(t time.Time) string {

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

// FmtTimeMedium returns the medium time representation of 't' for 'mer_KE'
func (mer *mer_KE) FmtTimeMedium(t time.Time) string {

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

// FmtTimeLong returns the long time representation of 't' for 'mer_KE'
func (mer *mer_KE) FmtTimeLong(t time.Time) string {

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

// FmtTimeFull returns the full time representation of 't' for 'mer_KE'
func (mer *mer_KE) FmtTimeFull(t time.Time) string {

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
