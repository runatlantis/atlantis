package ro

import (
	"math"
	"strconv"
	"time"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
)

type ro struct {
	locale                 string
	pluralsCardinal        []locales.PluralRule
	pluralsOrdinal         []locales.PluralRule
	pluralsRange           []locales.PluralRule
	decimal                string
	group                  string
	minus                  string
	percent                string
	percentSuffix          string
	perMille               string
	timeSeparator          string
	inifinity              string
	currencies             []string // idx = enum of currency code
	currencyPositiveSuffix string
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

// New returns a new instance of translator for the 'ro' locale
func New() locales.Translator {
	return &ro{
		locale:                 "ro",
		pluralsCardinal:        []locales.PluralRule{2, 4, 6},
		pluralsOrdinal:         []locales.PluralRule{2, 6},
		pluralsRange:           []locales.PluralRule{4, 6},
		decimal:                ",",
		group:                  ".",
		minus:                  "-",
		percent:                "%",
		perMille:               "‰",
		timeSeparator:          ":",
		inifinity:              "∞",
		currencies:             []string{"ADP", "AED", "AFA", "AFN", "ALK", "ALL", "AMD", "ANG", "AOA", "AOK", "AON", "AOR", "ARA", "ARL", "ARM", "ARP", "ARS", "ATS", "AUD", "AWG", "AZM", "AZN", "BAD", "BAM", "BAN", "BBD", "BDT", "BEC", "BEF", "BEL", "BGL", "BGM", "BGN", "BGO", "BHD", "BIF", "BMD", "BND", "BOB", "BOL", "BOP", "BOV", "BRB", "BRC", "BRE", "BRL", "BRN", "BRR", "BRZ", "BSD", "BTN", "BUK", "BWP", "BYB", "BYN", "BYR", "BZD", "CAD", "CDF", "CHE", "CHF", "CHW", "CLE", "CLF", "CLP", "CNH", "CNX", "CNY", "COP", "COU", "CRC", "CSD", "CSK", "CUC", "CUP", "CVE", "CYP", "CZK", "DDM", "DEM", "DJF", "DKK", "DOP", "DZD", "ECS", "ECV", "EEK", "EGP", "ERN", "ESA", "ESB", "ESP", "ETB", "EUR", "FIM", "FJD", "FKP", "FRF", "GBP", "GEK", "GEL", "GHC", "GHS", "GIP", "GMD", "GNF", "GNS", "GQE", "GRD", "GTQ", "GWE", "GWP", "GYD", "HKD", "HNL", "HRD", "HRK", "HTG", "HUF", "IDR", "IEP", "ILP", "ILR", "ILS", "INR", "IQD", "IRR", "ISJ", "ISK", "ITL", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRH", "KRO", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LTL", "LTT", "LUC", "LUF", "LUL", "LVL", "LVR", "LYD", "MAD", "MAF", "MCF", "MDC", "MDL", "MGA", "MGF", "MKD", "MKN", "MLF", "MMK", "MNT", "MOP", "MRO", "MTL", "MTP", "MUR", "MVP", "MVR", "MWK", "MXN", "MXP", "MXV", "MYR", "MZE", "MZM", "MZN", "NAD", "NGN", "NIC", "NIO", "NLG", "NOK", "NPR", "NZD", "OMR", "PAB", "PEI", "PEN", "PES", "PGK", "PHP", "PKR", "PLN", "PLZ", "PTE", "PYG", "QAR", "RHD", "ROL", "RON", "RSD", "RUB", "RUR", "RWF", "SAR", "SBD", "SCR", "SDD", "SDG", "SDP", "SEK", "SGD", "SHP", "SIT", "SKK", "SLL", "SOS", "SRD", "SRG", "SSP", "STD", "STN", "SUR", "SVC", "SYP", "SZL", "THB", "TJR", "TJS", "TMM", "TMT", "TND", "TOP", "TPE", "TRL", "TRY", "TTD", "TWD", "TZS", "UAH", "UAK", "UGS", "UGX", "USD", "USN", "USS", "UYI", "UYP", "UYU", "UZS", "VEB", "VEF", "VND", "VNN", "VUV", "WST", "FCFA", "XAG", "XAU", "XBA", "XBB", "XBC", "XBD", "XCD", "XDR", "XEU", "XFO", "XFU", "CFA", "XPD", "CFPF", "XPT", "XRE", "XSU", "XTS", "XUA", "XXX", "YDD", "YER", "YUD", "YUM", "YUN", "YUR", "ZAL", "ZAR", "ZMK", "ZMW", "ZRN", "ZRZ", "ZWD", "ZWL", "ZWR"},
		percentSuffix:          " ",
		currencyPositiveSuffix: " ",
		currencyNegativePrefix: "(",
		currencyNegativeSuffix: " )",
		monthsAbbreviated:      []string{"", "ian.", "feb.", "mar.", "apr.", "mai", "iun.", "iul.", "aug.", "sept.", "oct.", "nov.", "dec."},
		monthsNarrow:           []string{"", "I", "F", "M", "A", "M", "I", "I", "A", "S", "O", "N", "D"},
		monthsWide:             []string{"", "ianuarie", "februarie", "martie", "aprilie", "mai", "iunie", "iulie", "august", "septembrie", "octombrie", "noiembrie", "decembrie"},
		daysAbbreviated:        []string{"dum.", "lun.", "mar.", "mie.", "joi", "vin.", "sâm."},
		daysNarrow:             []string{"D", "L", "M", "M", "J", "V", "S"},
		daysShort:              []string{"du.", "lu.", "ma.", "mi.", "joi", "vi.", "sâ."},
		daysWide:               []string{"duminică", "luni", "marți", "miercuri", "joi", "vineri", "sâmbătă"},
		periodsAbbreviated:     []string{"a.m.", "p.m."},
		periodsNarrow:          []string{"a.m.", "p.m."},
		periodsWide:            []string{"a.m.", "p.m."},
		erasAbbreviated:        []string{"î.Hr.", "d.Hr."},
		erasNarrow:             []string{"î.Hr.", "d.Hr."},
		erasWide:               []string{"înainte de Hristos", "după Hristos"},
		timezones:              map[string]string{"HNEG": "Ora standard a Groenlandei orientale", "EST": "Ora standard orientală nord-americană", "TMT": "Ora standard din Turkmenistan", "HADT": "Ora de vară din Hawaii-Aleutine", "GMT": "Ora de Greenwhich", "PDT": "Ora de vară în zona Pacific nord-americană", "JST": "Ora standard a Japoniei", "CLT": "Ora standard din Chile", "MEZ": "Ora standard a Europei Centrale", "WARST": "Ora de vară a Argentinei Occidentale", "PST": "Ora standard în zona Pacific nord-americană", "ACDT": "Ora de vară a Australiei Centrale", "VET": "Ora Venezuelei", "WIT": "Ora Indoneziei de Est", "COST": "Ora de vară a Columbiei", "CHAST": "Ora standard din Chatham", "HEPMX": "Ora de vară a zonei Pacific mexicane", "CDT": "Ora de vară centrală nord-americană", "GFT": "Ora din Guyana Franceză", "HKST": "Ora de vară din Hong Kong", "HNPM": "Ora standard din Saint-Pierre și Miquelon", "UYT": "Ora standard a Uruguayului", "ChST": "Ora din Chamorro", "MST": "Ora standard în zona montană nord-americană", "WEZ": "Ora standard a Europei de Vest", "EDT": "Ora de vară orientală nord-americană", "CLST": "Ora de vară din Chile", "UYST": "Ora de vară a Uruguayului", "HECU": "Ora de vară a Cubei", "HEEG": "Ora de vară a Groenlandei orientale", "HEOG": "Ora de vară a Groenlandei occidentale", "NZST": "Ora standard a Noii Zeelande", "WITA": "Ora Indoneziei Centrale", "HAT": "Ora de vară din Newfoundland", "EAT": "Ora Africii Orientale", "ART": "Ora standard a Argentinei", "AEDT": "Ora de vară a Australiei Orientale", "MDT": "Ora de vară în zona montană nord-americană", "BOT": "Ora Boliviei", "LHDT": "Ora de vară din Lord Howe", "OEZ": "Ora standard a Europei de Est", "AEST": "Ora standard a Australiei Orientale", "SGT": "Ora din Singapore", "ACWDT": "Ora de vară a Australiei Central Occidentale", "HKT": "Ora standard din Hong Kong", "COT": "Ora standard a Columbiei", "OESZ": "Ora de vară a Europei de Est", "CHADT": "Ora de vară din Chatham", "HNNOMX": "Ora standard a Mexicului de nord-vest", "ARST": "Ora de vară a Argentinei", "SAST": "Ora Africii Meridionale", "JDT": "Ora de vară a Japoniei", "ACWST": "Ora standard a Australiei Central Occidentale", "WART": "Ora standard a Argentinei Occidentale", "BT": "Ora Bhutanului", "ACST": "Ora standard a Australiei Centrale", "HAST": "Ora standard din Hawaii-Aleutine", "HNCU": "Ora standard a Cubei", "AST": "Ora standard în zona Atlantic nord-americană", "ADT": "Ora de vară în zona Atlantic nord-americană", "WAST": "Ora de vară a Africii Occidentale", "CST": "Ora standard centrală nord-americană", "AWDT": "Ora de vară a Australiei Occidentale", "WIB": "Ora Indoneziei de Vest", "NZDT": "Ora de vară a Noii Zeelande", "HEPM": "Ora de vară din Saint-Pierre și Miquelon", "HENOMX": "Ora de vară a Mexicului de nord-vest", "TMST": "Ora de vară din Turkmenistan", "MYT": "Ora din Malaysia", "AKDT": "Ora de vară din Alaska", "LHST": "Ora standard din Lord Howe", "HNT": "Ora standard din Newfoundland", "MESZ": "Ora de vară a Europei Centrale", "GYT": "Ora din Guyana", "HNPMX": "Ora standard a zonei Pacific mexicane", "WAT": "Ora standard a Africii Occidentale", "AKST": "Ora standard din Alaska", "ECT": "Ora Ecuadorului", "CAT": "Ora Africii Centrale", "IST": "Ora Indiei", "SRT": "Ora Surinamului", "∅∅∅": "Ora de vară a Amazonului", "AWST": "Ora standard a Australiei Occidentale", "WESZ": "Ora de vară a Europei de Vest", "HNOG": "Ora standard a Groenlandei occidentale"},
	}
}

// Locale returns the current translators string locale
func (ro *ro) Locale() string {
	return ro.locale
}

// PluralsCardinal returns the list of cardinal plural rules associated with 'ro'
func (ro *ro) PluralsCardinal() []locales.PluralRule {
	return ro.pluralsCardinal
}

// PluralsOrdinal returns the list of ordinal plural rules associated with 'ro'
func (ro *ro) PluralsOrdinal() []locales.PluralRule {
	return ro.pluralsOrdinal
}

// PluralsRange returns the list of range plural rules associated with 'ro'
func (ro *ro) PluralsRange() []locales.PluralRule {
	return ro.pluralsRange
}

// CardinalPluralRule returns the cardinal PluralRule given 'num' and digits/precision of 'v' for 'ro'
func (ro *ro) CardinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)
	i := int64(n)
	nMod100 := math.Mod(n, 100)

	if i == 1 && v == 0 {
		return locales.PluralRuleOne
	} else if (v != 0) || (n == 0) || (n != 1 && nMod100 >= 1 && nMod100 <= 19) {
		return locales.PluralRuleFew
	}

	return locales.PluralRuleOther
}

