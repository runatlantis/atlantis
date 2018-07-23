package bem_ZM

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type bem_ZM struct {
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

// New returns a new instance of translator for the 'bem_ZM' locale
func New() locales.Translator {
	return &bem_ZM{
		locale:                 "bem_ZM",
		pluralsCardinal:        []locales.PluralRule{2, 6},
		pluralsOrdinal:         nil,
		pluralsRange:           nil,
		timeSeparator:          ":",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
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
		timezones:              map[string]string{"TMST": "TMST", "HECU": "HECU", "WAT": "WAT", "WAST": "WAST", "SGT": "SGT", "EDT": "EDT", "WARST": "WARST", "CST": "CST", "AWST": "AWST", "ACST": "ACST", "HEPM": "HEPM", "CLT": "CLT", "COST": "COST", "SAST": "SAST", "ACDT": "ACDT", "OEZ": "OEZ", "ARST": "ARST", "JDT": "JDT", "MST": "MST", "WIT": "WIT", "WIB": "WIB", "AKDT": "AKDT", "HEEG": "HEEG", "IST": "IST", "WART": "WART", "HAT": "HAT", "WEZ": "WEZ", "ECT": "ECT", "HNOG": "HNOG", "MDT": "MDT", "GYT": "GYT", "PST": "PST", "HNEG": "HNEG", "MEZ": "MEZ", "VET": "VET", "BOT": "BOT", "HKST": "HKST", "OESZ": "OESZ", "WITA": "WITA", "HENOMX": "HENOMX", "MYT": "MYT", "ACWST": "ACWST", "HEOG": "HEOG", "MESZ": "MESZ", "HKT": "HKT", "HNPM": "HNPM", "BT": "BT", "ACWDT": "ACWDT", "HAST": "HAST", "ART": "ART", "GMT": "GMT", "UYT": "UYT", "AEDT": "AEDT", "SRT": "SRT", "COT": "COT", "AWDT": "AWDT", "AST": "AST", "ADT": "ADT", "LHST": "LHST", "LHDT": "LHDT", "HNT": "HNT", "PDT": "PDT", "AEST": "AEST", "NZDT": "NZDT", "∅∅∅": "∅∅∅", "HNNOMX": "HNNOMX", "CAT": "CAT", "HNCU": "HNCU", "WESZ": "WESZ", "HNPMX": "HNPMX", "HEPMX": "HEPMX", "AKST": "AKST", "EAT": "EAT", "ChST": "ChST", "CHAST": "CHAST", "CDT": "CDT", "EST": "EST", "CLST": "CLST", "TMT": "TMT", "HADT": "HADT", "UYST": "UYST", "CHADT": "CHADT", "JST": "JST", "NZST": "NZST", "GFT": "GFT"},
	}
}

// Locale returns the current translators string locale
func (bem *bem_ZM) Locale() string {
	return bem.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'bem_ZM'
func (bem *bem_ZM) PluralsCardinal() []locales.PluralRule {
	return bem.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'bem_ZM'
func (bem *bem_ZM) PluralsOrdinal() []locales.PluralRule {
	return bem.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'bem_ZM'
func (bem *bem_ZM) PluralsRange() []locales.PluralRule {
	return bem.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'bem_ZM'
func (bem *bem_ZM) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'bem_ZM'
func (bem *bem_ZM) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'bem_ZM'
func (bem *bem_ZM) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (bem *bem_ZM) MonthAbbreviated(month time.Month) string {
	return bem.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (bem *bem_ZM) MonthsAbbreviated() []string {
	return bem.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (bem *bem_ZM) MonthNarrow(month time.Month) string {
	return bem.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (bem *bem_ZM) MonthsNarrow() []string {
	return bem.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (bem *bem_ZM) MonthWide(month time.Month) string {
	return bem.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (bem *bem_ZM) MonthsWide() []string {
	return bem.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (bem *bem_ZM) WeekdayAbbreviated(weekday time.Weekday) string {
	return bem.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (bem *bem_ZM) WeekdaysAbbreviated() []string {
	return bem.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (bem *bem_ZM) WeekdayNarrow(weekday time.Weekday) string {
	return bem.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (bem *bem_ZM) WeekdaysNarrow() []string {
	return bem.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (bem *bem_ZM) WeekdayShort(weekday time.Weekday) string {
	return bem.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (bem *bem_ZM) WeekdaysShort() []string {
	return bem.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (bem *bem_ZM) WeekdayWide(weekday time.Weekday) string {
	return bem.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (bem *bem_ZM) WeekdaysWide() []string {
	return bem.daysWide
}

// Decimal returns the decimal point of number
func (bem *bem_ZM) Decimal() string {
	return bem.decimal
}

// Group returns the group of number
func (bem *bem_ZM) Group() string {
	return bem.group
}

// Group returns the minus sign of number
func (bem *bem_ZM) Minus() string {
	return bem.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'bem_ZM' and handles both Whole and Real numbers based on 'v'
func (bem *bem_ZM) FmtNumber(num float64, v uint64) string {

	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'bem_ZM' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (bem *bem_ZM) FmtPercent(num float64, v uint64) string {
	return strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'bem_ZM'
func (bem *bem_ZM) FmtCurrency(num float64, v uint64, currency currency.Type) string {

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

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'bem_ZM'
// in accounting notation.
func (bem *bem_ZM) FmtAccounting(num float64, v uint64, currency currency.Type) string {

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

// FmtDateShort returns the short date representation of 't' for 'bem_ZM'
func (bem *bem_ZM) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'bem_ZM'
func (bem *bem_ZM) FmtDateMedium(t time.Time) string {

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

// FmtDateLong returns the long date representation of 't' for 'bem_ZM'
func (bem *bem_ZM) FmtDateLong(t time.Time) string {

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

// FmtDateFull returns the full date representation of 't' for 'bem_ZM'
func (bem *bem_ZM) FmtDateFull(t time.Time) string {

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

// FmtTimeShort returns the short time representation of 't' for 'bem_ZM'
func (bem *bem_ZM) FmtTimeShort(t time.Time) string {

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

// FmtTimeMedium returns the medium time representation of 't' for 'bem_ZM'
func (bem *bem_ZM) FmtTimeMedium(t time.Time) string {

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

// FmtTimeLong returns the long time representation of 't' for 'bem_ZM'
func (bem *bem_ZM) FmtTimeLong(t time.Time) string {

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

// FmtTimeFull returns the full time representation of 't' for 'bem_ZM'
func (bem *bem_ZM) FmtTimeFull(t time.Time) string {

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
