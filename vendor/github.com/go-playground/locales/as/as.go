package as

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type as struct {
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

// New returns a new instance of translator for the 'as' locale
func New() locales.Translator {
	return &as{
		locale:                 "as",
		pluralsCardinal:        []locales.PluralRule{2, 6},
		pluralsOrdinal:         []locales.PluralRule{2, 3, 4, 5, 6},
		pluralsRange:           []locales.PluralRule{2, 6},
		decimal:                ".",
		group:                  ",",
		minus:                  "-",
		percent:                "%",
		perMille:               "‰",
		timeSeparator:          ":",
		inifinity:              "∞",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "A$", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "R$", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CA$", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CN¥", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "€", "FIM", "FJD", "FKP", "FRF", "£", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HK$", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "₪", "₹", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JP¥", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "₩", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MX$", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZ$", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "NT$", "TZS", "UAH", "UAK", "UGS", "UGX", "US$", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "₫", "VNN", "VUV", "WST", "FCFA", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "EC$", "XDR", "XEU", "XFO", "XFU", "CFA", "XPD", "CFPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyPositivePrefix: " ",
		currencyNegativePrefix: " ",
		monthsAbbreviated:      []string{"", "জানু", "ফেব্ৰু", "মাৰ্চ", "এপ্ৰিল", "মে’", "জুন", "জুলাই", "আগ", "ছেপ্তে", "অক্টো", "নৱে", "ডিচে"},
		monthsNarrow:           []string{"", "জ", "ফ", "ম", "এ", "ম", "জ", "জ", "আ", "ছ", "অ", "ন", "ড"},
		monthsWide:             []string{"", "জানুৱাৰী", "ফেব্ৰুৱাৰী", "মাৰ্চ", "এপ্ৰিল", "মে’", "জুন", "জুলাই", "আগষ্ট", "ছেপ্তেম্বৰ", "অক্টোবৰ", "নৱেম্বৰ", "ডিচেম্বৰ"},
		daysAbbreviated:        []string{"দেও", "সোম", "মঙ্গল", "বুধ", "বৃহ", "শুক্ৰ", "শনি"},
		daysNarrow:             []string{"দ", "স", "ম", "ব", "ব", "শ", "শ"},
		daysShort:              []string{"দেও", "সোম", "মঙ্গল", "বুধ", "বৃহ", "শুক্ৰ", "শনি"},
		daysWide:               []string{"দেওবাৰ", "সোমবাৰ", "মঙ্গলবাৰ", "বুধবাৰ", "বৃহস্পতিবাৰ", "শুক্ৰবাৰ", "শনিবাৰ"},
		periodsAbbreviated:     []string{"পূৰ্বাহ্ণ", "অপৰাহ্ণ"},
		periodsNarrow:          []string{"পূৰ্বাহ্ণ", "অপৰাহ্ণ"},
		periodsWide:            []string{"পূৰ্বাহ্ণ", "অপৰাহ্ণ"},
		erasAbbreviated:        []string{"", ""},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"খ্ৰীষ্টপূৰ্ব", "খ্ৰীষ্টাব্দ"},
		timezones:              map[string]string{"AEDT": "অস্ট্রেলিয়ান পূর্ব দিবালোক সময়", "BT": "ভুটান টাইম", "WARST": "ওয়েস্টার্ন আর্জেন্টিনা গ্রীষ্মকালীন সময়", "IST": "ভাৰতীয় সময়", "WITA": "মধ্য ইন্দোনেশিয়া সময়", "WEZ": "পশ্চিম ইউরোপীয় মান সময়", "JDT": "জাপান দিনের হালকা সময়", "HEEG": "HEEG", "HENOMX": "HENOMX", "OEZ": "পূর্ব ইউরোপীয় মান সময়", "COT": "কলম্বিয়া মান সময়", "CST": "CST", "SAST": "দক্ষিণ আফ্রিকা মান সময়", "SGT": "সিঙ্গাপুর স্ট্যান্ডার্ড টাইম", "ACWST": "অস্ট্রেলিয়ান সেন্ট্রাল ওয়েস্টার্ন স্ট্যান্ডার্ড টাইম", "LHDT": "লর্ড হ্যালো দিবালোক সময়", "TMST": "তুর্কমেনিস্তান গ্রীষ্ম সময়", "ART": "আৰ্জেণ্টিনা মান সময়", "PDT": "PDT", "AWST": "অস্ট্রেলিয়ান ওয়েস্টার্ন স্ট্যান্ডার্ড টাইম", "ADT": "ADT", "OESZ": "পূর্ব ইউরোপীয় গ্রীষ্মকালীন সময়", "BOT": "বলিভিয়া সময়", "ACST": "অস্ট্রেলিয়ান কেন্দ্রীয় স্ট্যান্ডার্ড টাইম", "ACDT": "অস্ট্রেলিয়ান কেন্দ্রীয় দিবালোক সময়", "CHAST": "চ্যাথাম স্ট্যান্ডার্ড টাইম", "∅∅∅": "ব্ৰাছিলিয়া গ্ৰীষ্ম সময়", "WIB": "ওয়েস্টার্ন ইন্দোনেশিয়া সময়", "AWDT": "অস্ট্রেলিয়ান ওয়েস্টার্ন ডেলাইট টাইম", "HNT": "HNT", "WIT": "ইস্টার্ন ইন্দোনেশিয়া সময়", "UYT": "উৰুগুৱে মান সময়", "HNCU": "HNCU", "HECU": "HECU", "PST": "PST", "HNNOMX": "HNNOMX", "ARST": "আৰ্জেণ্টিনা গ্ৰীষ্ম সময়", "WESZ": "পশ্চিম ইউরোপীয় গ্রীষ্মকালীন সময়", "NZST": "নিউজিল্যান্ড স্ট্যান্ডার্ড টাইম", "JST": "জাপান স্ট্যান্ডার্ড টাইম", "HKST": "হংকং গ্রীষ্মকালীন সময়", "LHST": "লর্ড হাভী স্ট্যান্ডার্ড টাইম", "CLT": "চিলি স্ট্যান্ডার্ড টাইম", "GYT": "গায়ানা টাইম", "ChST": "চামেরো স্ট্যান্ডার্ড টাইম", "HNPMX": "HNPMX", "WAST": "পশ্চিম আফ্রিকার গ্রীষ্মকালীন সময়", "VET": "ভেনিজুয়েলা সময়", "CAT": "মধ্য আফ্রিকা সময়", "CLST": "চিলি গ্রীষ্মকালীন সময়", "AKST": "AKST", "AKDT": "AKDT", "MEZ": "কেন্দ্রীয় ইউরোপীয় স্ট্যান্ডার্ড টাইম", "AEST": "অস্ট্রেলিয়ান ইস্টার্ন স্ট্যান্ডার্ড টাইম", "MYT": "মালয়েশিয়া সময়", "COST": "কলম্বিয়া গ্ৰীষ্ম সময়", "TMT": "তুর্কমেনিস্তান মান সময়", "CHADT": "চ্যাথাম ডেইলাইট টাইম", "AST": "AST", "EST": "EST", "GMT": "মক্কার সময়", "ECT": "ইকুৱেডৰ সময়", "EDT": "EDT", "HNPM": "HNPM", "GFT": "ফরাসি গায়ানা সময়", "HEPM": "HEPM", "CDT": "CDT", "MST": "MST", "MDT": "MDT", "HADT": "HADT", "HNEG": "HNEG", "HKT": "হংকং স্ট্যান্ডার্ড টাইম", "HAST": "HAST", "ACWDT": "অস্ট্রেলিয়ান সেন্ট্রাল ওয়েস্টার্ন ডেলাইট টাইম", "HNOG": "HNOG", "HEOG": "HEOG", "MESZ": "মধ্য ইউরোপীয় গ্রীষ্মকালীন সময়", "WART": "ওয়েস্টার্ন আর্জেন্টিনা মান সময়", "HEPMX": "HEPMX", "WAT": "পশ্চিম আফ্রিকার মান সময়", "NZDT": "নিউজিল্যান্ড ডেলাইট টাইম", "UYST": "উৰুগুৱে গ্ৰীষ্ম সময়", "HAT": "HAT", "SRT": "সুরিনাম টাইম", "EAT": "পূর্ব আফ্রিকা সময়"},
	}
}

