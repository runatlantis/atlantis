package ur

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type ur struct {
	locale             string
	pluralsCardinal    []locales.PluralRule
	pluralsOrdinal     []locales.PluralRule
	pluralsRange       []locales.PluralRule
	decimal            string
	group              string
	minus              string
	percent            string
	perMille           string
	timeSeparator      string
	inifinity          string
	currencies         []string // idx = enum of currency code
	monthsAbbreviated  []string
	monthsNarrow       []string
	monthsWide         []string
	daysAbbreviated    []string
	daysNarrow         []string
	daysShort          []string
	daysWide           []string
	periodsAbbreviated []string
	periodsNarrow      []string
	periodsShort       []string
	periodsWide        []string
	erasAbbreviated    []string
	erasNarrow         []string
	erasWide           []string
	timezones          map[string]string
}

// New returns a new instance of translator for the 'ur' locale
func New() locales.Translator {
	return &ur{
		locale:             "ur",
		pluralsCardinal:    []locales.PluralRule{2, 6},
		pluralsOrdinal:     []locales.PluralRule{6},
		pluralsRange:       []locales.PluralRule{6},
		decimal:            ".",
		group:              ",",
		minus:              "‎-",
		percent:            "%",
		perMille:           "‰",
		timeSeparator:      ":",
		inifinity:          "∞",
		currencies:         []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "A$", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "R$", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CA$", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CN¥", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "€", "FIM", "FJD", "FKP", "FRF", "£", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HK$", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "₪", "₹", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JP¥", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "₩", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MX$", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZ$", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "Rs", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "฿", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "NT$", "TZS", "UAH", "UAK", "UGS", "UGX", "$", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "₫", "VNN", "VUV", "WST", "FCFA", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "EC$", "XDR", "XEU", "XFO", "XFU", "CFA", "XPD", "CFPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		monthsAbbreviated:  []string{"", "جنوری", "فروری", "مارچ", "اپریل", "مئی", "جون", "جولائی", "اگست", "ستمبر", "اکتوبر", "نومبر", "دسمبر"},
		monthsNarrow:       []string{"", "J", "F", "M", "A", "M", "J", "J", "A", "S", "O", "N", "D"},
		monthsWide:         []string{"", "جنوری", "فروری", "مارچ", "اپریل", "مئی", "جون", "جولائی", "اگست", "ستمبر", "اکتوبر", "نومبر", "دسمبر"},
		daysAbbreviated:    []string{"اتوار", "پیر", "منگل", "بدھ", "جمعرات", "جمعہ", "ہفتہ"},
		daysNarrow:         []string{"S", "M", "T", "W", "T", "F", "S"},
		daysShort:          []string{"اتوار", "پیر", "منگل", "بدھ", "جمعرات", "جمعہ", "ہفتہ"},
		daysWide:           []string{"اتوار", "پیر", "منگل", "بدھ", "جمعرات", "جمعہ", "ہفتہ"},
		periodsAbbreviated: []string{"AM", "PM"},
		periodsNarrow:      []string{"a", "p"},
		periodsWide:        []string{"AM", "PM"},
		erasAbbreviated:    []string{"قبل مسیح", "عیسوی"},
		erasNarrow:         []string{"", ""},
		erasWide:           []string{"قبل مسیح", "عیسوی"},
		timezones:          map[string]string{"EDT": "ایسٹرن ڈے لائٹ ٹائم", "ChST": "چامورو سٹینڈرڈ ٹائم", "NZST": "نیوزی لینڈ کا معیاری وقت", "ECT": "ایکواڈور کا وقت", "HEOG": "مغربی گرین لینڈ کا موسم گرما کا وقت", "TMT": "ترکمانستان کا معیاری وقت", "PDT": "پیسفک ڈے لائٹ ٹائم", "JDT": "جاپان ڈے لائٹ ٹائم", "LHDT": "لارڈ ہووے ڈے لائٹ ٹائم", "WARST": "مغربی ارجنٹینا کا موسم گرما کا وقت", "HENOMX": "شمال مغربی میکسیکو ڈے لائٹ ٹائم", "CAT": "وسطی افریقہ ٹائم", "WESZ": "مغربی یورپ کا موسم گرما کا وقت", "HEEG": "مشرقی گرین لینڈ کا موسم گرما کا وقت", "HNOG": "مغربی گرین لینڈ اسٹینڈرڈ ٹائم", "ACST": "آسٹریلین سنٹرل اسٹینڈرڈ ٹائم", "MEZ": "وسطی یورپ کا معیاری وقت", "HKT": "ہانگ کانگ سٹینڈرڈ ٹائم", "COT": "کولمبیا کا معیاری وقت", "GYT": "گیانا کا وقت", "PST": "پیسفک اسٹینڈرڈ ٹائم", "NZDT": "نیوزی لینڈ ڈے لائٹ ٹائم", "HNPM": "سینٹ پیئر اور مکلیئون اسٹینڈرڈ ٹائم", "CLT": "چلی کا معیاری وقت", "OEZ": "مشرقی یورپ کا معیاری وقت", "HECU": "کیوبا ڈے لائٹ ٹائم", "AEST": "آسٹریلین ایسٹرن اسٹینڈرڈ ٹائم", "AEDT": "آسٹریلین ایسٹرن ڈے لائٹ ٹائم", "MYT": "ملیشیا ٹائم", "HKST": "ہانگ کانگ سمر ٹائم", "WIT": "مشرقی انڈونیشیا ٹائم", "CLST": "چلی کا موسم گرما کا وقت", "JST": "جاپان سٹینڈرڈ ٹائم", "BT": "بھوٹان کا وقت", "HAT": "نیو فاؤنڈ لینڈ ڈے لائٹ ٹائم", "WAT": "مغربی افریقہ سٹینڈرڈ ٹائم", "HNT": "نیو فاؤنڈ لینڈ اسٹینڈرڈ ٹائم", "CHADT": "چیتھم ڈے لائٹ ٹائم", "HEPM": "سینٹ پیئر اور مکلیئون ڈے لائٹ ٹائم", "HNPMX": "میکسیکن پیسفک اسٹینڈرڈ ٹائم", "CDT": "سنٹرل ڈے لائٹ ٹائم", "EAT": "مشرقی افریقہ ٹائم", "AWST": "آسٹریلیا ویسٹرن اسٹینڈرڈ ٹائم", "∅∅∅": "∅∅∅", "ACDT": "آسٹریلین سنٹرل ڈے لائٹ ٹائم", "TMST": "ترکمانستان کا موسم گرما کا وقت", "HAST": "ہوائی الیوٹیئن اسٹینڈرڈ ٹائم", "CHAST": "چیتھم اسٹینڈرڈ ٹائم", "MDT": "ماؤنٹین ڈے لائٹ ٹائم", "WEZ": "مغربی یورپ کا معیاری وقت", "WIB": "مغربی انڈونیشیا ٹائم", "HNEG": "مشرقی گرین لینڈ اسٹینڈرڈ ٹائم", "EST": "ایسٹرن اسٹینڈرڈ ٹائم", "WART": "مغربی ارجنٹینا کا معیاری وقت", "WITA": "وسطی انڈونیشیا ٹائم", "COST": "کولمبیا کا موسم گرما کا وقت", "AWDT": "آسٹریلین ویسٹرن ڈے لائٹ ٹائم", "MST": "ماؤنٹین اسٹینڈرڈ ٹائم", "HNCU": "کیوبا اسٹینڈرڈ ٹائم", "OESZ": "مشرقی یورپ کا موسم گرما کا وقت", "ARST": "ارجنٹینا سمر ٹائم", "HADT": "ہوائی الیوٹیئن ڈے لائٹ ٹائم", "SAST": "جنوبی افریقہ سٹینڈرڈ ٹائم", "AKDT": "الاسکا ڈے لائٹ ٹائم", "BOT": "بولیویا کا وقت", "MESZ": "وسطی یورپ کا موسم گرما کا وقت", "UYST": "یوروگوئے کا موسم گرما کا وقت", "ADT": "اٹلانٹک ڈے لائٹ ٹائم", "WAST": "مغربی افریقہ سمر ٹائم", "SGT": "سنگاپور سٹینڈرڈ ٹائم", "LHST": "لارڈ ہووے اسٹینڈرڈ ٹائم", "VET": "وینزوئیلا کا وقت", "SRT": "سورینام کا وقت", "HEPMX": "میکسیکن پیسفک ڈے لائٹ ٹائم", "CST": "سنٹرل اسٹینڈرڈ ٹائم", "AKST": "الاسکا اسٹینڈرڈ ٹائم", "ACWST": "آسٹریلین سنٹرل ویسٹرن اسٹینڈرڈ ٹائم", "ACWDT": "آسٹریلین سنٹرل ویسٹرن ڈے لائٹ ٹائم", "IST": "ہندوستان کا معیاری وقت", "HNNOMX": "شمال مغربی میکسیکو اسٹینڈرڈ ٹائم", "ART": "ارجنٹینا سٹینڈرڈ ٹائم", "AST": "اٹلانٹک اسٹینڈرڈ ٹائم", "GFT": "فرینچ گیانا کا وقت", "GMT": "گرین وچ کا اصل وقت", "UYT": "یوروگوئے کا معیاری وقت"},
	}
}

