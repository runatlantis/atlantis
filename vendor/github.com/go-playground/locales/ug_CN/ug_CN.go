package ug_CN

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type ug_CN struct {
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

// New returns a new instance of translator for the 'ug_CN' locale
func New() locales.Translator {
	return &ug_CN{
		locale:                 "ug_CN",
		pluralsCardinal:        []locales.PluralRule{2, 6},
		pluralsOrdinal:         nil,
		pluralsRange:           []locales.PluralRule{2, 6},
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
		monthsAbbreviated:      []string{"", "يانۋار", "فېۋرال", "مارت", "ئاپرېل", "ماي", "ئىيۇن", "ئىيۇل", "ئاۋغۇست", "سېنتەبىر", "ئۆكتەبىر", "نويابىر", "دېكابىر"},
		monthsNarrow:           []string{"", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12"},
		monthsWide:             []string{"", "يانۋار", "فېۋرال", "مارت", "ئاپرېل", "ماي", "ئىيۇن", "ئىيۇل", "ئاۋغۇست", "سېنتەبىر", "ئۆكتەبىر", "نويابىر", "دېكابىر"},
		daysAbbreviated:        []string{"يە", "دۈ", "سە", "چا", "پە", "جۈ", "شە"},
		daysNarrow:             []string{"ي", "د", "س", "چ", "پ", "ج", "ش"},
		daysShort:              []string{"ي", "د", "س", "چ", "پ", "ج", "ش"},
		daysWide:               []string{"يەكشەنبە", "دۈشەنبە", "سەيشەنبە", "چارشەنبە", "پەيشەنبە", "جۈمە", "شەنبە"},
		periodsAbbreviated:     []string{"چ.ب", "چ.ك"},
		periodsNarrow:          []string{"ب", "ك"},
		periodsWide:            []string{"چۈشتىن بۇرۇن", "چۈشتىن كېيىن"},
		erasAbbreviated:        []string{"BCE", "مىلادىيە"},
		erasNarrow:             []string{"BCE", "مىلادىيە"},
		erasWide:               []string{"مىلادىيەدىن بۇرۇن", "مىلادىيە"},
		timezones:              map[string]string{"CST": "ئوتتۇرا قىسىم ئۆلچەملىك ۋاقتى", "CDT": "ئوتتۇرا قىسىم يازلىق ۋاقتى", "WAT": "غەربىي ئافرىقا ئۆلچەملىك ۋاقتى", "VET": "ۋېنېزۇئېلا ۋاقتى", "EAT": "شەرقىي ئافرىقا ۋاقتى", "COT": "كولومبىيە ئۆلچەملىك ۋاقتى", "COST": "كولومبىيە يازلىق ۋاقتى", "GYT": "گىۋىيانا ۋاقتى", "WAST": "غەربىي ئافرىقا يازلىق ۋاقتى", "WEZ": "غەربىي ياۋروپا ئۆلچەملىك ۋاقتى", "WESZ": "غەربىي ياۋروپا يازلىق ۋاقتى", "SGT": "سىنگاپور ۋاقتى", "HEOG": "غەربىي گىرېنلاند يازلىق ۋاقتى", "MESZ": "ئوتتۇرا ياۋروپا يازلىق ۋاقتى", "WITA": "ئوتتۇرا ھىندونېزىيە ۋاقتى", "ARST": "ئارگېنتىنا يازلىق ۋاقتى", "UYT": "ئۇرۇگۋاي ئۆلچەملىك ۋاقتى", "BT": "بۇتان ۋاقتى", "ECT": "ئېكۋادور ۋاقتى", "EDT": "شەرقىي قىسىم يازلىق ۋاقتى", "HNEG": "شەرقىي گىرېنلاند ئۆلچەملىك ۋاقتى", "CAT": "ئوتتۇرا ئافرىقا ۋاقتى", "HAST": "ھاۋاي-ئالېيۇت ئۆلچەملىك ۋاقتى", "AKDT": "ئالياسكا يازلىق ۋاقتى", "ACST": "ئاۋسترالىيە ئوتتۇرا قىسىم ئۆلچەملىك ۋاقتى", "WIT": "شەرقىي ھىندونېزىيە ۋاقتى", "NZDT": "يېڭى زېلاندىيە يازلىق ۋاقتى", "MST": "ئاۋمېن ئۆلچەملىك ۋاقتى", "SRT": "سۇرىنام ۋاقتى", "ART": "ئارگېنتىنا ئۆلچەملىك ۋاقتى", "CHADT": "چاتام يازلىق ۋاقتى", "HEPMX": "مېكسىكا تىنچ ئوكيان يازلىق ۋاقتى", "AEST": "ئاۋسترالىيە شەرقىي قىسىم ئۆلچەملىك ۋاقتى", "MYT": "مالايشىيا ۋاقتى", "AKST": "ئالياسكا ئۆلچەملىك ۋاقتى", "HKST": "شياڭگاڭ يازلىق ۋاقتى", "WART": "غەربىي ئارگېنتىنا ئۆلچەملىك ۋاقتى", "HNNOMX": "مېكسىكا غەربىي شىمالىي قىسىم ئۆلچەملىك ۋاقتى", "HECU": "كۇبا يازلىق ۋاقتى", "MDT": "ئاۋمېن يازلىق ۋاقتى", "GMT": "گىرىنۋىچ ۋاقتى", "HNCU": "كۇبا ئۆلچەملىك ۋاقتى", "AWST": "ئاۋسترالىيە غەربىي قىسىم ئۆلچەملىك ۋاقتى", "SAST": "جەنۇبىي ئافرىقا ئۆلچەملىك ۋاقتى", "∅∅∅": "ئاكرى يازلىق ۋاقتى", "WARST": "غەربىي ئارگېنتىنا يازلىق ۋاقتى", "HNPM": "ساينىت پىئېر ۋە مىكېلون ئۆلچەملىك ۋاقتى", "UYST": "ئۇرۇگۋاي يازلىق ۋاقتى", "PDT": "تىنچ ئوكيان يازلىق ۋاقتى", "WIB": "غەربىي ھىندونېزىيە ۋاقتى", "JDT": "ياپونىيە يازلىق ۋاقتى", "LHST": "لورد-خاي ئۆلچەملىك ۋاقتى", "LHDT": "لورد-خاي يازلىق ۋاقتى", "OEZ": "شەرقىي ياۋروپا ئۆلچەملىك ۋاقتى", "HADT": "ھاۋاي-ئالېيۇت يازلىق ۋاقتى", "ACDT": "ئاۋسترالىيە ئوتتۇرا قىسىم يازلىق ۋاقتى", "CLST": "چىلى يازلىق ۋاقتى", "PST": "تىنچ ئوكيان ئۆلچەملىك ۋاقتى", "AWDT": "ئاۋسترالىيە غەربىي قىسىم يازلىق ۋاقتى", "JST": "ياپونىيە ئۆلچەملىك ۋاقتى", "HEEG": "شەرقىي گىرېنلاند يازلىق ۋاقتى", "HAT": "نىۋفوئۇنلاند يازلىق ۋاقتى", "TMST": "تۈركمەنىستان يازلىق ۋاقتى", "HNPMX": "مېكسىكا تىنچ ئوكيان ئۆلچەملىك ۋاقتى", "ADT": "ئاتلانتىك ئوكيان يازلىق ۋاقتى", "HNOG": "غەربىي گىرېنلاند ئۆلچەملىك ۋاقتى", "HNT": "نىۋفوئۇنلاند ئۆلچەملىك ۋاقتى", "ChST": "چاموررو ئۆلچەملىك ۋاقتى", "NZST": "يېڭى زېلاندىيە ئۆلچەملىك ۋاقتى", "ACWST": "ئاۋستىرالىيە ئوتتۇرا غەربىي قىسىم ئۆلچەملىك ۋاقتى", "MEZ": "ئوتتۇرا ياۋروپا ئۆلچەملىك ۋاقتى", "IST": "ھىندىستان ئۆلچەملىك ۋاقتى", "CHAST": "چاتام ئۆلچەملىك ۋاقتى", "OESZ": "شەرقىي ياۋروپا يازلىق ۋاقتى", "AEDT": "ئاۋسترالىيە شەرقىي قىسىم يازلىق ۋاقتى", "BOT": "بولىۋىيە ۋاقتى", "GFT": "فىرانسىيەگە قاراشلىق گىۋىيانا ۋاقتى", "EST": "شەرقىي قىسىم ئۆلچەملىك ۋاقتى", "ACWDT": "ئاۋسترالىيە ئوتتۇرا غەربىي قىسىم يازلىق ۋاقتى", "HEPM": "ساينىت پىئېر ۋە مىكېلون يازلىق ۋاقتى", "CLT": "چىلى ئۆلچەملىك ۋاقتى", "HKT": "شياڭگاڭ ئۆلچەملىك ۋاقتى", "HENOMX": "مېكسىكا غەربىي شىمالىي قىسىم يازلىق ۋاقتى", "TMT": "تۈركمەنىستان ئۆلچەملىك ۋاقتى", "AST": "ئاتلانتىك ئوكيان ئۆلچەملىك ۋاقتى"},
	}
}

// Locale returns the current translators string locale
func (ug *ug_CN) Locale() string {
	return ug.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'ug_CN'
func (ug *ug_CN) PluralsCardinal() []locales.PluralRule {
	return ug.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'ug_CN'
func (ug *ug_CN) PluralsOrdinal() []locales.PluralRule {
	return ug.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'ug_CN'
func (ug *ug_CN) PluralsRange() []locales.PluralRule {
	return ug.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'ug_CN'
func (ug *ug_CN) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'ug_CN'
func (ug *ug_CN) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return locales.PluralRuleUnknown
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'ug_CN'
func (ug *ug_CN) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {

	start := ug.CardinalPluralRule(num1, v1)
	end := ug.CardinalPluralRule(num2, v2)

	if start == locales.PluralRuleOne && end == locales.PluralRuleOther {
		return locales.PluralRuleOther
	} else if start == locales.PluralRuleOther && end == locales.PluralRuleOne {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther

}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (ug *ug_CN) MonthAbbreviated(month time.Month) string {
	return ug.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (ug *ug_CN) MonthsAbbreviated() []string {
	return ug.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (ug *ug_CN) MonthNarrow(month time.Month) string {
	return ug.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (ug *ug_CN) MonthsNarrow() []string {
	return ug.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (ug *ug_CN) MonthWide(month time.Month) string {
	return ug.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (ug *ug_CN) MonthsWide() []string {
	return ug.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (ug *ug_CN) WeekdayAbbreviated(weekday time.Weekday) string {
	return ug.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (ug *ug_CN) WeekdaysAbbreviated() []string {
	return ug.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (ug *ug_CN) WeekdayNarrow(weekday time.Weekday) string {
	return ug.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (ug *ug_CN) WeekdaysNarrow() []string {
	return ug.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (ug *ug_CN) WeekdayShort(weekday time.Weekday) string {
	return ug.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (ug *ug_CN) WeekdaysShort() []string {
	return ug.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (ug *ug_CN) WeekdayWide(weekday time.Weekday) string {
	return ug.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (ug *ug_CN) WeekdaysWide() []string {
	return ug.daysWide
}

// Decimal returns the decimal point of number
func (ug *ug_CN) Decimal() string {
	return ug.decimal
}

// Group returns the group of number
func (ug *ug_CN) Group() string {
	return ug.group
}

// Group returns the minus sign of number
func (ug *ug_CN) Minus() string {
	return ug.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'ug_CN' and handles both Whole and Real numbers based on 'v'
func (ug *ug_CN) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ug.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, ug.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, ug.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'ug_CN' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (ug *ug_CN) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 3
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ug.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, ug.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, ug.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'ug_CN'
func (ug *ug_CN) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := ug.currencies[currency]
	l := len(s) + len(symbol) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ug.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, ug.group[0])
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
		b = append(b, ug.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, ug.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'ug_CN'
// in accounting notation.
func (ug *ug_CN) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := ug.currencies[currency]
	l := len(s) + len(symbol) + 4 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ug.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, ug.group[0])
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

		b = append(b, ug.currencyNegativePrefix[0])

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
			b = append(b, ug.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, ug.currencyNegativeSuffix...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'ug_CN'
func (ug *ug_CN) FmtDateShort(t time.Time) string {

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

// FmtDateMedium returns the medium date representation of 't' for 'ug_CN'
func (ug *ug_CN) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2d}...)
	b = append(b, ug.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0xd8, 0x8c, 0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'ug_CN'
func (ug *ug_CN) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2d}...)
	b = append(b, ug.monthsWide[t.Month()]...)
	b = append(b, []byte{0xd8, 0x8c, 0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'ug_CN'
func (ug *ug_CN) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	b = append(b, []byte{0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2d}...)
	b = append(b, ug.monthsWide[t.Month()]...)
	b = append(b, []byte{0xd8, 0x8c, 0x20}...)
	b = append(b, ug.daysWide[t.Weekday()]...)

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'ug_CN'
func (ug *ug_CN) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, ug.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, ug.periodsAbbreviated[0]...)
	} else {
		b = append(b, ug.periodsAbbreviated[1]...)
	}

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'ug_CN'
func (ug *ug_CN) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, ug.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, ug.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, ug.periodsAbbreviated[0]...)
	} else {
		b = append(b, ug.periodsAbbreviated[1]...)
	}

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'ug_CN'
func (ug *ug_CN) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, ug.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, ug.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, ug.periodsAbbreviated[0]...)
	} else {
		b = append(b, ug.periodsAbbreviated[1]...)
	}

	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'ug_CN'
func (ug *ug_CN) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	h := t.Hour()

	if h > 12 {
		h -= 12
	}

	b = strconv.AppendInt(b, int64(h), 10)
	b = append(b, ug.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, ug.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	if t.Hour() < 12 {
		b = append(b, ug.periodsAbbreviated[0]...)
	} else {
		b = append(b, ug.periodsAbbreviated[1]...)
	}

	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := ug.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
