package twq_NE

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type twq_NE struct {
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

// New returns a new instance of translator for the 'twq_NE' locale
func New() locales.Translator {
	return &twq_NE{
		locale:             "twq_NE",
		pluralsCardinal:    nil,
		pluralsOrdinal:     nil,
		pluralsRange:       nil,
		decimal:            ".",
		group:              " ",
		timeSeparator:      ":",
		currencies:         []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		monthsAbbreviated:  []string{"", "Žan", "Fee", "Mar", "Awi", "Me", "Žuw", "Žuy", "Ut", "Sek", "Okt", "Noo", "Dee"},
		monthsNarrow:       []string{"", "Ž", "F", "M", "A", "M", "Ž", "Ž", "U", "S", "O", "N", "D"},
		monthsWide:         []string{"", "Žanwiye", "Feewiriye", "Marsi", "Awiril", "Me", "Žuweŋ", "Žuyye", "Ut", "Sektanbur", "Oktoobur", "Noowanbur", "Deesanbur"},
		daysAbbreviated:    []string{"Alh", "Ati", "Ata", "Ala", "Alm", "Alz", "Asi"},
		daysNarrow:         []string{"H", "T", "T", "L", "L", "L", "S"},
		daysWide:           []string{"Alhadi", "Atinni", "Atalaata", "Alarba", "Alhamiisa", "Alzuma", "Asibti"},
		periodsAbbreviated: []string{"Subbaahi", "Zaarikay b"},
		periodsWide:        []string{"Subbaahi", "Zaarikay b"},
		erasAbbreviated:    []string{"IJ", "IZ"},
		erasNarrow:         []string{"", ""},
		erasWide:           []string{"Isaa jine", "Isaa zamanoo"},
		timezones:          map[string]string{"CLT": "CLT", "CHAST": "CHAST", "HECU": "HECU", "BOT": "BOT", "ACWST": "ACWST", "IST": "IST", "ADT": "ADT", "ECT": "ECT", "GYT": "GYT", "WIB": "WIB", "HNOG": "HNOG", "EAT": "EAT", "HAST": "HAST", "ACDT": "ACDT", "MESZ": "MESZ", "VET": "VET", "∅∅∅": "∅∅∅", "MST": "MST", "WESZ": "WESZ", "GFT": "GFT", "JST": "JST", "HEPM": "HEPM", "HNNOMX": "HNNOMX", "HNPMX": "HNPMX", "MEZ": "MEZ", "LHST": "LHST", "HNT": "HNT", "HNPM": "HNPM", "GMT": "GMT", "UYT": "UYT", "AWDT": "AWDT", "AST": "AST", "WAST": "WAST", "NZDT": "NZDT", "AKST": "AKST", "SAST": "SAST", "WEZ": "WEZ", "MYT": "MYT", "PST": "PST", "WAT": "WAT", "HEOG": "HEOG", "TMT": "TMT", "TMST": "TMST", "CLST": "CLST", "COT": "COT", "HEPMX": "HEPMX", "CDT": "CDT", "PDT": "PDT", "HEEG": "HEEG", "WITA": "WITA", "ChST": "ChST", "CHADT": "CHADT", "NZST": "NZST", "EST": "EST", "UYST": "UYST", "AWST": "AWST", "MDT": "MDT", "LHDT": "LHDT", "WARST": "WARST", "HADT": "HADT", "HNCU": "HNCU", "CST": "CST", "SGT": "SGT", "EDT": "EDT", "ACST": "ACST", "AEDT": "AEDT", "HKT": "HKT", "WIT": "WIT", "CAT": "CAT", "ART": "ART", "ACWDT": "ACWDT", "HAT": "HAT", "HENOMX": "HENOMX", "SRT": "SRT", "COST": "COST", "AEST": "AEST", "AKDT": "AKDT", "HNEG": "HNEG", "HKST": "HKST", "WART": "WART", "JDT": "JDT", "BT": "BT", "OEZ": "OEZ", "OESZ": "OESZ", "ARST": "ARST"},
	}
}

// Locale returns the current translators string locale
func (twq *twq_NE) Locale() string {
	return twq.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'twq_NE'
func (twq *twq_NE) PluralsCardinal() []locales.PluralRule {
	return twq.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'twq_NE'
func (twq *twq_NE) PluralsOrdinal() []locales.PluralRule {
	return twq.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'twq_NE'
func (twq *twq_NE) PluralsRange() []locales.PluralRule {
	return twq.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'twq_NE'
func (twq *twq_NE) CardinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'twq_NE'
func (twq *twq_NE) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'twq_NE'
func (twq *twq_NE) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (twq *twq_NE) MonthAbbreviated(month time.Month) string {
	return twq.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (twq *twq_NE) MonthsAbbreviated() []string {
	return twq.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (twq *twq_NE) MonthNarrow(month time.Month) string {
	return twq.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (twq *twq_NE) MonthsNarrow() []string {
	return twq.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (twq *twq_NE) MonthWide(month time.Month) string {
	return twq.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (twq *twq_NE) MonthsWide() []string {
	return twq.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (twq *twq_NE) WeekdayAbbreviated(weekday time.Weekday) string {
	return twq.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (twq *twq_NE) WeekdaysAbbreviated() []string {
	return twq.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (twq *twq_NE) WeekdayNarrow(weekday time.Weekday) string {
	return twq.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (twq *twq_NE) WeekdaysNarrow() []string {
	return twq.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (twq *twq_NE) WeekdayShort(weekday time.Weekday) string {
	return twq.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (twq *twq_NE) WeekdaysShort() []string {
	return twq.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (twq *twq_NE) WeekdayWide(weekday time.Weekday) string {
	return twq.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (twq *twq_NE) WeekdaysWide() []string {
	return twq.daysWide
}

// Decimal returns the decimal point of number
func (twq *twq_NE) Decimal() string {
	return twq.decimal
}

// Group returns the group of number
func (twq *twq_NE) Group() string {
	return twq.group
}

// Group returns the minus sign of number
func (twq *twq_NE) Minus() string {
	return twq.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'twq_NE' and handles both Whole and Real numbers based on 'v'
func (twq *twq_NE) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 1 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, twq.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(twq.group) - 1; j >= 0; j-- {
					b = append(b, twq.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, twq.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'twq_NE' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (twq *twq_NE) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 1
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, twq.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, twq.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, twq.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'twq_NE'
func (twq *twq_NE) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := twq.currencies[currency]
	l := len(s) + len(symbol) + 1 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, twq.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(twq.group) - 1; j >= 0; j-- {
					b = append(b, twq.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, twq.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, twq.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	b = append(b, symbol...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'twq_NE'
// in accounting notation.
func (twq *twq_NE) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := twq.currencies[currency]
	l := len(s) + len(symbol) + 1 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, twq.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(twq.group) - 1; j >= 0; j-- {
					b = append(b, twq.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {

		b = append(b, twq.minus[0])

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, twq.decimal...)
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

// FmtDateShort returns the short date representation of 't' for 'twq_NE'
func (twq *twq_NE) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'twq_NE'
func (twq *twq_NE) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, twq.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'twq_NE'
func (twq *twq_NE) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, twq.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'twq_NE'
func (twq *twq_NE) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, twq.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, twq.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'twq_NE'
func (twq *twq_NE) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, twq.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'twq_NE'
func (twq *twq_NE) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, twq.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, twq.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'twq_NE'
func (twq *twq_NE) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, twq.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, twq.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'twq_NE'
func (twq *twq_NE) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, twq.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, twq.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := twq.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
