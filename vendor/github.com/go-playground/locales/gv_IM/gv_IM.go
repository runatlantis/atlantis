package gv_IM

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type gv_IM struct {
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

// New returns a new instance of translator for the 'gv_IM' locale
func New() locales.Translator {
	return &gv_IM{
		locale:             "gv_IM",
		pluralsCardinal:    []locales.PluralRule{2, 3, 4, 5, 6},
		pluralsOrdinal:     nil,
		pluralsRange:       nil,
		timeSeparator:      ":",
		currencies:         []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		monthsAbbreviated:  []string{"", "J-guer", "T-arree", "Mayrnt", "Avrril", "Boaldyn", "M-souree", "J-souree", "Luanistyn", "M-fouyir", "J-fouyir", "M-Houney", "M-Nollick"},
		monthsWide:         []string{"", "Jerrey-geuree", "Toshiaght-arree", "Mayrnt", "Averil", "Boaldyn", "Mean-souree", "Jerrey-souree", "Luanistyn", "Mean-fouyir", "Jerrey-fouyir", "Mee Houney", "Mee ny Nollick"},
		daysAbbreviated:    []string{"Jed", "Jel", "Jem", "Jerc", "Jerd", "Jeh", "Jes"},
		daysWide:           []string{"Jedoonee", "Jelhein", "Jemayrt", "Jercean", "Jerdein", "Jeheiney", "Jesarn"},
		periodsAbbreviated: []string{"a.m.", "p.m."},
		periodsWide:        []string{"a.m.", "p.m."},
		erasAbbreviated:    []string{"RC", "AD"},
		erasNarrow:         []string{"", ""},
		erasWide:           []string{"", ""},
		timezones:          map[string]string{"MST": "MST", "WIB": "WIB", "SGT": "SGT", "COT": "COT", "WEZ": "WEZ", "BT": "BT", "ACWDT": "ACWDT", "WART": "WART", "HAT": "HAT", "GFT": "GFT", "ECT": "ECT", "CHADT": "CHADT", "ADT": "ADT", "MDT": "MDT", "WAST": "WAST", "HEOG": "HEOG", "TMST": "TMST", "GYT": "GYT", "HNOG": "HNOG", "ART": "ART", "UYT": "UYT", "AWST": "AWST", "∅∅∅": "∅∅∅", "JST": "JST", "ACWST": "ACWST", "MESZ": "MESZ", "VET": "VET", "HNCU": "HNCU", "HNEG": "HNEG", "MEZ": "MEZ", "HKT": "HKT", "WARST": "WARST", "HNNOMX": "HNNOMX", "HEPM": "HEPM", "CLST": "CLST", "AEDT": "AEDT", "SAST": "SAST", "WESZ": "WESZ", "EDT": "EDT", "HKST": "HKST", "LHDT": "LHDT", "WAT": "WAT", "JDT": "JDT", "LHST": "LHST", "CAT": "CAT", "CST": "CST", "AWDT": "AWDT", "AST": "AST", "NZST": "NZST", "HNT": "HNT", "HADT": "HADT", "HNPMX": "HNPMX", "HNPM": "HNPM", "COST": "COST", "OEZ": "OEZ", "OESZ": "OESZ", "HEPMX": "HEPMX", "BOT": "BOT", "AKDT": "AKDT", "ACDT": "ACDT", "SRT": "SRT", "EAT": "EAT", "HECU": "HECU", "ChST": "ChST", "MYT": "MYT", "EST": "EST", "IST": "IST", "WITA": "WITA", "WIT": "WIT", "ARST": "ARST", "AEST": "AEST", "ACST": "ACST", "CLT": "CLT", "UYST": "UYST", "CHAST": "CHAST", "HAST": "HAST", "GMT": "GMT", "CDT": "CDT", "PST": "PST", "PDT": "PDT", "AKST": "AKST", "HENOMX": "HENOMX", "TMT": "TMT", "NZDT": "NZDT", "HEEG": "HEEG"},
	}
}

// Locale returns the current translators string locale
func (gv *gv_IM) Locale() string {
	return gv.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'gv_IM'
func (gv *gv_IM) PluralsCardinal() []locales.PluralRule {
	return gv.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'gv_IM'
func (gv *gv_IM) PluralsOrdinal() []locales.PluralRule {
	return gv.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'gv_IM'
func (gv *gv_IM) PluralsRange() []locales.PluralRule {
	return gv.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'gv_IM'
func (gv *gv_IM) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)
	i := int64(n)
	iMod100 := i % 100
	iMod10 := i % 10

	if v == 0 && iMod10 == 1 {
		return locales.PluralRuleOne
	} else if v == 0 && iMod10 == 2 {
		return locales.PluralRuleTwo
	} else if v == 0 && (iMod100 == 0 || iMod100 == 20 || iMod100 == 40 || iMod100 == 60 || iMod100 == 80) {
		return locales.PluralRuleFew
	} else if v != 0 {
		return locales.PluralRuleMany
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'gv_IM'
func (gv *gv_IM) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'gv_IM'
func (gv *gv_IM) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (gv *gv_IM) MonthAbbreviated(month time.Month) string {
	return gv.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (gv *gv_IM) MonthsAbbreviated() []string {
	return gv.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (gv *gv_IM) MonthNarrow(month time.Month) string {
	return gv.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (gv *gv_IM) MonthsNarrow() []string {
	return nil
}

// MonthWide returns the locales wide month given the 'month' provided
func (gv *gv_IM) MonthWide(month time.Month) string {
	return gv.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (gv *gv_IM) MonthsWide() []string {
	return gv.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (gv *gv_IM) WeekdayAbbreviated(weekday time.Weekday) string {
	return gv.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (gv *gv_IM) WeekdaysAbbreviated() []string {
	return gv.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (gv *gv_IM) WeekdayNarrow(weekday time.Weekday) string {
	return gv.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (gv *gv_IM) WeekdaysNarrow() []string {
	return gv.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (gv *gv_IM) WeekdayShort(weekday time.Weekday) string {
	return gv.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (gv *gv_IM) WeekdaysShort() []string {
	return gv.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (gv *gv_IM) WeekdayWide(weekday time.Weekday) string {
	return gv.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (gv *gv_IM) WeekdaysWide() []string {
	return gv.daysWide
}

// Decimal returns the decimal point of number
func (gv *gv_IM) Decimal() string {
	return gv.decimal
}

// Group returns the group of number
func (gv *gv_IM) Group() string {
	return gv.group
}

// Group returns the minus sign of number
func (gv *gv_IM) Minus() string {
	return gv.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'gv_IM' and handles both Whole and Real numbers based on 'v'
func (gv *gv_IM) FmtNumber(num float64, v uint64) string {

	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'gv_IM' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (gv *gv_IM) FmtPercent(num float64, v uint64) string {
	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'gv_IM'
func (gv *gv_IM) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := gv.currencies[currency]
	l := len(s) + len(symbol) + 0
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, gv.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, gv.group[0])
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
		b = append(b, gv.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, gv.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'gv_IM'
// in accounting notation.
func (gv *gv_IM) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := gv.currencies[currency]
	l := len(s) + len(symbol) + 0
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, gv.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, gv.group[0])
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

		b = append(b, gv.minus[0])

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
			b = append(b, gv.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'gv_IM'
func (gv *gv_IM) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'gv_IM'
func (gv *gv_IM) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, gv.monthsAbbreviated[t.Month()]...)
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

// FmtDateLong returns the long date representation of 't' for 'gv_IM'
func (gv *gv_IM) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Day() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, gv.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'gv_IM'
func (gv *gv_IM) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, gv.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x20}...)

	if t.Day() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, gv.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'gv_IM'
func (gv *gv_IM) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, gv.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'gv_IM'
func (gv *gv_IM) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, gv.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, gv.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'gv_IM'
func (gv *gv_IM) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, gv.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, gv.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'gv_IM'
func (gv *gv_IM) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, gv.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, gv.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := gv.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
