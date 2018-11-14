package ga_IE

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type ga_IE struct {
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

// New returns a new instance of translator for the 'ga_IE' locale
func New() locales.Translator {
	return &ga_IE{
		locale:                 "ga_IE",
		pluralsCardinal:        []locales.PluralRule{2, 3, 4, 5, 6},
		pluralsOrdinal:         []locales.PluralRule{2, 6},
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
		monthsAbbreviated:      []string{"", "Ean", "Feabh", "Márta", "Aib", "Beal", "Meith", "Iúil", "Lún", "MFómh", "DFómh", "Samh", "Noll"},
		monthsNarrow:           []string{"", "E", "F", "M", "A", "B", "M", "I", "L", "M", "D", "S", "N"},
		monthsWide:             []string{"", "Eanáir", "Feabhra", "Márta", "Aibreán", "Bealtaine", "Meitheamh", "Iúil", "Lúnasa", "Meán Fómhair", "Deireadh Fómhair", "Samhain", "Nollaig"},
		daysAbbreviated:        []string{"Domh", "Luan", "Máirt", "Céad", "Déar", "Aoine", "Sath"},
		daysNarrow:             []string{"D", "L", "M", "C", "D", "A", "S"},
		daysShort:              []string{"Do", "Lu", "Má", "Cé", "Dé", "Ao", "Sa"},
		daysWide:               []string{"Dé Domhnaigh", "Dé Luain", "Dé Máirt", "Dé Céadaoin", "Déardaoin", "Dé hAoine", "Dé Sathairn"},
		periodsAbbreviated:     []string{"r.n.", "i.n."},
		periodsNarrow:          []string{"a", "p"},
		periodsWide:            []string{"r.n.", "i.n."},
		erasAbbreviated:        []string{"RC", "AD"},
		erasNarrow:             []string{"RC", "AD"},
		erasWide:               []string{"Roimh Chríost", "Anno Domini"},
		timezones:              map[string]string{"SRT": "Am Shuranam", "MDT": "Am Samhraidh Mhacao", "HADT": "Am Samhraidh Haváí-Ailiúit", "BOT": "Am na Bolaive", "ECT": "Am Eacuadór", "ACST": "Am Caighdeánach Lár na hAstráile", "IST": "Am Caighdeánach na hIndia", "WIB": "Am Iarthar na hIndinéise", "AKDT": "Am Samhraidh Alasca", "HNEG": "Am Caighdeánach Oirthear na Graonlainne", "AWDT": "Am Samhraidh Iarthar na hAstráile", "HENOMX": "Am Samhraidh Iarthuaisceart Mheicsiceo", "CLST": "Am Samhraidh na Sile", "TMST": "Am Samhraidh na Tuircméanastáine", "PST": "Am Caighdeánach an Aigéin Chiúin", "WAT": "Am Caighdeánach Iarthar na hAfraice", "NZDT": "Am Samhraidh na Nua-Shéalainne", "WITA": "Am Lár na hIndinéise", "HEOG": "Am Samhraidh Iarthar na Graonlainne", "HEPMX": "Am Samhraidh Meicsiceach an Aigéin Chiúin", "LHST": "Am Caighdeánach Lord Howe", "WIT": "Am Oirthear na hIndinéise", "ARST": "Am Samhraidh na hAirgintíne", "ACWST": "Am Caighdeánach Mheániarthar na hAstráile", "∅∅∅": "Am Samhraidh na nAsór", "OESZ": "Am Samhraidh Oirthear na hEorpa", "WESZ": "Am Samhraidh Iarthar na hEorpa", "SAST": "Am Caighdeánach na hAfraice Theas", "JDT": "Am Samhraidh na Seapáine", "LHDT": "Am Samhraidh Lord Howe", "VET": "Am Veiniséala", "HECU": "Am Samhraidh Chúba", "CST": "Am Caighdeánach Lárnach", "ACDT": "Am Samhraidh Lár na hAstráile", "EST": "Am Caighdeánach an Oirthir", "HKST": "Am Samhraidh Hong Cong", "EAT": "Am Oirthear na hAfraice", "TMT": "Am Caighdeánach na Tuircméanastáine", "AEDT": "Am Samhraidh Oirthear na hAstráile", "AST": "Am Caighdeánach an Atlantaigh", "WAST": "Am Samhraidh Iarthar na hAfraice", "GFT": "Am Ghuáin na Fraince", "HEEG": "Am Samhraidh Oirthear na Graonlainne", "HNNOMX": "Am Caighdeánach Iarthuaisceart Mheicsiceo", "COST": "Am Samhraidh na Colóime", "CHAST": "Am Caighdeánach Chatham", "HAT": "Am Samhraidh Thalamh an Éisc", "MEZ": "Am Caighdeánach Lár na hEorpa", "CAT": "Am Lár na hAfraice", "HNPMX": "Am Caighdeánach Meicsiceach an Aigéin Chiúin", "EDT": "Am Samhraidh an Oirthir", "HKT": "Am Caighdeánach Hong Cong", "ADT": "Am Samhraidh an Atlantaigh", "HNOG": "Am Caighdeánach Iarthar na Graonlainne", "HNPM": "Am Caighdeánach Saint-Pierre-et-Miquelon", "OEZ": "Am Caighdeánach Oirthear na hEorpa", "AKST": "Am Caighdeánach Alasca", "CLT": "Am Caighdeánach na Sile", "HAST": "Am Caighdeánach Haváí-Ailiúit", "ART": "Am Caighdeánach na hAirgintíne", "GYT": "Am na Guáine", "UYT": "Am Caighdeánach Uragua", "CHADT": "Am Samhraidh Chatham", "WEZ": "Am Caighdeánach Iarthar na hEorpa", "MESZ": "Am Samhraidh Lár na hEorpa", "PDT": "Am Samhraidh an Aigéin Chiúin", "AEST": "Am Caighdeánach Oirthear na hAstráile", "NZST": "Am Caighdeánach na Nua-Shéalainne", "MYT": "Am na Malaeisia", "BT": "Am na Bútáine", "SGT": "Am Caighdeánach Shingeapór", "WARST": "Am Samhraidh Iarthar na hAirgintíne", "HNT": "Am Caighdeánach Thalamh an Éisc", "UYST": "Am Samhraidh Uragua", "ChST": "Am Caighdeánach Seamórach", "HNCU": "Am Caighdeánach Chúba", "CDT": "Am Samhraidh Lárnach", "JST": "Am Caighdeánach na Seapáine", "ACWDT": "Am Samhraidh Mheániarthar na hAstráile", "HEPM": "Am Samhraidh Saint-Pierre-et-Miquelon", "MST": "Am Caighdeánach Mhacao", "COT": "Am Caighdeánach na Colóime", "GMT": "Meán-Am Greenwich", "AWST": "Am Caighdeánach Iarthar na hAstráile", "WART": "Am Caighdeánach Iarthar na hAirgintíne"},
	}
}

// Locale returns the current translators string locale
func (ga *ga_IE) Locale() string {
	return ga.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'ga_IE'
func (ga *ga_IE) PluralsCardinal() []locales.PluralRule {
	return ga.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'ga_IE'
func (ga *ga_IE) PluralsOrdinal() []locales.PluralRule {
	return ga.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'ga_IE'
func (ga *ga_IE) PluralsRange() []locales.PluralRule {
	return ga.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'ga_IE'
func (ga *ga_IE) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 {
		return locales.PluralRuleOne
	} else if n == 2 {
		return locales.PluralRuleTwo
	} else if n >= 3 && n <= 6 {
		return locales.PluralRuleFew
	} else if n >= 7 && n <= 10 {
		return locales.PluralRuleMany
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'ga_IE'
func (ga *ga_IE) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'ga_IE'
func (ga *ga_IE) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {

	start := ga.CardinalPluralRule(num1, v1)
	end := ga.CardinalPluralRule(num2, v2)

	if start == locales.PluralRuleOne && end == locales.PluralRuleTwo {
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
	} else if start == locales.PluralRuleFew && end == locales.PluralRuleFew {
		return locales.PluralRuleFew
	} else if start == locales.PluralRuleFew && end == locales.PluralRuleMany {
		return locales.PluralRuleMany
	} else if start == locales.PluralRuleFew && end == locales.PluralRuleOther {
		return locales.PluralRuleOther
	} else if start == locales.PluralRuleMany && end == locales.PluralRuleMany {
		return locales.PluralRuleMany
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
func (ga *ga_IE) MonthAbbreviated(month time.Month) string {
	return ga.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (ga *ga_IE) MonthsAbbreviated() []string {
	return ga.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (ga *ga_IE) MonthNarrow(month time.Month) string {
	return ga.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (ga *ga_IE) MonthsNarrow() []string {
	return ga.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (ga *ga_IE) MonthWide(month time.Month) string {
	return ga.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (ga *ga_IE) MonthsWide() []string {
	return ga.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (ga *ga_IE) WeekdayAbbreviated(weekday time.Weekday) string {
	return ga.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (ga *ga_IE) WeekdaysAbbreviated() []string {
	return ga.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (ga *ga_IE) WeekdayNarrow(weekday time.Weekday) string {
	return ga.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (ga *ga_IE) WeekdaysNarrow() []string {
	return ga.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (ga *ga_IE) WeekdayShort(weekday time.Weekday) string {
	return ga.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (ga *ga_IE) WeekdaysShort() []string {
	return ga.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (ga *ga_IE) WeekdayWide(weekday time.Weekday) string {
	return ga.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (ga *ga_IE) WeekdaysWide() []string {
	return ga.daysWide
}

// Decimal returns the decimal point of number
func (ga *ga_IE) Decimal() string {
	return ga.decimal
}

// Group returns the group of number
func (ga *ga_IE) Group() string {
	return ga.group
}

// Group returns the minus sign of number
func (ga *ga_IE) Minus() string {
	return ga.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'ga_IE' and handles both Whole and Real numbers based on 'v'
func (ga *ga_IE) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ga.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, ga.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, ga.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'ga_IE' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (ga *ga_IE) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 3
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ga.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, ga.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, ga.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'ga_IE'
func (ga *ga_IE) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := ga.currencies[currency]
	l := len(s) + len(symbol) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ga.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, ga.group[0])
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
		b = append(b, ga.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, ga.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'ga_IE'
// in accounting notation.
func (ga *ga_IE) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := ga.currencies[currency]
	l := len(s) + len(symbol) + 4 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ga.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, ga.group[0])
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

		b = append(b, ga.currencyNegativePrefix[0])

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
			b = append(b, ga.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, ga.currencyNegativeSuffix...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'ga_IE'
func (ga *ga_IE) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'ga_IE'
func (ga *ga_IE) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, ga.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'ga_IE'
func (ga *ga_IE) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, ga.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'ga_IE'
func (ga *ga_IE) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, ga.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, ga.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'ga_IE'
func (ga *ga_IE) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, ga.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'ga_IE'
func (ga *ga_IE) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, ga.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, ga.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'ga_IE'
func (ga *ga_IE) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, ga.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, ga.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'ga_IE'
func (ga *ga_IE) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, ga.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, ga.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := ga.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
