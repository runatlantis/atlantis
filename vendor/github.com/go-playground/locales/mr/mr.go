package mr

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type mr struct {
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

// New returns a new instance of translator for the 'mr' locale
func New() locales.Translator {
	return &mr{
		locale:             "mr",
		pluralsCardinal:    []locales.PluralRule{2, 6},
		pluralsOrdinal:     []locales.PluralRule{2, 3, 4, 6},
		pluralsRange:       []locales.PluralRule{2, 6},
		decimal:            ".",
		group:              ",",
		minus:              "-",
		percent:            "%",
		perMille:           "‰",
		timeSeparator:      ":",
		inifinity:          "∞",
		currencies:         []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "A$", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "R$", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CA$", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CN¥", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "€", "FIM", "FJD", "FKP", "FRF", "£", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HK$", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "₪", "₹", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JP¥", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "₩", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MX$", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZ$", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "฿", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "NT$", "TZS", "UAH", "UAK", "UGS", "UGX", "$", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "₫", "VNN", "VUV", "WST", "FCFA", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "EC$", "XDR", "XEU", "XFO", "XFU", "CFA", "XPD", "CFPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		monthsAbbreviated:  []string{"", "जाने", "फेब्रु", "मार्च", "एप्रि", "मे", "जून", "जुलै", "ऑग", "सप्टें", "ऑक्टो", "नोव्हें", "डिसें"},
		monthsNarrow:       []string{"", "जा", "फे", "मा", "ए", "मे", "जू", "जु", "ऑ", "स", "ऑ", "नो", "डि"},
		monthsWide:         []string{"", "जानेवारी", "फेब्रुवारी", "मार्च", "एप्रिल", "मे", "जून", "जुलै", "ऑगस्ट", "सप्टेंबर", "ऑक्टोबर", "नोव्हेंबर", "डिसेंबर"},
		daysAbbreviated:    []string{"रवि", "सोम", "मंगळ", "बुध", "गुरु", "शुक्र", "शनि"},
		daysNarrow:         []string{"र", "सो", "मं", "बु", "गु", "शु", "श"},
		daysShort:          []string{"र", "सो", "मं", "बु", "गु", "शु", "श"},
		daysWide:           []string{"रविवार", "सोमवार", "मंगळवार", "बुधवार", "गुरुवार", "शुक्रवार", "शनिवार"},
		periodsAbbreviated: []string{"म.पू.", "म.उ."},
		periodsNarrow:      []string{"स", "सं"},
		periodsWide:        []string{"म.पू.", "म.उ."},
		erasAbbreviated:    []string{"इ. स. पू.", "इ. स."},
		erasNarrow:         []string{"", ""},
		erasWide:           []string{"ईसवीसनपूर्व", "ईसवीसन"},
		timezones:          map[string]string{"AEST": "ऑस्ट्रेलियन पूर्व प्रमाण वेळ", "AEDT": "ऑस्ट्रेलियन पूर्व सूर्यप्रकाश वेळ", "SGT": "सिंगापूर प्रमाण वेळ", "WART": "पश्चिमी अर्जेंटिना प्रमाण वेळ", "VET": "व्हेनेझुएला वेळ", "HENOMX": "वायव्य मेक्सिको सूर्यप्रकाश वेळ", "SRT": "सुरिनाम वेळ", "COST": "कोलंबिया उन्हाळी वेळ", "HKT": "हाँग काँग प्रमाण वेळ", "CLST": "चिली उन्हाळी वेळ", "PDT": "पॅसिफिक सूर्यप्रकाश वेळ", "GFT": "फ्रेंच गयाना वेळ", "HNOG": "पश्चिम ग्रीनलँड प्रमाण वेळ", "HEOG": "पश्चिम ग्रीनलँड उन्हाळी वेळ", "MESZ": "मध्\u200dय युरोपियन उन्हाळी वेळ", "TMT": "तुर्कमेनिस्तान प्रमाण वेळ", "UYT": "उरुग्वे प्रमाण वेळ", "ECT": "इक्वेडोर वेळ", "WARST": "पश्चिमी अर्जेंटिना उन्हाळी वेळ", "CLT": "चिली प्रमाण वेळ", "AWDT": "ऑस्ट्रेलियन पश्चिम सूर्यप्रकाश वेळ", "HEPM": "सेंट पियरे आणि मिक्वेलोन सूर्यप्रकाश वेळ", "WITA": "मध्\u200dय इंडोनेशिया वेळ", "MDT": "मकाऊ ग्रीष्मकालीन वेळ", "EAT": "पूर्व आफ्रिका वेळ", "HNCU": "क्यूबा प्रमाण वेळ", "AST": "अटलांटिक प्रमाण वेळ", "AKST": "अलास्का प्रमाण वेळ", "LHST": "लॉर्ड होवे प्रमाण वेळ", "HAT": "न्यू फाउंडलंड सूर्यप्रकाश वेळ", "PST": "पॅसिफिक प्रमाण वेळ", "WIB": "पश्चिमी इंडोनेशिया वेळ", "MST": "मकाऊ प्रमाणवेळ", "WIT": "पौर्वात्य इंडोनेशिया वेळ", "HADT": "हवाई-अलूशन सूर्यप्रकाश वेळ", "GMT": "ग्रीनिच प्रमाण वेळ", "CHAST": "चॅथम प्रमाण वेळ", "CHADT": "चॅथम सूर्यप्रकाश वेळ", "JDT": "जपान सूर्यप्रकाश वेळ", "ACST": "ऑस्ट्रेलियन मध्य प्रमाण वेळ", "ACDT": "ऑस्ट्रेलियन मध्य सूर्यप्रकाश वेळ", "TMST": "तुर्कमेनिस्तान उन्हाळी वेळ", "ART": "अर्जेंटिना प्रमाण वेळ", "HNPM": "सेंट पियरे आणि मिक्वेलोन प्रमाण वेळ", "HECU": "क्यूबा सूर्यप्रकाश वेळ", "CDT": "केंद्रीय सूर्यप्रकाश वेळ", "HNPMX": "मेक्सिको पॅसिफिक प्रमाण वेळ", "WESZ": "पश्चिम युरोपियन उन्हाळी वेळ", "MYT": "मलेशिया वेळ", "AKDT": "अलास्का सूर्यप्रकाश वेळ", "ACWST": "ऑस्ट्रेलियन मध्य-पश्चिम प्रमाण वेळ", "MEZ": "मध्\u200dय युरोपियन प्रमाण वेळ", "SAST": "दक्षिण आफ्रिका प्रमाण वेळ", "WAT": "पश्चिम आफ्रिका प्रमाण वेळ", "NZDT": "न्यूझीलंड सूर्यप्रकाश वेळ", "EST": "पौर्वात्य प्रमाण वेळ", "HEEG": "पूर्व ग्रीनलँड उन्हाळी वेळ", "GYT": "गयाना वेळ", "ChST": "चामोरो प्रमाण वेळ", "AWST": "ऑस्ट्रेलियन पश्चिम प्रमाण वेळ", "HKST": "हाँग काँग उन्हाळी वेळ", "LHDT": "लॉर्ड होवे सूर्यप्रकाश वेळ", "OEZ": "पूर्व युरोपियन प्रमाण वेळ", "HAST": "हवाई-अलूशन प्रमाण वेळ", "UYST": "उरुग्वे उन्हाळी वेळ", "BOT": "बोलिव्हिया वेळ", "HNEG": "पूर्व ग्रीनलँड प्रमाण वेळ", "HNT": "न्यू फाउंडलंड प्रमाण वेळ", "OESZ": "पूर्व युरोपियन उन्हाळी वेळ", "WAST": "पश्चिम आफ्रिका उन्हाळी वेळ", "WEZ": "पश्चिम युरोपियन प्रमाण वेळ", "HNNOMX": "वायव्य मेक्सिको प्रमाण वेळ", "HEPMX": "मेक्सिको पॅसिफिक सूर्यप्रकाश वेळ", "JST": "जपान प्रमाण वेळ", "EDT": "पौर्वात्य सूर्यप्रकाश वेळ", "IST": "भारतीय प्रमाण वेळ", "ARST": "अर्जेंटिना उन्हाळी वेळ", "CST": "केंद्रीय प्रमाण वेळ", "ADT": "अटलांटिक सूर्यप्रकाश वेळ", "BT": "भूतान वेळ", "NZST": "न्यूझीलंड प्रमाण वेळ", "ACWDT": "ऑस्ट्रेलियन मध्य-पश्चिम सूर्यप्रकाश वेळ", "∅∅∅": "अ\u200dॅझोरेस उन्हाळी वेळ", "CAT": "मध्\u200dय आफ्रिका वेळ", "COT": "कोलंबिया प्रमाण वेळ"},
	}
}

// Locale returns the current translators string locale
func (mr *mr) Locale() string {
	return mr.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'mr'
func (mr *mr) PluralsCardinal() []locales.PluralRule {
	return mr.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'mr'
func (mr *mr) PluralsOrdinal() []locales.PluralRule {
	return mr.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'mr'
func (mr *mr) PluralsRange() []locales.PluralRule {
	return mr.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'mr'
func (mr *mr) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)
	i := int64(n)

	if (i == 0) || (n == 1) {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'mr'
func (mr *mr) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 {
		return locales.PluralRuleOne
	} else if n == 2 || n == 3 {
		return locales.PluralRuleTwo
	} else if n == 4 {
		return locales.PluralRuleFew
	}

	return locales.PluralRuleOther
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'mr'
func (mr *mr) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {

	start := mr.CardinalPluralRule(num1, v1)
	end := mr.CardinalPluralRule(num2, v2)

	if start == locales.PluralRuleOne && end == locales.PluralRuleOne {
		return locales.PluralRuleOne
	} else if start == locales.PluralRuleOne && end == locales.PluralRuleOther {
		return locales.PluralRuleOther
	}

	return locales.PluralRuleOther

}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (mr *mr) MonthAbbreviated(month time.Month) string {
	return mr.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (mr *mr) MonthsAbbreviated() []string {
	return mr.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (mr *mr) MonthNarrow(month time.Month) string {
	return mr.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (mr *mr) MonthsNarrow() []string {
	return mr.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (mr *mr) MonthWide(month time.Month) string {
	return mr.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (mr *mr) MonthsWide() []string {
	return mr.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (mr *mr) WeekdayAbbreviated(weekday time.Weekday) string {
	return mr.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (mr *mr) WeekdaysAbbreviated() []string {
	return mr.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (mr *mr) WeekdayNarrow(weekday time.Weekday) string {
	return mr.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (mr *mr) WeekdaysNarrow() []string {
	return mr.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (mr *mr) WeekdayShort(weekday time.Weekday) string {
	return mr.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (mr *mr) WeekdaysShort() []string {
	return mr.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (mr *mr) WeekdayWide(weekday time.Weekday) string {
	return mr.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (mr *mr) WeekdaysWide() []string {
	return mr.daysWide
}

// Decimal returns the decimal point of number
func (mr *mr) Decimal() string {
	return mr.decimal
}

// Group returns the group of number
func (mr *mr) Group() string {
	return mr.group
}

// Group returns the minus sign of number
func (mr *mr) Minus() string {
	return mr.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'mr' and handles both Whole and Real numbers based on 'v'
func (mr *mr) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	inSecondary := false
	groupThreshold := 3

	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, mr.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {

			if count == groupThreshold {
				b = append(b, mr.group[0])
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
		b = append(b, mr.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'mr' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (mr *mr) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 3
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, mr.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, mr.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, mr.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'mr'
func (mr *mr) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := mr.currencies[currency]
	l := len(s) + len(symbol) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, mr.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, mr.group[0])
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
		b = append(b, mr.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, mr.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'mr'
// in accounting notation.
func (mr *mr) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := mr.currencies[currency]
	l := len(s) + len(symbol) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, mr.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, mr.group[0])
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

		b = append(b, mr.minus[0])

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
			b = append(b, mr.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'mr'
func (mr *mr) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'mr'
func (mr *mr) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, mr.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x2c, 0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'mr'
func (mr *mr) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, mr.monthsWide[t.Month()]...)
	b = append(b, []byte{0x2c, 0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'mr'
func (mr *mr) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, mr.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, mr.monthsWide[t.Month()]...)
	b = append(b, []byte{0x2c, 0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'mr'
func (mr *mr) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, mr.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, mr.periodsAbbreviated[0]...)
	} else {
		b = append(b, mr.periodsAbbreviated[1]...)
	}

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'mr'
func (mr *mr) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, mr.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, mr.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, mr.periodsAbbreviated[0]...)
	} else {
		b = append(b, mr.periodsAbbreviated[1]...)
	}

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'mr'
func (mr *mr) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, mr.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, mr.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, mr.periodsAbbreviated[0]...)
	} else {
		b = append(b, mr.periodsAbbreviated[1]...)
	}

	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'mr'
func (mr *mr) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, mr.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, mr.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, mr.periodsAbbreviated[0]...)
	} else {
		b = append(b, mr.periodsAbbreviated[1]...)
	}

	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := mr.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
