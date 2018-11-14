package luo_KE

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type luo_KE struct {
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

// New returns a new instance of translator for the 'luo_KE' locale
func New() locales.Translator {
	return &luo_KE{
		locale:             "luo_KE",
		pluralsCardinal:    nil,
		pluralsOrdinal:     nil,
		pluralsRange:       nil,
		timeSeparator:      ":",
		currencies:         []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		monthsAbbreviated:  []string{"", "DAC", "DAR", "DAD", "DAN", "DAH", "DAU", "DAO", "DAB", "DOC", "DAP", "DGI", "DAG"},
		monthsNarrow:       []string{"", "C", "R", "D", "N", "B", "U", "B", "B", "C", "P", "C", "P"},
		monthsWide:         []string{"", "Dwe mar Achiel", "Dwe mar Ariyo", "Dwe mar Adek", "Dwe mar Ang’wen", "Dwe mar Abich", "Dwe mar Auchiel", "Dwe mar Abiriyo", "Dwe mar Aboro", "Dwe mar Ochiko", "Dwe mar Apar", "Dwe mar gi achiel", "Dwe mar Apar gi ariyo"},
		daysAbbreviated:    []string{"JMP", "WUT", "TAR", "TAD", "TAN", "TAB", "NGS"},
		daysNarrow:         []string{"J", "W", "T", "T", "T", "T", "N"},
		daysWide:           []string{"Jumapil", "Wuok Tich", "Tich Ariyo", "Tich Adek", "Tich Ang’wen", "Tich Abich", "Ngeso"},
		periodsAbbreviated: []string{"OD", "OT"},
		periodsWide:        []string{"OD", "OT"},
		erasAbbreviated:    []string{"BC", "AD"},
		erasNarrow:         []string{"", ""},
		erasWide:           []string{"Kapok Kristo obiro", "Ka Kristo osebiro"},
		timezones:          map[string]string{"PST": "PST", "AST": "AST", "ADT": "ADT", "NZDT": "NZDT", "SGT": "SGT", "TMST": "TMST", "GYT": "GYT", "WEZ": "WEZ", "ECT": "ECT", "LHST": "LHST", "HEPM": "HEPM", "WITA": "WITA", "HENOMX": "HENOMX", "TMT": "TMT", "AWST": "AWST", "CLST": "CLST", "HKT": "HKT", "IST": "IST", "ACWDT": "ACWDT", "HNCU": "HNCU", "MDT": "MDT", "WART": "WART", "OESZ": "OESZ", "HADT": "HADT", "GMT": "GMT", "HEPMX": "HEPMX", "WIB": "WIB", "ACDT": "ACDT", "HNPMX": "HNPMX", "∅∅∅": "∅∅∅", "PDT": "PDT", "AWDT": "AWDT", "MYT": "MYT", "ACWST": "ACWST", "CHAST": "CHAST", "HEOG": "HEOG", "EDT": "EDT", "ACST": "ACST", "MESZ": "MESZ", "VET": "VET", "SRT": "SRT", "CLT": "CLT", "ChST": "ChST", "ARST": "ARST", "HNPM": "HNPM", "AKDT": "AKDT", "AEST": "AEST", "JDT": "JDT", "BT": "BT", "BOT": "BOT", "HEEG": "HEEG", "CDT": "CDT", "GFT": "GFT", "CAT": "CAT", "OEZ": "OEZ", "COT": "COT", "MST": "MST", "HNT": "HNT", "SAST": "SAST", "LHDT": "LHDT", "HNOG": "HNOG", "MEZ": "MEZ", "WARST": "WARST", "HAT": "HAT", "EAT": "EAT", "WIT": "WIT", "WAT": "WAT", "CST": "CST", "WESZ": "WESZ", "HKST": "HKST", "ART": "ART", "COST": "COST", "HECU": "HECU", "UYST": "UYST", "CHADT": "CHADT", "AEDT": "AEDT", "JST": "JST", "EST": "EST", "HNEG": "HNEG", "HAST": "HAST", "UYT": "UYT", "NZST": "NZST", "AKST": "AKST", "HNNOMX": "HNNOMX", "WAST": "WAST"},
	}
}

// Locale returns the current translators string locale
func (luo *luo_KE) Locale() string {
	return luo.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'luo_KE'
func (luo *luo_KE) PluralsCardinal() []locales.PluralRule {
	return luo.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'luo_KE'
func (luo *luo_KE) PluralsOrdinal() []locales.PluralRule {
	return luo.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'luo_KE'
func (luo *luo_KE) PluralsRange() []locales.PluralRule {
	return luo.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'luo_KE'
func (luo *luo_KE) CardinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'luo_KE'
func (luo *luo_KE) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'luo_KE'
func (luo *luo_KE) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (luo *luo_KE) MonthAbbreviated(month time.Month) string {
	return luo.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (luo *luo_KE) MonthsAbbreviated() []string {
	return luo.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (luo *luo_KE) MonthNarrow(month time.Month) string {
	return luo.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (luo *luo_KE) MonthsNarrow() []string {
	return luo.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (luo *luo_KE) MonthWide(month time.Month) string {
	return luo.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (luo *luo_KE) MonthsWide() []string {
	return luo.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (luo *luo_KE) WeekdayAbbreviated(weekday time.Weekday) string {
	return luo.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (luo *luo_KE) WeekdaysAbbreviated() []string {
	return luo.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (luo *luo_KE) WeekdayNarrow(weekday time.Weekday) string {
	return luo.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (luo *luo_KE) WeekdaysNarrow() []string {
	return luo.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (luo *luo_KE) WeekdayShort(weekday time.Weekday) string {
	return luo.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (luo *luo_KE) WeekdaysShort() []string {
	return luo.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (luo *luo_KE) WeekdayWide(weekday time.Weekday) string {
	return luo.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (luo *luo_KE) WeekdaysWide() []string {
	return luo.daysWide
}

// Decimal returns the decimal point of number
func (luo *luo_KE) Decimal() string {
	return luo.decimal
}

// Group returns the group of number
func (luo *luo_KE) Group() string {
	return luo.group
}

// Group returns the minus sign of number
func (luo *luo_KE) Minus() string {
	return luo.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'luo_KE' and handles both Whole and Real numbers based on 'v'
func (luo *luo_KE) FmtNumber(num float64, v uint64) string {

	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'luo_KE' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (luo *luo_KE) FmtPercent(num float64, v uint64) string {
	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'luo_KE'
func (luo *luo_KE) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := luo.currencies[currency]
	l := len(s) + len(symbol) + 0
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, luo.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, luo.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, luo.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, luo.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	b = append(b, symbol...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'luo_KE'
// in accounting notation.
func (luo *luo_KE) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := luo.currencies[currency]
	l := len(s) + len(symbol) + 0
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, luo.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, luo.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {

		b = append(b, luo.minus[0])

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, luo.decimal...)
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

// FmtDateShort returns the short date representation of 't' for 'luo_KE'
func (luo *luo_KE) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'luo_KE'
func (luo *luo_KE) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, luo.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'luo_KE'
func (luo *luo_KE) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, luo.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'luo_KE'
func (luo *luo_KE) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, luo.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, luo.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'luo_KE'
func (luo *luo_KE) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, luo.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'luo_KE'
func (luo *luo_KE) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, luo.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, luo.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'luo_KE'
func (luo *luo_KE) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, luo.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, luo.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'luo_KE'
func (luo *luo_KE) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, luo.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, luo.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := luo.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
