package guz_KE

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type guz_KE struct {
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

// New returns a new instance of translator for the 'guz_KE' locale
func New() locales.Translator {
	return &guz_KE{
		locale:                 "guz_KE",
		pluralsCardinal:        nil,
		pluralsOrdinal:         nil,
		pluralsRange:           nil,
		timeSeparator:          ":",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyNegativePrefix: "(",
		currencyNegativeSuffix: ")",
		monthsAbbreviated:      []string{"", "Can", "Feb", "Mac", "Apr", "Mei", "Jun", "Cul", "Agt", "Sep", "Okt", "Nob", "Dis"},
		monthsNarrow:           []string{"", "C", "F", "M", "A", "M", "J", "C", "A", "S", "O", "N", "D"},
		monthsWide:             []string{"", "Chanuari", "Feburari", "Machi", "Apiriri", "Mei", "Juni", "Chulai", "Agosti", "Septemba", "Okitoba", "Nobemba", "Disemba"},
		daysAbbreviated:        []string{"Cpr", "Ctt", "Cmn", "Cmt", "Ars", "Icm", "Est"},
		daysNarrow:             []string{"C", "C", "C", "C", "A", "I", "E"},
		daysWide:               []string{"Chumapiri", "Chumatato", "Chumaine", "Chumatano", "Aramisi", "Ichuma", "Esabato"},
		periodsAbbreviated:     []string{"Ma", "Mo"},
		periodsWide:            []string{"Mambia", "Mog"},
		erasAbbreviated:        []string{"YA", "YK"},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"Yeso ataiborwa", "Yeso kaiboirwe"},
		timezones:              map[string]string{"HNPM": "HNPM", "EAT": "EAT", "WIT": "WIT", "ACWST": "ACWST", "HKST": "HKST", "HAT": "HAT", "AKDT": "AKDT", "ACWDT": "ACWDT", "HEEG": "HEEG", "HKT": "HKT", "LHDT": "LHDT", "CDT": "CDT", "HNPMX": "HNPMX", "∅∅∅": "∅∅∅", "HEPM": "HEPM", "ChST": "ChST", "CHAST": "CHAST", "TMT": "TMT", "TMST": "TMST", "CHADT": "CHADT", "AWDT": "AWDT", "MDT": "MDT", "VET": "VET", "UYT": "UYT", "JST": "JST", "ACST": "ACST", "HAST": "HAST", "MYT": "MYT", "EST": "EST", "WITA": "WITA", "SRT": "SRT", "PST": "PST", "WEZ": "WEZ", "WESZ": "WESZ", "HNT": "HNT", "CAT": "CAT", "OESZ": "OESZ", "AWST": "AWST", "WIB": "WIB", "LHST": "LHST", "GMT": "GMT", "PDT": "PDT", "AST": "AST", "HADT": "HADT", "WAST": "WAST", "HNEG": "HNEG", "CLT": "CLT", "CST": "CST", "AEDT": "AEDT", "WAT": "WAT", "COT": "COT", "GYT": "GYT", "HNCU": "HNCU", "SAST": "SAST", "JDT": "JDT", "EDT": "EDT", "AEST": "AEST", "HEOG": "HEOG", "MEZ": "MEZ", "BT": "BT", "ECT": "ECT", "IST": "IST", "HNNOMX": "HNNOMX", "HECU": "HECU", "MST": "MST", "SGT": "SGT", "MESZ": "MESZ", "NZST": "NZST", "NZDT": "NZDT", "HNOG": "HNOG", "ART": "ART", "ARST": "ARST", "HEPMX": "HEPMX", "ADT": "ADT", "GFT": "GFT", "UYST": "UYST", "ACDT": "ACDT", "WART": "WART", "OEZ": "OEZ", "AKST": "AKST", "HENOMX": "HENOMX", "COST": "COST", "BOT": "BOT", "WARST": "WARST", "CLST": "CLST"},
	}
}

// Locale returns the current translators string locale
func (guz *guz_KE) Locale() string {
	return guz.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'guz_KE'
func (guz *guz_KE) PluralsCardinal() []locales.PluralRule {
	return guz.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'guz_KE'
func (guz *guz_KE) PluralsOrdinal() []locales.PluralRule {
	return guz.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'guz_KE'
func (guz *guz_KE) PluralsRange() []locales.PluralRule {
	return guz.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'guz_KE'
func (guz *guz_KE) CardinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'guz_KE'
func (guz *guz_KE) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'guz_KE'
func (guz *guz_KE) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (guz *guz_KE) MonthAbbreviated(month time.Month) string {
	return guz.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (guz *guz_KE) MonthsAbbreviated() []string {
	return guz.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (guz *guz_KE) MonthNarrow(month time.Month) string {
	return guz.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (guz *guz_KE) MonthsNarrow() []string {
	return guz.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (guz *guz_KE) MonthWide(month time.Month) string {
	return guz.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (guz *guz_KE) MonthsWide() []string {
	return guz.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (guz *guz_KE) WeekdayAbbreviated(weekday time.Weekday) string {
	return guz.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (guz *guz_KE) WeekdaysAbbreviated() []string {
	return guz.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (guz *guz_KE) WeekdayNarrow(weekday time.Weekday) string {
	return guz.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (guz *guz_KE) WeekdaysNarrow() []string {
	return guz.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (guz *guz_KE) WeekdayShort(weekday time.Weekday) string {
	return guz.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (guz *guz_KE) WeekdaysShort() []string {
	return guz.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (guz *guz_KE) WeekdayWide(weekday time.Weekday) string {
	return guz.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (guz *guz_KE) WeekdaysWide() []string {
	return guz.daysWide
}

// Decimal returns the decimal point of number
func (guz *guz_KE) Decimal() string {
	return guz.decimal
}

// Group returns the group of number
func (guz *guz_KE) Group() string {
	return guz.group
}

// Group returns the minus sign of number
func (guz *guz_KE) Minus() string {
	return guz.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'guz_KE' and handles both Whole and Real numbers based on 'v'
func (guz *guz_KE) FmtNumber(num float64, v uint64) string {

	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'guz_KE' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (guz *guz_KE) FmtPercent(num float64, v uint64) string {
	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'guz_KE'
func (guz *guz_KE) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := guz.currencies[currency]
	l := len(s) + len(symbol) + 0
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, guz.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, guz.group[0])
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
		b = append(b, guz.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, guz.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'guz_KE'
// in accounting notation.
func (guz *guz_KE) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := guz.currencies[currency]
	l := len(s) + len(symbol) + 2
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, guz.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, guz.group[0])
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

		b = append(b, guz.currencyNegativePrefix[0])

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
			b = append(b, guz.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, guz.currencyNegativeSuffix...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'guz_KE'
func (guz *guz_KE) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'guz_KE'
func (guz *guz_KE) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, guz.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'guz_KE'
func (guz *guz_KE) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, guz.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'guz_KE'
func (guz *guz_KE) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, guz.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, guz.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'guz_KE'
func (guz *guz_KE) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, guz.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'guz_KE'
func (guz *guz_KE) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, guz.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, guz.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'guz_KE'
func (guz *guz_KE) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, guz.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, guz.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'guz_KE'
func (guz *guz_KE) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, guz.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, guz.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := guz.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
