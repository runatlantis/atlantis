package lg

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type lg struct {
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

// New returns a new instance of translator for the 'lg' locale
func New() locales.Translator {
	return &lg{
		locale:            "lg",
		pluralsCardinal:   []locales.PluralRule{2, 6},
		pluralsOrdinal:    nil,
		pluralsRange:      nil,
		timeSeparator:     ":",
		currencies:        []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "USh", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		monthsAbbreviated: []string{"", "Jan", "Feb", "Mar", "Apu", "Maa", "Juu", "Jul", "Agu", "Seb", "Oki", "Nov", "Des"},
		monthsNarrow:      []string{"", "J", "F", "M", "A", "M", "J", "J", "A", "S", "O", "N", "D"},
		monthsWide:        []string{"", "Janwaliyo", "Febwaliyo", "Marisi", "Apuli", "Maayi", "Juuni", "Julaayi", "Agusito", "Sebuttemba", "Okitobba", "Novemba", "Desemba"},
		daysAbbreviated:   []string{"Sab", "Bal", "Lw2", "Lw3", "Lw4", "Lw5", "Lw6"},
		daysNarrow:        []string{"S", "B", "L", "L", "L", "L", "L"},
		daysWide:          []string{"Sabbiiti", "Balaza", "Lwakubiri", "Lwakusatu", "Lwakuna", "Lwakutaano", "Lwamukaaga"},
		erasAbbreviated:   []string{"BC", "AD"},
		erasNarrow:        []string{"", ""},
		erasWide:          []string{"Kulisito nga tannaza", "Bukya Kulisito Azaal"},
		timezones:         map[string]string{"LHDT": "LHDT", "VET": "VET", "OESZ": "OESZ", "∅∅∅": "∅∅∅", "PST": "PST", "HNPMX": "HNPMX", "HEOG": "HEOG", "WARST": "WARST", "EAT": "EAT", "ChST": "ChST", "EST": "EST", "ACST": "ACST", "WAT": "WAT", "ARST": "ARST", "HNCU": "HNCU", "AEST": "AEST", "AEDT": "AEDT", "HAT": "HAT", "WIT": "WIT", "UYT": "UYT", "AST": "AST", "ACDT": "ACDT", "MYT": "MYT", "AKDT": "AKDT", "ECT": "ECT", "ART": "ART", "HECU": "HECU", "SAST": "SAST", "WESZ": "WESZ", "OEZ": "OEZ", "CDT": "CDT", "JST": "JST", "GFT": "GFT", "BT": "BT", "AKST": "AKST", "ACWDT": "ACWDT", "CAT": "CAT", "CST": "CST", "PDT": "PDT", "WIB": "WIB", "ADT": "ADT", "LHST": "LHST", "HNPM": "HNPM", "HEPM": "HEPM", "TMT": "TMT", "HAST": "HAST", "COT": "COT", "CHADT": "CHADT", "SGT": "SGT", "HNEG": "HNEG", "CLST": "CLST", "MST": "MST", "MDT": "MDT", "BOT": "BOT", "WART": "WART", "CLT": "CLT", "TMST": "TMST", "NZST": "NZST", "MESZ": "MESZ", "WEZ": "WEZ", "ACWST": "ACWST", "SRT": "SRT", "COST": "COST", "UYST": "UYST", "CHAST": "CHAST", "AWST": "AWST", "JDT": "JDT", "HKST": "HKST", "IST": "IST", "HNT": "HNT", "HENOMX": "HENOMX", "HEPMX": "HEPMX", "NZDT": "NZDT", "EDT": "EDT", "HEEG": "HEEG", "HNOG": "HNOG", "MEZ": "MEZ", "HNNOMX": "HNNOMX", "GMT": "GMT", "GYT": "GYT", "AWDT": "AWDT", "WAST": "WAST", "HADT": "HADT", "HKT": "HKT", "WITA": "WITA"},
	}
}

// Locale returns the current translators string locale
func (lg *lg) Locale() string {
	return lg.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'lg'
func (lg *lg) PluralsCardinal() []locales.PluralRule {
	return lg.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'lg'
func (lg *lg) PluralsOrdinal() []locales.PluralRule {
	return lg.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'lg'
func (lg *lg) PluralsRange() []locales.PluralRule {
	return lg.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'lg'
func (lg *lg) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'lg'
func (lg *lg) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'lg'
func (lg *lg) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (lg *lg) MonthAbbreviated(month time.Month) string {
	return lg.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (lg *lg) MonthsAbbreviated() []string {
	return lg.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (lg *lg) MonthNarrow(month time.Month) string {
	return lg.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (lg *lg) MonthsNarrow() []string {
	return lg.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (lg *lg) MonthWide(month time.Month) string {
	return lg.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (lg *lg) MonthsWide() []string {
	return lg.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (lg *lg) WeekdayAbbreviated(weekday time.Weekday) string {
	return lg.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (lg *lg) WeekdaysAbbreviated() []string {
	return lg.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (lg *lg) WeekdayNarrow(weekday time.Weekday) string {
	return lg.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (lg *lg) WeekdaysNarrow() []string {
	return lg.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (lg *lg) WeekdayShort(weekday time.Weekday) string {
	return lg.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (lg *lg) WeekdaysShort() []string {
	return lg.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (lg *lg) WeekdayWide(weekday time.Weekday) string {
	return lg.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (lg *lg) WeekdaysWide() []string {
	return lg.daysWide
}

// Decimal returns the decimal point of number
func (lg *lg) Decimal() string {
	return lg.decimal
}

// Group returns the group of number
func (lg *lg) Group() string {
	return lg.group
}

// Group returns the minus sign of number
func (lg *lg) Minus() string {
	return lg.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'lg' and handles both Whole and Real numbers based on 'v'
func (lg *lg) FmtNumber(num float64, v uint64) string {

	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'lg' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (lg *lg) FmtPercent(num float64, v uint64) string {
	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'lg'
func (lg *lg) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := lg.currencies[currency]
	l := len(s) + len(symbol) + 0
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, lg.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, lg.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, lg.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, lg.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	b = append(b, symbol...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'lg'
// in accounting notation.
func (lg *lg) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := lg.currencies[currency]
	l := len(s) + len(symbol) + 0
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, lg.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, lg.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {

		b = append(b, lg.minus[0])

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, lg.decimal...)
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

// FmtDateShort returns the short date representation of 't' for 'lg'
func (lg *lg) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'lg'
func (lg *lg) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, lg.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'lg'
func (lg *lg) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, lg.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'lg'
func (lg *lg) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, lg.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, lg.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'lg'
func (lg *lg) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, lg.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'lg'
func (lg *lg) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, lg.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, lg.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'lg'
func (lg *lg) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, lg.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, lg.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'lg'
func (lg *lg) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, lg.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, lg.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := lg.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
