package or_IN

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type or_IN struct {
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
	currencyNegativePrefix string
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

// New returns a new instance of translator for the 'or_IN' locale
func New() locales.Translator {
	return &or_IN{
		locale:                 "or_IN",
		pluralsCardinal:        []locales.PluralRule{2, 6},
		pluralsOrdinal:         []locales.PluralRule{2, 3, 4, 5, 6},
		pluralsRange:           []locales.PluralRule{6},
		decimal:                ".",
		group:                  ",",
		minus:                  "-",
		percent:                "%",
		perMille:               "‰",
		timeSeparator:          ":",
		inifinity:              "∞",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyNegativePrefix: "(",
		currencyNegativeSuffix: ")",
		monthsAbbreviated:      []string{"", "ଜାନୁଆରୀ", "ଫେବୃଆରୀ", "ମାର୍ଚ୍ଚ", "ଅପ୍ରେଲ", "ମଇ", "ଜୁନ", "ଜୁଲାଇ", "ଅଗଷ୍ଟ", "ସେପ୍ଟେମ୍ବର", "ଅକ୍ଟୋବର", "ନଭେମ୍ବର", "ଡିସେମ୍ବର"},
		monthsNarrow:           []string{"", "ଜା", "ଫେ", "ମା", "ଅ", "ମଇ", "ଜୁ", "ଜୁ", "ଅ", "ସେ", "ଅ", "ନ", "ଡି"},
		monthsWide:             []string{"", "ଜାନୁଆରୀ", "ଫେବୃଆରୀ", "ମାର୍ଚ୍ଚ", "ଅପ୍ରେଲ", "ମଇ", "ଜୁନ", "ଜୁଲାଇ", "ଅଗଷ୍ଟ", "ସେପ୍ଟେମ୍ବର", "ଅକ୍ଟୋବର", "ନଭେମ୍ବର", "ଡିସେମ୍ବର"},
		daysAbbreviated:        []string{"ରବି", "ସୋମ", "ମଙ୍ଗଳ", "ବୁଧ", "ଗୁରୁ", "ଶୁକ୍ର", "ଶନି"},
		daysNarrow:             []string{"ର", "ସୋ", "ମ", "ବୁ", "ଗୁ", "ଶୁ", "ଶ"},
		daysShort:              []string{"ରବି", "ସୋମ", "ମଙ୍ଗଳ", "ବୁଧ", "ଗୁରୁ", "ଶୁକ୍ର", "ଶନି"},
		daysWide:               []string{"ରବିବାର", "ସୋମବାର", "ମଙ୍ଗଳବାର", "ବୁଧବାର", "ଗୁରୁବାର", "ଶୁକ୍ରବାର", "ଶନିବାର"},
		periodsAbbreviated:     []string{"am", "pm"},
		periodsNarrow:          []string{"ପୂ", "ଅ"},
		periodsWide:            []string{"AM", "PM"},
		erasAbbreviated:        []string{"BC", "AD"},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"ଖ୍ରୀଷ୍ଟପୂର୍ବ", "ଖ୍ରୀଷ୍ଟାବ୍ଦ"},
		timezones:              map[string]string{"JDT": "ଜାପାନ ଦିବାଲୋକ ସମୟ", "ACDT": "ଅଷ୍ଟ୍ରେଲିୟ ମଧ୍ୟ ଦିବାଲୋକ ସମୟ", "VET": "ଭେନିଜୁଏଲା ସମୟ", "PST": "ପାସିଫିକ୍ ମାନାଙ୍କ ସମୟ", "HEEG": "ପୂର୍ବ ଗ୍ରୀନଲ୍ୟାଣ୍ଡ୍ ଗ୍ରୀଷ୍ମକାଳ ସମୟ", "ChST": "ଚାମୋରୋ ମାନାଙ୍କ ସମୟ", "WEZ": "ପଶ୍ଚିମାଞ୍ଚଳ ୟୁରୋପୀୟ ମାନାଙ୍କ ସମୟ", "GYT": "ଗୁଏନା ସମୟ", "AEDT": "ଅଷ୍ଟ୍ରେଲିୟ ପୂର୍ବ ଦିବାଲୋକ ସମୟ", "SGT": "ସିଙ୍ଗାପୁର୍\u200c ମାନାଙ୍କ ସମୟ", "HKT": "ହଂ କଂ ମାନାଙ୍କ ସମୟ", "COST": "କଲମ୍ବିଆ ଗ୍ରୀଷ୍ମକାଳ ସମୟ", "SAST": "ଦକ୍ଷିଣ ଆଫ୍ରିକା ମାନାଙ୍କ ସମୟ", "ECT": "ଇକ୍ୱେଡର ସମୟ", "LHDT": "ଲର୍ଡ ହୋୱେ ଦିବାଲୋକ ସମୟ", "ADT": "ଆଟଲାଣ୍ଟିକ୍ ଦିବାଲୋକ ସମୟ", "WITA": "ମଧ୍ୟ ଇଣ୍ଡୋନେସିଆ ସମୟ", "UYST": "ଉରୁଗୁଏ ଗ୍ରୀଷ୍ମକାଳ ସମୟ", "WESZ": "ପଶ୍ଚିମାଞ୍ଚଳ ୟୁରୋପୀୟ ଗ୍ରୀଷ୍ମକାଳ ସମୟ", "JST": "ଜାପାନ ମାନାଙ୍କ ସମୟ", "WARST": "ପଶ୍ଚିମ ଆର୍ଜେଣ୍ଟିନା ଗ୍ରୀଷ୍ମକାଳ ସମୟ", "HNT": "ନ୍ୟୁଫାଉଣ୍ଡଲ୍ୟାଣ୍ଡ୍ ମାନାଙ୍କ ସମୟ", "HEPM": "ସେଣ୍ଟ. ପିଏରେ ଏବଂ ମିକ୍ୟୁଲୋନ୍ ଦିବାଲୋକ ସମୟ", "MYT": "ମାଲେସିଆ ସମୟ", "ACWDT": "ଅଷ୍ଟ୍ରେଲିୟ ମଧ୍ୟ ପଶ୍ଚିମ ଦିବାଲୋକ ସମୟ", "HNEG": "ପୂର୍ବ ଗ୍ରୀନଲ୍ୟାଣ୍ଡ୍ ମାନାଙ୍କ ସମୟ", "HENOMX": "ଉତ୍ତରପଶ୍ଚିମ ମେକ୍ସିକୋ ଦିବାଲୋକ ସମୟ", "AWST": "ଅଷ୍ଟ୍ରେଲିୟ ପଶ୍ଚିମ ମାନାଙ୍କ ସମୟ", "AWDT": "ଅଷ୍ଟ୍ରେଲିୟ ପଶ୍ଚିମ ଦିବାଲୋକ ସମୟ", "HEPMX": "ମେକ୍ସିକୋ ପାସିଫିକ୍ ଦିବାଲୋକ ସମୟ", "NZDT": "ନ୍ୟୁଜିଲାଣ୍ଡ ଦିବାଲୋକ ସମୟ", "HAT": "ନ୍ୟୁଫାଉଣ୍ଡଲ୍ୟାଣ୍ଡ୍ ଦିବାଲୋକ ସମୟ", "WIT": "ପୂର୍ବ ଇଣ୍ଡୋନେସିଆ ସମୟ", "HECU": "କ୍ୟୁବା ଦିବାଲୋକ ସମୟ", "EAT": "ପୂର୍ବ ଆଫ୍ରିକା ସମୟ", "HNCU": "କ୍ୟୁବା ମାନାଙ୍କ ସମୟ", "HNPMX": "ମେକ୍ସିକୋ ପାସିଫିକ୍ ମାନାଙ୍କ ସମୟ", "∅∅∅": "ଆଜୋରେସ୍ ଗ୍ରୀଷ୍ମକାଳ ସମୟ", "IST": "ଭାରତ ମାନାଙ୍କ ସମୟ", "HNPM": "ସେଣ୍ଟ. ପିଏରେ ଏବଂ ମିକ୍ୟୁଲୋନ୍ ମାନାଙ୍କ ସମୟ", "HKST": "ହଂ କଂ ଗ୍ରୀଷ୍ମକାଳ ସମୟ", "CAT": "ମଧ୍ୟ ଆଫ୍ରିକା ସମୟ", "PDT": "ପାସିଫିକ୍ ଦିବାଲୋକ ସମୟ", "AEST": "ଅଷ୍ଟ୍ରେଲିୟ ପୂର୍ବ ମାନାଙ୍କ ସମୟ", "AKST": "ଆଲାସ୍କା ମାନାଙ୍କ ସମୟ", "EST": "ପୂର୍ବାଞ୍ଚଳ ମାନାଙ୍କ ସମୟ", "ACST": "ଅଷ୍ଟ୍ରେଲିୟ ମଧ୍ୟ ମାନାଙ୍କ ସମୟ", "GMT": "ଗ୍ରୀନୱିଚ୍ ମିନ୍ ସମୟ", "CST": "କେନ୍ଦ୍ରୀୟ ମାନାଙ୍କ ସମୟ", "CDT": "କେନ୍ଦ୍ରୀୟ ଦିବାଲୋକ ସମୟ", "TMST": "ତୁର୍କମେନିସ୍ତାନ ଖରାଦିନ ସମୟ", "COT": "କଲମ୍ବିଆ ମାନାଙ୍କ ସମୟ", "UYT": "ଉରୁଗୁଏ ମାନାଙ୍କ ସମୟ", "CHAST": "ଚାଥାମ୍\u200c ମାନାଙ୍କ ସମୟ", "BOT": "ବଲିଭିଆ ସମୟ", "AKDT": "ଆଲାସ୍କା ଦିବାଲୋକ ସମୟ", "HEOG": "ପଶ୍ଚିମ ଗ୍ରୀନଲ୍ୟାଣ୍ଡ୍ ଗ୍ରୀଷ୍ମ ସମୟ", "CLT": "ଚିଲି ମାନାଙ୍କ ସମୟ", "CLST": "ଚିଲି ଗ୍ରୀଷ୍ମକାଳ ସମୟ", "HAST": "ହୱାଇ-ଆଲେଉଟିୟ ମାନାଙ୍କ ସମୟ", "WIB": "ପଶ୍ଚିମ ଇଣ୍ଡୋନେସିଆ ସମୟ", "BT": "ଭୁଟାନ ସମୟ", "NZST": "ନ୍ୟୁଜିଲାଣ୍ଡ ମାନାଙ୍କ ସମୟ", "EDT": "ପୂର୍ବାଞ୍ଚଳ ଦିବାଲୋକ ସମୟ", "ACWST": "ଅଷ୍ଟ୍ରେଲିୟ ମଧ୍ୟ ପଶ୍ଚିମ ମାନାଙ୍କ ସମୟ", "HNOG": "ପଶ୍ଚିମ ଗ୍ରୀନଲ୍ୟାଣ୍ଡ୍ ମାନାଙ୍କ ସମୟ", "MEZ": "କେନ୍ଦ୍ରୀୟ ୟୁରୋପୀୟ ମାନାଙ୍କ ସମୟ", "LHST": "ଲର୍ଡ ହୋୱେ ମାନାଙ୍କ ସମୟ", "TMT": "ତୁର୍କମେନିସ୍ତାନ ମାନାଙ୍କ ସମୟ", "ART": "ଆର୍ଜେଣ୍ଟିନା ମାନାଙ୍କ ସମୟ", "ARST": "ଆର୍ଜେଣ୍ଟିନା ଗ୍ରୀଷ୍ମକାଳ ସମୟ", "AST": "ଆଟଲାଣ୍ଟିକ୍ ମାନାଙ୍କ ସମୟ", "WAT": "ପଶ୍ଚିମ ଆଫ୍ରିକା ମାନାଙ୍କ ସମୟ", "GFT": "ଫ୍ରେଞ୍ଚ ଗୁଆନା ସମୟ", "WART": "ପଶ୍ଚିମ ଆର୍ଜେଣ୍ଟିନା ମାନାଙ୍କ ସମୟ", "OEZ": "ପୂର୍ବାଞ୍ଚଳ ୟୁରୋପୀୟ ମାନାଙ୍କ ସମୟ", "OESZ": "ପୂର୍ବାଞ୍ଚଳ ୟୁରୋପୀୟ ଗ୍ରୀଷ୍ମକାଳ ସମୟ", "HADT": "ହୱାଇ-ଆଲେଉଟିୟ ଦିବାଲୋକ ସମୟ", "CHADT": "ଚାଥାମ୍\u200c ଦିବାଲୋକ ସମୟ", "MESZ": "କେନ୍ଦ୍ରୀୟ ୟୁରୋପୀୟ ଗ୍ରୀଷ୍ମକାଳ ସମୟ", "MST": "MST", "SRT": "ସୁରିନେମ୍\u200c ସମୟ", "HNNOMX": "ଉତ୍ତରପଶ୍ଚିମ ମେକ୍ସିକୋ ମାନାଙ୍କ ସମୟ", "MDT": "MDT", "WAST": "ପଶ୍ଚିମ ଆଫ୍ରିକା ଖରାଦିନ ସମୟ"},
	}
}

