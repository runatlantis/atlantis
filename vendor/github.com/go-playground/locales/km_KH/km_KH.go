package km_KH

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type km_KH struct {
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

// New returns a new instance of translator for the 'km_KH' locale
func New() locales.Translator {
	return &km_KH{
		locale:                 "km_KH",
		pluralsCardinal:        []locales.PluralRule{6},
		pluralsOrdinal:         []locales.PluralRule{6},
		pluralsRange:           []locales.PluralRule{6},
		decimal:                ",",
		group:                  ".",
		minus:                  "-",
		percent:                "%",
		perMille:               "‰",
		timeSeparator:          ":",
		inifinity:              "∞",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "XAF", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "XOF", "XPD", "XPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		currencyNegativePrefix: "(",
		currencyNegativeSuffix: ")",
		monthsAbbreviated:      []string{"", "មករា", "កុម្ភៈ", "មីនា", "មេសា", "ឧសភា", "មិថុនា", "កក្កដា", "សីហា", "កញ្ញា", "តុលា", "វិច្ឆិកា", "ធ្នូ"},
		monthsNarrow:           []string{"", "ម", "ក", "ម", "ម", "ឧ", "ម", "ក", "ស", "ក", "ត", "វ", "ធ"},
		monthsWide:             []string{"", "មករា", "កុម្ភៈ", "មីនា", "មេសា", "ឧសភា", "មិថុនា", "កក្កដា", "សីហា", "កញ្ញា", "តុលា", "វិច្ឆិកា", "ធ្នូ"},
		daysAbbreviated:        []string{"អាទិត្យ", "ច័ន្ទ", "អង្គារ", "ពុធ", "ព្រហស្បតិ៍", "សុក្រ", "សៅរ៍"},
		daysNarrow:             []string{"អ", "ច", "អ", "ព", "ព", "ស", "ស"},
		daysShort:              []string{"អា", "ច", "អ", "ពុ", "ព្រ", "សុ", "ស"},
		daysWide:               []string{"អាទិត្យ", "ច័ន្ទ", "អង្គារ", "ពុធ", "ព្រហស្បតិ៍", "សុក្រ", "សៅរ៍"},
		periodsAbbreviated:     []string{"AM", "PM"},
		periodsNarrow:          []string{"a", "p"},
		periodsWide:            []string{"AM", "PM"},
		erasAbbreviated:        []string{"មុន គ.ស.", "គ.ស."},
		erasNarrow:             []string{"", ""},
		erasWide:               []string{"មុន\u200bគ្រិស្តសករាជ", "គ្រិស្តសករាជ"},
		timezones:              map[string]string{"EST": "ម៉ោងស្តង់ដារនៅទ្វីបអាមេរិកខាងជើងភាគខាងកើត", "ACWDT": "ម៉ោង\u200bពេល\u200bថ្ងៃ\u200bនៅ\u200b\u200bភាគ\u200bខាង\u200bលិច\u200bនៃ\u200bអូស្ត្រាលី\u200bកណ្ដាល", "CLST": "ម៉ោងនៅស៊ីលីនារដូវក្តៅ", "OESZ": "ម៉ោង\u200bនៅ\u200bអឺរ៉ុប\u200b\u200bខាង\u200bកើត\u200bនា\u200bរដូវ\u200bក្ដៅ", "CHAST": "ម៉ោង\u200bស្តង់ដារ\u200bនៅ\u200bចាថាំ", "AEDT": "ម៉ោង\u200bពេល\u200bថ្ងៃ\u200bនៅ\u200bអូស្ត្រាលី\u200bខាង\u200bកើត", "BT": "ម៉ោងនៅប៊ូតង់", "NZDT": "ម៉ោង\u200bពេល\u200bថ្ងៃ\u200bនៅ\u200bនូវែលសេឡង់", "HNPMX": "ម៉ោង\u200bស្តង់ដា\u200bនៅ\u200bប៉ាស៊ីហ្វិក\u200bម៉ិកស៊ិក", "MEZ": "ម៉ោង\u200bស្តង់ដារ\u200bនៅ\u200bអឺរ៉ុប\u200bកណ្ដាល", "HEPM": "ម៉ោង\u200bពេល\u200bថ្ងៃ\u200bនៅសង់\u200bព្យែរ និង\u200bមីគុយឡុង", "WITA": "ម៉ោង\u200bនៅ\u200bឥណ្ឌូណេស៊ី\u200b\u200b\u200bកណ្ដាល", "AKST": "ម៉ោង\u200bស្តង់ដារ\u200bនៅ\u200bអាឡាស្កា", "EDT": "ម៉ោងពេលថ្ងៃនៅទ្វីបអាមេរិកខាងជើងភាគខាងកើត", "IST": "ម៉ោង\u200bស្តង់ដារនៅ\u200bឥណ្ឌា", "HNT": "ម៉ោង\u200b\u200bស្តង់ដារ\u200b\u200bនៅ\u200bញូវហ្វោនឡែន", "COST": "ម៉ោង\u200bនៅ\u200bកូឡុំប៊ីនា\u200bរដូវ\u200bក្ដៅ", "AWST": "ម៉ោង\u200b\u200bស្តង់ដារ\u200bនៅ\u200bអូស្ត្រាលី\u200bខាង\u200bលិច", "ADT": "ម៉ោង\u200bពេល\u200bថ្ងៃ\u200bនៅ\u200bអាត្លង់ទិក", "WESZ": "ម៉ោង\u200bនៅ\u200bអឺរ៉ុប\u200bខាង\u200bលិចនារដូវ\u200bក្ដៅ\u200b", "LHDT": "ម៉ោង\u200bពេល\u200bថ្ងៃ\u200bនៅ\u200bឡតហៅ", "HNPM": "ម៉ោង\u200bស្តង់ដារ\u200bនៅសង់\u200bព្យែរ និង\u200bមីគុយឡុង", "BOT": "ម៉ោង\u200bនៅ\u200bបូលីវី", "GFT": "ម៉ោង\u200bនៅ\u200bហ្គីយ៉ាន\u200bបារាំង", "MESZ": "ម៉ោង\u200bនៅ\u200bអឺរ៉ុប\u200bកណ្ដាលនា\u200bរដូវ\u200bក្ដៅ", "HKT": "ម៉ោង\u200bស្តង់ដារ\u200bនៅ\u200bហុងកុង", "HAT": "ម៉ោង\u200bពេលថ្ងៃ\u200bនៅ\u200bញូវហ្វោនឡែន", "MDT": "MDT", "COT": "ម៉ោង\u200bស្តង់ដារ\u200bនៅ\u200bកូឡុំប៊ី", "CHADT": "ម៉ោង\u200bពេល\u200bថ្ងៃ\u200bនៅ\u200bចាថាំ", "ECT": "ម៉ោង\u200bនៅ\u200bអេក្វាទ័រ", "HKST": "ម៉ោងនៅ\u200bហុងកុងនា\u200bរដូវ\u200bក្ដៅ\u200b", "CAT": "ម៉ោង\u200bនៅ\u200bអាហ្វ្រិក\u200bកណ្ដាល", "OEZ": "ម៉ោង\u200bស្តង់ដារ\u200b\u200bនៅ\u200bអឺរ៉ុប\u200b\u200bខាង\u200bកើត\u200b", "AKDT": "ម៉ោង\u200bពេល\u200bថ្ងៃ\u200bនៅ\u200b\u200bអាឡាស្កា", "HEEG": "ម៉ោង\u200bនៅ\u200bហ្គ្រីនលែនខាង\u200bកើតនា\u200bរដូវ\u200bក្ដៅ", "PDT": "ម៉ោង\u200b\u200bពេលថ្ងៃនៅទ្វីប\u200bអាមរិក\u200bខាង\u200bជើងភាគខាងលិច", "SAST": "ម៉ោង\u200bនៅ\u200bអាហ្វ្រិក\u200bខាង\u200bត្បូង", "NZST": "ម៉ោង\u200bស្តង់ដារ\u200bនៅ\u200bនូវែលសេឡង់", "WART": "ម៉ោងស្តង់ដារ\u200bនៅ\u200bអាហ្សង់ទីនភាគខាងលិច", "EAT": "ម៉ោង\u200bនៅ\u200bអាហ្វ្រិក\u200bខាង\u200bកើត", "CLT": "ម៉ោងស្តង់ដារនៅស៊ីលី", "UYT": "ម៉ោង\u200bស្តង់ដារ\u200bនៅ\u200bអ៊ុយរូហ្គាយ", "HNCU": "ម៉ោង\u200bស្តង់ដារ\u200bនៅ\u200bគុយបា", "WARST": "ម៉ោង\u200bនៅ\u200bអាហ្សង់ទីនភាគខាងលិចនារដូវក្តៅ", "ACST": "ម៉ោង\u200bស្តង់ដារ\u200bនៅ\u200bអូស្ត្រាលី\u200bកណ្ដាល", "ACWST": "ម៉ោង\u200bស្តង់ដារ\u200bនៅ\u200bភាគ\u200bខាង\u200bលិច\u200bនៃ\u200bអូស្ត្រាលី\u200bកណ្ដាល", "MST": "MST", "PST": "ម៉ោងស្តង់ដារ\u200b\u200b\u200bនៅទ្វីប\u200bអាមរិក\u200bខាង\u200bជើងភាគខាងលិច", "WEZ": "ម៉ោង\u200bស្តង់ដារ\u200bនៅ\u200bអឺរ៉ុប\u200bខាង\u200bលិច", "MYT": "ម៉ោង\u200bនៅ\u200bម៉ាឡេស៊ី", "WIB": "ម៉ោង\u200bនៅ\u200bឥណ្ឌូណេស៊ី\u200b\u200bខាង\u200bលិច", "HEOG": "ម៉ោងនៅហ្គ្រីនលែនខាងលិចនារដូវក្តៅ", "HAST": "ម៉ោង\u200bស្តង់ដារ\u200b\u200bនៅ\u200bហាវៃ-អាល់ដ្យូសិន", "ARST": "ម៉ោង\u200bនៅ\u200bអាហ្សង់ទីននារដូវក្តៅ", "ChST": "ម៉ោង\u200bស្តង់ដារនៅ\u200bចាំម៉ូរ៉ូ", "AST": "ម៉ោង\u200bស្តង់ដារ\u200bនៅ\u200bអាត្លង់ទិក", "JDT": "ម៉ោង\u200bពេល\u200bថ្ងៃ\u200bនៅជប៉ុន", "HNOG": "ម៉ោងស្តង់ដារនៅហ្គ្រីនលែនខាងលិច", "HNNOMX": "ម៉ោង\u200bស្តង់ដារនៅ\u200bម៉ិកស៊ិកភាគពាយព្យ", "SRT": "ម៉ោង\u200bនៅ\u200bសូរីណាម", "TMST": "ម៉ោង\u200bរដូវ\u200bក្ដៅ\u200bនៅ\u200bតួកម៉េនីស្ថាន\u200b", "ART": "ម៉ោងស្តង់ដារ\u200bនៅ\u200bអាហ្សង់ទីន", "ACDT": "ម៉ោង\u200bពេលថ្ងៃ\u200b\u200b\u200b\u200bនៅ\u200bអូស្ត្រាលី\u200bកណ្ដាល", "HNEG": "ម៉ោង\u200b\u200b\u200bស្តង់ដារ\u200bនៅ\u200b\u200bហ្គ្រីនលែន\u200bខាង\u200bកើត", "VET": "ម៉ោង\u200bនៅ\u200bវ៉េណេស៊ុយអេឡា", "UYST": "ម៉ោង\u200bនៅ\u200bអ៊ុយរូហ្គាយនា\u200b\u200bរដូវ\u200bក្ដៅ", "CST": "ម៉ោង\u200b\u200bស្តង់ដារនៅ\u200bទ្វីបអាមេរិក\u200bខាង\u200bជើងភាគកណ្តាល", "AEST": "ម៉ោង\u200bស្តង់ដារ\u200bនៅ\u200bអូស្ត្រាលី\u200bខាង\u200bកើត", "JST": "ម៉ោង\u200bស្តង់ដារ\u200bនៅ\u200bជប៉ុន", "WAST": "ម៉ោងនៅ\u200bអាហ្វ្រិក\u200b\u200b\u200bខាងលិច\u200b\u200bនារដូវ\u200bក្ដៅ\u200b", "WIT": "ម៉ោង\u200bនៅ\u200bឥណ្ឌូណេស៊ី\u200b\u200bខាង\u200bកើត", "∅∅∅": "ម៉ោង\u200bនៅ\u200bអាម៉ាហ្សូននារដូវក្តៅ", "GYT": "ម៉ោង\u200bនៅ\u200bហ្គីយ៉ាន", "AWDT": "ម៉ោង\u200bពេល\u200bថ្ងៃ\u200bនៅ\u200bអូស្ត្រាលី\u200bខាង\u200bលិច", "CDT": "ម៉ោង\u200b\u200bពេលថ្ងៃនៅ\u200bទ្វីបអាមេរិក\u200bខាង\u200bជើងភាគកណ្តាល", "WAT": "ម៉ោង\u200bស្តង់ដារ\u200bនៅ\u200bអាហ្វ្រិក\u200bខាង\u200bលិច", "HENOMX": "ម៉ោង\u200bពេល\u200bថ្ងៃ\u200bនៅ\u200bម៉ិកស៊ិកភាគពាយព្យ", "TMT": "ម៉ោង\u200bស្តង់ដារ\u200bនៅតួកម៉េនីស្ថាន", "HADT": "ម៉ោង\u200bពេល\u200bថ្ងៃ\u200bនៅ\u200bហាវៃ-អាល់ដ្យូសិន", "GMT": "ម៉ោងនៅគ្រីនវិច", "HECU": "ម៉ោង\u200bពេល\u200bថ្ងៃ\u200bនៅ\u200bគុយបា", "HEPMX": "ម៉ោង\u200bពេល\u200bថ្ងៃ\u200bនៅ\u200bប៉ាស៊ីហ្វិក\u200bម៉ិកស៊ិក", "SGT": "ម៉ោង\u200bនៅ\u200bសិង្ហបូរី", "LHST": "ម៉ោង\u200bស្តង់ដារ\u200bនៅ\u200bឡត\u200bហៅ"},
	}
}

// Locale returns the current translators string locale
func (km *km_KH) Locale() string {
	return km.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'km_KH'
func (km *km_KH) PluralsCardinal() []locales.PluralRule {
	return km.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'km_KH'
func (km *km_KH) PluralsOrdinal() []locales.PluralRule {
	return km.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'km_KH'
func (km *km_KH) PluralsRange() []locales.PluralRule {
	return km.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'km_KH'
func (km *km_KH) CardinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'km_KH'
func (km *km_KH) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'km_KH'
func (km *km_KH) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return locales.PluralRuleOther
}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (km *km_KH) MonthAbbreviated(month time.Month) string {
	return km.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (km *km_KH) MonthsAbbreviated() []string {
	return km.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (km *km_KH) MonthNarrow(month time.Month) string {
	return km.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (km *km_KH) MonthsNarrow() []string {
	return km.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (km *km_KH) MonthWide(month time.Month) string {
	return km.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (km *km_KH) MonthsWide() []string {
	return km.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (km *km_KH) WeekdayAbbreviated(weekday time.Weekday) string {
	return km.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (km *km_KH) WeekdaysAbbreviated() []string {
	return km.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (km *km_KH) WeekdayNarrow(weekday time.Weekday) string {
	return km.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (km *km_KH) WeekdaysNarrow() []string {
	return km.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (km *km_KH) WeekdayShort(weekday time.Weekday) string {
	return km.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (km *km_KH) WeekdaysShort() []string {
	return km.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (km *km_KH) WeekdayWide(weekday time.Weekday) string {
	return km.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (km *km_KH) WeekdaysWide() []string {
	return km.daysWide
}

// Decimal returns the decimal point of number
func (km *km_KH) Decimal() string {
	return km.decimal
}

// Group returns the group of number
func (km *km_KH) Group() string {
	return km.group
}

// Group returns the minus sign of number
func (km *km_KH) Minus() string {
	return km.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'km_KH' and handles both Whole and Real numbers based on 'v'
func (km *km_KH) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, km.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, km.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, km.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'km_KH' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (km *km_KH) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 3
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, km.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, km.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, km.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'km_KH'
func (km *km_KH) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := km.currencies[currency]
	l := len(s) + len(symbol) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, km.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, km.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, km.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, km.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	b = append(b, symbol...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'km_KH'
// in accounting notation.
func (km *km_KH) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := km.currencies[currency]
	l := len(s) + len(symbol) + 4 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, km.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, km.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {

		b = append(b, km.currencyNegativePrefix[0])

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, km.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, km.currencyNegativeSuffix...)
		b = append(b, symbol...)
	} else {

		b = append(b, symbol...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'km_KH'
func (km *km_KH) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'km_KH'
func (km *km_KH) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, km.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'km_KH'
func (km *km_KH) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, km.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'km_KH'
func (km *km_KH) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, km.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, km.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'km_KH'
func (km *km_KH) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, km.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, km.periodsAbbreviated[0]...)
	} else {
		b = append(b, km.periodsAbbreviated[1]...)
	}

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'km_KH'
func (km *km_KH) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, km.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, km.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, km.periodsAbbreviated[0]...)
	} else {
		b = append(b, km.periodsAbbreviated[1]...)
	}

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'km_KH'
func (km *km_KH) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, km.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, km.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, km.periodsAbbreviated[0]...)
	} else {
		b = append(b, km.periodsAbbreviated[1]...)
	}

	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'km_KH'
func (km *km_KH) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, km.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, km.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, km.periodsAbbreviated[0]...)
	} else {
		b = append(b, km.periodsAbbreviated[1]...)
	}

	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := km.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