// OrdinalPluralRule returns the ordinal PluralRule given 'num' and digits/precision of 'v' for 'ro'
func (ro *ro) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {

	n := math.Abs(num)

	if n == 1 {
		return locales.PluralRuleOne
	}

	return locales.PluralRuleOther
}

// RangePluralRule returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for 'ro'
func (ro *ro) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {

	start := ro.CardinalPluralRule(num1, v1)
	end := ro.CardinalPluralRule(num2, v2)

	if start == locales.PluralRuleOne && end == locales.PluralRuleFew {
		return locales.PluralRuleFew
	} else if start == locales.PluralRuleOne && end == locales.PluralRuleOther {
		return locales.PluralRuleOther
	} else if start == locales.PluralRuleFew && end == locales.PluralRuleOne {
		return locales.PluralRuleFew
	} else if start == locales.PluralRuleFew && end == locales.PluralRuleFew {
		return locales.PluralRuleFew
	} else if start == locales.PluralRuleFew && end == locales.PluralRuleOther {
		return locales.PluralRuleOther
	} else if start == locales.PluralRuleOther && end == locales.PluralRuleFew {
		return locales.PluralRuleFew
	}

	return locales.PluralRuleOther

}

// MonthAbbreviated returns the locales abbreviated month given the 'month' provided
func (ro *ro) MonthAbbreviated(month time.Month) string {
	return ro.monthsAbbreviated[month]
}

