package ast

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type ast struct {
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
	currencyPositiveSuffix string
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

// New returns a new instance of translator for the 'ast' locale
func New() locales.Translator {
	return &ast{
		locale:                 "ast",
		pluralsCardinal:        []locales.PluralRule{2, 6},
		pluralsOrdinal:         nil,
		pluralsRange:           nil,
		decimal:                ",",
		group:                  ".",
		minus:                  "-",
		percent:                "%",
		perMille:               "‰",
		timeSeparator:          ":",
		inifinity:              "∞",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "A$", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "R$", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CA$", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CN¥", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "€", "FIM", "FJD", "FKP", "FRF", "£", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HK$", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "₪", "₹", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "¥", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "₩", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MX$", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZ$", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "฿", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "NT$", "TZS", "UAH", "UAK", "UGS", "UGX", "$", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "₫", "VNN", "VUV", "WST", "FCFA", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "EC$", "XDR", "XEU", "XFO", "XFU", "CFA", "XPD", "CFPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyPositiveSuffix: " ",
		currencyNegativeSuffix: " ",
		monthsAbbreviated:      []string{"", "xin", "feb", "mar", "abr", "may", "xun", "xnt", "ago", "set", "och", "pay", "avi"},
		monthsNarrow:           []string{"", "X", "F", "M", "A", "M", "X", "X", "A", "S", "O", "P", "A"},
		monthsWide:             []string{"", "de xineru", "de febreru", "de marzu", "d’abril", "de mayu", "de xunu", "de xunetu", "d’agostu", "de setiembre", "d’ochobre", "de payares", "d’avientu"},
		daysAbbreviated:        []string{"dom", "llu", "mar", "mié", "xue", "vie", "sáb"},
		daysNarrow:             []string{"D", "L", "M", "M", "X", "V", "S"},
		daysShort:              []string{"do", "ll", "ma", "mi", "xu", "vi", "sá"},
		daysWide:               []string{"domingu", "llunes", "martes", "miércoles", "xueves", "vienres", "sábadu"},
		periodsAbbreviated:     []string{"AM", "PM"},
		periodsNarrow:          []string{"a", "p"},
		periodsWide:            []string{"de la mañana", "de la tarde"},
		erasAbbreviated:        []string{"e.C.", "d.C."},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"enantes de Cristu", "después de Cristu"},
		timezones:              map[string]string{"PDT": "Hora braniega del Pacíficu norteamericanu", "EDT": "Hora braniega del este norteamericanu", "ACWST": "Hora estándar d’Australia central del oeste", "HKST": "Hora braniega de Ḥong Kong", "SRT": "Hora del Surinam", "CHAST": "Hora estándar de Chatham", "SGT": "Hora estándar de Singapur", "HEPM": "Hora braniega de Saint Pierre y Miquelon", "EAT": "Hora d’África del este", "COT": "Hora estándar de Colombia", "COST": "Hora braniega de Colombia", "AWDT": "Hora braniega d’Australia del oeste", "AST": "Hora estándar del Atlánticu", "AKST": "Hora estándar d’Alaska", "CLST": "Hora braniega de Chile", "CHADT": "Hora braniega de Chatham", "LHDT": "Hora braniega de Lord Howe", "SAST": "Hora de Sudáfrica", "WESZ": "Hora braniega d’Europa Occidental", "JST": "Hora estándar de Xapón", "ACST": "Hora estándar d’Australia central", "HNOG": "Hora estándar de Groenlandia occidental", "MESZ": "Hora braniega d’Europa Central", "LHST": "Hora estándar de Lord Howe", "WART": "Hora estándar occidental d’Arxentina", "TMST": "Hora braniega del Turkmenistán", "AWST": "Hora estándar d’Australia del oeste", "NZDT": "Hora braniega de Nueva Zelanda", "HEOG": "Hora braniega de Groenlandia occidental", "TMT": "Hora estándar del Turkmenistán", "ART": "Hora estándar d’Arxentina", "UYT": "Hora estándar del Uruguái", "NZST": "Hora estándar de Nueva Zelanda", "MYT": "Hora de Malasia", "AKDT": "Hora braniega d’Alaska", "HADT": "Hora braniega de Hawaii-Aleutianes", "ChST": "Hora estándar de Chamorro", "HNPMX": "Hora estándar del Pacíficu de Méxicu", "WAST": "Hora braniega d’África del oeste", "IST": "Hora estándar de la India", "HNT": "Hora estándar de Newfoundland", "CAT": "Hora d’África central", "UYST": "Hora braniega del Uruguái", "HECU": "Hora braniega de Cuba", "PST": "Hora estándar del Pacíficu norteamericanu", "ADT": "Hora braniega del Atlánticu", "EST": "Hora estándar del este norteamericanu", "CLT": "Hora estándar de Chile", "OESZ": "Hora braniega d’Europa del Este", "HNCU": "Hora estándar de Cuba", "HEPMX": "Hora braniega del Pacíficu de Méxicu", "ACWDT": "Hora braniega d’Australia central del oeste", "GYT": "Hora de La Guyana", "GFT": "Hora de La Guyana Francesa", "HENOMX": "Hora braniega del noroeste de Méxicu", "MST": "Hora estándar de les montañes norteamericanes", "MDT": "Hora braniega de les montañes norteamericanes", "HEEG": "Hora braniega de Groenlandia oriental", "HKT": "Hora estándar de Ḥong Kong", "HNEG": "Hora estándar de Groenlandia oriental", "CST": "Hora estándar central norteamericana", "AEST": "Hora estándar d’Australia del este", "WAT": "Hora estándar d’África del oeste", "WEZ": "Hora estándar d’Europa Occidental", "BT": "Hora de Bután", "JDT": "Hora braniega de Xapón", "ACDT": "Hora braniega d’Australia central", "HAT": "Hora braniega de Newfoundland", "HNPM": "Hora estándar de Saint Pierre y Miquelon", "OEZ": "Hora estándar d’Europa del Este", "HAST": "Hora estándar de Hawaii-Aleutianes", "AEDT": "Hora braniega d’Australia del este", "WIB": "Hora d’Indonesia del oeste", "∅∅∅": "hora braniega d’Acre", "VET": "Hora de Venezuela", "CDT": "Hora braniega central norteamericana", "WITA": "Hora d’Indonesia central", "WIT": "Hora d’Indonesia del este", "WARST": "Hora braniega occidental d’Arxentina", "ARST": "Hora braniega d’Arxentina", "GMT": "Hora media de Greenwich", "BOT": "Hora de Bolivia", "ECT": "Hora d’Ecuador", "MEZ": "Hora estándar d’Europa Central", "HNNOMX": "Hora estándar del noroeste de Méxicu"},
	}
}

