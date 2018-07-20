package lt_LT

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type lt_LT struct {
	locale                 string
	pluralsCardinal        []locales.PluralRule
	pluralsOrdinal         []locales.PluralRule
	pluralsRange           []locales.PluralRule
	decimal                string
	group                  string
	minus                  string
	percent                string
	percentSuffix          string
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

// New returns a new instance of translator for the 'lt_LT' locale
func New() locales.Translator {
	return &lt_LT{
		locale:                 "lt_LT",
		pluralsCardinal:        []locales.PluralRule{2, 4, 5, 6},
		pluralsOrdinal:         []locales.PluralRule{6},
		pluralsRange:           []locales.PluralRule{2, 4, 5, 6},
		decimal:                ",",
		group:                  " ",
		minus:                  "−",
		percent:                "%",
		perMille:               "‰",
		timeSeparator:          ":",
		inifinity:              "∞",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		percentSuffix:          " ",
		currencyPositiveSuffix: " ",
		currencyNegativeSuffix: " ",
		monthsAbbreviated:      []string{"", "saus.", "vas.", "kov.", "bal.", "geg.", "birž.", "liep.", "rugp.", "rugs.", "spal.", "lapkr.", "gruod."},
		monthsNarrow:           []string{"", "S", "V", "K", "B", "G", "B", "L", "R", "R", "S", "L", "G"},
		monthsWide:             []string{"", "sausio", "vasario", "kovo", "balandžio", "gegužės", "birželio", "liepos", "rugpjūčio", "rugsėjo", "spalio", "lapkričio", "gruodžio"},
		daysAbbreviated:        []string{"sk", "pr", "an", "tr", "kt", "pn", "št"},
		daysNarrow:             []string{"S", "P", "A", "T", "K", "P", "Š"},
		daysShort:              []string{"Sk", "Pr", "An", "Tr", "Kt", "Pn", "Št"},
		daysWide:               []string{"sekmadienis", "pirmadienis", "antradienis", "trečiadienis", "ketvirtadienis", "penktadienis", "šeštadienis"},
		periodsAbbreviated:     []string{"priešpiet", "popiet"},
		periodsNarrow:          []string{"pr. p.", "pop."},
		periodsWide:            []string{"priešpiet", "popiet"},
		erasAbbreviated:        []string{"pr. Kr.", "po Kr."},
		erasNarrow:             []string{"pr. Kr.", "po Kr."},
		erasWide:               []string{"prieš Kristų", "po Kristaus"},
		timezones:              map[string]string{"WIB": "Vakarų Indonezijos laikas", "NZST": "Naujosios Zelandijos žiemos laikas", "HNOG": "Grenlandijos vakarų žiemos laikas", "HEPM": "Sen Pjero ir Mikelono vasaros laikas", "HAST": "Havajų–Aleutų žiemos laikas", "GMT": "Grinvičo laikas", "UYST": "Urugvajaus vasaros laikas", "AKDT": "Aliaskos vasaros laikas", "ACWDT": "Centrinės vakarų Australijos vasaros laikas", "MEZ": "Vidurio Europos žiemos laikas", "CAT": "Centrinės Afrikos laikas", "HECU": "Kubos vasaros laikas", "HNPMX": "Meksikos Ramiojo vandenyno žiemos laikas", "AEDT": "Rytų Australijos vasaros laikas", "ECT": "Ekvadoro laikas", "GYT": "Gajanos laikas", "AEST": "Rytų Australijos žiemos laikas", "SAST": "Pietų Afrikos laikas", "WAT": "Vakarų Afrikos žiemos laikas", "MYT": "Malaizijos laikas", "JDT": "Japonijos vasaros laikas", "EST": "Šiaurės Amerikos rytų žiemos laikas", "WART": "Vakarų Argentinos žiemos laikas", "ADT": "Atlanto vasaros laikas", "ACST": "Centrinės Australijos žiemos laikas", "HAT": "Niufaundlendo vasaros laikas", "WITA": "Centrinės Indonezijos laikas", "TMST": "Turkmėnistano vasaros laikas", "OEZ": "Rytų Europos žiemos laikas", "CHAST": "Čatamo žiemos laikas", "HNCU": "Kubos žiemos laikas", "MDT": "Šiaurės Amerikos kalnų vasaros laikas", "WEZ": "Vakarų Europos žiemos laikas", "∅∅∅": "Ako vasaros laikas", "WARST": "Vakarų Argentinos vasaros laikas", "HNT": "Niufaundlendo žiemos laikas", "WIT": "Rytų Indonezijos laikas", "CHADT": "Čatamo vasaros laikas", "MST": "Šiaurės Amerikos kalnų žiemos laikas", "WESZ": "Vakarų Europos vasaros laikas", "HENOMX": "Šiaurės Vakarų Meksikos vasaros laikas", "CLT": "Čilės žiemos laikas", "HEPMX": "Meksikos Ramiojo vandenyno vasaros laikas", "ACDT": "Centrinės Australijos vasaros laikas", "HEOG": "Grenlandijos vakarų vasaros laikas", "HNPM": "Sen Pjero ir Mikelono žiemos laikas", "BOT": "Bolivijos laikas", "EDT": "Šiaurės Amerikos rytų vasaros laikas", "IST": "Indijos laikas", "CLST": "Čilės vasaros laikas", "JST": "Japonijos žiemos laikas", "SRT": "Surinamo laikas", "WAST": "Vakarų Afrikos vasaros laikas", "HNEG": "Grenlandijos rytų žiemos laikas", "MESZ": "Vidurio Europos vasaros laikas", "HNNOMX": "Šiaurės Vakarų Meksikos žiemos laikas", "TMT": "Turkmėnistano žiemos laikas", "ARST": "Argentinos vasaros laikas", "UYT": "Urugvajaus žiemos laikas", "HKST": "Honkongo vasaros laikas", "CST": "Šiaurės Amerikos centro žiemos laikas", "AWST": "Vakarų Australijos žiemos laikas", "PDT": "Šiaurės Amerikos Ramiojo vandenyno vasaros laikas", "BT": "Butano laikas", "NZDT": "Naujosios Zelandijos vasaros laikas", "AKST": "Aliaskos žiemos laikas", "ACWST": "Centrinės vakarų Australijos žiemos laikas", "LHST": "Lordo Hau žiemos laikas", "COST": "Kolumbijos vasaros laikas", "ChST": "Čamoro laikas", "SGT": "Singapūro laikas", "HEEG": "Grenlandijos rytų vasaros laikas", "HKT": "Honkongo žiemos laikas", "LHDT": "Lordo Hau vasaros laikas", "VET": "Venesuelos laikas", "OESZ": "Rytų Europos vasaros laikas", "ART": "Argentinos žiemos laikas", "AST": "Atlanto žiemos laikas", "GFT": "Prancūzijos Gvianos laikas", "EAT": "Rytų Afrikos laikas", "HADT": "Havajų–Aleutų vasaros laikas", "COT": "Kolumbijos žiemos laikas", "CDT": "Šiaurės Amerikos centro vasaros laikas", "PST": "Šiaurės Amerikos Ramiojo vandenyno žiemos laikas", "AWDT": "Vakarų Australijos vasaros laikas"},
	}
}

// Locale returns the current translators string locale
func (lt *lt_LT) Locale() string {
	return lt.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'lt_LT'
func (lt *lt_LT) PluralsCardinal() []locales.PluralRule {
	return lt.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'lt_LT'
func (lt *lt_LT) PluralsOrdinal() []locales.PluralRule {
	return lt.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'lt_LT'
func (lt *lt_LT) PluralsRange() []locales.PluralRule {
	return lt.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'lt_LT'
func (lt *lt_LT) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)
	f := locales.F(n, v)
	nMod10 := math.Mod(n, 10)
	nMod100 := math.Mod(n, 100)

	if nMod10 == 1 && (nMod100 < 11 || nMod100 > 19) {
		return locales.PluralRuleOne
	} else if nMod10 >= 2 && nMod10 <= 9 && (nMod100 < 11 || nMod100 > 19) {
		return locales.PluralRuleFew
	} else if f != 0 {
		return locales.PluralRuleMany
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'lt_LT'
func (lt *lt_LT) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'lt_LT'
func (lt *lt_LT) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {

	start := lt.CardinalPluralRule(num1, v1)
	end := lt.CardinalPluralRule(num2, v2)

	if start == locales.PluralRuleOne && end == locales.PluralRuleOne {
		return locales.PluralRuleOne
	} else if start == locales.PluralRuleOne && end == locales.PluralRuleFew {
		return locales.PluralRuleFew
	} else if start == locales.PluralRuleOne && end == locales.PluralRuleMany {
		return locales.PluralRuleMany
	} else if start == locales.PluralRuleOne && end == locales.PluralRuleOther {
		return locales.PluralRuleOther
	} else if start == locales.PluralRuleFew && end == locales.PluralRuleOne {
		return locales.PluralRuleOne
	} else if start == locales.PluralRuleFew && end == locales.PluralRuleFew {
		return locales.PluralRuleFew
	} else if start == locales.PluralRuleFew && end == locales.PluralRuleMany {
		return locales.PluralRuleMany
	} else if start == locales.PluralRuleFew && end == locales.PluralRuleOther {
		return locales.PluralRuleOther
	} else if start == locales.PluralRuleMany && end == locales.PluralRuleOne {
		return locales.PluralRuleOne
	} else if start == locales.PluralRuleMany && end == locales.PluralRuleFew {
		return locales.PluralRuleFew
	} else if start == locales.PluralRuleMany && end == locales.PluralRuleMany {
		return locales.PluralRuleMany
	} else if start == locales.PluralRuleMany && end == locales.PluralRuleOther {
		return locales.PluralRuleOther
	} else if start == locales.PluralRuleOther && end == locales.PluralRuleOne {
		return locales.PluralRuleOne
	} else if start == locales.PluralRuleOther && end == locales.PluralRuleFew {
		return locales.PluralRuleFew
	} else if start == locales.PluralRuleOther && end == locales.PluralRuleMany {
		return locales.PluralRuleMany
	}

	return locales.PluralRuleOther

}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (lt *lt_LT) MonthAbbreviated(month time.Month) string {
	return lt.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (lt *lt_LT) MonthsAbbreviated() []string {
	return lt.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (lt *lt_LT) MonthNarrow(month time.Month) string {
	return lt.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (lt *lt_LT) MonthsNarrow() []string {
	return lt.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (lt *lt_LT) MonthWide(month time.Month) string {
	return lt.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (lt *lt_LT) MonthsWide() []string {
	return lt.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (lt *lt_LT) WeekdayAbbreviated(weekday time.Weekday) string {
	return lt.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (lt *lt_LT) WeekdaysAbbreviated() []string {
	return lt.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (lt *lt_LT) WeekdayNarrow(weekday time.Weekday) string {
	return lt.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (lt *lt_LT) WeekdaysNarrow() []string {
	return lt.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (lt *lt_LT) WeekdayShort(weekday time.Weekday) string {
	return lt.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (lt *lt_LT) WeekdaysShort() []string {
	return lt.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (lt *lt_LT) WeekdayWide(weekday time.Weekday) string {
	return lt.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (lt *lt_LT) WeekdaysWide() []string {
	return lt.daysWide
}

// Decimal returns the decimal point of number
func (lt *lt_LT) Decimal() string {
	return lt.decimal
}

// Group returns the group of number
func (lt *lt_LT) Group() string {
	return lt.group
}

// Group returns the minus sign of number
func (lt *lt_LT) Minus() string {
	return lt.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'lt_LT' and handles both Whole and Real numbers based on 'v'
func (lt *lt_LT) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 4 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, lt.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(lt.group) - 1; j >= 0; j-- {
					b = append(b, lt.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		for j := len(lt.minus) - 1; j >= 0; j-- {
			b = append(b, lt.minus[j])
		}
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'lt_LT' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (lt *lt_LT) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 7
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, lt.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		for j := len(lt.minus) - 1; j >= 0; j-- {
			b = append(b, lt.minus[j])
		}
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, lt.percentSuffix...)

	b = append(b, lt.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'lt_LT'
func (lt *lt_LT) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := lt.currencies[currency]
	l := len(s) + len(symbol) + 6 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, lt.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(lt.group) - 1; j >= 0; j-- {
					b = append(b, lt.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		for j := len(lt.minus) - 1; j >= 0; j-- {
			b = append(b, lt.minus[j])
		}
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, lt.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	b = append(b, lt.currencyPositiveSuffix...)

	b = append(b, symbol...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'lt_LT'
// in accounting notation.
func (lt *lt_LT) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := lt.currencies[currency]
	l := len(s) + len(symbol) + 6 + 2*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, lt.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				for j := len(lt.group) - 1; j >= 0; j-- {
					b = append(b, lt.group[j])
				}
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {

		for j := len(lt.minus) - 1; j >= 0; j-- {
			b = append(b, lt.minus[j])
		}

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, lt.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, lt.currencyNegativeSuffix...)
		b = append(b, symbol...)
	} else {

		b = append(b, lt.currencyPositiveSuffix...)
		b = append(b, symbol...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'lt_LT'
func (lt *lt_LT) FmtDateShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x2d}...)

	if t.Month() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Month()), 10)

	b = append(b, []byte{0x2d}...)

	if t.Day() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Day()), 10)

	return string(b)
}

// FmtDateMedium returns the medium date representation of 't' for 'lt_LT'
func (lt *lt_LT) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x2d}...)

	if t.Month() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Month()), 10)

	b = append(b, []byte{0x2d}...)

	if t.Day() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Day()), 10)

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'lt_LT'
func (lt *lt_LT) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x20, 0x6d}...)
	b = append(b, []byte{0x2e, 0x20}...)
	b = append(b, lt.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20, 0x64}...)
	b = append(b, []byte{0x2e}...)

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'lt_LT'
func (lt *lt_LT) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x20, 0x6d}...)
	b = append(b, []byte{0x2e, 0x20}...)
	b = append(b, lt.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20, 0x64}...)
	b = append(b, []byte{0x2e, 0x2c, 0x20}...)
	b = append(b, lt.daysWide[t.Weekday()]...)

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'lt_LT'
func (lt *lt_LT) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, lt.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'lt_LT'
func (lt *lt_LT) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, lt.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, lt.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'lt_LT'
func (lt *lt_LT) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, lt.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, lt.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'lt_LT'
func (lt *lt_LT) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, lt.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, lt.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := lt.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
