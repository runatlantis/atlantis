package gd

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type gd struct {
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

// New returns a new instance of translator for the 'gd' locale
func New() locales.Translator {
	return &gd{
		locale:                 "gd",
		pluralsCardinal:        []locales.PluralRule{2, 3, 4, 6},
		pluralsOrdinal:         nil,
		pluralsRange:           nil,
		decimal:                ".",
		group:                  ",",
		minus:                  "-",
		percent:                "%",
		perMille:               "‰",
		timeSeparator:          ":",
		inifinity:              "∞",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "A$", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "R$", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CA$", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CN¥", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "€", "FIM", "FJD", "FKP", "FRF", "£", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HK$", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "₪", "₹", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JP¥", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "₩", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MX$", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZ$", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "฿", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "NT$", "TZS", "UAH", "UAK", "UGS", "UGX", "$", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "₫", "VNN", "VUV", "WST", "FCFA", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "EC$", "XDR", "XEU", "XFO", "XFU", "CFA", "XPD", "CFPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyNegativePrefix: "(",
		currencyNegativeSuffix: ")",
		monthsAbbreviated:      []string{"", "Faoi", "Gearr", "Màrt", "Gibl", "Cèit", "Ògmh", "Iuch", "Lùna", "Sult", "Dàmh", "Samh", "Dùbh"},
		monthsNarrow:           []string{"", "F", "G", "M", "G", "C", "Ò", "I", "L", "S", "D", "S", "D"},
		monthsWide:             []string{"", "dhen Fhaoilleach", "dhen Ghearran", "dhen Mhàrt", "dhen Ghiblean", "dhen Chèitean", "dhen Ògmhios", "dhen Iuchar", "dhen Lùnastal", "dhen t-Sultain", "dhen Dàmhair", "dhen t-Samhain", "dhen Dùbhlachd"},
		daysAbbreviated:        []string{"DiD", "DiL", "DiM", "DiC", "Dia", "Dih", "DiS"},
		daysNarrow:             []string{"D", "L", "M", "C", "A", "H", "S"},
		daysShort:              []string{"Dò", "Lu", "Mà", "Ci", "Da", "hA", "Sa"},
		daysWide:               []string{"DiDòmhnaich", "DiLuain", "DiMàirt", "DiCiadain", "DiarDaoin", "DihAoine", "DiSathairne"},
		periodsAbbreviated:     []string{"m", "f"},
		periodsNarrow:          []string{"m", "f"},
		periodsWide:            []string{"m", "f"},
		erasAbbreviated:        []string{"RC", "AD"},
		erasNarrow:             []string{"R", "A"},
		erasWide:               []string{"Ro Chrìosta", "An dèidh Chrìosta"},
		timezones:              map[string]string{"NZST": "Bun-àm Shealainn Nuaidh", "HNOG": "Bun-àm na Graonlainn an Iar", "MEZ": "Bun-àm Meadhan na Roinn-Eòrpa", "HKT": "Bun-àm Hong Kong", "WIB": "Àm nan Innd-Innse an Iar", "AKST": "Bun-àm Alaska", "HNEG": "Bun-àm na Graonlainn an Ear", "ACDT": "Tìde samhraidh Meadhan Astràilia", "WARST": "Tìde samhraidh na h-Argantaine Siaraich", "CST": "Bun-àm Meadhan Aimeireaga a Tuath", "WESZ": "Tìde samhraidh na Roinn-Eòrpa an Iar", "BT": "Àm Butàin", "EDT": "Tìde samhraidh Aimeireaga a Tuath an Ear", "MESZ": "Tìde samhraidh Meadhan na Roinn-Eòrpa", "ART": "Bun-àm na h-Argantaine", "PST": "Bun-àm a’ Chuain Sèimh", "COT": "Bun-àm Coloimbia", "HADT": "Tìde Samhraidh nan Eileanan Hawai’i ’s Aleutach", "LHST": "Bun-àm Lord Howe", "TMST": "Tìde samhraidh Turcmanastàin", "EAT": "Àm Afraga an Ear", "AEDT": "Tìde samhraidh Astràilia an Ear", "JST": "Bun-àm na Seapaine", "NZDT": "Tìde samhraidh Shealainn Nuaidh", "ECT": "Àm Eacuadoir", "EST": "Bun-àm Aimeireaga a Tuath an Ear", "CLT": "Bun-àm na Sile", "UYT": "Bun-àm Uruguaidh", "HNPMX": "Bun-àm a’ Chuain Sèimh Mheagsago", "OESZ": "Tìde samhraidh na Roinn-Eòrpa an Ear", "CDT": "Tìde samhraidh Meadhan Aimeireaga a Tuath", "PDT": "Tìde samhraidh a’ Chuain Sèimh", "WEZ": "Bun-àm na Roinn-Eòrpa an Iar", "MYT": "Àm Mhalaidhsea", "HAT": "Tìde samhraidh Talamh an Èisg", "HEPM": "Tìde Samhraidh Saint Pierre agus Miquelon", "HENOMX": "Tìde samhraidh Mheagsago an Iar-thuath", "AKDT": "Tìde samhraidh Alaska", "WAT": "Bun-àm Afraga an Iar", "HNNOMX": "Bun-àm Mheagsago an Iar-thuath", "AWDT": "Tìde samhraidh Astràilia an Iar", "ADT": "Tìde samhraidh a’ Chuain Siar", "ACWDT": "Tìde samhraidh Meadhan Astràilia an Iar", "SAST": "Àm Afraga a Deas", "BOT": "Àm Boilibhia", "GFT": "Àm Guidheàna na Frainge", "CHAST": "Bun-àm Chatham", "HKST": "Tìde samhraidh Hong Kong", "CAT": "Àm Meadhan Afraga", "CLST": "Tìde samhraidh na Sile", "HEOG": "Tìde samhraidh na Graonlainn an Iar", "WART": "Bun-àm na h-Argantaine Siaraich", "MST": "Bun-àm Macàthu", "SGT": "Àm Singeapòr", "ACST": "Bun-àm Meadhan Astràilia", "SRT": "Àm Suranaim", "ARST": "Tìde samhraidh na h-Argantaine", "JDT": "Tìde samhraidh na Seapaine", "HNCU": "Bun-àm Cùba", "AST": "Bun-àm a’ Chuain Siar", "WIT": "Àm nan Innd-Innse an Ear", "COST": "Tìde samhraidh Coloimbia", "UYST": "Tìde samhraidh Uruguaidh", "MDT": "Tìde samhraidh Macàthu", "HAST": "Bun-àm nan Eileanan Hawai’i ’s Aleutach", "HEPMX": "Tìde samhraidh a’ Chuain Sèimh Mheagsago", "ChST": "Àm Chamorro", "AEST": "Bun-àm Astràilia an Ear", "WITA": "Àm Meadhan nan Innd-Innse", "HNPM": "Bun-àm Saint Pierre agus Miquelon", "TMT": "Bun-àm Turcmanastàin", "GYT": "Àm Guidheàna", "WAST": "Tìde Samhraidh Afraga an Iar", "ACWST": "Bun-àm Meadhan Astràilia an Iar", "IST": "Àm nan Innseachan", "OEZ": "Bun-àm na Roinn-Eòrpa an Ear", "GMT": "Greenwich Mean Time", "VET": "Àm na Bheiniseala", "CHADT": "Tìde samhraidh Chatham", "HECU": "Tìde samhraidh Cùba", "AWST": "Bun-àm Astràilia an Iar", "HEEG": "Tìde samhraidh na Graonlainn an Ear", "∅∅∅": "Tìde samhraidh nan Eileanan Asorach", "LHDT": "Tìde samhraidh Lord Howe", "HNT": "Bun-àm Talamh an Èisg"},
	}
}

// Locale returns the current translators string locale
func (gd *gd) Locale() string {
	return gd.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'gd'
func (gd *gd) PluralsCardinal() []locales.PluralRule {
	return gd.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'gd'
func (gd *gd) PluralsOrdinal() []locales.PluralRule {
	return gd.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'gd'
func (gd *gd) PluralsRange() []locales.PluralRule {
	return gd.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'gd'
func (gd *gd) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 || n == 11 {
		return locales.PluralRuleOne
	} else if n == 2 || n == 12 {
		return locales.PluralRuleTwo
	} else if (n >= 3 && n <= 10) || (n >= 13 && n <= 19) {
		return locales.PluralRuleFew
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'gd'
func (gd *gd) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'gd'
func (gd *gd) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (gd *gd) MonthAbbreviated(month time.Month) string {
	return gd.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (gd *gd) MonthsAbbreviated() []string {
	return gd.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (gd *gd) MonthNarrow(month time.Month) string {
	return gd.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (gd *gd) MonthsNarrow() []string {
	return gd.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (gd *gd) MonthWide(month time.Month) string {
	return gd.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (gd *gd) MonthsWide() []string {
	return gd.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (gd *gd) WeekdayAbbreviated(weekday time.Weekday) string {
	return gd.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (gd *gd) WeekdaysAbbreviated() []string {
	return gd.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (gd *gd) WeekdayNarrow(weekday time.Weekday) string {
	return gd.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (gd *gd) WeekdaysNarrow() []string {
	return gd.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (gd *gd) WeekdayShort(weekday time.Weekday) string {
	return gd.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (gd *gd) WeekdaysShort() []string {
	return gd.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (gd *gd) WeekdayWide(weekday time.Weekday) string {
	return gd.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (gd *gd) WeekdaysWide() []string {
	return gd.daysWide
}

// Decimal returns the decimal point of number
func (gd *gd) Decimal() string {
	return gd.decimal
}

// Group returns the group of number
func (gd *gd) Group() string {
	return gd.group
}

// Group returns the minus sign of number
func (gd *gd) Minus() string {
	return gd.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'gd' and handles both Whole and Real numbers based on 'v'
func (gd *gd) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, gd.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, gd.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, gd.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'gd' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (gd *gd) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 3
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, gd.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, gd.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, gd.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'gd'
func (gd *gd) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := gd.currencies[currency]
	l := len(s) + len(symbol) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, gd.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, gd.group[0])
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
		b = append(b, gd.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, gd.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'gd'
// in accounting notation.
func (gd *gd) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := gd.currencies[currency]
	l := len(s) + len(symbol) + 4 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, gd.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, gd.group[0])
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

		b = append(b, gd.currencyNegativePrefix[0])

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
			b = append(b, gd.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, gd.currencyNegativeSuffix...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'gd'
func (gd *gd) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'gd'
func (gd *gd) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, gd.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'gd'
func (gd *gd) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x6d, 0x68}...)
	b = append(b, []byte{0x20}...)
	b = append(b, gd.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'gd'
func (gd *gd) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, gd.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x6d, 0x68}...)
	b = append(b, []byte{0x20}...)
	b = append(b, gd.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'gd'
func (gd *gd) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, gd.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'gd'
func (gd *gd) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, gd.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, gd.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'gd'
func (gd *gd) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, gd.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, gd.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'gd'
func (gd *gd) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, gd.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, gd.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := gd.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
