package hi

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type hi struct {
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

// New returns a new instance of translator for the 'hi' locale
func New() locales.Translator {
	return &hi{
		locale:             "hi",
		pluralsCardinal:    []locales.PluralRule{2, 6},
		pluralsOrdinal:     []locales.PluralRule{2, 3, 4, 5, 6},
		pluralsRange:       []locales.PluralRule{2, 6},
		decimal:            ".",
		group:              ",",
		minus:              "-",
		percent:            "%",
		perMille:           "‰",
		timeSeparator:      ":",
		inifinity:          "∞",
		currencies:         []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "A$", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "R$", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CA$", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CN¥", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "€", "FIM", "FJD", "FKP", "FRF", "£", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HK$", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "₪", "₹", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JP¥", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "₩", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MX$", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZ$", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "฿", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "NT$", "TZS", "UAH", "UAK", "UGS", "UGX", "$", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "₫", "VNN", "VUV", "WST", "FCFA", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "EC$", "XDR", "XEU", "XFO", "XFU", "CFA", "XPD", "CFPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		monthsAbbreviated:  []string{"", "जन॰", "फ़र॰", "मार्च", "अप्रैल", "मई", "जून", "जुल॰", "अग॰", "सित॰", "अक्तू॰", "नव॰", "दिस॰"},
		monthsNarrow:       []string{"", "ज", "फ़", "मा", "अ", "म", "जू", "जु", "अ", "सि", "अ", "न", "दि"},
		monthsWide:         []string{"", "जनवरी", "फ़रवरी", "मार्च", "अप्रैल", "मई", "जून", "जुलाई", "अगस्त", "सितंबर", "अक्तूबर", "नवंबर", "दिसंबर"},
		daysAbbreviated:    []string{"रवि", "सोम", "मंगल", "बुध", "गुरु", "शुक्र", "शनि"},
		daysNarrow:         []string{"र", "सो", "मं", "बु", "गु", "शु", "श"},
		daysShort:          []string{"र", "सो", "मं", "बु", "गु", "शु", "श"},
		daysWide:           []string{"रविवार", "सोमवार", "मंगलवार", "बुधवार", "गुरुवार", "शुक्रवार", "शनिवार"},
		periodsAbbreviated: []string{"पूर्वाह्न", "अपराह्न"},
		periodsNarrow:      []string{"पू", "अ"},
		periodsWide:        []string{"पूर्वाह्न", "अपराह्न"},
		erasAbbreviated:    []string{"ईसा-पूर्व", "ईस्वी"},
		erasNarrow:         []string{"", ""},
		erasWide:           []string{"ईसा-पूर्व", "ईसवी सन"},
		timezones:          map[string]string{"SGT": "सिंगापुर समय", "IST": "भारतीय मानक समय", "HNT": "न्यूफ़ाउंडलैंड मानक समय", "ChST": "चामोरो मानक समय", "HECU": "क्यूबा डेलाइट समय", "MST": "उत्तरी अमेरिकी माउंटेन मानक समय", "BOT": "बोलीविया समय", "AKST": "अलास्\u200dका मानक समय", "CST": "उत्तरी अमेरिकी केंद्रीय मानक समय", "PDT": "उत्तरी अमेरिकी प्रशांत डेलाइट समय", "HKT": "हाँग काँग मानक समय", "UYT": "उरुग्वे मानक समय", "ART": "अर्जेंटीना मानक समय", "COST": "कोलंबिया ग्रीष्मकालीन समय", "WIB": "पश्चिमी इंडोनेशिया समय", "ACWST": "ऑस्\u200dट्रेलियाई केंद्रीय पश्चिमी मानक समय", "HAST": "हवाई–आल्यूशन मानक समय", "ECT": "इक्वाडोर समय", "EST": "उत्तरी अमेरिकी पूर्वी मानक समय", "WITA": "मध्य इंडोनेशिया समय", "CDT": "उत्तरी अमेरिकी केंद्रीय डेलाइट समय", "WESZ": "पश्चिमी यूरोपीय ग्रीष्\u200dमकालीन समय", "NZST": "न्यूज़ीलैंड मानक समय", "NZDT": "न्यूज़ीलैंड डेलाइट समय", "JDT": "जापान डेलाइट समय", "HNOG": "पश्चिमी ग्रीनलैंड मानक समय", "WART": "पश्चिमी अर्जेंटीना मानक समय", "VET": "वेनेज़ुएला समय", "CLST": "चिली ग्रीष्मकालीन समय", "ADT": "अटलांटिक डेलाइट समय", "WAST": "पश्चिम अफ़्रीका ग्रीष्मकालीन समय", "WEZ": "पश्चिमी यूरोपीय मानक समय", "OESZ": "पूर्वी यूरोपीय ग्रीष्मकालीन समय", "MDT": "उत्तरी अमेरिकी माउंटेन डेलाइट समय", "TMT": "तुर्कमेनिस्तान मानक समय", "TMST": "तुर्कमेनिस्तान ग्रीष्मकालीन समय", "PST": "उत्तरी अमेरिकी प्रशांत मानक समय", "HKST": "हाँग काँग ग्रीष्मकालीन समय", "OEZ": "पूर्वी यूरोपीय मानक समय", "HADT": "हवाई–आल्यूशन डेलाइट समय", "MESZ": "मध्\u200dय यूरोपीय ग्रीष्\u200dमकालीन समय", "HEPM": "सेंट पिएरे और मिक्वेलान डेलाइट समय", "HNCU": "क्यूबा मानक समय", "AWDT": "ऑस्ट्रेलियाई पश्चिमी डेलाइट समय", "WAT": "पश्चिम अफ़्रीका मानक समय", "∅∅∅": "∅∅∅", "AKDT": "अलास्\u200dका डेलाइट समय", "CAT": "मध्य अफ़्रीका समय", "CHAST": "चैथम मानक समय", "GFT": "फ़्रेंच गुयाना समय", "ACDT": "ऑस्\u200dट्रेलियाई केंद्रीय डेलाइट समय", "HNPM": "सेंट पिएरे और मिक्वेलान मानक समय", "JST": "जापान मानक समय", "ARST": "अर्जेंटीना ग्रीष्मकालीन समय", "HEEG": "पूर्वी ग्रीनलैंड ग्रीष्मकालीन समय", "WARST": "पश्चिमी अर्जेंटीना ग्रीष्मकालीन समय", "EAT": "पूर्वी अफ़्रीका समय", "GMT": "ग्रीनविच मीन टाइम", "AEDT": "ऑस्\u200dट्रेलियाई पूर्वी डेलाइट समय", "SAST": "दक्षिण अफ़्रीका मानक समय", "EDT": "उत्तरी अमेरिकी पूर्वी डेलाइट समय", "GYT": "गुयाना समय", "UYST": "उरुग्वे ग्रीष्मकालीन समय", "HNPMX": "मेक्सिकन प्रशांत मानक समय", "BT": "भूटान समय", "HEOG": "पश्चिमी ग्रीनलैंड ग्रीष्मकालीन समय", "WIT": "पूर्वी इंडोनेशिया समय", "SRT": "सूरीनाम समय", "COT": "कोलंबिया मानक समय", "CHADT": "चैथम डेलाइट समय", "HEPMX": "मेक्सिकन प्रशांत डेलाइट समय", "MYT": "मलेशिया समय", "LHST": "लॉर्ड होवे मानक समय", "HNNOMX": "उत्तर पश्चिमी मेक्सिको मानक समय", "LHDT": "लॉर्ड होवे डेलाइट समय", "AST": "अटलांटिक मानक समय", "ACST": "ऑस्\u200dट्रेलियाई केंद्रीय मानक समय", "MEZ": "मध्य यूरोपीय मानक समय", "HENOMX": "उत्तर पश्चिमी मेक्सिको डेलाइट समय", "CLT": "चिली मानक समय", "AEST": "ऑस्\u200dट्रेलियाई पूर्वी मानक समय", "ACWDT": "ऑस्\u200dट्रेलियाई केंद्रीय पश्चिमी डेलाइट समय", "HNEG": "पूर्वी ग्रीनलैंड मानक समय", "HAT": "न्यूफ़ाउंडलैंड डेलाइट समय", "AWST": "ऑस्ट्रेलियाई पश्चिमी मानक समय"},
	}
}

// Locale returns the current translators string locale
func (hi *hi) Locale() string {
	return hi.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'hi'
func (hi *hi) PluralsCardinal() []locales.PluralRule {
	return hi.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'hi'
func (hi *hi) PluralsOrdinal() []locales.PluralRule {
	return hi.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'hi'
func (hi *hi) PluralsRange() []locales.PluralRule {
	return hi.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'hi'
func (hi *hi) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)
	i := int64(n)

	if (i == 0) || (n == 1) {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'hi'
func (hi *hi) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 {
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

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'hi'
func (hi *hi) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {

	start := hi.CardinalPluralRule(num1, v1)
	end := hi.CardinalPluralRule(num2, v2)

	if start == locales.PluralRuleOne && end == locales.PluralRuleOne {
		return locales.PluralRuleOne
	} else if start == locales.PluralRuleOne && end == locales.PluralRuleOther {
		return locales.PluralRuleOther
	}

	return locales.PluralRuleOther

}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (hi *hi) MonthAbbreviated(month time.Month) string {
	return hi.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (hi *hi) MonthsAbbreviated() []string {
	return hi.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (hi *hi) MonthNarrow(month time.Month) string {
	return hi.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (hi *hi) MonthsNarrow() []string {
	return hi.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (hi *hi) MonthWide(month time.Month) string {
	return hi.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (hi *hi) MonthsWide() []string {
	return hi.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (hi *hi) WeekdayAbbreviated(weekday time.Weekday) string {
	return hi.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (hi *hi) WeekdaysAbbreviated() []string {
	return hi.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (hi *hi) WeekdayNarrow(weekday time.Weekday) string {
	return hi.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (hi *hi) WeekdaysNarrow() []string {
	return hi.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (hi *hi) WeekdayShort(weekday time.Weekday) string {
	return hi.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (hi *hi) WeekdaysShort() []string {
	return hi.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (hi *hi) WeekdayWide(weekday time.Weekday) string {
	return hi.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (hi *hi) WeekdaysWide() []string {
	return hi.daysWide
}

// Decimal returns the decimal point of number
func (hi *hi) Decimal() string {
	return hi.decimal
}

// Group returns the group of number
func (hi *hi) Group() string {
	return hi.group
}

// Group returns the minus sign of number
func (hi *hi) Minus() string {
	return hi.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'hi' and handles both Whole and Real numbers based on 'v'
func (hi *hi) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	inSecondary := false
	groupThreshold := 3

	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, hi.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {

			if count == groupThreshold {
				b = append(b, hi.group[0])
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
		b = append(b, hi.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'hi' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (hi *hi) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 3
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, hi.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, hi.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, hi.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'hi'
func (hi *hi) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := hi.currencies[currency]
	l := len(s) + len(symbol) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	inSecondary := false
	groupThreshold := 3

	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, hi.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {

			if count == groupThreshold {
				b = append(b, hi.group[0])
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

	if num < 0 {
		b = append(b, hi.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, hi.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'hi'
// in accounting notation.
func (hi *hi) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := hi.currencies[currency]
	l := len(s) + len(symbol) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	inSecondary := false
	groupThreshold := 3

	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, hi.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {

			if count == groupThreshold {
				b = append(b, hi.group[0])
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

		b = append(b, hi.minus[0])

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
			b = append(b, hi.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'hi'
func (hi *hi) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'hi'
func (hi *hi) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, hi.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'hi'
func (hi *hi) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, hi.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'hi'
func (hi *hi) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, hi.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, hi.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'hi'
func (hi *hi) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, hi.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, hi.periodsAbbreviated[0]...)
	} else {
		b = append(b, hi.periodsAbbreviated[1]...)
	}

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'hi'
func (hi *hi) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, hi.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, hi.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, hi.periodsAbbreviated[0]...)
	} else {
		b = append(b, hi.periodsAbbreviated[1]...)
	}

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'hi'
func (hi *hi) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, hi.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, hi.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, hi.periodsAbbreviated[0]...)
	} else {
		b = append(b, hi.periodsAbbreviated[1]...)
	}

	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'hi'
func (hi *hi) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, hi.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, hi.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, hi.periodsAbbreviated[0]...)
	} else {
		b = append(b, hi.periodsAbbreviated[1]...)
	}

	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := hi.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
