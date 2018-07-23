package mt_MT

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type mt_MT struct {
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

// New returns a new instance of translator for the 'mt_MT' locale
func New() locales.Translator {
	return &mt_MT{
		locale:             "mt_MT",
		pluralsCardinal:    []locales.PluralRule{2, 4, 5, 6},
		pluralsOrdinal:     nil,
		pluralsRange:       nil,
		decimal:            ".",
		group:              ",",
		minus:              "-",
		percent:            "%",
		perMille:           "‰",
		timeSeparator:      ":",
		inifinity:          "∞",
		currencies:         []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		monthsAbbreviated:  []string{"", "Jan", "Fra", "Mar", "Apr", "Mej", "Ġun", "Lul", "Aww", "Set", "Ott", "Nov", "Diċ"},
		monthsNarrow:       []string{"", "J", "F", "M", "A", "M", "Ġ", "L", "A", "S", "O", "N", "D"},
		monthsWide:         []string{"", "Jannar", "Frar", "Marzu", "April", "Mejju", "Ġunju", "Lulju", "Awwissu", "Settembru", "Ottubru", "Novembru", "Diċembru"},
		daysAbbreviated:    []string{"Ħad", "Tne", "Tli", "Erb", "Ħam", "Ġim", "Sib"},
		daysNarrow:         []string{"Ħd", "T", "Tl", "Er", "Ħm", "Ġm", "Sb"},
		daysShort:          []string{"Ħad", "Tne", "Tli", "Erb", "Ħam", "Ġim", "Sib"},
		daysWide:           []string{"Il-Ħadd", "It-Tnejn", "It-Tlieta", "L-Erbgħa", "Il-Ħamis", "Il-Ġimgħa", "Is-Sibt"},
		periodsAbbreviated: []string{"AM", "PM"},
		periodsNarrow:      []string{"am", "pm"},
		periodsWide:        []string{"AM", "PM"},
		erasAbbreviated:    []string{"QK", "WK"},
		erasNarrow:         []string{"", ""},
		erasWide:           []string{"Qabel Kristu", "Wara Kristu"},
		timezones:          map[string]string{"JDT": "JDT", "AKDT": "AKDT", "HKT": "HKT", "HAST": "HAST", "HNNOMX": "HNNOMX", "HECU": "HECU", "ACWST": "ACWST", "MESZ": "Ħin Ċentrali Ewropew tas-Sajf", "HEPM": "HEPM", "HNT": "HNT", "HAT": "HAT", "HNCU": "HNCU", "CDT": "CDT", "ADT": "ADT", "BT": "BT", "EST": "EST", "LHST": "LHST", "COT": "COT", "UYST": "UYST", "HKST": "HKST", "MDT": "MDT", "AWDT": "AWDT", "MEZ": "Ħin Ċentrali Ewropew Standard", "VET": "VET", "PST": "PST", "AST": "AST", "SAST": "SAST", "WAT": "WAT", "ACST": "ACST", "CLST": "CLST", "∅∅∅": "∅∅∅", "TMST": "TMST", "GFT": "GFT", "ECT": "ECT", "HEOG": "HEOG", "IST": "IST", "GYT": "GYT", "CST": "CST", "WAST": "WAST", "HNOG": "HNOG", "MST": "MST", "HNPMX": "HNPMX", "WEZ": "WEZ", "EDT": "EDT", "EAT": "EAT", "UYT": "UYT", "AEDT": "AEDT", "WIB": "WIB", "ACWDT": "ACWDT", "HEEG": "HEEG", "WART": "WART", "CLT": "CLT", "OEZ": "OEZ", "AEST": "AEST", "AKST": "AKST", "HNPM": "HNPM", "WITA": "WITA", "HADT": "HADT", "CHAST": "CHAST", "SGT": "SGT", "LHDT": "LHDT", "GMT": "GMT", "JST": "JST", "NZST": "NZST", "WARST": "WARST", "WIT": "WIT", "OESZ": "OESZ", "ART": "ART", "ARST": "ARST", "WESZ": "WESZ", "MYT": "MYT", "CAT": "CAT", "TMT": "TMT", "COST": "COST", "CHADT": "CHADT", "HEPMX": "HEPMX", "ACDT": "ACDT", "HNEG": "HNEG", "HENOMX": "HENOMX", "ChST": "ChST", "PDT": "PDT", "AWST": "AWST", "NZDT": "NZDT", "BOT": "BOT", "SRT": "SRT"},
	}
}

// Locale returns the current translators string locale
func (mt *mt_MT) Locale() string {
	return mt.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'mt_MT'
func (mt *mt_MT) PluralsCardinal() []locales.PluralRule {
	return mt.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'mt_MT'
func (mt *mt_MT) PluralsOrdinal() []locales.PluralRule {
	return mt.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'mt_MT'
func (mt *mt_MT) PluralsRange() []locales.PluralRule {
	return mt.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'mt_MT'
func (mt *mt_MT) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)
	nMod100 := math.Mod(n, 100)

	if n == 1 {
		return locales.PluralRuleOne
	} else if (n == 0) || (nMod100 >= 2 && nMod100 <= 10) {
		return locales.PluralRuleFew
	} else if nMod100 >= 11 && nMod100 <= 19 {
		return locales.PluralRuleMany
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'mt_MT'
func (mt *mt_MT) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'mt_MT'
func (mt *mt_MT) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (mt *mt_MT) MonthAbbreviated(month time.Month) string {
	return mt.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (mt *mt_MT) MonthsAbbreviated() []string {
	return mt.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (mt *mt_MT) MonthNarrow(month time.Month) string {
	return mt.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (mt *mt_MT) MonthsNarrow() []string {
	return mt.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (mt *mt_MT) MonthWide(month time.Month) string {
	return mt.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (mt *mt_MT) MonthsWide() []string {
	return mt.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (mt *mt_MT) WeekdayAbbreviated(weekday time.Weekday) string {
	return mt.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (mt *mt_MT) WeekdaysAbbreviated() []string {
	return mt.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (mt *mt_MT) WeekdayNarrow(weekday time.Weekday) string {
	return mt.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (mt *mt_MT) WeekdaysNarrow() []string {
	return mt.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (mt *mt_MT) WeekdayShort(weekday time.Weekday) string {
	return mt.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (mt *mt_MT) WeekdaysShort() []string {
	return mt.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (mt *mt_MT) WeekdayWide(weekday time.Weekday) string {
	return mt.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (mt *mt_MT) WeekdaysWide() []string {
	return mt.daysWide
}

// Decimal returns the decimal point of number
func (mt *mt_MT) Decimal() string {
	return mt.decimal
}

// Group returns the group of number
func (mt *mt_MT) Group() string {
	return mt.group
}

// Group returns the minus sign of number
func (mt *mt_MT) Minus() string {
	return mt.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'mt_MT' and handles both Whole and Real numbers based on 'v'
func (mt *mt_MT) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, mt.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, mt.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, mt.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'mt_MT' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (mt *mt_MT) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 3
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, mt.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, mt.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, mt.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'mt_MT'
func (mt *mt_MT) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := mt.currencies[currency]
	l := len(s) + len(symbol) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, mt.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, mt.group[0])
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
		b = append(b, mt.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, mt.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'mt_MT'
// in accounting notation.
func (mt *mt_MT) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := mt.currencies[currency]
	l := len(s) + len(symbol) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, mt.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, mt.group[0])
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

		b = append(b, mt.minus[0])

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
			b = append(b, mt.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'mt_MT'
func (mt *mt_MT) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'mt_MT'
func (mt *mt_MT) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Day() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, mt.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'mt_MT'
func (mt *mt_MT) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20, 0x74, 0x61}...)
	b = append(b, []byte{0xe2, 0x80, 0x99, 0x20}...)
	b = append(b, mt.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'mt_MT'
func (mt *mt_MT) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, mt.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20, 0x74, 0x61}...)
	b = append(b, []byte{0xe2, 0x80, 0x99, 0x20}...)
	b = append(b, mt.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'mt_MT'
func (mt *mt_MT) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, mt.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'mt_MT'
func (mt *mt_MT) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, mt.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, mt.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'mt_MT'
func (mt *mt_MT) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, mt.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, mt.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'mt_MT'
func (mt *mt_MT) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, mt.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, mt.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := mt.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
