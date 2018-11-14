package cy_GB

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type cy_GB struct {
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

// New returns a new instance of translator for the 'cy_GB' locale
func New() locales.Translator {
	return &cy_GB{
		locale:                 "cy_GB",
		pluralsCardinal:        []locales.PluralRule{1, 2, 3, 4, 5, 6},
		pluralsOrdinal:         []locales.PluralRule{1, 2, 3, 4, 5, 6},
		pluralsRange:           []locales.PluralRule{2, 3, 4, 5, 6},
		decimal:                ".",
		group:                  ",",
		minus:                  "-",
		percent:                "%",
		perMille:               "‰",
		timeSeparator:          ":",
		inifinity:              "∞",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyNegativePrefix: "(",
		currencyNegativeSuffix: ")",
		monthsAbbreviated:      []string{"", "Ion", "Chwef", "Maw", "Ebrill", "Mai", "Meh", "Gorff", "Awst", "Medi", "Hyd", "Tach", "Rhag"},
		monthsNarrow:           []string{"", "I", "Ch", "M", "E", "M", "M", "G", "A", "M", "H", "T", "Rh"},
		monthsWide:             []string{"", "Ionawr", "Chwefror", "Mawrth", "Ebrill", "Mai", "Mehefin", "Gorffennaf", "Awst", "Medi", "Hydref", "Tachwedd", "Rhagfyr"},
		daysAbbreviated:        []string{"Sul", "Llun", "Maw", "Mer", "Iau", "Gwen", "Sad"},
		daysNarrow:             []string{"S", "Ll", "M", "M", "I", "G", "S"},
		daysShort:              []string{"Su", "Ll", "Ma", "Me", "Ia", "Gw", "Sa"},
		daysWide:               []string{"Dydd Sul", "Dydd Llun", "Dydd Mawrth", "Dydd Mercher", "Dydd Iau", "Dydd Gwener", "Dydd Sadwrn"},
		periodsAbbreviated:     []string{"yb", "yh"},
		periodsNarrow:          []string{"b", "h"},
		periodsWide:            []string{"yb", "yh"},
		erasAbbreviated:        []string{"CC", "OC"},
		erasNarrow:             []string{"C", "O"},
		erasWide:               []string{"Cyn Crist", "Oed Crist"},
		timezones:              map[string]string{"MDT": "Amser Haf Mynyddoedd Gogledd America", "HEOG": "Amser Haf Gorllewin yr Ynys Las", "HNT": "Amser Safonol Newfoundland", "OESZ": "Amser Haf Dwyrain Ewrop", "HECU": "Amser Haf Cuba", "HEEG": "Amser Haf Dwyrain yr Ynys Las", "ART": "Amser Safonol Ariannin", "LHDT": "Amser Haf yr Arglwydd Howe", "EAT": "Amser Dwyrain Affrica", "HADT": "Amser Haf Hawaii-Aleutian", "CST": "Amser Safonol Canolbarth Gogledd America", "WAST": "Amser Haf Gorllewin Affrica", "SAST": "Amser Safonol De Affrica", "MYT": "Amser Malaysia", "HKST": "Amser Haf Hong Kong", "COT": "Amser Safonol Colombia", "GYT": "Amser Guyana", "CHAST": "Amser Safonol Chatham", "CHADT": "Amser Haf Chatham", "MEZ": "Amser Safonol Canolbarth Ewrop", "LHST": "Amser Safonol yr Arglwydd Howe", "WART": "Amser Safonol Gorllewin Ariannin", "UYST": "Amser Haf Uruguay", "CLT": "Amser Safonol Chile", "HAST": "Amser Safonol Hawaii-Aleutian", "ACWDT": "Amser Haf Canolbarth Gorllewin Awstralia", "HKT": "Amser Safonol Hong Kong", "WARST": "Amser Haf Gorllewin Ariannin", "VET": "Amser Venezuela", "SRT": "Amser Suriname", "IST": "Amser India", "HNPM": "Amser Safonol Saint-Pierre-et-Miquelon", "ChST": "Amser Chamorro", "PDT": "Amser Haf Cefnfor Tawel Gogledd America", "AWDT": "Amser Haf Gorllewin Awstralia", "HEPMX": "Amser Haf Pasiffig Mecsico", "ACWST": "Amser Safonol Canolbarth Gorllewin Awstralia", "MESZ": "Amser Haf Canolbarth Ewrop", "HNCU": "Amser Safonol Cuba", "ADT": "Amser Haf Cefnfor yr Iwerydd", "WAT": "Amser Safonol Gorllewin Affrica", "HEPM": "Amser Haf Saint-Pierre-et-Miquelon", "HNNOMX": "Amser Safonol Gogledd Orllewin Mecsico", "OEZ": "Amser Safonol Dwyrain Ewrop", "HENOMX": "Amser Haf Gogledd Orllewin Mecsico", "AWST": "Amser Safonol Gorllewin Awstralia", "WESZ": "Amser Haf Gorllewin Ewrop", "∅∅∅": "∅∅∅", "GFT": "Amser Guyane Ffrengig", "HNOG": "Amser Safonol Gorllewin yr Ynys Las", "CDT": "Amser Haf Canolbarth Gogledd America", "SGT": "Amser Singapore", "HAT": "Amser Haf Newfoundland", "COST": "Amser Haf Colombia", "JDT": "Amser Haf Siapan", "AKDT": "Amser Haf Alaska", "WITA": "Amser Canolbarth Indonesia", "CLST": "Amser Haf Chile", "WIT": "Amser Dwyrain Indonesia", "PST": "Amser Safonol Cefnfor Tawel Gogledd America", "BOT": "Amser Bolivia", "TMST": "Amser Haf Tyrcmenistan", "GMT": "Amser Safonol Greenwich", "HNPMX": "Amser Safonol Pasiffig Mecsico", "AEDT": "Amser Haf Dwyrain Awstralia", "AST": "Amser Safonol Cefnfor yr Iwerydd", "CAT": "Amser Canolbarth Affrica", "WEZ": "Amser Safonol Gorllewin Ewrop", "BT": "Amser Bhutan", "ACDT": "Amser Haf Canolbarth Awstralia", "HNEG": "Amser Safonol Dwyrain yr Ynys Las", "EST": "Amser Safonol Dwyrain Gogledd America", "MST": "Amser Safonol Mynyddoedd Gogledd America", "WIB": "Amser Gorllewin Indonesia", "AKST": "Amser Safonol Alaska", "ARST": "Amser Haf Ariannin", "UYT": "Amser Safonol Uruguay", "JST": "Amser Safonol Siapan", "ACST": "Amser Safonol Canolbarth Awstralia", "TMT": "Amser Safonol Tyrcmenistan", "AEST": "Amser Safonol Dwyrain Awstralia", "NZST": "Amser Safonol Seland Newydd", "NZDT": "Amser Haf Seland Newydd", "ECT": "Amser Ecuador", "EDT": "Amser Haf Dwyrain Gogledd America"},
	}
}

// Locale returns the current translators string locale
func (cy *cy_GB) Locale() string {
	return cy.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'cy_GB'
func (cy *cy_GB) PluralsCardinal() []locales.PluralRule {
	return cy.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'cy_GB'
func (cy *cy_GB) PluralsOrdinal() []locales.PluralRule {
	return cy.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'cy_GB'
func (cy *cy_GB) PluralsRange() []locales.PluralRule {
	return cy.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'cy_GB'
func (cy *cy_GB) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 0 {
		return locales.PluralRuleZero
	} else if n == 1 {
		return locales.PluralRuleOne
	} else if n == 2 {
		return locales.PluralRuleTwo
	} else if n == 3 {
		return locales.PluralRuleFew
	} else if n == 6 {
		return locales.PluralRuleMany
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'cy_GB'
func (cy *cy_GB) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 0 || n == 7 || n == 8 || n == 9 {
		return locales.PluralRuleZero
	} else if n == 1 {
		return locales.PluralRuleOne
	} else if n == 2 {
		return locales.PluralRuleTwo
	} else if n == 3 || n == 4 {
		return locales.PluralRuleFew
	} else if n == 5 || n == 6 {
		return locales.PluralRuleMany
	}

	return locales.PluralRuleOther
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'cy_GB'
func (cy *cy_GB) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {

	start := cy.CardinalPluralRule(num1, v1)
	end := cy.CardinalPluralRule(num2, v2)

	if start == locales.PluralRuleZero && end == locales.PluralRuleOne {
		return locales.PluralRuleOne
	} else if start == locales.PluralRuleZero && end == locales.PluralRuleTwo {
		return locales.PluralRuleTwo
	} else if start == locales.PluralRuleZero && end == locales.PluralRuleFew {
		return locales.PluralRuleFew
	} else if start == locales.PluralRuleZero && end == locales.PluralRuleMany {
		return locales.PluralRuleMany
	} else if start == locales.PluralRuleZero && end == locales.PluralRuleOther {
		return locales.PluralRuleOther
	} else if start == locales.PluralRuleOne && end == locales.PluralRuleTwo {
		return locales.PluralRuleTwo
	} else if start == locales.PluralRuleOne && end == locales.PluralRuleFew {
		return locales.PluralRuleFew
	} else if start == locales.PluralRuleOne && end == locales.PluralRuleMany {
		return locales.PluralRuleMany
	} else if start == locales.PluralRuleOne && end == locales.PluralRuleOther {
		return locales.PluralRuleOther
	} else if start == locales.PluralRuleTwo && end == locales.PluralRuleFew {
		return locales.PluralRuleFew
	} else if start == locales.PluralRuleTwo && end == locales.PluralRuleMany {
		return locales.PluralRuleMany
	} else if start == locales.PluralRuleTwo && end == locales.PluralRuleOther {
		return locales.PluralRuleOther
	} else if start == locales.PluralRuleFew && end == locales.PluralRuleMany {
		return locales.PluralRuleMany
	} else if start == locales.PluralRuleFew && end == locales.PluralRuleOther {
		return locales.PluralRuleOther
	} else if start == locales.PluralRuleMany && end == locales.PluralRuleOther {
		return locales.PluralRuleOther
	} else if start == locales.PluralRuleOther && end == locales.PluralRuleOne {
		return locales.PluralRuleOne
	} else if start == locales.PluralRuleOther && end == locales.PluralRuleTwo {
		return locales.PluralRuleTwo
	} else if start == locales.PluralRuleOther && end == locales.PluralRuleFew {
		return locales.PluralRuleFew
	} else if start == locales.PluralRuleOther && end == locales.PluralRuleMany {
		return locales.PluralRuleMany
	}

	return locales.PluralRuleOther

}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (cy *cy_GB) MonthAbbreviated(month time.Month) string {
	return cy.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (cy *cy_GB) MonthsAbbreviated() []string {
	return cy.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (cy *cy_GB) MonthNarrow(month time.Month) string {
	return cy.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (cy *cy_GB) MonthsNarrow() []string {
	return cy.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (cy *cy_GB) MonthWide(month time.Month) string {
	return cy.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (cy *cy_GB) MonthsWide() []string {
	return cy.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (cy *cy_GB) WeekdayAbbreviated(weekday time.Weekday) string {
	return cy.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (cy *cy_GB) WeekdaysAbbreviated() []string {
	return cy.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (cy *cy_GB) WeekdayNarrow(weekday time.Weekday) string {
	return cy.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (cy *cy_GB) WeekdaysNarrow() []string {
	return cy.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (cy *cy_GB) WeekdayShort(weekday time.Weekday) string {
	return cy.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (cy *cy_GB) WeekdaysShort() []string {
	return cy.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (cy *cy_GB) WeekdayWide(weekday time.Weekday) string {
	return cy.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (cy *cy_GB) WeekdaysWide() []string {
	return cy.daysWide
}

// Decimal returns the decimal point of number
func (cy *cy_GB) Decimal() string {
	return cy.decimal
}

// Group returns the group of number
func (cy *cy_GB) Group() string {
	return cy.group
}

// Group returns the minus sign of number
func (cy *cy_GB) Minus() string {
	return cy.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'cy_GB' and handles both Whole and Real numbers based on 'v'
func (cy *cy_GB) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, cy.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, cy.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, cy.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'cy_GB' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (cy *cy_GB) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 3
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, cy.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, cy.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, cy.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'cy_GB'
func (cy *cy_GB) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := cy.currencies[currency]
	l := len(s) + len(symbol) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, cy.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, cy.group[0])
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
		b = append(b, cy.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, cy.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'cy_GB'
// in accounting notation.
func (cy *cy_GB) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := cy.currencies[currency]
	l := len(s) + len(symbol) + 4 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, cy.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, cy.group[0])
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

		b = append(b, cy.currencyNegativePrefix[0])

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
			b = append(b, cy.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, cy.currencyNegativeSuffix...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'cy_GB'
func (cy *cy_GB) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'cy_GB'
func (cy *cy_GB) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, cy.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'cy_GB'
func (cy *cy_GB) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, cy.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'cy_GB'
func (cy *cy_GB) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, cy.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, cy.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'cy_GB'
func (cy *cy_GB) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, cy.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'cy_GB'
func (cy *cy_GB) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, cy.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, cy.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'cy_GB'
func (cy *cy_GB) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, cy.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, cy.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'cy_GB'
func (cy *cy_GB) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, cy.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, cy.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := cy.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
