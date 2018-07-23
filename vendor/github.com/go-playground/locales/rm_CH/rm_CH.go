package rm_CH

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type rm_CH struct {
	locale                 string
	pluralsCardinal        []locales.PluralRule
	pluralsOrdinal         []locales.PluralRule
	pluralsRange           []locales.PluralRule
	decimal                string
	group                  string
	minus                  string
	percent                string
	percentSuffix          string
	perMille               string
	timeSeparator          string
	inifinity              string
	currencies             []string // idx = enum of currency code
	currencyPositiveSuffix string
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

// New returns a new instance of translator for the 'rm_CH' locale
func New() locales.Translator {
	return &rm_CH{
		locale:                 "rm_CH",
		pluralsCardinal:        []locales.PluralRule{2, 6},
		pluralsOrdinal:         nil,
		pluralsRange:           nil,
		decimal:                ".",
		group:                  "’",
		minus:                  "−",
		percent:                "%",
		perMille:               "‰",
		timeSeparator:          ":",
		inifinity:              "∞",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		percentSuffix:          " ",
		currencyPositiveSuffix: " ",
		currencyNegativeSuffix: " ",
		monthsAbbreviated:      []string{"", "schan.", "favr.", "mars", "avr.", "matg", "zercl.", "fan.", "avust", "sett.", "oct.", "nov.", "dec."},
		monthsNarrow:           []string{"", "S", "F", "M", "A", "M", "Z", "F", "A", "S", "O", "N", "D"},
		monthsWide:             []string{"", "schaner", "favrer", "mars", "avrigl", "matg", "zercladur", "fanadur", "avust", "settember", "october", "november", "december"},
		daysAbbreviated:        []string{"du", "gli", "ma", "me", "gie", "ve", "so"},
		daysNarrow:             []string{"D", "G", "M", "M", "G", "V", "S"},
		daysShort:              []string{"du", "gli", "ma", "me", "gie", "ve", "so"},
		daysWide:               []string{"dumengia", "glindesdi", "mardi", "mesemna", "gievgia", "venderdi", "sonda"},
		periodsAbbreviated:     []string{"AM", "PM"},
		periodsNarrow:          []string{"AM", "PM"},
		periodsWide:            []string{"AM", "PM"},
		erasAbbreviated:        []string{"av. Cr.", "s. Cr."},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"avant Cristus", "suenter Cristus"},
		timezones:              map[string]string{"GYT": "GYT", "AEST": "AEST", "JDT": "JDT", "ACWST": "ACWST", "HNNOMX": "HNNOMX", "ARST": "ARST", "UYST": "UYST", "WAST": "WAST", "JST": "JST", "WIT": "WIT", "HEPMX": "HEPMX", "AKST": "AKST", "HEEG": "HEEG", "MEZ": "MEZ", "TMST": "TMST", "HAST": "HAST", "PDT": "PDT", "CST": "CST", "AST": "AST", "HKST": "HKST", "WITA": "WITA", "OEZ": "OEZ", "BT": "BT", "HEPM": "HEPM", "HNT": "HNT", "SRT": "SRT", "EAT": "EAT", "COT": "COT", "AWST": "AWST", "WIB": "WIB", "SAST": "SAST", "COST": "COST", "HADT": "HADT", "GMT": "GMT", "HNPMX": "HNPMX", "MDT": "MDT", "WEZ": "WEZ", "NZST": "NZST", "VET": "VET", "SGT": "SGT", "ACST": "ACST", "HNEG": "HNEG", "EST": "EST", "CLST": "CLST", "TMT": "TMT", "IST": "IST", "PST": "PST", "WAT": "WAT", "WESZ": "WESZ", "GFT": "GFT", "EDT": "EDT", "LHDT": "LHDT", "∅∅∅": "∅∅∅", "ADT": "ADT", "BOT": "BOT", "HKT": "HKT", "LHST": "LHST", "ART": "ART", "NZDT": "NZDT", "ECT": "ECT", "ACDT": "ACDT", "HEOG": "HEOG", "WARST": "WARST", "HECU": "HECU", "CHAST": "CHAST", "AWDT": "AWDT", "CDT": "CDT", "AEDT": "AEDT", "ACWDT": "ACWDT", "CHADT": "CHADT", "MYT": "MYT", "AKDT": "AKDT", "HAT": "HAT", "CAT": "CAT", "OESZ": "OESZ", "UYT": "UYT", "HNCU": "HNCU", "MST": "MST", "MESZ": "MESZ", "HENOMX": "HENOMX", "CLT": "CLT", "ChST": "ChST", "HNOG": "HNOG", "WART": "WART", "HNPM": "HNPM"},
	}
}

