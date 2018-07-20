package zu

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type zu struct {
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

// New returns a new instance of translator for the 'zu' locale
func New() locales.Translator {
	return &zu{
		locale:             "zu",
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
		currencies:         []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "A$", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "R$", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CA$", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CN¥", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "€", "FIM", "FJD", "FKP", "FRF", "£", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HK$", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "₪", "₹", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JP¥", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "₩", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MX$", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZ$", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "฿", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "NT$", "TZS", "UAH", "UAK", "UGS", "UGX", "US$", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "₫", "VNN", "VUV", "WST", "FCFA", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "EC$", "XDR", "XEU", "XFO", "XFU", "CFA", "XPD", "CFPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "R", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
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
		timezones:          map[string]string{"HEPM": "Isikhathi sase-Saint Pierre nase-Miquelon sasemini", "WIT": "Isikhathi sase-Eastern Indonesia", "HNCU": "Isikhathi sase-Cuba esijwayelekile", "CDT": "Isikhathi sase-North American Central sasemini", "NZST": "Isikhathi esivamile sase-New Zealand", "GFT": "Isikhathi sase-French Guiana", "ACWST": "Isikhathi sase-Australian Central West esivamile", "CLT": "Isikhathi sase-Chile esijwayelekile", "CST": "Isikhathi sase-North American Central esijwayelekile", "PST": "Isikhathi sase-North American Pacific esijwayelekile", "BOT": "Isikhathi sase-Bolivia", "HNEG": "Isikhathi sase-East Greenland esijwayelekile", "HNPM": "Iikhathi sase-Saint Pierre nase-Miquelon esijwayelekile", "AEDT": "Isikhathi sasemini sase-Australian East", "SAST": "Isikhathi esijwayelekile saseNingizimu Afrika", "BT": "Isikhathi sase-Bhutan", "NZDT": "Isikhathi sasemini sase-New Zealand", "ECT": "Isikhathi sase-Ecuador", "HNNOMX": "Isikhathi sase-Northwest Mexico esijwayelekile", "MDT": "MDT", "CLST": "Isikhathi sase-Chile sasehlobo", "AWDT": "Isikhathi sase-Australian Western sasemini", "WAST": "Isikhathi sasehlobo saseNtshonalanga Afrika", "JST": "Isikhathi esivamile sase-Japan", "MESZ": "Isikhathi sasehlobo sase-Central Europe", "CHAST": "Isikhathi esivamile sase-Chatham", "JDT": "Isikhathi semini sase-Japan", "ACDT": "Isikhathi sase-Australian Central sasemini", "HEEG": "Isikhathi sase-East Greenland sasemini", "LHST": "Isikhathi sase-Lord Howe esivamile", "LHDT": "Isikhathi sase-Lord Howe sasemini", "EAT": "Isikhathi saseMpumalanga Afrika", "OESZ": "Isikhathi sasehlobo sase-Eastern Europe", "MEZ": "Isikhathi esijwayelekile sase-Central Europe", "HNT": "Isikhathi sase-Newfoundland esijwayelekile", "COST": "Isikhathi sase-Colombia sasehlobo", "ChST": "Isikhathi esijwayelekile sase-Chamorro", "HEPMX": "Isikhathi sase-Mexican Pacific sasemini", "WAT": "Isikhathi esijwayelekile saseNtshonalanga Afrika", "WART": "Isikhathi saseNyakatho ne-Argentina esijwayelekile", "CAT": "Isikhathi sase-Central Africa", "TMT": "Isikhathi esivamile sase-Turkmenistan", "COT": "Isikhathi sase-Colombia esijwayelekile", "WEZ": "Isikhathi esijwayelekile sase-Western Europe", "HNOG": "Isikhathi sase-West Greenland esijwayelekile", "AEST": "Isikhathi esivamile sase-Australian East", "ACST": "Isikhathi sase-Australian Central esivamile", "HKT": "Isikhathi esivamile sase-Hong Kong", "∅∅∅": "Isikhathi sasehlobo sase-Azores", "WITA": "Isikhathi sase-Central Indonesia", "TMST": "Isikhathi sehlobo sase-Turkmenistan", "MYT": "Isikhathi sase-Malaysia", "EST": "Isikhathi sase-North American East esijwayelekile", "EDT": "Isikhathi sase-North American East sasemini", "HAT": "Isikhathi sase-Newfoundland sasemini", "UYST": "Isikhathi sase-Uruguay sasehlobo", "HECU": "Isikhathi sase-Cuba sasemini", "AWST": "Isikhathi sase-Australian Western esivamile", "SGT": "Isikhathi esivamile sase-Singapore", "WARST": "Isikhathi saseNyakatho ne-Argentina sasehlobo", "HADT": "Isikhathi sase-Hawaii-Aleutia sasemini", "ARST": "Isikhathi sase-Argentina sasehlobo", "UYT": "Isikhathi sase-Uruguay esijwayelekile", "HNPMX": "Isikhathi sase-Mexican Pacific esijwayelekile", "AKDT": "Isikhathi sase-Alaska sasemini", "HENOMX": "Isikhathi sase-Northwest Mexico sasemini", "WIB": "Isikhathi sase-Western Indonesia", "ART": "Isikhathi sase-Argentina esijwayelekile", "PDT": "Isikhathi sase-North American Pacific sasemini", "ACWDT": "Isikhathi sasemini sase-Australian Central West", "HEOG": "Isikhathi sase-West Greenland sasehlobo", "HKST": "Isikhathi sehlobo sase-Hong Kong", "VET": "Isikhathi sase-Venezuela", "SRT": "Isikhathi sase-Suriname", "HAST": "Isikhathi sase-Hawaii-Aleutia esijwayelekile", "ADT": "Isikhathi sase-Atlantic sasemini", "WESZ": "Isikhathi sasehlobo sase-Western Europe", "OEZ": "Isikhathi esijwayelekile sase-Eastern Europe", "GMT": "Isikhathi sase-Greenwich Mean", "CHADT": "Isikhathi sasemini sase-Chatham", "IST": "Isikhathi sase-India esivamile", "MST": "MST", "GYT": "Isikhathi sase-Guyana", "AST": "Isikhathi sase-Atlantic esijwayelekile", "AKST": "Isikhathi sase-Alaska esijwayelekile"},
	}
}

// Locale returns the current translators string locale
func (zu *zu) Locale() string {
	return zu.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'zu'
func (zu *zu) PluralsCardinal() []locales.PluralRule {
	return zu.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'zu'
func (zu *zu) PluralsOrdinal() []locales.PluralRule {
	return zu.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'zu'
func (zu *zu) PluralsRange() []locales.PluralRule {
	return zu.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'zu'
func (zu *zu) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)
	i := int64(n)

	if (i == 0) || (n == 1) {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'zu'
func (zu *zu) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'zu'
func (zu *zu) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {

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
func (zu *zu) MonthAbbreviated(month time.Month) string {
	return zu.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (zu *zu) MonthsAbbreviated() []string {
	return zu.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (zu *zu) MonthNarrow(month time.Month) string {
	return zu.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (zu *zu) MonthsNarrow() []string {
	return zu.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (zu *zu) MonthWide(month time.Month) string {
	return zu.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (zu *zu) MonthsWide() []string {
	return zu.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (zu *zu) WeekdayAbbreviated(weekday time.Weekday) string {
	return zu.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (zu *zu) WeekdaysAbbreviated() []string {
	return zu.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (zu *zu) WeekdayNarrow(weekday time.Weekday) string {
	return zu.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (zu *zu) WeekdaysNarrow() []string {
	return zu.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (zu *zu) WeekdayShort(weekday time.Weekday) string {
	return zu.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (zu *zu) WeekdaysShort() []string {
	return zu.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (zu *zu) WeekdayWide(weekday time.Weekday) string {
	return zu.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (zu *zu) WeekdaysWide() []string {
	return zu.daysWide
}

// Decimal returns the decimal point of number
func (zu *zu) Decimal() string {
	return zu.decimal
}

// Group returns the group of number
func (zu *zu) Group() string {
	return zu.group
}

// Group returns the minus sign of number
func (zu *zu) Minus() string {
	return zu.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'zu' and handles both Whole and Real numbers based on 'v'
func (zu *zu) FmtNumber(num float64, v uint64) string {

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

// FmtPercent returns 'num' with digits/precision of 'v' for 'zu' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (zu *zu) FmtPercent(num float64, v uint64) string {
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

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'zu'
func (zu *zu) FmtCurrency(num float64, v uint64, currency currency.Type) string {

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

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'zu'
// in accounting notation.
func (zu *zu) FmtAccounting(num float64, v uint64, currency currency.Type) string {

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

// FmtDateShort returns the short date representation of 't' for 'zu'
func (zu *zu) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'zu'
func (zu *zu) FmtDateMedium(t time.Time) string {

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

// FmtDateLong returns the long date representation of 't' for 'zu'
func (zu *zu) FmtDateLong(t time.Time) string {

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

// FmtDateFull returns the full date representation of 't' for 'zu'
func (zu *zu) FmtDateFull(t time.Time) string {

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

// FmtTimeShort returns the short time representation of 't' for 'zu'
func (zu *zu) FmtTimeShort(t time.Time) string {

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

// FmtTimeMedium returns the medium time representation of 't' for 'zu'
func (zu *zu) FmtTimeMedium(t time.Time) string {

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

// FmtTimeLong returns the long time representation of 't' for 'zu'
func (zu *zu) FmtTimeLong(t time.Time) string {

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

// FmtTimeFull returns the full time representation of 't' for 'zu'
func (zu *zu) FmtTimeFull(t time.Time) string {

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