// MonthsAbbreviated returns the locales abbreviated months
func (ro *ro) MonthsAbbreviated() []string {
	return ro.monthsAbbreviated[1:]
}

// MonthNarrow returns the locales narrow month given the 'month' provided
func (ro *ro) MonthNarrow(month time.Month) string {
	return ro.monthsNarrow[month]
}

// MonthsNarrow returns the locales narrow months
func (ro *ro) MonthsNarrow() []string {
	return ro.monthsNarrow[1:]
}

// MonthWide returns the locales wide month given the 'month' provided
func (ro *ro) MonthWide(month time.Month) string {
	return ro.monthsWide[month]
}

// MonthsWide returns the locales wide months
func (ro *ro) MonthsWide() []string {
	return ro.monthsWide[1:]
}

// WeekdayAbbreviated returns the locales abbreviated weekday given the 'weekday' provided
func (ro *ro) WeekdayAbbreviated(weekday time.Weekday) string {
	return ro.daysAbbreviated[weekday]
}

// WeekdaysAbbreviated returns the locales abbreviated weekdays
func (ro *ro) WeekdaysAbbreviated() []string {
	return ro.daysAbbreviated
}

// WeekdayNarrow returns the locales narrow weekday given the 'weekday' provided
func (ro *ro) WeekdayNarrow(weekday time.Weekday) string {
	return ro.daysNarrow[weekday]
}

