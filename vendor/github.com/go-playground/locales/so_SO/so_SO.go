package so_SO

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type so_SO struct {
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

// New returns a new instance of translator for the 'so_SO' locale
func New() locales.Translator {
	return &so_SO{
		locale:             "so_SO",
		pluralsCardinal:    []locales.PluralRule{2, 6},
		pluralsOrdinal:     nil,
		pluralsRange:       nil,
		decimal:            ".",
		group:              ",",
		timeSeparator:      ":",
		currencies:         []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
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
		timezones:          map[string]string{"HENOMX": "HENOMX", "GYT": "GYT", "AST": "AST", "LHDT": "LHDT", "CLST": "CLST", "OESZ": "OESZ", "AKDT": "AKDT", "MEZ": "MEZ", "MESZ": "MESZ", "CAT": "CAT", "ARST": "ARST", "ChST": "ChST", "MST": "MST", "HKST": "HKST", "TMST": "TMST", "ECT": "ECT", "WITA": "WITA", "VET": "VET", "OEZ": "OEZ", "ART": "ART", "HEPMX": "HEPMX", "AWST": "AWST", "AEDT": "AEDT", "EST": "EST", "ACWDT": "ACWDT", "TMT": "TMT", "HNCU": "HNCU", "SGT": "SGT", "JDT": "JDT", "HEOG": "HEOG", "ACWST": "ACWST", "HEPM": "HEPM", "COST": "Waqtiyada Xagaaga Kolambiya", "WIB": "WIB", "∅∅∅": "∅∅∅", "HNNOMX": "HNNOMX", "CLT": "CLT", "HAST": "HAST", "PDT": "PDT", "MDT": "MDT", "WART": "WART", "UYST": "UYST", "HECU": "HECU", "HNPMX": "HNPMX", "JST": "JST", "NZST": "NZST", "NZDT": "NZDT", "ACDT": "ACDT", "WIT": "WIT", "CST": "CST", "WAST": "WAST", "MYT": "MYT", "HNOG": "HNOG", "EDT": "EDT", "ACST": "ACST", "LHST": "LHST", "WARST": "WARST", "HAT": "HAT", "BT": "BT", "CHAST": "CHAST", "AEST": "AEST", "SAST": "SAST", "WESZ": "WESZ", "BOT": "BOT", "AKST": "AKST", "HNEG": "HNEG", "ADT": "ADT", "PST": "PST", "HKT": "HKT", "IST": "IST", "SRT": "SRT", "COT": "Waqtiyada Caadiga ah ee kolambiya", "CHADT": "CHADT", "AWDT": "AWDT", "HEEG": "HEEG", "GFT": "GFT", "HNT": "HNT", "HADT": "HADT", "GMT": "GMT", "UYT": "UYT", "CDT": "CDT", "WAT": "WAT", "HNPM": "HNPM", "EAT": "EAT", "WEZ": "WEZ"},
	}
}

// Locale returns the current translators string locale
func (so *so_SO) Locale() string {
	return so.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'so_SO'
func (so *so_SO) PluralsCardinal() []locales.PluralRule {
	return so.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'so_SO'
func (so *so_SO) PluralsOrdinal() []locales.PluralRule {
	return so.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'so_SO'
func (so *so_SO) PluralsRange() []locales.PluralRule {
	return so.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'so_SO'
func (so *so_SO) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'so_SO'
func (so *so_SO) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'so_SO'
func (so *so_SO) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (so *so_SO) MonthAbbreviated(month time.Month) string {
	return so.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (so *so_SO) MonthsAbbreviated() []string {
	return so.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (so *so_SO) MonthNarrow(month time.Month) string {
	return so.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (so *so_SO) MonthsNarrow() []string {
	return so.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (so *so_SO) MonthWide(month time.Month) string {
	return so.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (so *so_SO) MonthsWide() []string {
	return so.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (so *so_SO) WeekdayAbbreviated(weekday time.Weekday) string {
	return so.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (so *so_SO) WeekdaysAbbreviated() []string {
	return so.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (so *so_SO) WeekdayNarrow(weekday time.Weekday) string {
	return so.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (so *so_SO) WeekdaysNarrow() []string {
	return so.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (so *so_SO) WeekdayShort(weekday time.Weekday) string {
	return so.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (so *so_SO) WeekdaysShort() []string {
	return so.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (so *so_SO) WeekdayWide(weekday time.Weekday) string {
	return so.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (so *so_SO) WeekdaysWide() []string {
	return so.daysWide
}

// Decimal returns the decimal point of number
func (so *so_SO) Decimal() string {
	return so.decimal
}

// Group returns the group of number
func (so *so_SO) Group() string {
	return so.group
}

// Group returns the minus sign of number
func (so *so_SO) Minus() string {
	return so.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'so_SO' and handles both Whole and Real numbers based on 'v'
func (so *so_SO) FmtNumber(num float64, v uint64) string {

	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'so_SO' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (so *so_SO) FmtPercent(num float64, v uint64) string {
	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'so_SO'
func (so *so_SO) FmtCurrency(num float64, v uint64, currency currency.Type) string {

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

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'so_SO'
// in accounting notation.
func (so *so_SO) FmtAccounting(num float64, v uint64, currency currency.Type) string {

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

// FmtDateShort returns the short date representation of 't' for 'so_SO'
func (so *so_SO) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'so_SO'
func (so *so_SO) FmtDateMedium(t time.Time) string {

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

// FmtDateLong returns the long date representation of 't' for 'so_SO'
func (so *so_SO) FmtDateLong(t time.Time) string {

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

// FmtDateFull returns the full date representation of 't' for 'so_SO'
func (so *so_SO) FmtDateFull(t time.Time) string {

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

// FmtTimeShort returns the short time representation of 't' for 'so_SO'
func (so *so_SO) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, so.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, so.periodsAbbreviated[0]...)
	} else {
		b = append(b, so.periodsAbbreviated[1]...)
	}

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'so_SO'
func (so *so_SO) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
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

	if t.Hour() < 12 {
		b = append(b, so.periodsAbbreviated[0]...)
	} else {
		b = append(b, so.periodsAbbreviated[1]...)
	}

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'so_SO'
func (so *so_SO) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
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

	if t.Hour() < 12 {
		b = append(b, so.periodsAbbreviated[0]...)
	} else {
		b = append(b, so.periodsAbbreviated[1]...)
	}

	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'so_SO'
func (so *so_SO) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
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

	if t.Hour() < 12 {
		b = append(b, so.periodsAbbreviated[0]...)
	} else {
		b = append(b, so.periodsAbbreviated[1]...)
	}

	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := so.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
