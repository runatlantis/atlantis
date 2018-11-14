package pt_ST

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type pt_ST struct {
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

// New returns a new instance of translator for the 'pt_ST' locale
func New() locales.Translator {
	return &pt_ST{
		locale:                 "pt_ST",
		pluralsCardinal:        []locales.PluralRule{2, 6},
		pluralsOrdinal:         []locales.PluralRule{6},
		pluralsRange:           []locales.PluralRule{2, 6},
		decimal:                ",",
		group:                  " ",
		minus:                  "-",
		percent:                "%",
		perMille:               "‰",
		timeSeparator:          ":",
		inifinity:              "∞",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "Db", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyPositiveSuffix: " ",
		currencyNegativePrefix: "(",
		currencyNegativeSuffix: " )",
		monthsAbbreviated:      []string{"", "jan", "fev", "mar", "abr", "mai", "jun", "jul", "ago", "set", "out", "nov", "dez"},
		monthsNarrow:           []string{"", "J", "F", "M", "A", "M", "J", "J", "A", "S", "O", "N", "D"},
		monthsWide:             []string{"", "janeiro", "fevereiro", "março", "abril", "maio", "junho", "julho", "agosto", "setembro", "outubro", "novembro", "dezembro"},
		daysAbbreviated:        []string{"domingo", "segunda", "terça", "quarta", "quinta", "sexta", "sábado"},
		daysNarrow:             []string{"D", "S", "T", "Q", "Q", "S", "S"},
		daysShort:              []string{"dom", "seg", "ter", "qua", "qui", "sex", "sáb"},
		daysWide:               []string{"domingo", "segunda-feira", "terça-feira", "quarta-feira", "quinta-feira", "sexta-feira", "sábado"},
		periodsAbbreviated:     []string{"a.m.", "p.m."},
		periodsNarrow:          []string{"a.m.", "p.m."},
		periodsWide:            []string{"da manhã", "da tarde"},
		erasAbbreviated:        []string{"a.E.C.", "E.C."},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"", ""},
		timezones:              map[string]string{"MDT": "Hora de verão da Montanha", "HKT": "Hora padrão de Hong Kong", "IST": "Hora padrão da Índia", "HAT": "Hora de verão da Terra Nova", "∅∅∅": "Hora de verão de Brasília", "AKST": "Hora padrão do Alasca", "WART": "Hora padrão da Argentina Ocidental", "CLST": "Hora de verão do Chile", "UYST": "Hora de verão do Uruguai", "TMST": "Hora de verão do Turquemenistão", "ChST": "Hora padrão de Chamorro", "CST": "Hora padrão Central", "SAST": "Hora da África do Sul", "EDT": "Hora de verão Oriental", "HNEG": "Hora padrão da Gronelândia Oriental", "WITA": "Hora da Indonésia Central", "CAT": "Hora da África Central", "UYT": "Hora padrão do Uruguai", "PST": "Hora padrão do Pacífico", "WAST": "Hora de verão da África Ocidental", "BT": "Hora do Butão", "MEZ": "Hora padrão da Europa Central", "CLT": "Hora padrão do Chile", "OESZ": "Hora de verão da Europa Oriental", "HADT": "Hora de verão do Havai e Aleutas", "WIT": "Hora da Indonésia Oriental", "AWDT": "Hora de verão da Austrália Ocidental", "JDT": "Hora de verão do Japão", "ACDT": "Hora de verão da Austrália Central", "HNOG": "Hora padrão da Gronelândia Ocidental", "HNT": "Hora padrão da Terra Nova", "HNNOMX": "Hora padrão do Noroeste do México", "HENOMX": "Hora de verão do Noroeste do México", "HEPMX": "Hora de verão do Pacífico Mexicano", "EST": "Hora padrão Oriental", "MESZ": "Hora de verão da Europa Central", "CHAST": "Hora padrão de Chatham", "MYT": "Hora da Malásia", "ACWDT": "Hora de verão da Austrália Central Ocidental", "WARST": "Hora de verão da Argentina Ocidental", "COST": "Hora de verão da Colômbia", "HECU": "Hora de verão de Cuba", "WEZ": "Hora padrão da Europa Ocidental", "WESZ": "Hora de verão da Europa Ocidental", "BOT": "Hora da Bolívia", "HKST": "Hora de verão de Hong Kong", "LHST": "Hora padrão de Lord Howe", "ARST": "Hora de verão da Argentina", "AWST": "Hora padrão da Austrália Ocidental", "NZST": "Hora padrão da Nova Zelândia", "AKDT": "Hora de verão do Alasca", "ART": "Hora padrão da Argentina", "HAST": "Hora padrão do Havai e Aleutas", "CHADT": "Hora de verão de Chatham", "PDT": "Hora de verão do Pacífico", "AEST": "Hora padrão da Austrália Oriental", "AEDT": "Hora de verão da Austrália Oriental", "JST": "Hora padrão do Japão", "ACWST": "Hora padrão da Austrália Central Ocidental", "HNPM": "Hora padrão de São Pedro e Miquelão", "GMT": "Hora de Greenwich", "MST": "Hora padrão da Montanha", "ACST": "Hora padrão da Austrália Central", "WIB": "Hora da Indonésia Ocidental", "TMT": "Hora padrão do Turquemenistão", "COT": "Hora padrão da Colômbia", "GYT": "Hora da Guiana", "CDT": "Hora de verão Central", "GFT": "Hora da Guiana Francesa", "ADT": "Hora de verão do Atlântico", "NZDT": "Hora de verão da Nova Zelândia", "HEOG": "Hora de verão da Gronelândia Ocidental", "HEPM": "Hora de verão de São Pedro e Miquelão", "SRT": "Hora do Suriname", "EAT": "Hora da África Oriental", "OEZ": "Hora padrão da Europa Oriental", "HNPMX": "Hora padrão do Pacífico Mexicano", "WAT": "Hora padrão da África Ocidental", "ECT": "Hora do Equador", "HNCU": "Hora padrão de Cuba", "AST": "Hora padrão do Atlântico", "SGT": "Hora padrão de Singapura", "HEEG": "Hora de verão da Gronelândia Oriental", "LHDT": "Hora de verão de Lord Howe", "VET": "Hora da Venezuela"},
	}
}

