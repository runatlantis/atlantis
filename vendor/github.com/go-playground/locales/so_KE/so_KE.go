package so_KE

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type so_KE struct {
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

// New returns a new instance of translator for the 'so_KE' locale
func New() locales.Translator {
	return &so_KE{
		locale:             "so_KE",
		pluralsCardinal:    []locales.PluralRule{2, 6},
		pluralsOrdinal:     nil,
		pluralsRange:       nil,
		decimal:            ".",
		group:              ",",
		timeSeparator:      ":",
		currencies:         []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "Ksh", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		monthsAbbreviated:  []string{"", "Kob", "Lab", "Sad", "Afr", "Sha", "Lix", "Tod", "Sid", "Sag", "Tob", "KIT", "LIT"},
		monthsNarrow:       []string{"", "K", "L", "S", "A", "S", "L", "T", "S", "S", "T", "K", "L"},
		monthsWide:         []string{"", "Bisha Koobaad", "Bisha Labaad", "Bisha Saddexaad", "Bisha Afraad", "Bisha Shanaad", "Bisha Lixaad", "Bisha Todobaad", "Bisha Sideedaad", "Bisha Sagaalaad", "Bisha Tobnaad", "Bisha Kow iyo Tobnaad", "Bisha Laba iyo Tobnaad"},
		daysAbbreviated:    []string{"Axd", "Isn", "Tal", "Arb", "Kha", "Jim", "Sab"},
		daysNarrow:         []string{"A", "I", "T", "A", "Kh", "J", "S"},
		daysShort:          []string{"Axd", "Isn", "Tal", "Arb", "Kha", "Jim", "Sab"},
		daysWide:           []string{"Axad", "Isniin", "Talaado", "Arbaco", "Khamiis", "Jimco", "Sabti"},
		periodsAbbreviated: []string{"sn.", "gn."},
		periodsNarrow:      []string{"sn.", "gn."},
		periodsWide:        []string{"sn.", "gn."},
		erasAbbreviated:    []string{"CK", "CD"},
		erasNarrow:         []string{"", ""},
		erasWide:           []string{"CK", "CD"},
		timezones:          map[string]string{"IST": "IST", "TMT": "TMT", "WIB": "WIB", "BT": "BT", "GFT": "GFT", "GYT": "GYT", "UYST": "UYST", "AST": "AST", "SAST": "SAST", "JDT": "JDT", "EST": "EST", "HEPM": "HEPM", "OEZ": "OEZ", "HADT": "HADT", "CDT": "CDT", "HEPMX": "HEPMX", "PDT": "PDT", "JST": "JST", "EDT": "EDT", "ACWST": "ACWST", "HEEG": "HEEG", "HNNOMX": "HNNOMX", "OESZ": "OESZ", "CST": "CST", "BOT": "BOT", "SGT": "SGT", "ECT": "ECT", "HKST": "HKST", "HECU": "HECU", "WESZ": "WESZ", "AWST": "AWST", "AKST": "AKST", "LHST": "LHST", "EAT": "EAT", "CLT": "CLT", "ARST": "ARST", "CHAST": "CHAST", "∅∅∅": "∅∅∅", "HAT": "HAT", "CAT": "CAT", "COST": "Waqtiyada Xagaaga Kolambiya", "WEZ": "WEZ", "HNEG": "HNEG", "WART": "WART", "MDT": "MDT", "SRT": "SRT", "ACST": "ACST", "MESZ": "MESZ", "CHADT": "CHADT", "AEST": "AEST", "WAT": "WAT", "NZST": "NZST", "HNT": "HNT", "WIT": "WIT", "GMT": "GMT", "AKDT": "AKDT", "ACDT": "ACDT", "MEZ": "MEZ", "HKT": "HKT", "NZDT": "NZDT", "MYT": "MYT", "TMST": "TMST", "UYT": "UYT", "ChST": "ChST", "PST": "PST", "ADT": "ADT", "WAST": "WAST", "ACWDT": "ACWDT", "WARST": "WARST", "HAST": "HAST", "ART": "ART", "AEDT": "AEDT", "HEOG": "HEOG", "LHDT": "LHDT", "VET": "VET", "HENOMX": "HENOMX", "COT": "Waqtiyada Caadiga ah ee kolambiya", "AWDT": "AWDT", "HNPM": "HNPM", "WITA": "WITA", "MST": "MST", "HNCU": "HNCU", "HNOG": "HNOG", "CLST": "CLST", "HNPMX": "HNPMX"},
	}
}

// Locale returns the current translators string locale
func (so *so_KE) Locale() string {
	return so.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'so_KE'
func (so *so_KE) PluralsCardinal() []locales.PluralRule {
	return so.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'so_KE'
func (so *so_KE) PluralsOrdinal() []locales.PluralRule {
	return so.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'so_KE'
func (so *so_KE) PluralsRange() []locales.PluralRule {
	return so.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'so_KE'
func (so *so_KE) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'so_KE'
func (so *so_KE) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'so_KE'
func (so *so_KE) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (so *so_KE) MonthAbbreviated(month time.Month) string {
	return so.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (so *so_KE) MonthsAbbreviated() []string {
	return so.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (so *so_KE) MonthNarrow(month time.Month) string {
	return so.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (so *so_KE) MonthsNarrow() []string {
	return so.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (so *so_KE) MonthWide(month time.Month) string {
	return so.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (so *so_KE) MonthsWide() []string {
	return so.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (so *so_KE) WeekdayAbbreviated(weekday time.Weekday) string {
	return so.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (so *so_KE) WeekdaysAbbreviated() []string {
	return so.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (so *so_KE) WeekdayNarrow(weekday time.Weekday) string {
	return so.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (so *so_KE) WeekdaysNarrow() []string {
	return so.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (so *so_KE) WeekdayShort(weekday time.Weekday) string {
	return so.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (so *so_KE) WeekdaysShort() []string {
	return so.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (so *so_KE) WeekdayWide(weekday time.Weekday) string {
	return so.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (so *so_KE) WeekdaysWide() []string {
	return so.daysWide
}

// Decimal returns the decimal point of number
func (so *so_KE) Decimal() string {
	return so.decimal
}

// Group returns the group of number
func (so *so_KE) Group() string {
	return so.group
}

// Group returns the minus sign of number
func (so *so_KE) Minus() string {
	return so.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'so_KE' and handles both Whole and Real numbers based on 'v'
func (so *so_KE) FmtNumber(num float64, v uint64) string {

	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'so_KE' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (so *so_KE) FmtPercent(num float64, v uint64) string {
	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'so_KE'
func (so *so_KE) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := so.currencies[currency]
	l := len(s) + len(symbol) + 1 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, so.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, so.group[0])
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
		b = append(b, so.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, so.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'so_KE'
// in accounting notation.
func (so *so_KE) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := so.currencies[currency]
	l := len(s) + len(symbol) + 1 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, so.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, so.group[0])
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

		b = append(b, so.minus[0])

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
			b = append(b, so.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'so_KE'
func (so *so_KE) FmtDateShort(t time.Time) string {

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

	if t.Year() > 9 {
		b = append(b, strconv.Itoa(t.Year())[2:]...)
	} else {
		b = append(b, strconv.Itoa(t.Year())[1:]...)
	}

	return string(b)
}

// FmtDateMedium returns the medium date representation of 't' for 'so_KE'
func (so *so_KE) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Day() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2d}...)
	b = append(b, so.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x2d}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'so_KE'
func (so *so_KE) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Day() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, so.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'so_KE'
func (so *so_KE) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, so.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)
	b = append(b, so.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Day() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2c, 0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'so_KE'
func (so *so_KE) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, so.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'so_KE'
func (so *so_KE) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, so.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, so.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'so_KE'
func (so *so_KE) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, so.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, so.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'so_KE'
func (so *so_KE) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, so.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, so.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := so.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
