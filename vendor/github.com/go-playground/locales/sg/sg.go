package sg

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type sg struct {
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

// New returns a new instance of translator for the 'sg' locale
func New() locales.Translator {
	return &sg{
		locale:             "sg",
		pluralsCardinal:    []locales.PluralRule{6},
		pluralsOrdinal:     nil,
		pluralsRange:       nil,
		decimal:            ",",
		group:              ".",
		timeSeparator:      ":",
		currencies:         []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		monthsAbbreviated:  []string{"", "Nye", "Ful", "Mbä", "Ngu", "Bêl", "Fön", "Len", "Kük", "Mvu", "Ngb", "Nab", "Kak"},
		monthsNarrow:       []string{"", "N", "F", "M", "N", "B", "F", "L", "K", "M", "N", "N", "K"},
		monthsWide:         []string{"", "Nyenye", "Fulundïgi", "Mbängü", "Ngubùe", "Bêläwü", "Föndo", "Lengua", "Kükürü", "Mvuka", "Ngberere", "Nabändüru", "Kakauka"},
		daysAbbreviated:    []string{"Bk1", "Bk2", "Bk3", "Bk4", "Bk5", "Lâp", "Lây"},
		daysNarrow:         []string{"K", "S", "T", "S", "K", "P", "Y"},
		daysWide:           []string{"Bikua-ôko", "Bïkua-ûse", "Bïkua-ptâ", "Bïkua-usïö", "Bïkua-okü", "Lâpôsö", "Lâyenga"},
		periodsAbbreviated: []string{"ND", "LK"},
		periodsWide:        []string{"ND", "LK"},
		erasAbbreviated:    []string{"KnK", "NpK"},
		erasNarrow:         []string{"", ""},
		erasWide:           []string{"Kôzo na Krîstu", "Na pekô tî Krîstu"},
		timezones:          map[string]string{"AST": "AST", "NZDT": "NZDT", "TMT": "TMT", "JDT": "JDT", "ART": "ART", "AEDT": "AEDT", "MST": "MST", "SAST": "SAST", "CLT": "CLT", "EDT": "EDT", "LHDT": "LHDT", "HEPM": "HEPM", "HNPM": "HNPM", "SRT": "SRT", "ARST": "ARST", "HNCU": "HNCU", "HECU": "HECU", "MDT": "MDT", "WIB": "WIB", "HKST": "HKST", "CST": "CST", "HNEG": "HNEG", "WART": "WART", "GMT": "GMT", "GYT": "GYT", "AWDT": "AWDT", "WEZ": "WEZ", "WESZ": "WESZ", "BT": "BT", "IST": "IST", "UYT": "UYT", "AWST": "AWST", "HNNOMX": "HNNOMX", "ADT": "ADT", "HEOG": "HEOG", "EST": "EST", "OESZ": "OESZ", "HADT": "HADT", "CHAST": "CHAST", "HKT": "HKT", "HNT": "HNT", "CLST": "CLST", "TMST": "TMST", "EAT": "EAT", "CDT": "CDT", "NZST": "NZST", "HNOG": "HNOG", "WIT": "WIT", "AKDT": "AKDT", "SGT": "SGT", "MEZ": "MEZ", "PDT": "PDT", "AEST": "AEST", "AKST": "AKST", "HAT": "HAT", "LHST": "LHST", "CHADT": "CHADT", "BOT": "BOT", "ECT": "ECT", "ACWDT": "ACWDT", "JST": "JST", "MESZ": "MESZ", "ACST": "ACST", "ACWST": "ACWST", "UYST": "UYST", "PST": "PST", "WAT": "WAT", "WAST": "WAST", "HEEG": "HEEG", "HAST": "HAST", "HEPMX": "HEPMX", "ACDT": "ACDT", "COT": "COT", "OEZ": "OEZ", "WARST": "WARST", "WITA": "WITA", "VET": "VET", "HENOMX": "HENOMX", "CAT": "CAT", "∅∅∅": "∅∅∅", "GFT": "GFT", "MYT": "MYT", "COST": "COST", "ChST": "ChST", "HNPMX": "HNPMX"},
	}
}

// Locale returns the current translators string locale
func (sg *sg) Locale() string {
	return sg.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'sg'
func (sg *sg) PluralsCardinal() []locales.PluralRule {
	return sg.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'sg'
func (sg *sg) PluralsOrdinal() []locales.PluralRule {
	return sg.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'sg'
func (sg *sg) PluralsRange() []locales.PluralRule {
	return sg.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'sg'
func (sg *sg) CardinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'sg'
func (sg *sg) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'sg'
func (sg *sg) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (sg *sg) MonthAbbreviated(month time.Month) string {
	return sg.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (sg *sg) MonthsAbbreviated() []string {
	return sg.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (sg *sg) MonthNarrow(month time.Month) string {
	return sg.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (sg *sg) MonthsNarrow() []string {
	return sg.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (sg *sg) MonthWide(month time.Month) string {
	return sg.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (sg *sg) MonthsWide() []string {
	return sg.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (sg *sg) WeekdayAbbreviated(weekday time.Weekday) string {
	return sg.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (sg *sg) WeekdaysAbbreviated() []string {
	return sg.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (sg *sg) WeekdayNarrow(weekday time.Weekday) string {
	return sg.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (sg *sg) WeekdaysNarrow() []string {
	return sg.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (sg *sg) WeekdayShort(weekday time.Weekday) string {
	return sg.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (sg *sg) WeekdaysShort() []string {
	return sg.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (sg *sg) WeekdayWide(weekday time.Weekday) string {
	return sg.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (sg *sg) WeekdaysWide() []string {
	return sg.daysWide
}

// Decimal returns the decimal point of number
func (sg *sg) Decimal() string {
	return sg.decimal
}

// Group returns the group of number
func (sg *sg) Group() string {
	return sg.group
}

// Group returns the minus sign of number
func (sg *sg) Minus() string {
	return sg.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'sg' and handles both Whole and Real numbers based on 'v'
func (sg *sg) FmtNumber(num float64, v uint64) string {

	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'sg' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (sg *sg) FmtPercent(num float64, v uint64) string {
	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'sg'
func (sg *sg) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := sg.currencies[currency]
	l := len(s) + len(symbol) + 1 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, sg.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, sg.group[0])
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
		b = append(b, sg.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, sg.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'sg'
// in accounting notation.
func (sg *sg) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := sg.currencies[currency]
	l := len(s) + len(symbol) + 1 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, sg.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, sg.group[0])
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

		b = append(b, sg.minus[0])

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
			b = append(b, sg.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'sg'
func (sg *sg) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'sg'
func (sg *sg) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, sg.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x2c, 0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'sg'
func (sg *sg) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, sg.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'sg'
func (sg *sg) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, sg.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, sg.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'sg'
func (sg *sg) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, sg.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'sg'
func (sg *sg) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, sg.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, sg.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'sg'
func (sg *sg) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, sg.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, sg.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'sg'
func (sg *sg) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, sg.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, sg.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := sg.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
