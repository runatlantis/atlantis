package sl

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type sl struct {
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

// New returns a new instance of translator for the 'sl' locale
func New() locales.Translator {
	return &sl{
		locale:                 "sl",
		pluralsCardinal:        []locales.PluralRule{2, 3, 4, 6},
		pluralsOrdinal:         []locales.PluralRule{6},
		pluralsRange:           []locales.PluralRule{3, 4, 6},
		decimal:                ",",
		group:                  ".",
		minus:                  "−",
		percent:                "%",
		perMille:               "‰",
		timeSeparator:          ":",
		inifinity:              "∞",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CN¥", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "€", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HK$", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "₪", "₹", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "¥", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "₩", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "$", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "₫", "VNN", "VUV", "WST", "FCFA", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "CFA", "XPD", "CFPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		percentSuffix:          " ",
		currencyPositiveSuffix: " ",
		currencyNegativePrefix: "(",
		currencyNegativeSuffix: " )",
		monthsAbbreviated:      []string{"", "jan.", "feb.", "mar.", "apr.", "maj", "jun.", "jul.", "avg.", "sep.", "okt.", "nov.", "dec."},
		monthsNarrow:           []string{"", "j", "f", "m", "a", "m", "j", "j", "a", "s", "o", "n", "d"},
		monthsWide:             []string{"", "januar", "februar", "marec", "april", "maj", "junij", "julij", "avgust", "september", "oktober", "november", "december"},
		daysAbbreviated:        []string{"ned.", "pon.", "tor.", "sre.", "čet.", "pet.", "sob."},
		daysNarrow:             []string{"n", "p", "t", "s", "č", "p", "s"},
		daysShort:              []string{"ned.", "pon.", "tor.", "sre.", "čet.", "pet.", "sob."},
		daysWide:               []string{"nedelja", "ponedeljek", "torek", "sreda", "četrtek", "petek", "sobota"},
		periodsAbbreviated:     []string{"dop.", "pop."},
		periodsNarrow:          []string{"d", "p"},
		periodsWide:            []string{"dop.", "pop."},
		erasAbbreviated:        []string{"pr. Kr.", "po Kr."},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"pred Kristusom", "po Kristusu"},
		timezones:              map[string]string{"JST": "Japonski standardni čas", "HAT": "Novofundlandski poletni čas", "CAT": "Centralnoafriški čas", "ECT": "Ekvadorski čas", "EST": "Vzhodni standardni čas", "HKT": "Hongkonški standardni čas", "HEPMX": "Mehiški pacifiški poletni čas", "BOT": "Bolivijski čas", "COST": "Kolumbijski poletni čas", "PST": "Pacifiški standardni čas", "EDT": "Vzhodni poletni čas", "HENOMX": "Mehiški severozahodni poletni čas", "OESZ": "Vzhodnoevropski poletni čas", "CST": "Centralni standardni čas", "AEDT": "Avstralski vzhodni poletni čas", "SGT": "Singapurski standardni čas", "HNOG": "Zahodnogrenlandski standardni čas", "TMST": "Turkmenistanski poletni čas", "CLT": "Čilski standardni čas", "TMT": "Turkmenistanski standardni čas", "OEZ": "Vzhodnoevropski standardni čas", "AEST": "Avstralski vzhodni standardni čas", "GFT": "Čas: Francoska Gvajana", "HNEG": "Vzhodnogrenlandski standardni čas", "HKST": "Hongkonški poletni čas", "EAT": "Vzhodnoafriški čas", "GMT": "Greenwiški srednji čas", "HEPM": "Poletni čas: Saint Pierre in Miquelon", "HADT": "Havajski aleutski poletni čas", "ChST": "Čamorski standardni čas", "CDT": "Centralni poletni čas", "NZDT": "Novozelandski poletni čas", "AKST": "Aljaški standardni čas", "ACST": "Avstralski centralni standardni čas", "LHDT": "Poletni čas otoka Lord Howe", "CHAST": "Čatamski standardni čas", "VET": "Venezuelski čas", "PDT": "Pacifiški poletni čas", "WEZ": "Zahodnoevropski standardni čas", "MYT": "Malezijski čas", "ACDT": "Avstralski centralni poletni čas", "HEOG": "Zahodnogrenlandski poletni čas", "ART": "Argentinski standardni čas", "COT": "Kolumbijski standardni čas", "UYT": "Urugvajski standardni čas", "AST": "Atlantski standardni čas", "ACWDT": "Avstralski centralni zahodni poletni čas", "LHST": "Standardni čas otoka Lord Howe", "SRT": "Surinamski čas", "CLST": "Čilski poletni čas", "HNCU": "Kubanski standardni čas", "MDT": "Gorski poletni čas", "WIB": "Indonezijski zahodni čas", "NZST": "Novozelandski standardni čas", "JDT": "Japonski poletni čas", "HNT": "Novofundlandski standardni čas", "BT": "Butanski čas", "IST": "Indijski standardni čas", "HNPM": "Standardni čas: Saint Pierre in Miquelon", "HNNOMX": "Mehiški severozahodni standardni čas", "WIT": "Indonezijski vzhodni čas", "MESZ": "Srednjeevropski poletni čas", "WART": "Argentinski zahodni standardni čas", "UYST": "Urugvajski poletni čas", "AWDT": "Avstralski zahodni poletni čas", "HNPMX": "Mehiški pacifiški standardni čas", "MST": "Gorski standardni čas", "WESZ": "Zahodnoevropski poletni čas", "∅∅∅": "∅∅∅", "CHADT": "Čatamski poletni čas", "ACWST": "Avstralski centralni zahodni standardni čas", "HEEG": "Vzhodnogrenlandski poletni čas", "WARST": "Argentinski zahodni poletni čas", "HAST": "Havajski aleutski standardni čas", "SAST": "Južnoafriški čas", "WAST": "Zahodnoafriški poletni čas", "AKDT": "Aljaški poletni čas", "MEZ": "Srednjeevropski standardni čas", "WITA": "Indonezijski osrednji čas", "HECU": "Kubanski poletni čas", "AWST": "Avstralski zahodni standardni čas", "ADT": "Atlantski poletni čas", "WAT": "Zahodnoafriški standardni čas", "ARST": "Argentinski poletni čas", "GYT": "Gvajanski čas"},
	}
}

