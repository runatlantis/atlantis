package rof

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type rof struct {
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

// New returns a new instance of translator for the 'rof' locale
func New() locales.Translator {
	return &rof{
		locale:             "rof",
		pluralsCardinal:    []locales.PluralRule{2, 6},
		pluralsOrdinal:     nil,
		pluralsRange:       nil,
		timeSeparator:      ":",
		currencies:         []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TSh", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
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
		timezones:          map[string]string{"HNOG": "HNOG", "HAT": "HAT", "HNPM": "HNPM", "HNNOMX": "HNNOMX", "SRT": "SRT", "AWST": "AWST", "EAT": "EAT", "CDT": "CDT", "NZDT": "NZDT", "MYT": "MYT", "LHST": "LHST", "HENOMX": "HENOMX", "ChST": "ChST", "PST": "PST", "GFT": "GFT", "WARST": "WARST", "∅∅∅": "∅∅∅", "WITA": "WITA", "HADT": "HADT", "HNEG": "HNEG", "VET": "VET", "UYT": "UYT", "HNPMX": "HNPMX", "AEDT": "AEDT", "AKST": "AKST", "WAST": "WAST", "WESZ": "WESZ", "WIB": "WIB", "OEZ": "OEZ", "AWDT": "AWDT", "HEPMX": "HEPMX", "ADT": "ADT", "SAST": "SAST", "HEEG": "HEEG", "ACWST": "ACWST", "WART": "WART", "CLT": "CLT", "WIT": "WIT", "ART": "ART", "ACDT": "ACDT", "TMST": "TMST", "HNCU": "HNCU", "HKT": "HKT", "IST": "IST", "HAST": "HAST", "COT": "COT", "JDT": "JDT", "UYST": "UYST", "PDT": "PDT", "AEST": "AEST", "WEZ": "WEZ", "NZST": "NZST", "BOT": "BOT", "JST": "JST", "SGT": "SGT", "HNT": "HNT", "TMT": "TMT", "GMT": "GMT", "AST": "AST", "WAT": "WAT", "ACWDT": "ACWDT", "MESZ": "MESZ", "LHDT": "LHDT", "HEPM": "HEPM", "OESZ": "OESZ", "COST": "COST", "BT": "BT", "MEZ": "MEZ", "ARST": "ARST", "CST": "CST", "AKDT": "AKDT", "EDT": "EDT", "HEOG": "HEOG", "MST": "MST", "MDT": "MDT", "CLST": "CLST", "CHAST": "CHAST", "HKST": "HKST", "EST": "EST", "ACST": "ACST", "CAT": "CAT", "GYT": "GYT", "CHADT": "CHADT", "HECU": "HECU", "ECT": "ECT"},
	}
}

// Locale returns the current translators string locale
func (rof *rof) Locale() string {
	return rof.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'rof'
func (rof *rof) PluralsCardinal() []locales.PluralRule {
	return rof.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'rof'
func (rof *rof) PluralsOrdinal() []locales.PluralRule {
	return rof.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'rof'
func (rof *rof) PluralsRange() []locales.PluralRule {
	return rof.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'rof'
func (rof *rof) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'rof'
func (rof *rof) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'rof'
func (rof *rof) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (rof *rof) MonthAbbreviated(month time.Month) string {
	return rof.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (rof *rof) MonthsAbbreviated() []string {
	return rof.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (rof *rof) MonthNarrow(month time.Month) string {
	return rof.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (rof *rof) MonthsNarrow() []string {
	return rof.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (rof *rof) MonthWide(month time.Month) string {
	return rof.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (rof *rof) MonthsWide() []string {
	return rof.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (rof *rof) WeekdayAbbreviated(weekday time.Weekday) string {
	return rof.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (rof *rof) WeekdaysAbbreviated() []string {
	return rof.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (rof *rof) WeekdayNarrow(weekday time.Weekday) string {
	return rof.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (rof *rof) WeekdaysNarrow() []string {
	return rof.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (rof *rof) WeekdayShort(weekday time.Weekday) string {
	return rof.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (rof *rof) WeekdaysShort() []string {
	return rof.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (rof *rof) WeekdayWide(weekday time.Weekday) string {
	return rof.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (rof *rof) WeekdaysWide() []string {
	return rof.daysWide
}

// Decimal returns the decimal point of number
func (rof *rof) Decimal() string {
	return rof.decimal
}

// Group returns the group of number
func (rof *rof) Group() string {
	return rof.group
}

// Group returns the minus sign of number
func (rof *rof) Minus() string {
	return rof.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'rof' and handles both Whole and Real numbers based on 'v'
func (rof *rof) FmtNumber(num float64, v uint64) string {

	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'rof' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (rof *rof) FmtPercent(num float64, v uint64) string {
	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'rof'
func (rof *rof) FmtCurrency(num float64, v uint64, currency currency.Type) string {

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

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'rof'
// in accounting notation.
func (rof *rof) FmtAccounting(num float64, v uint64, currency currency.Type) string {

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

// FmtDateShort returns the short date representation of 't' for 'rof'
func (rof *rof) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'rof'
func (rof *rof) FmtDateMedium(t time.Time) string {

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

// FmtDateLong returns the long date representation of 't' for 'rof'
func (rof *rof) FmtDateLong(t time.Time) string {

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

// FmtDateFull returns the full date representation of 't' for 'rof'
func (rof *rof) FmtDateFull(t time.Time) string {

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

// FmtTimeShort returns the short time representation of 't' for 'rof'
func (rof *rof) FmtTimeShort(t time.Time) string {

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

// FmtTimeMedium returns the medium time representation of 't' for 'rof'
func (rof *rof) FmtTimeMedium(t time.Time) string {

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

// FmtTimeLong returns the long time representation of 't' for 'rof'
func (rof *rof) FmtTimeLong(t time.Time) string {

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

// FmtTimeFull returns the full time representation of 't' for 'rof'
func (rof *rof) FmtTimeFull(t time.Time) string {

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