// Locale returns the current translators string locale
func (pt *pt_ST) Locale() string {
	return pt.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'pt_ST'
func (pt *pt_ST) PluralsCardinal() []locales.PluralRule {
	return pt.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'pt_ST'
func (pt *pt_ST) PluralsOrdinal() []locales.PluralRule {
	return pt.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'pt_ST'
func (pt *pt_ST) PluralsRange() []locales.PluralRule {
	return pt.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'pt_ST'
func (pt *pt_ST) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)
	i := int64(n)

	if i >= 0 && i <= 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'pt_ST'
func (pt *pt_ST) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'pt_ST'
func (pt *pt_ST) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {

	start := pt.CardinalPluralRule(num1, v1)
	end := pt.CardinalPluralRule(num2, v2)

	if start == locales.PluralRuleOne && end == locales.PluralRuleOne {
		return locales.PluralRuleOne
	} else if start == locales.PluralRuleOne && end == locales.PluralRuleOther {
		return locales.PluralRuleOther
	}

	return locales.PluralRuleOther

}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (pt *pt_ST) MonthAbbreviated(month time.Month) string {
	return pt.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (pt *pt_ST) MonthsAbbreviated() []string {
	return pt.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (pt *pt_ST) MonthNarrow(month time.Month) string {
	return pt.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (pt *pt_ST) MonthsNarrow() []string {
	return pt.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (pt *pt_ST) MonthWide(month time.Month) string {
	return pt.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (pt *pt_ST) MonthsWide() []string {
	return pt.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (pt *pt_ST) WeekdayAbbreviated(weekday time.Weekday) string {
	return pt.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (pt *pt_ST) WeekdaysAbbreviated() []string {
	return pt.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (pt *pt_ST) WeekdayNarrow(weekday time.Weekday) string {
	return pt.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (pt *pt_ST) WeekdaysNarrow() []string {
	return pt.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (pt *pt_ST) WeekdayShort(weekday time.Weekday) string {
	return pt.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (pt *pt_ST) WeekdaysShort() []string {
	return pt.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (pt *pt_ST) WeekdayWide(weekday time.Weekday) string {
	return pt.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (pt *pt_ST) WeekdaysWide() []string {
	return pt.daysWide
}

// Decimal returns the decimal point of number
func (pt *pt_ST) Decimal() string {
	return pt.decimal
}

// Group returns the group of number
func (pt *pt_ST) Group() string {
	return pt.group
}

// Group returns the minus sign of number
func (pt *pt_ST) Minus() string {
	return pt.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'pt_ST' and handles both Whole and Real numbers based on 'v'
func (pt *pt_ST) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, pt.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(pt.group) - 1; j >= 0; j-- {
					b = append(b, pt.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, pt.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'pt_ST' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (pt *pt_ST) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 3
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, pt.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, pt.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, pt.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'pt_ST'
func (pt *pt_ST) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := pt.currencies[currency]
	l := len(s) + len(symbol) + 4 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, pt.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(pt.group) - 1; j >= 0; j-- {
					b = append(b, pt.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, pt.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, pt.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	b = append(b, pt.currencyPositiveSuffix...)

	b = append(b, symbol...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'pt_ST'
// in accounting notation.
func (pt *pt_ST) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := pt.currencies[currency]
	l := len(s) + len(symbol) + 6 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, pt.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(pt.group) - 1; j >= 0; j-- {
					b = append(b, pt.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {

		b = append(b, pt.currencyNegativePrefix[0])

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, pt.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, pt.currencyNegativeSuffix...)
		b = append(b, symbol...)
	} else {

		b = append(b, pt.currencyPositiveSuffix...)
		b = append(b, symbol...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'pt_ST'
func (pt *pt_ST) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'pt_ST'
func (pt *pt_ST) FmtDateMedium(t time.Time) string {

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

// FmtDateLong returns the long date representation of 't' for 'pt_ST'
func (pt *pt_ST) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20, 0x64, 0x65}...)
	b = append(b, []byte{0x20}...)
	b = append(b, pt.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20, 0x64, 0x65}...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'pt_ST'
func (pt *pt_ST) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, pt.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20, 0x64, 0x65}...)
	b = append(b, []byte{0x20}...)
	b = append(b, pt.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20, 0x64, 0x65}...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'pt_ST'
func (pt *pt_ST) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, pt.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'pt_ST'
func (pt *pt_ST) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, pt.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, pt.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'pt_ST'
func (pt *pt_ST) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, pt.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, pt.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'pt_ST'
func (pt *pt_ST) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, pt.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, pt.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := pt.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
