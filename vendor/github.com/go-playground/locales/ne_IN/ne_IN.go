package ne_IN

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type ne_IN struct {
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

// New returns a new instance of translator for the 'ne_IN' locale
func New() locales.Translator {
	return &ne_IN{
		locale:                 "ne_IN",
		pluralsCardinal:        []locales.PluralRule{2, 6},
		pluralsOrdinal:         []locales.PluralRule{2, 6},
		pluralsRange:           []locales.PluralRule{2, 6},
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
		monthsAbbreviated:      []string{"", "जनवरी", "फेब्रुअरी", "मार्च", "अप्रिल", "मे", "जुन", "जुलाई", "अगस्ट", "सेप्टेम्बर", "अक्टोबर", "नोभेम्बर", "डिसेम्बर"},
		monthsNarrow:           []string{"", "जन", "फेब", "मार्च", "अप्र", "मे", "जुन", "जुल", "अग", "सेप", "अक्टो", "नोभे", "डिसे"},
		monthsWide:             []string{"", "जनवरी", "फेब्रुअरी", "मार्च", "अप्रिल", "मे", "जुन", "जुलाई", "अगस्ट", "सेप्टेम्बर", "अक्टोबर", "नोभेम्बर", "डिसेम्बर"},
		daysAbbreviated:        []string{"आइत", "सोम", "मङ्गल", "बुध", "बिहि", "शुक्र", "शनि"},
		daysNarrow:             []string{"आ", "सो", "म", "बु", "बि", "शु", "श"},
		daysShort:              []string{"आइत", "सोम", "मङ्गल", "बुध", "बिहि", "शुक्र", "शनि"},
		daysWide:               []string{"आइतबार", "सोमबार", "मङ्गलबार", "बुधबार", "बिहिबार", "शुक्रबार", "शनिबार"},
		periodsAbbreviated:     []string{"पूर्वाह्न", "अपराह्न"},
		periodsNarrow:          []string{"पूर्वाह्न", "अपराह्न"},
		periodsWide:            []string{"पूर्वाह्न", "अपराह्न"},
		erasAbbreviated:        []string{"ईसा पूर्व", "सन्"},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"ईसा पूर्व", "सन्"},
		timezones:              map[string]string{"ACWDT": "केन्द्रीय पश्चिमी अस्ट्रेलिया दिवा समय", "WART": "पश्चिमी अर्जेनटिनी मानक समय", "ART": "अर्जेनटिनी मानक समय", "UYST": "उरुग्वे ग्रीष्मकालीन समय", "AST": "एट्लान्टिक मानक समय", "AKST": "अलस्काको मानक समय", "OEZ": "पूर्वी युरोपेली मानक समय", "ARST": "अर्जेनटिनी ग्रीष्मकालीन समय", "PDT": "प्यासिफिक दिवा समय", "∅∅∅": "∅∅∅", "UYT": "उरूग्वे मानक समय", "AEDT": "पूर्वी अस्ट्रेलिया दिवा समय", "CLT": "चिली मानक समय", "OESZ": "पूर्वी युरोपेली ग्रीष्मकालीन समय", "COT": "कोलम्बियाली मानक समय", "CHAST": "चाथाम मानक समय", "ADT": "एट्लान्टिक दिवा समय", "EAT": "पूर्वी अफ्रिकी समय", "ECT": "ईक्वोडोर समय", "HEEG": "पूर्वी ग्रीनल्यान्डको ग्रीष्मकालीन समय", "SRT": "सुरिनामा समय", "WAT": "पश्चिम अफ्रिकी मानक समय", "ChST": "चामोर्रो मानक समय", "CST": "केन्द्रीय मानक समय", "LHDT": "लर्ड हावे दिवा समय", "HNT": "न्यूफाउनडल्यान्डको मानक समय", "HADT": "हवाई-एलुटियन दिवा समय", "GYT": "गुयाना समय", "HECU": "क्यूबाको दिवा समय", "AEST": "पूर्वी अस्ट्रेलिया मानक समय", "JST": "जापान मानक समय", "MEZ": "केन्द्रीय युरोपेली मानक समय", "IST": "भारतीय मानक समय", "CDT": "केन्द्रीय दिवा समय", "HNOG": "पश्चिमी ग्रीनल्यान्डको मानक समय", "TMT": "तुर्कमेनिस्तान मानक समय", "CAT": "केन्द्रीय अफ्रिकी समय", "WIT": "पूर्वी इन्डोनेशिया समय", "GMT": "ग्रीनविच मिन समय", "AWDT": "पश्चिमी अस्ट्रेलिया दिवा समय", "JDT": "जापान दिवा समय", "WAST": "पश्चिम अफ्रिकी ग्रीष्मकालीन समय", "SGT": "सिंगापुर मानक समय", "ACST": "केन्द्रीय अस्ट्रेलिया मानक समय", "HKST": "हङकङ ग्रीष्मकालीन समय", "VET": "भेनेज्युएला समय", "TMST": "तुर्कमेनिस्तान ग्रीष्मकालीन मानक समय", "HAST": "हवाई-एलुटियन मानक समय", "SAST": "दक्षिण अफ्रिकी समय", "HNCU": "क्यूबाको मानक समय", "COST": "कोलम्बियाली ग्रीष्मकालीन समय", "MYT": "मलेसिया समय", "HEOG": "पश्चिमी ग्रीनल्यान्डको ग्रीष्मकालीन समय", "ACDT": "केन्द्रीय अस्ट्रेलिया दिवा समय", "HEPM": "सेन्ट पियर्रे र मिक्युलोनको दिवा समय", "WESZ": "युरोपेली ग्रीष्मकालीन समय", "EST": "पूर्वी मानक समय", "EDT": "पूर्वी दिवा समय", "ACWST": "केन्द्रीय पश्चिमी अस्ट्रेलिया मानक समय", "WITA": "केन्द्रीय इन्डोनेशिया समय", "MST": "MST", "MDT": "MDT", "PST": "प्यासिफिक मानक समय", "BOT": "बोलिभिया समय", "AKDT": "अलस्काको दिवा समय", "HNEG": "पूर्वी ग्रीनल्यान्डको मानक समय", "HKT": "हङकङ मानक समय", "HNPM": "सेन्ट पियर्रे र मिक्युलोनको मानक समय", "CLST": "चिली ग्रीष्मकालीन समय", "AWST": "पश्चिमी अस्ट्रेलिया मानक समय", "WIB": "पश्चिमी इन्डोनेशिया समय", "GFT": "फ्रेन्च ग्वाना समय", "NZST": "न्यूजिल्यान्ड मानक समय", "WARST": "पश्चिमी अर्जेनटिनी ग्रीष्मकालीन समय", "HAT": "न्यूफाउनल्यान्डको दिवा समय", "HNNOMX": "उत्तर पश्चिम मेक्सिकोको मानक समय", "HENOMX": "उत्तर पश्चिम मेक्सिकोको दिवा समय", "HEPMX": "मेक्सिकन प्यासिफिक दिवा समय", "WEZ": "पश्चिमी युरोपेली मानक समय", "LHST": "लर्ड हावे मानक समय", "CHADT": "चाथाम दिवा समय", "HNPMX": "मेक्सिकन प्यासिफिक मानक समय", "MESZ": "केन्द्रीय युरोपेली ग्रीष्मकालीन समय", "NZDT": "न्यूजिल्यान्ड दिवा समय", "BT": "भुटानी समय"},
	}
}

