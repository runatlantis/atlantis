package jgo

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type jgo struct {
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
	currencyPositivePrefix string
	currencyNegativePrefix string
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

// New returns a new instance of translator for the 'jgo' locale
func New() locales.Translator {
	return &jgo{
		locale:                 "jgo",
		pluralsCardinal:        []locales.PluralRule{2, 6},
		pluralsOrdinal:         nil,
		pluralsRange:           nil,
		decimal:                ",",
		group:                  ".",
		minus:                  "-",
		percent:                "%",
		perMille:               "‰",
		timeSeparator:          ":",
		inifinity:              "∞",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "FCFA", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyPositivePrefix: " ",
		currencyNegativePrefix: " ",
		monthsAbbreviated:      []string{"", "Nduŋmbi Saŋ", "Pɛsaŋ Pɛ́pá", "Pɛsaŋ Pɛ́tát", "Pɛsaŋ Pɛ́nɛ́kwa", "Pɛsaŋ Pataa", "Pɛsaŋ Pɛ́nɛ́ntúkú", "Pɛsaŋ Saambá", "Pɛsaŋ Pɛ́nɛ́fɔm", "Pɛsaŋ Pɛ́nɛ́pfúꞋú", "Pɛsaŋ Nɛgɛ́m", "Pɛsaŋ Ntsɔ̌pmɔ́", "Pɛsaŋ Ntsɔ̌ppá"},
		monthsNarrow:           []string{"", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12"},
		monthsWide:             []string{"", "Nduŋmbi Saŋ", "Pɛsaŋ Pɛ́pá", "Pɛsaŋ Pɛ́tát", "Pɛsaŋ Pɛ́nɛ́kwa", "Pɛsaŋ Pataa", "Pɛsaŋ Pɛ́nɛ́ntúkú", "Pɛsaŋ Saambá", "Pɛsaŋ Pɛ́nɛ́fɔm", "Pɛsaŋ Pɛ́nɛ́pfúꞋú", "Pɛsaŋ Nɛgɛ́m", "Pɛsaŋ Ntsɔ̌pmɔ́", "Pɛsaŋ Ntsɔ̌ppá"},
		daysAbbreviated:        []string{"Sɔ́ndi", "Mɔ́ndi", "Ápta Mɔ́ndi", "Wɛ́nɛsɛdɛ", "Tɔ́sɛdɛ", "Fɛlâyɛdɛ", "Sásidɛ"},
		daysNarrow:             []string{"Sɔ́", "Mɔ́", "ÁM", "Wɛ́", "Tɔ́", "Fɛ", "Sá"},
		daysWide:               []string{"Sɔ́ndi", "Mɔ́ndi", "Ápta Mɔ́ndi", "Wɛ́nɛsɛdɛ", "Tɔ́sɛdɛ", "Fɛlâyɛdɛ", "Sásidɛ"},
		periodsAbbreviated:     []string{"mbaꞌmbaꞌ", "ŋka mbɔ́t nji"},
		periodsWide:            []string{"mbaꞌmbaꞌ", "ŋka mbɔ́t nji"},
		erasAbbreviated:        []string{"BCE", "CE"},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"tsɛttsɛt mɛŋguꞌ mi ɛ́ lɛɛnɛ Kɛlísɛtɔ gɔ ńɔ́", "tsɛttsɛt mɛŋguꞌ mi ɛ́ fúnɛ Kɛlísɛtɔ tɔ́ mɔ́"},
		timezones:              map[string]string{"LHDT": "LHDT", "WITA": "WITA", "MST": "MST", "WIT": "WIT", "UYST": "UYST", "HNPMX": "HNPMX", "AKDT": "AKDT", "WART": "WART", "VET": "VET", "COT": "COT", "GMT": "GMT", "HNCU": "HNCU", "HNEG": "HNEG", "HADT": "HADT", "GYT": "GYT", "CHADT": "CHADT", "AEDT": "AEDT", "WAST": "WAST", "SAST": "SAST", "HNOG": "HNOG", "ACST": "ACST", "MESZ": "MESZ", "HNNOMX": "HNNOMX", "OEZ": "OEZ", "HECU": "HECU", "EAT": "EAT", "COST": "COST", "CDT": "CDT", "ADT": "ADT", "WAT": "WAT", "SGT": "SGT", "∅∅∅": "∅∅∅", "CAT": "CAT", "UYT": "UYT", "BOT": "BOT", "ACWST": "ACWST", "HAT": "HAT", "ARST": "ARST", "CHAST": "CHAST", "JST": "JST", "HEOG": "HEOG", "HKT": "HKT", "HEPM": "HEPM", "TMT": "TMT", "CLT": "CLT", "ACWDT": "ACWDT", "LHST": "LHST", "WARST": "WARST", "ACDT": "ACDT", "MEZ": "MEZ", "SRT": "SRT", "ChST": "ChST", "HNT": "HNT", "HEEG": "HEEG", "EST": "EST", "MDT": "MDT", "TMST": "TMST", "WESZ": "WESZ", "JDT": "JDT", "NZDT": "NZDT", "AKST": "AKST", "HKST": "HKST", "HNPM": "HNPM", "HENOMX": "HENOMX", "HEPMX": "HEPMX", "GFT": "GFT", "WEZ": "WEZ", "NZST": "NZST", "ECT": "ECT", "CLST": "CLST", "OESZ": "OESZ", "PST": "PST", "AWST": "AWST", "AWDT": "AWDT", "HAST": "HAST", "PDT": "PDT", "AEST": "AEST", "BT": "BT", "MYT": "MYT", "EDT": "EDT", "IST": "IST", "ART": "ART", "CST": "CST", "AST": "AST", "WIB": "WIB"},
	}
}

// Locale returns the current translators string locale
func (jgo *jgo) Locale() string {
	return jgo.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'jgo'
func (jgo *jgo) PluralsCardinal() []locales.PluralRule {
	return jgo.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'jgo'
func (jgo *jgo) PluralsOrdinal() []locales.PluralRule {
	return jgo.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'jgo'
func (jgo *jgo) PluralsRange() []locales.PluralRule {
	return jgo.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'jgo'
func (jgo *jgo) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'jgo'
func (jgo *jgo) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'jgo'
func (jgo *jgo) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (jgo *jgo) MonthAbbreviated(month time.Month) string {
	return jgo.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (jgo *jgo) MonthsAbbreviated() []string {
	return jgo.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (jgo *jgo) MonthNarrow(month time.Month) string {
	return jgo.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (jgo *jgo) MonthsNarrow() []string {
	return jgo.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (jgo *jgo) MonthWide(month time.Month) string {
	return jgo.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (jgo *jgo) MonthsWide() []string {
	return jgo.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (jgo *jgo) WeekdayAbbreviated(weekday time.Weekday) string {
	return jgo.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (jgo *jgo) WeekdaysAbbreviated() []string {
	return jgo.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (jgo *jgo) WeekdayNarrow(weekday time.Weekday) string {
	return jgo.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (jgo *jgo) WeekdaysNarrow() []string {
	return jgo.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (jgo *jgo) WeekdayShort(weekday time.Weekday) string {
	return jgo.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (jgo *jgo) WeekdaysShort() []string {
	return jgo.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (jgo *jgo) WeekdayWide(weekday time.Weekday) string {
	return jgo.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (jgo *jgo) WeekdaysWide() []string {
	return jgo.daysWide
}

// Decimal returns the decimal point of number
func (jgo *jgo) Decimal() string {
	return jgo.decimal
}

// Group returns the group of number
func (jgo *jgo) Group() string {
	return jgo.group
}

// Group returns the minus sign of number
func (jgo *jgo) Minus() string {
	return jgo.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'jgo' and handles both Whole and Real numbers based on 'v'
func (jgo *jgo) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, jgo.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, jgo.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, jgo.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'jgo' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (jgo *jgo) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 3
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, jgo.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, jgo.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, jgo.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'jgo'
func (jgo *jgo) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := jgo.currencies[currency]
	l := len(s) + len(symbol) + 4 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, jgo.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, jgo.group[0])
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

	for j := len(jgo.currencyPositivePrefix) - 1; j >= 0; j-- {
		b = append(b, jgo.currencyPositivePrefix[j])
	}

	if num < 0 {
		b = append(b, jgo.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, jgo.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'jgo'
// in accounting notation.
func (jgo *jgo) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := jgo.currencies[currency]
	l := len(s) + len(symbol) + 4 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, jgo.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, jgo.group[0])
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

		for j := len(jgo.currencyNegativePrefix) - 1; j >= 0; j-- {
			b = append(b, jgo.currencyNegativePrefix[j])
		}

		b = append(b, jgo.minus[0])

	} else {

		for j := len(symbol) - 1; j >= 0; j-- {
			b = append(b, symbol[j])
		}

		for j := len(jgo.currencyPositivePrefix) - 1; j >= 0; j-- {
			b = append(b, jgo.currencyPositivePrefix[j])
		}

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, jgo.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'jgo'
func (jgo *jgo) FmtDateShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x2d}...)

	if t.Month() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Month()), 10)

	b = append(b, []byte{0x2d}...)

	if t.Day() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Day()), 10)

	return string(b)
}

// FmtDateMedium returns the medium date representation of 't' for 'jgo'
func (jgo *jgo) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x20}...)
	b = append(b, jgo.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'jgo'
func (jgo *jgo) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x20}...)
	b = append(b, jgo.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'jgo'
func (jgo *jgo) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, jgo.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x20}...)
	b = append(b, jgo.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Day() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Day()), 10)

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'jgo'
func (jgo *jgo) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, jgo.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'jgo'
func (jgo *jgo) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, jgo.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, jgo.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'jgo'
func (jgo *jgo) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, jgo.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, jgo.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'jgo'
func (jgo *jgo) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, jgo.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, jgo.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := jgo.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
