package he

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type he struct {
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

// New returns a new instance of translator for the 'he' locale
func New() locales.Translator {
	return &he{
		locale:                 "he",
		pluralsCardinal:        []locales.PluralRule{2, 3, 5, 6},
		pluralsOrdinal:         []locales.PluralRule{6},
		pluralsRange:           []locales.PluralRule{5, 6},
		decimal:                ".",
		group:                  ",",
		minus:                  "‎-",
		percent:                "%",
		perMille:               "‰",
		timeSeparator:          ":",
		inifinity:              "∞",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "A$", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "R$", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CA$", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "\u200eCN¥\u200e", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "€", "FIM", "FJD", "FKP", "FRF", "£", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HK$", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ל״י", "ILR", "₪", "₹", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "¥", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "₩", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MX$", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZ$", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "฿", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "NT$", "TZS", "UAH", "UAK", "UGS", "UGX", "$", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "₫", "VNN", "VUV", "WST", "FCFA", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "EC$", "XDR", "XEU", "XFO", "XFU", "CFA", "XPD", "CFPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyPositiveSuffix: " ",
		currencyNegativeSuffix: " ",
		monthsAbbreviated:      []string{"", "ינו׳", "פבר׳", "מרץ", "אפר׳", "מאי", "יוני", "יולי", "אוג׳", "ספט׳", "אוק׳", "נוב׳", "דצמ׳"},
		monthsNarrow:           []string{"", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12"},
		monthsWide:             []string{"", "ינואר", "פברואר", "מרץ", "אפריל", "מאי", "יוני", "יולי", "אוגוסט", "ספטמבר", "אוקטובר", "נובמבר", "דצמבר"},
		daysAbbreviated:        []string{"יום א׳", "יום ב׳", "יום ג׳", "יום ד׳", "יום ה׳", "יום ו׳", "שבת"},
		daysNarrow:             []string{"א׳", "ב׳", "ג׳", "ד׳", "ה׳", "ו׳", "ש׳"},
		daysShort:              []string{"א׳", "ב׳", "ג׳", "ד׳", "ה׳", "ו׳", "ש׳"},
		daysWide:               []string{"יום ראשון", "יום שני", "יום שלישי", "יום רביעי", "יום חמישי", "יום שישי", "יום שבת"},
		periodsAbbreviated:     []string{"לפנה״צ", "אחה״צ"},
		periodsNarrow:          []string{"לפנה״צ", "אחה״צ"},
		periodsWide:            []string{"לפנה״צ", "אחה״צ"},
		erasAbbreviated:        []string{"לפנה״ס", "לספירה"},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"לפני הספירה", "לספירה"},
		timezones:              map[string]string{"CDT": "שעון מרכז ארה״ב (קיץ)", "HNPMX": "שעון מערב מקסיקו (חורף)", "AKST": "שעון אלסקה (חורף)", "WESZ": "שעון מערב אירופה (קיץ)", "ACDT": "שעון מרכז אוסטרליה (קיץ)", "HNOG": "שעון מערב גרינלנד (חורף)", "HKST": "שעון הונג קונג (קיץ)", "WITA": "שעון מרכז אינדונזיה", "AST": "שעון האוקיינוס האטלנטי (חורף)", "MYT": "שעון מלזיה", "SGT": "שעון סינגפור", "MEZ": "שעון מרכז אירופה (חורף)", "PST": "שעון מערב ארה״ב (חורף)", "AKDT": "שעון אלסקה (קיץ)", "HEEG": "שעון מזרח גרינלנד (קיץ)", "HNPM": "שעון סנט פייר ומיקלון (חורף)", "HAST": "שעון האיים האלאוטיים הוואי (חורף)", "NZDT": "שעון ניו זילנד (קיץ)", "HNEG": "שעון מזרח גרינלנד (חורף)", "WARST": "שעון מערב ארגנטינה (קיץ)", "COST": "שעון קולומביה (קיץ)", "HNNOMX": "שעון צפון-מערב מקסיקו (חורף)", "CHADT": "שעון צ׳טהאם (קיץ)", "NZST": "שעון ניו זילנד (חורף)", "BOT": "שעון בוליביה", "EDT": "שעון החוף המזרחי (קיץ)", "IST": "שעון הודו", "AWST": "שעון מערב אוסטרליה (חורף)", "HEPM": "שעון סנט פייר ומיקלון (קיץ)", "CLST": "שעון צ׳ילה (קיץ)", "GMT": "שעון גריניץ׳\u200f", "HKT": "שעון הונג קונג (חורף)", "WART": "שעון מערב ארגנטינה (חורף)", "SRT": "שעון סורינאם", "MDT": "שעון אזור ההרים בארה״ב (קיץ)", "WAT": "שעון מערב אפריקה (חורף)", "WEZ": "שעון מערב אירופה (חורף)", "JST": "שעון יפן (חורף)", "EST": "שעון החוף המזרחי (חורף)", "CAT": "שעון מרכז אפריקה", "TMT": "שעון טורקמניסטן (חורף)", "COT": "שעון קולומביה (חורף)", "HEPMX": "שעון מערב מקסיקו (קיץ)", "ACWDT": "שעון מרכז-מערב אוסטרליה (קיץ)", "LHDT": "שעון אי הלורד האו (קיץ)", "UYT": "שעון אורוגוואי (חורף)", "UYST": "שעון אורוגוואי (קיץ)", "WAST": "שעון מערב אפריקה (קיץ)", "ACST": "שעון מרכז אוסטרליה (חורף)", "CLT": "שעון צ׳ילה (חורף)", "OEZ": "שעון מזרח אירופה (חורף)", "GYT": "שעון גיאנה", "HNT": "שעון ניופאונדלנד (חורף)", "HAT": "שעון ניופאונדלנד (קיץ)", "AWDT": "שעון מערב אוסטרליה (קיץ)", "WIB": "שעון מערב אינדונזיה", "HEOG": "שעון מערב גרינלנד (קיץ)", "MESZ": "שעון מרכז אירופה (קיץ)", "LHST": "שעון אי הלורד האו (חורף)", "VET": "שעון ונצואלה", "HENOMX": "שעון צפון-מערב מקסיקו (קיץ)", "TMST": "שעון טורקמניסטן (קיץ)", "CHAST": "שעון צ׳טהאם (חורף)", "HECU": "שעון קובה (קיץ)", "∅∅∅": "שעון ברזיליה (קיץ)", "MST": "שעון אזור ההרים בארה״ב (חורף)", "ACWST": "שעון מרכז-מערב אוסטרליה (חורף)", "EAT": "שעון מזרח אפריקה", "ARST": "שעון ארגנטינה (קיץ)", "OESZ": "שעון מזרח אירופה (קיץ)", "HNCU": "שעון קובה (חורף)", "PDT": "שעון מערב ארה״ב (קיץ)", "AEDT": "שעון מזרח אוסטרליה (קיץ)", "SAST": "שעון דרום אפריקה", "WIT": "שעון מזרח אינדונזיה", "JDT": "שעון יפן (קיץ)", "ECT": "שעון אקוודור", "HADT": "שעון האיים האלאוטיים הוואי (קיץ)", "ART": "שעון ארגנטינה (חורף)", "ChST": "שעון צ׳אמורו", "CST": "שעון מרכז ארה״ב (חורף)", "ADT": "שעון האוקיינוס האטלנטי (קיץ)", "AEST": "שעון מזרח אוסטרליה (חורף)", "BT": "שעון בהוטן", "GFT": "שעון גיאנה הצרפתית"},
	}
}

// Locale returns the current translators string locale
func (he *he) Locale() string {
	return he.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'he'
func (he *he) PluralsCardinal() []locales.PluralRule {
	return he.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'he'
func (he *he) PluralsOrdinal() []locales.PluralRule {
	return he.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'he'
func (he *he) PluralsRange() []locales.PluralRule {
	return he.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'he'
func (he *he) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)
	i := int64(n)
	nMod10 := math.Mod(n, 10)

	if i == 1 && v == 0 {
		return locales.PluralRuleOne
	} else if i == 2 && v == 0 {
		return locales.PluralRuleTwo
	} else if v == 0 && (n < 0 || n > 10) && nMod10 == 0 {
		return locales.PluralRuleMany
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'he'
func (he *he) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'he'
func (he *he) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {

	start := he.CardinalPluralRule(num1, v1)
	end := he.CardinalPluralRule(num2, v2)

	if start == locales.PluralRuleOne && end == locales.PluralRuleTwo {
		return locales.PluralRuleOther
	} else if start == locales.PluralRuleOne && end == locales.PluralRuleMany {
		return locales.PluralRuleMany
	} else if start == locales.PluralRuleOne && end == locales.PluralRuleOther {
		return locales.PluralRuleOther
	} else if start == locales.PluralRuleTwo && end == locales.PluralRuleMany {
		return locales.PluralRuleOther
	} else if start == locales.PluralRuleTwo && end == locales.PluralRuleOther {
		return locales.PluralRuleOther
	} else if start == locales.PluralRuleMany && end == locales.PluralRuleMany {
		return locales.PluralRuleMany
	} else if start == locales.PluralRuleMany && end == locales.PluralRuleOther {
		return locales.PluralRuleMany
	} else if start == locales.PluralRuleOther && end == locales.PluralRuleOne {
		return locales.PluralRuleOther
	} else if start == locales.PluralRuleOther && end == locales.PluralRuleTwo {
		return locales.PluralRuleOther
	} else if start == locales.PluralRuleOther && end == locales.PluralRuleMany {
		return locales.PluralRuleMany
	}

	return locales.PluralRuleOther

}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (he *he) MonthAbbreviated(month time.Month) string {
	return he.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (he *he) MonthsAbbreviated() []string {
	return he.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (he *he) MonthNarrow(month time.Month) string {
	return he.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (he *he) MonthsNarrow() []string {
	return he.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (he *he) MonthWide(month time.Month) string {
	return he.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (he *he) MonthsWide() []string {
	return he.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (he *he) WeekdayAbbreviated(weekday time.Weekday) string {
	return he.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (he *he) WeekdaysAbbreviated() []string {
	return he.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (he *he) WeekdayNarrow(weekday time.Weekday) string {
	return he.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (he *he) WeekdaysNarrow() []string {
	return he.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (he *he) WeekdayShort(weekday time.Weekday) string {
	return he.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (he *he) WeekdaysShort() []string {
	return he.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (he *he) WeekdayWide(weekday time.Weekday) string {
	return he.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (he *he) WeekdaysWide() []string {
	return he.daysWide
}

// Decimal returns the decimal point of number
func (he *he) Decimal() string {
	return he.decimal
}

// Group returns the group of number
func (he *he) Group() string {
	return he.group
}

// Group returns the minus sign of number
func (he *he) Minus() string {
	return he.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'he' and handles both Whole and Real numbers based on 'v'
func (he *he) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 5 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, he.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, he.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		for j := len(he.minus) - 1; j >= 0; j-- {
			b = append(b, he.minus[j])
		}
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'he' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (he *he) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 6
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, he.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		for j := len(he.minus) - 1; j >= 0; j-- {
			b = append(b, he.minus[j])
		}
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, he.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'he'
func (he *he) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := he.currencies[currency]
	l := len(s) + len(symbol) + 7 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, he.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, he.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		for j := len(he.minus) - 1; j >= 0; j-- {
			b = append(b, he.minus[j])
		}
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, he.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	b = append(b, he.currencyPositiveSuffix...)

	b = append(b, symbol...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'he'
// in accounting notation.
func (he *he) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := he.currencies[currency]
	l := len(s) + len(symbol) + 7 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, he.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, he.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {

		for j := len(he.minus) - 1; j >= 0; j-- {
			b = append(b, he.minus[j])
		}

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, he.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, he.currencyNegativeSuffix...)
		b = append(b, symbol...)
	} else {

		b = append(b, he.currencyPositiveSuffix...)
		b = append(b, symbol...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'he'
func (he *he) FmtDateShort(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2e}...)
	b = strconv.AppendInt(b, int64(t.Month()), 10)
	b = append(b, []byte{0x2e}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateMedium returns the medium date representation of 't' for 'he'
func (he *he) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20, 0xd7, 0x91}...)
	b = append(b, he.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'he'
func (he *he) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20, 0xd7, 0x91}...)
	b = append(b, he.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'he'
func (he *he) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, he.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20, 0xd7, 0x91}...)
	b = append(b, he.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'he'
func (he *he) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, he.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'he'
func (he *he) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, he.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, he.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'he'
func (he *he) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, he.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, he.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'he'
func (he *he) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, he.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, he.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := he.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
