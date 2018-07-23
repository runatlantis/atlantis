package cgg

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type cgg struct {
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

// New returns a new instance of translator for the 'cgg' locale
func New() locales.Translator {
	return &cgg{
		locale:            "cgg",
		pluralsCardinal:   []locales.PluralRule{2, 6},
		pluralsOrdinal:    nil,
		pluralsRange:      nil,
		timeSeparator:     ":",
		currencies:        []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "USh", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		monthsAbbreviated: []string{"", "KBZ", "KBR", "KST", "KKN", "KTN", "KMK", "KMS", "KMN", "KMW", "KKM", "KNK", "KNB"},
		monthsNarrow:      []string{"", "J", "F", "M", "A", "M", "J", "J", "A", "S", "O", "N", "D"},
		monthsWide:        []string{"", "Okwokubanza", "Okwakabiri", "Okwakashatu", "Okwakana", "Okwakataana", "Okwamukaaga", "Okwamushanju", "Okwamunaana", "Okwamwenda", "Okwaikumi", "Okwaikumi na kumwe", "Okwaikumi na ibiri"},
		daysAbbreviated:   []string{"SAN", "ORK", "OKB", "OKS", "OKN", "OKT", "OMK"},
		daysNarrow:        []string{"S", "K", "R", "S", "N", "T", "M"},
		daysWide:          []string{"Sande", "Orwokubanza", "Orwakabiri", "Orwakashatu", "Orwakana", "Orwakataano", "Orwamukaaga"},
		erasAbbreviated:   []string{"BC", "AD"},
		erasNarrow:        []string{"", ""},
		erasWide:          []string{"Kurisito Atakaijire", "Kurisito Yaijire"},
		timezones:         map[string]string{"EST": "EST", "TMT": "TMT", "ADT": "ADT", "ECT": "ECT", "JDT": "JDT", "MEZ": "MEZ", "MST": "MST", "TMST": "TMST", "GMT": "GMT", "UYT": "UYT", "WEZ": "WEZ", "JST": "JST", "WIT": "WIT", "COST": "COST", "AST": "AST", "NZDT": "NZDT", "ACWDT": "ACWDT", "HEOG": "HEOG", "BOT": "BOT", "AKDT": "AKDT", "CLST": "CLST", "HAST": "HAST", "CST": "CST", "HEPMX": "HEPMX", "AEST": "AEST", "WAT": "WAT", "NZST": "NZST", "MDT": "MDT", "SRT": "SRT", "CLT": "CLT", "OESZ": "OESZ", "∅∅∅": "∅∅∅", "PDT": "PDT", "LHDT": "LHDT", "HEPM": "HEPM", "CAT": "CAT", "HECU": "HECU", "AWDT": "AWDT", "HNPMX": "HNPMX", "AEDT": "AEDT", "MESZ": "MESZ", "OEZ": "OEZ", "COT": "COT", "CHAST": "CHAST", "SGT": "SGT", "HNCU": "HNCU", "CDT": "CDT", "MYT": "MYT", "EDT": "EDT", "HEEG": "HEEG", "HNPM": "HNPM", "HNNOMX": "HNNOMX", "PST": "PST", "GFT": "GFT", "ACDT": "ACDT", "HKT": "HKT", "WART": "WART", "WITA": "WITA", "EAT": "EAT", "HADT": "HADT", "ARST": "ARST", "CHADT": "CHADT", "AKST": "AKST", "ART": "ART", "SAST": "SAST", "WAST": "WAST", "ACWST": "ACWST", "VET": "VET", "BT": "BT", "IST": "IST", "AWST": "AWST", "HNOG": "HNOG", "LHST": "LHST", "HAT": "HAT", "HENOMX": "HENOMX", "GYT": "GYT", "UYST": "UYST", "ChST": "ChST", "WIB": "WIB", "HKST": "HKST", "WARST": "WARST", "WESZ": "WESZ", "ACST": "ACST", "HNEG": "HNEG", "HNT": "HNT"},
	}
}

// Locale returns the current translators string locale
func (cgg *cgg) Locale() string {
	return cgg.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'cgg'
func (cgg *cgg) PluralsCardinal() []locales.PluralRule {
	return cgg.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'cgg'
func (cgg *cgg) PluralsOrdinal() []locales.PluralRule {
	return cgg.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'cgg'
func (cgg *cgg) PluralsRange() []locales.PluralRule {
	return cgg.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'cgg'
func (cgg *cgg) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'cgg'
func (cgg *cgg) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'cgg'
func (cgg *cgg) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (cgg *cgg) MonthAbbreviated(month time.Month) string {
	return cgg.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (cgg *cgg) MonthsAbbreviated() []string {
	return cgg.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (cgg *cgg) MonthNarrow(month time.Month) string {
	return cgg.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (cgg *cgg) MonthsNarrow() []string {
	return cgg.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (cgg *cgg) MonthWide(month time.Month) string {
	return cgg.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (cgg *cgg) MonthsWide() []string {
	return cgg.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (cgg *cgg) WeekdayAbbreviated(weekday time.Weekday) string {
	return cgg.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (cgg *cgg) WeekdaysAbbreviated() []string {
	return cgg.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (cgg *cgg) WeekdayNarrow(weekday time.Weekday) string {
	return cgg.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (cgg *cgg) WeekdaysNarrow() []string {
	return cgg.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (cgg *cgg) WeekdayShort(weekday time.Weekday) string {
	return cgg.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (cgg *cgg) WeekdaysShort() []string {
	return cgg.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (cgg *cgg) WeekdayWide(weekday time.Weekday) string {
	return cgg.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (cgg *cgg) WeekdaysWide() []string {
	return cgg.daysWide
}

// Decimal returns the decimal point of number
func (cgg *cgg) Decimal() string {
	return cgg.decimal
}

// Group returns the group of number
func (cgg *cgg) Group() string {
	return cgg.group
}

// Group returns the minus sign of number
func (cgg *cgg) Minus() string {
	return cgg.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'cgg' and handles both Whole and Real numbers based on 'v'
func (cgg *cgg) FmtNumber(num float64, v uint64) string {

	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'cgg' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (cgg *cgg) FmtPercent(num float64, v uint64) string {
	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'cgg'
func (cgg *cgg) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := cgg.currencies[currency]
	l := len(s) + len(symbol) + 0
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, cgg.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, cgg.group[0])
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
		b = append(b, cgg.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, cgg.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'cgg'
// in accounting notation.
func (cgg *cgg) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := cgg.currencies[currency]
	l := len(s) + len(symbol) + 0
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, cgg.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, cgg.group[0])
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

		b = append(b, cgg.minus[0])

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
			b = append(b, cgg.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'cgg'
func (cgg *cgg) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'cgg'
func (cgg *cgg) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, cgg.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'cgg'
func (cgg *cgg) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, cgg.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'cgg'
func (cgg *cgg) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, cgg.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, cgg.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'cgg'
func (cgg *cgg) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, cgg.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'cgg'
func (cgg *cgg) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, cgg.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, cgg.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'cgg'
func (cgg *cgg) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, cgg.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, cgg.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'cgg'
func (cgg *cgg) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, cgg.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, cgg.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := cgg.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
