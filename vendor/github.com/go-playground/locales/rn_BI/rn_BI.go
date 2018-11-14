package rn_BI

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type rn_BI struct {
	locale             string
	pluralsCardinal    []locales.PluralRule
	pluralsOrdinal     []locales.PluralRule
	pluralsRange       []locales.PluralRule
	decimal            string
	group              string
	minus              string
	percent            string
	percentSuffix      string
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

// New returns a new instance of translator for the 'rn_BI' locale
func New() locales.Translator {
	return &rn_BI{
		locale:             "rn_BI",
		pluralsCardinal:    nil,
		pluralsOrdinal:     nil,
		pluralsRange:       nil,
		decimal:            ",",
		group:              ".",
		timeSeparator:      ":",
		currencies:         []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		percentSuffix:      " ",
		monthsAbbreviated:  []string{"", "Mut.", "Gas.", "Wer.", "Mat.", "Gic.", "Kam.", "Nya.", "Kan.", "Nze.", "Ukw.", "Ugu.", "Uku."},
		monthsWide:         []string{"", "Nzero", "Ruhuhuma", "Ntwarante", "Ndamukiza", "Rusama", "Ruheshi", "Mukakaro", "Nyandagaro", "Nyakanga", "Gitugutu", "Munyonyo", "Kigarama"},
		daysAbbreviated:    []string{"cu.", "mbe.", "kab.", "gtu.", "kan.", "gnu.", "gnd."},
		daysWide:           []string{"Ku w’indwi", "Ku wa mbere", "Ku wa kabiri", "Ku wa gatatu", "Ku wa kane", "Ku wa gatanu", "Ku wa gatandatu"},
		periodsAbbreviated: []string{"Z.MU.", "Z.MW."},
		periodsWide:        []string{"Z.MU.", "Z.MW."},
		erasAbbreviated:    []string{"Mb.Y.", "Ny.Y"},
		erasNarrow:         []string{"", ""},
		erasWide:           []string{"Mbere ya Yezu", "Nyuma ya Yezu"},
		timezones:          map[string]string{"JDT": "JDT", "EDT": "EDT", "MDT": "MDT", "ARST": "ARST", "UYT": "UYT", "PST": "PST", "WESZ": "WESZ", "HNNOMX": "HNNOMX", "ChST": "ChST", "HNPMX": "HNPMX", "WEZ": "WEZ", "EST": "EST", "WARST": "WARST", "MST": "MST", "CST": "CST", "AEST": "AEST", "WAST": "WAST", "∅∅∅": "∅∅∅", "IST": "IST", "NZST": "NZST", "HEOG": "HEOG", "SAST": "SAST", "JST": "JST", "HNEG": "HNEG", "VET": "VET", "OESZ": "OESZ", "CHADT": "CHADT", "AWDT": "AWDT", "WIB": "WIB", "HKT": "HKT", "HEPM": "HEPM", "HADT": "HADT", "ART": "ART", "PDT": "PDT", "GYT": "GYT", "HECU": "HECU", "AST": "AST", "NZDT": "NZDT", "ACST": "ACST", "LHDT": "LHDT", "WART": "WART", "EAT": "EAT", "ECT": "ECT", "HNOG": "HNOG", "HNT": "HNT", "TMT": "TMT", "COT": "COT", "WAT": "WAT", "HEEG": "HEEG", "HNPM": "HNPM", "HAST": "HAST", "AWST": "AWST", "ACWST": "ACWST", "CLT": "CLT", "CAT": "CAT", "OEZ": "OEZ", "CDT": "CDT", "AKST": "AKST", "CLST": "CLST", "TMST": "TMST", "UYST": "UYST", "ADT": "ADT", "SGT": "SGT", "HKST": "HKST", "HENOMX": "HENOMX", "SRT": "SRT", "CHAST": "CHAST", "AEDT": "AEDT", "AKDT": "AKDT", "MEZ": "MEZ", "WITA": "WITA", "MYT": "MYT", "BOT": "BOT", "BT": "BT", "ACWDT": "ACWDT", "HEPMX": "HEPMX", "GFT": "GFT", "HAT": "HAT", "WIT": "WIT", "COST": "COST", "GMT": "GMT", "HNCU": "HNCU", "LHST": "LHST", "ACDT": "ACDT", "MESZ": "MESZ"},
	}
}

// Locale returns the current translators string locale
func (rn *rn_BI) Locale() string {
	return rn.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'rn_BI'
func (rn *rn_BI) PluralsCardinal() []locales.PluralRule {
	return rn.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'rn_BI'
func (rn *rn_BI) PluralsOrdinal() []locales.PluralRule {
	return rn.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'rn_BI'
func (rn *rn_BI) PluralsRange() []locales.PluralRule {
	return rn.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'rn_BI'
func (rn *rn_BI) CardinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'rn_BI'
func (rn *rn_BI) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'rn_BI'
func (rn *rn_BI) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (rn *rn_BI) MonthAbbreviated(month time.Month) string {
	return rn.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (rn *rn_BI) MonthsAbbreviated() []string {
	return rn.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (rn *rn_BI) MonthNarrow(month time.Month) string {
	return rn.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (rn *rn_BI) MonthsNarrow() []string {
	return nil
}

// MonthWide returns the locales wide month given the 'month' provided
func (rn *rn_BI) MonthWide(month time.Month) string {
	return rn.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (rn *rn_BI) MonthsWide() []string {
	return rn.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (rn *rn_BI) WeekdayAbbreviated(weekday time.Weekday) string {
	return rn.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (rn *rn_BI) WeekdaysAbbreviated() []string {
	return rn.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (rn *rn_BI) WeekdayNarrow(weekday time.Weekday) string {
	return rn.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (rn *rn_BI) WeekdaysNarrow() []string {
	return rn.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (rn *rn_BI) WeekdayShort(weekday time.Weekday) string {
	return rn.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (rn *rn_BI) WeekdaysShort() []string {
	return rn.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (rn *rn_BI) WeekdayWide(weekday time.Weekday) string {
	return rn.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (rn *rn_BI) WeekdaysWide() []string {
	return rn.daysWide
}

// Decimal returns the decimal point of number
func (rn *rn_BI) Decimal() string {
	return rn.decimal
}

// Group returns the group of number
func (rn *rn_BI) Group() string {
	return rn.group
}

// Group returns the minus sign of number
func (rn *rn_BI) Minus() string {
	return rn.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'rn_BI' and handles both Whole and Real numbers based on 'v'
func (rn *rn_BI) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 1 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, rn.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, rn.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, rn.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'rn_BI' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (rn *rn_BI) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 3
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, rn.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, rn.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, rn.percentSuffix...)

	b = append(b, rn.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'rn_BI'
func (rn *rn_BI) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := rn.currencies[currency]
	l := len(s) + len(symbol) + 1 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, rn.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, rn.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, rn.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, rn.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	b = append(b, symbol...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'rn_BI'
// in accounting notation.
func (rn *rn_BI) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := rn.currencies[currency]
	l := len(s) + len(symbol) + 1 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, rn.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, rn.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {

		b = append(b, rn.minus[0])

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, rn.decimal...)
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

// FmtDateShort returns the short date representation of 't' for 'rn_BI'
func (rn *rn_BI) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'rn_BI'
func (rn *rn_BI) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, rn.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'rn_BI'
func (rn *rn_BI) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, rn.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'rn_BI'
func (rn *rn_BI) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, rn.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, rn.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'rn_BI'
func (rn *rn_BI) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, rn.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'rn_BI'
func (rn *rn_BI) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, rn.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, rn.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'rn_BI'
func (rn *rn_BI) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, rn.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, rn.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'rn_BI'
func (rn *rn_BI) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, rn.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, rn.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := rn.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
