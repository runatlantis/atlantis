package si_LK

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type si_LK struct {
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

// New returns a new instance of translator for the 'si_LK' locale
func New() locales.Translator {
	return &si_LK{
		locale:                 "si_LK",
		pluralsCardinal:        []locales.PluralRule{2, 6},
		pluralsOrdinal:         []locales.PluralRule{6},
		pluralsRange:           []locales.PluralRule{2, 6},
		decimal:                ".",
		group:                  ",",
		minus:                  "-",
		percent:                "%",
		perMille:               "‰",
		timeSeparator:          ".",
		inifinity:              "∞",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyNegativePrefix: "(",
		currencyNegativeSuffix: ")",
		monthsAbbreviated:      []string{"", "ජන", "පෙබ", "මාර්තු", "අප්\u200dරේල්", "මැයි", "ජූනි", "ජූලි", "අගෝ", "සැප්", "ඔක්", "නොවැ", "දෙසැ"},
		monthsNarrow:           []string{"", "ජ", "පෙ", "මා", "අ", "මැ", "ජූ", "ජූ", "අ", "සැ", "ඔ", "නෙ", "දෙ"},
		monthsWide:             []string{"", "ජනවාරි", "පෙබරවාරි", "මාර්තු", "අප්\u200dරේල්", "මැයි", "ජූනි", "ජූලි", "අගෝස්තු", "සැප්තැම්බර්", "ඔක්තෝබර්", "නොවැම්බර්", "දෙසැම්බර්"},
		daysAbbreviated:        []string{"ඉරිදා", "සඳුදා", "අඟහ", "බදාදා", "බ්\u200dරහස්", "සිකු", "සෙන"},
		daysNarrow:             []string{"ඉ", "ස", "අ", "බ", "බ්\u200dර", "සි", "සෙ"},
		daysShort:              []string{"ඉරි", "සඳු", "අඟ", "බදා", "බ්\u200dරහ", "සිකු", "සෙන"},
		daysWide:               []string{"ඉරිදා", "සඳුදා", "අඟහරුවාදා", "බදාදා", "බ්\u200dරහස්පතින්දා", "සිකුරාදා", "සෙනසුරාදා"},
		periodsAbbreviated:     []string{"පෙ.ව.", "ප.ව."},
		periodsNarrow:          []string{"පෙ", "ප"},
		periodsWide:            []string{"පෙ.ව.", "ප.ව."},
		erasAbbreviated:        []string{"ක්\u200dරි.පූ.", "ක්\u200dරි.ව."},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"ක්\u200dරිස්තු පූර්ව", "ක්\u200dරිස්තු වර්ෂ"},
		timezones:              map[string]string{"WAT": "බටහිර අප්\u200dරිකානු සම්මත වේලාව", "OEZ": "නැගෙනහිර යුරෝපීය සම්මත වේලාව", "AKDT": "ඇලස්කා දිවාආලෝක වේලාව", "HAT": "නිව්ෆවුන්ලන්ත දිවාආලෝක වේලාව", "COST": "කොලොම්බියා ග්\u200dරීෂ්ම කාලය", "GMT": "ග්\u200dරිනිච් මධ්\u200dයම වේලාව", "PST": "උතුරු ඇමරිකානු පැසිෆික් සම්මත වේලාව", "WEZ": "බටහිර යුරෝපීය සම්මත වේලාව", "WART": "බටහිර ආර්ජන්ටිනා සම්මත වේලාව", "WITA": "මධ්\u200dයම ඉන්දුනීසියානු වේලාව", "CST": "උතුරු ඇමරිකානු මධ්\u200dයම සම්මත වේලාව", "GFT": "ප්\u200dරංශ ගයනා වේලාව", "ACWST": "මධ්\u200dයම බටහිර ඔස්ට්\u200dරේලියානු සම්මත වේලාව", "LHDT": "ලෝර්ඩ් හෝව් දිවා වේලාව", "AEDT": "නැඟෙනහිර ඕස්ට්\u200dරේලියානු දහවල් වේලාව", "EST": "උතුරු ඇමරිකානු නැගෙනහිර සම්මත වේලාව", "CLT": "චිලී සම්මත වේලාව", "CLST": "චිලී ග්\u200dරීෂ්ම කාලය", "OESZ": "නැගෙනහිර යුරෝපීය ග්\u200dරීෂ්ම වේලාව", "ART": "ආර්ජන්ටිනා සම්මත වේලාව", "UYT": "උරුගුවේ සම්මත වේලාව", "AWST": "බටහිර ඕස්ට්\u200dරේලියානු සම්මත වේලාව", "MESZ": "මධ්\u200dයම යුරෝපීය ග්\u200dරීෂ්ම වේලාව", "HKST": "හොංකොං ග්\u200dරීෂ්ම වේලාව", "MDT": "උතුරු ඇමරිකානු කඳුකර දිවාආලෝක වේලාව", "WAST": "බටහිර අප්\u200dරිකානු ග්\u200dරීෂ්ම කාලය", "WIB": "බටහිර ඉන්දුනීසියානු වේලාව", "HEEG": "නැගෙනහිර ග්\u200dරීන්ලන්ත ග්\u200dරීෂ්ම කාලය", "TMST": "ටර්ක්මෙනිස්තාන ග්\u200dරීෂ්ම වේලාව", "ECT": "ඉක්වදෝර් වේලාව", "EDT": "උතුරු ඇමරිකානු නැගෙනහිර දිවාආලෝක වේලාව", "ACDT": "මධ්\u200dයම ඔස්ට්\u200dරේලියානු දහවල් වේලාව", "HNPM": "ශාන්ත පියරේ සහ මැකෝලන් සම්මත වේලාව", "HAST": "හවායි-අලෙයුතියාන් සම්මත වේලාව", "HNCU": "කියුබානු සම්මත වේලාව", "HECU": "කියුබානු දිවාආලෝක වේලාව", "PDT": "උතුරු ඇමරිකානු පැසිෆික් දිවාආලෝක වේලාව", "HNT": "නිව්ෆවුන්ලන්ත සම්මත වේලාව", "VET": "වෙනිසියුලා වේලාව", "JDT": "ජපාන දහවල් වේලාව", "IST": "ඉන්දියානු වේලාව", "∅∅∅": "ඇමර්සන් ග්\u200dරීෂ්ම කාලය", "CDT": "උතුරු ඇමරිකානු මධ්\u200dයම දිවාආලෝක වේලාව", "HNPMX": "මෙක්සිකෝ පැසිෆික් සම්මත වේලාව", "HEPMX": "මෙක්සිකෝ පැසිෆික් දිවාආලෝක වේලාව", "NZDT": "නවසීලන්ත දිවා වේලාව", "JST": "ජපාන සම්මත වේලාව", "HEPM": "ශාන්ත පියරේ සහ මැකෝලන් දිවාආලෝක වේලාව", "SRT": "සුරිනාම වේලාව", "ACWDT": "මධ්\u200dයම බටහිර ඔස්ට්\u200dරේලියානු දහවල් වේලාව", "HEOG": "බටහිර ග්\u200dරීන්ලන්ත ග්\u200dරීෂ්ම කාලය", "TMT": "ටර්ක්මෙනිස්තාන සම්මත වේලාව", "ARST": "ආර්ජන්ටිනා ග්\u200dරීෂ්ම කාලය", "GYT": "ගයනා වේලාව", "ChST": "චමොරෝ වේලාව", "AST": "අත්ලාන්තික් සම්මත වේලාව", "AKST": "ඇලස්කා සම්මත වේලාව", "MEZ": "මධ්\u200dයම යුරෝපීය සම්මත වේලාව", "LHST": "ලෝර්ඩ් හෝව් සම්මත වේලාව", "HADT": "හවායි-අලෙයුතියාන් දිවාආලෝක වේලාව", "WESZ": "බටහිර යුරෝපීය ග්\u200dරීෂ්ම වේලාව", "BOT": "බොලිවියා වේලාව", "HENOMX": "වයඹ මෙක්සිකෝ දිවාආලෝක වේලාව", "MST": "උතුරු ඇමරිකානු කඳුකර සම්මත වේලාව", "BT": "භුතාන වේලාව", "NZST": "නවසීලන්ත සම්මත වේලාව", "HNOG": "බටහිර ග්\u200dරීන්ලන්ත සම්මත වේලාව", "HNNOMX": "වයඹ මෙක්සිකෝ සම්මත වේලාව", "HNEG": "නැගෙනහිර ග්\u200dරීන්ලන්ත සම්මත වේලාව", "CAT": "මධ්\u200dයම අප්\u200dරිකානු වේලාව", "WIT": "නැගෙනහිර ඉන්දුනීසියානු වේලාව", "COT": "කොලොම්බියා සම්මත වේලාව", "UYST": "උරුගුවේ ග්\u200dරීෂ්ම කාලය", "CHADT": "චැතම් දිවා වේලාව", "ACST": "ඕස්ට්\u200dරේලියානු සම්මත වේලාව", "ADT": "අත්ලාන්තික් දිවාආලෝක වේලාව", "AEST": "නැගෙනහිර ඕස්ට්\u200dරේලියානු සම්මත වේලාව", "HKT": "හොංකොං සම්මත වේලාව", "EAT": "නැගෙනහිර අප්\u200dරිකානු වේලාව", "CHAST": "චැතම් සම්මත වේලාව", "AWDT": "බටහිර ඔස්ට්\u200dරේලියානු දහවල් වේලාව", "SGT": "සිංගප්පුරු වේලාව", "WARST": "බටහිර ආර්ජන්ටිනා ග්\u200dරීෂ්ම කාලය", "SAST": "දකුණු අප්\u200dරිකානු වේලාව", "MYT": "මැලේසියානු වේලාව"},
	}
}

