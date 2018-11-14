package bem

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type bem struct {
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

// New returns a new instance of translator for the 'bem' locale
func New() locales.Translator {
	return &bem{
		locale:                 "bem",
		pluralsCardinal:        []locales.PluralRule{2, 6},
		pluralsOrdinal:         nil,
		pluralsRange:           nil,
		timeSeparator:          ":",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "K", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyNegativePrefix: "(",
		currencyNegativeSuffix: ")",
		monthsAbbreviated:      []string{"", "Jan", "Feb", "Mac", "Epr", "Mei", "Jun", "Jul", "Oga", "Sep", "Okt", "Nov", "Dis"},
		monthsNarrow:           []string{"", "J", "F", "M", "E", "M", "J", "J", "O", "S", "O", "N", "D"},
		monthsWide:             []string{"", "Januari", "Februari", "Machi", "Epreo", "Mei", "Juni", "Julai", "Ogasti", "Septemba", "Oktoba", "Novemba", "Disemba"},
		daysWide:               []string{"Pa Mulungu", "Palichimo", "Palichibuli", "Palichitatu", "Palichine", "Palichisano", "Pachibelushi"},
		periodsAbbreviated:     []string{"uluchelo", "akasuba"},
		periodsWide:            []string{"uluchelo", "akasuba"},
		erasAbbreviated:        []string{"BC", "AD"},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"Before Yesu", "After Yesu"},
		timezones:              map[string]string{"HADT": "HADT", "HECU": "HECU", "AWST": "AWST", "AWDT": "AWDT", "MDT": "MDT", "WIB": "WIB", "COST": "COST", "HAST": "HAST", "EST": "EST", "WART": "WART", "SGT": "SGT", "ACWDT": "ACWDT", "NZDT": "NZDT", "VET": "VET", "WAT": "WAT", "JST": "JST", "JDT": "JDT", "MEZ": "MEZ", "HENOMX": "HENOMX", "HNCU": "HNCU", "WEZ": "WEZ", "CHAST": "CHAST", "HEPMX": "HEPMX", "HEPM": "HEPM", "CLT": "CLT", "GMT": "GMT", "MST": "MST", "COT": "COT", "CHADT": "CHADT", "ACST": "ACST", "OESZ": "OESZ", "WAST": "WAST", "BOT": "BOT", "HEOG": "HEOG", "WARST": "WARST", "EAT": "EAT", "CDT": "CDT", "HNT": "HNT", "UYST": "UYST", "PDT": "PDT", "AKST": "AKST", "ACDT": "ACDT", "HNNOMX": "HNNOMX", "SRT": "SRT", "WESZ": "WESZ", "SAST": "SAST", "ARST": "ARST", "HNPM": "HNPM", "CLST": "CLST", "TMT": "TMT", "HNPMX": "HNPMX", "PST": "PST", "MESZ": "MESZ", "WITA": "WITA", "CAT": "CAT", "UYT": "UYT", "BT": "BT", "LHDT": "LHDT", "ART": "ART", "AEST": "AEST", "AST": "AST", "GFT": "GFT", "ECT": "ECT", "ACWST": "ACWST", "HEEG": "HEEG", "HNOG": "HNOG", "GYT": "GYT", "CST": "CST", "HKST": "HKST", "IST": "IST", "LHST": "LHST", "EDT": "EDT", "HKT": "HKT", "AEDT": "AEDT", "AKDT": "AKDT", "ChST": "ChST", "ADT": "ADT", "NZST": "NZST", "HNEG": "HNEG", "HAT": "HAT", "WIT": "WIT", "TMST": "TMST", "MYT": "MYT", "∅∅∅": "∅∅∅", "OEZ": "OEZ"},
	}
}

// Locale returns the current translators string locale
func (bem *bem) Locale() string {
	return bem.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'bem'
func (bem *bem) PluralsCardinal() []locales.PluralRule {
	return bem.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'bem'
func (bem *bem) PluralsOrdinal() []locales.PluralRule {
	return bem.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'bem'
func (bem *bem) PluralsRange() []locales.PluralRule {
	return bem.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'bem'
func (bem *bem) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'bem'
func (bem *bem) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'bem'
func (bem *bem) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (bem *bem) MonthAbbreviated(month time.Month) string {
	return bem.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (bem *bem) MonthsAbbreviated() []string {
	return bem.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (bem *bem) MonthNarrow(month time.Month) string {
	return bem.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (bem *bem) MonthsNarrow() []string {
	return bem.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (bem *bem) MonthWide(month time.Month) string {
	return bem.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (bem *bem) MonthsWide() []string {
	return bem.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (bem *bem) WeekdayAbbreviated(weekday time.Weekday) string {
	return bem.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (bem *bem) WeekdaysAbbreviated() []string {
	return bem.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (bem *bem) WeekdayNarrow(weekday time.Weekday) string {
	return bem.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (bem *bem) WeekdaysNarrow() []string {
	return bem.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (bem *bem) WeekdayShort(weekday time.Weekday) string {
	return bem.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (bem *bem) WeekdaysShort() []string {
	return bem.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (bem *bem) WeekdayWide(weekday time.Weekday) string {
	return bem.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (bem *bem) WeekdaysWide() []string {
	return bem.daysWide
}

// Decimal returns the decimal point of number
func (bem *bem) Decimal() string {
	return bem.decimal
}

// Group returns the group of number
func (bem *bem) Group() string {
	return bem.group
}

// Group returns the minus sign of number
func (bem *bem) Minus() string {
	return bem.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'bem' and handles both Whole and Real numbers based on 'v'
func (bem *bem) FmtNumber(num float64, v uint64) string {

	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'bem' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (bem *bem) FmtPercent(num float64, v uint64) string {
	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'bem'
func (bem *bem) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := bem.currencies[currency]
	l := len(s) + len(symbol) + 0
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, bem.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, bem.group[0])
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
		b = append(b, bem.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, bem.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'bem'
// in accounting notation.
func (bem *bem) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := bem.currencies[currency]
	l := len(s) + len(symbol) + 2
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, bem.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, bem.group[0])
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

		b = append(b, bem.currencyNegativePrefix[0])

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
			b = append(b, bem.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, bem.currencyNegativeSuffix...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'bem'
func (bem *bem) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'bem'
func (bem *bem) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, bem.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'bem'
func (bem *bem) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, bem.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'bem'
func (bem *bem) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, bem.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, bem.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'bem'
func (bem *bem) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, bem.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, bem.periodsAbbreviated[0]...)
	} else {
		b = append(b, bem.periodsAbbreviated[1]...)
	}

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'bem'
func (bem *bem) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, bem.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, bem.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, bem.periodsAbbreviated[0]...)
	} else {
		b = append(b, bem.periodsAbbreviated[1]...)
	}

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'bem'
func (bem *bem) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, bem.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, bem.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, bem.periodsAbbreviated[0]...)
	} else {
		b = append(b, bem.periodsAbbreviated[1]...)
	}

	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'bem'
func (bem *bem) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, bem.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, bem.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, bem.periodsAbbreviated[0]...)
	} else {
		b = append(b, bem.periodsAbbreviated[1]...)
	}

	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := bem.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
