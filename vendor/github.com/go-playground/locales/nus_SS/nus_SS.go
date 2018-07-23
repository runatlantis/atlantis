package nus_SS

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type nus_SS struct {
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

// New returns a new instance of translator for the 'nus_SS' locale
func New() locales.Translator {
	return &nus_SS{
		locale:                 "nus_SS",
		pluralsCardinal:        nil,
		pluralsOrdinal:         nil,
		pluralsRange:           nil,
		decimal:                ".",
		group:                  ",",
		timeSeparator:          ":",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyNegativePrefix: "(",
		currencyNegativeSuffix: ")",
		monthsAbbreviated:      []string{"", "Tiop", "Pɛt", "Duɔ̱ɔ̱", "Guak", "Duä", "Kor", "Pay", "Thoo", "Tɛɛ", "Laa", "Kur", "Tid"},
		monthsNarrow:           []string{"", "T", "P", "D", "G", "D", "K", "P", "T", "T", "L", "K", "T"},
		monthsWide:             []string{"", "Tiop thar pɛt", "Pɛt", "Duɔ̱ɔ̱ŋ", "Guak", "Duät", "Kornyoot", "Pay yie̱tni", "Tho̱o̱r", "Tɛɛr", "Laath", "Kur", "Tio̱p in di̱i̱t"},
		daysAbbreviated:        []string{"Cäŋ", "Jiec", "Rɛw", "Diɔ̱k", "Ŋuaan", "Dhieec", "Bäkɛl"},
		daysNarrow:             []string{"C", "J", "R", "D", "Ŋ", "D", "B"},
		daysWide:               []string{"Cäŋ kuɔth", "Jiec la̱t", "Rɛw lätni", "Diɔ̱k lätni", "Ŋuaan lätni", "Dhieec lätni", "Bäkɛl lätni"},
		periodsAbbreviated:     []string{"RW", "TŊ"},
		periodsWide:            []string{"RW", "TŊ"},
		erasAbbreviated:        []string{"AY", "ƐY"},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"A ka̱n Yecu ni dap", "Ɛ ca Yecu dap"},
		timezones:              map[string]string{"SGT": "SGT", "HKST": "HKST", "MST": "MST", "SRT": "SRT", "GYT": "GYT", "NZST": "NZST", "CLST": "CLST", "LHST": "LHST", "CAT": "CAT", "CLT": "CLT", "PST": "PST", "WAST": "WAST", "WESZ": "WESZ", "ACWST": "ACWST", "MEZ": "MEZ", "WARST": "WARST", "HAT": "HAT", "VET": "VET", "HNNOMX": "HNNOMX", "HAST": "HAST", "CST": "CST", "EDT": "EDT", "WAT": "WAT", "ACWDT": "ACWDT", "HNT": "HNT", "GMT": "GMT", "AEDT": "AEDT", "SAST": "SAST", "EST": "EST", "MESZ": "MESZ", "TMT": "TMT", "CHADT": "CHADT", "HECU": "HECU", "WEZ": "WEZ", "ECT": "ECT", "EAT": "EAT", "HADT": "HADT", "HNPMX": "HNPMX", "HEPM": "HEPM", "JST": "JST", "HNEG": "HNEG", "LHDT": "LHDT", "WIT": "WIT", "COST": "COST", "ChST": "ChST", "HEPMX": "HEPMX", "GFT": "GFT", "ACDT": "ACDT", "HEEG": "HEEG", "TMST": "TMST", "CDT": "CDT", "ACST": "ACST", "HENOMX": "HENOMX", "ART": "ART", "HNCU": "HNCU", "AST": "AST", "WART": "WART", "AKDT": "AKDT", "MDT": "MDT", "OEZ": "OEZ", "COT": "COT", "CHAST": "CHAST", "AEST": "AEST", "BOT": "BOT", "IST": "IST", "HNPM": "HNPM", "UYST": "UYST", "AWST": "AWST", "ADT": "ADT", "MYT": "MYT", "AKST": "AKST", "HKT": "HKT", "ARST": "ARST", "AWDT": "AWDT", "NZDT": "NZDT", "HNOG": "HNOG", "PDT": "PDT", "WIB": "WIB", "JDT": "JDT", "OESZ": "OESZ", "BT": "BT", "∅∅∅": "∅∅∅", "WITA": "WITA", "UYT": "UYT", "HEOG": "HEOG"},
	}
}