// WeekdaysNarrow returns the locales narrow weekdays
func (ro *ro) WeekdaysNarrow() []string {
	return ro.daysNarrow
}

// WeekdayShort returns the locales short weekday given the 'weekday' provided
func (ro *ro) WeekdayShort(weekday time.Weekday) string {
	return ro.daysShort[weekday]
}

// WeekdaysShort returns the locales short weekdays
func (ro *ro) WeekdaysShort() []string {
	return ro.daysShort
}

// WeekdayWide returns the locales wide weekday given the 'weekday' provided
func (ro *ro) WeekdayWide(weekday time.Weekday) string {
	return ro.daysWide[weekday]
}

// WeekdaysWide returns the locales wide weekdays
func (ro *ro) WeekdaysWide() []string {
	return ro.daysWide
}

// Decimal returns the decimal point of number
func (ro *ro) Decimal() string {
	return ro.decimal
}

// Group returns the group of number
func (ro *ro) Group() string {
	return ro.group
}

// Group returns the minus sign of number
func (ro *ro) Minus() string {
	return ro.minus
}

// FmtNumber returns 'num' with digits/precision of 'v' for 'ro' and handles both Whole and Real numbers based on 'v'
func (ro *ro) FmtNumber(num float64, v uint64) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 2 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ro.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, ro.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, ro.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return string(b)
}

// FmtPercent returns 'num' with digits/precision of 'v' for 'ro' and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (ro *ro) FmtPercent(num float64, v uint64) string {
	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	l := len(s) + 5
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ro.decimal[0])
			continue
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, ro.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	b = append(b, ro.percentSuffix...)

	b = append(b, ro.percent...)

	return string(b)
}

