package twq

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type twq struct {
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

// New returns a new instance of translator for the 'twq' locale
func New() locales.Translator {
	return &twq{
		locale:             "twq",
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
		timezones:          map[string]string{"CDT": "CDT", "AWST": "AWST", "WITA": "WITA", "HEEG": "HEEG", "AWDT": "AWDT", "AEST": "AEST", "EDT": "EDT", "HNNOMX": "HNNOMX", "OESZ": "OESZ", "ART": "ART", "UYST": "UYST", "ChST": "ChST", "NZDT": "NZDT", "ACWST": "ACWST", "MEZ": "MEZ", "WART": "WART", "HAST": "HAST", "PDT": "PDT", "WIB": "WIB", "JST": "JST", "ACWDT": "ACWDT", "HKT": "HKT", "CAT": "CAT", "COST": "COST", "ACST": "ACST", "HKST": "HKST", "HNT": "HNT", "HNPM": "HNPM", "TMST": "TMST", "OEZ": "OEZ", "GMT": "GMT", "WAST": "WAST", "HEOG": "HEOG", "MYT": "MYT", "MST": "MST", "HNCU": "HNCU", "WEZ": "WEZ", "AKDT": "AKDT", "SGT": "SGT", "MESZ": "MESZ", "CLT": "CLT", "SAST": "SAST", "EAT": "EAT", "MDT": "MDT", "SRT": "SRT", "CHADT": "CHADT", "HNEG": "HNEG", "ARST": "ARST", "AKST": "AKST", "HAT": "HAT", "VET": "VET", "HEPM": "HEPM", "HENOMX": "HENOMX", "COT": "COT", "GYT": "GYT", "HECU": "HECU", "HEPMX": "HEPMX", "LHST": "LHST", "NZST": "NZST", "TMT": "TMT", "CST": "CST", "ADT": "ADT", "BOT": "BOT", "ACDT": "ACDT", "WIT": "WIT", "UYT": "UYT", "PST": "PST", "HNPMX": "HNPMX", "AST": "AST", "AEDT": "AEDT", "GFT": "GFT", "∅∅∅": "∅∅∅", "IST": "IST", "WARST": "WARST", "HADT": "HADT", "CHAST": "CHAST", "ECT": "ECT", "HNOG": "HNOG", "LHDT": "LHDT", "CLST": "CLST", "WAT": "WAT", "WESZ": "WESZ", "JDT": "JDT", "BT": "BT", "EST": "EST"},
	}
}

// Locale returns the current translators string locale
func (twq *twq) Locale() string {
	return twq.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'twq'
func (twq *twq) PluralsCardinal() []locales.PluralRule {
	return twq.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'twq'
func (twq *twq) PluralsOrdinal() []locales.PluralRule {
	return twq.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'twq'
func (twq *twq) PluralsRange() []locales.PluralRule {
	return twq.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'twq'
func (twq *twq) CardinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'twq'
func (twq *twq) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'twq'
func (twq *twq) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (twq *twq) MonthAbbreviated(month time.Month) string {
	return twq.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (twq *twq) MonthsAbbreviated() []string {
	return twq.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (twq *twq) MonthNarrow(month time.Month) string {
	return twq.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (twq *twq) MonthsNarrow() []string {
	return twq.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (twq *twq) MonthWide(month time.Month) string {
	return twq.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (twq *twq) MonthsWide() []string {
	return twq.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (twq *twq) WeekdayAbbreviated(weekday time.Weekday) string {
	return twq.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (twq *twq) WeekdaysAbbreviated() []string {
	return twq.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (twq *twq) WeekdayNarrow(weekday time.Weekday) string {
	return twq.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (twq *twq) WeekdaysNarrow() []string {
	return twq.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (twq *twq) WeekdayShort(weekday time.Weekday) string {
	return twq.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (twq *twq) WeekdaysShort() []string {
	return twq.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (twq *twq) WeekdayWide(weekday time.Weekday) string {
	return twq.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (twq *twq) WeekdaysWide() []string {
	return twq.daysWide
}

// Decimal returns the decimal point of number
func (twq *twq) Decimal() string {
	return twq.decimal
}

// Group returns the group of number
func (twq *twq) Group() string {
	return twq.group
}

// Group returns the minus sign of number
func (twq *twq) Minus() string {
	return twq.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'twq' and handles both Whole and Real numbers based on 'v'
func (twq *twq) FmtNumber(num float64, v uint64) string {

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

// FmtPercent returns 'num' with digits/precision of 'v' for 'twq' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (twq *twq) FmtPercent(num float64, v uint64) string {
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

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'twq'
func (twq *twq) FmtCurrency(num float64, v uint64, currency currency.Type) string {

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

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'twq'
// in accounting notation.
func (twq *twq) FmtAccounting(num float64, v uint64, currency currency.Type) string {

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

// FmtDateShort returns the short date representation of 't' for 'twq'
func (twq *twq) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'twq'
func (twq *twq) FmtDateMedium(t time.Time) string {

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

// FmtDateLong returns the long date representation of 't' for 'twq'
func (twq *twq) FmtDateLong(t time.Time) string {

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

// FmtDateFull returns the full date representation of 't' for 'twq'
func (twq *twq) FmtDateFull(t time.Time) string {

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

// FmtTimeShort returns the short time representation of 't' for 'twq'
func (twq *twq) FmtTimeShort(t time.Time) string {

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

// FmtTimeMedium returns the medium time representation of 't' for 'twq'
func (twq *twq) FmtTimeMedium(t time.Time) string {

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

// FmtTimeLong returns the long time representation of 't' for 'twq'
func (twq *twq) FmtTimeLong(t time.Time) string {

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

// FmtTimeFull returns the full time representation of 't' for 'twq'
func (twq *twq) FmtTimeFull(t time.Time) string {

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
