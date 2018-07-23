package rof_TZ

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type rof_TZ struct {
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

// New returns a new instance of translator for the 'rof_TZ' locale
func New() locales.Translator {
	return &rof_TZ{
		locale:             "rof_TZ",
		pluralsCardinal:    []locales.PluralRule{2, 6},
		pluralsOrdinal:     nil,
		pluralsRange:       nil,
		timeSeparator:      ":",
		currencies:         []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		monthsAbbreviated:  []string{"", "M1", "M2", "M3", "M4", "M5", "M6", "M7", "M8", "M9", "M10", "M11", "M12"},
		monthsNarrow:       []string{"", "K", "K", "K", "K", "T", "S", "S", "N", "T", "I", "I", "I"},
		monthsWide:         []string{"", "Mweri wa kwanza", "Mweri wa kaili", "Mweri wa katatu", "Mweri wa kaana", "Mweri wa tanu", "Mweri wa sita", "Mweri wa saba", "Mweri wa nane", "Mweri wa tisa", "Mweri wa ikumi", "Mweri wa ikumi na moja", "Mweri wa ikumi na mbili"},
		daysAbbreviated:    []string{"Ijp", "Ijt", "Ijn", "Ijtn", "Alh", "Iju", "Ijm"},
		daysNarrow:         []string{"2", "3", "4", "5", "6", "7", "1"},
		daysWide:           []string{"Ijumapili", "Ijumatatu", "Ijumanne", "Ijumatano", "Alhamisi", "Ijumaa", "Ijumamosi"},
		periodsAbbreviated: []string{"kang’ama", "kingoto"},
		periodsWide:        []string{"kang’ama", "kingoto"},
		erasAbbreviated:    []string{"KM", "BM"},
		erasNarrow:         []string{"", ""},
		erasWide:           []string{"Kabla ya Mayesu", "Baada ya Mayesu"},
		timezones:          map[string]string{"AWDT": "AWDT", "ADT": "ADT", "ECT": "ECT", "CST": "CST", "CDT": "CDT", "HEPM": "HEPM", "UYST": "UYST", "HNCU": "HNCU", "HECU": "HECU", "WEZ": "WEZ", "JST": "JST", "HNOG": "HNOG", "HEOG": "HEOG", "BT": "BT", "HNEG": "HNEG", "HKST": "HKST", "WIT": "WIT", "ChST": "ChST", "PST": "PST", "PDT": "PDT", "AEST": "AEST", "∅∅∅": "∅∅∅", "JDT": "JDT", "ACDT": "ACDT", "AEDT": "AEDT", "HNT": "HNT", "GFT": "GFT", "CLT": "CLT", "OESZ": "OESZ", "CHAST": "CHAST", "MDT": "MDT", "WESZ": "WESZ", "CAT": "CAT", "GMT": "GMT", "ACWST": "ACWST", "ACWDT": "ACWDT", "TMT": "TMT", "COST": "COST", "AST": "AST", "WAT": "WAT", "HAST": "HAST", "COT": "COT", "VET": "VET", "HNPM": "HNPM", "CHADT": "CHADT", "MYT": "MYT", "AKST": "AKST", "MEZ": "MEZ", "WARST": "WARST", "CLST": "CLST", "AWST": "AWST", "NZDT": "NZDT", "BOT": "BOT", "IST": "IST", "WITA": "WITA", "SRT": "SRT", "EAT": "EAT", "WAST": "WAST", "AKDT": "AKDT", "ACST": "ACST", "MESZ": "MESZ", "EST": "EST", "EDT": "EDT", "HAT": "HAT", "HENOMX": "HENOMX", "HADT": "HADT", "ARST": "ARST", "UYT": "UYT", "MST": "MST", "SGT": "SGT", "LHDT": "LHDT", "WART": "WART", "HKT": "HKT", "ART": "ART", "HNPMX": "HNPMX", "LHST": "LHST", "HNNOMX": "HNNOMX", "TMST": "TMST", "OEZ": "OEZ", "SAST": "SAST", "WIB": "WIB", "NZST": "NZST", "HEEG": "HEEG", "GYT": "GYT", "HEPMX": "HEPMX"},
	}
}

// Locale returns the current translators string locale
func (rof *rof_TZ) Locale() string {
	return rof.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'rof_TZ'
func (rof *rof_TZ) PluralsCardinal() []locales.PluralRule {
	return rof.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'rof_TZ'
func (rof *rof_TZ) PluralsOrdinal() []locales.PluralRule {
	return rof.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'rof_TZ'
func (rof *rof_TZ) PluralsRange() []locales.PluralRule {
	return rof.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'rof_TZ'
func (rof *rof_TZ) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'rof_TZ'
func (rof *rof_TZ) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'rof_TZ'
func (rof *rof_TZ) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (rof *rof_TZ) MonthAbbreviated(month time.Month) string {
	return rof.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (rof *rof_TZ) MonthsAbbreviated() []string {
	return rof.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (rof *rof_TZ) MonthNarrow(month time.Month) string {
	return rof.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (rof *rof_TZ) MonthsNarrow() []string {
	return rof.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (rof *rof_TZ) MonthWide(month time.Month) string {
	return rof.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (rof *rof_TZ) MonthsWide() []string {
	return rof.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (rof *rof_TZ) WeekdayAbbreviated(weekday time.Weekday) string {
	return rof.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (rof *rof_TZ) WeekdaysAbbreviated() []string {
	return rof.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (rof *rof_TZ) WeekdayNarrow(weekday time.Weekday) string {
	return rof.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (rof *rof_TZ) WeekdaysNarrow() []string {
	return rof.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (rof *rof_TZ) WeekdayShort(weekday time.Weekday) string {
	return rof.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (rof *rof_TZ) WeekdaysShort() []string {
	return rof.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (rof *rof_TZ) WeekdayWide(weekday time.Weekday) string {
	return rof.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (rof *rof_TZ) WeekdaysWide() []string {
	return rof.daysWide
}

// Decimal returns the decimal point of number
func (rof *rof_TZ) Decimal() string {
	return rof.decimal
}

// Group returns the group of number
func (rof *rof_TZ) Group() string {
	return rof.group
}

// Group returns the minus sign of number
func (rof *rof_TZ) Minus() string {
	return rof.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'rof_TZ' and handles both Whole and Real numbers based on 'v'
func (rof *rof_TZ) FmtNumber(num float64, v uint64) string {

	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'rof_TZ' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (rof *rof_TZ) FmtPercent(num float64, v uint64) string {
	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'rof_TZ'
func (rof *rof_TZ) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := rof.currencies[currency]
	l := len(s) + len(symbol) + 0
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, rof.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, rof.group[0])
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
		b = append(b, rof.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, rof.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'rof_TZ'
// in accounting notation.
func (rof *rof_TZ) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := rof.currencies[currency]
	l := len(s) + len(symbol) + 0
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, rof.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, rof.group[0])
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

		b = append(b, rof.minus[0])

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
			b = append(b, rof.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'rof_TZ'
func (rof *rof_TZ) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'rof_TZ'
func (rof *rof_TZ) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, rof.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'rof_TZ'
func (rof *rof_TZ) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, rof.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'rof_TZ'
func (rof *rof_TZ) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, rof.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, rof.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'rof_TZ'
func (rof *rof_TZ) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, rof.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'rof_TZ'
func (rof *rof_TZ) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, rof.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, rof.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'rof_TZ'
func (rof *rof_TZ) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, rof.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, rof.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'rof_TZ'
func (rof *rof_TZ) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, rof.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, rof.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := rof.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