// Locale returns the current translators string locale
func (ast *ast) Locale() string {
	return ast.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'ast'
func (ast *ast) PluralsCardinal() []locales.PluralRule {
	return ast.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'ast'
func (ast *ast) PluralsOrdinal() []locales.PluralRule {
	return ast.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'ast'
func (ast *ast) PluralsRange() []locales.PluralRule {
	return ast.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'ast'
func (ast *ast) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)
	i := int64(n)

	if i == 1 && v == 0 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'ast'
func (ast *ast) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'ast'
func (ast *ast) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (ast *ast) MonthAbbreviated(month time.Month) string {
	return ast.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (ast *ast) MonthsAbbreviated() []string {
	return ast.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (ast *ast) MonthNarrow(month time.Month) string {
	return ast.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (ast *ast) MonthsNarrow() []string {
	return ast.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (ast *ast) MonthWide(month time.Month) string {
	return ast.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (ast *ast) MonthsWide() []string {
	return ast.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (ast *ast) WeekdayAbbreviated(weekday time.Weekday) string {
	return ast.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (ast *ast) WeekdaysAbbreviated() []string {
	return ast.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (ast *ast) WeekdayNarrow(weekday time.Weekday) string {
	return ast.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (ast *ast) WeekdaysNarrow() []string {
	return ast.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (ast *ast) WeekdayShort(weekday time.Weekday) string {
	return ast.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (ast *ast) WeekdaysShort() []string {
	return ast.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (ast *ast) WeekdayWide(weekday time.Weekday) string {
	return ast.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (ast *ast) WeekdaysWide() []string {
	return ast.daysWide
}

// Decimal returns the decimal point of number
func (ast *ast) Decimal() string {
	return ast.decimal
}

// Group returns the group of number
func (ast *ast) Group() string {
	return ast.group
}

// Group returns the minus sign of number
func (ast *ast) Minus() string {
	return ast.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'ast' and handles both Whole and Real numbers based on 'v'
func (ast *ast) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ast.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, ast.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, ast.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'ast' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (ast *ast) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 3
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ast.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, ast.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, ast.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'ast'
func (ast *ast) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := ast.currencies[currency]
	l := len(s) + len(symbol) + 4 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ast.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, ast.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, ast.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, ast.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	b = append(b, ast.currencyPositiveSuffix...)

	b = append(b, symbol...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'ast'
// in accounting notation.
func (ast *ast) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := ast.currencies[currency]
	l := len(s) + len(symbol) + 4 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ast.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, ast.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {

		b = append(b, ast.minus[0])

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, ast.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, ast.currencyNegativeSuffix...)
		b = append(b, symbol...)
	} else {

		b = append(b, ast.currencyPositiveSuffix...)
		b = append(b, symbol...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'ast'
func (ast *ast) FmtDateShort(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2f}...)
	b = strconv.AppendInt(b, int64(t.Month()), 10)
	b = append(b, []byte{0x2f}...)

	if t.Year() > 9 {
		b = append(b, strconv.Itoa(t.Year())[2:]...)
	} else {
		b = append(b, strconv.Itoa(t.Year())[1:]...)
	}

	return string(b)
}

// FmtDateMedium returns the medium date representation of 't' for 'ast'
func (ast *ast) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, ast.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'ast'
func (ast *ast) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, ast.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20, 0x64, 0x65}...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'ast'
func (ast *ast) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, ast.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, ast.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20, 0x64, 0x65}...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'ast'
func (ast *ast) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, ast.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'ast'
func (ast *ast) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, ast.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, ast.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'ast'
func (ast *ast) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, ast.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, ast.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'ast'
func (ast *ast) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, ast.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, ast.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := ast.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