// FmtCurrency returns the currency representation of 'num' with digits/precision of 'v' for 'ro'
func (ro *ro) FmtCurrency(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := ro.currencies[currency]
	l := len(s) + len(symbol) + 4 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ro.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, ro.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {
		b = append(b, ro.minus[0])
	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, ro.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	b = append(b, ro.currencyPositiveSuffix...)

	b = append(b, symbol...)

	return string(b)
}

// FmtAccounting returns the currency representation of 'num' with digits/precision of 'v' for 'ro'
// in accounting notation.
func (ro *ro) FmtAccounting(num float64, v uint64, currency currency.Type) string {

	s := strconv.FormatFloat(math.Abs(num), 'f', int(v), 64)
	symbol := ro.currencies[currency]
	l := len(s) + len(symbol) + 6 + 1*len(s[:len(s)-int(v)-1])/3
	count := 0
	inWhole := v == 0
	b := make([]byte, 0, l)

	for i := len(s) - 1; i >= 0; i-- {

		if s[i] == '.' {
			b = append(b, ro.decimal[0])
			inWhole = true
			continue
		}

		if inWhole {
			if count == 3 {
				b = append(b, ro.group[0])
				count = 1
			} else {
				count++
			}
		}

		b = append(b, s[i])
	}

	if num < 0 {

		b = append(b, ro.currencyNegativePrefix[0])

	}

	// reverse
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	if int(v) < 2 {

		if v == 0 {
			b = append(b, ro.decimal...)
		}

		for i := 0; i < 2-int(v); i++ {
			b = append(b, '0')
		}
	}

	if num < 0 {
		b = append(b, ro.currencyNegativeSuffix...)
		b = append(b, symbol...)
	} else {

		b = append(b, ro.currencyPositiveSuffix...)
		b = append(b, symbol...)
	}

	return string(b)
}

// FmtDateShort returns the short date representation of 't' for 'ro'
func (ro *ro) FmtDateShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Day() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x2e}...)

	if t.Month() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Month()), 10)

	b = append(b, []byte{0x2e}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateMedium returns the medium date representation of 't' for 'ro'
func (ro *ro) FmtDateMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, ro.monthsAbbreviated[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateLong returns the long date representation of 't' for 'ro'
func (ro *ro) FmtDateLong(t time.Time) string {

	b := make([]byte, 0, 32)

	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, ro.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtDateFull returns the full date representation of 't' for 'ro'
func (ro *ro) FmtDateFull(t time.Time) string {

	b := make([]byte, 0, 32)

	b = append(b, ro.daysWide[t.Weekday()]...)
	b = append(b, []byte{0x2c, 0x20}...)
	b = strconv.AppendInt(b, int64(t.Day()), 10)
	b = append(b, []byte{0x20}...)
	b = append(b, ro.monthsWide[t.Month()]...)
	b = append(b, []byte{0x20}...)

	if t.Year() > 0 {
		b = strconv.AppendInt(b, int64(t.Year()), 10)
	} else {
		b = strconv.AppendInt(b, int64(-t.Year()), 10)
	}

	return string(b)
}

// FmtTimeShort returns the short time representation of 't' for 'ro'
func (ro *ro) FmtTimeShort(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, ro.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)

	return string(b)
}

// FmtTimeMedium returns the medium time representation of 't' for 'ro'
func (ro *ro) FmtTimeMedium(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, ro.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, ro.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)

	return string(b)
}

// FmtTimeLong returns the long time representation of 't' for 'ro'
func (ro *ro) FmtTimeLong(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, ro.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, ro.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()
	b = append(b, tz...)

	return string(b)
}

// FmtTimeFull returns the full time representation of 't' for 'ro'
func (ro *ro) FmtTimeFull(t time.Time) string {

	b := make([]byte, 0, 32)

	if t.Hour() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Hour()), 10)
	b = append(b, ro.timeSeparator...)

	if t.Minute() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Minute()), 10)
	b = append(b, ro.timeSeparator...)

	if t.Second() < 10 {
		b = append(b, '0')
	}

	b = strconv.AppendInt(b, int64(t.Second()), 10)
	b = append(b, []byte{0x20}...)

	tz, _ := t.Zone()

	if btz, ok := ro.timezones[tz]; ok {
		b = append(b, btz...)
	} else {
		b = append(b, tz...)
	}

	return string(b)
}
