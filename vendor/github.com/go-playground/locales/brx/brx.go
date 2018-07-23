package brx

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type brx struct {
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

// New returns a new instance of translator for the 'brx' locale
func New() locales.Translator {
	return &brx{
		locale:                 "brx",
		pluralsCardinal:        []locales.PluralRule{2, 6},
		pluralsOrdinal:         nil,
		pluralsRange:           nil,
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
		monthsNarrow:           []string{"", "ज", "फे", "मा", "ए", "मे", "जु", "जु", "आ", "से", "अ", "न", "दि"},
		monthsWide:             []string{"", "जानुवारी", "फेब्रुवारी", "मार्स", "एफ्रिल", "मे", "जुन", "जुलाइ", "आगस्थ", "सेबथेज्ब़र", "अखथबर", "नबेज्ब़र", "दिसेज्ब़र"},
		daysAbbreviated:        []string{"रबि", "सम", "मंगल", "बुद", "बिसथि", "सुखुर", "सुनि"},
		daysNarrow:             []string{"र", "स", "मं", "बु", "बि", "सु", "सु"},
		daysWide:               []string{"रबिबार", "समबार", "मंगलबार", "बुदबार", "बिसथिबार", "सुखुरबार", "सुनिबार"},
		periodsAbbreviated:     []string{"फुं", "बेलासे"},
		periodsWide:            []string{"फुं", "बेलासे"},
		erasAbbreviated:        []string{"ईसा.पूर्व", "सन"},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"", ""},
		timezones:              map[string]string{"HNEG": "ग्रीनलैण्ड ईस्टर्न स्टैंडर्ड टाईम", "HAST": "हवाई आलटन स्टैंडर्ड टाईम", "PST": "पैसीफीक स्टैंडर्ड टाईम", "WAST": "पश्चीम अफ्रीका समर टाईम", "AKST": "अलास्का स्टैंडर्ड टाईम", "MST": "माकाऊ स्टैंडर्ड टाईम", "WART": "पश्चीम अर्जण्टिना स्टैंडर्ड टाईम", "HENOMX": "HENOMX", "HNPMX": "HNPMX", "ECT": "एक्वाडौर स्टैंडर्ड टाईम", "AEST": "पूर्वी ओस्ट्रेलिया स्टैंडर्ड टाईम", "MDT": "माकाऊ समर टाईम", "CLT": "चीली स्टैंडर्ड टाईम", "TMT": "तुर्कमेनीस्तान स्टैंडर्ड टाईम", "GYT": "गुयाना स्टैंडर्ड टाईम", "CAT": "मध्य अफ्रीका स्टैंडर्ड टाईम", "OESZ": "ईस्टर्न यूरोप समर टाईम", "UYST": "ऊरुगुए समर टाईम", "HNOG": "ग्रीनलैण्ड वेस्टर्न स्टैंडर्ड टाईम", "HNNOMX": "HNNOMX", "UYT": "ऊरुगुए स्टैंडर्ड टाईम", "ChST": "चामरो स्टैंडर्ड टाईम", "WESZ": "वेस्टर्न यूरोप समर टाईम", "LHST": "लार्ड़ होव स्टैंडर्ड टाईम", "VET": "वेनेज़ुएला स्टैंडर्ड टाईम", "CHADT": "चैथम डेलाईट टाईम", "CDT": "सैंट्रल अमरिका डेलाईट टाईम", "PDT": "पैसीफीक डेलाईट टाईम", "ACWDT": "मध्य-पश्चीम ओस्ट्रेलिया डेलाईट टाईम", "HEOG": "ग्रीनलैण्ड वेस्टर्न समर टाईम", "EDT": "ईस्टर्न अमरिका डेलाईट टाईम", "IST": "भारतीय स्टैंडर्ड टाईम", "SRT": "सुरीनाम स्टैंडर्ड टाईम", "AST": "अटलांटीक स्टैंडर्ड टाईम", "ACWST": "मध्य-पश्चीम ओस्ट्रेलिया स्टैंडर्ड टाईम", "HEEG": "ग्रीनलैण्ड ईस्टर्न समर टाईम", "NZST": "न्युज़ीलैण्ड स्टैंडर्ड टाईम", "WIT": "ईस्टर्न ईंडोनीशिया स्टैंडर्ड टाईम", "TMST": "तुर्कमेनीस्तान समर टाईम", "JDT": "जपान डेलाईट टाईम", "BT": "भुटान स्टैंडर्ड टाईम", "HEPM": "सेँ पीयॅर एवं मीकलों डेलाईट टाईम", "OEZ": "ईस्टर्न यूरोप स्टैंडर्ड टाईम", "GMT": "ग्रीनीच स्टैंडर्ड टाईम", "AWDT": "दक्षिण ओस्ट्रेलिया डेलाईट टाईम", "WEZ": "वेस्टर्न यूरोप स्टैंडर्ड टाईम", "MYT": "मलेशिया स्टैंडर्ड टाईम", "JST": "जपान स्टैंडर्ड टाईम", "AKDT": "अलास्का डेलाईट टाईम", "SGT": "सींगापुर स्टैंडर्ड टाईम", "COST": "कोलंबिया समर टाईम", "∅∅∅": "अमाज़ोन समर टाईम", "CHAST": "चैथम स्टैंडर्ड टाईम", "AWST": "दक्षिण ओस्ट्रेलिया स्टैंडर्ड टाईम", "ACST": "मध्य ओस्ट्रेलिया स्टैंडर्ड टाईम", "ACDT": "मध्य ओस्ट्रेलिया डेलाईट टाईम", "HNT": "न्युफाऊंडलैण्ड स्टैंडर्ड टाईम", "HNCU": "क्युबा स्टैंडर्ड टाईम", "SAST": "दक्षिण अफ्रीका स्टैंडर्ड टाईम", "WAT": "पश्चीम अफ्रीका स्टैंडर्ड टाईम", "WARST": "पश्चीम अर्जण्टिना समर टाईम", "NZDT": "न्युज़ीलैण्ड डेलाईट टाईम", "EST": "ईस्टर्न अमरिका स्टैंडर्ड टाईम", "EAT": "पूर्वी अफ्रीका स्टैंडर्ड टाईम", "HADT": "हवाई आलटन डेलाईट टाईम", "AEDT": "पूर्वी ओस्ट्रेलिया डेलाईट टाईम", "BOT": "बोलिविया स्टैंडर्ड टाईम", "CST": "सैंट्रल अमरिका स्टैंडर्ड टाईम", "WIB": "वेस्टर्न ईंडोनीशिया स्टैंडर्ड टाईम", "GFT": "फ्रान्सीसी गुयाना स्टैंडर्ड टाईम", "MESZ": "मध्य यूरोप समर टाईम", "ART": "अर्जनटिना स्टैंडर्ड टाईम", "COT": "कोलंबिया स्टैंडर्ड टाईम", "HECU": "क्युबा डेलाईट टाईम", "HEPMX": "HEPMX", "HNPM": "सेँ पीयॅर एवं मीकलों स्टैंडर्ड टाईम", "HAT": "न्युफाऊंडलैण्ड डेलाईट टाईम", "WITA": "ईंडोनीशिया स्टैंडर्ड टाईम", "CLST": "चीली समर टाईम", "ARST": "अर्जण्टिना समर टाईम", "MEZ": "मध्य यूरोप स्टैंडर्ड टाईम", "LHDT": "लार्ड़ होव डेलाईट टाईम", "ADT": "अटलांटीक डेलाईट टाईम", "HKT": "हाँगकॉंग स्टैंडर्ड टाईम", "HKST": "हाँगकॉंग समर टाईम"},
	}
}

