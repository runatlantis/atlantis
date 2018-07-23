package seh

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type seh struct {
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

// New returns a new instance of translator for the 'seh' locale
func New() locales.Translator {
	return &seh{
		locale:            "seh",
		pluralsCardinal:   []locales.PluralRule{2, 6},
		pluralsOrdinal:    nil,
		pluralsRange:      nil,
		decimal:           ",",
		group:             ".",
		timeSeparator:     ":",
		currencies:        []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MTn", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		monthsAbbreviated: []string{"", "Jan", "Fev", "Mar", "Abr", "Mai", "Jun", "Jul", "Aug", "Set", "Otu", "Nov", "Dec"},
		monthsNarrow:      []string{"", "J", "F", "M", "A", "M", "J", "J", "A", "S", "O", "N", "D"},
		monthsWide:        []string{"", "Janeiro", "Fevreiro", "Marco", "Abril", "Maio", "Junho", "Julho", "Augusto", "Setembro", "Otubro", "Novembro", "Decembro"},
		daysAbbreviated:   []string{"Dim", "Pos", "Pir", "Tat", "Nai", "Sha", "Sab"},
		daysNarrow:        []string{"D", "P", "C", "T", "N", "S", "S"},
		daysWide:          []string{"Dimingu", "Chiposi", "Chipiri", "Chitatu", "Chinai", "Chishanu", "Sabudu"},
		erasAbbreviated:   []string{"AC", "AD"},
		erasNarrow:        []string{"", ""},
		erasWide:          []string{"Antes de Cristo", "Anno Domini"},
		timezones:         map[string]string{"UYT": "UYT", "HEPMX": "HEPMX", "JDT": "JDT", "EST": "EST", "SRT": "SRT", "WIT": "WIT", "ARST": "ARST", "HECU": "HECU", "CST": "CST", "NZST": "NZST", "SGT": "SGT", "HNEG": "HNEG", "ART": "ART", "HKST": "HKST", "MEZ": "MEZ", "BOT": "BOT", "ECT": "ECT", "GMT": "GMT", "ADT": "ADT", "ACWST": "ACWST", "WARST": "WARST", "HNT": "HNT", "CLST": "CLST", "MESZ": "MESZ", "HEPM": "HEPM", "WITA": "WITA", "HAT": "HAT", "TMST": "TMST", "ChST": "ChST", "AWST": "AWST", "GFT": "GFT", "BT": "BT", "∅∅∅": "∅∅∅", "WART": "WART", "AKST": "AKST", "EAT": "EAT", "TMT": "TMT", "PDT": "PDT", "CAT": "CAT", "UYST": "UYST", "WAT": "WAT", "NZDT": "NZDT", "ACWDT": "ACWDT", "HEEG": "HEEG", "LHDT": "LHDT", "COT": "COT", "CHAST": "CHAST", "WIB": "WIB", "WESZ": "WESZ", "IST": "IST", "HENOMX": "HENOMX", "CLT": "CLT", "AEDT": "AEDT", "HNCU": "HNCU", "AWDT": "AWDT", "JST": "JST", "ACST": "ACST", "HEOG": "HEOG", "HNPM": "HNPM", "OEZ": "OEZ", "OESZ": "OESZ", "CHADT": "CHADT", "CDT": "CDT", "EDT": "EDT", "HKT": "HKT", "LHST": "LHST", "MST": "MST", "AST": "AST", "WAST": "WAST", "AEST": "AEST", "SAST": "SAST", "ACDT": "ACDT", "HAST": "HAST", "COST": "COST", "HNPMX": "HNPMX", "MYT": "MYT", "VET": "VET", "HNNOMX": "HNNOMX", "PST": "PST", "GYT": "GYT", "WEZ": "WEZ", "AKDT": "AKDT", "HNOG": "HNOG", "MDT": "MDT", "HADT": "HADT"},
	}
}

// Locale returns the current translators string locale
func (seh *seh) Locale() string {
	return seh.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'seh'
func (seh *seh) PluralsCardinal() []locales.PluralRule {
	return seh.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'seh'
func (seh *seh) PluralsOrdinal() []locales.PluralRule {
	return seh.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'seh'
func (seh *seh) PluralsRange() []locales.PluralRule {
	return seh.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'seh'
func (seh *seh) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'seh'
func (seh *seh) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'seh'
func (seh *seh) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (seh *seh) MonthAbbreviated(month time.Month) string {
	return seh.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (seh *seh) MonthsAbbreviated() []string {
	return seh.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (seh *seh) MonthNarrow(month time.Month) string {
	return seh.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (seh *seh) MonthsNarrow() []string {
	return seh.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (seh *seh) MonthWide(month time.Month) string {
	return seh.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (seh *seh) MonthsWide() []string {
	return seh.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (seh *seh) WeekdayAbbreviated(weekday time.Weekday) string {
	return seh.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (seh *seh) WeekdaysAbbreviated() []string {
	return seh.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (seh *seh) WeekdayNarrow(weekday time.Weekday) string {
	return seh.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (seh *seh) WeekdaysNarrow() []string {
	return seh.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (seh *seh) WeekdayShort(weekday time.Weekday) string {
	return seh.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (seh *seh) WeekdaysShort() []string {
	return seh.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (seh *seh) WeekdayWide(weekday time.Weekday) string {
	return seh.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (seh *seh) WeekdaysWide() []string {
	return seh.daysWide
}

// Decimal returns the decimal point of number
func (seh *seh) Decimal() string {
	return seh.decimal
}

// Group returns the group of number
func (seh *seh) Group() string {
	return seh.group
}

// Group returns the minus sign of number
func (seh *seh) Minus() string {
	return seh.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'seh' and handles both Whole and Real numbers based on 'v'
func (seh *seh) FmtNumber(num float64, v uint64) string {

	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'seh' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (seh *seh) FmtPercent(num float64, v uint64) string {
	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'seh'
func (seh *seh) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := seh.currencies[currency]
	l := len(s) + len(symbol) + 1 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, seh.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, seh.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, seh.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, seh.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	b = append(b, symbol...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'seh'
// in accounting notation.
func (seh *seh) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := seh.currencies[currency]
	l := len(s) + len(symbol) + 1 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, seh.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, seh.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {

		b = append(b, seh.minus[0])

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, seh.decimal...)
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

// FmtDateShort returns the short date representation of 't' for 'seh'
func (seh *seh) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'seh'
func (seh *seh) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20, 0x64, 0x65}...)
	b = append(b, []byte{0x20}...)
	b = append(b, seh.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20, 0x64, 0x65}...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'seh'
func (seh *seh) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20, 0x64, 0x65}...)
	b = append(b, []byte{0x20}...)
	b = append(b, seh.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20, 0x64, 0x65}...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'seh'
func (seh *seh) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, seh.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20, 0x64, 0x65}...)
	b = append(b, []byte{0x20}...)
	b = append(b, seh.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20, 0x64, 0x65}...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'seh'
func (seh *seh) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, seh.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'seh'
func (seh *seh) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, seh.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, seh.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'seh'
func (seh *seh) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, seh.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, seh.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'seh'
func (seh *seh) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, seh.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, seh.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := seh.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
