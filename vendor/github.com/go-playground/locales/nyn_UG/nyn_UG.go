package nyn_UG

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type nyn_UG struct {
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

// New returns a new instance of translator for the 'nyn_UG' locale
func New() locales.Translator {
	return &nyn_UG{
		locale:            "nyn_UG",
		pluralsCardinal:   []locales.PluralRule{2, 6},
		pluralsOrdinal:    nil,
		pluralsRange:      nil,
		timeSeparator:     ":",
		currencies:        []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		monthsAbbreviated: []string{"", "KBZ", "KBR", "KST", "KKN", "KTN", "KMK", "KMS", "KMN", "KMW", "KKM", "KNK", "KNB"},
		monthsNarrow:      []string{"", "J", "F", "M", "A", "M", "J", "J", "A", "S", "O", "N", "D"},
		monthsWide:        []string{"", "Okwokubanza", "Okwakabiri", "Okwakashatu", "Okwakana", "Okwakataana", "Okwamukaaga", "Okwamushanju", "Okwamunaana", "Okwamwenda", "Okwaikumi", "Okwaikumi na kumwe", "Okwaikumi na ibiri"},
		daysAbbreviated:   []string{"SAN", "ORK", "OKB", "OKS", "OKN", "OKT", "OMK"},
		daysNarrow:        []string{"S", "K", "R", "S", "N", "T", "M"},
		daysWide:          []string{"Sande", "Orwokubanza", "Orwakabiri", "Orwakashatu", "Orwakana", "Orwakataano", "Orwamukaaga"},
		erasAbbreviated:   []string{"BC", "AD"},
		erasNarrow:        []string{"", ""},
		erasWide:          []string{"Kurisito Atakaijire", "Kurisito Yaijire"},
		timezones:         map[string]string{"HKST": "HKST", "TMT": "TMT", "COST": "COST", "HEPMX": "HEPMX", "ADT": "ADT", "ACDT": "ACDT", "WIT": "WIT", "HECU": "HECU", "AWST": "AWST", "JDT": "JDT", "MEZ": "MEZ", "WARST": "WARST", "HADT": "HADT", "GYT": "GYT", "HNCU": "HNCU", "∅∅∅": "∅∅∅", "WAT": "WAT", "MESZ": "MESZ", "MST": "MST", "AKST": "AKST", "ECT": "ECT", "AST": "AST", "ACWST": "ACWST", "HEEG": "HEEG", "CLT": "CLT", "ARST": "ARST", "WESZ": "WESZ", "GFT": "GFT", "CAT": "CAT", "HAT": "HAT", "HNPM": "HNPM", "CDT": "CDT", "WEZ": "WEZ", "MYT": "MYT", "VET": "VET", "PST": "PST", "AEST": "AEST", "ACST": "ACST", "EST": "EST", "ACWDT": "ACWDT", "HEOG": "HEOG", "WITA": "WITA", "HENOMX": "HENOMX", "SAST": "SAST", "WIB": "WIB", "SGT": "SGT", "EDT": "EDT", "HNEG": "HNEG", "HNOG": "HNOG", "ART": "ART", "HNPMX": "HNPMX", "AEDT": "AEDT", "MDT": "MDT", "WART": "WART", "SRT": "SRT", "TMST": "TMST", "COT": "COT", "AWDT": "AWDT", "JST": "JST", "HKT": "HKT", "EAT": "EAT", "CLST": "CLST", "OESZ": "OESZ", "BT": "BT", "NZDT": "NZDT", "LHDT": "LHDT", "HEPM": "HEPM", "HNNOMX": "HNNOMX", "CHADT": "CHADT", "CST": "CST", "BOT": "BOT", "OEZ": "OEZ", "WAST": "WAST", "IST": "IST", "HNT": "HNT", "LHST": "LHST", "HAST": "HAST", "UYST": "UYST", "NZST": "NZST", "AKDT": "AKDT", "GMT": "GMT", "UYT": "UYT", "ChST": "ChST", "CHAST": "CHAST", "PDT": "PDT"},
	}
}

// Locale returns the current translators string locale
func (nyn *nyn_UG) Locale() string {
	return nyn.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'nyn_UG'
func (nyn *nyn_UG) PluralsCardinal() []locales.PluralRule {
	return nyn.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'nyn_UG'
func (nyn *nyn_UG) PluralsOrdinal() []locales.PluralRule {
	return nyn.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'nyn_UG'
func (nyn *nyn_UG) PluralsRange() []locales.PluralRule {
	return nyn.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'nyn_UG'
func (nyn *nyn_UG) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'nyn_UG'
func (nyn *nyn_UG) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'nyn_UG'
func (nyn *nyn_UG) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (nyn *nyn_UG) MonthAbbreviated(month time.Month) string {
	return nyn.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (nyn *nyn_UG) MonthsAbbreviated() []string {
	return nyn.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (nyn *nyn_UG) MonthNarrow(month time.Month) string {
	return nyn.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (nyn *nyn_UG) MonthsNarrow() []string {
	return nyn.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (nyn *nyn_UG) MonthWide(month time.Month) string {
	return nyn.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (nyn *nyn_UG) MonthsWide() []string {
	return nyn.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (nyn *nyn_UG) WeekdayAbbreviated(weekday time.Weekday) string {
	return nyn.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (nyn *nyn_UG) WeekdaysAbbreviated() []string {
	return nyn.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (nyn *nyn_UG) WeekdayNarrow(weekday time.Weekday) string {
	return nyn.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (nyn *nyn_UG) WeekdaysNarrow() []string {
	return nyn.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (nyn *nyn_UG) WeekdayShort(weekday time.Weekday) string {
	return nyn.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (nyn *nyn_UG) WeekdaysShort() []string {
	return nyn.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (nyn *nyn_UG) WeekdayWide(weekday time.Weekday) string {
	return nyn.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (nyn *nyn_UG) WeekdaysWide() []string {
	return nyn.daysWide
}

// Decimal returns the decimal point of number
func (nyn *nyn_UG) Decimal() string {
	return nyn.decimal
}

// Group returns the group of number
func (nyn *nyn_UG) Group() string {
	return nyn.group
}

// Group returns the minus sign of number
func (nyn *nyn_UG) Minus() string {
	return nyn.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'nyn_UG' and handles both Whole and Real numbers based on 'v'
func (nyn *nyn_UG) FmtNumber(num float64, v uint64) string {

	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'nyn_UG' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (nyn *nyn_UG) FmtPercent(num float64, v uint64) string {
	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'nyn_UG'
func (nyn *nyn_UG) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := nyn.currencies[currency]
	l := len(s) + len(symbol) + 0
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, nyn.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, nyn.group[0])
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
		b = append(b, nyn.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, nyn.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'nyn_UG'
// in accounting notation.
func (nyn *nyn_UG) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := nyn.currencies[currency]
	l := len(s) + len(symbol) + 0
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, nyn.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, nyn.group[0])
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

		b = append(b, nyn.minus[0])

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
			b = append(b, nyn.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'nyn_UG'
func (nyn *nyn_UG) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'nyn_UG'
func (nyn *nyn_UG) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, nyn.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'nyn_UG'
func (nyn *nyn_UG) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, nyn.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'nyn_UG'
func (nyn *nyn_UG) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, nyn.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, nyn.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'nyn_UG'
func (nyn *nyn_UG) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, nyn.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'nyn_UG'
func (nyn *nyn_UG) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, nyn.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, nyn.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'nyn_UG'
func (nyn *nyn_UG) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, nyn.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, nyn.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'nyn_UG'
func (nyn *nyn_UG) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, nyn.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, nyn.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := nyn.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
