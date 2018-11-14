package sd_PK

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type sd_PK struct {
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
	currencyPositivePrefix string
	currencyNegativePrefix string
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

// New returns a new instance of translator for the 'sd_PK' locale
func New() locales.Translator {
	return &sd_PK{
		locale:                 "sd_PK",
		pluralsCardinal:        []locales.PluralRule{2, 6},
		pluralsOrdinal:         []locales.PluralRule{6},
		pluralsRange:           []locales.PluralRule{6},
		decimal:                ".",
		group:                  ",",
		minus:                  "-",
		percent:                "%",
		perMille:               "‰",
		timeSeparator:          ":",
		inifinity:              "∞",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyPositivePrefix: " ",
		currencyNegativePrefix: " ",
		monthsAbbreviated:      []string{"", "جنوري", "فيبروري", "مارچ", "اپريل", "مئي", "جون", "جولاءِ", "آگسٽ", "سيپٽمبر", "آڪٽوبر", "نومبر", "ڊسمبر"},
		monthsNarrow:           []string{"", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12"},
		monthsWide:             []string{"", "جنوري", "فيبروري", "مارچ", "اپريل", "مئي", "جون", "جولاءِ", "آگسٽ", "سيپٽمبر", "آڪٽوبر", "نومبر", "ڊسمبر"},
		daysAbbreviated:        []string{"آچر", "سومر", "اڱارو", "اربع", "خميس", "جمعو", "ڇنڇر"},
		daysNarrow:             []string{"آچر", "سو", "اڱارو", "اربع", "خم", "جمعو", "ڇنڇر"},
		daysShort:              []string{"آچر", "سومر", "اڱارو", "اربع", "خميس", "جمعو", "ڇنڇر"},
		daysWide:               []string{"آچر", "سومر", "اڱارو", "اربع", "خميس", "جمعو", "ڇنڇر"},
		periodsAbbreviated:     []string{"صبح، منجهند", "شام، منجهند"},
		periodsNarrow:          []string{"صبح، منجهند", "منجهند، شام"},
		periodsWide:            []string{"صبح، منجهند", "منجهند، شام"},
		erasAbbreviated:        []string{"BCE", "CE"},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"مسيح کان اڳ", "عيسوي کان پهرين"},
		timezones:              map[string]string{"COT": "ڪولمبيا جو معياري وقت", "CST": "مرڪزي معياري وقت", "AKST": "الاسڪا جو معياري وقت", "ACWDT": "آسٽريليا جو مرڪزي مغربي ڏينهن جو وقت", "LHDT": "لورڊ هووي جي ڏينهن جو وقت", "HEPM": "سینٽ پیئر و میڪوئیلون جي ڏينهن جو وقت", "WIT": "اوڀر انڊونيشيا جو وقت", "WAST": "اولهه آفريقا جي اونهاري جو وقت", "MYT": "ملائيشيا جو وقت", "HKST": "هانگ ڪانگ جي اونهاري جو وقت", "VET": "وينزويلا جو وقت", "PST": "پيسيفڪ معياري وقت", "∅∅∅": "∅∅∅", "HNEG": "مشرقي گرين لينڊ جو معياري وقت", "CAT": "مرڪزي آفريقا جو وقت", "HNPMX": "ميڪسيڪن پيسيفڪ جو معياري وقت", "PDT": "پيسيفڪ ڏينهن جو وقت", "AEDT": "آسٽريليا جو مشرقي ڏينهن جو وقت", "NZDT": "نيوزيلينڊ جي ڏينهن جو وقت", "ACWST": "آسٽريليا جو مرڪزي مغربي معياري وقت", "HAT": "نيو فائونڊ لينڊ جي ڏينهن جو وقت", "CLST": "چلي جي اونهاري جو وقت", "AKDT": "الاسڪا جي ڏينهن جو وقت", "ACDT": "آسٽريليا جو مرڪزي ڏينهن جو وقت", "WARST": "مغربي ارجنٽائن جي اونهاري جو وقت", "HENOMX": "شمالي مغربي ميڪسيڪو جي ڏينهن جو وقت", "TMST": "ترڪمانستان جي اونهاري جو وقت", "UYT": "يوروگائي جو معياري وقت", "UYST": "يوروگائي جي اونهاري جو وقت", "HNCU": "ڪيوبا جو معياري وقت", "CHADT": "چئٿم جي ڏينهن جو وقت", "CDT": "مرڪزي ڏينهن جو وقت", "AWDT": "آسٽريليا جو مغربي ڏينهن جو وقت", "BOT": "بولیویا جو وقت", "JDT": "جاپان جي ڏينهن جو وقت", "HNOG": "مغربي گرين لينڊ جو معياري وقت", "HKT": "هانگ ڪانگ جو معياري وقت", "IST": "ڀارت جو معياري وقت", "ADT": "ايٽلانٽڪ جي ڏينهن جو وقت", "AEST": "آسٽريليا جو مشرقي معياري وقت", "ACST": "آسٽريليا جو مرڪزي معياري وقت", "CLT": "چلي جو معياري وقت", "HECU": "ڪيوبا جي ڏينهن جو وقت", "TMT": "ترڪمانستان جو معياري وقت", "MDT": "پهاڙي ڏينهن جو وقت", "HEOG": "مغربي گرين لينڊ جي اونهاري جو وقت", "MEZ": "مرڪزي يورپي معياري وقت", "WART": "مغربي ارجنٽائن جو معياري وقت", "EAT": "اوڀر آفريڪا جو وقت", "AWST": "آسٽريليا جو مغربي معياري وقت", "JST": "جاپان جو معياري وقت", "EST": "مشرقي معياري وقت", "MESZ": "مرڪزي يورپي اونهاري جو وقت", "OEZ": "مشرقي يورپي معياري وقت", "COST": "ڪولمبيا جي اونهاري جو وقت", "BT": "ڀوٽان جو وقت", "WITA": "مرڪزي انڊونيشيا جو وقت", "ART": "ارجنٽائن معياري وقت", "WIB": "اولهه انڊونيشيا جو وقت", "ChST": "چمورو جو معياري وقت", "HEPMX": "ميڪسيڪن پيسيفڪ جي ڏينهن جو وقت", "LHST": "لورڊ هووي جو معياري وقت", "GYT": "گیانا جو وقت", "HAST": "هوائي اليوٽين جو معياري وقت", "CHAST": "چئٿم جو معياري وقت", "MST": "پهاڙي معياري وقت", "SGT": "سنگاپور جو معياري وقت", "HNPM": "سینٽ پیئر و میڪوئیلون جو معياري وقت", "SRT": "سوري نام جو وقت", "OESZ": "مشرقي يورپي اونهاري جو وقت", "ECT": "ايڪواڊور جو وقت", "HNT": "نيو فائونڊ لينڊ جو معياري وقت", "HNNOMX": "شمالي مغربي ميڪسيڪو جو معياري وقت", "AST": "ايٽلانٽڪ جو معياري وقت", "SAST": "ڏکڻ آفريڪا جو معياري وقت", "WAT": "اولهه آفريقا جو معياري وقت", "WEZ": "مغربي يورپي معياري وقت", "WESZ": "مغربي يورپي ڏينهن جو وقت", "HADT": "هوائي اليوٽين جي ڏينهن جو وقت", "ARST": "ارجنٽائن جي اونهاري جو وقت", "GMT": "گرين وچ مين ٽائيم", "NZST": "نيوزيلينڊ جو معياري وقت", "GFT": "فرانسيسي گيانا جو وقت", "EDT": "مشرقي ڏينهن جو وقت", "HEEG": "مشرقي گرين لينڊ جي اونهاري جو وقت"},
	}
}

// Locale returns the current translators string locale
func (sd *sd_PK) Locale() string {
	return sd.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'sd_PK'
func (sd *sd_PK) PluralsCardinal() []locales.PluralRule {
	return sd.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'sd_PK'
func (sd *sd_PK) PluralsOrdinal() []locales.PluralRule {
	return sd.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'sd_PK'
func (sd *sd_PK) PluralsRange() []locales.PluralRule {
	return sd.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'sd_PK'
func (sd *sd_PK) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'sd_PK'
func (sd *sd_PK) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'sd_PK'
func (sd *sd_PK) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (sd *sd_PK) MonthAbbreviated(month time.Month) string {
	return sd.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (sd *sd_PK) MonthsAbbreviated() []string {
	return sd.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (sd *sd_PK) MonthNarrow(month time.Month) string {
	return sd.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (sd *sd_PK) MonthsNarrow() []string {
	return sd.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (sd *sd_PK) MonthWide(month time.Month) string {
	return sd.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (sd *sd_PK) MonthsWide() []string {
	return sd.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (sd *sd_PK) WeekdayAbbreviated(weekday time.Weekday) string {
	return sd.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (sd *sd_PK) WeekdaysAbbreviated() []string {
	return sd.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (sd *sd_PK) WeekdayNarrow(weekday time.Weekday) string {
	return sd.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (sd *sd_PK) WeekdaysNarrow() []string {
	return sd.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (sd *sd_PK) WeekdayShort(weekday time.Weekday) string {
	return sd.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (sd *sd_PK) WeekdaysShort() []string {
	return sd.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (sd *sd_PK) WeekdayWide(weekday time.Weekday) string {
	return sd.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (sd *sd_PK) WeekdaysWide() []string {
	return sd.daysWide
}

// Decimal returns the decimal point of number
func (sd *sd_PK) Decimal() string {
	return sd.decimal
}

// Group returns the group of number
func (sd *sd_PK) Group() string {
	return sd.group
}

// Group returns the minus sign of number
func (sd *sd_PK) Minus() string {
	return sd.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'sd_PK' and handles both Whole and Real numbers based on 'v'
func (sd *sd_PK) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, sd.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, sd.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, sd.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'sd_PK' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (sd *sd_PK) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 3
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, sd.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, sd.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, sd.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'sd_PK'
func (sd *sd_PK) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := sd.currencies[currency]
	l := len(s) + len(symbol) + 4 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, sd.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, sd.group[0])
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

	for j := len(sd.currencyPositivePrefix) - 1; j >= 0; j-- {
		b = append(b, sd.currencyPositivePrefix[j])
	}

	if num < 0 {
		b = append(b, sd.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, sd.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'sd_PK'
// in accounting notation.
func (sd *sd_PK) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := sd.currencies[currency]
	l := len(s) + len(symbol) + 4 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, sd.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, sd.group[0])
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

		for j := len(sd.currencyNegativePrefix) - 1; j >= 0; j-- {
			b = append(b, sd.currencyNegativePrefix[j])
		}

		b = append(b, sd.minus[0])

	} else {

		for j := len(symbol) - 1; j >= 0; j-- {
			b = append(b, symbol[j])
		}

		for j := len(sd.currencyPositivePrefix) - 1; j >= 0; j-- {
			b = append(b, sd.currencyPositivePrefix[j])
		}

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, sd.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'sd_PK'
func (sd *sd_PK) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'sd_PK'
func (sd *sd_PK) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x20}...)
	b = append(b, sd.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'sd_PK'
func (sd *sd_PK) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x20}...)
	b = append(b, sd.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'sd_PK'
func (sd *sd_PK) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x20}...)
	b = append(b, sd.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2c, 0x20}...)
	b = append(b, sd.daysWide[t.Weekday()]...)

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'sd_PK'
func (sd *sd_PK) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, sd.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, sd.periodsAbbreviated[0]...)
	} else {
		b = append(b, sd.periodsAbbreviated[1]...)
	}

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'sd_PK'
func (sd *sd_PK) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, sd.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, sd.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, sd.periodsAbbreviated[0]...)
	} else {
		b = append(b, sd.periodsAbbreviated[1]...)
	}

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'sd_PK'
func (sd *sd_PK) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, sd.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, sd.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, sd.periodsAbbreviated[0]...)
	} else {
		b = append(b, sd.periodsAbbreviated[1]...)
	}

	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'sd_PK'
func (sd *sd_PK) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, sd.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, sd.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, sd.periodsAbbreviated[0]...)
	} else {
		b = append(b, sd.periodsAbbreviated[1]...)
	}

	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := sd.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