// Locale returns the current translators string locale
func (nus *nus_SS) Locale() string {
	return nus.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'nus_SS'
func (nus *nus_SS) PluralsCardinal() []locales.PluralRule {
	return nus.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'nus_SS'
func (nus *nus_SS) PluralsOrdinal() []locales.PluralRule {
	return nus.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'nus_SS'
func (nus *nus_SS) PluralsRange() []locales.PluralRule {
	return nus.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'nus_SS'
func (nus *nus_SS) CardinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'nus_SS'
func (nus *nus_SS) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'nus_SS'
func (nus *nus_SS) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (nus *nus_SS) MonthAbbreviated(month time.Month) string {
	return nus.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (nus *nus_SS) MonthsAbbreviated() []string {
	return nus.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (nus *nus_SS) MonthNarrow(month time.Month) string {
	return nus.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (nus *nus_SS) MonthsNarrow() []string {
	return nus.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (nus *nus_SS) MonthWide(month time.Month) string {
	return nus.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (nus *nus_SS) MonthsWide() []string {
	return nus.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (nus *nus_SS) WeekdayAbbreviated(weekday time.Weekday) string {
	return nus.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (nus *nus_SS) WeekdaysAbbreviated() []string {
	return nus.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (nus *nus_SS) WeekdayNarrow(weekday time.Weekday) string {
	return nus.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (nus *nus_SS) WeekdaysNarrow() []string {
	return nus.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (nus *nus_SS) WeekdayShort(weekday time.Weekday) string {
	return nus.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (nus *nus_SS) WeekdaysShort() []string {
	return nus.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (nus *nus_SS) WeekdayWide(weekday time.Weekday) string {
	return nus.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (nus *nus_SS) WeekdaysWide() []string {
	return nus.daysWide
}

// Decimal returns the decimal point of number
func (nus *nus_SS) Decimal() string {
	return nus.decimal
}

// Group returns the group of number
func (nus *nus_SS) Group() string {
	return nus.group
}

// Group returns the minus sign of number
func (nus *nus_SS) Minus() string {
	return nus.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'nus_SS' and handles both Whole and Real numbers based on 'v'
func (nus *nus_SS) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 1 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, nus.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, nus.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, nus.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'nus_SS' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (nus *nus_SS) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 1
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, nus.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, nus.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, nus.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'nus_SS'
func (nus *nus_SS) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := nus.currencies[currency]
	l := len(s) + len(symbol) + 1 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, nus.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, nus.group[0])
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
		b = append(b, nus.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, nus.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'nus_SS'
// in accounting notation.
func (nus *nus_SS) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := nus.currencies[currency]
	l := len(s) + len(symbol) + 3 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, nus.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, nus.group[0])
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

		b = append(b, nus.currencyNegativePrefix[0])

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
			b = append(b, nus.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, nus.currencyNegativeSuffix...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'nus_SS'
func (nus *nus_SS) FmtDateShort(t time.Time) string {

	b := make([]byte, 0, 32)

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

// FmtDateMedium returns the medium date representation of 't' for 'nus_SS'
func (nus *nus_SS) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, nus.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'nus_SS'
func (nus *nus_SS) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, nus.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'nus_SS'
func (nus *nus_SS) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, nus.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, nus.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'nus_SS'
func (nus *nus_SS) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, nus.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, nus.periodsAbbreviated[0]...)
	} else {
		b = append(b, nus.periodsAbbreviated[1]...)
	}

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'nus_SS'
func (nus *nus_SS) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, nus.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, nus.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, nus.periodsAbbreviated[0]...)
	} else {
		b = append(b, nus.periodsAbbreviated[1]...)
	}

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'nus_SS'
func (nus *nus_SS) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	tz, _ := t.Zone()
	b = append(b, tz...)

	b = append(b, []byte{0x20}...)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, nus.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, nus.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, nus.periodsAbbreviated[0]...)
	} else {
		b = append(b, nus.periodsAbbreviated[1]...)
	}

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'nus_SS'
func (nus *nus_SS) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	tz, _ := t.Zone()

	if btz, ok := nus.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	b = append(b, []byte{0x20}...)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, nus.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, nus.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, nus.periodsAbbreviated[0]...)
	} else {
		b = append(b, nus.periodsAbbreviated[1]...)
	}

	return string(b)
}
