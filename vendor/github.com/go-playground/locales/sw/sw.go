package sw

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type sw struct {
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

// New returns a new instance of translator for the 'sw' locale
func New() locales.Translator {
	return &sw{
		locale:                 "sw",
		pluralsCardinal:        []locales.PluralRule{2, 6},
		pluralsOrdinal:         []locales.PluralRule{6},
		pluralsRange:           []locales.PluralRule{2, 6},
		decimal:                ".",
		group:                  ",",
		minus:                  "-",
		percent:                "%",
		perMille:               "‰",
		timeSeparator:          ":",
		inifinity:              "∞",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "A$", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "R$", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CA$", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CN¥", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "€", "FIM", "FJD", "FKP", "FRF", "£", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HK$", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "₪", "₹", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JP¥", "Ksh", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "₩", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MX$", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZ$", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "฿", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "NT$", "TSh", "UAH", "UAK", "UGS", "UGX", "US$", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "₫", "VNN", "VUV", "WST", "FCFA", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "EC$", "XDR", "XEU", "XFO", "XFU", "CFA", "XPD", "CFPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyNegativePrefix: "(",
		currencyNegativeSuffix: ")",
		monthsAbbreviated:      []string{"", "Jan", "Feb", "Mac", "Apr", "Mei", "Jun", "Jul", "Ago", "Sep", "Okt", "Nov", "Des"},
		monthsNarrow:           []string{"", "J", "F", "M", "A", "M", "J", "J", "A", "S", "O", "N", "D"},
		monthsWide:             []string{"", "Januari", "Februari", "Machi", "Aprili", "Mei", "Juni", "Julai", "Agosti", "Septemba", "Oktoba", "Novemba", "Desemba"},
		daysAbbreviated:        []string{"Jumapili", "Jumatatu", "Jumanne", "Jumatano", "Alhamisi", "Ijumaa", "Jumamosi"},
		daysNarrow:             []string{"S", "M", "T", "W", "T", "F", "S"},
		daysShort:              []string{"Jumapili", "Jumatatu", "Jumanne", "Jumatano", "Alhamisi", "Ijumaa", "Jumamosi"},
		daysWide:               []string{"Jumapili", "Jumatatu", "Jumanne", "Jumatano", "Alhamisi", "Ijumaa", "Jumamosi"},
		periodsAbbreviated:     []string{"AM", "PM"},
		periodsNarrow:          []string{"am", "pm"},
		periodsWide:            []string{"AM", "PM"},
		erasAbbreviated:        []string{"KK", "BK"},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"Kabla ya Kristo", "Baada ya Kristo"},
		timezones:              map[string]string{"HNOG": "Saa za Wastani za Greenland Magharibi", "EST": "Saa za Wastani za Mashariki", "HEPM": "Saa za Mchana za Saint-Pierre na Miquelon", "ADT": "Saa za Mchana za Atlantiki", "SAST": "Saa za Wastani za Afrika Kusini", "WAST": "Saa za Majira ya joto za Afrika Magharibi", "WEZ": "Saa za Wastani za Magharibi mwa Ulaya", "MYT": "Saa za Malaysia", "ARST": "Saa za Majira ya joto za Argentina", "EAT": "Saa za Afrika Mashariki", "GFT": "Saa za Guiana ya Ufaransa", "JDT": "Saa za Mchana za Japan", "ACST": "Saa za Wastani za Australia ya Kati", "ACWST": "Saa za Wastani za Magharibi ya Kati ya Australia", "HKST": "Saa za Majira ya joto za Hong Kong", "MDT": "MDT", "HENOMX": "Saa za mchana za Mexico Kaskazini Magharibi", "WIT": "Saa za Mashariki mwa Indonesia", "OEZ": "Saa za Wastani za Mashariki mwa Ulaya", "WIB": "Saa za Magharibi mwa Indonesia", "SGT": "Saa za Wastani za Singapore", "HNNOMX": "Saa za Wastani za Mexico Kaskazini Magharibi", "CAT": "Saa za Afrika ya Kati", "UYST": "Saa za Majira ya joto za Uruguay", "ChST": "Saa za Wastani za Chamorro", "CHAST": "Saa za Wastani za Chatham", "AST": "Saa za Wastani za Atlantiki", "WAT": "Saa za Wastani za Afrika Magharibi", "IST": "Saa Wastani za India", "TMST": "Saa za Majira ya joto za Turkmenistan", "VET": "Saa za Venezuela", "HAT": "Saa za Mchana za Newfoundland", "EDT": "Saa za Mchana za Mashariki", "HKT": "Saa za Wastani za Hong Kong", "WART": "Saa za Wastani za Magharibi mwa Argentina", "HADT": "Saa za Mchana za Hawaii-Aleutian", "HNPMX": "Saa za wastani za pasifiki za Mexico", "HNCU": "Saa za Wastani ya Cuba", "ECT": "Saa za Ecuador", "HNT": "Saa za Wastani za Newfoundland", "COT": "Saa za Wastani za Colombia", "HECU": "Saa za Mchana za Cuba", "BOT": "Saa za Bolivia", "MEZ": "Saa za Wastani za Ulaya ya kati", "MESZ": "Saa za Majira ya joto za Ulaya ya Kati", "LHST": "Saa za Wastani za Lord Howe", "HNPM": "Saa za Wastani ya Saint-Pierre na Miquelon", "CLST": "Saa za Majira ya joto za Chile", "PST": "Saa za Wastani za Pasifiki", "PDT": "Saa za Mchana za Pasifiki", "WESZ": "Saa za Majira ya joto za Magharibi mwa Ulaya", "ACDT": "Saa za Mchana za Australia ya Kati", "MST": "MST", "CHADT": "Saa za Mchana za Chatham", "AWST": "Saa za Wastani za Australia Magharibi", "HEPMX": "Saa za mchana za pasifiki za Mexico", "NZST": "Saa za Wastani za New Zealand", "NZDT": "Saa za Mchana za New Zealand", "AKST": "Saa za Wastani za Alaska", "WARST": "Saa za Majira ya joto za Magharibi mwa Argentina", "COST": "Saa za Majira ya joto za Colombia", "GYT": "Saa za Guyana", "UYT": "Saa za Wastani za Uruguay", "AEST": "Saa za Wastani za Mashariki mwa Australia", "JST": "Saa Wastani za Japan", "ACWDT": "Saa za Mchana za Magharibi ya Kati ya Australia", "LHDT": "Saa za Mchana za Lord Howe", "WITA": "Saa za Indonesia ya Kati", "HAST": "Saa za Wastani za Hawaii-Aleutian", "HEOG": "Saa za Majira ya joto za Greenland Magharibi", "CST": "Saa za Wastani za Kati", "CLT": "Saa za Wastani za Chile", "∅∅∅": "Saa za Majira ya joto za Amazon", "OESZ": "Saa za Majira ya joto za Mashariki mwa Ulaya", "AEDT": "Saa za Mchana za Mashariki mwa Australia", "BT": "Saa za Bhutan", "HNEG": "Saa za Wastani za Greenland Mashariki", "TMT": "Saa za Wastani za Turkmenistan", "CDT": "Saa za Mchana za Kati", "AKDT": "Saa za Mchana za Alaska", "HEEG": "Saa za Majira ya joto za Greenland Mashariki", "ART": "Saa za Wastani za Argentina", "GMT": "Saa za Greenwich", "AWDT": "Saa za Mchana za Australia Magharibi", "SRT": "Saa za Suriname"},
	}
}

// Locale returns the current translators string locale
func (sw *sw) Locale() string {
	return sw.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'sw'
func (sw *sw) PluralsCardinal() []locales.PluralRule {
	return sw.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'sw'
func (sw *sw) PluralsOrdinal() []locales.PluralRule {
	return sw.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'sw'
func (sw *sw) PluralsRange() []locales.PluralRule {
	return sw.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'sw'
func (sw *sw) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)
	i := int64(n)

	if i == 1 && v == 0 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'sw'
func (sw *sw) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'sw'
func (sw *sw) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {

	start := sw.CardinalPluralRule(num1, v1)
	end := sw.CardinalPluralRule(num2, v2)

	if start == locales.PluralRuleOne && end == locales.PluralRuleOther {
		return locales.PluralRuleOther
	} else if start == locales.PluralRuleOther && end == locales.PluralRuleOne {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther

}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (sw *sw) MonthAbbreviated(month time.Month) string {
	return sw.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (sw *sw) MonthsAbbreviated() []string {
	return sw.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (sw *sw) MonthNarrow(month time.Month) string {
	return sw.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (sw *sw) MonthsNarrow() []string {
	return sw.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (sw *sw) MonthWide(month time.Month) string {
	return sw.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (sw *sw) MonthsWide() []string {
	return sw.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (sw *sw) WeekdayAbbreviated(weekday time.Weekday) string {
	return sw.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (sw *sw) WeekdaysAbbreviated() []string {
	return sw.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (sw *sw) WeekdayNarrow(weekday time.Weekday) string {
	return sw.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (sw *sw) WeekdaysNarrow() []string {
	return sw.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (sw *sw) WeekdayShort(weekday time.Weekday) string {
	return sw.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (sw *sw) WeekdaysShort() []string {
	return sw.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (sw *sw) WeekdayWide(weekday time.Weekday) string {
	return sw.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (sw *sw) WeekdaysWide() []string {
	return sw.daysWide
}

// Decimal returns the decimal point of number
func (sw *sw) Decimal() string {
	return sw.decimal
}

// Group returns the group of number
func (sw *sw) Group() string {
	return sw.group
}

// Group returns the minus sign of number
func (sw *sw) Minus() string {
	return sw.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'sw' and handles both Whole and Real numbers based on 'v'
func (sw *sw) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, sw.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, sw.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, sw.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'sw' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (sw *sw) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 3
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, sw.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, sw.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, sw.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'sw'
func (sw *sw) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := sw.currencies[currency]
	l := len(s) + len(symbol) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, sw.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, sw.group[0])
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
		b = append(b, sw.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, sw.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'sw'
// in accounting notation.
func (sw *sw) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := sw.currencies[currency]
	l := len(s) + len(symbol) + 4 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, sw.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, sw.group[0])
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

		b = append(b, sw.currencyNegativePrefix[0])

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
			b = append(b, sw.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, sw.currencyNegativeSuffix...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'sw'
func (sw *sw) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'sw'
func (sw *sw) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, sw.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'sw'
func (sw *sw) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, sw.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'sw'
func (sw *sw) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, sw.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, sw.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'sw'
func (sw *sw) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, sw.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'sw'
func (sw *sw) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, sw.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, sw.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'sw'
func (sw *sw) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, sw.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, sw.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'sw'
func (sw *sw) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, sw.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, sw.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := sw.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