// Locale returns the current translators string locale
func (as *as) Locale() string {
	return as.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'as'
func (as *as) PluralsCardinal() []locales.PluralRule {
	return as.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'as'
func (as *as) PluralsOrdinal() []locales.PluralRule {
	return as.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'as'
func (as *as) PluralsRange() []locales.PluralRule {
	return as.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'as'
func (as *as) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)
	i := int64(n)

	if (i == 0) || (n == 1) {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'as'
func (as *as) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 || n == 5 || n == 7 || n == 8 || n == 9 || n == 10 {
		return locales.PluralRuleOne
	} else if n == 2 || n == 3 {
		return locales.PluralRuleTwo
	} else if n == 4 {
		return locales.PluralRuleFew
	} else if n == 6 {
		return locales.PluralRuleMany
	}

	return locales.PluralRuleOther
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'as'
func (as *as) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {

	start := as.CardinalPluralRule(num1, v1)
	end := as.CardinalPluralRule(num2, v2)

	if start == locales.PluralRuleOne && end == locales.PluralRuleOne {
		return locales.PluralRuleOne
	} else if start == locales.PluralRuleOne && end == locales.PluralRuleOther {
		return locales.PluralRuleOther
	}

	return locales.PluralRuleOther

}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (as *as) MonthAbbreviated(month time.Month) string {
	return as.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (as *as) MonthsAbbreviated() []string {
	return as.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (as *as) MonthNarrow(month time.Month) string {
	return as.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (as *as) MonthsNarrow() []string {
	return as.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (as *as) MonthWide(month time.Month) string {
	return as.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (as *as) MonthsWide() []string {
	return as.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (as *as) WeekdayAbbreviated(weekday time.Weekday) string {
	return as.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (as *as) WeekdaysAbbreviated() []string {
	return as.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (as *as) WeekdayNarrow(weekday time.Weekday) string {
	return as.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (as *as) WeekdaysNarrow() []string {
	return as.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (as *as) WeekdayShort(weekday time.Weekday) string {
	return as.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (as *as) WeekdaysShort() []string {
	return as.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (as *as) WeekdayWide(weekday time.Weekday) string {
	return as.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (as *as) WeekdaysWide() []string {
	return as.daysWide
}

// Decimal returns the decimal point of number
func (as *as) Decimal() string {
	return as.decimal
}

// Group returns the group of number
func (as *as) Group() string {
	return as.group
}

// Group returns the minus sign of number
func (as *as) Minus() string {
	return as.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'as' and handles both Whole and Real numbers based on 'v'
func (as *as) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	inSecondary := false
	groupThreshold := 3

	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, as.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {

			if count == groupThreshold {
				b = append(b, as.group[0])
				count = 1

				if !inSecondary {
					inSecondary = true
					groupThreshold = 2
				}
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, as.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'as' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (as *as) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 3
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, as.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, as.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, as.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'as'
func (as *as) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := as.currencies[currency]
	l := len(s) + len(symbol) + 4 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	inSecondary := false
	groupThreshold := 3

	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, as.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {

			if count == groupThreshold {
				b = append(b, as.group[0])
				count = 1

				if !inSecondary {
					inSecondary = true
					groupThreshold = 2
				}
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	for j := len(symbol) - 1; j >= 0; j-- {
		b = append(b, symbol[j])
	}

	for j := len(as.currencyPositivePrefix) - 1; j >= 0; j-- {
		b = append(b, as.currencyPositivePrefix[j])
	}

	if num < 0 {
		b = append(b, as.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, as.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'as'
// in accounting notation.
func (as *as) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := as.currencies[currency]
	l := len(s) + len(symbol) + 4 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	inSecondary := false
	groupThreshold := 3

	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, as.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {

			if count == groupThreshold {
				b = append(b, as.group[0])
				count = 1

				if !inSecondary {
					inSecondary = true
					groupThreshold = 2
				}
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

		for j := len(as.currencyNegativePrefix) - 1; j >= 0; j-- {
			b = append(b, as.currencyNegativePrefix[j])
		}

		b = append(b, as.minus[0])

	} else {

		for j := len(symbol) - 1; j >= 0; j-- {
			b = append(b, symbol[j])
		}

		for j := len(as.currencyPositivePrefix) - 1; j >= 0; j-- {
			b = append(b, as.currencyPositivePrefix[j])
		}

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, as.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'as'
func (as *as) FmtDateShort(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2d}...)
	b = strconv.AppendInt(b, int64(t.Month()), 10)
	b = append(b, []byte{0x2d}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateMedium returns the medium date representation of 't' for 'as'
func (as *as) FmtDateMedium(t time.Time) string {

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

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'as'
func (as *as) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, as.monthsWide[t.Month()]...)
	b = append(b, []byte{0x2c, 0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'as'
func (as *as) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, as.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, as.monthsWide[t.Month()]...)
	b = append(b, []byte{0x2c, 0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'as'
func (as *as) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, []byte{0x2e}...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, []byte{0x2e, 0x20}...)

	if t.Hour() < 12 {
		b = append(b, as.periodsAbbreviated[0]...)
	} else {
		b = append(b, as.periodsAbbreviated[1]...)
	}

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'as'
func (as *as) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
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

	if t.Hour() < 12 {
		b = append(b, as.periodsAbbreviated[0]...)
	} else {
		b = append(b, as.periodsAbbreviated[1]...)
	}

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'as'
func (as *as) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
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

	if t.Hour() < 12 {
		b = append(b, as.periodsAbbreviated[0]...)
	} else {
		b = append(b, as.periodsAbbreviated[1]...)
	}

	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'as'
func (as *as) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
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

	if t.Hour() < 12 {
		b = append(b, as.periodsAbbreviated[0]...)
	} else {
		b = append(b, as.periodsAbbreviated[1]...)
	}

	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := as.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
