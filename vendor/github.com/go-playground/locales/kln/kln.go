package kln

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type kln struct {
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

// New returns a new instance of translator for the 'kln' locale
func New() locales.Translator {
	return &kln{
		locale:                 "kln",
		pluralsCardinal:        nil,
		pluralsOrdinal:         nil,
		pluralsRange:           nil,
		timeSeparator:          ":",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "Ksh", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyNegativePrefix: "(",
		currencyNegativeSuffix: ")",
		monthsAbbreviated:      []string{"", "Mul", "Ngat", "Taa", "Iwo", "Mam", "Paa", "Nge", "Roo", "Bur", "Epe", "Kpt", "Kpa"},
		monthsNarrow:           []string{"", "M", "N", "T", "I", "M", "P", "N", "R", "B", "E", "K", "K"},
		monthsWide:             []string{"", "Mulgul", "Ng’atyaato", "Kiptaamo", "Iwootkuut", "Mamuut", "Paagi", "Ng’eiyeet", "Rooptui", "Bureet", "Epeeso", "Kipsuunde ne taai", "Kipsuunde nebo aeng’"},
		daysAbbreviated:        []string{"Kts", "Kot", "Koo", "Kos", "Koa", "Kom", "Kol"},
		daysNarrow:             []string{"T", "T", "O", "S", "A", "M", "L"},
		daysWide:               []string{"Kotisap", "Kotaai", "Koaeng’", "Kosomok", "Koang’wan", "Komuut", "Kolo"},
		periodsAbbreviated:     []string{"krn", "koosk"},
		periodsWide:            []string{"karoon", "kooskoliny"},
		erasAbbreviated:        []string{"AM", "KO"},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"Amait kesich Jesu", "Kokakesich Jesu"},
		timezones:              map[string]string{"EST": "EST", "WART": "WART", "HNPM": "HNPM", "VET": "VET", "CAT": "CAT", "PDT": "PDT", "AEDT": "AEDT", "BT": "BT", "ACDT": "ACDT", "HKST": "HKST", "LHST": "LHST", "HEPMX": "HEPMX", "MST": "MST", "GFT": "GFT", "CLST": "CLST", "ChST": "ChST", "PST": "PST", "AKDT": "AKDT", "HNNOMX": "HNNOMX", "GYT": "GYT", "UYST": "UYST", "ECT": "ECT", "HNEG": "HNEG", "WITA": "WITA", "TMST": "TMST", "GMT": "GMT", "SRT": "SRT", "WIT": "WIT", "ARST": "ARST", "MDT": "MDT", "WAST": "WAST", "ACST": "ACST", "HKT": "HKT", "LHDT": "LHDT", "HNCU": "HNCU", "HNPMX": "HNPMX", "∅∅∅": "∅∅∅", "SGT": "SGT", "IST": "IST", "TMT": "TMT", "UYT": "UYT", "AEST": "AEST", "WAT": "WAT", "MYT": "MYT", "ACWDT": "ACWDT", "CHAST": "CHAST", "MEZ": "MEZ", "COT": "COT", "AWDT": "AWDT", "JDT": "JDT", "MESZ": "MESZ", "EAT": "EAT", "WESZ": "WESZ", "AKST": "AKST", "HEOG": "HEOG", "WARST": "WARST", "HNT": "HNT", "CLT": "CLT", "CHADT": "CHADT", "AST": "AST", "SAST": "SAST", "NZST": "NZST", "HAT": "HAT", "HENOMX": "HENOMX", "HEPM": "HEPM", "OEZ": "OEZ", "OESZ": "OESZ", "ADT": "ADT", "BOT": "BOT", "HADT": "HADT", "CST": "CST", "WEZ": "WEZ", "JST": "JST", "ACWST": "ACWST", "HNOG": "HNOG", "COST": "COST", "ART": "ART", "HECU": "HECU", "CDT": "CDT", "WIB": "WIB", "NZDT": "NZDT", "EDT": "EDT", "HEEG": "HEEG", "HAST": "HAST", "AWST": "AWST"},
	}
}

// Locale returns the current translators string locale
func (kln *kln) Locale() string {
	return kln.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'kln'
func (kln *kln) PluralsCardinal() []locales.PluralRule {
	return kln.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'kln'
func (kln *kln) PluralsOrdinal() []locales.PluralRule {
	return kln.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'kln'
func (kln *kln) PluralsRange() []locales.PluralRule {
	return kln.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'kln'
func (kln *kln) CardinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'kln'
func (kln *kln) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'kln'
func (kln *kln) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (kln *kln) MonthAbbreviated(month time.Month) string {
	return kln.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (kln *kln) MonthsAbbreviated() []string {
	return kln.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (kln *kln) MonthNarrow(month time.Month) string {
	return kln.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (kln *kln) MonthsNarrow() []string {
	return kln.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (kln *kln) MonthWide(month time.Month) string {
	return kln.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (kln *kln) MonthsWide() []string {
	return kln.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (kln *kln) WeekdayAbbreviated(weekday time.Weekday) string {
	return kln.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (kln *kln) WeekdaysAbbreviated() []string {
	return kln.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (kln *kln) WeekdayNarrow(weekday time.Weekday) string {
	return kln.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (kln *kln) WeekdaysNarrow() []string {
	return kln.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (kln *kln) WeekdayShort(weekday time.Weekday) string {
	return kln.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (kln *kln) WeekdaysShort() []string {
	return kln.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (kln *kln) WeekdayWide(weekday time.Weekday) string {
	return kln.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (kln *kln) WeekdaysWide() []string {
	return kln.daysWide
}

// Decimal returns the decimal point of number
func (kln *kln) Decimal() string {
	return kln.decimal
}

// Group returns the group of number
func (kln *kln) Group() string {
	return kln.group
}

// Group returns the minus sign of number
func (kln *kln) Minus() string {
	return kln.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'kln' and handles both Whole and Real numbers based on 'v'
func (kln *kln) FmtNumber(num float64, v uint64) string {

	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'kln' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (kln *kln) FmtPercent(num float64, v uint64) string {
	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'kln'
func (kln *kln) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := kln.currencies[currency]
	l := len(s) + len(symbol) + 0
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, kln.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, kln.group[0])
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
		b = append(b, kln.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, kln.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'kln'
// in accounting notation.
func (kln *kln) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := kln.currencies[currency]
	l := len(s) + len(symbol) + 2
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, kln.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, kln.group[0])
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

		b = append(b, kln.currencyNegativePrefix[0])

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
			b = append(b, kln.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, kln.currencyNegativeSuffix...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'kln'
func (kln *kln) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'kln'
func (kln *kln) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, kln.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'kln'
func (kln *kln) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, kln.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'kln'
func (kln *kln) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, kln.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, kln.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'kln'
func (kln *kln) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, kln.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'kln'
func (kln *kln) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, kln.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, kln.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'kln'
func (kln *kln) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, kln.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, kln.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'kln'
func (kln *kln) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, kln.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, kln.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := kln.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
