package zu_ZA

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type zu_ZA struct {
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

// New returns a new instance of translator for the 'zu_ZA' locale
func New() locales.Translator {
	return &zu_ZA{
		locale:             "zu_ZA",
		pluralsCardinal:    []locales.PluralRule{2, 6},
		pluralsOrdinal:     []locales.PluralRule{6},
		pluralsRange:       []locales.PluralRule{2, 6},
		decimal:            ".",
		group:              ",",
		minus:              "-",
		percent:            "%",
		perMille:           "‰",
		timeSeparator:      ":",
		inifinity:          "∞",
		currencies:         []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		monthsAbbreviated:  []string{"", "Jan", "Feb", "Mas", "Eph", "Mey", "Jun", "Jul", "Aga", "Sep", "Okt", "Nov", "Dis"},
		monthsNarrow:       []string{"", "J", "F", "M", "E", "M", "J", "J", "A", "S", "O", "N", "D"},
		monthsWide:         []string{"", "Januwari", "Februwari", "Mashi", "Ephreli", "Meyi", "Juni", "Julayi", "Agasti", "Septhemba", "Okthoba", "Novemba", "Disemba"},
		daysAbbreviated:    []string{"Son", "Mso", "Bil", "Tha", "Sin", "Hla", "Mgq"},
		daysNarrow:         []string{"S", "M", "B", "T", "S", "H", "M"},
		daysShort:          []string{"Son", "Mso", "Bil", "Tha", "Sin", "Hla", "Mgq"},
		daysWide:           []string{"ISonto", "UMsombuluko", "ULwesibili", "ULwesithathu", "ULwesine", "ULwesihlanu", "UMgqibelo"},
		periodsAbbreviated: []string{"AM", "PM"},
		periodsNarrow:      []string{"a", "p"},
		periodsWide:        []string{"AM", "PM"},
		erasAbbreviated:    []string{"BC", "AD"},
		erasNarrow:         []string{"", ""},
		erasWide:           []string{"BC", "AD"},
		timezones:          map[string]string{"HNPM": "Iikhathi sase-Saint Pierre nase-Miquelon esijwayelekile", "EDT": "Isikhathi sase-North American East sasemini", "ECT": "Isikhathi sase-Ecuador", "EST": "Isikhathi sase-North American East esijwayelekile", "ACWST": "Isikhathi sase-Australian Central West esivamile", "ACWDT": "Isikhathi sasemini sase-Australian Central West", "TMST": "Isikhathi sehlobo sase-Turkmenistan", "MST": "Isikhathi sase-North American Mountain esijwayelekile", "HADT": "Isikhathi sase-Hawaii-Aleutia sasemini", "AWST": "Isikhathi sase-Australian Western esivamile", "HENOMX": "Isikhathi sase-Northwest Mexico sasemini", "BT": "Isikhathi sase-Bhutan", "AKDT": "Isikhathi sase-Alaska sasemini", "HKT": "Isikhathi esivamile sase-Hong Kong", "SRT": "Isikhathi sase-Suriname", "OESZ": "Isikhathi sasehlobo sase-Eastern Europe", "AST": "Isikhathi sase-Atlantic esijwayelekile", "WAST": "Isikhathi sasehlobo saseNtshonalanga Afrika", "JDT": "Isikhathi semini sase-Japan", "HKST": "Isikhathi sehlobo sase-Hong Kong", "HNNOMX": "Isikhathi sase-Northwest Mexico esijwayelekile", "COST": "Isikhathi sase-Colombia sasehlobo", "CHAST": "Isikhathi esivamile sase-Chatham", "AEST": "Isikhathi esivamile sase-Australian East", "HNOG": "Isikhathi sase-West Greenland esijwayelekile", "LHDT": "Isikhathi sase-Lord Howe sasemini", "HNT": "Isikhathi sase-Newfoundland esijwayelekile", "EAT": "Isikhathi saseMpumalanga Afrika", "CLST": "Isikhathi sase-Chile sasehlobo", "HEPMX": "Isikhathi sase-Mexican Pacific sasemini", "AEDT": "Isikhathi sasemini sase-Australian East", "BOT": "Isikhathi sase-Bolivia", "ACDT": "Isikhathi sase-Australian Central sasemini", "HEEG": "Isikhathi sase-East Greenland sasemini", "LHST": "Isikhathi sase-Lord Howe esivamile", "TMT": "Isikhathi esivamile sase-Turkmenistan", "HAST": "Isikhathi sase-Hawaii-Aleutia esijwayelekile", "AWDT": "Isikhathi sase-Australian Western sasemini", "ADT": "Isikhathi sase-Atlantic sasemini", "CAT": "Isikhathi sase-Central Africa", "OEZ": "Isikhathi esijwayelekile sase-Eastern Europe", "PDT": "Isikhathi sase-North American Pacific sasemini", "WEZ": "Isikhathi esijwayelekile sase-Western Europe", "HNPMX": "Isikhathi sase-Mexican Pacific esijwayelekile", "UYST": "Isikhathi sase-Uruguay sasehlobo", "JST": "Isikhathi esivamile sase-Japan", "WART": "Isikhathi saseNyakatho ne-Argentina esijwayelekile", "HAT": "Isikhathi sase-Newfoundland sasemini", "WIT": "Isikhathi sase-Eastern Indonesia", "HECU": "Isikhathi sase-Cuba sasemini", "PST": "Isikhathi sase-North American Pacific esijwayelekile", "GFT": "Isikhathi sase-French Guiana", "IST": "Isikhathi sase-India esivamile", "COT": "Isikhathi sase-Colombia esijwayelekile", "UYT": "Isikhathi sase-Uruguay esijwayelekile", "CDT": "Isikhathi sase-North American Central sasemini", "∅∅∅": "∅∅∅", "WESZ": "Isikhathi sasehlobo sase-Western Europe", "WIB": "Isikhathi sase-Western Indonesia", "NZDT": "Isikhathi sasemini sase-New Zealand", "HNEG": "Isikhathi sase-East Greenland esijwayelekile", "HNCU": "Isikhathi sase-Cuba esijwayelekile", "SAST": "Isikhathi esijwayelekile saseNingizimu Afrika", "SGT": "Isikhathi esivamile sase-Singapore", "ACST": "Isikhathi sase-Australian Central esivamile", "HEOG": "Isikhathi sase-West Greenland sasehlobo", "MEZ": "Isikhathi esijwayelekile sase-Central Europe", "MESZ": "Isikhathi sasehlobo sase-Central Europe", "VET": "Isikhathi sase-Venezuela", "HEPM": "Isikhathi sase-Saint Pierre nase-Miquelon sasemini", "MDT": "Isikhathi sase-North American Mountain sasemini", "ChST": "Isikhathi esijwayelekile sase-Chamorro", "CST": "Isikhathi sase-North American Central esijwayelekile", "GYT": "Isikhathi sase-Guyana", "MYT": "Isikhathi sase-Malaysia", "AKST": "Isikhathi sase-Alaska esijwayelekile", "WITA": "Isikhathi sase-Central Indonesia", "ARST": "Isikhathi sase-Argentina sasehlobo", "GMT": "Isikhathi sase-Greenwich Mean", "WAT": "Isikhathi esijwayelekile saseNtshonalanga Afrika", "ART": "Isikhathi sase-Argentina esijwayelekile", "WARST": "Isikhathi saseNyakatho ne-Argentina sasehlobo", "CLT": "Isikhathi sase-Chile esijwayelekile", "CHADT": "Isikhathi sasemini sase-Chatham", "NZST": "Isikhathi esivamile sase-New Zealand"},
	}
}

// Locale returns the current translators string locale
func (zu *zu_ZA) Locale() string {
	return zu.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'zu_ZA'
func (zu *zu_ZA) PluralsCardinal() []locales.PluralRule {
	return zu.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'zu_ZA'
func (zu *zu_ZA) PluralsOrdinal() []locales.PluralRule {
	return zu.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'zu_ZA'
func (zu *zu_ZA) PluralsRange() []locales.PluralRule {
	return zu.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'zu_ZA'
func (zu *zu_ZA) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)
	i := int64(n)

	if (i == 0) || (n == 1) {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'zu_ZA'
func (zu *zu_ZA) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'zu_ZA'
func (zu *zu_ZA) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {

	start := zu.CardinalPluralRule(num1, v1)
	end := zu.CardinalPluralRule(num2, v2)

	if start == locales.PluralRuleOne && end == locales.PluralRuleOne {
		return locales.PluralRuleOne
	} else if start == locales.PluralRuleOne && end == locales.PluralRuleOther {
		return locales.PluralRuleOther
	}

	return locales.PluralRuleOther

}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (zu *zu_ZA) MonthAbbreviated(month time.Month) string {
	return zu.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (zu *zu_ZA) MonthsAbbreviated() []string {
	return zu.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (zu *zu_ZA) MonthNarrow(month time.Month) string {
	return zu.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (zu *zu_ZA) MonthsNarrow() []string {
	return zu.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (zu *zu_ZA) MonthWide(month time.Month) string {
	return zu.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (zu *zu_ZA) MonthsWide() []string {
	return zu.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (zu *zu_ZA) WeekdayAbbreviated(weekday time.Weekday) string {
	return zu.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (zu *zu_ZA) WeekdaysAbbreviated() []string {
	return zu.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (zu *zu_ZA) WeekdayNarrow(weekday time.Weekday) string {
	return zu.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (zu *zu_ZA) WeekdaysNarrow() []string {
	return zu.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (zu *zu_ZA) WeekdayShort(weekday time.Weekday) string {
	return zu.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (zu *zu_ZA) WeekdaysShort() []string {
	return zu.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (zu *zu_ZA) WeekdayWide(weekday time.Weekday) string {
	return zu.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (zu *zu_ZA) WeekdaysWide() []string {
	return zu.daysWide
}

// Decimal returns the decimal point of number
func (zu *zu_ZA) Decimal() string {
	return zu.decimal
}

// Group returns the group of number
func (zu *zu_ZA) Group() string {
	return zu.group
}

// Group returns the minus sign of number
func (zu *zu_ZA) Minus() string {
	return zu.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'zu_ZA' and handles both Whole and Real numbers based on 'v'
func (zu *zu_ZA) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, zu.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, zu.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, zu.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'zu_ZA' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (zu *zu_ZA) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 3
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, zu.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, zu.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, zu.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'zu_ZA'
func (zu *zu_ZA) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := zu.currencies[currency]
	l := len(s) + len(symbol) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, zu.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, zu.group[0])
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
		b = append(b, zu.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, zu.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'zu_ZA'
// in accounting notation.
func (zu *zu_ZA) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := zu.currencies[currency]
	l := len(s) + len(symbol) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, zu.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, zu.group[0])
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

		b = append(b, zu.minus[0])

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
			b = append(b, zu.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'zu_ZA'
func (zu *zu_ZA) FmtDateShort(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Month()), 10)
	b = append(b, []byte{0x2f}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2f}...)

	if t.Year() > 9 {
		b = append(b, strconv.Itoa(t.Year())[2:]...)
	} else {
		b = append(b, strconv.Itoa(t.Year())[1:]...)
	}

	return string(b)
}

// FmtDateMedium returns the medium date representation of 't' for 'zu_ZA'
func (zu *zu_ZA) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, zu.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2c, 0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'zu_ZA'
func (zu *zu_ZA) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, zu.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2c, 0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'zu_ZA'
func (zu *zu_ZA) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, zu.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)
	b = append(b, zu.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2c, 0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'zu_ZA'
func (zu *zu_ZA) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, zu.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'zu_ZA'
func (zu *zu_ZA) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, zu.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, zu.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'zu_ZA'
func (zu *zu_ZA) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, zu.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, zu.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'zu_ZA'
func (zu *zu_ZA) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, zu.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, zu.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := zu.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
