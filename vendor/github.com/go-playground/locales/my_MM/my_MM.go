package my_MM

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type my_MM struct {
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

// New returns a new instance of translator for the 'my_MM' locale
func New() locales.Translator {
	return &my_MM{
		locale:                 "my_MM",
		pluralsCardinal:        []locales.PluralRule{6},
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
		monthsAbbreviated:      []string{"", "ဇန်", "ဖေ", "မတ်", "ဧ", "မေ", "ဇွန်", "ဇူ", "ဩ", "စက်", "အောက်", "နို", "ဒီ"},
		monthsNarrow:           []string{"", "ဇ", "ဖ", "မ", "ဧ", "မ", "ဇ", "ဇ", "ဩ", "စ", "အ", "န", "ဒ"},
		monthsWide:             []string{"", "ဇန်နဝါရီ", "ဖေဖော်ဝါရီ", "မတ်", "ဧပြီ", "မေ", "ဇွန်", "ဇူလိုင်", "ဩဂုတ်", "စက်တင်ဘာ", "အောက်တိုဘာ", "နိုဝင်ဘာ", "ဒီဇင်ဘာ"},
		daysAbbreviated:        []string{"တနင်္ဂနွေ", "တနင်္လာ", "အင်္ဂါ", "ဗုဒ္ဓဟူး", "ကြာသပတေး", "သောကြာ", "စနေ"},
		daysNarrow:             []string{"တ", "တ", "အ", "ဗ", "က", "သ", "စ"},
		daysShort:              []string{"တနင်္ဂနွေ", "တနင်္လာ", "အင်္ဂါ", "ဗုဒ္ဓဟူး", "ကြာသပတေး", "သောကြာ", "စနေ"},
		daysWide:               []string{"တနင်္ဂနွေ", "တနင်္လာ", "အင်္ဂါ", "ဗုဒ္ဓဟူး", "ကြာသပတေး", "သောကြာ", "စနေ"},
		periodsAbbreviated:     []string{"နံနက်", "ညနေ"},
		periodsNarrow:          []string{"နံနက်", "ညနေ"},
		periodsWide:            []string{"နံနက်", "ညနေ"},
		erasAbbreviated:        []string{"ဘီစီ", "အေဒီ"},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"ခရစ်တော် မပေါ်မီနှစ်", "ခရစ်နှစ်"},
		timezones:              map[string]string{"MST": "မြောက်အမေရိက တောင်တန်းဒေသစံတော်ချိန်", "BOT": "ဘိုလီးဘီးယား အချိန်", "JDT": "ဂျပန် နွေရာသီ အချိန်", "HENOMX": "အနောက်တောင် မက္ကစီကို နွေရာသီ စံတော်ချိန်", "GMT": "ဂရင်းနစ် စံတော်ချိန်", "EDT": "မြောက်အမေရိက အရှေ့ပိုင်း နွေရာသီစံတော်ချိန်", "HKST": "ဟောင်ကောင် နွေရာသီ အချိန်", "WIT": "အရှေ့ပိုင်း အင်ဒိုနီးရှား အချိန်", "OESZ": "အရှေ့ဥရောပ နွေရာသီ အချိန်", "AWDT": "ဩစတြေးလျ နွေရာသီ အချိန်", "AEDT": "အရှေ့ဩစတြေးလျ နွေရာသီ အချိန်", "ECT": "အီကွေဒေါ အချိန်", "ACWDT": "သြစတြေးလျား အနောက်အလယ်ပိုင်း နွေရာသီ အချိန်", "HEPM": "စိန့်ပီအဲနှင့် မီခွီလွန် နွေရာသီ စံတော်ချိန်", "HNT": "နယူးဖောင်လန် စံတော်ချိန်", "CAT": "အလယ်အာဖရိက အချိန်", "GYT": "ဂိုင်ယာနာ အချိန်", "AKST": "အလာစကာ စံတော်ချိန်", "MEZ": "ဥရောပအလယ်ပိုင်း စံတော်ချိန်", "HECU": "ကျူးဘား နွေရာသီ စံတော်ချိန်", "WAT": "အနောက်အာဖရိက စံတော်ချိန်", "WAST": "အနောက်အာဖရိက နွေရာသီ အချိန်", "SAST": "တောင်အာဖရိက အချိန်", "BT": "ဘူတန် အချိန်", "SGT": "စင်္ကာပူ အချိန်", "LHDT": "လော့ဒ်ဟောင် နွေရာသီ အချိန်", "WART": "အနောက် အာဂျင်တီးနား စံတော်ချိန်", "CLT": "ချီလီ စံတော်ချိန်", "ART": "အာဂျင်တီးနား စံတော်ချိန်", "AEST": "အရှေ့ဩစတြေးလျ စံတော်ချိန်", "MYT": "မလေးရှား အချိန်", "ACST": "ဩစတြေးလျ အလယ်ပိုင်း စံတော်ချိန်", "TMT": "တာ့ခ်မင်နစ္စတန် စံတော်ချိန်", "HNCU": "ကျူးဘား စံတော်ချိန်", "AKDT": "အလာစကာ နွေရာသီ စံတော်ချိန်", "LHST": "လော့ဒ်ဟောင် စံတော်ချိန်", "WARST": "အနောက် အာဂျင်တီးနား နွေရာသီ အချိန်", "TMST": "တာ့ခ်မင်နစ္စတန် နွေရာသီ အချိန်", "HADT": "ဟာဝိုင်ယီ အယ်လူးရှန်း နွေရာသီ စံတော်ချိန်", "AWST": "အနောက်ဩစတြေးလျ စံတော်ချိန်", "GFT": "ပြင်သစ် ဂီအားနား အချိန်", "HNOG": "အနောက် ဂရင်းလန်း စံတော်ချိန်", "HEOG": "အနောက် ဂရင်းလန် နွေရာသီ စံတော်ချိန်", "WITA": "အလယ်ပိုင်း အင်ဒိုနီးရှား အချိန်", "HAT": "နယူးဖောင်လန် နွေရာသီ စံတော်ချိန်", "COST": "ကိုလံဘီယာ နွေရာသီ အချိန်", "PST": "မြောက်အမေရိက ပစိဖိတ်စံတော်ချိန်", "CST": "မြောက်အမေရိက အလယ်ပိုင်းစံတော်ချိန်", "WEZ": "အနောက်ဥရောပ စံတော်ချိန်", "WIB": "အနောက်ပိုင်း အင်ဒိုနီးရှား အချိန်", "NZDT": "နယူးဇီလန် နွေရာသီ အချိန်", "EAT": "အရှေ့အာဖရိက အချိန်", "CDT": "မြောက်အမေရိက အလယ်ပိုင်း နွေရာသီစံတော်ချိန်", "ACDT": "ဩစတြေးလျ အလယ်ပိုင်း နွေရာသီ အချိန်", "COT": "ကိုလံဘီယာ စံတော်ချိန်", "OEZ": "အရှေ့ဥရောပ စံတော်ချိန်", "UYT": "ဥရုဂွေး စံတော်ချိန်", "ChST": "ချာမိုရို အချိန်", "WESZ": "အနောက်ဥရောပ နွေရာသီ အချိန်", "JST": "ဂျပန် စံတော်ချိန်", "ACWST": "သြစတြေးလျား အနောက်အလယ်ပိုင်း စံတော်ချိန်", "HKT": "ဟောင်ကောင် စံတော်ချိန်", "SRT": "စူးရီနာမ်အချိန်", "UYST": "ဥရုဂွေး နွေရာသီ အချိန်", "HEPMX": "မက္ကစီကန် ပစိဖိတ် နွေရာသီ စံတော်ချိန်", "MDT": "မြောက်အမေရိက တောင်တန်းဒေသ နွေရာသီစံတော်ချိန်", "EST": "မြောက်အမေရိက အရှေ့ပိုင်းစံတော်ချိန်", "ARST": "အာဂျင်တီးနား နွေရာသီအချိန်", "NZST": "နယူးဇီလန် စံတော်ချိန်", "VET": "ဗင်နီဇွဲလား အချိန်", "HNPM": "စိန့်ပီအဲနှင့်မီခွီလွန်စံတော်ချိန်", "CHAST": "ချားသမ်စံတော်ချိန်", "CHADT": "ချာသမ် နွေရာသီ အချိန်", "PDT": "မြောက်အမေရိက ပစိဖိတ် နွေရာသီစံတော်ချိန်", "ADT": "အတ္တလန်တစ် နွေရာသီ စံတော်ချိန်", "MESZ": "ဥရောပအလယ်ပိုင်း နွေရာသီ အချိန်", "CLST": "ချီလီ နွေရာသီ အချိန်", "∅∅∅": "ဘရာဇီး နွေရာသီ အချိန်", "HNPMX": "မက္ကဆီကန် ပစိဖိတ် စံတော်ချိန်", "AST": "အတ္တလန်တစ် စံတော်ချိန်", "HEEG": "အရှေ့ဂရင်းလန် နွေရာသီ စံတော်ချိန်", "IST": "အိန္ဒိယ စံတော်ချိန်", "HNNOMX": "အနောက်တောင် မက္ကဆီကို စံတော်ချိန်", "HNEG": "အရှေ့ဂရင်းလန်း စံတော်ချိန်", "HAST": "ဟာဝိုင်ယီ အယ်လူးရှန်း စံတော်ချိန်"},
	}
}