// Locale returns the current translators string locale
func (rm *rm_CH) Locale() string {
	return rm.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'rm_CH'
func (rm *rm_CH) PluralsCardinal() []locales.PluralRule {
	return rm.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'rm_CH'
func (rm *rm_CH) PluralsOrdinal() []locales.PluralRule {
	return rm.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'rm_CH'
func (rm *rm_CH) PluralsRange() []locales.PluralRule {
	return rm.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'rm_CH'
func (rm *rm_CH) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'rm_CH'
func (rm *rm_CH) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'rm_CH'
func (rm *rm_CH) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (rm *rm_CH) MonthAbbreviated(month time.Month) string {
	return rm.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (rm *rm_CH) MonthsAbbreviated() []string {
	return rm.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (rm *rm_CH) MonthNarrow(month time.Month) string {
	return rm.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (rm *rm_CH) MonthsNarrow() []string {
	return rm.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (rm *rm_CH) MonthWide(month time.Month) string {
	return rm.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (rm *rm_CH) MonthsWide() []string {
	return rm.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (rm *rm_CH) WeekdayAbbreviated(weekday time.Weekday) string {
	return rm.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (rm *rm_CH) WeekdaysAbbreviated() []string {
	return rm.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (rm *rm_CH) WeekdayNarrow(weekday time.Weekday) string {
	return rm.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (rm *rm_CH) WeekdaysNarrow() []string {
	return rm.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (rm *rm_CH) WeekdayShort(weekday time.Weekday) string {
	return rm.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (rm *rm_CH) WeekdaysShort() []string {
	return rm.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (rm *rm_CH) WeekdayWide(weekday time.Weekday) string {
	return rm.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (rm *rm_CH) WeekdaysWide() []string {
	return rm.daysWide
}

// Decimal returns the decimal point of number
func (rm *rm_CH) Decimal() string {
	return rm.decimal
}

// Group returns the group of number
func (rm *rm_CH) Group() string {
	return rm.group
}

// Group returns the minus sign of number
func (rm *rm_CH) Minus() string {
	return rm.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'rm_CH' and handles both Whole and Real numbers based on 'v'
func (rm *rm_CH) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 4 + 3*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, rm.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(rm.group) - 1; j >= 0; j-- {
					b = append(b, rm.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		for j := len(rm.minus) - 1; j >= 0; j-- {
			b = append(b, rm.minus[j])
		}
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'rm_CH' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (rm *rm_CH) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 7
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, rm.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		for j := len(rm.minus) - 1; j >= 0; j-- {
			b = append(b, rm.minus[j])
		}
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, rm.percentSuffix...)

	b = append(b, rm.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'rm_CH'
func (rm *rm_CH) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := rm.currencies[currency]
	l := len(s) + len(symbol) + 6 + 3*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, rm.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(rm.group) - 1; j >= 0; j-- {
					b = append(b, rm.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		for j := len(rm.minus) - 1; j >= 0; j-- {
			b = append(b, rm.minus[j])
		}
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, rm.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	b = append(b, rm.currencyPositiveSuffix...)

	b = append(b, symbol...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'rm_CH'
// in accounting notation.
func (rm *rm_CH) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := rm.currencies[currency]
	l := len(s) + len(symbol) + 6 + 3*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, rm.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(rm.group) - 1; j >= 0; j-- {
					b = append(b, rm.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {

		for j := len(rm.minus) - 1; j >= 0; j-- {
			b = append(b, rm.minus[j])
		}

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, rm.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, rm.currencyNegativeSuffix...)
		b = append(b, symbol...)
	} else {

		b = append(b, rm.currencyPositiveSuffix...)
		b = append(b, symbol...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'rm_CH'
func (rm *rm_CH) FmtDateShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Day() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2d}...)

	if t.Month() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Month()), 10)

	b = append(b, []byte{0x2d}...)

	if t.Year() > 9 {
		b = append(b, strconv.Itoa(t.Year())[2:]...)
	} else {
		b = append(b, strconv.Itoa(t.Year())[1:]...)
	}

	return string(b)
}

// FmtDateMedium returns the medium date representation of 't' for 'rm_CH'
func (rm *rm_CH) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Day() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2d}...)

	if t.Month() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Month()), 10)

	b = append(b, []byte{0x2d}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'rm_CH'
func (rm *rm_CH) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20, 0x64, 0x61}...)
	b = append(b, []byte{0x20}...)
	b = append(b, rm.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'rm_CH'
func (rm *rm_CH) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, rm.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20, 0x69, 0x6c, 0x73}...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20, 0x64, 0x61}...)
	b = append(b, []byte{0x20}...)
	b = append(b, rm.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'rm_CH'
func (rm *rm_CH) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, rm.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'rm_CH'
func (rm *rm_CH) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, rm.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, rm.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'rm_CH'
func (rm *rm_CH) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, rm.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, rm.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'rm_CH'
func (rm *rm_CH) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, rm.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, rm.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := rm.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