// Locale returns the current translators string locale
func (ne *ne_IN) Locale() string {
	return ne.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'ne_IN'
func (ne *ne_IN) PluralsCardinal() []locales.PluralRule {
	return ne.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'ne_IN'
func (ne *ne_IN) PluralsOrdinal() []locales.PluralRule {
	return ne.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'ne_IN'
func (ne *ne_IN) PluralsRange() []locales.PluralRule {
	return ne.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'ne_IN'
func (ne *ne_IN) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'ne_IN'
func (ne *ne_IN) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n >= 1 && n <= 4 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'ne_IN'
func (ne *ne_IN) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {

	start := ne.CardinalPluralRule(num1, v1)
	end := ne.CardinalPluralRule(num2, v2)

	if start == locales.PluralRuleOne && end == locales.PluralRuleOther {
		return locales.PluralRuleOther
	} else if start == locales.PluralRuleOther && end == locales.PluralRuleOne {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther

}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (ne *ne_IN) MonthAbbreviated(month time.Month) string {
	return ne.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (ne *ne_IN) MonthsAbbreviated() []string {
	return ne.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (ne *ne_IN) MonthNarrow(month time.Month) string {
	return ne.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (ne *ne_IN) MonthsNarrow() []string {
	return ne.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (ne *ne_IN) MonthWide(month time.Month) string {
	return ne.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (ne *ne_IN) MonthsWide() []string {
	return ne.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (ne *ne_IN) WeekdayAbbreviated(weekday time.Weekday) string {
	return ne.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (ne *ne_IN) WeekdaysAbbreviated() []string {
	return ne.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (ne *ne_IN) WeekdayNarrow(weekday time.Weekday) string {
	return ne.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (ne *ne_IN) WeekdaysNarrow() []string {
	return ne.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (ne *ne_IN) WeekdayShort(weekday time.Weekday) string {
	return ne.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (ne *ne_IN) WeekdaysShort() []string {
	return ne.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (ne *ne_IN) WeekdayWide(weekday time.Weekday) string {
	return ne.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (ne *ne_IN) WeekdaysWide() []string {
	return ne.daysWide
}

// Decimal returns the decimal point of number
func (ne *ne_IN) Decimal() string {
	return ne.decimal
}

// Group returns the group of number
func (ne *ne_IN) Group() string {
	return ne.group
}

// Group returns the minus sign of number
func (ne *ne_IN) Minus() string {
	return ne.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'ne_IN' and handles both Whole and Real numbers based on 'v'
func (ne *ne_IN) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ne.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, ne.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, ne.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'ne_IN' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (ne *ne_IN) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 3
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ne.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, ne.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, ne.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'ne_IN'
func (ne *ne_IN) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := ne.currencies[currency]
	l := len(s) + len(symbol) + 4 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ne.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, ne.group[0])
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

	for j := len(ne.currencyPositivePrefix) - 1; j >= 0; j-- {
		b = append(b, ne.currencyPositivePrefix[j])
	}

	if num < 0 {
		b = append(b, ne.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, ne.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'ne_IN'
// in accounting notation.
func (ne *ne_IN) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := ne.currencies[currency]
	l := len(s) + len(symbol) + 4 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ne.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, ne.group[0])
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

		for j := len(ne.currencyNegativePrefix) - 1; j >= 0; j-- {
			b = append(b, ne.currencyNegativePrefix[j])
		}

		b = append(b, ne.minus[0])

	} else {

		for j := len(symbol) - 1; j >= 0; j-- {
			b = append(b, symbol[j])
		}

		for j := len(ne.currencyPositivePrefix) - 1; j >= 0; j-- {
			b = append(b, ne.currencyPositivePrefix[j])
		}

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, ne.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'ne_IN'
func (ne *ne_IN) FmtDateShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 9 {
		b = append(b, strconv.Itoa(t.Year())[2:]...)
	} else {
		b = append(b, strconv.Itoa(t.Year())[1:]...)
	}

	b = append(b, []byte{0x2f}...)
	b = strconv.AppendInt(b, int64(t.Month()), 10)
	b = append(b, []byte{0x2f}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)

	return string(b)
}

// FmtDateMedium returns the medium date representation of 't' for 'ne_IN'
func (ne *ne_IN) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x20}...)
	b = append(b, ne.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'ne_IN'
func (ne *ne_IN) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x20}...)
	b = append(b, ne.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'ne_IN'
func (ne *ne_IN) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x20}...)
	b = append(b, ne.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2c, 0x20}...)
	b = append(b, ne.daysWide[t.Weekday()]...)

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'ne_IN'
func (ne *ne_IN) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, ne.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, ne.periodsAbbreviated[0]...)
	} else {
		b = append(b, ne.periodsAbbreviated[1]...)
	}

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'ne_IN'
func (ne *ne_IN) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, ne.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, ne.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, ne.periodsAbbreviated[0]...)
	} else {
		b = append(b, ne.periodsAbbreviated[1]...)
	}

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'ne_IN'
func (ne *ne_IN) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, ne.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, ne.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, ne.periodsAbbreviated[0]...)
	} else {
		b = append(b, ne.periodsAbbreviated[1]...)
	}

	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'ne_IN'
func (ne *ne_IN) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, ne.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, ne.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, ne.periodsAbbreviated[0]...)
	} else {
		b = append(b, ne.periodsAbbreviated[1]...)
	}

	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := ne.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
