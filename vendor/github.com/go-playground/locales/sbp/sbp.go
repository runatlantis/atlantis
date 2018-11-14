package sbp

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type sbp struct {
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

// New returns a new instance of translator for the 'sbp' locale
func New() locales.Translator {
	return &sbp{
		locale:             "sbp",
		pluralsCardinal:    nil,
		pluralsOrdinal:     nil,
		pluralsRange:       nil,
		decimal:            ".",
		group:              ",",
		timeSeparator:      ":",
		currencies:         []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TSh", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		monthsAbbreviated:  []string{"", "Mup", "Mwi", "Msh", "Mun", "Mag", "Muj", "Msp", "Mpg", "Mye", "Mok", "Mus", "Muh"},
		monthsWide:         []string{"", "Mupalangulwa", "Mwitope", "Mushende", "Munyi", "Mushende Magali", "Mujimbi", "Mushipepo", "Mupuguto", "Munyense", "Mokhu", "Musongandembwe", "Muhaano"},
		daysAbbreviated:    []string{"Mul", "Jtt", "Jnn", "Jtn", "Alh", "Iju", "Jmo"},
		daysNarrow:         []string{"M", "J", "J", "J", "A", "I", "J"},
		daysWide:           []string{"Mulungu", "Jumatatu", "Jumanne", "Jumatano", "Alahamisi", "Ijumaa", "Jumamosi"},
		periodsAbbreviated: []string{"Lwamilawu", "Pashamihe"},
		periodsWide:        []string{"Lwamilawu", "Pashamihe"},
		erasAbbreviated:    []string{"AK", "PK"},
		erasNarrow:         []string{"", ""},
		erasWide:           []string{"Ashanali uKilisito", "Pamwandi ya Kilisto"},
		timezones:          map[string]string{"CST": "CST", "PST": "PST", "JDT": "JDT", "CLST": "CLST", "CHAST": "CHAST", "AWST": "AWST", "AST": "AST", "BOT": "BOT", "EST": "EST", "ACWST": "ACWST", "HEPM": "HEPM", "EAT": "EAT", "ART": "ART", "ACWDT": "ACWDT", "HNCU": "HNCU", "HEPMX": "HEPMX", "JST": "JST", "ACST": "ACST", "GMT": "GMT", "HNPMX": "HNPMX", "MDT": "MDT", "WARST": "WARST", "GYT": "GYT", "AWDT": "AWDT", "SGT": "SGT", "EDT": "EDT", "HNOG": "HNOG", "VET": "VET", "TMST": "TMST", "HADT": "HADT", "MST": "MST", "ACDT": "ACDT", "WAT": "WAT", "WEZ": "WEZ", "BT": "BT", "NZDT": "NZDT", "WART": "WART", "HENOMX": "HENOMX", "CAT": "CAT", "OESZ": "OESZ", "WAST": "WAST", "WESZ": "WESZ", "NZST": "NZST", "OEZ": "OEZ", "UYT": "UYT", "AKDT": "AKDT", "MEZ": "MEZ", "LHDT": "LHDT", "HAT": "HAT", "COT": "COT", "WIT": "WIT", "UYST": "UYST", "ADT": "ADT", "AKST": "AKST", "HNEG": "HNEG", "HEEG": "HEEG", "MESZ": "MESZ", "LHST": "LHST", "HNT": "HNT", "TMT": "TMT", "CHADT": "CHADT", "CDT": "CDT", "AEST": "AEST", "SAST": "SAST", "∅∅∅": "∅∅∅", "GFT": "GFT", "HKT": "HKT", "HKST": "HKST", "HNPM": "HNPM", "HAST": "HAST", "ARST": "ARST", "COST": "COST", "HECU": "HECU", "AEDT": "AEDT", "SRT": "SRT", "CLT": "CLT", "ChST": "ChST", "PDT": "PDT", "ECT": "ECT", "HEOG": "HEOG", "IST": "IST", "WIB": "WIB", "MYT": "MYT", "WITA": "WITA", "HNNOMX": "HNNOMX"},
	}
}

// Locale returns the current translators string locale
func (sbp *sbp) Locale() string {
	return sbp.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'sbp'
func (sbp *sbp) PluralsCardinal() []locales.PluralRule {
	return sbp.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'sbp'
func (sbp *sbp) PluralsOrdinal() []locales.PluralRule {
	return sbp.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'sbp'
func (sbp *sbp) PluralsRange() []locales.PluralRule {
	return sbp.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'sbp'
func (sbp *sbp) CardinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'sbp'
func (sbp *sbp) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'sbp'
func (sbp *sbp) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (sbp *sbp) MonthAbbreviated(month time.Month) string {
	return sbp.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (sbp *sbp) MonthsAbbreviated() []string {
	return sbp.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (sbp *sbp) MonthNarrow(month time.Month) string {
	return sbp.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (sbp *sbp) MonthsNarrow() []string {
	return nil
}

// MonthWide returns the locales wide month given the 'month' provided
func (sbp *sbp) MonthWide(month time.Month) string {
	return sbp.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (sbp *sbp) MonthsWide() []string {
	return sbp.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (sbp *sbp) WeekdayAbbreviated(weekday time.Weekday) string {
	return sbp.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (sbp *sbp) WeekdaysAbbreviated() []string {
	return sbp.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (sbp *sbp) WeekdayNarrow(weekday time.Weekday) string {
	return sbp.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (sbp *sbp) WeekdaysNarrow() []string {
	return sbp.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (sbp *sbp) WeekdayShort(weekday time.Weekday) string {
	return sbp.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (sbp *sbp) WeekdaysShort() []string {
	return sbp.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (sbp *sbp) WeekdayWide(weekday time.Weekday) string {
	return sbp.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (sbp *sbp) WeekdaysWide() []string {
	return sbp.daysWide
}

// Decimal returns the decimal point of number
func (sbp *sbp) Decimal() string {
	return sbp.decimal
}

// Group returns the group of number
func (sbp *sbp) Group() string {
	return sbp.group
}

// Group returns the minus sign of number
func (sbp *sbp) Minus() string {
	return sbp.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'sbp' and handles both Whole and Real numbers based on 'v'
func (sbp *sbp) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 1 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, sbp.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, sbp.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, sbp.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'sbp' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (sbp *sbp) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 1
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, sbp.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, sbp.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, sbp.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'sbp'
func (sbp *sbp) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := sbp.currencies[currency]
	l := len(s) + len(symbol) + 1 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, sbp.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, sbp.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, sbp.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, sbp.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	b = append(b, symbol...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'sbp'
// in accounting notation.
func (sbp *sbp) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := sbp.currencies[currency]
	l := len(s) + len(symbol) + 1 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, sbp.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, sbp.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {

		b = append(b, sbp.minus[0])

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, sbp.decimal...)
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

// FmtDateShort returns the short date representation of 't' for 'sbp'
func (sbp *sbp) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'sbp'
func (sbp *sbp) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, sbp.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'sbp'
func (sbp *sbp) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, sbp.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'sbp'
func (sbp *sbp) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, sbp.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, sbp.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'sbp'
func (sbp *sbp) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, sbp.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'sbp'
func (sbp *sbp) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, sbp.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, sbp.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'sbp'
func (sbp *sbp) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, sbp.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, sbp.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'sbp'
func (sbp *sbp) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, sbp.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, sbp.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := sbp.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
