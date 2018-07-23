package ur_PK

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type ur_PK struct {
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

// New returns a new instance of translator for the 'ur_PK' locale
func New() locales.Translator {
	return &ur_PK{
		locale:             "ur_PK",
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
		currencies:         []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
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
		timezones:          map[string]string{"GFT": "فرینچ گیانا کا وقت", "MST": "ماؤنٹین اسٹینڈرڈ ٹائم", "BOT": "بولیویا کا وقت", "EST": "ایسٹرن اسٹینڈرڈ ٹائم", "CHAST": "چیتھم اسٹینڈرڈ ٹائم", "HAT": "نیو فاؤنڈ لینڈ ڈے لائٹ ٹائم", "HNNOMX": "شمال مغربی میکسیکو اسٹینڈرڈ ٹائم", "HADT": "ہوائی الیوٹیئن ڈے لائٹ ٹائم", "PDT": "پیسفک ڈے لائٹ ٹائم", "COST": "کولمبیا کا موسم گرما کا وقت", "MDT": "ماؤنٹین ڈے لائٹ ٹائم", "WAT": "مغربی افریقہ سٹینڈرڈ ٹائم", "ACWDT": "آسٹریلین سنٹرل ویسٹرن ڈے لائٹ ٹائم", "EDT": "ایسٹرن ڈے لائٹ ٹائم", "WIT": "مشرقی انڈونیشیا ٹائم", "HENOMX": "شمال مغربی میکسیکو ڈے لائٹ ٹائم", "UYT": "یوروگوئے کا معیاری وقت", "CST": "سنٹرل اسٹینڈرڈ ٹائم", "ADT": "اٹلانٹک ڈے لائٹ ٹائم", "WESZ": "مغربی یورپ کا موسم گرما کا وقت", "ACDT": "آسٹریلین سنٹرل ڈے لائٹ ٹائم", "ACWST": "آسٹریلین سنٹرل ویسٹرن اسٹینڈرڈ ٹائم", "HKST": "ہانگ کانگ سمر ٹائم", "GYT": "گیانا کا وقت", "HEPMX": "میکسیکن پیسفک ڈے لائٹ ٹائم", "AEST": "آسٹریلین ایسٹرن اسٹینڈرڈ ٹائم", "HEOG": "مغربی گرین لینڈ کا موسم گرما کا وقت", "IST": "ہندوستان کا معیاری وقت", "WITA": "وسطی انڈونیشیا ٹائم", "ARST": "ارجنٹینا سمر ٹائم", "MYT": "ملیشیا ٹائم", "HAST": "ہوائی الیوٹیئن اسٹینڈرڈ ٹائم", "OEZ": "مشرقی یورپ کا معیاری وقت", "GMT": "گرین وچ کا اصل وقت", "CHADT": "چیتھم ڈے لائٹ ٹائم", "EAT": "مشرقی افریقہ ٹائم", "TMST": "ترکمانستان کا موسم گرما کا وقت", "CAT": "وسطی افریقہ ٹائم", "UYST": "یوروگوئے کا موسم گرما کا وقت", "OESZ": "مشرقی یورپ کا موسم گرما کا وقت", "AWST": "آسٹریلیا ویسٹرن اسٹینڈرڈ ٹائم", "AWDT": "آسٹریلین ویسٹرن ڈے لائٹ ٹائم", "WEZ": "مغربی یورپ کا معیاری وقت", "WIB": "مغربی انڈونیشیا ٹائم", "NZDT": "نیوزی لینڈ ڈے لائٹ ٹائم", "ECT": "ایکواڈور کا وقت", "LHST": "لارڈ ہووے اسٹینڈرڈ ٹائم", "HNPMX": "میکسیکن پیسفک اسٹینڈرڈ ٹائم", "JST": "جاپان سٹینڈرڈ ٹائم", "HNEG": "مشرقی گرین لینڈ اسٹینڈرڈ ٹائم", "HNOG": "مغربی گرین لینڈ اسٹینڈرڈ ٹائم", "HEPM": "سینٹ پیئر اور مکلیئون ڈے لائٹ ٹائم", "TMT": "ترکمانستان کا معیاری وقت", "AST": "اٹلانٹک اسٹینڈرڈ ٹائم", "AKDT": "الاسکا ڈے لائٹ ٹائم", "WAST": "مغربی افریقہ سمر ٹائم", "HNPM": "سینٹ پیئر اور مکلیئون اسٹینڈرڈ ٹائم", "HECU": "کیوبا ڈے لائٹ ٹائم", "PST": "پیسفک اسٹینڈرڈ ٹائم", "MESZ": "وسطی یورپ کا موسم گرما کا وقت", "LHDT": "لارڈ ہووے ڈے لائٹ ٹائم", "VET": "وینزوئیلا کا وقت", "HNT": "نیو فاؤنڈ لینڈ اسٹینڈرڈ ٹائم", "CLT": "چلی کا معیاری وقت", "CLST": "چلی کا موسم گرما کا وقت", "ChST": "چامورو سٹینڈرڈ ٹائم", "HNCU": "کیوبا اسٹینڈرڈ ٹائم", "∅∅∅": "∅∅∅", "NZST": "نیوزی لینڈ کا معیاری وقت", "AKST": "الاسکا اسٹینڈرڈ ٹائم", "ACST": "آسٹریلین سنٹرل اسٹینڈرڈ ٹائم", "HEEG": "مشرقی گرین لینڈ کا موسم گرما کا وقت", "CDT": "سنٹرل ڈے لائٹ ٹائم", "JDT": "جاپان ڈے لائٹ ٹائم", "SGT": "سنگاپور سٹینڈرڈ ٹائم", "MEZ": "وسطی یورپ کا معیاری وقت", "SRT": "سورینام کا وقت", "COT": "کولمبیا کا معیاری وقت", "WARST": "مغربی ارجنٹینا کا موسم گرما کا وقت", "ART": "ارجنٹینا سٹینڈرڈ ٹائم", "AEDT": "آسٹریلین ایسٹرن ڈے لائٹ ٹائم", "SAST": "جنوبی افریقہ سٹینڈرڈ ٹائم", "BT": "بھوٹان کا وقت", "HKT": "ہانگ کانگ سٹینڈرڈ ٹائم", "WART": "مغربی ارجنٹینا کا معیاری وقت"},
	}
}

// Locale returns the current translators string locale
func (ur *ur_PK) Locale() string {
	return ur.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'ur_PK'
func (ur *ur_PK) PluralsCardinal() []locales.PluralRule {
	return ur.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'ur_PK'
func (ur *ur_PK) PluralsOrdinal() []locales.PluralRule {
	return ur.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'ur_PK'
func (ur *ur_PK) PluralsRange() []locales.PluralRule {
	return ur.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'ur_PK'
func (ur *ur_PK) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)
	i := int64(n)

	if i == 1 && v == 0 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'ur_PK'
func (ur *ur_PK) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'ur_PK'
func (ur *ur_PK) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (ur *ur_PK) MonthAbbreviated(month time.Month) string {
	return ur.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (ur *ur_PK) MonthsAbbreviated() []string {
	return ur.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (ur *ur_PK) MonthNarrow(month time.Month) string {
	return ur.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (ur *ur_PK) MonthsNarrow() []string {
	return ur.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (ur *ur_PK) MonthWide(month time.Month) string {
	return ur.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (ur *ur_PK) MonthsWide() []string {
	return ur.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (ur *ur_PK) WeekdayAbbreviated(weekday time.Weekday) string {
	return ur.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (ur *ur_PK) WeekdaysAbbreviated() []string {
	return ur.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (ur *ur_PK) WeekdayNarrow(weekday time.Weekday) string {
	return ur.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (ur *ur_PK) WeekdaysNarrow() []string {
	return ur.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (ur *ur_PK) WeekdayShort(weekday time.Weekday) string {
	return ur.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (ur *ur_PK) WeekdaysShort() []string {
	return ur.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (ur *ur_PK) WeekdayWide(weekday time.Weekday) string {
	return ur.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (ur *ur_PK) WeekdaysWide() []string {
	return ur.daysWide
}

// Decimal returns the decimal point of number
func (ur *ur_PK) Decimal() string {
	return ur.decimal
}

// Group returns the group of number
func (ur *ur_PK) Group() string {
	return ur.group
}

// Group returns the minus sign of number
func (ur *ur_PK) Minus() string {
	return ur.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'ur_PK' and handles both Whole and Real numbers based on 'v'
func (ur *ur_PK) FmtNumber(num float64, v uint64) string {

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

// FmtPercent returns 'num' with digits/precision of 'v' for 'ur_PK' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (ur *ur_PK) FmtPercent(num float64, v uint64) string {
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

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'ur_PK'
func (ur *ur_PK) FmtCurrency(num float64, v uint64, currency currency.Type) string {

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

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'ur_PK'
// in accounting notation.
func (ur *ur_PK) FmtAccounting(num float64, v uint64, currency currency.Type) string {

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

// FmtDateShort returns the short date representation of 't' for 'ur_PK'
func (ur *ur_PK) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'ur_PK'
func (ur *ur_PK) FmtDateMedium(t time.Time) string {

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

// FmtDateLong returns the long date representation of 't' for 'ur_PK'
func (ur *ur_PK) FmtDateLong(t time.Time) string {

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

// FmtDateFull returns the full date representation of 't' for 'ur_PK'
func (ur *ur_PK) FmtDateFull(t time.Time) string {

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

// FmtTimeShort returns the short time representation of 't' for 'ur_PK'
func (ur *ur_PK) FmtTimeShort(t time.Time) string {

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

// FmtTimeMedium returns the medium time representation of 't' for 'ur_PK'
func (ur *ur_PK) FmtTimeMedium(t time.Time) string {

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

// FmtTimeLong returns the long time representation of 't' for 'ur_PK'
func (ur *ur_PK) FmtTimeLong(t time.Time) string {

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

// FmtTimeFull returns the full time representation of 't' for 'ur_PK'
func (ur *ur_PK) FmtTimeFull(t time.Time) string {

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
