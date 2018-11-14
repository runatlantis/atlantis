package khq_ML

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type khq_ML struct {
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

// New returns a new instance of translator for the 'khq_ML' locale
func New() locales.Translator {
	return &khq_ML{
		locale:             "khq_ML",
		pluralsCardinal:    nil,
		pluralsOrdinal:     nil,
		pluralsRange:       nil,
		group:              " ",
		timeSeparator:      ":",
		currencies:         []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		monthsAbbreviated:  []string{"", "Žan", "Fee", "Mar", "Awi", "Me", "Žuw", "Žuy", "Ut", "Sek", "Okt", "Noo", "Dee"},
		monthsNarrow:       []string{"", "Ž", "F", "M", "A", "M", "Ž", "Ž", "U", "S", "O", "N", "D"},
		monthsWide:         []string{"", "Žanwiye", "Feewiriye", "Marsi", "Awiril", "Me", "Žuweŋ", "Žuyye", "Ut", "Sektanbur", "Oktoobur", "Noowanbur", "Deesanbur"},
		daysAbbreviated:    []string{"Alh", "Ati", "Ata", "Ala", "Alm", "Alj", "Ass"},
		daysNarrow:         []string{"H", "T", "T", "L", "L", "L", "S"},
		daysWide:           []string{"Alhadi", "Atini", "Atalata", "Alarba", "Alhamiisa", "Aljuma", "Assabdu"},
		periodsAbbreviated: []string{"Adduha", "Aluula"},
		periodsWide:        []string{"Adduha", "Aluula"},
		erasAbbreviated:    []string{"IJ", "IZ"},
		erasNarrow:         []string{"", ""},
		erasWide:           []string{"Isaa jine", "Isaa jamanoo"},
		timezones:          map[string]string{"HEPM": "HEPM", "UYT": "UYT", "NZDT": "NZDT", "EDT": "EDT", "MEZ": "MEZ", "LHST": "LHST", "HECU": "HECU", "AKDT": "AKDT", "ARST": "ARST", "HKST": "HKST", "HAT": "HAT", "MST": "MST", "MDT": "MDT", "AEDT": "AEDT", "SGT": "SGT", "∅∅∅": "∅∅∅", "HNPM": "HNPM", "ART": "ART", "AWDT": "AWDT", "AEST": "AEST", "WESZ": "WESZ", "WIB": "WIB", "JST": "JST", "MESZ": "MESZ", "HKT": "HKT", "ECT": "ECT", "ACWST": "ACWST", "JDT": "JDT", "VET": "VET", "HENOMX": "HENOMX", "HNPMX": "HNPMX", "SAST": "SAST", "GFT": "GFT", "HNT": "HNT", "OEZ": "OEZ", "HEEG": "HEEG", "HNOG": "HNOG", "SRT": "SRT", "WAT": "WAT", "CDT": "CDT", "HNNOMX": "HNNOMX", "COST": "COST", "WIT": "WIT", "CAT": "CAT", "UYST": "UYST", "ChST": "ChST", "HNCU": "HNCU", "ADT": "ADT", "IST": "IST", "WITA": "WITA", "HEOG": "HEOG", "MYT": "MYT", "HNEG": "HNEG", "TMT": "TMT", "PDT": "PDT", "CHAST": "CHAST", "CLST": "CLST", "GMT": "GMT", "CLT": "CLT", "BT": "BT", "ACST": "ACST", "WART": "WART", "LHDT": "LHDT", "GYT": "GYT", "CST": "CST", "PST": "PST", "WEZ": "WEZ", "AKST": "AKST", "WARST": "WARST", "TMST": "TMST", "HEPMX": "HEPMX", "AST": "AST", "WAST": "WAST", "HADT": "HADT", "AWST": "AWST", "OESZ": "OESZ", "HAST": "HAST", "CHADT": "CHADT", "BOT": "BOT", "NZST": "NZST", "ACWDT": "ACWDT", "EAT": "EAT", "COT": "COT", "EST": "EST", "ACDT": "ACDT"},
	}
}

// Locale returns the current translators string locale
func (khq *khq_ML) Locale() string {
	return khq.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'khq_ML'
func (khq *khq_ML) PluralsCardinal() []locales.PluralRule {
	return khq.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'khq_ML'
func (khq *khq_ML) PluralsOrdinal() []locales.PluralRule {
	return khq.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'khq_ML'
func (khq *khq_ML) PluralsRange() []locales.PluralRule {
	return khq.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'khq_ML'
func (khq *khq_ML) CardinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'khq_ML'
func (khq *khq_ML) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'khq_ML'
func (khq *khq_ML) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (khq *khq_ML) MonthAbbreviated(month time.Month) string {
	return khq.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (khq *khq_ML) MonthsAbbreviated() []string {
	return khq.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (khq *khq_ML) MonthNarrow(month time.Month) string {
	return khq.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (khq *khq_ML) MonthsNarrow() []string {
	return khq.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (khq *khq_ML) MonthWide(month time.Month) string {
	return khq.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (khq *khq_ML) MonthsWide() []string {
	return khq.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (khq *khq_ML) WeekdayAbbreviated(weekday time.Weekday) string {
	return khq.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (khq *khq_ML) WeekdaysAbbreviated() []string {
	return khq.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (khq *khq_ML) WeekdayNarrow(weekday time.Weekday) string {
	return khq.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (khq *khq_ML) WeekdaysNarrow() []string {
	return khq.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (khq *khq_ML) WeekdayShort(weekday time.Weekday) string {
	return khq.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (khq *khq_ML) WeekdaysShort() []string {
	return khq.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (khq *khq_ML) WeekdayWide(weekday time.Weekday) string {
	return khq.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (khq *khq_ML) WeekdaysWide() []string {
	return khq.daysWide
}

// Decimal returns the decimal point of number
func (khq *khq_ML) Decimal() string {
	return khq.decimal
}

// Group returns the group of number
func (khq *khq_ML) Group() string {
	return khq.group
}

// Group returns the minus sign of number
func (khq *khq_ML) Minus() string {
	return khq.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'khq_ML' and handles both Whole and Real numbers based on 'v'
func (khq *khq_ML) FmtNumber(num float64, v uint64) string {

	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'khq_ML' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (khq *khq_ML) FmtPercent(num float64, v uint64) string {
	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'khq_ML'
func (khq *khq_ML) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := khq.currencies[currency]
	l := len(s) + len(symbol) + 0 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, khq.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(khq.group) - 1; j >= 0; j-- {
					b = append(b, khq.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, khq.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, khq.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	b = append(b, symbol...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'khq_ML'
// in accounting notation.
func (khq *khq_ML) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := khq.currencies[currency]
	l := len(s) + len(symbol) + 0 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, khq.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(khq.group) - 1; j >= 0; j-- {
					b = append(b, khq.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {

		b = append(b, khq.minus[0])

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, khq.decimal...)
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

// FmtDateShort returns the short date representation of 't' for 'khq_ML'
func (khq *khq_ML) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'khq_ML'
func (khq *khq_ML) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, khq.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x2c, 0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'khq_ML'
func (khq *khq_ML) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, khq.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'khq_ML'
func (khq *khq_ML) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, khq.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, khq.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'khq_ML'
func (khq *khq_ML) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, khq.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'khq_ML'
func (khq *khq_ML) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, khq.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, khq.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'khq_ML'
func (khq *khq_ML) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, khq.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, khq.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'khq_ML'
func (khq *khq_ML) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, khq.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, khq.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := khq.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