// Locale returns the current translators string locale
func (sl *sl) Locale() string {
	return sl.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'sl'
func (sl *sl) PluralsCardinal() []locales.PluralRule {
	return sl.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'sl'
func (sl *sl) PluralsOrdinal() []locales.PluralRule {
	return sl.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'sl'
func (sl *sl) PluralsRange() []locales.PluralRule {
	return sl.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'sl'
func (sl *sl) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)
	i := int64(n)
	iMod100 := i % 100

	if v == 0 && iMod100 == 1 {
		return locales.PluralRuleOne
	} else if v == 0 && iMod100 == 2 {
		return locales.PluralRuleTwo
	} else if (v == 0 && iMod100 >= 3 && iMod100 <= 4) || (v != 0) {
		return locales.PluralRuleFew
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'sl'
func (sl *sl) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'sl'
func (sl *sl) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {

	start := sl.CardinalPluralRule(num1, v1)
	end := sl.CardinalPluralRule(num2, v2)

	if start == locales.PluralRuleOne && end == locales.PluralRuleOne {
		return locales.PluralRuleFew
	} else if start == locales.PluralRuleOne && end == locales.PluralRuleTwo {
		return locales.PluralRuleTwo
	} else if start == locales.PluralRuleOne && end == locales.PluralRuleFew {
		return locales.PluralRuleFew
	} else if start == locales.PluralRuleOne && end == locales.PluralRuleOther {
		return locales.PluralRuleOther
	} else if start == locales.PluralRuleTwo && end == locales.PluralRuleOne {
		return locales.PluralRuleFew
	} else if start == locales.PluralRuleTwo && end == locales.PluralRuleTwo {
		return locales.PluralRuleTwo
	} else if start == locales.PluralRuleTwo && end == locales.PluralRuleFew {
		return locales.PluralRuleFew
	} else if start == locales.PluralRuleTwo && end == locales.PluralRuleOther {
		return locales.PluralRuleOther
	} else if start == locales.PluralRuleFew && end == locales.PluralRuleOne {
		return locales.PluralRuleFew
	} else if start == locales.PluralRuleFew && end == locales.PluralRuleTwo {
		return locales.PluralRuleTwo
	} else if start == locales.PluralRuleFew && end == locales.PluralRuleFew {
		return locales.PluralRuleFew
	} else if start == locales.PluralRuleFew && end == locales.PluralRuleOther {
		return locales.PluralRuleOther
	} else if start == locales.PluralRuleOther && end == locales.PluralRuleOne {
		return locales.PluralRuleFew
	} else if start == locales.PluralRuleOther && end == locales.PluralRuleTwo {
		return locales.PluralRuleTwo
	} else if start == locales.PluralRuleOther && end == locales.PluralRuleFew {
		return locales.PluralRuleFew
	}

	return locales.PluralRuleOther

}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (sl *sl) MonthAbbreviated(month time.Month) string {
	return sl.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (sl *sl) MonthsAbbreviated() []string {
	return sl.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (sl *sl) MonthNarrow(month time.Month) string {
	return sl.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (sl *sl) MonthsNarrow() []string {
	return sl.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (sl *sl) MonthWide(month time.Month) string {
	return sl.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (sl *sl) MonthsWide() []string {
	return sl.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (sl *sl) WeekdayAbbreviated(weekday time.Weekday) string {
	return sl.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (sl *sl) WeekdaysAbbreviated() []string {
	return sl.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (sl *sl) WeekdayNarrow(weekday time.Weekday) string {
	return sl.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (sl *sl) WeekdaysNarrow() []string {
	return sl.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (sl *sl) WeekdayShort(weekday time.Weekday) string {
	return sl.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (sl *sl) WeekdaysShort() []string {
	return sl.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (sl *sl) WeekdayWide(weekday time.Weekday) string {
	return sl.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (sl *sl) WeekdaysWide() []string {
	return sl.daysWide
}

// Decimal returns the decimal point of number
func (sl *sl) Decimal() string {
	return sl.decimal
}

// Group returns the group of number
func (sl *sl) Group() string {
	return sl.group
}

// Group returns the minus sign of number
func (sl *sl) Minus() string {
	return sl.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'sl' and handles both Whole and Real numbers based on 'v'
func (sl *sl) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 4 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, sl.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, sl.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		for j := len(sl.minus) - 1; j >= 0; j-- {
			b = append(b, sl.minus[j])
		}
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'sl' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (sl *sl) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 7
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, sl.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		for j := len(sl.minus) - 1; j >= 0; j-- {
			b = append(b, sl.minus[j])
		}
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, sl.percentSuffix...)

	b = append(b, sl.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'sl'
func (sl *sl) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := sl.currencies[currency]
	l := len(s) + len(symbol) + 6 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, sl.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, sl.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		for j := len(sl.minus) - 1; j >= 0; j-- {
			b = append(b, sl.minus[j])
		}
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, sl.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	b = append(b, sl.currencyPositiveSuffix...)

	b = append(b, symbol...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'sl'
// in accounting notation.
func (sl *sl) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := sl.currencies[currency]
	l := len(s) + len(symbol) + 8 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, sl.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, sl.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {

		b = append(b, sl.currencyNegativePrefix[0])

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, sl.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, sl.currencyNegativeSuffix...)
		b = append(b, symbol...)
	} else {

		b = append(b, sl.currencyPositiveSuffix...)
		b = append(b, symbol...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'sl'
func (sl *sl) FmtDateShort(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2e, 0x20}...)

	if t.Month() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Month()), 10)

	b = append(b, []byte{0x2e, 0x20}...)

	if t.Year() > 9 {
		b = append(b, strconv.Itoa(t.Year())[2:]...)
	} else {
		b = append(b, strconv.Itoa(t.Year())[1:]...)
	}

	return string(b)
}

// FmtDateMedium returns the medium date representation of 't' for 'sl'
func (sl *sl) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2e, 0x20}...)
	b = append(b, sl.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'sl'
func (sl *sl) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Day() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2e, 0x20}...)
	b = append(b, sl.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'sl'
func (sl *sl) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, sl.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)

	if t.Day() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2e, 0x20}...)
	b = append(b, sl.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'sl'
func (sl *sl) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, sl.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'sl'
func (sl *sl) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, sl.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, sl.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'sl'
func (sl *sl) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, sl.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, sl.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'sl'
func (sl *sl) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, sl.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, sl.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := sl.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