// Locale returns the current translators string locale
func (si *si_LK) Locale() string {
	return si.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'si_LK'
func (si *si_LK) PluralsCardinal() []locales.PluralRule {
	return si.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'si_LK'
func (si *si_LK) PluralsOrdinal() []locales.PluralRule {
	return si.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'si_LK'
func (si *si_LK) PluralsRange() []locales.PluralRule {
	return si.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'si_LK'
func (si *si_LK) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)
	i := int64(n)
	f := locales.F(n, v)

	if (n == 0 || n == 1) || (i == 0 && f == 1) {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'si_LK'
func (si *si_LK) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'si_LK'
func (si *si_LK) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {

	start := si.CardinalPluralRule(num1, v1)
	end := si.CardinalPluralRule(num2, v2)

	if start == locales.PluralRuleOne && end == locales.PluralRuleOne {
		return locales.PluralRuleOne
	} else if start == locales.PluralRuleOne && end == locales.PluralRuleOther {
		return locales.PluralRuleOther
	} else if start == locales.PluralRuleOther && end == locales.PluralRuleOne {
		return locales.PluralRuleOther
	}

	return locales.PluralRuleOther

}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (si *si_LK) MonthAbbreviated(month time.Month) string {
	return si.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (si *si_LK) MonthsAbbreviated() []string {
	return si.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (si *si_LK) MonthNarrow(month time.Month) string {
	return si.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (si *si_LK) MonthsNarrow() []string {
	return si.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (si *si_LK) MonthWide(month time.Month) string {
	return si.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (si *si_LK) MonthsWide() []string {
	return si.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (si *si_LK) WeekdayAbbreviated(weekday time.Weekday) string {
	return si.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (si *si_LK) WeekdaysAbbreviated() []string {
	return si.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (si *si_LK) WeekdayNarrow(weekday time.Weekday) string {
	return si.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (si *si_LK) WeekdaysNarrow() []string {
	return si.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (si *si_LK) WeekdayShort(weekday time.Weekday) string {
	return si.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (si *si_LK) WeekdaysShort() []string {
	return si.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (si *si_LK) WeekdayWide(weekday time.Weekday) string {
	return si.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (si *si_LK) WeekdaysWide() []string {
	return si.daysWide
}

// Decimal returns the decimal point of number
func (si *si_LK) Decimal() string {
	return si.decimal
}

// Group returns the group of number
func (si *si_LK) Group() string {
	return si.group
}

// Group returns the minus sign of number
func (si *si_LK) Minus() string {
	return si.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'si_LK' and handles both Whole and Real numbers based on 'v'
func (si *si_LK) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, si.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, si.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, si.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'si_LK' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (si *si_LK) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 3
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, si.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, si.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, si.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'si_LK'
func (si *si_LK) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := si.currencies[currency]
	l := len(s) + len(symbol) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, si.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, si.group[0])
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
		b = append(b, si.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, si.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'si_LK'
// in accounting notation.
func (si *si_LK) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := si.currencies[currency]
	l := len(s) + len(symbol) + 4 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, si.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, si.group[0])
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

		b = append(b, si.currencyNegativePrefix[0])

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
			b = append(b, si.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, si.currencyNegativeSuffix...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'si_LK'
func (si *si_LK) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'si_LK'
func (si *si_LK) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x20}...)
	b = append(b, si.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'si_LK'
func (si *si_LK) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x20}...)
	b = append(b, si.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'si_LK'
func (si *si_LK) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x20}...)
	b = append(b, si.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2c, 0x20}...)
	b = append(b, si.daysWide[t.Weekday()]...)

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'si_LK'
func (si *si_LK) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, []byte{0x2e}...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'si_LK'
func (si *si_LK) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, []byte{0x2e}...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, []byte{0x2e}...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'si_LK'
func (si *si_LK) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, []byte{0x2e}...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, []byte{0x2e}...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'si_LK'
func (si *si_LK) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, []byte{0x2e}...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, []byte{0x2e}...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := si.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