// Locale returns the current translators string locale
func (or *or_IN) Locale() string {
	return or.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'or_IN'
func (or *or_IN) PluralsCardinal() []locales.PluralRule {
	return or.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'or_IN'
func (or *or_IN) PluralsOrdinal() []locales.PluralRule {
	return or.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'or_IN'
func (or *or_IN) PluralsRange() []locales.PluralRule {
	return or.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'or_IN'
func (or *or_IN) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'or_IN'
func (or *or_IN) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 || n == 5 || n >= 7 && n <= 9 {
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

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'or_IN'
func (or *or_IN) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (or *or_IN) MonthAbbreviated(month time.Month) string {
	return or.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (or *or_IN) MonthsAbbreviated() []string {
	return or.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (or *or_IN) MonthNarrow(month time.Month) string {
	return or.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (or *or_IN) MonthsNarrow() []string {
	return or.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (or *or_IN) MonthWide(month time.Month) string {
	return or.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (or *or_IN) MonthsWide() []string {
	return or.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (or *or_IN) WeekdayAbbreviated(weekday time.Weekday) string {
	return or.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (or *or_IN) WeekdaysAbbreviated() []string {
	return or.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (or *or_IN) WeekdayNarrow(weekday time.Weekday) string {
	return or.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (or *or_IN) WeekdaysNarrow() []string {
	return or.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (or *or_IN) WeekdayShort(weekday time.Weekday) string {
	return or.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (or *or_IN) WeekdaysShort() []string {
	return or.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (or *or_IN) WeekdayWide(weekday time.Weekday) string {
	return or.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (or *or_IN) WeekdaysWide() []string {
	return or.daysWide
}

// Decimal returns the decimal point of number
func (or *or_IN) Decimal() string {
	return or.decimal
}

// Group returns the group of number
func (or *or_IN) Group() string {
	return or.group
}

// Group returns the minus sign of number
func (or *or_IN) Minus() string {
	return or.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'or_IN' and handles both Whole and Real numbers based on 'v'
func (or *or_IN) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	inSecondary := false
	groupThreshold := 3

	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, or.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {

			if count == groupThreshold {
				b = append(b, or.group[0])
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
		b = append(b, or.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'or_IN' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (or *or_IN) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 3
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, or.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, or.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, or.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'or_IN'
func (or *or_IN) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := or.currencies[currency]
	l := len(s) + len(symbol) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	inSecondary := false
	groupThreshold := 3

	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, or.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {

			if count == groupThreshold {
				b = append(b, or.group[0])
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
		b = append(b, or.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, or.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'or_IN'
// in accounting notation.
func (or *or_IN) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := or.currencies[currency]
	l := len(s) + len(symbol) + 4 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	inSecondary := false
	groupThreshold := 3

	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, or.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {

			if count == groupThreshold {
				b = append(b, or.group[0])
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

		b = append(b, or.currencyNegativePrefix[0])

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
			b = append(b, or.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, or.currencyNegativeSuffix...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'or_IN'
func (or *or_IN) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'or_IN'
func (or *or_IN) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, or.monthsAbbreviated[t.Month()]...)
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

// FmtDateLong returns the long date representation of 't' for 'or_IN'
func (or *or_IN) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, or.monthsWide[t.Month()]...)
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

// FmtDateFull returns the full date representation of 't' for 'or_IN'
func (or *or_IN) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, or.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)
	b = append(b, or.monthsWide[t.Month()]...)
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

// FmtTimeShort returns the short time representation of 't' for 'or_IN'
func (or *or_IN) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, or.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, or.periodsAbbreviated[0]...)
	} else {
		b = append(b, or.periodsAbbreviated[1]...)
	}

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'or_IN'
func (or *or_IN) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, or.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, or.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, or.periodsAbbreviated[0]...)
	} else {
		b = append(b, or.periodsAbbreviated[1]...)
	}

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'or_IN'
func (or *or_IN) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, or.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, or.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, or.periodsAbbreviated[0]...)
	} else {
		b = append(b, or.periodsAbbreviated[1]...)
	}

	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'or_IN'
func (or *or_IN) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, or.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, or.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, or.periodsAbbreviated[0]...)
	} else {
		b = append(b, or.periodsAbbreviated[1]...)
	}

	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := or.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