// Locale returns the current translators string locale
func (brx *brx) Locale() string {
	return brx.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'brx'
func (brx *brx) PluralsCardinal() []locales.PluralRule {
	return brx.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'brx'
func (brx *brx) PluralsOrdinal() []locales.PluralRule {
	return brx.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'brx'
func (brx *brx) PluralsRange() []locales.PluralRule {
	return brx.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'brx'
func (brx *brx) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'brx'
func (brx *brx) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'brx'
func (brx *brx) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (brx *brx) MonthAbbreviated(month time.Month) string {
	return brx.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (brx *brx) MonthsAbbreviated() []string {
	return nil
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (brx *brx) MonthNarrow(month time.Month) string {
	return brx.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (brx *brx) MonthsNarrow() []string {
	return brx.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (brx *brx) MonthWide(month time.Month) string {
	return brx.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (brx *brx) MonthsWide() []string {
	return brx.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (brx *brx) WeekdayAbbreviated(weekday time.Weekday) string {
	return brx.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (brx *brx) WeekdaysAbbreviated() []string {
	return brx.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (brx *brx) WeekdayNarrow(weekday time.Weekday) string {
	return brx.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (brx *brx) WeekdaysNarrow() []string {
	return brx.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (brx *brx) WeekdayShort(weekday time.Weekday) string {
	return brx.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (brx *brx) WeekdaysShort() []string {
	return brx.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (brx *brx) WeekdayWide(weekday time.Weekday) string {
	return brx.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (brx *brx) WeekdaysWide() []string {
	return brx.daysWide
}

// Decimal returns the decimal point of number
func (brx *brx) Decimal() string {
	return brx.decimal
}

// Group returns the group of number
func (brx *brx) Group() string {
	return brx.group
}

// Group returns the minus sign of number
func (brx *brx) Minus() string {
	return brx.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'brx' and handles both Whole and Real numbers based on 'v'
func (brx *brx) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	inSecondary := false
	groupThreshold := 3

	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, brx.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {

			if count == groupThreshold {
				b = append(b, brx.group[0])
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
		b = append(b, brx.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'brx' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (brx *brx) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 3
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, brx.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, brx.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, brx.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'brx'
func (brx *brx) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := brx.currencies[currency]
	l := len(s) + len(symbol) + 4 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	inSecondary := false
	groupThreshold := 3

	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, brx.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {

			if count == groupThreshold {
				b = append(b, brx.group[0])
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

	for j := len(brx.currencyPositivePrefix) - 1; j >= 0; j-- {
		b = append(b, brx.currencyPositivePrefix[j])
	}

	if num < 0 {
		b = append(b, brx.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, brx.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'brx'
// in accounting notation.
func (brx *brx) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := brx.currencies[currency]
	l := len(s) + len(symbol) + 4 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	inSecondary := false
	groupThreshold := 3

	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, brx.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {

			if count == groupThreshold {
				b = append(b, brx.group[0])
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

		for j := len(brx.currencyNegativePrefix) - 1; j >= 0; j-- {
			b = append(b, brx.currencyNegativePrefix[j])
		}

		b = append(b, brx.minus[0])

	} else {

		for j := len(symbol) - 1; j >= 0; j-- {
			b = append(b, symbol[j])
		}

		for j := len(brx.currencyPositivePrefix) - 1; j >= 0; j-- {
			b = append(b, brx.currencyPositivePrefix[j])
		}

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, brx.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'brx'
func (brx *brx) FmtDateShort(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Month()), 10)
	b = append(b, []byte{0x2f}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2f}...)

	if t.Year() > 9 {
		b = append(b, strconv.Itoa(t.Year())[2:]...)
	} else {
		b = append(b, strconv.Itoa(t.Year())[1:]...)
	}

	return string(b)
}

// FmtDateMedium returns the medium date representation of 't' for 'brx'
func (brx *brx) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, brx.monthsAbbreviated[t.Month()]...)
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

// FmtDateLong returns the long date representation of 't' for 'brx'
func (brx *brx) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, brx.monthsWide[t.Month()]...)
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

// FmtDateFull returns the full date representation of 't' for 'brx'
func (brx *brx) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, brx.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)
	b = append(b, brx.monthsWide[t.Month()]...)
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

// FmtTimeShort returns the short time representation of 't' for 'brx'
func (brx *brx) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, brx.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, brx.periodsAbbreviated[0]...)
	} else {
		b = append(b, brx.periodsAbbreviated[1]...)
	}

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'brx'
func (brx *brx) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, brx.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, brx.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, brx.periodsAbbreviated[0]...)
	} else {
		b = append(b, brx.periodsAbbreviated[1]...)
	}

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'brx'
func (brx *brx) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, brx.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, brx.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, brx.periodsAbbreviated[0]...)
	} else {
		b = append(b, brx.periodsAbbreviated[1]...)
	}

	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'brx'
func (brx *brx) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, brx.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, brx.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, brx.periodsAbbreviated[0]...)
	} else {
		b = append(b, brx.periodsAbbreviated[1]...)
	}

	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := brx.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
