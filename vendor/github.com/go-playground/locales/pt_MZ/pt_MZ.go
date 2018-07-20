package pt_MZ

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type pt_MZ struct {
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

// New returns a new instance of translator for the 'pt_MZ' locale
func New() locales.Translator {
	return &pt_MZ{
		locale:                 "pt_MZ",
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
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MTn", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
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
		timezones:              map[string]string{"OESZ": "Hora de verão da Europa Oriental", "UYST": "Hora de verão do Uruguai", "HECU": "Hora de verão de Cuba", "AWDT": "Hora de verão da Austrália Ocidental", "JST": "Hora padrão do Japão", "LHDT": "Hora de verão de Lord Howe", "HEPM": "Hora de verão de São Pedro e Miquelão", "HNPMX": "Hora padrão do Pacífico Mexicano", "WAST": "Hora de verão da África Ocidental", "NZST": "Hora padrão da Nova Zelândia", "ACWDT": "Hora de verão da Austrália Central Ocidental", "AKDT": "Hora de verão do Alasca", "EST": "Hora padrão Oriental", "EDT": "Hora de verão Oriental", "HENOMX": "Hora de verão do Noroeste do México", "UYT": "Hora padrão do Uruguai", "NZDT": "Hora de verão da Nova Zelândia", "MYT": "Hora da Malásia", "ECT": "Hora do Equador", "HNNOMX": "Hora padrão do Noroeste do México", "TMST": "Hora de verão do Turquemenistão", "OEZ": "Hora padrão da Europa Oriental", "ART": "Hora padrão da Argentina", "HEPMX": "Hora de verão do Pacífico Mexicano", "SAST": "Hora da África do Sul", "IST": "Hora padrão da Índia", "SRT": "Hora do Suriname", "TMT": "Hora padrão do Turquemenistão", "HNCU": "Hora padrão de Cuba", "HKST": "Hora de verão de Hong Kong", "HNPM": "Hora padrão de São Pedro e Miquelão", "WITA": "Hora da Indonésia Central", "COT": "Hora padrão da Colômbia", "CST": "Hora padrão Central", "WIB": "Hora da Indonésia Ocidental", "WIT": "Hora da Indonésia Oriental", "AKST": "Hora padrão do Alasca", "CAT": "Hora da África Central", "CHADT": "Hora de verão de Chatham", "AST": "Hora padrão do Atlântico", "MDT": "Hora de verão da Montanha", "WAT": "Hora padrão da África Ocidental", "JDT": "Hora de verão do Japão", "BOT": "Hora da Bolívia", "HEOG": "Hora de verão da Gronelândia Ocidental", "WART": "Hora padrão da Argentina Ocidental", "GMT": "Hora de Greenwich", "CLT": "Hora padrão do Chile", "WEZ": "Hora padrão da Europa Ocidental", "SGT": "Hora padrão de Singapura", "HNEG": "Hora padrão da Gronelândia Oriental", "HEEG": "Hora de verão da Gronelândia Oriental", "EAT": "Hora da África Oriental", "COST": "Hora de verão da Colômbia", "PST": "Hora padrão do Pacífico", "AEST": "Hora padrão da Austrália Oriental", "MESZ": "Hora de verão da Europa Central", "LHST": "Hora padrão de Lord Howe", "WARST": "Hora de verão da Argentina Ocidental", "ADT": "Hora de verão do Atlântico", "HNOG": "Hora padrão da Gronelândia Ocidental", "HNT": "Hora padrão da Terra Nova", "ARST": "Hora de verão da Argentina", "ChST": "Hora padrão de Chamorro", "ACWST": "Hora padrão da Austrália Central Ocidental", "MEZ": "Hora padrão da Europa Central", "VET": "Hora da Venezuela", "∅∅∅": "Hora de verão de Brasília", "AWST": "Hora padrão da Austrália Ocidental", "MST": "Hora padrão da Montanha", "BT": "Hora do Butão", "ACDT": "Hora de verão da Austrália Central", "CLST": "Hora de verão do Chile", "PDT": "Hora de verão do Pacífico", "ACST": "Hora padrão da Austrália Central", "CHAST": "Hora padrão de Chatham", "CDT": "Hora de verão Central", "GFT": "Hora da Guiana Francesa", "HKT": "Hora padrão de Hong Kong", "GYT": "Hora da Guiana", "AEDT": "Hora de verão da Austrália Oriental", "WESZ": "Hora de verão da Europa Ocidental", "HAT": "Hora de verão da Terra Nova", "HAST": "Hora padrão do Havai e Aleutas", "HADT": "Hora de verão do Havai e Aleutas"},
	}
}

// Locale returns the current translators string locale
func (pt *pt_MZ) Locale() string {
	return pt.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'pt_MZ'
func (pt *pt_MZ) PluralsCardinal() []locales.PluralRule {
	return pt.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'pt_MZ'
func (pt *pt_MZ) PluralsOrdinal() []locales.PluralRule {
	return pt.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'pt_MZ'
func (pt *pt_MZ) PluralsRange() []locales.PluralRule {
	return pt.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'pt_MZ'
func (pt *pt_MZ) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)
	i := int64(n)

	if i >= 0 && i <= 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'pt_MZ'
func (pt *pt_MZ) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'pt_MZ'
func (pt *pt_MZ) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {

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
func (pt *pt_MZ) MonthAbbreviated(month time.Month) string {
	return pt.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (pt *pt_MZ) MonthsAbbreviated() []string {
	return pt.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (pt *pt_MZ) MonthNarrow(month time.Month) string {
	return pt.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (pt *pt_MZ) MonthsNarrow() []string {
	return pt.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (pt *pt_MZ) MonthWide(month time.Month) string {
	return pt.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (pt *pt_MZ) MonthsWide() []string {
	return pt.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (pt *pt_MZ) WeekdayAbbreviated(weekday time.Weekday) string {
	return pt.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (pt *pt_MZ) WeekdaysAbbreviated() []string {
	return pt.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (pt *pt_MZ) WeekdayNarrow(weekday time.Weekday) string {
	return pt.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (pt *pt_MZ) WeekdaysNarrow() []string {
	return pt.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (pt *pt_MZ) WeekdayShort(weekday time.Weekday) string {
	return pt.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (pt *pt_MZ) WeekdaysShort() []string {
	return pt.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (pt *pt_MZ) WeekdayWide(weekday time.Weekday) string {
	return pt.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (pt *pt_MZ) WeekdaysWide() []string {
	return pt.daysWide
}

// Decimal returns the decimal point of number
func (pt *pt_MZ) Decimal() string {
	return pt.decimal
}

// Group returns the group of number
func (pt *pt_MZ) Group() string {
	return pt.group
}

// Group returns the minus sign of number
func (pt *pt_MZ) Minus() string {
	return pt.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'pt_MZ' and handles both Whole and Real numbers based on 'v'
func (pt *pt_MZ) FmtNumber(num float64, v uint64) string {

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

// FmtPercent returns 'num' with digits/precision of 'v' for 'pt_MZ' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (pt *pt_MZ) FmtPercent(num float64, v uint64) string {
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

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'pt_MZ'
func (pt *pt_MZ) FmtCurrency(num float64, v uint64, currency currency.Type) string {

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

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'pt_MZ'
// in accounting notation.
func (pt *pt_MZ) FmtAccounting(num float64, v uint64, currency currency.Type) string {

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

// FmtDateShort returns the short date representation of 't' for 'pt_MZ'
func (pt *pt_MZ) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'pt_MZ'
func (pt *pt_MZ) FmtDateMedium(t time.Time) string {

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

// FmtDateLong returns the long date representation of 't' for 'pt_MZ'
func (pt *pt_MZ) FmtDateLong(t time.Time) string {

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

// FmtDateFull returns the full date representation of 't' for 'pt_MZ'
func (pt *pt_MZ) FmtDateFull(t time.Time) string {

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

// FmtTimeShort returns the short time representation of 't' for 'pt_MZ'
func (pt *pt_MZ) FmtTimeShort(t time.Time) string {

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

// FmtTimeMedium returns the medium time representation of 't' for 'pt_MZ'
func (pt *pt_MZ) FmtTimeMedium(t time.Time) string {

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

// FmtTimeLong returns the long time representation of 't' for 'pt_MZ'
func (pt *pt_MZ) FmtTimeLong(t time.Time) string {

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

// FmtTimeFull returns the full time representation of 't' for 'pt_MZ'
func (pt *pt_MZ) FmtTimeFull(t time.Time) string {

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