// Locale returns the current translators string locale
func (ur *ur) Locale() string {
	return ur.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'ur'
func (ur *ur) PluralsCardinal() []locales.PluralRule {
	return ur.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'ur'
func (ur *ur) PluralsOrdinal() []locales.PluralRule {
	return ur.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'ur'
func (ur *ur) PluralsRange() []locales.PluralRule {
	return ur.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'ur'
func (ur *ur) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)
	i := int64(n)

	if i == 1 && v == 0 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'ur'
func (ur *ur) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'ur'
func (ur *ur) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (ur *ur) MonthAbbreviated(month time.Month) string {
	return ur.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (ur *ur) MonthsAbbreviated() []string {
	return ur.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (ur *ur) MonthNarrow(month time.Month) string {
	return ur.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (ur *ur) MonthsNarrow() []string {
	return ur.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (ur *ur) MonthWide(month time.Month) string {
	return ur.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (ur *ur) MonthsWide() []string {
	return ur.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (ur *ur) WeekdayAbbreviated(weekday time.Weekday) string {
	return ur.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (ur *ur) WeekdaysAbbreviated() []string {
	return ur.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (ur *ur) WeekdayNarrow(weekday time.Weekday) string {
	return ur.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (ur *ur) WeekdaysNarrow() []string {
	return ur.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (ur *ur) WeekdayShort(weekday time.Weekday) string {
	return ur.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (ur *ur) WeekdaysShort() []string {
	return ur.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (ur *ur) WeekdayWide(weekday time.Weekday) string {
	return ur.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (ur *ur) WeekdaysWide() []string {
	return ur.daysWide
}

// Decimal returns the decimal point of number
func (ur *ur) Decimal() string {
	return ur.decimal
}

// Group returns the group of number
func (ur *ur) Group() string {
	return ur.group
}

// Group returns the minus sign of number
func (ur *ur) Minus() string {
	return ur.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'ur' and handles both Whole and Real numbers based on 'v'
func (ur *ur) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 5 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ur.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, ur.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		for j := len(ur.minus) - 1; j >= 0; j-- {
			b = append(b, ur.minus[j])
		}
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'ur' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (ur *ur) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 6
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ur.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		for j := len(ur.minus) - 1; j >= 0; j-- {
			b = append(b, ur.minus[j])
		}
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, ur.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'ur'
func (ur *ur) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := ur.currencies[currency]
	l := len(s) + len(symbol) + 5 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ur.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, ur.group[0])
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
		for j := len(ur.minus) - 1; j >= 0; j-- {
			b = append(b, ur.minus[j])
		}
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, ur.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'ur'
// in accounting notation.
func (ur *ur) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := ur.currencies[currency]
	l := len(s) + len(symbol) + 5 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ur.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, ur.group[0])
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

		for j := len(ur.minus) - 1; j >= 0; j-- {
			b = append(b, ur.minus[j])
		}

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
			b = append(b, ur.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'ur'
func (ur *ur) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'ur'
func (ur *ur) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, ur.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2c, 0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'ur'
func (ur *ur) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, ur.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2c, 0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'ur'
func (ur *ur) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, ur.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)
	b = append(b, ur.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2c, 0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'ur'
func (ur *ur) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, ur.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, ur.periodsAbbreviated[0]...)
	} else {
		b = append(b, ur.periodsAbbreviated[1]...)
	}

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'ur'
func (ur *ur) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, ur.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, ur.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, ur.periodsAbbreviated[0]...)
	} else {
		b = append(b, ur.periodsAbbreviated[1]...)
	}

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'ur'
func (ur *ur) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, ur.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, ur.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, ur.periodsAbbreviated[0]...)
	} else {
		b = append(b, ur.periodsAbbreviated[1]...)
	}

	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'ur'
func (ur *ur) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, ur.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, ur.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, ur.periodsAbbreviated[0]...)
	} else {
		b = append(b, ur.periodsAbbreviated[1]...)
	}

	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := ur.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
