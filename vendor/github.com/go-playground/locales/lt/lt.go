package lt

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type lt struct {
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

// New returns a new instance of translator for the 'lt' locale
func New() locales.Translator {
	return &lt{
		locale:                 "lt",
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
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "€", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
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
		timezones:              map[string]string{"CDT": "Šiaurės Amerikos centro vasaros laikas", "WAT": "Vakarų Afrikos žiemos laikas", "HNNOMX": "Šiaurės Vakarų Meksikos žiemos laikas", "HAST": "Havajų–Aleutų žiemos laikas", "ARST": "Argentinos vasaros laikas", "UYST": "Urugvajaus vasaros laikas", "MDT": "Makau vasaros laikas", "CLST": "Čilės vasaros laikas", "ART": "Argentinos žiemos laikas", "JDT": "Japonijos vasaros laikas", "ACDT": "Centrinės Australijos vasaros laikas", "OEZ": "Rytų Europos žiemos laikas", "UYT": "Urugvajaus žiemos laikas", "HNCU": "Kubos žiemos laikas", "HECU": "Kubos vasaros laikas", "AST": "Atlanto žiemos laikas", "HKT": "Honkongo žiemos laikas", "TMST": "Turkmėnistano vasaros laikas", "COT": "Kolumbijos žiemos laikas", "AKST": "Aliaskos žiemos laikas", "EST": "Šiaurės Amerikos rytų žiemos laikas", "WIT": "Rytų Indonezijos laikas", "MESZ": "Vidurio Europos vasaros laikas", "HKST": "Honkongo vasaros laikas", "CAT": "Centrinės Afrikos laikas", "AEST": "Rytų Australijos žiemos laikas", "ECT": "Ekvadoro laikas", "ACST": "Centrinės Australijos žiemos laikas", "HNEG": "Grenlandijos rytų žiemos laikas", "HNT": "Niufaundlendo žiemos laikas", "SRT": "Surinamo laikas", "CHADT": "Čatamo vasaros laikas", "AWDT": "Vakarų Australijos vasaros laikas", "ADT": "Atlanto vasaros laikas", "NZST": "Naujosios Zelandijos žiemos laikas", "GFT": "Prancūzijos Gvianos laikas", "HEEG": "Grenlandijos rytų vasaros laikas", "AEDT": "Rytų Australijos vasaros laikas", "SAST": "Pietų Afrikos laikas", "MYT": "Malaizijos laikas", "HAT": "Niufaundlendo vasaros laikas", "OESZ": "Rytų Europos vasaros laikas", "WARST": "Vakarų Argentinos vasaros laikas", "VET": "Venesuelos laikas", "HENOMX": "Šiaurės Vakarų Meksikos vasaros laikas", "EAT": "Rytų Afrikos laikas", "GMT": "Grinvičo laikas", "BT": "Butano laikas", "AKDT": "Aliaskos vasaros laikas", "SGT": "Singapūro laikas", "AWST": "Vakarų Australijos žiemos laikas", "HADT": "Havajų–Aleutų vasaros laikas", "PST": "Šiaurės Amerikos Ramiojo vandenyno žiemos laikas", "WEZ": "Vakarų Europos žiemos laikas", "JST": "Japonijos žiemos laikas", "BOT": "Bolivijos laikas", "TMT": "Turkmėnistano žiemos laikas", "WIB": "Vakarų Indonezijos laikas", "MST": "Makau žiemos laikas", "COST": "Kolumbijos vasaros laikas", "WAST": "Vakarų Afrikos vasaros laikas", "CST": "Šiaurės Amerikos centro žiemos laikas", "MEZ": "Vidurio Europos žiemos laikas", "HEPM": "Sen Pjero ir Mikelono vasaros laikas", "CLT": "Čilės žiemos laikas", "WESZ": "Vakarų Europos vasaros laikas", "EDT": "Šiaurės Amerikos rytų vasaros laikas", "ACWDT": "Centrinės vakarų Australijos vasaros laikas", "LHDT": "Lordo Hau vasaros laikas", "HEPMX": "Meksikos Ramiojo vandenyno vasaros laikas", "WART": "Vakarų Argentinos žiemos laikas", "GYT": "Gajanos laikas", "CHAST": "Čatamo žiemos laikas", "HEOG": "Grenlandijos vakarų vasaros laikas", "PDT": "Šiaurės Amerikos Ramiojo vandenyno vasaros laikas", "NZDT": "Naujosios Zelandijos vasaros laikas", "ACWST": "Centrinės vakarų Australijos žiemos laikas", "HNOG": "Grenlandijos vakarų žiemos laikas", "HNPM": "Sen Pjero ir Mikelono žiemos laikas", "WITA": "Centrinės Indonezijos laikas", "ChST": "Čamoro laikas", "HNPMX": "Meksikos Ramiojo vandenyno žiemos laikas", "∅∅∅": "Azorų Salų vasaros laikas", "IST": "Indijos laikas", "LHST": "Lordo Hau žiemos laikas"},
	}
}

// Locale returns the current translators string locale
func (lt *lt) Locale() string {
	return lt.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'lt'
func (lt *lt) PluralsCardinal() []locales.PluralRule {
	return lt.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'lt'
func (lt *lt) PluralsOrdinal() []locales.PluralRule {
	return lt.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'lt'
func (lt *lt) PluralsRange() []locales.PluralRule {
	return lt.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'lt'
func (lt *lt) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

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

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'lt'
func (lt *lt) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'lt'
func (lt *lt) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {

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
func (lt *lt) MonthAbbreviated(month time.Month) string {
	return lt.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (lt *lt) MonthsAbbreviated() []string {
	return lt.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (lt *lt) MonthNarrow(month time.Month) string {
	return lt.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (lt *lt) MonthsNarrow() []string {
	return lt.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (lt *lt) MonthWide(month time.Month) string {
	return lt.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (lt *lt) MonthsWide() []string {
	return lt.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (lt *lt) WeekdayAbbreviated(weekday time.Weekday) string {
	return lt.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (lt *lt) WeekdaysAbbreviated() []string {
	return lt.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (lt *lt) WeekdayNarrow(weekday time.Weekday) string {
	return lt.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (lt *lt) WeekdaysNarrow() []string {
	return lt.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (lt *lt) WeekdayShort(weekday time.Weekday) string {
	return lt.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (lt *lt) WeekdaysShort() []string {
	return lt.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (lt *lt) WeekdayWide(weekday time.Weekday) string {
	return lt.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (lt *lt) WeekdaysWide() []string {
	return lt.daysWide
}

// Decimal returns the decimal point of number
func (lt *lt) Decimal() string {
	return lt.decimal
}

// Group returns the group of number
func (lt *lt) Group() string {
	return lt.group
}

// Group returns the minus sign of number
func (lt *lt) Minus() string {
	return lt.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'lt' and handles both Whole and Real numbers based on 'v'
func (lt *lt) FmtNumber(num float64, v uint64) string {

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

// FmtPercent returns 'num' with digits/precision of 'v' for 'lt' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (lt *lt) FmtPercent(num float64, v uint64) string {
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

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'lt'
func (lt *lt) FmtCurrency(num float64, v uint64, currency currency.Type) string {

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

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'lt'
// in accounting notation.
func (lt *lt) FmtAccounting(num float64, v uint64, currency currency.Type) string {

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

// FmtDateShort returns the short date representation of 't' for 'lt'
func (lt *lt) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'lt'
func (lt *lt) FmtDateMedium(t time.Time) string {

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

// FmtDateLong returns the long date representation of 't' for 'lt'
func (lt *lt) FmtDateLong(t time.Time) string {

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

// FmtDateFull returns the full date representation of 't' for 'lt'
func (lt *lt) FmtDateFull(t time.Time) string {

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

// FmtTimeShort returns the short time representation of 't' for 'lt'
func (lt *lt) FmtTimeShort(t time.Time) string {

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

// FmtTimeMedium returns the medium time representation of 't' for 'lt'
func (lt *lt) FmtTimeMedium(t time.Time) string {

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

// FmtTimeLong returns the long time representation of 't' for 'lt'
func (lt *lt) FmtTimeLong(t time.Time) string {

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

// FmtTimeFull returns the full time representation of 't' for 'lt'
func (lt *lt) FmtTimeFull(t time.Time) string {

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