// Locale returns the current translators string locale
func (my *my_MM) Locale() string {
	return my.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'my_MM'
func (my *my_MM) PluralsCardinal() []locales.PluralRule {
	return my.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'my_MM'
func (my *my_MM) PluralsOrdinal() []locales.PluralRule {
	return my.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'my_MM'
func (my *my_MM) PluralsRange() []locales.PluralRule {
	return my.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'my_MM'
func (my *my_MM) CardinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'my_MM'
func (my *my_MM) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'my_MM'
func (my *my_MM) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (my *my_MM) MonthAbbreviated(month time.Month) string {
	return my.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (my *my_MM) MonthsAbbreviated() []string {
	return my.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (my *my_MM) MonthNarrow(month time.Month) string {
	return my.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (my *my_MM) MonthsNarrow() []string {
	return my.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (my *my_MM) MonthWide(month time.Month) string {
	return my.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (my *my_MM) MonthsWide() []string {
	return my.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (my *my_MM) WeekdayAbbreviated(weekday time.Weekday) string {
	return my.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (my *my_MM) WeekdaysAbbreviated() []string {
	return my.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (my *my_MM) WeekdayNarrow(weekday time.Weekday) string {
	return my.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (my *my_MM) WeekdaysNarrow() []string {
	return my.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (my *my_MM) WeekdayShort(weekday time.Weekday) string {
	return my.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (my *my_MM) WeekdaysShort() []string {
	return my.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (my *my_MM) WeekdayWide(weekday time.Weekday) string {
	return my.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (my *my_MM) WeekdaysWide() []string {
	return my.daysWide
}

// Decimal returns the decimal point of number
func (my *my_MM) Decimal() string {
	return my.decimal
}

// Group returns the group of number
func (my *my_MM) Group() string {
	return my.group
}

// Group returns the minus sign of number
func (my *my_MM) Minus() string {
	return my.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'my_MM' and handles both Whole and Real numbers based on 'v'
func (my *my_MM) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, my.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, my.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, my.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'my_MM' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (my *my_MM) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 3
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, my.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, my.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, my.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'my_MM'
func (my *my_MM) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := my.currencies[currency]
	l := len(s) + len(symbol) + 4 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, my.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, my.group[0])
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

	for j := len(my.currencyPositivePrefix) - 1; j >= 0; j-- {
		b = append(b, my.currencyPositivePrefix[j])
	}

	if num < 0 {
		b = append(b, my.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, my.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'my_MM'
// in accounting notation.
func (my *my_MM) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := my.currencies[currency]
	l := len(s) + len(symbol) + 4 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, my.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, my.group[0])
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

		for j := len(my.currencyNegativePrefix) - 1; j >= 0; j-- {
			b = append(b, my.currencyNegativePrefix[j])
		}

		b = append(b, my.minus[0])

	} else {

		for j := len(symbol) - 1; j >= 0; j-- {
			b = append(b, symbol[j])
		}

		for j := len(my.currencyPositivePrefix) - 1; j >= 0; j-- {
			b = append(b, my.currencyPositivePrefix[j])
		}

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, my.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'my_MM'
func (my *my_MM) FmtDateShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Day() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2d}...)

	if t.Month() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Month()), 10)

	b = append(b, []byte{0x2d}...)

	if t.Year() > 9 {
		b = append(b, strconv.Itoa(t.Year())[2:]...)
	} else {
		b = append(b, strconv.Itoa(t.Year())[1:]...)
	}

	return string(b)
}

// FmtDateMedium returns the medium date representation of 't' for 'my_MM'
func (my *my_MM) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0xe1, 0x81, 0x8a, 0x20}...)
	b = append(b, my.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'my_MM'
func (my *my_MM) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0xe1, 0x81, 0x8a, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, my.monthsWide[t.Month()]...)

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'my_MM'
func (my *my_MM) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0xe1, 0x81, 0x8a, 0x20}...)
	b = append(b, my.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0xe1, 0x81, 0x8a, 0x20}...)
	b = append(b, my.daysWide[t.Weekday()]...)

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'my_MM'
func (my *my_MM) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, []byte{0x42, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, my.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'my_MM'
func (my *my_MM) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, []byte{0x42, 0x20}...)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, my.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, my.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'my_MM'
func (my *my_MM) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	tz, _ := t.Zone()
	b = append(b, tz...)

	b = append(b, []byte{0x20}...)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, my.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, my.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'my_MM'
func (my *my_MM) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	tz, _ := t.Zone()

	if btz, ok := my.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	b = append(b, []byte{0x20}...)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, my.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, my.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}
